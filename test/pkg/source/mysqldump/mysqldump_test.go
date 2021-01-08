package mysqldump_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
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
const mySQLRootPW = "mysqlroot"
const mySQLDatabase = "mysql"
const mySQLUser = "mysqluser"
const mySQLRoot = "root"
const mySQLPw = "mysql"
const dataDir = "data"
const dumpKind = "mysqldump"
const restoreKind = "mysqlrestore"
const dbDriver = "mysql"
const tableName = "testTable"
const hostName = "127.0.0.1" // mysql does not like localhost, therefore use this as address
const logString = "ready for connections"
const mysqlImage = "quay.io/bitnami/mysql:latest"

type MySQLDumpTestSuite struct {
	suite.Suite
}

// struct for test data
type TestStruct struct {
	ID   int
	Name string
}

//v--default-authentication-plugin=mysql_native_password"
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

// SetupTest resets and
func (mySQLDumpTestSuite *MySQLDumpTestSuite) SetupTest() {
	commons.TestSetup()
}

// TearDownTest resets viper after a test
func (mySQLDumpTestSuite *MySQLDumpTestSuite) TearDownTest() {
	viper.Reset()
}

// TestBasicMySQLDump performs an integration test for mysqldump, without restic
func (mySQLDumpTestSuite *MySQLDumpTestSuite) TestBasicMySQLDump() {
	ctx := context.Background()

	defer func() {
		removeErr := os.Remove(backupPath)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to clean up mysql backup files")
		}
	}()

	// backup test data with brudi and remember test data for verification
	testData, err := mySQLDoBackup(ctx, false, commons.TestContainerSetup{Port: "", Address: ""})
	mySQLDumpTestSuite.Require().NoError(err)

	// restore database from backup and pull test data from it for verification
	var restoreResult []TestStruct
	restoreResult, err = mySQLDoRestore(ctx, false, commons.TestContainerSetup{Port: "", Address: ""})
	mySQLDumpTestSuite.Require().NoError(err)

	assert.DeepEqual(mySQLDumpTestSuite.T(), testData, restoreResult)
}

// TestMySQLDumpRestic performs an integration test for mysqldump with restic
func (mySQLDumpTestSuite *MySQLDumpTestSuite) TestMySQLDumpRestic() {
	ctx := context.Background()

	defer func() {
		removeErr := os.Remove(backupPath)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to clean up mysql backup files")
		}
	}()

	// setup a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	mySQLDumpTestSuite.Require().NoError(err)
	defer func() {
		resticErr := resticContainer.Container.Terminate(ctx)
		mySQLDumpTestSuite.Require().NoError(resticErr)
	}()

	// backup test data with brudi and remember test data for verification
	var testData []TestStruct
	testData, err = mySQLDoBackup(ctx, true, resticContainer)

	// restore database from backup and pull test data from it for verification
	var restoreResult []TestStruct
	restoreResult, err = mySQLDoRestore(ctx, true, resticContainer)
	mySQLDumpTestSuite.Require().NoError(err)

	assert.DeepEqual(mySQLDumpTestSuite.T(), testData, restoreResult)
}

func TestMySQLDumpTestSuite(t *testing.T) {
	suite.Run(t, new(MySQLDumpTestSuite))
}

// mySQLDoBackup inserts test data into the given database and then executes brudi's `mysqldump`
func mySQLDoBackup(ctx context.Context, useRestic bool,
	resticContainer commons.TestContainerSetup) ([]TestStruct, error) {

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
	backupConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=skip-verify",
		mySQLRoot, mySQLRootPW, mySQLBackupTarget.Address, mySQLBackupTarget.Port, mySQLDatabase)
	db, err := sql.Open(dbDriver, backupConnectionString)
	if err != nil {
		return []TestStruct{}, err
	}
	defer func() {
		dbErr := db.Close()
		if dbErr != nil {
			log.WithError(dbErr).Error("failed to close connection to mysql backup database")
		}
	}()

	time.Sleep(1 * time.Second)
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
	MySQLBackupConfig := createMySQLConfig(mySQLBackupTarget, useRestic, resticContainer.Address, resticContainer.Port)
	err = viper.ReadConfig(bytes.NewBuffer(MySQLBackupConfig))
	if err != nil {
		return []TestStruct{}, err
	}

	// use brudi to create dump
	err = source.DoBackupForKind(ctx, dumpKind, false, useRestic, false)
	if err != nil {
		return []TestStruct{}, err
	}
	return testData, nil
}

// mySQLDoRestore restores data from backup and retrieves it for verification, optionally using restic
func mySQLDoRestore(ctx context.Context, useRestic bool,
	resticContainer commons.TestContainerSetup) ([]TestStruct, error) {
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

	MySQLRestoreConfig := createMySQLConfig(mySQLRestoreTarget, useRestic, resticContainer.Address, resticContainer.Port)
	err = viper.ReadConfig(bytes.NewBuffer(MySQLRestoreConfig))
	if err != nil {
		return []TestStruct{}, err
	}

	time.Sleep(1 * time.Second)
	// restore server from mysqldump
	err = source.DoRestoreForKind(ctx, restoreKind, false, useRestic, false)
	if err != nil {
		return []TestStruct{}, err
	}

	// establish connection for restoring data
	restoreConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=skip-verify",
		mySQLRoot, mySQLRootPW, mySQLRestoreTarget.Address, mySQLRestoreTarget.Port, mySQLDatabase)
	dbRestore, err := sql.Open(dbDriver, restoreConnectionString)
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
	result, err := dbRestore.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
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
func createMySQLConfig(container commons.TestContainerSetup, useRestic bool, resticIP, resticPort string) []byte {
	var resticConfig string
	if useRestic {
		resticConfig = fmt.Sprintf(`
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
  restore:
    flags:
      target: "/"
    id: "latest"
`, resticIP, resticPort)
	}

	result := []byte(fmt.Sprintf(`
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
    sourceFile: %s%s
`, hostName, container.Port, mySQLRootPW, mySQLRoot, backupPath,
		hostName, container.Port, mySQLRootPW, mySQLRoot, mySQLDatabase, backupPath,
		resticConfig))
	return result
}

// prepareTestData creates test data and inserts it into the given database
func prepareTestData(database *sql.DB) ([]TestStruct, error) {
	var err error
	testStruct1 := TestStruct{2, "TEST"}
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
