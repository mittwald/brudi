package pgdump_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/mittwald/brudi/pkg/source/pgrestore"
	"github.com/mittwald/brudi/pkg/source/psql"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
	"time"

	"github.com/mittwald/brudi/pkg/source"
	commons "github.com/mittwald/brudi/test/pkg/source/internal"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gotest.tools/assert"
)

const pgPort = "5432/tcp"
const backupPath = "/tmp/postgres.dump.tar"
const backupPathPlain = "/tmp/postgres.dump"
const postgresPW = "postgresroot"
const postgresUser = "postgresuser"
const postgresDB = "postgres"
const tableName = "test"
const dumpKind = "pgdump"
const hostName = "127.0.0.1"
const dbDriver = "pgx"
const pgImage = "quay.io/bitnami/postgresql:latest"
const plainKind = "plain"

type PGDumpTestSuite struct {
	suite.Suite
}

type TestStruct struct {
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

func (pgDumpTestSuite *PGDumpTestSuite) SetupTest() {
	commons.TestSetup()
}

// TearDownTest resets viper after test
func (pgDumpTestSuite *PGDumpTestSuite) TearDownTest() {
	viper.Reset()
}

// TestBasicPGDump performs an integration test for brudi pgdump, without use of restic
func (pgDumpTestSuite *PGDumpTestSuite) TestBasicPGDump() {
	ctx := context.Background()

	defer func() {
		removeErr := os.Remove(backupPath)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to remove pgdump backup files")
		}
	}()
	log.Info("Testing postgres restoration with tar dump via pg_restore")
	testData, err := pgDoBackup(ctx, false, commons.TestContainerSetup{Port: "", Address: ""},
		"tar", backupPath)
	pgDumpTestSuite.Require().NoError(err)
	// setup second postgres container to test if correct data is restored
	var restoreResult []TestStruct
	restoreResult, err = pgDoRestore(ctx, false, commons.TestContainerSetup{Port: "", Address: ""},
		"tar", backupPath)
	pgDumpTestSuite.Require().NoError(err)

	assert.DeepEqual(pgDumpTestSuite.T(), testData, restoreResult)

	log.Info("Testing postgres restoration with plain-text dump via psql")
	var testDataPlain []TestStruct
	testDataPlain, err = pgDoBackup(ctx, false, commons.TestContainerSetup{Port: "", Address: ""},
		"plain", backupPathPlain)
	pgDumpTestSuite.Require().NoError(err)
	// setup second postgres container to test if correct data is restored
	var restoreResultPlain []TestStruct
	restoreResultPlain, err = pgDoRestore(ctx, false, commons.TestContainerSetup{Port: "", Address: ""},
		"plain", backupPathPlain)
	pgDumpTestSuite.Require().NoError(err)

	assert.DeepEqual(pgDumpTestSuite.T(), testDataPlain, restoreResultPlain)
}

// TestPGDumpRestic performs an integration test for brudi pgdump with restic
func (pgDumpTestSuite *PGDumpTestSuite) TestPGDumpRestic() {
	ctx := context.Background()

	defer func() {
		// delete folder with backup file
		removeErr := os.RemoveAll(backupPath)
		log.WithError(removeErr).Error("failed to remove pgdump backup files")
	}()

	// setup a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		resticErr := resticContainer.Container.Terminate(ctx)
		log.WithError(resticErr).Error("failed to terminate pgdump restic container")
	}()

	var testData []TestStruct
	testData, err = pgDoBackup(ctx, true, resticContainer,
		"tar", backupPath)
	pgDumpTestSuite.Require().NoError(err)

	var restoreResult []TestStruct
	restoreResult, err = pgDoRestore(ctx, true, resticContainer,
		"tar", backupPath)
	pgDumpTestSuite.Require().NoError(err)

	assert.DeepEqual(pgDumpTestSuite.T(), testData, restoreResult)
}

func TestPGDumpTestSuite(t *testing.T) {
	suite.Run(t, new(PGDumpTestSuite))
}

// pgDoBackup populates a database with data and performs a backup, optionally with restic
func pgDoBackup(ctx context.Context, useRestic bool,
	resticContainer commons.TestContainerSetup, format, path string) ([]TestStruct, error) {
	// create a postgres container to test backup function
	pgBackupTarget, err := commons.NewTestContainerSetup(ctx, &pgRequest, pgPort)
	if err != nil {
		return []TestStruct{}, err
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
		return []TestStruct{}, err
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
		return []TestStruct{}, err
	}
	// create test data and write it to database
	var testData []TestStruct
	testData, err = prepareTestData(db)
	if err != nil {
		return []TestStruct{}, err
	}

	testPGConfig := createPGConfig(pgBackupTarget, useRestic, resticContainer.Address, resticContainer.Port, format, path)
	err = viper.ReadConfig(bytes.NewBuffer(testPGConfig))
	if err != nil {
		return []TestStruct{}, err
	}

	// perform backup action on first postgres container
	err = source.DoBackupForKind(ctx, dumpKind, false, useRestic, false)
	if err != nil {
		return []TestStruct{}, err
	}

	return testData, nil
}

func pgDoRestore(ctx context.Context, useRestic bool, resticContainer commons.TestContainerSetup,
	format, path string) ([]TestStruct, error) {

	// setup second postgres container to test if correct data is restored
	pgRestoreTarget, err := commons.NewTestContainerSetup(ctx, &pgRequest, pgPort)
	if err != nil {
		return []TestStruct{}, err
	}
	defer func() {
		restoreErr := pgRestoreTarget.Container.Terminate(ctx)
		if restoreErr != nil {
			log.WithError(restoreErr).Error("failed to terminate pgdump restore container")
		}
	}()

	restorePGConfig := createPGConfig(pgRestoreTarget, useRestic, resticContainer.Address, resticContainer.Port, format, path)
	err = viper.ReadConfig(bytes.NewBuffer(restorePGConfig))
	if err != nil {
		return []TestStruct{}, err
	}

	restoreConnectionString := createConnectionString(pgRestoreTarget)
	var dbRestore *sql.DB
	dbRestore, err = sql.Open(dbDriver, restoreConnectionString)
	if err != nil {
		return []TestStruct{}, err
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

	if format == "plain" {
		psqlErr := source.DoRestoreForKind(ctx, psql.Kind, false, useRestic, false)
		if psqlErr != nil {
			return []TestStruct{}, psqlErr
		}
	} else {
		pgErr := source.DoRestoreForKind(ctx, pgrestore.Kind, false, useRestic, false)
		if pgErr != nil {
			return []TestStruct{}, pgErr
		}
	}

	// retrieve data from db
	var result *sql.Rows
	result, err = dbRestore.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		return []TestStruct{}, err
	}
	if result.Err() != nil {
		return []TestStruct{}, result.Err()
	}
	defer func() {
		resultErr := result.Close()
		if resultErr != nil {
			log.WithError(resultErr).Error("failed to cloe pgdump restore result")
		}
	}()

	var restoreResult []TestStruct
	restoreResult, err = scanResult(result)
	if err != nil {
		return []TestStruct{}, err
	}

	return restoreResult, nil
}

// createPGConfig creates a brudi config for the pgdump and the correct restoration command based on format
func createPGConfig(container commons.TestContainerSetup, useRestic bool, resticIP, resticPort, format, path string) []byte {
	var restoreConfig string
	if format != plainKind {
		restoreConfig = fmt.Sprintf(`pgrestore:		
  options:
    flags:
      host: %s
      port: %s
      password: %s
      username: %s
      dbname: %s
    additionalArgs: []
    sourcefile: %s
`, hostName, container.Port, postgresPW, postgresUser, postgresDB, path)
	} else {
		restoreConfig = fmt.Sprintf(`psql:
  options:
    flags:
      host: %s
      port: %s
      user: %s
      password: %s
      dbname: %s
    additionalArgs: []
    sourcefile: %s 
`, hostName, container.Port, postgresUser, postgresPW, postgresDB, path)
	}

	var resticConfig string
	if useRestic {
		resticConfig = fmt.Sprintf(`restic:
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
`, resticIP, resticPort)
	}

	result := []byte(fmt.Sprintf(`
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

`, hostName, container.Port, postgresPW, postgresUser, postgresDB, path, format, restoreConfig, resticConfig))
	return result
}

// prepareTestData creates and isnerts testdata into the specified pg database
func prepareTestData(database *sql.DB) ([]TestStruct, error) {
	var err error
	testStruct1 := TestStruct{2, "TEST"}
	testData := []TestStruct{testStruct1}
	var insert *sql.Rows
	for idx := range testData {
		insert, err = database.Query(fmt.Sprintf("INSERT INTO %s (id, name) VALUES ( %d, '%s' )", tableName, testData[idx].ID, testData[idx].Name))
		if err != nil {
			return []TestStruct{}, err
		}
		if insert.Err() != nil {
			return []TestStruct{}, insert.Err()
		}
	}
	err = insert.Close()
	if err != nil {
		return []TestStruct{}, err
	}
	return testData, nil
}

// scanResult parses the output from a database query back into TestStructs
func scanResult(result *sql.Rows) ([]TestStruct, error) {
	var restoreResult []TestStruct
	for result.Next() {
		var test TestStruct
		err := result.Scan(&test.ID, &test.Name)
		if err != nil {
			return []TestStruct{}, err
		}
		restoreResult = append(restoreResult, test)
	}
	return restoreResult, nil
}

// createConnectionString returns a connection string for the given testcontainer
func createConnectionString(target commons.TestContainerSetup) string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%s database=%s sslmode=disable", postgresUser,
		postgresPW, target.Address, target.Port, postgresDB)
}
