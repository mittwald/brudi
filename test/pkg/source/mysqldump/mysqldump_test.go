package mysqldump_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mittwald/brudi/pkg/source"
	commons "github.com/mittwald/brudi/test/pkg/source/internal"

	_ "github.com/go-sql-driver/mysql"
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
const dbDriver = "mysql"
const tableName = "testTable"
const hostName = "127.0.0.1" // mysql does not like localhost, therefore use this as address
const logString = "ready for connections"

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
	Image:        "quay.io/bitnami/mysql:latest",
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

// createMySQLConfig creates a brudi config for the `mysqlodump` command.
func createMySQLConfig(container commons.TestContainerSetup, useRestic bool, resticIP, resticPort string) []byte {
	if !useRestic {
		return []byte(fmt.Sprintf(`
mysqldump:
  options:
    flags:
      host: %s
      port: %s
      password: %s
      user: %s
      opt: true
      allDatabases: true
      resultFile: %s
    additionalArgs: []
`, hostName, container.Port, mySQLRootPW, mySQLRoot, backupPath))
	}
	return []byte(fmt.Sprintf(`
mysqldump:
  options:
    flags:
      host: %s
      port: %s
      password: %s
      user: %s
      opt: true
      allDatabases: true
      resultFile: %s
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
`, hostName, container.Port, mySQLRootPW, mySQLRoot, backupPath, resticIP, resticPort))
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

// restoreSQLFromBackup restores the given database from sqldump file
func restoreSQLFromBackup(filename string, database *sql.DB) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	requests := strings.Split(string(file), ";\n")
	for _, request := range requests {
		_, err := database.Exec(request)
		if err != nil {
			return err
		}
	}
	return nil
}

// mySQLDoBackup inserts test data into the given database and then executes brudi's `mysqldump`
func mySQLDoBackup(ctx context.Context, mySQLDumpTestSuite *MySQLDumpTestSuite, useRestic bool,
	resticContainer commons.TestContainerSetup) []TestStruct {

	mySQLBackupTarget, err := commons.NewTestContainerSetup(ctx, &mySQLRequest, sqlPort)
	mySQLDumpTestSuite.Require().NoError(err)
	defer func() {
		backupErr := mySQLBackupTarget.Container.Terminate(ctx)
		mySQLDumpTestSuite.Require().NoError(backupErr)
	}()

	// establish connection
	backupConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=skip-verify",
		mySQLRoot, mySQLRootPW, mySQLBackupTarget.Address, mySQLBackupTarget.Port, mySQLDatabase)
	db, err := sql.Open(dbDriver, backupConnectionString)
	mySQLDumpTestSuite.Require().NoError(err)
	defer func() {
		dbErr := db.Close()
		mySQLDumpTestSuite.Require().NoError(dbErr)
	}()
	time.Sleep(1 * time.Second)
	// create table for test data
	_, err = db.Exec(fmt.Sprintf("CREATE TABLE %s(id INT NOT NULL AUTO_INCREMENT, name VARCHAR(100) NOT NULL, PRIMARY KEY ( id ));", tableName))
	mySQLDumpTestSuite.Require().NoError(err)

	// insert test data
	testData, err := prepareTestData(db)
	mySQLDumpTestSuite.Require().NoError(err)

	MySQLBackupConfig := createMySQLConfig(mySQLBackupTarget, useRestic, resticContainer.Address, resticContainer.Port)
	err = viper.ReadConfig(bytes.NewBuffer(MySQLBackupConfig))
	mySQLDumpTestSuite.Require().NoError(err)

	err = source.DoBackupForKind(ctx, dumpKind, false, useRestic, false)
	mySQLDumpTestSuite.Require().NoError(err)
	return testData
}

// TestBasicMySQLDump performs an integration test for mysqldump, without restic
func (mySQLDumpTestSuite *MySQLDumpTestSuite) TestBasicMySQLDump() {
	ctx := context.Background()

	testData := mySQLDoBackup(ctx, mySQLDumpTestSuite, false, commons.TestContainerSetup{Port: "", Address: ""})

	// setup second mysql container to test if correct data is restored
	mySQLRestoreTarget, err := commons.NewTestContainerSetup(ctx, &mySQLRequest, sqlPort)
	mySQLDumpTestSuite.Require().NoError(err)
	defer func() {
		restoreErr := mySQLRestoreTarget.Container.Terminate(ctx)
		mySQLDumpTestSuite.Require().NoError(restoreErr)
	}()

	// establish connection for restoring data
	restoreConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=skip-verify",
		mySQLRoot, mySQLRootPW, mySQLRestoreTarget.Address, mySQLRestoreTarget.Port, mySQLDatabase)
	dbRestore, err := sql.Open(dbDriver, restoreConnectionString)
	mySQLDumpTestSuite.Require().NoError(err)
	defer func() {
		dbErr := dbRestore.Close()
		mySQLDumpTestSuite.Require().NoError(dbErr)
	}()
	time.Sleep(1 * time.Second)
	// restore server from mysqldump
	err = restoreSQLFromBackup(backupPath, dbRestore)
	mySQLDumpTestSuite.Require().NoError(err)

	// remove backup files
	err = os.Remove(backupPath)
	mySQLDumpTestSuite.Require().NoError(err)

	// check if data was restored correctly
	result, err := dbRestore.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	mySQLDumpTestSuite.Require().NoError(err)
	mySQLDumpTestSuite.Require().NoError(result.Err())
	defer func() {
		resultErr := result.Close()
		mySQLDumpTestSuite.Require().NoError(resultErr)
	}()

	var restoreResult []TestStruct
	for result.Next() {
		var test TestStruct
		err := result.Scan(&test.ID, &test.Name)
		mySQLDumpTestSuite.Require().NoError(err)
		restoreResult = append(restoreResult, test)
	}

	assert.DeepEqual(mySQLDumpTestSuite.T(), testData, restoreResult)
}

// TestMySQLDumpRestic performs an integration test for mysqldump with restic
func (mySQLDumpTestSuite *MySQLDumpTestSuite) TestMySQLDumpRestic() {
	ctx := context.Background()

	defer func() {
		// delete folder with backup file
		removeErr := os.RemoveAll(dataDir)
		mySQLDumpTestSuite.Require().NoError(removeErr)
	}()

	// setup a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	mySQLDumpTestSuite.Require().NoError(err)
	defer func() {
		resticErr := resticContainer.Container.Terminate(ctx)
		mySQLDumpTestSuite.Require().NoError(resticErr)
	}()

	testData := mySQLDoBackup(ctx, mySQLDumpTestSuite, true, resticContainer)

	mySQLRestoreTarget, err := commons.NewTestContainerSetup(ctx, &mySQLRequest, sqlPort)
	mySQLDumpTestSuite.Require().NoError(err)
	defer func() {
		restoreErr := mySQLRestoreTarget.Container.Terminate(ctx)
		mySQLDumpTestSuite.Require().NoError(restoreErr)
	}()

	restoreConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=skip-verify",
		mySQLRoot, mySQLRootPW, mySQLRestoreTarget.Address, mySQLRestoreTarget.Port, mySQLDatabase)
	dbRestore, err := sql.Open(dbDriver, restoreConnectionString)
	mySQLDumpTestSuite.Require().NoError(err)
	defer func() {
		dbErr := dbRestore.Close()
		mySQLDumpTestSuite.Require().NoError(dbErr)
	}()

	time.Sleep(1 * time.Second)
	// restore backup file from restic repository
	err = commons.DoResticRestore(ctx, resticContainer, dataDir)
	mySQLDumpTestSuite.Require().NoError(err)

	err = restoreSQLFromBackup(fmt.Sprintf("%s/%s", dataDir, backupPath), dbRestore)
	mySQLDumpTestSuite.Require().NoError(err)

	// check if data was restored correctly
	result, err := dbRestore.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	mySQLDumpTestSuite.Require().NoError(err)
	mySQLDumpTestSuite.Require().NoError(result.Err())
	defer func() {
		resultErr := result.Close()
		mySQLDumpTestSuite.Require().NoError(resultErr)
	}()

	var restoreResult []TestStruct
	for result.Next() {
		var test TestStruct
		err := result.Scan(&test.ID, &test.Name)
		mySQLDumpTestSuite.Require().NoError(err)
		restoreResult = append(restoreResult, test)
	}

	assert.DeepEqual(mySQLDumpTestSuite.T(), testData, restoreResult)
}

func TestMySQLDumpTestSuite(t *testing.T) {
	suite.Run(t, new(MySQLDumpTestSuite))
}
