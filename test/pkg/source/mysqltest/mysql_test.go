package mysql_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"reflect"
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
const mysqlImage = "docker.io/bitnami/mysql:5.7"

type MySQLDumpAndRestoreTestSuite struct {
	suite.Suite
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
		}, backupPath,
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
		}, backupPathZip,
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

// TestMySQLDumpRestic performs an integration test for mysqldump with restic
func (mySQLDumpAndRestoreTestSuite *MySQLDumpAndRestoreTestSuite) TestMySQLDumpAndRestoreRestic() {
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
	testData, backupErr := mySQLDoBackup(ctx, true, resticContainer, backupPath)
	mySQLDumpAndRestoreTestSuite.Require().NoError(backupErr)

	// restore database from backup and pull test data from it for verification
	restoreResult, restoreErr := mySQLDoRestore(ctx, true, resticContainer, backupPath)
	mySQLDumpAndRestoreTestSuite.Require().NoError(restoreErr)

	assert.DeepEqual(mySQLDumpAndRestoreTestSuite.T(), testData, restoreResult)
}

// TestMySQLDumpResticGzip performs an integration test for mysqldump with restic and gzip
func (mySQLDumpAndRestoreTestSuite *MySQLDumpAndRestoreTestSuite) TestMySQLDumpAndRestoreResticGzip() {
	ctx := context.Background()

	defer func() {
		removeErr := os.Remove(backupPathZip)
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
	testData, backupErr := mySQLDoBackup(ctx, true, resticContainer, backupPathZip)
	mySQLDumpAndRestoreTestSuite.Require().NoError(backupErr)

	// restore database from backup and pull test data from it for verification
	restoreResult, restoreErr := mySQLDoRestore(ctx, true, resticContainer, backupPathZip)
	mySQLDumpAndRestoreTestSuite.Require().NoError(restoreErr)

	mySQLDumpAndRestoreTestSuite.Require().True(reflect.DeepEqual(testData, restoreResult))
}

func TestMySQLDumpAndRestoreTestSuite(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	suite.Run(t, new(MySQLDumpAndRestoreTestSuite))
}

// mySQLDoBackup inserts test data into the given database and then executes brudi's `mysqldump`
func mySQLDoBackup(
	ctx context.Context, useRestic bool,
	resticContainer commons.TestContainerSetup, path string,
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
	// sleep to give mysql server time to get ready
	time.Sleep(5 * time.Second)

	// create table for test data
	_, err = db.Exec(
		fmt.Sprintf(
			"CREATE TABLE %s(id INT NOT NULL AUTO_INCREMENT, name VARCHAR(100) NOT NULL, PRIMARY KEY ( id ));",
			tableName,
		),
	)
	if err != nil {
		return []TestStruct{}, err
	}

	// insert test data
	testData, err := prepareTestData(db)
	if err != nil {
		return []TestStruct{}, err
	}

	// create brudi config for mysqldump
	MySQLBackupConfig := createMySQLConfig(
		mySQLBackupTarget,
		useRestic,
		resticContainer.Address,
		resticContainer.Port,
		path,
	)
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
	MySQLRestoreConfig := createMySQLConfig(
		mySQLRestoreTarget,
		useRestic,
		resticContainer.Address,
		resticContainer.Port,
		path,
	)
	err = viper.ReadConfig(bytes.NewBuffer(MySQLRestoreConfig))
	if err != nil {
		return []TestStruct{}, err
	}

	// sleep to give mysql time to get ready
	time.Sleep(1 * time.Second)

	// restore server from mysqldump
	err = source.DoBackupForKind(ctx, dumpKind, false, useRestic, false, false)
	if err != nil {
		return []TestStruct{}, err
	}

	// establish connection for retrieving restored data
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
			log.WithError(dbErr).Error("failed to close connection to mysql restore database")
		}
	}()

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
func createMySQLConfig(container commons.TestContainerSetup, useRestic bool, resticIP, resticPort, path string) []byte {
	var resticConfig string
	if useRestic {
		resticConfig = fmt.Sprintf(
			`
restic:
  global:
    flags:
      repo: rest:http://%s:%s/
      skipSsl: "foo"
  forget:
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
`, resticIP, resticPort,
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
      skipSsl: true
mysqlrestore:
  options:
    flags:
      host: %s
      port: %s
      password: %s
      user: %s
      Database: %s
      skipSsl: true
    sourceFile: %s%s
`, hostName, container.Port, mySQLRootPW, mySQLRoot, path,
		hostName, container.Port, mySQLRootPW, mySQLRoot, mySQLDatabase, path,
		resticConfig,
	))
	return result
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
	defer func() {
		err = insert.Close()
		if err != nil {
			log.WithError(err).Error("failed to close insert")
		}
	}()
	for idx := range testData {
		insert, err = database.Query(
			fmt.Sprintf(
				"INSERT INTO %s VALUES ( %d, '%s' )",
				tableName,
				testData[idx].ID,
				testData[idx].Name,
			),
		)
		if err != nil {
			return []TestStruct{}, err
		}
		if insert.Err() != nil {
			return []TestStruct{}, insert.Err()
		}
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
		"JDBC_PARAMS":         "useSSL=false",
		"MYSQL_EXTRA:FLAGS":   "--default-authentication-plugin=mysql_native_password --skip-ssl",
	},
	WaitingFor: wait.ForLog(logString),
}
