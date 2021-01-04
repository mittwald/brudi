package mongodump_test

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

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gotest.tools/assert"
)

const mongoPort = "27017/tcp"
const backupPath = "/tmp/dump.tar.gz"
const mongoPW = "mongodbroot"
const mongoUser = "root"
const dataDir = "data"

type TestColl struct {
	Name string
	Age  int
}

type MongoDumpTestSuite struct {
	suite.Suite
}

func (mongoDumpTestSuite *MongoDumpTestSuite) SetupTest() {
	viper.Reset()
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func (mongoDumpTestSuite *MongoDumpTestSuite) TearDownTest() {
	viper.Reset()
}

var mongoRequest = testcontainers.ContainerRequest{
	Image:        "mongo:latest",
	ExposedPorts: []string{mongoPort},
	Env: map[string]string{
		"MONGO_INITDB_ROOT_USERNAME": mongoUser,
		"MONGO_INITDB_ROOT_PASSWORD": mongoUser,
	},
}

// execCommand executes a given command within a context and with specified arguments
func execCommand(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, cmd, args...)
	out, err := command.CombinedOutput()
	if err != nil {
		return out, err
	}
	return out, nil
}

// newMongoClient creates a mongo client connected to the database specified by the provided TestContainerSetup
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

// createMongoConfig creates a brudi config for the mongodump command
func createMongoConfig(container commons.TestContainerSetup, useRestic bool, resticIP, resticPort string) []byte {
	if !useRestic {
		return []byte(fmt.Sprintf(`
      mongodump:
        options:
          flags:
            host: %s
            port: %s
            username: %s
            password: %s
            gzip: true
            archive: %s
          additionalArgs: []
`, container.Address, container.Port, mongoUser, mongoPW, backupPath))
	}
	return []byte(fmt.Sprintf(`
      mongodump:
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
`, container.Address, container.Port, mongoUser, mongoPW, backupPath, resticIP, resticPort))
}

// prepareTestData creates test data and writes it into a database using the provided client
func prepareTestData(client *mongo.Client) ([]interface{}, error) {
	fooColl := TestColl{"Foo", 10}
	barColl := TestColl{"Bar", 13}
	gopherColl := TestColl{"Gopher", 42}
	testData := []interface{}{fooColl, barColl, gopherColl}
	collection := client.Database("test").Collection("testColl")
	_, err := collection.InsertMany(context.TODO(), testData)
	if err != nil {
		return []interface{}{}, err
	}
	return testData, nil
}

// getResultsFromCursor iterates over a mongo.Cursor and returns the result.It also closes the cursor when done.
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
	err := cur.Close(context.TODO())
	if err != nil {
		return []interface{}{}, err
	}
	return results, nil
}

func mongoDoBackup(ctx context.Context, mongoDumpTestSuite *MongoDumpTestSuite, useRestic bool,
	resticContainer commons.TestContainerSetup) []interface{} {
	// create a mongo container to test backup function
	mongoBackupTarget, err := commons.NewTestContainerSetup(ctx, &mongoRequest, mongoPort)
	mongoDumpTestSuite.Require().NoError(err)
	defer func() {
		backErr := mongoBackupTarget.Container.Terminate(ctx)
		mongoDumpTestSuite.Require().NoError(backErr)
	}()
	var backupClient mongo.Client
	backupClient, err = newMongoClient(&mongoBackupTarget)
	mongoDumpTestSuite.Require().NoError(err)
	defer func() {
		clientErr := backupClient.Disconnect(ctx)
		mongoDumpTestSuite.Require().NoError(clientErr)
	}()

	// write test data into database and retain it for later assertion
	var testData []interface{}
	testData, err = prepareTestData(&backupClient)
	mongoDumpTestSuite.Require().NoError(err)

	testMongoConfig := createMongoConfig(mongoBackupTarget, useRestic, resticContainer.Address, resticContainer.Port)
	err = viper.ReadConfig(bytes.NewBuffer(testMongoConfig))
	mongoDumpTestSuite.Require().NoError(err)

	// perform backup action on first mongo container
	err = source.DoBackupForKind(ctx, "mongodump", false, useRestic, false)
	mongoDumpTestSuite.Require().NoError(err)

	return testData
}

// TestBasicMongoDBDump performs an integration test for the `mongodump` command
func (mongoDumpTestSuite *MongoDumpTestSuite) TestBasicMongoDBDump() {
	ctx := context.Background()

	defer func() {
		removeErr := os.Remove(backupPath)
		mongoDumpTestSuite.Require().NoError(removeErr)
	}()

	testData := mongoDoBackup(ctx, mongoDumpTestSuite, false, commons.TestContainerSetup{Port: "", Address: ""})

	// setup a new mongo container which will be used to ensure data was backed up correctly
	mongoRestoreTarget, err := commons.NewTestContainerSetup(ctx, &mongoRequest, mongoPort)
	mongoDumpTestSuite.Require().NoError(err)
	defer func() {
		restoreErr := mongoRestoreTarget.Container.Terminate(ctx)
		mongoDumpTestSuite.Require().NoError(restoreErr)
	}()

	// use `mongorestore` to restore backed up data to new container
	_, err = execCommand(ctx, "mongorestore", fmt.Sprintf("--host=%s", mongoRestoreTarget.Address),
		fmt.Sprintf("--port=%s", mongoRestoreTarget.Port), fmt.Sprintf("--archive=%s", backupPath),
		"--gzip", fmt.Sprintf("--username=%s", mongoUser),
		fmt.Sprintf("--password=%s", mongoPW))
	mongoDumpTestSuite.Require().NoError(err)

	restoreClient, err := newMongoClient(&mongoRestoreTarget)
	mongoDumpTestSuite.Require().NoError(err)
	defer func() {
		clientErr := restoreClient.Disconnect(ctx)
		mongoDumpTestSuite.Require().NoError(clientErr)
	}()

	restoredCollection := restoreClient.Database("test").Collection("testColl")

	findOptions := options.Find()
	var cur *mongo.Cursor
	cur, err = restoredCollection.Find(context.TODO(), bson.D{{}}, findOptions)
	mongoDumpTestSuite.Require().NoError(err)

	var results []interface{}
	results, err = getResultsFromCursor(cur)
	mongoDumpTestSuite.Require().NoError(err)

	// check if the original data was restored
	assert.DeepEqual(mongoDumpTestSuite.T(), testData, results)
}

func (mongoDumpTestSuite *MongoDumpTestSuite) TestBasicMongoDBDumpRestic() {
	ctx := context.Background()

	// create a container running the restic rest-server
	resticContainer, err := commons.NewTestContainerSetup(ctx, &commons.ResticReq, commons.ResticPort)
	mongoDumpTestSuite.Require().NoError(err)
	defer func() {
		resticErr := resticContainer.Container.Terminate(ctx)
		mongoDumpTestSuite.Require().NoError(resticErr)
	}()

	testData := mongoDoBackup(ctx, mongoDumpTestSuite, true, resticContainer)

	var mongoRestoreTarget commons.TestContainerSetup
	mongoRestoreTarget, err = commons.NewTestContainerSetup(ctx, &mongoRequest, mongoPort)
	mongoDumpTestSuite.Require().NoError(err)
	defer func() {
		restoreErr := mongoRestoreTarget.Container.Terminate(ctx)
		mongoDumpTestSuite.Require().NoError(restoreErr)
	}()

	// restore backed up data from restic repository
	cmd := exec.CommandContext(ctx, "restic", "restore", "-r", fmt.Sprintf("rest:http://%s:%s/",
		resticContainer.Address, resticContainer.Port),
		"--target", dataDir, "latest")
	_, err = cmd.CombinedOutput()
	mongoDumpTestSuite.Require().NoError(err)

	cmd = exec.CommandContext(ctx, "mongorestore", fmt.Sprintf("--host=%s", mongoRestoreTarget.Address),
		fmt.Sprintf("--port=%s", mongoRestoreTarget.Port),
		fmt.Sprintf("--archive=%s/%s", dataDir, backupPath), "--gzip", fmt.Sprintf("--username=%s", mongoUser),
		fmt.Sprintf("--password=%s", mongoPW))
	_, err = cmd.CombinedOutput()
	mongoDumpTestSuite.Require().NoError(err)

	// remove backup directory
	err = os.RemoveAll(dataDir)
	mongoDumpTestSuite.Require().NoError(err)

	var restoreClient mongo.Client
	restoreClient, err = newMongoClient(&mongoRestoreTarget)
	mongoDumpTestSuite.Require().NoError(err)
	defer func() {
		clientErr := restoreClient.Disconnect(ctx)
		mongoDumpTestSuite.Require().NoError(clientErr)
	}()

	restoredCollection := restoreClient.Database("test").Collection("testColl")

	findOptions := options.Find()
	var cur *mongo.Cursor
	cur, err = restoredCollection.Find(context.TODO(), bson.D{{}}, findOptions)
	mongoDumpTestSuite.Require().NoError(err)

	var results []interface{}
	results, err = getResultsFromCursor(cur)
	mongoDumpTestSuite.Require().NoError(err)

	assert.DeepEqual(mongoDumpTestSuite.T(), testData, results)
}

func TestMongoDumpTestSuite(t *testing.T) {
	suite.Run(t, new(MongoDumpTestSuite))
}
