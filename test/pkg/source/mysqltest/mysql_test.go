package mysql_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"testing"
	"time"

	"github.com/mittwald/brudi/pkg/source"
	commons "github.com/mittwald/brudi/test/pkg/source/internal"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gotest.tools/assert"
)

const sqlPort = "3306/tcp"
const backupPath = "/tmp/test.sqldump"
const backupPathZip = "/tmp/test.sqldump.gz"
const mySQLRootPW = "mysqlroot"
const mySQLDatabase = "mysql"
const mySQLUser = "mysqluser"
const mySQLRoot = "root"
const mySQLPw = "mysql"
const dumpKind = "mysqldump"
const restoreKind = "mysqlrestore"
const dbDriver = "mysql"
const tableName = "testTable"

// mysql and psql are a bit picky when it comes to localhost, use ip instead
const hostName = "127.0.0.1"
const logString = "ready for connections"
const mysqlImage = "docker.io/bitnami/mysql:latest"

type MySQLDumpAndRestoreTestSuite struct {
	suite.Suite
	resticExists bool
}

// struct for test data
type TestStruct struct {
	ID   int
	Name string
}

// SetupTest resets and
func (mySQLDumpAndRestoreTestSuite *MySQLDumpAndRestoreTestSuite) SetupTest() {
	commons.TestSetup()
}

// TearDownTest resets viper after a test
func (mySQLDumpAndRestoreTestSuite *MySQLDumpAndRestoreTestSuite) TearDownTest() {
	viper.Reset()
}

// TestBasicMySQLDump performs an integration test for mysqldump, without restic
func (mySQLDumpAndRestoreTestSuite *MySQLDumpAndRestoreTestSuite) TestBasicMySQLDumpAndRestore() {
	ctx := context.Background()

	// remove backup files after test
	defer func() {
		removeErr := os.Remove(backupPath)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to clean up mysql backup files")
		}
	}()

	// backup test data with brudi and retain test data for verification
	testData, err := mySQLDoBackup(
		ctx, false, commons.TestContainerSetup{
			Port:    "",
			Address: "",
		}, backupPath, false,
	)
	mySQLDumpAndRestoreTestSuite.Require().NoError(err)

	// restore test data with brudi and retrieve it from the db for verification
	var restoreResult []TestStruct
	restoreResult, err = mySQLDoRestore(
		ctx, false, commons.TestContainerSetup{
			Port:    "",
			Address: "",
		}, backupPath,
	)
	mySQLDumpAndRestoreTestSuite.Require().NoError(err)

	assert.DeepEqual(mySQLDumpAndRestoreTestSuite.T(), testData, restoreResult)
}

// TestBasicMySQLDumpGzip performs an integration test for mysqldump, with gzip and without restic
func (mySQLDumpAndRestoreTestSuite *MySQLDumpAndRestoreTestSuite) TestBasicMySQLDumpAndRestoreGzip() {
	ctx := context.Background()

	// remove backup files after test
	defer func() {
		removeErr := os.Remove(backupPathZip)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to clean up mysql backup files")
		}
	}()

	// backup test data with brudi and retain test data for verification
	testData, err := mySQLDoBackup(
		ctx, false, commons.TestContainerSetup{
			Port:    "",
			Address: "",
		}, backupPathZip, false,
	)
	mySQLDumpAndRestoreTestSuite.Require().NoError(err)

	// restore test data with brudi and retrieve it from the db for verification
	var restoreResult []TestStruct
	restoreResult, err = mySQLDoRestore(
		ctx, false, commons.TestContainerSetup{
			Port:    "",
			Address: "",
		}, backupPathZip,
	)
	mySQLDumpAndRestoreTestSuite.Require().NoError(err)

	assert.DeepEqual(mySQLDumpAndRestoreTestSuite.T(), testData, restoreResult)
}

func (mySQLDumpAndRestoreTestSuite *MySQLDumpAndRestoreTestSuite) mySQLDumpAndRestoreRestic(backupPath string, useStdin bool) {
	mySQLDumpAndRestoreTestSuite.True(mySQLDumpAndRestoreTestSuite.resticExists, "can't use restic on this machine")
	if !mySQLDumpAndRestoreTestSuite.resticExists {
		return
	}
	ctx := context.Background()

	defer func() {
		removeErr := os.Remove(backupPath)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to clean up mysql backup files")
		}
	}()

	// setup a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	mySQLDumpAndRestoreTestSuite.Require().NoError(err)
	defer func() {
		resticErr := resticContainer.Container.Terminate(ctx)
		if resticErr != nil {
			log.WithError(resticErr).Error("failed to terminate mysql restic container")
		}
	}()

	// backup test data with brudi and retain test data for verification
	var testData []TestStruct
	testData, err = mySQLDoBackup(ctx, true, resticContainer, backupPath, useStdin)
	mySQLDumpAndRestoreTestSuite.Require().NoError(err)

	// restore database from backup and pull test data from it for verification
	var restoreResult []TestStruct
	restoreResult, err = mySQLDoRestore(ctx, true, resticContainer, backupPath)
	mySQLDumpAndRestoreTestSuite.Require().NoError(err)

	assert.DeepEqual(mySQLDumpAndRestoreTestSuite.T(), testData, restoreResult)
}

// TestMySQLDumpRestic performs an integration test for mysqldump with restic
func (mySQLDumpAndRestoreTestSuite *MySQLDumpAndRestoreTestSuite) TestMySQLDumpAndRestoreRestic() {
	mySQLDumpAndRestoreTestSuite.mySQLDumpAndRestoreRestic(backupPath, false)
}

// TestMySQLDumpResticGzip performs an integration test for mysqldump with restic and gzip
func (mySQLDumpAndRestoreTestSuite *MySQLDumpAndRestoreTestSuite) TestMySQLDumpAndRestoreResticGzip() {
	mySQLDumpAndRestoreTestSuite.mySQLDumpAndRestoreRestic(backupPathZip, false)
}

// TestMySQLDumpResticStdin performs an integration test for mysqldump with restic using STDIN
func (mySQLDumpAndRestoreTestSuite *MySQLDumpAndRestoreTestSuite) TestMySQLDumpAndRestoreResticStdin() {
	mySQLDumpAndRestoreTestSuite.mySQLDumpAndRestoreRestic(backupPath, true)
}

func TestMySQLDumpAndRestoreTestSuite(t *testing.T) {
	_, resticExists := commons.CheckProgramsAndRestic(t, "mysqldump", "", "mysql", "")
	testSuite := &MySQLDumpAndRestoreTestSuite{
		resticExists: resticExists,
	}
	suite.Run(t, testSuite)
}

// mySQLDoBackup inserts test data into the given database and then executes brudi's `mysqldump`
func mySQLDoBackup(
	ctx context.Context, useRestic bool,
	resticContainer commons.TestContainerSetup, path string, useStdinBackup bool,
) ([]TestStruct, error) {
	// setup a mysql container to backup from
	mySQLBackupTarget, err := commons.NewTestContainerSetup(ctx, &mySQLRequest, sqlPort)
	if err != nil {
		return []TestStruct{}, err
	}
	defer func() {
		backupErr := mySQLBackupTarget.Container.Terminate(ctx)
		if backupErr != nil {
			log.WithError(backupErr).Error("failed to terminate mysql backup container")
		}
	}()

	// establish connection
	backupConnectionString := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?tls=skip-verify",
		mySQLRoot, mySQLRootPW, mySQLBackupTarget.Address, mySQLBackupTarget.Port, mySQLDatabase,
	)
	var db *sql.DB
	db, err = sql.Open(dbDriver, backupConnectionString)
	if err != nil {
		return []TestStruct{}, err
	}
	defer func() {
		dbErr := db.Close()
		if dbErr != nil {
			log.WithError(dbErr).Error("failed to close connection to mysql backup database")
		}
	}()
	err = waitForDb(db)
	if err != nil {
		return []TestStruct{}, err
	}

	// create table for test data
	_, err = db.Exec(fmt.Sprintf("CREATE TABLE %s(id INT NOT NULL AUTO_INCREMENT, name VARCHAR(100) NOT NULL, PRIMARY KEY ( id ));", tableName))
	if err != nil {
		return []TestStruct{}, err
	}

	// insert test data
	testData, err := prepareTestData(db)
	if err != nil {
		return []TestStruct{}, err
	}

	// create brudi config for mysqldump
	MySQLBackupConfig := createMySQLConfig(mySQLBackupTarget, useRestic, resticContainer.Address, resticContainer.Port, path, useStdinBackup)
	err = viper.ReadConfig(bytes.NewBuffer(MySQLBackupConfig))
	if err != nil {
		return []TestStruct{}, err
	}

	// use brudi to create dump
	err = source.DoBackupForKind(ctx, dumpKind, false, useRestic, false, false)
	if err != nil {
		return []TestStruct{}, err
	}
	return testData, nil
}

// mySQLDoRestore restores data from backup and retrieves it for verification, optionally using restic
func mySQLDoRestore(
	ctx context.Context, useRestic bool,
	resticContainer commons.TestContainerSetup, path string,
) ([]TestStruct, error) {
	// create a mysql container to restore data to
	mySQLRestoreTarget, err := commons.NewTestContainerSetup(ctx, &mySQLRequest, sqlPort)
	if err != nil {
		return []TestStruct{}, err
	}
	defer func() {
		restoreErr := mySQLRestoreTarget.Container.Terminate(ctx)
		if restoreErr != nil {
			log.WithError(restoreErr).Error("failed to terminate mysql restore container")
		}
	}()

	// create a brudi config for mysql restore
	MySQLRestoreConfig := createMySQLConfig(mySQLRestoreTarget, useRestic, resticContainer.Address, resticContainer.Port, path, false)
	err = viper.ReadConfig(bytes.NewBuffer(MySQLRestoreConfig))
	if err != nil {
		return []TestStruct{}, err
	}

	// establish connection to be able to wait for the DB and retrieving restored data
	restoreConnectionString := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?tls=skip-verify",
		mySQLRoot, mySQLRootPW, mySQLRestoreTarget.Address, mySQLRestoreTarget.Port, mySQLDatabase,
	)
	var dbRestore *sql.DB
	dbRestore, err = sql.Open(dbDriver, restoreConnectionString)
	if err != nil {
		return []TestStruct{}, err
	}
	defer func() {
		dbErr := dbRestore.Close()
		if dbErr != nil {
			log.WithError(dbErr).Error("failed to close connection to mysql backup database")
		}
	}()
	err = waitForDb(dbRestore)
	if err != nil {
		return []TestStruct{}, err
	}

	// restore server from mysqldump
	err = source.DoRestoreForKind(ctx, restoreKind, false, useRestic)
	if err != nil {
		return []TestStruct{}, err
	}

	// attempt to retrieve test data from database
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
			log.WithError(resultErr).Error("failed to close mysql restore result")
		}
	}()

	// convert mysql result into a list of TestStructs
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

// createMySQLConfig creates a brudi config for mysqldump and mysqlrestore command.
func createMySQLConfig(container commons.TestContainerSetup, useRestic bool, resticIP, resticPort, filepath string, doStdinBackup bool) []byte {
	var resticConfig string
	if useRestic {
		//restoreTarget := "/"
		stdinFilename := ""
		if doStdinBackup {
			stdinFilename = fmt.Sprintf("  backup:\n    flags:\n      stdinFilename: %s\n", filepath)
			//restoreTarget = path.Join(restoreTarget, filepath)
		}
		resticConfig = fmt.Sprintf(
			`doPipingBackup: %t
restic:
  global:
    flags:
      repo: rest:http://%s:%s/
%s  forget:
    flags:
      keepLast: 1
      keepHourly: 0
      keepDaily: 0
      keepWeekly: 0
      keepMonthly: 0
      keepYearly: 0
  restore:
    flags:
      target: "/"
    id: "latest"
`, doStdinBackup, resticIP, resticPort, stdinFilename,
		)
	}

	result := []byte(fmt.Sprintf(
		`
mysqldump:
  options:
    flags:
      host: %s
      port: %s
      password: %s
      user: %s
      opt: true
      force: true
      allDatabases: true
      resultFile: %s
    additionalArgs: []
mysqlrestore:
  options:
    flags:
      host: %s
      port: %s
      password: %s
      user: %s
      Database: %s
    additionalArgs: []
    sourceFile: %s
%s
`, hostName, container.Port, mySQLRootPW, mySQLRoot, filepath,
		hostName, container.Port, mySQLRootPW, mySQLRoot, mySQLDatabase, filepath,
		resticConfig,
	))
	return result
}

func waitForDb(db *sql.DB) error {
	var err error
	// sleep to give mysql server time to get ready
	time.Sleep(10 * time.Second)
	// Ping until ready or 30 seconds are over
	for i := 0; i < 20; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return errors.Wrap(err, "can't ping database after 30 seconds")
	}
	return nil
}

// prepareTestData creates test data and inserts it into the given database
func prepareTestData(database *sql.DB) ([]TestStruct, error) {
	var err error
	testStruct1 := TestStruct{
		2,
		"TEST",
	}
	testData := []TestStruct{testStruct1}
	var insert *sql.Rows
	for idx := range testData {
		insert, err = database.Query(fmt.Sprintf("INSERT INTO %s VALUES ( %d, '%s' )", tableName, testData[idx].ID, testData[idx].Name))
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

// testcontainer request for a mysql container
var mySQLRequest = testcontainers.ContainerRequest{
	Image:        mysqlImage,
	ExposedPorts: []string{sqlPort},
	Env: map[string]string{
		"MYSQL_ROOT_PASSWORD": mySQLRootPW,
		"MYSQL_DATABASE":      mySQLDatabase,
		"MYSQL_USER":          mySQLUser,
		"MYSQL_PASSWORD":      mySQLPw,
		"MYSQL_EXTRA:FLAGS":   "--default-authentication-plugin=mysql_native_password",
	},
	WaitingFor: wait.ForLog(logString),
}
