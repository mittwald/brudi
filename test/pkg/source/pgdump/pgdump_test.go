package pgdump_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

type PGDumpTestSuite struct {
	suite.Suite
}

type TestStruct struct {
	ID   int
	Name string
}

// testcontainers request for a postgres testcontainer
var pgRequest = testcontainers.ContainerRequest{
	Image:        "postgres:12",
	ExposedPorts: []string{pgPort},
	Env: map[string]string{
		"POSTGRES_PASSWORD": "postgresroot",
		"POSTGRES_USER":     "postgresuser",
		"POSTGRES_DB":       "postgres",
	},
	WaitingFor: wait.ForLog("database system is ready to accept connections"),
}

func (pgDumpTestSuite *PGDumpTestSuite) SetupTest() {
	viper.Reset()
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

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
      password: postgresroot
      username: postgresuser
      dbName: postgres
      file: /tmp/postgres.dump.tar
      format: tar
    additionalArgs: []
`, "127.0.0.1", container.Port))
	}
	return []byte(fmt.Sprintf(`
pgdump:
  options:
    flags:
      host: %s
      port: %s
      password: postgresroot
      username: postgresuser
      dbName: postgres
      format: tar
      file: /tmp/postgres.dump.tar
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
`, "127.0.0.1", container.Port, resticIP, resticPort))
}

// prepareTestData creates and isnerts testdata into the specified pg database
func prepareTestData(database *sql.DB) ([]TestStruct, error) {
	var err error
	testStruct1 := TestStruct{2, "TEST"}
	testData := []TestStruct{testStruct1}
	var insert *sql.Rows
	for idx := range testData {
		insert, err = database.Query(fmt.Sprintf("INSERT INTO test (id, name) VALUES ( %d, '%s' )", testData[idx].ID, testData[idx].Name))
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

func restorePGDump(ctx context.Context, resticContainer, restoreTarget commons.TestContainerSetup) error {
	cmd := exec.CommandContext(ctx, "restic", "restore", "-r", fmt.Sprintf("rest:http://%s:%s/",
		resticContainer.Address, resticContainer.Port),
		"--target", "data", "latest")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	// restore server from pgdump
	command := exec.CommandContext(ctx, "pg_restore", "--dbname=postgres",
		"--host=127.0.0.1", fmt.Sprintf("--port=%s", restoreTarget.Port),
		"--username=postgresuser", "data/tmp/postgres.dump.tar")
	_, err = command.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func pgDoBackup(ctx context.Context, pgDumpTestSuite *PGDumpTestSuite, useRestic bool,
	resticContainer commons.TestContainerSetup) []TestStruct {
	// create a postgres container to test backup function
	pgBackupTarget, err := commons.NewTestContainerSetup(ctx, &pgRequest, pgPort)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		err = pgBackupTarget.Container.Terminate(ctx)
		pgDumpTestSuite.Require().NoError(err)
	}()

	// connect to postgres database using the driver
	connectionString := fmt.Sprintf("user=postgresuser password=postgresroot host=%s port=%s database=%s sslmode=disable",
		pgBackupTarget.Address, pgBackupTarget.Port, "postgres")
	db, err := sql.Open("pgx", connectionString)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		err = db.Close()
		pgDumpTestSuite.Require().NoError(err)
	}()

	// these are necessary, otherwise pgserver resets connections
	time.Sleep(1 * time.Second)
	err = db.Ping()
	pgDumpTestSuite.Require().NoError(err)

	// Create test table
	_, err = db.Exec("CREATE TABLE test(id serial PRIMARY KEY, name VARCHAR(100) NOT NULL)")
	pgDumpTestSuite.Require().NoError(err)

	// create test data and write it to database
	testData, err := prepareTestData(db)
	pgDumpTestSuite.Require().NoError(err)

	testPGConfig := createPGConfig(pgBackupTarget, useRestic, resticContainer.Address, resticContainer.Port)
	err = viper.ReadConfig(bytes.NewBuffer(testPGConfig))
	pgDumpTestSuite.Require().NoError(err)

	// perform backup action on first postgres container
	err = source.DoBackupForKind(ctx, "pgdump", false, useRestic, false)
	pgDumpTestSuite.Require().NoError(err)
	return testData
}

// TestBasicPGDump performs an integration test for brudi pgdump, without use of restic
func (pgDumpTestSuite *PGDumpTestSuite) TestBasicPGDump() {
	ctx := context.Background()

	testData := pgDoBackup(ctx, pgDumpTestSuite, false, commons.TestContainerSetup{Port: "", Address: ""})
	// setup second postgres container to test if correct data is restored
	pgRestoreTarget, err := commons.NewTestContainerSetup(ctx, &pgRequest, pgPort)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		err = pgRestoreTarget.Container.Terminate(ctx)
		pgDumpTestSuite.Require().NoError(err)
	}()

	connectionString2 := fmt.Sprintf("user=postgresuser password=postgresroot host=%s port=%s database=%s sslmode=disable",
		pgRestoreTarget.Address, pgRestoreTarget.Port, "postgres")
	dbRestore, err := sql.Open("pgx", connectionString2)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		err = dbRestore.Close()
		pgDumpTestSuite.Require().NoError(err)
	}()

	time.Sleep(1 * time.Second)
	err = dbRestore.Ping()
	pgDumpTestSuite.Require().NoError(err)

	// restore server from pgdump
	command := exec.CommandContext(ctx, "pg_restore", "--dbname=postgres",
		"--host=127.0.0.1", fmt.Sprintf("--port=%s", pgRestoreTarget.Port), "--username=postgresuser", "/tmp/postgres.dump.tar")
	_, err = command.CombinedOutput()
	pgDumpTestSuite.Require().NoError(err)

	err = os.Remove("/tmp/postgres.dump.tar")
	pgDumpTestSuite.Require().NoError(err)

	// check if data was restored correctly
	result, err := dbRestore.Query("SELECT * FROM test")
	pgDumpTestSuite.Require().NoError(err)
	pgDumpTestSuite.Require().NoError(result.Err())
	defer result.Close()

	restoreResult, err := scanResult(result)
	pgDumpTestSuite.Require().NoError(err)

	assert.DeepEqual(pgDumpTestSuite.T(), testData, restoreResult)
}

// TestPGDumpRestic performs an integration test for brudi pgdump with restic
func (pgDumpTestSuite *PGDumpTestSuite) TestPGDumpRestic() {
	ctx := context.Background()

	// setup a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		err = resticContainer.Container.Terminate(ctx)
		pgDumpTestSuite.Require().NoError(err)
	}()

	testData := pgDoBackup(ctx, pgDumpTestSuite, true, resticContainer)

	// setup second postgres container to test if correct data is restored
	pgRestoreTarget, err := commons.NewTestContainerSetup(ctx, &pgRequest, pgPort)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		err = pgRestoreTarget.Container.Terminate(ctx)
		pgDumpTestSuite.Require().NoError(err)
	}()

	connectionString2 := fmt.Sprintf("user=postgresuser password=postgresroot host=%s port=%s database=%s sslmode=disable",
		pgRestoreTarget.Address, pgRestoreTarget.Port, "postgres")
	dbRestore, err := sql.Open("pgx", connectionString2)
	pgDumpTestSuite.Require().NoError(err)
	defer func() {
		err = dbRestore.Close()
		pgDumpTestSuite.Require().NoError(err)
	}()

	time.Sleep(1 * time.Second)
	err = dbRestore.Ping()
	pgDumpTestSuite.Require().NoError(err)

	err = restorePGDump(ctx, resticContainer, pgRestoreTarget)
	pgDumpTestSuite.Require().NoError(err)
	// delete folder with backup file
	err = os.RemoveAll("data")
	pgDumpTestSuite.Require().NoError(err)

	// check if data was restored correctly
	result, err := dbRestore.Query("SELECT * FROM test")
	pgDumpTestSuite.Require().NoError(err)
	pgDumpTestSuite.Require().NoError(result.Err())
	defer result.Close()

	restoreResult, err := scanResult(result)
	pgDumpTestSuite.Require().NoError(err)

	assert.DeepEqual(pgDumpTestSuite.T(), testData, restoreResult)
}

func TestPGDumpTestSuite(t *testing.T) {
	suite.Run(t, new(PGDumpTestSuite))
}
