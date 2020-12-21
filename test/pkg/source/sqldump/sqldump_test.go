package sqldump_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/docker/go-connections/nat"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mittwald/brudi/pkg/source"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gotest.tools/assert"
	"io/ioutil"
	"strings"
	"testing"
)

type MySQLDumpTestSuite struct {
	suite.Suite
}

type TestStruct struct {
	ID   int
	Name string
}

var mySQLRequest = testcontainers.ContainerRequest{
	Image:        "mysql:8",
	ExposedPorts: []string{"3306/tcp"},
	Env: map[string]string{
		"MYSQL_ROOT_PASSWORD": "mysqlroot",
		"MYSQL_DATABASE":      "mysql",
		"MYSQL_USER":          "mysqluser",
		"MYSQL_PASSWORD":      "mysql",
	},
	Cmd:        []string{"--default-authentication-plugin=mysql_native_password"},
	WaitingFor: wait.ForLog("port: 3306  MySQL Community Server - GPL"),
}

var resticReq = testcontainers.ContainerRequest{
	Image:        "restic/rest-server:latest",
	ExposedPorts: []string{"8000/tcp"},
	Env: map[string]string{
		"OPTIONS":         "--no-auth",
		"RESTIC_PASSWORD": "mongorepo",
	},
	VolumeMounts: map[string]string{
		"mysql-data": "/var/lib/mysql",
	},
}

type TestContainerSetup struct {
	Container testcontainers.Container
	Address   string
	Port      string
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

func newTestContainerSetup(ctx context.Context, request *testcontainers.ContainerRequest, port nat.Port) (TestContainerSetup, error) {
	result := TestContainerSetup{}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: *request,
		Started:          true,
	})
	if err != nil {
		return TestContainerSetup{}, err
	}
	result.Container = container
	contPort, err := container.MappedPort(ctx, port)
	if err != nil {
		return TestContainerSetup{}, err
	}
	result.Port = fmt.Sprint(contPort.Int())
	host, err := container.Host(ctx)
	if err != nil {
		return TestContainerSetup{}, err
	}
	result.Address = host

	return result, nil
}

// createMongoConfig creates a brudi config for the mongodump command
func createMySQLConfig(container TestContainerSetup, useRestic bool, resticIP, resticPort string) []byte {
	if !useRestic {
		return []byte(fmt.Sprintf(`
mysqldump:
  options:
    flags:
      host: %s
      port: %s
      password: mysqlroot
      user: root
      opt: true
      allDatabases: true
      resultFile: /tmp/test.sqldump
    additionalArgs: []
`, "127.0.0.1", container.Port))
	}
	return []byte(fmt.Sprintf(`
mysqldump:
  options:
    flags:
      host: %s
      port: %s
      password: mysqlroot
      user: root
      opt: true
      allDatabases: true
      resultFile: /tmp/test.sqldump
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
`, container.Address, container.Port, resticIP, resticPort))
}

func (mySQLDumpTestSuite *MySQLDumpTestSuite) TestBasicMySQLDump() {
	ctx := context.Background()

	// create a mongo container to test backup function
	mySQLBackupTarget, err := newTestContainerSetup(ctx, &mySQLRequest, "3306/tcp")
	mySQLDumpTestSuite.Require().NoError(err)

	connectionString := fmt.Sprintf("root:mysqlroot@tcp(%s:%s)/%s?tls=skip-verify",
		mySQLBackupTarget.Address, mySQLBackupTarget.Port, "mysql")
	db, err := sql.Open("mysql", connectionString)
	mySQLDumpTestSuite.Require().NoError(err)

	_, err = db.Exec("CREATE TABLE test(id INT NOT NULL AUTO_INCREMENT, name VARCHAR(100) NOT NULL, PRIMARY KEY ( id ));")
	mySQLDumpTestSuite.Require().NoError(err)

	testStruct1 := TestStruct{2, "TEST"}
	testData := []TestStruct{testStruct1}
	var insert *sql.Rows
	for idx := range testData {
		insert, err = db.Query(fmt.Sprintf("INSERT INTO test VALUES ( %d, '%s' )", testData[idx].ID, testData[idx].Name))
		mySQLDumpTestSuite.Require().NoError(insert.Err())
		mySQLDumpTestSuite.Require().NoError(err)
	}
	err = insert.Close()
	mySQLDumpTestSuite.Require().NoError(err)

	err = db.Close()
	mySQLDumpTestSuite.Require().NoError(err)

	testMySQLConfig := createMySQLConfig(mySQLBackupTarget, false, "", "")
	err = viper.ReadConfig(bytes.NewBuffer(testMySQLConfig))
	mySQLDumpTestSuite.Require().NoError(err)

	// perform backup action on first mongo container
	err = source.DoBackupForKind(ctx, "mysqldump", false, false, false)
	mySQLDumpTestSuite.Require().NoError(err)

	mySQLBackupTarget.Container.Terminate(ctx)

	mySQLRestoreTarget, err := newTestContainerSetup(ctx, &mySQLRequest, "3306/tcp")
	mySQLDumpTestSuite.Require().NoError(err)

	_, err = mySQLRestoreTarget.Container.Exec(context.TODO(), []string{"mysql", "--user=root",
		"--database=mysql", "--password=mysqlroot", "<", "/tmp/test.sqldump"})
	mySQLDumpTestSuite.Require().NoError(err)

	connectionString2 := fmt.Sprintf("root:mysqlroot@tcp(%s:%s)/%s?tls=skip-verify",
		mySQLRestoreTarget.Address, mySQLRestoreTarget.Port, "mysql")
	dbRestore, err := sql.Open("mysql", connectionString2)
	mySQLDumpTestSuite.Require().NoError(err)
	file, err := ioutil.ReadFile("/tmp/test.sqldump")
	mySQLDumpTestSuite.Require().NoError(err)

	requests := strings.Split(string(file), ";\n")

	for _, request := range requests {
		_, err := dbRestore.Exec(request)
		mySQLDumpTestSuite.Require().NoError(err)
	}

	result, err := dbRestore.Query("SELECT * FROM test")
	mySQLDumpTestSuite.Require().NoError(err)
	mySQLDumpTestSuite.Require().NoError(result.Err())
	defer result.Close()

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
