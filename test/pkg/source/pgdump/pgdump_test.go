package pgdump_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
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
const postgresPW = "postgresroot"
const postgresUser = "postgresuser"
const postgresDB = "postgres"
const dataDir = "data"
const tableName = "test"
const restoreKind = "pg_restore"
const dumpKind = "pgdump"
const hostName = "127.0.0.1"
const dbDriver = "pgx"

type PGDumpTestSuite struct {
	suite.Suite
}

type TestStruct struct {
	ID   int
	Name string
}

// testcontainers request for a postgres testcontainer
var pgRequest = testcontainers.ContainerRequest{
	Image:        "quay.io/bitnami/postgresql:latest",
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

// createPGConfig creates a brudi config for the pgdump command
func createPGConfig(container commons.TestContainerSetup, useRestic bool, resticIP, resticPort string) []byte {
	if !useRestic {
		return []byte(fmt.Sprintf(`
pgdump:
  options:
    flags:
      host: %s
      port: %s
      password: %s
      username: %s
      dbName: %s
      file: %s
      format: tar
    additionalArgs: []
`, hostName, container.Port, postgresPW, postgresUser, postgresDB, backupPath))
	}
	return []byte(fmt.Sprintf(`
pgdump:
  options:
    flags:
      host: %s
      port: %s
      password: %s
      username: %s
      dbName: %s
      format: tar
      file: %s
    additionalArgs: []
restic:
  global:
    flags:
      repo: rest:http://%s:%s/
  forget:
    flags:
      keepLast: 1
      keepHourly: 0
      keepDaily: 0
      keepWeekly: 0
      keepMonthly: 0
      keepYearly: 0
`, hostName, container.Port, postgresPW, postgresUser, postgresDB, backupPath, resticIP, resticPort))
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

// restorePGDump pulls data from restic repository and uses pg_restore to restore it to the given databse
func restorePGDump(ctx context.Context, resticContainer, restoreTarget commons.TestContainerSetup) error {
	err := commons.DoResticRestore(ctx, resticContainer, dataDir)
	if err != nil {
		return err
	}
	// restore server from pgdump
	command := exec.CommandContext(ctx, restoreKind, fmt.Sprintf("--dbname=%s", postgresDB),
		fmt.Sprintf("--host=%s", hostName), fmt.Sprintf("--port=%s", restoreTarget.Port),
		fmt.Sprintf("--username=%s", postgresUser), fmt.Sprintf("%s/%s", dataDir, backupPath))
	_, err = command.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func createConnectionString(target commons.TestContainerSetup) string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%s database=%s sslmode=disable", postgresUser,
		postgresPW, target.Address, target.Port, postgresDB)
}

// pgDoBackup populates a database with data and performs a backup, optionally with restic
func pgDoBackup(ctx context.Context, pgDumpTestSuite *PGDumpTestSuite, useRestic bool,
	resticContainer commons.TestContainerSetup) []TestStruct {
	// create a postgres container to test backup function
	pgBackupTarget, err := commons.NewTestContainerSetup(ctx, &pgRequest, pgPort)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		backupErr := pgBackupTarget.Container.Terminate(ctx)
		pgDumpTestSuite.Require().NoError(backupErr)
	}()

	// connect to postgres database using the driver
	backupConnectionString := createConnectionString(pgBackupTarget)
	db, err := sql.Open(dbDriver, backupConnectionString)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		dbErr := db.Close()
		pgDumpTestSuite.Require().NoError(dbErr)
	}()

	// wait for postgres to be ready for connections
	for ok := true; ok; ok = db.Ping() != nil {
		time.Sleep(1 * time.Second)
	}

	// Create test table
	_, err = db.Exec(fmt.Sprintf("CREATE TABLE %s(id serial PRIMARY KEY, name VARCHAR(100) NOT NULL)", tableName))
	pgDumpTestSuite.Require().NoError(err)

	// create test data and write it to database
	testData, err := prepareTestData(db)
	pgDumpTestSuite.Require().NoError(err)

	testPGConfig := createPGConfig(pgBackupTarget, useRestic, resticContainer.Address, resticContainer.Port)
	err = viper.ReadConfig(bytes.NewBuffer(testPGConfig))
	pgDumpTestSuite.Require().NoError(err)

	// perform backup action on first postgres container
	err = source.DoBackupForKind(ctx, dumpKind, false, useRestic, false)
	pgDumpTestSuite.Require().NoError(err)
	return testData
}

// TestBasicPGDump performs an integration test for brudi pgdump, without use of restic
func (pgDumpTestSuite *PGDumpTestSuite) TestBasicPGDump() {
	ctx := context.Background()

	defer func() {
		removeErr := os.Remove(backupPath)
		pgDumpTestSuite.Require().NoError(removeErr)
	}()

	testData := pgDoBackup(ctx, pgDumpTestSuite, false, commons.TestContainerSetup{Port: "", Address: ""})
	// setup second postgres container to test if correct data is restored
	pgRestoreTarget, err := commons.NewTestContainerSetup(ctx, &pgRequest, pgPort)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		restoreErr := pgRestoreTarget.Container.Terminate(ctx)
		pgDumpTestSuite.Require().NoError(restoreErr)
	}()

	restoreConnectionString := createConnectionString(pgRestoreTarget)
	dbRestore, err := sql.Open(dbDriver, restoreConnectionString)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		dbErr := dbRestore.Close()
		pgDumpTestSuite.Require().NoError(dbErr)
	}()

	// wait for postgres to be ready for connections
	for ok := true; ok; ok = dbRestore.Ping() != nil {
		time.Sleep(1 * time.Second)
	}

	// restore server from pgdump
	command := exec.CommandContext(ctx, restoreKind, fmt.Sprintf("--dbname=%s", postgresDB),
		fmt.Sprintf("--host=%s", hostName), fmt.Sprintf("--port=%s", pgRestoreTarget.Port), fmt.Sprintf("--username=%s", postgresUser),
		backupPath)
	_, err = command.CombinedOutput()
	pgDumpTestSuite.Require().NoError(err)

	// check if data was restored correctly
	result, err := dbRestore.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	pgDumpTestSuite.Require().NoError(err)
	pgDumpTestSuite.Require().NoError(result.Err())
	defer func() {
		resultErr := result.Close()
		pgDumpTestSuite.Require().NoError(resultErr)
	}()

	restoreResult, err := scanResult(result)
	pgDumpTestSuite.Require().NoError(err)

	assert.DeepEqual(pgDumpTestSuite.T(), testData, restoreResult)
}

// TestPGDumpRestic performs an integration test for brudi pgdump with restic
func (pgDumpTestSuite *PGDumpTestSuite) TestPGDumpRestic() {
	ctx := context.Background()

	defer func() {
		// delete folder with backup file
		removeErr := os.RemoveAll(dataDir)
		pgDumpTestSuite.Require().NoError(removeErr)
	}()

	// setup a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		resticErr := resticContainer.Container.Terminate(ctx)
		pgDumpTestSuite.Require().NoError(resticErr)
	}()

	testData := pgDoBackup(ctx, pgDumpTestSuite, true, resticContainer)

	// setup second postgres container to test if correct data is restored
	pgRestoreTarget, err := commons.NewTestContainerSetup(ctx, &pgRequest, pgPort)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		restoreErr := pgRestoreTarget.Container.Terminate(ctx)
		pgDumpTestSuite.Require().NoError(restoreErr)
	}()

	restoreConnectionString := createConnectionString(pgRestoreTarget)
	dbRestore, err := sql.Open(dbDriver, restoreConnectionString)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		dbErr := dbRestore.Close()
		pgDumpTestSuite.Require().NoError(dbErr)
	}()

	// wait for postgres to be ready for connections
	for ok := true; ok; ok = dbRestore.Ping() != nil {
		time.Sleep(1 * time.Second)
	}

	err = restorePGDump(ctx, resticContainer, pgRestoreTarget)
	pgDumpTestSuite.Require().NoError(err)

	// check if data was restored correctly
	result, err := dbRestore.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	pgDumpTestSuite.Require().NoError(err)
	pgDumpTestSuite.Require().NoError(result.Err())
	defer func() {
		resultErr := result.Close()
		pgDumpTestSuite.Require().NoError(resultErr)
	}()

	restoreResult, err := scanResult(result)
	pgDumpTestSuite.Require().NoError(err)

	assert.DeepEqual(pgDumpTestSuite.T(), testData, restoreResult)
}

func TestPGDumpTestSuite(t *testing.T) {
	suite.Run(t, new(PGDumpTestSuite))
}
