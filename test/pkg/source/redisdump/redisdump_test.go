package redisdump_test

import (
	"bytes"
	"context"
	"fmt"

	"os"
	"os/exec"
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

type RedisDumpTestSuite struct {
	suite.Suite
}

var redisRequest = testcontainers.ContainerRequest{
	Image:        "redis:alpine",
	ExposedPorts: []string{redisPort},
	WaitingFor:   wait.ForLog("Ready to accept connections"),
}

func createRedisConfig(container commons.TestContainerSetup, useRestic bool, resticIP, resticPort string) []byte {
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

func createContainerFromCompose() (*testcontainers.LocalDockerCompose, error) {
	composeFilePaths := []string{"../../../testdata/testredis.yml"}
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

func (redisDumpTestSuite *RedisDumpTestSuite) TestBasicRedisDump() {
	ctx := context.Background()

	// create a redis container to test backup function
	redisBackupTarget, err := commons.NewTestContainerSetup(ctx, &redisRequest, redisPort)
	redisDumpTestSuite.Require().NoError(err)
	defer func() {
		err = redisBackupTarget.Container.Terminate(ctx)
		redisDumpTestSuite.Require().NoError(err)
	}()

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisBackupTarget.Address, redisBackupTarget.Port),
	})
	defer func() {
		err = redisClient.Close()
		redisDumpTestSuite.Require().NoError(err)
	}()

	_, err = redisClient.Ping().Result()
	redisDumpTestSuite.Require().NoError(err)

	err = redisClient.Set("name", testName, 0).Err()
	redisDumpTestSuite.Require().NoError(err)

	err = redisClient.Set("type", testType, 0).Err()
	redisDumpTestSuite.Require().NoError(err)

	testRedisConfig := createRedisConfig(redisBackupTarget, false, "", "")
	err = viper.ReadConfig(bytes.NewBuffer(testRedisConfig))
	redisDumpTestSuite.Require().NoError(err)

	// perform backup action on first redis container
	err = source.DoBackupForKind(ctx, "redisdump", false, false, false)
	redisDumpTestSuite.Require().NoError(err)

	compose, err := createContainerFromCompose()
	redisDumpTestSuite.Require().NoError(err)
	defer func() {
		err = compose.Down().Error
		redisDumpTestSuite.Require().NoError(err)
	}()

	redisRestoreClient := redis.NewClient(&redis.Options{Password: "redisdb",
		Addr: fmt.Sprintf("%s:%s", "0.0.0.0", "6379"),
	})
	defer func() {
		err = redisRestoreClient.Close()
		redisDumpTestSuite.Require().NoError(err)
	}()

	_, err = redisRestoreClient.Ping().Result()
	redisDumpTestSuite.Require().NoError(err)

	nameVal, err := redisRestoreClient.Get("name").Result()
	redisDumpTestSuite.Require().NoError(err)

	typeVal, err := redisRestoreClient.Get("type").Result()
	redisDumpTestSuite.Require().NoError(err)

	assert.Equal(redisDumpTestSuite.T(), testName, nameVal)
	assert.Equal(redisDumpTestSuite.T(), testType, typeVal)

	err = os.Remove(backupPath)
	redisDumpTestSuite.Require().NoError(err)
}

func (redisDumpTestSuite *RedisDumpTestSuite) TestRedisDumpRestic() {
	ctx := context.Background()

	// create a redis container to test backup function
	redisBackupTarget, err := commons.NewTestContainerSetup(ctx, &redisRequest, redisPort)
	redisDumpTestSuite.Require().NoError(err)
	defer func() {
		err = redisBackupTarget.Container.Terminate(ctx)
		redisDumpTestSuite.Require().NoError(err)
	}()

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisBackupTarget.Address, redisBackupTarget.Port),
	})
	_, err = redisClient.Ping().Result()
	redisDumpTestSuite.Require().NoError(err)
	defer func() {
		err = redisClient.Close()
		redisDumpTestSuite.Require().NoError(err)
	}()

	// setup a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	redisDumpTestSuite.Require().NoError(err)
	defer func() {
		err = resticContainer.Container.Terminate(ctx)
		redisDumpTestSuite.Require().NoError(err)
	}()

	err = redisClient.Set("name", testName, 0).Err()
	redisDumpTestSuite.Require().NoError(err)

	err = redisClient.Set("type", testType, 0).Err()
	redisDumpTestSuite.Require().NoError(err)

	testRedisConfig := createRedisConfig(redisBackupTarget, true, resticContainer.Address, resticContainer.Port)
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
	compose, err := createContainerFromCompose()
	redisDumpTestSuite.Require().NoError(err)
	defer func() {
		err = compose.Down().Error
		redisDumpTestSuite.Require().NoError(err)
	}()

	redisRestoreClient := redis.NewClient(&redis.Options{Password: "redisdb",
		Addr: fmt.Sprintf("%s:%s", "0.0.0.0", "6379"),
	})
	defer func() {
		err = redisRestoreClient.Close()
		redisDumpTestSuite.Require().NoError(err)
	}()

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
