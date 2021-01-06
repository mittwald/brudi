package redisdump_test

import (
	"bytes"
	"context"
	"fmt"

	"os"
	"strings"
	"testing"

	"github.com/mittwald/brudi/pkg/source"
	commons "github.com/mittwald/brudi/test/pkg/source/internal"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gotest.tools/assert"
)

const redisPort = "6379/tcp"
const testName = "test"
const testType = "gopher"
const backupPath = "/tmp/redisdump.rdb"
const backupPathRestic = "/tmp/redisdump_restic.rdb"
const redisPW = "redisdb"
const composeLocation = "../../../testdata/testredis.yml"
const dataDir = "data"
const nameKey = "name"
const typeKey = "type"
const logString = "Ready to accept connections"
const dumpKind = "redisdump"
const redisImage = "quay.io/bitnami/redis:latest"

type RedisDumpTestSuite struct {
	suite.Suite
}

// redisRequest is a request for a blank redis container
var redisRequest = testcontainers.ContainerRequest{
	Image:        redisImage,
	ExposedPorts: []string{redisPort},
	WaitingFor:   wait.ForLog("Ready to accept connections"),
	Env: map[string]string{"ALLOW_EMPTY_PASSWORD": "yes",
		"REDIS_AOF_ENABLED": "no"},
}

// redisRestoreRequest is a request for a redis container that mounts an rdb-file from backupPath to initialize the database
var redisRestoreRequest = testcontainers.ContainerRequest{
	Image:        redisImage,
	ExposedPorts: []string{redisPort},
	WaitingFor:   wait.ForLog(logString),
	BindMounts:   map[string]string{backupPath: "/bitnami/redis/data/dump.rdb"},
	Env: map[string]string{"ALLOW_EMPTY_PASSWORD": "yes",
		"REDIS_AOF_ENABLED": "no"},
}

func createResticRestoreRequest(backupFilePath string) *testcontainers.ContainerRequest {
	request := testcontainers.ContainerRequest{
		Image:        redisImage,
		ExposedPorts: []string{redisPort},
		WaitingFor:   wait.ForLog(logString),
		BindMounts:   map[string]string{backupFilePath: "/bitnami/redis/data/dump.rdb"},
		Env: map[string]string{"ALLOW_EMPTY_PASSWORD": "yes",
			"REDIS_AOF_ENABLED": "no"},
	}
	return &request
}

// createRedisConfig returns a brudi config for redis
func createRedisConfig(container commons.TestContainerSetup, useRestic bool, resticIP, resticPort string) []byte {
	if !useRestic {
		return []byte(fmt.Sprintf(`
redisdump:
  options:
    flags:
      host: %s
      port: %s
      password: %s
      rdb: %s
    additionalArgs: []
`, container.Address, container.Port, redisPW, backupPath))
	}
	return []byte(fmt.Sprintf(`
redisdump:
  options:
    flags:
      host: %s
      port: %s
      password: %s
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
`, container.Address, container.Port, redisPW, backupPathRestic, resticIP, resticPort))
}

func (redisDumpTestSuite *RedisDumpTestSuite) SetupTest() {
	commons.TestSetup()
}

// TearDownTest resets viper after test
func (redisDumpTestSuite *RedisDumpTestSuite) TearDownTest() {
	viper.Reset()
}

func createContainerFromCompose() (*testcontainers.LocalDockerCompose, error) {
	composeFilePaths := []string{composeLocation}
	identifier := strings.ToLower(uuid.New().String())

	compose := testcontainers.NewLocalDockerCompose(composeFilePaths, identifier)
	execError := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	err := execError.Error
	if err != nil {
		return &testcontainers.LocalDockerCompose{}, err
	}
	return compose, nil
}

// doResticBackup  populates a database with test data and performs a backup
func doRedisBackup(ctx context.Context, redisDumpTestSuite *RedisDumpTestSuite, useRestic bool,
	resticContainer commons.TestContainerSetup) {
	redisBackupTarget, err := commons.NewTestContainerSetup(ctx, &redisRequest, redisPort)
	redisDumpTestSuite.Require().NoError(err)
	defer func() {
		backupErr := redisBackupTarget.Container.Terminate(ctx)
		redisDumpTestSuite.Require().NoError(backupErr)
	}()

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisBackupTarget.Address, redisBackupTarget.Port),
	})
	defer func() {
		redisErr := redisClient.Close()
		redisDumpTestSuite.Require().NoError(redisErr)
	}()

	_, err = redisClient.Ping().Result()
	redisDumpTestSuite.Require().NoError(err)

	err = redisClient.Set(nameKey, testName, 0).Err()
	redisDumpTestSuite.Require().NoError(err)

	err = redisClient.Set(typeKey, testType, 0).Err()
	redisDumpTestSuite.Require().NoError(err)

	redisBackupConfig := createRedisConfig(redisBackupTarget, useRestic, resticContainer.Address, resticContainer.Port)
	err = viper.ReadConfig(bytes.NewBuffer(redisBackupConfig))
	redisDumpTestSuite.Require().NoError(err)

	// perform backup action on first redis container
	err = source.DoBackupForKind(ctx, dumpKind, false, useRestic, false)
	redisDumpTestSuite.Require().NoError(err)
}

// TestBasicRedisDump performs an integration test for brudi's `redisdump` command without restic
func (redisDumpTestSuite *RedisDumpTestSuite) TestBasicRedisDump() {
	ctx := context.Background()

	// create a redis container to test backup function
	doRedisBackup(ctx, redisDumpTestSuite, false, commons.TestContainerSetup{Port: "", Address: ""})

	redisRestoreTarget, err := commons.NewTestContainerSetup(ctx, createResticRestoreRequest(backupPath), redisPort)
	redisDumpTestSuite.Require().NoError(err)
	defer func() {
		restoreErr := redisRestoreTarget.Container.Terminate(ctx)
		redisDumpTestSuite.Require().NoError(restoreErr)
	}()

	redisRestoreClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisRestoreTarget.Address, redisRestoreTarget.Port),
	})
	defer func() {
		redisErr := redisRestoreClient.Close()
		redisDumpTestSuite.Require().NoError(redisErr)
	}()

	_, err = redisRestoreClient.Ping().Result()
	redisDumpTestSuite.Require().NoError(err)

	nameVal, err := redisRestoreClient.Get(nameKey).Result()
	redisDumpTestSuite.Require().NoError(err)

	typeVal, err := redisRestoreClient.Get(typeKey).Result()
	redisDumpTestSuite.Require().NoError(err)

	assert.Equal(redisDumpTestSuite.T(), testName, nameVal)
	assert.Equal(redisDumpTestSuite.T(), testType, typeVal)
}

// TestBasicRedisDumpRestic performs an integration test for brudi's `redisdump` command with restic
func (redisDumpTestSuite *RedisDumpTestSuite) TestRedisDumpRestic() {
	ctx := context.Background()

	// setup a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	redisDumpTestSuite.Require().NoError(err)
	defer func() {
		resticErr := resticContainer.Container.Terminate(ctx)
		redisDumpTestSuite.Require().NoError(resticErr)
	}()

	// setup a redis container, populate it with test data and perform a backup
	doRedisBackup(ctx, redisDumpTestSuite, true, resticContainer)

	// pull backup from restic repo
	err = commons.DoResticRestore(ctx, resticContainer, dataDir)
	redisDumpTestSuite.Require().NoError(err)
	defer func() {
		removeErr := os.RemoveAll(dataDir)
		redisDumpTestSuite.Require().NoError(removeErr)
	}()

	// setup a new redis-container which loads the backup as a volume
	redisRestoreTarget, err := commons.NewTestContainerSetup(ctx, createResticRestoreRequest(backupPathRestic), redisPort)
	redisDumpTestSuite.Require().NoError(err)
	defer func() {
		restoreErr := redisRestoreTarget.Container.Terminate(ctx)
		redisDumpTestSuite.Require().NoError(restoreErr)
	}()

	// redis-client to retrieve restored data from database
	redisRestoreClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisRestoreTarget.Address, redisRestoreTarget.Port),
	})
	defer func() {
		redisErr := redisRestoreClient.Close()
		redisDumpTestSuite.Require().NoError(redisErr)
	}()

	_, err = redisRestoreClient.Ping().Result()
	redisDumpTestSuite.Require().NoError(err)

	nameVal, err := redisRestoreClient.Get(nameKey).Result()
	redisDumpTestSuite.Require().NoError(err)

	typeVal, err := redisRestoreClient.Get(typeKey).Result()
	redisDumpTestSuite.Require().NoError(err)

	assert.Equal(redisDumpTestSuite.T(), testName, nameVal)
	assert.Equal(redisDumpTestSuite.T(), testType, typeVal)
}

func TestRedisDumpTestSuite(t *testing.T) {
	suite.Run(t, new(RedisDumpTestSuite))
}
