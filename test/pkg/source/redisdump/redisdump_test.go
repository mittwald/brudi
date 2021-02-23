package redisdump_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mittwald/brudi/pkg/cli"
	"os"
	"strings"
	"testing"

	"github.com/mittwald/brudi/pkg/source"
	commons "github.com/mittwald/brudi/test/pkg/source/internal"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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
const backupPathZip = "/tmp/redisdump.rdb.gz"
const redisPW = "redisdb"
const nameKey = "name"
const typeKey = "type"
const logString = "Ready to accept connections"
const dumpKind = "redisdump"
const redisImage = "quay.io/bitnami/redis:latest"

type RedisDumpTestSuite struct {
	suite.Suite
}

func (redisDumpTestSuite *RedisDumpTestSuite) SetupTest() {
	commons.TestSetup()
}

// TearDownTest resets viper after test
func (redisDumpTestSuite *RedisDumpTestSuite) TearDownTest() {
	viper.Reset()
}

// TestBasicRedisDump performs an integration test for brudi's `redisdump` command without restic
func (redisDumpTestSuite *RedisDumpTestSuite) TestBasicRedisDump() {
	ctx := context.Background()

	// remove backup files after test
	defer func() {
		removeErr := os.RemoveAll(backupPath)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to remove redis backup files")
		}
	}()

	// create a redis container to test backup function
	testData, err := redisDoBackup(ctx, false, commons.TestContainerSetup{Port: "", Address: ""}, backupPath)
	redisDumpTestSuite.Require().NoError(err)

	var restoredData testStruct
	restoredData, err = redisDoRestore(ctx, false, commons.TestContainerSetup{Port: "", Address: ""}, backupPath)
	redisDumpTestSuite.Require().NoError(err)

	assert.Equal(redisDumpTestSuite.T(), testData.Name, restoredData.Name)
	assert.Equal(redisDumpTestSuite.T(), testData.Type, restoredData.Type)
}

// TestBasicRedisDumpGzip performs an integration test for brudi's `redisdump` command with gzip and without restic
func (redisDumpTestSuite *RedisDumpTestSuite) TestBasicRedisDumpGzip() {
	ctx := context.Background()

	// remove backup files after test
	defer func() {
		removeErr := os.RemoveAll(backupPathZip)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to remove redis backup files")
		}
	}()

	// create a redis container to test backup function
	testData, err := redisDoBackup(ctx, false, commons.TestContainerSetup{Port: "", Address: ""}, backupPathZip)
	redisDumpTestSuite.Require().NoError(err)

	var restoredData testStruct
	restoredData, err = redisDoRestore(ctx, false, commons.TestContainerSetup{Port: "", Address: ""}, backupPathZip)
	redisDumpTestSuite.Require().NoError(err)

	assert.Equal(redisDumpTestSuite.T(), testData.Name, restoredData.Name)
	assert.Equal(redisDumpTestSuite.T(), testData.Type, restoredData.Type)
}

// TestBasicRedisDumpRestic performs an integration test for brudi's `redisdump` command with restic
func (redisDumpTestSuite *RedisDumpTestSuite) TestRedisDumpRestic() {
	ctx := context.Background()

	// remove backup files after test
	defer func() {
		removeErr := os.RemoveAll(backupPath)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to remove redis backup files")
		}
	}()

	// setup a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	redisDumpTestSuite.Require().NoError(err)
	defer func() {
		resticErr := resticContainer.Container.Terminate(ctx)
		if resticErr != nil {
			log.WithError(resticErr).Error("failed to terminate redis restic container")
		}
	}()

	// populate a redis database with test data, back it up with brudi and retain the test data it for verification
	var testData testStruct
	testData, err = redisDoBackup(ctx, true, resticContainer, backupPath)
	redisDumpTestSuite.Require().NoError(err)

	// restore a redis database from backup and pull test data for verification
	var restoredData testStruct
	restoredData, err = redisDoRestore(ctx, true, resticContainer, backupPath)
	redisDumpTestSuite.Require().NoError(err)

	assert.Equal(redisDumpTestSuite.T(), testData.Name, restoredData.Name)
	assert.Equal(redisDumpTestSuite.T(), testData.Type, restoredData.Type)
}

// TestBasicRedisDumpRestic performs an integration test for brudi's `redisdump` command with restic and gzip
func (redisDumpTestSuite *RedisDumpTestSuite) TestRedisDumpResticGzip() {
	ctx := context.Background()

	// remove backup files after test
	defer func() {
		removeErr := os.RemoveAll(backupPathZip)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to remove redis backup files")
		}
	}()

	// setup a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	redisDumpTestSuite.Require().NoError(err)
	defer func() {
		resticErr := resticContainer.Container.Terminate(ctx)
		if resticErr != nil {
			log.WithError(resticErr).Error("failed to terminate redis restic container")
		}
	}()

	// populate a redis database with test data, back it up with brudi and retain the test data it for verification
	var testData testStruct
	testData, err = redisDoBackup(ctx, true, resticContainer, backupPathZip)
	redisDumpTestSuite.Require().NoError(err)

	// restore a redis database from backup and pull test data for verification
	var restoredData testStruct
	restoredData, err = redisDoRestore(ctx, true, resticContainer, backupPathZip)
	redisDumpTestSuite.Require().NoError(err)

	assert.Equal(redisDumpTestSuite.T(), testData.Name, restoredData.Name)
	assert.Equal(redisDumpTestSuite.T(), testData.Type, restoredData.Type)
}

func TestRedisDumpTestSuite(t *testing.T) {
	suite.Run(t, new(RedisDumpTestSuite))
}

// redisDoBackup populates a database with test data and performs a backup
func redisDoBackup(ctx context.Context, useRestic bool,
	resticContainer commons.TestContainerSetup, path string) (testStruct, error) {
	// setup a redis container to backup from
	redisBackupTarget, err := commons.NewTestContainerSetup(ctx, &redisRequest, redisPort)
	if err != nil {
		return testStruct{}, errors.WithStack(err)
	}
	defer func() {
		backupErr := redisBackupTarget.Container.Terminate(ctx)
		if backupErr != nil {
			log.WithError(backupErr).Error("failed to terminate redis backup container")
		}
	}()

	// connect to database to prepare for test data insertion
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisBackupTarget.Address, redisBackupTarget.Port),
	})
	defer func() {
		redisErr := redisClient.Close()
		if redisErr != nil {
			log.WithError(redisErr).Error("failed to close redis backup client")
		}
	}()

	// test connection
	_, err = redisClient.Ping().Result()
	if err != nil {
		return testStruct{}, errors.WithStack(err)
	}

	//  setup test data and write it to database
	testData := testStruct{Name: testName, Type: testType}
	err = redisClient.Set(nameKey, testData.Name, 0).Err()
	if err != nil {
		return testStruct{}, errors.WithStack(err)
	}

	err = redisClient.Set(typeKey, testData.Type, 0).Err()
	if err != nil {
		return testStruct{}, errors.WithStack(err)
	}

	// create a brudi config for redisdump
	redisBackupConfig := createRedisConfig(redisBackupTarget, useRestic, resticContainer.Address, resticContainer.Port, path)
	err = viper.ReadConfig(bytes.NewBuffer(redisBackupConfig))
	if err != nil {
		return testStruct{}, errors.WithStack(err)
	}

	// perform backup action on first redis container
	err = source.DoBackupForKind(ctx, dumpKind, false, useRestic, true)
	if err != nil {
		return testStruct{}, errors.WithStack(err)
	}

	return testData, nil
}

// redisDoRestore restores data from backup and retrieves it for verification, optionally using restic
func redisDoRestore(ctx context.Context, useRestic bool,
	resticContainer commons.TestContainerSetup, path string) (testStruct, error) {
	// unzip file if necessary
	_, err := cli.CheckAndGunzipFile(path)
	if err != nil {
		return testStruct{}, err
	}

	// setup container to restore data to
	var redisRestoreTarget commons.TestContainerSetup
	redisRestoreTarget, err = commons.NewTestContainerSetup(ctx, &redisRestoreRequest, redisPort)
	if err != nil {
		return testStruct{}, errors.WithStack(err)
	}
	defer func() {
		restoreErr := redisRestoreTarget.Container.Terminate(ctx)
		if restoreErr != nil {
			log.WithError(restoreErr).Error("failed to terminate redis restore container")
		}
	}()

	// pull data from restic repository if needed
	if useRestic {
		err = commons.DoResticRestore(ctx, resticContainer, backupPath)
		if err != nil {
			return testStruct{}, errors.WithStack(err)
		}
	}

	// connect to database to prepare for restoration
	redisRestoreClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisRestoreTarget.Address, redisRestoreTarget.Port),
	})
	defer func() {
		redisErr := redisRestoreClient.Close()
		if redisErr != nil {
			log.WithError(redisErr).Error("failed to close redis restore client")
		}
	}()

	// check connection
	_, err = redisRestoreClient.Ping().Result()
	if err != nil {
		return testStruct{}, errors.WithStack(err)
	}

	// retrieve first test value
	var nameVal string
	nameVal, err = redisRestoreClient.Get(nameKey).Result()
	if err != nil {
		return testStruct{}, errors.WithStack(err)
	}

	// retrive second test value
	var typeVal string
	typeVal, err = redisRestoreClient.Get(typeKey).Result()
	if err != nil {
		return testStruct{}, errors.WithStack(err)
	}

	return testStruct{Name: nameVal, Type: typeVal}, err
}

// redisRequest is a request for a blank redis container
var redisRequest = testcontainers.ContainerRequest{
	Image:        redisImage,
	ExposedPorts: []string{redisPort},
	WaitingFor:   wait.ForLog(logString),
	Env: map[string]string{"ALLOW_EMPTY_PASSWORD": "yes",
		"REDIS_AOF_ENABLED": "no"},
}

// redisRestoreRequest is a request for a redis container that mounts an rdb-file from backupPath to initialize the database
var redisRestoreRequest = testcontainers.ContainerRequest{
	Image:        redisImage,
	ExposedPorts: []string{redisPort},
	WaitingFor:   wait.ForLog(logString),
	BindMounts:   map[string]string{strings.TrimSuffix(backupPath, ".gz"): "/bitnami/redis/data/dump.rdb"},
	Env: map[string]string{"ALLOW_EMPTY_PASSWORD": "yes",
		"REDIS_AOF_ENABLED": "no"},
}

// createRedisConfig returns a brudi config for redis
func createRedisConfig(container commons.TestContainerSetup, useRestic bool, resticIP, resticPort, path string) []byte {
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
`, container.Address, container.Port, redisPW, path))
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
      keepDaily: 7
      keepHourly: 24
      keepLast: 48
      keepMonthly: 6
      keepWeekly: 2
      keepYearly: 2
    ids: []
  restore:
    flags:
      target: "/"
    id: "latest"
`, container.Address, container.Port, redisPW, path, resticIP, resticPort))
}

type testStruct struct {
	Name string
	Type string
}
