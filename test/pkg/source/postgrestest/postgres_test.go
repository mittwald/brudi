package postgres_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mittwald/brudi/pkg/source"
	"github.com/mittwald/brudi/pkg/source/pgrestore"
	"github.com/mittwald/brudi/pkg/source/psql"
	commons "github.com/mittwald/brudi/test/pkg/source/internal"

	_ "github.com/jackc/pgx/stdlib"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gotest.tools/assert"
)

const pgPort = "5432/tcp"
const backupPath = "/tmp/postgres.dump.tar"
const backupPathPlain = "/tmp/postgres.dump"
const backupPathZip = "/tmp/postgres.dump.tar.gz"
const backupPathPlainZip = "/tmp/postgres.dump.gz"
const postgresPW = "postgresroot"
const postgresUser = "postgresuser"
const postgresDB = "postgres"
const tableName = "test"
const dumpKind = "pgdump"

// mysql and psql are a bit picky when it comes to localhost, use ip instead
const hostName = "127.0.0.1"
const dbDriver = "pgx"
const pgImage = "docker.io/bitnami/postgresql:latest"
const plainKind = "plain"

type PGDumpAndRestoreTestSuite struct {
	suite.Suite
}

func (pgDumpAndRestoreTestSuite *PGDumpAndRestoreTestSuite) SetupTest() {
	commons.TestSetup()
}

// TearDownTest resets viper after test
func (pgDumpAndRestoreTestSuite *PGDumpAndRestoreTestSuite) TearDownTest() {
	viper.Reset()
}

// TestBasicPGDump performs an integration test for brudi pgdump, without use of restic
func (pgDumpAndRestoreTestSuite *PGDumpAndRestoreTestSuite) TestBasicPGDumpAndRestore() {
	ctx := context.Background()

	// remove backup files after test
	defer func() {
		removeErr := os.Remove(backupPath)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to remove pgdump backup files")
		}
	}()

	// remove backup files for plain dump after test
	defer func() {
		removePlainErr := os.Remove(backupPathPlain)
		if removePlainErr != nil {
			log.WithError(removePlainErr).Error("failed to remove psql backup files")
		}
	}()

	log.Info("Testing postgres restoration with tar dump via pg_restore")
	// backup test data with brudi and retain test data for verification
	testData, err := pgDoBackup(
		ctx, false, commons.TestContainerSetup{
			Port:    "",
			Address: "",
		},
		"tar", backupPath,
	)
	pgDumpAndRestoreTestSuite.Require().NoError(err)

	// setup second postgres container to test if correct data is restored
	var restoreResult []testStruct
	restoreResult, err = pgDoRestore(
		ctx, false, commons.TestContainerSetup{
			Port:    "",
			Address: "",
		},
		"tar", backupPath,
	)
	pgDumpAndRestoreTestSuite.Require().NoError(err)

	assert.DeepEqual(pgDumpAndRestoreTestSuite.T(), testData, restoreResult)

	log.Info("Testing postgres restoration with plain-text dump via psql")
	var testDataPlain []testStruct
	testDataPlain, err = pgDoBackup(
		ctx, false, commons.TestContainerSetup{
			Port:    "",
			Address: "",
		},
		"plain", backupPathPlain,
	)
	pgDumpAndRestoreTestSuite.Require().NoError(err)

	// restore test data with brudi and retreive it from the db for verification
	var restoreResultPlain []testStruct
	restoreResultPlain, err = pgDoRestore(
		ctx, false, commons.TestContainerSetup{
			Port:    "",
			Address: "",
		},
		"plain", backupPathPlain,
	)
	pgDumpAndRestoreTestSuite.Require().NoError(err)

	assert.DeepEqual(pgDumpAndRestoreTestSuite.T(), testDataPlain, restoreResultPlain)
}

// TestBasicPGDumpZip performs an integration test for brudi pgdump, with gzip and without use of restic
func (pgDumpAndRestoreTestSuite *PGDumpAndRestoreTestSuite) TestBasicPGDumpAndRestoreGzip() {
	ctx := context.Background()

	// remove backup files after test
	defer func() {
		removeErr := os.Remove(backupPathZip)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to remove pgdump backup files")
		}
	}()

	// remove backup files for plain dump after test
	defer func() {
		removePlainErr := os.Remove(backupPathPlainZip)
		if removePlainErr != nil {
			log.WithError(removePlainErr).Error("failed to remove psql backup files")
		}
	}()

	log.Info("Testing postgres restoration with tar dump via pg_restore")
	// backup test data with brudi and retain test data for verification
	testData, err := pgDoBackup(
		ctx, false, commons.TestContainerSetup{
			Port:    "",
			Address: "",
		},
		"tar", backupPathZip,
	)
	pgDumpAndRestoreTestSuite.Require().NoError(err)

	// setup second postgres container to test if correct data is restored
	var restoreResult []testStruct
	restoreResult, err = pgDoRestore(
		ctx, false, commons.TestContainerSetup{
			Port:    "",
			Address: "",
		},
		"tar", backupPathZip,
	)
	pgDumpAndRestoreTestSuite.Require().NoError(err)

	assert.DeepEqual(pgDumpAndRestoreTestSuite.T(), testData, restoreResult)

	log.Info("Testing postgres restoration with plain-text dump via psql")
	var testDataPlain []testStruct
	testDataPlain, err = pgDoBackup(
		ctx, false, commons.TestContainerSetup{
			Port:    "",
			Address: "",
		},
		"plain", backupPathPlainZip,
	)
	pgDumpAndRestoreTestSuite.Require().NoError(err)

	// restore test data with brudi and retreive it from the db for verification
	var restoreResultPlain []testStruct
	restoreResultPlain, err = pgDoRestore(
		ctx, false, commons.TestContainerSetup{
			Port:    "",
			Address: "",
		},
		"plain", backupPathPlainZip,
	)
	pgDumpAndRestoreTestSuite.Require().NoError(err)

	assert.DeepEqual(pgDumpAndRestoreTestSuite.T(), testDataPlain, restoreResultPlain)
}

// TestPGDumpRestic performs an integration test for brudi pgdump with restic
func (pgDumpAndRestoreTestSuite *PGDumpAndRestoreTestSuite) TestPGDumpAndRestoreRestic() {
	ctx := context.Background()

	// remove backup files after test
	defer func() {
		// delete folder with backup file
		removeErr := os.RemoveAll(backupPath)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to remove pgdump backup files")
		}
	}()

	// setup a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	pgDumpAndRestoreTestSuite.Require().NoError(err)
	defer func() {
		resticErr := resticContainer.Container.Terminate(ctx)
		if resticErr != nil {
			log.WithError(resticErr).Error("failed to terminate pgdump restic container")
		}
	}()

	// backup test data with brudi and retain test data for verification
	var testData []testStruct
	testData, err = pgDoBackup(
		ctx, true, resticContainer,
		"tar", backupPath,
	)
	pgDumpAndRestoreTestSuite.Require().NoError(err)

	// restore test data with brudi and retrieve it from the db for verification
	var restoreResult []testStruct
	restoreResult, err = pgDoRestore(
		ctx, true, resticContainer,
		"tar", backupPath,
	)
	pgDumpAndRestoreTestSuite.Require().NoError(err)

	assert.DeepEqual(pgDumpAndRestoreTestSuite.T(), testData, restoreResult)
}

// TestPGDumpResticGzip performs an integration test for brudi pgdump with restic and gzip
func (pgDumpAndRestoreTestSuite *PGDumpAndRestoreTestSuite) TestPGDumpAndRestoreResticGzip() {
	ctx := context.Background()

	// remove backup files after test
	defer func() {
		// delete folder with backup file
		removeErr := os.RemoveAll(backupPathZip)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to remove pgdump backup files")
		}
	}()

	// setup a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	pgDumpAndRestoreTestSuite.Require().NoError(err)
	defer func() {
		resticErr := resticContainer.Container.Terminate(ctx)
		if resticErr != nil {
			log.WithError(resticErr).Error("failed to terminate pgdump restic container")
		}
	}()

	// backup test data with brudi and retain test data for verification
	var testData []testStruct
	testData, err = pgDoBackup(
		ctx, true, resticContainer,
		"tar", backupPathZip,
	)
	pgDumpAndRestoreTestSuite.Require().NoError(err)

	// restore test data with brudi and retrieve it from the db for verification
	var restoreResult []testStruct
	restoreResult, err = pgDoRestore(
		ctx, true, resticContainer,
		"tar", backupPathZip,
	)
	pgDumpAndRestoreTestSuite.Require().NoError(err)

	assert.DeepEqual(pgDumpAndRestoreTestSuite.T(), testData, restoreResult)
}

func TestPGDumpAndRestoreTestSuite(t *testing.T) {
	suite.Run(t, new(PGDumpAndRestoreTestSuite))
}

// pgDoBackup populates a database with data and performs a backup, optionally with restic
func pgDoBackup(
	ctx context.Context, useRestic bool,
	resticContainer commons.TestContainerSetup, format, path string,
) ([]testStruct, error) {
	// create a postgres container to test backup function
	pgBackupTarget, err := commons.NewTestContainerSetup(ctx, &pgRequest, pgPort)
	if err != nil {
		return []testStruct{}, err
	}
	defer func() {
		backupErr := pgBackupTarget.Container.Terminate(ctx)
		if backupErr != nil {
			log.WithError(backupErr).Error("failed to termninate pgdump backup container")
		}
	}()

	// connect to postgres database using the driver
	backupConnectionString := createConnectionString(pgBackupTarget)
	var db *sql.DB
	db, err = sql.Open(dbDriver, backupConnectionString)
	if err != nil {
		return []testStruct{}, err
	}
	defer func() {
		dbErr := db.Close()
		if dbErr != nil {
			log.WithError(dbErr).Error("failed to close connection to pgdump backup database")
		}
	}()

	// wait for postgres to be ready for connections
	for ok := true; ok; ok = db.Ping() != nil {
		time.Sleep(1 * time.Second)
	}

	// Create test table
	_, err = db.Exec(fmt.Sprintf("CREATE TABLE %s(id serial PRIMARY KEY, name VARCHAR(100) NOT NULL)", tableName))
	if err != nil {
		return []testStruct{}, err
	}
	// create test data and write it to database
	var testData []testStruct
	testData, err = prepareTestData(db)
	if err != nil {
		return []testStruct{}, err
	}

	// create a brudi config for pgdump
	testPGConfig := createPGConfig(pgBackupTarget, useRestic, resticContainer.Address, resticContainer.Port, format, path)
	err = viper.ReadConfig(bytes.NewBuffer(testPGConfig))
	if err != nil {
		return []testStruct{}, err
	}

	// perform backup action on database
	err = source.DoBackupForKind(ctx, dumpKind, false, useRestic, false, false)
	if err != nil {
		return []testStruct{}, err
	}

	return testData, nil
}

// pgDoRestore restores data from backup and retrieves it for verification, optionally using restic
func pgDoRestore(
	ctx context.Context, useRestic bool, resticContainer commons.TestContainerSetup,
	format, path string,
) ([]testStruct, error) {

	// setup second postgres container to test if correct data is restored
	pgRestoreTarget, err := commons.NewTestContainerSetup(ctx, &pgRequest, pgPort)
	if err != nil {
		return []testStruct{}, err
	}
	defer func() {
		restoreErr := pgRestoreTarget.Container.Terminate(ctx)
		if restoreErr != nil {
			log.WithError(restoreErr).Error("failed to terminate pgdump restore container")
		}
	}()

	// create a brudi configuration for pgrestore, depending on backup format
	restorePGConfig := createPGConfig(pgRestoreTarget, useRestic, resticContainer.Address, resticContainer.Port, format, path)
	err = viper.ReadConfig(bytes.NewBuffer(restorePGConfig))
	if err != nil {
		return []testStruct{}, err
	}

	// connect to restored database
	restoreConnectionString := createConnectionString(pgRestoreTarget)
	var dbRestore *sql.DB
	dbRestore, err = sql.Open(dbDriver, restoreConnectionString)
	if err != nil {
		return []testStruct{}, err
	}
	defer func() {
		dbErr := dbRestore.Close()
		if dbErr != nil {
			log.WithError(dbErr).Error("failed to disconnect from pgdump restore database")
		}
	}()

	// wait for postgres to be ready for connections
	for ok := true; ok; ok = dbRestore.Ping() != nil {
		time.Sleep(1 * time.Second)
	}

	// use correct restoration function based on backup format
	if format == "plain" {
		psqlErr := source.DoRestoreForKind(ctx, psql.Kind, false, useRestic)
		if psqlErr != nil {
			return []testStruct{}, psqlErr
		}
	} else {
		pgErr := source.DoRestoreForKind(ctx, pgrestore.Kind, false, useRestic)
		if pgErr != nil {
			return []testStruct{}, pgErr
		}
	}

	// retrieve data from db
	var result *sql.Rows
	result, err = dbRestore.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		return []testStruct{}, err
	}
	if result.Err() != nil {
		return []testStruct{}, result.Err()
	}
	defer func() {
		resultErr := result.Close()
		if resultErr != nil {
			log.WithError(resultErr).Error("failed to cloe pgdump restore result")
		}
	}()

	// scan query result into testStructs
	var restoreResult []testStruct
	restoreResult, err = scanResult(result)
	if err != nil {
		return []testStruct{}, err
	}

	return restoreResult, nil
}

// createPGConfig creates a brudi config for the pgdump and the correct restoration command based on format
func createPGConfig(container commons.TestContainerSetup, useRestic bool, resticIP, resticPort, format, path string) []byte {
	var restoreConfig string
	if format != plainKind {
		restoreConfig = fmt.Sprintf(
			`pgrestore:		
  options:
    flags:
      host: %s
      port: %s
      password: %s
      username: %s
      dbname: %s
    additionalArgs: []
    sourcefile: %s
`, hostName, container.Port, postgresPW, postgresUser, postgresDB, path,
		)
	} else {
		restoreConfig = fmt.Sprintf(
			`psql:
  options:
    flags:
      host: %s
      port: %s
      user: %s
      password: %s
      dbname: %s
    additionalArgs: []
    sourcefile: %s 
`, hostName, container.Port, postgresUser, postgresPW, postgresDB, path,
		)
	}

	var resticConfig string
	if useRestic {
		resticConfig = fmt.Sprintf(
			`restic:
  global:
    flags:
      repo: rest:http://%s:%s/
  forget:
    flags:
      keepLast: 1
      keepHourly: 0
      keepDaily: 0
      eepWeekly: 0
      keepMonthly: 0
      keepYearly: 0
  restore:
    flags:
      target: "/"
    id: "latest"
`, resticIP, resticPort,
		)
	}

	result := []byte(fmt.Sprintf(
		`
pgdump:
  options:
    flags:
      host: %s
      port: %s
      password: %s
      username: %s
      dbName: %s
      file: %s
      format: %s
    additionalArgs: []
%s
%s

`, hostName, container.Port, postgresPW, postgresUser, postgresDB, path, format, restoreConfig, resticConfig,
	))
	return result
}

// prepareTestData creates and inserts testdata into the specified pg database
func prepareTestData(database *sql.DB) ([]testStruct, error) {
	var err error
	testStruct1 := testStruct{
		2,
		"TEST",
	}
	testData := []testStruct{testStruct1}
	var insert *sql.Rows
	for idx := range testData {
		insert, err = database.Query(fmt.Sprintf("INSERT INTO %s (id, name) VALUES ( %d, '%s' )", tableName, testData[idx].ID, testData[idx].Name))
		if err != nil {
			return []testStruct{}, err
		}
		if insert.Err() != nil {
			return []testStruct{}, insert.Err()
		}
	}
	err = insert.Close()
	if err != nil {
		return []testStruct{}, err
	}
	return testData, nil
}

// scanResult parses the output from a database query back into TestStructs
func scanResult(result *sql.Rows) ([]testStruct, error) {
	var restoreResult []testStruct
	for result.Next() {
		var test testStruct
		err := result.Scan(&test.ID, &test.Name)
		if err != nil {
			return []testStruct{}, err
		}
		restoreResult = append(restoreResult, test)
	}
	return restoreResult, nil
}

// createConnectionString returns a connection string for the given testcontainer
func createConnectionString(target commons.TestContainerSetup) string {
	return fmt.Sprintf(
		"user=%s password=%s host=%s port=%s database=%s sslmode=disable", postgresUser,
		postgresPW, target.Address, target.Port, postgresDB,
	)
}

type testStruct struct {
	ID   int
	Name string
}

// testcontainers request for a postgres testcontainer
var pgRequest = testcontainers.ContainerRequest{
	Image:        pgImage,
	ExposedPorts: []string{pgPort},
	Env: map[string]string{
		"POSTGRES_PASSWORD": postgresPW,
		"POSTGRES_USER":     postgresUser,
		"POSTGRES_DB":       postgresDB,
	},
	WaitingFor: wait.ForLog("database system is ready to accept connections"),
}
