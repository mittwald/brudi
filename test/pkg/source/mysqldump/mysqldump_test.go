package mysqldump_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

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
const mySQLPW = "mysql"
const dataDir = "data"

type MySQLDumpTestSuite struct {
	suite.Suite
}

// struct for test data
type TestStruct struct {
	ID   int
	Name string
}

// testcontainer request for a mysql container
var mySQLRequest = testcontainers.ContainerRequest{
	Image:        "mysql:8",
	ExposedPorts: []string{sqlPort},
	Env: map[string]string{
		"MYSQL_ROOT_PASSWORD": mySQLRootPW,
		"MYSQL_DATABASE":      mySQLDatabase,
		"MYSQL_USER":          mySQLUser,
		"MYSQL_PASSWORD":      mySQLPW,
	},
	Cmd:        []string{"--default-authentication-plugin=mysql_native_password"},
	WaitingFor: wait.ForLog("port: 3306  MySQL Community Server - GPL"),
}

func (mySQLDumpTestSuite *MySQLDumpTestSuite) SetupTest() {
	viper.Reset()
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func (mySQLDumpTestSuite *MySQLDumpTestSuite) TearDownTest() {
	viper.Reset()
}

// createMySQLConfig creates a brudi config for the mysqlodump command.
func createMySQLConfig(container commons.TestContainerSetup, useRestic bool, resticIP, resticPort string) []byte {
	fmt.Println(resticIP)
	fmt.Println(resticPort)
	if !useRestic {
		return []byte(fmt.Sprintf(`
mysqldump:
  options:
    flags:
      host: %s
      port: %s
      password: %s
      user: root
      opt: true
      allDatabases: true
      resultFile: %s
    additionalArgs: []
`, "127.0.0.1", container.Port, mySQLRootPW, backupPath)) // address is hardcoded because the sql driver doesn't like 'localhost'
	}
	return []byte(fmt.Sprintf(`
mysqldump:
  options:
    flags:
      host: %s
      port: %s
      password: %s
      user: root
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
`, "127.0.0.1", container.Port, mySQLRootPW, backupPath, resticIP, resticPort))
}

// prepareTestData creates test data and inserts it into the given database
func prepareTestData(database *sql.DB) ([]TestStruct, error) {
	var err error
	testStruct1 := TestStruct{2, "TEST"}
	testData := []TestStruct{testStruct1}
	var insert *sql.Rows
	for idx := range testData {
		insert, err = database.Query(fmt.Sprintf("INSERT INTO test VALUES ( %d, '%s' )", testData[idx].ID, testData[idx].Name))
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

func mySQLDoBackup(ctx context.Context, mySQLDumpTestSuite *MySQLDumpTestSuite, useRestic bool,
	resticContainer commons.TestContainerSetup) []TestStruct {

	mySQLBackupTarget, err := commons.NewTestContainerSetup(ctx, &mySQLRequest, sqlPort)
	mySQLDumpTestSuite.Require().NoError(err)
	defer func() {
		backupErr := mySQLBackupTarget.Container.Terminate(ctx)
		mySQLDumpTestSuite.Require().NoError(backupErr)
	}()

	connectionString := fmt.Sprintf("root:%s@tcp(%s:%s)/%s?tls=skip-verify",
		mySQLRootPW, mySQLBackupTarget.Address, mySQLBackupTarget.Port, mySQLDatabase)
	db, err := sql.Open("mysql", connectionString)
	mySQLDumpTestSuite.Require().NoError(err)
	defer func() {
		dbErr := db.Close()
		mySQLDumpTestSuite.Require().NoError(dbErr)
	}()

	_, err = db.Exec("CREATE TABLE test(id INT NOT NULL AUTO_INCREMENT, name VARCHAR(100) NOT NULL, PRIMARY KEY ( id ));")
	mySQLDumpTestSuite.Require().NoError(err)

	testData, err := prepareTestData(db)
	mySQLDumpTestSuite.Require().NoError(err)

	testMySQLConfig := createMySQLConfig(mySQLBackupTarget, useRestic, resticContainer.Address, resticContainer.Port)
	err = viper.ReadConfig(bytes.NewBuffer(testMySQLConfig))
	mySQLDumpTestSuite.Require().NoError(err)

	err = source.DoBackupForKind(ctx, "mysqldump", false, useRestic, false)
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

	connectionString2 := fmt.Sprintf("root:%s@tcp(%s:%s)/%s?tls=skip-verify",
		mySQLRootPW, mySQLRestoreTarget.Address, mySQLRestoreTarget.Port, mySQLDatabase)
	dbRestore, err := sql.Open("mysql", connectionString2)
	mySQLDumpTestSuite.Require().NoError(err)
	defer func() {
		dbErr := dbRestore.Close()
		mySQLDumpTestSuite.Require().NoError(dbErr)
	}()

	// restore server from mysqldump
	err = restoreSQLFromBackup(backupPath, dbRestore)
	mySQLDumpTestSuite.Require().NoError(err)

	err = os.Remove(backupPath)
	mySQLDumpTestSuite.Require().NoError(err)

	// check if data was restored correctly
	result, err := dbRestore.Query("SELECT * FROM test")
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

	connectionString2 := fmt.Sprintf("root:%s@tcp(%s:%s)/%s?tls=skip-verify",
		mySQLRootPW, mySQLRestoreTarget.Address, mySQLRestoreTarget.Port, "mysql")
	dbRestore, err := sql.Open("mysql", connectionString2)
	mySQLDumpTestSuite.Require().NoError(err)
	defer func() {
		dbErr := dbRestore.Close()
		mySQLDumpTestSuite.Require().NoError(dbErr)
	}()

	// restore backup file from restic repository
	cmd := exec.CommandContext(ctx, "restic", "restore", "-r", fmt.Sprintf("rest:http://%s:%s/",
		resticContainer.Address, resticContainer.Port),
		"--target", dataDir, "latest")
	_, err = cmd.CombinedOutput()
	mySQLDumpTestSuite.Require().NoError(err)

	err = restoreSQLFromBackup(fmt.Sprintf("%s/%s", dataDir, backupPath), dbRestore)
	mySQLDumpTestSuite.Require().NoError(err)

	result, err := dbRestore.Query("SELECT * FROM test")
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
