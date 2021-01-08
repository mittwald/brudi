package mongodump_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mittwald/brudi/pkg/source"
	commons "github.com/mittwald/brudi/test/pkg/source/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gotest.tools/assert"
	"os"
	"testing"
)

const mongoPort = "27017/tcp"
const backupPath = "/tmp/dump.tar.gz"
const mongoPW = "mongodbroot"
const mongoUser = "root"
const dumpKind = "mongodump"
const restoreKind = "mongorestore"
const dbName = "test"
const collName = "testColl"
const mongoImage = "quay.io/bitnami/mongodb:latest"
const logString = "Waiting for connections"

type MongoDumpTestSuite struct {
	suite.Suite
}

func (mongoDumpTestSuite *MongoDumpTestSuite) SetupTest() {
	commons.TestSetup()
}

// TearDownTest resets viper after a test
func (mongoDumpTestSuite *MongoDumpTestSuite) TearDownTest() {
	viper.Reset()
}

// TestBasicMongoDBDump performs an integration test for the `mongodump` command
func (mongoDumpTestSuite *MongoDumpTestSuite) TestBasicMongoDBDump() {
	ctx := context.Background()

	// remove files after test is done
	defer func() {
		removeErr := os.Remove(backupPath)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to clean up mongodb backup files")
		}
	}()

	// backup test data with brudi and retain test data for verification
	testData, err := mongoDoBackup(ctx, false, commons.TestContainerSetup{Port: "", Address: ""})
	mongoDumpTestSuite.Require().NoError(err)

	// restore database from backup and pull test data from it for verification
	var results []interface{}
	results, err = mongoDoRestore(ctx, false, commons.TestContainerSetup{Port: "", Address: ""})
	mongoDumpTestSuite.Require().NoError(err)

	// check if the original data was restored
	assert.DeepEqual(mongoDumpTestSuite.T(), testData, results)
}

// TestBasicMongoDBDumpRestic performs an integration test for the `mongodump` command with restic support
func (mongoDumpTestSuite *MongoDumpTestSuite) TestBasicMongoDBDumpRestic() {
	ctx := context.Background()

	// remove files after test is done
	defer func() {
		removeErr := os.Remove(backupPath)
		if removeErr != nil {
			log.WithError(removeErr).Error("failed to clean up mongodb backup files")
		}
	}()
	// create a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	mongoDumpTestSuite.Require().NoError(err)
	defer func() {
		resticErr := resticContainer.Container.Terminate(ctx)
		if resticErr != nil {
			log.WithError(resticErr).Error("failed to terminate mongodb restic container")
		}
	}()

	// backup database and retain test data for verification
	var testData []interface{}
	testData, err = mongoDoBackup(ctx, true, resticContainer)
	mongoDumpTestSuite.Require().NoError(err)

	// restore database from backup and pull test data for verification
	var results []interface{}
	results, err = mongoDoRestore(ctx, true, resticContainer)

	assert.DeepEqual(mongoDumpTestSuite.T(), testData, results)
}

func TestMongoDumpTestSuite(t *testing.T) {
	suite.Run(t, new(MongoDumpTestSuite))
}

// mongoDoBackup performs a mongodump and returns the test data that was used for verification purposes
func mongoDoBackup(ctx context.Context, useRestic bool,
	resticContainer commons.TestContainerSetup) ([]interface{}, error) {
	// create a mongodb-container to test backup function
	mongoBackupTarget, err := commons.NewTestContainerSetup(ctx, &mongoRequest, mongoPort)
	if err != nil {
		return []interface{}{}, err
	}
	defer func() {
		backErr := mongoBackupTarget.Container.Terminate(ctx)
		if backErr != nil {
			log.WithError(backErr).Error("failed to terminate mongodb backup container")
		}
	}()

	// client to insert test data into database
	var backupClient mongo.Client
	backupClient, err = newMongoClient(&mongoBackupTarget)
	if err != nil {
		return []interface{}{}, err
	}
	defer func() {
		clientErr := backupClient.Disconnect(ctx)
		if clientErr != nil {
			log.WithError(clientErr).Error("failed to disconnect mongodb backup-client")
		}
	}()

	// write test data into database and retain it for later assertion
	var testData []interface{}
	testData, err = prepareTestData(&backupClient)
	if err != nil {
		return []interface{}{}, err
	}

	// create brudi config for backup
	backupMongoConfig := createMongoConfig(mongoBackupTarget, useRestic, resticContainer.Address, resticContainer.Port, dumpKind)
	err = viper.ReadConfig(bytes.NewBuffer(backupMongoConfig))
	if err != nil {
		return []interface{}{}, err
	}

	// perform backup action on mongodb-container
	err = source.DoBackupForKind(ctx, dumpKind, false, useRestic, false)
	if err != nil {
		return []interface{}{}, err
	}

	return testData, nil
}

// mongoDoRestore restores data from backup using brudi and retrieves it for verification, optionally using restic
func mongoDoRestore(ctx context.Context, useRestic bool,
	resticContainer commons.TestContainerSetup) ([]interface{}, error) {
	// setup a new mongodb container which will be used to ensure data was backed up correctly
	mongoRestoreTarget, err := commons.NewTestContainerSetup(ctx, &mongoRequest, mongoPort)
	if err != nil {
		return []interface{}{}, err
	}
	defer func() {
		restoreErr := mongoRestoreTarget.Container.Terminate(ctx)
		if restoreErr != nil {
			log.WithError(restoreErr).Error("failed to terminate mongodb restore container")
		}
	}()

	// create brudi config for restoration
	restoreMongoConfig := createMongoConfig(mongoRestoreTarget, useRestic, resticContainer.Address, resticContainer.Port, restoreKind)
	err = viper.ReadConfig(bytes.NewBuffer(restoreMongoConfig))
	if err != nil {
		return []interface{}{}, err
	}

	// use `mongorestore` to restore backed up data to new container
	err = source.DoRestoreForKind(ctx, restoreKind, false, useRestic, false)
	if err != nil {
		return []interface{}{}, err
	}

	// client to retrieve restored data
	var restoreClient mongo.Client
	restoreClient, err = newMongoClient(&mongoRestoreTarget)
	if err != nil {
		return []interface{}{}, err
	}
	defer func() {
		clientErr := restoreClient.Disconnect(ctx)
		if clientErr != nil {
			log.WithError(clientErr).Error("failed to disconnect mongodb restore-client")
		}
	}()

	// pull restored data from database
	restoredCollection := restoreClient.Database(dbName).Collection(collName)
	findOptions := options.Find()
	var cur *mongo.Cursor
	cur, err = restoredCollection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		return []interface{}{}, err
	}
	defer func() {
		curErr := cur.Close(context.TODO())
		if curErr != nil {
			log.WithError(curErr).Error("failed to close mongosb cursor")
		}
	}()

	// get results from cursor
	var results []interface{}
	results, err = getResultsFromCursor(cur)
	if err != nil {
		return []interface{}{}, err
	}

	return results, nil
}

// mongorequest is a testcontainers.ContainerRequest for a basic mongodb testcontainer
var mongoRequest = testcontainers.ContainerRequest{
	Image:        mongoImage,
	ExposedPorts: []string{mongoPort},
	Env: map[string]string{
		"MONGODB_ROOT_USERNAME": mongoUser,
		"MONGODB_ROOT_PASSWORD": mongoPW,
	},
	WaitingFor: wait.ForLog(logString),
}

// newMongoClient creates a mongo client connected to the provided commons.TestContainerSetup
func newMongoClient(target *commons.TestContainerSetup) (mongo.Client, error) {
	backupClientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", target.Address,
		target.Port))
	clientAuth := options.Client().SetAuth(options.Credential{Username: mongoUser, Password: mongoPW})
	client, err := mongo.Connect(context.TODO(), backupClientOptions, clientAuth)
	if err != nil {

		return mongo.Client{}, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return mongo.Client{}, err
	}
	return *client, nil
}

// createMongoConfig creates a brudi config for the brudi command specified via kind
func createMongoConfig(container commons.TestContainerSetup, useRestic bool, resticIP, resticPort, kind string) []byte {
	if !useRestic {
		return []byte(fmt.Sprintf(`
      %s:
        options:
          flags:
            host: %s
            port: %s
            username: %s
            password: %s
            gzip: true
            archive: %s
          additionalArgs: []
`, kind, container.Address, container.Port, mongoUser, mongoPW, backupPath))
	}
	return []byte(fmt.Sprintf(`
      %s:
        options:
          flags:
            host: %s
            port: %s
            username: %s
            password: %s
            gzip: true
            archive: %s
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
        restore:
          flags:
            target: "/"
          id: "latest"
`, kind, container.Address, container.Port, mongoUser, mongoPW, backupPath, resticIP, resticPort))
}

// prepareTestData creates test data and writes it into a database using the provided client
func prepareTestData(client *mongo.Client) ([]interface{}, error) {
	fooColl := TestColl{"Foo", 10}
	barColl := TestColl{"Bar", 13}
	gopherColl := TestColl{"Gopher", 42}
	testData := []interface{}{fooColl, barColl, gopherColl}

	collection := client.Database(dbName).Collection(collName)
	_, err := collection.InsertMany(context.TODO(), testData)
	if err != nil {
		return []interface{}{}, err
	}
	return testData, nil
}

// getResultsFromCursor iterates over a mongo.Cursor and returns the result
func getResultsFromCursor(cur *mongo.Cursor) ([]interface{}, error) {
	var results []interface{}
	for cur.Next(context.TODO()) {
		var elem TestColl
		err := cur.Decode(&elem)
		if err != nil {
			return []interface{}{}, err
		}
		results = append(results, elem)
	}
	return results, nil
}

// TestColl holds test data for integration tests
type TestColl struct {
	Name string
	Age  int
}
