package redisdump_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/go-redis/redis"
	"github.com/mittwald/brudi/pkg/source"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gotest.tools/assert"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const redisPort = "6379"
const testName = "test"
const testType = "gopher"
const backupPath = "/tmp/redisdump.rdb"
const implemented = false

type RedisDumpTestSuite struct {
	suite.Suite
}

type TestContainerSetup struct {
	Container testcontainers.Container
	Address   string
	Port      string
}

var resticReq = testcontainers.ContainerRequest{
	Image:        "restic/rest-server:latest",
	ExposedPorts: []string{"8000/tcp"},
	Env: map[string]string{
		"OPTIONS":         "--no-auth",
		"RESTIC_PASSWORD": "mongorepo",
	},
}

var redisRequest = testcontainers.ContainerRequest{
	Image:        "redis:alpine",
	ExposedPorts: []string{fmt.Sprintf("%s/tcp", redisPort)},
	WaitingFor:   wait.ForLog("Ready to accept connections"),
}

var redisRestoreRequest = testcontainers.ContainerRequest{
	Image:        "redis:alpine",
	ExposedPorts: []string{fmt.Sprintf("%s/tcp", redisPort)},
	//BindMounts: map[string]string{
	//   backupPath:"/data/dump.rdb" ,
	//},
	WaitingFor: wait.ForLog("Ready to accept connections"),
}

var redisRestoreRequestRestic = testcontainers.ContainerRequest{
	Image:        "redis:alpine",
	ExposedPorts: []string{fmt.Sprintf("%s/tcp", redisPort)},
	WaitingFor:   wait.ForLog("Ready to accept connections"),
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

func createRedisConfig(container TestContainerSetup, useRestic bool, resticIP, resticPort string) []byte {
	if !useRestic {
		return []byte(fmt.Sprintf(`
redisdump:
  options:
    flags:
      host: %s
      port: %s
      password: redisdb
      rdb: %s
    additionalArgs: []
`, container.Address, container.Port, backupPath))
	}
	return []byte(fmt.Sprintf(`
redisdump:
  options:
    flags:
      host: %s
      port: %s
      password: redisdb
      rdb: %s
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
`, container.Address, container.Port, backupPath, resticIP, resticPort))
}

func (redisDumpTestSuite *RedisDumpTestSuite) SetupTest() {
	viper.Reset()
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func (redisDumpTestSuite *RedisDumpTestSuite) TearDownTest() {
	viper.Reset()
}

func (redisDumpTestSuite *RedisDumpTestSuite) TestBasicRedisDump() {
	ctx := context.Background()
	port, err := nat.NewPort("tcp", redisPort)
	redisDumpTestSuite.Require().NoError(err)

	// create a redis container to test backup function
	redisBackupTarget, err := newTestContainerSetup(ctx, &redisRequest, port)
	redisDumpTestSuite.Require().NoError(err)
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisBackupTarget.Address, redisBackupTarget.Port),
	})
	_, err = redisClient.Ping().Result()
	redisDumpTestSuite.Require().NoError(err)

	err = redisClient.Set("name", testName, 0).Err()
	redisDumpTestSuite.Require().NoError(err)

	err = redisClient.Set("type", testType, 0).Err()
	redisDumpTestSuite.Require().NoError(err)

	err = redisClient.Close()
	redisDumpTestSuite.Require().NoError(err)

	testRedisConfig := createRedisConfig(redisBackupTarget, false, "", "")
	err = viper.ReadConfig(bytes.NewBuffer(testRedisConfig))
	redisDumpTestSuite.Require().NoError(err)

	// perform backup action on first redis container
	err = source.DoBackupForKind(ctx, "redisdump", false, false, false)
	redisDumpTestSuite.Require().NoError(err)

	redisBackupTarget.Container.Terminate(ctx)

	// Checking of backed up data is currently unavailable due to implementation issues

	// create second redis container to test dumped values. link dump.rdb as volume
	//redisRestoreTarget, err := newTestContainerSetup(ctx, &redisRestoreRequest, port)
	//redisDumpTestSuite.Require().NoError(err)
	//redisRestoreClient := redis.NewClient(&redis.Options{
	//	Addr: fmt.Sprintf("%s:%s", redisRestoreTarget.Address, redisRestoreTarget.Port),
	//})
	//
	//_, err = redisRestoreClient.Ping().Result()
	//redisDumpTestSuite.Require().NoError(err)
	//
	//nameVal, err := redisRestoreClient.Get("name").Result()
	//redisDumpTestSuite.Require().NoError(err)
	//
	//typeVal, err := redisRestoreClient.Get("type").Result()
	//redisDumpTestSuite.Require().NoError(err)
	//
	//assert.Equal(redisDumpTestSuite.T(), testName, nameVal)
	//assert.Equal(redisDumpTestSuite.T(), testType, typeVal)
	//
	//err = redisRestoreClient.Close()
	//redisDumpTestSuite.Require().NoError(err)
	//
	//err = redisRestoreTarget.Container.Terminate(ctx)
	//redisDumpTestSuite.Require().NoError(err)
	err = os.Remove(backupPath)
	redisDumpTestSuite.Require().NoError(err)
}

func (redisDumpTestSuite *RedisDumpTestSuite) TestRedisDumpRestic() {
	if !implemented {
		return
	}

	ctx := context.Background()
	port, err := nat.NewPort("tcp", redisPort)
	redisDumpTestSuite.Require().NoError(err)

	// create a redis container to test backup function
	redisBackupTarget, err := newTestContainerSetup(ctx, &redisRequest, port)
	redisDumpTestSuite.Require().NoError(err)
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisBackupTarget.Address, redisBackupTarget.Port),
	})
	_, err = redisClient.Ping().Result()
	redisDumpTestSuite.Require().NoError(err)

	// setup a container running the restic rest-server
	resticContainer, err := newTestContainerSetup(ctx, &resticReq, "8000/tcp")
	redisDumpTestSuite.Require().NoError(err)

	err = redisClient.Set("name", testName, 0).Err()
	redisDumpTestSuite.Require().NoError(err)

	err = redisClient.Set("type", testType, 0).Err()
	redisDumpTestSuite.Require().NoError(err)

	err = redisClient.Close()
	redisDumpTestSuite.Require().NoError(err)

	testRedisConfig := createRedisConfig(redisBackupTarget, false, resticContainer.Address, resticContainer.Port)
	err = viper.ReadConfig(bytes.NewBuffer(testRedisConfig))
	redisDumpTestSuite.Require().NoError(err)

	// perform backup action on first redis container
	err = source.DoBackupForKind(ctx, "redisdump", false, true, false)
	redisDumpTestSuite.Require().NoError(err)

	cmd := exec.CommandContext(ctx, "restic", "restore", "-r", fmt.Sprintf("rest:http://%s:%s/",
		resticContainer.Address, resticContainer.Port),
		"--target", "data", "latest")
	_, err = cmd.CombinedOutput()
	redisDumpTestSuite.Require().NoError(err)

	// create second redis container to test dumped values. link dump.rdb as volume
	redisRestoreTarget, err := newTestContainerSetup(ctx, &redisRestoreRequestRestic, port)
	redisDumpTestSuite.Require().NoError(err)
	redisRestoreClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisRestoreTarget.Address, redisRestoreTarget.Port),
	})

	_, err = redisRestoreClient.Ping().Result()
	redisDumpTestSuite.Require().NoError(err)

	nameVal, err := redisRestoreClient.Get("name").Result()
	redisDumpTestSuite.Require().NoError(err)

	typeVal, err := redisRestoreClient.Get("type").Result()
	redisDumpTestSuite.Require().NoError(err)

	assert.Equal(redisDumpTestSuite.T(), testName, nameVal)
	assert.Equal(redisDumpTestSuite.T(), testType, typeVal)

}

func TestRedisDumpTestSuite(t *testing.T) {
	suite.Run(t, new(RedisDumpTestSuite))
}
