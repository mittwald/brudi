package mongodump_test

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/mittwald/brudi/pkg/source"

	"github.com/docker/go-connections/nat"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gotest.tools/assert"
)

type TestColl struct {
	Name string
	Age  int
}

type TestContainerSetup struct {
	Container testcontainers.Container
	Address   string
	Port      string
}

type MongoDumpTestSuite struct {
	suite.Suite
}

type TestLogConsumer struct {
	Msgs []string
}

var mongoRequest = testcontainers.ContainerRequest{
	Image:        "mongo:latest",
	ExposedPorts: []string{"27017/tcp"},
	Env: map[string]string{
		"MONGO_INITDB_ROOT_USERNAME": "root",
		"MONGO_INITDB_ROOT_PASSWORD": "mongodbroot",
	},
}

var resticReq = testcontainers.ContainerRequest{
	Image:        "restic/rest-server:latest",
	ExposedPorts: []string{"8000/tcp"},
	Env: map[string]string{
		"OPTIONS":         "--no-auth",
		"RESTIC_PASSWORD": "mongorepo",
	},
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

func (g *TestLogConsumer) Accept(l testcontainers.Log) {
	g.Msgs = append(g.Msgs, string(l.Content))
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

func execCommand(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, cmd, args...)
	out, err := command.CombinedOutput()
	if err != nil {
		return out, err
	}
	return out, nil
}

func newMongoClient(target *TestContainerSetup) (mongo.Client, error) {
	backupClientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", target.Address,
		target.Port))
	clientAuth := options.Client().SetAuth(options.Credential{Username: "root", Password: "mongodbroot"})

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

func createMongoConfig(container TestContainerSetup, useRestic bool, resticIP, resticPort string) []byte {
	if !useRestic {
		return []byte(fmt.Sprintf(`
      mongodump:
        options:
          flags:
            host: %s
            port: %s
            username: root
            password: mongodbroot
            gzip: true
            archive: /tmp/dump.tar.gz
          additionalArgs: []
`, container.Address, container.Port))
	}
	return []byte(fmt.Sprintf(`
      mongodump:
        options:
          flags:
            host: %s
            port: %s
            username: root
            password: mongodbroot
            gzip: true
            archive: /tmp/dump.tar.gz
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

func newMongoCollection(client *mongo.Client) ([]interface{}, error) {
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

func (mongoDumpTestSuite *MongoDumpTestSuite) TestBasicMongoDBDump() {
	ctx := context.Background()
	// Create a database to test backup function
	mongoBackupTarget, err := newTestContainerSetup(ctx, &mongoRequest, "27017/tcp")
	mongoDumpTestSuite.Require().NoError(err)

	backupClient, err := newMongoClient(&mongoBackupTarget)
	mongoDumpTestSuite.Require().NoError(err)

	testData, err := newMongoCollection(&backupClient)
	mongoDumpTestSuite.Require().NoError(err)

	err = backupClient.Disconnect(context.TODO())
	mongoDumpTestSuite.Require().NoError(err)

	testMongoConfig := createMongoConfig(mongoBackupTarget, false, "", "")

	err = viper.ReadConfig(bytes.NewBuffer(testMongoConfig))
	mongoDumpTestSuite.Require().NoError(err)
	err = source.DoBackupForKind(ctx, "mongodump", false, false, false)
	mongoDumpTestSuite.Require().NoError(err)

	err = mongoBackupTarget.Container.Terminate(ctx)
	mongoDumpTestSuite.Require().NoError(err)

	mongoRestoreTarget, err := newTestContainerSetup(ctx, &mongoRequest, "27017/tcp")
	mongoDumpTestSuite.Require().NoError(err)

	_, err = execCommand(ctx, "mongorestore", fmt.Sprintf("--host=%s", mongoRestoreTarget.Address),
		fmt.Sprintf("--port=%s", mongoRestoreTarget.Port), "--archive=/tmp/dump.tar.gz", "--gzip", "--username=root",
		"--password=mongodbroot")
	mongoDumpTestSuite.Require().NoError(err)

	restoreClient, err := newMongoClient(&mongoRestoreTarget)
	mongoDumpTestSuite.Require().NoError(err)

	restoredCollection := restoreClient.Database("test").Collection("testColl")

	findOptions := options.Find()
	cur, err := restoredCollection.Find(context.TODO(), bson.D{{}}, findOptions)
	mongoDumpTestSuite.Require().NoError(err)

	results, err := getResultsFromCursor(cur)
	mongoDumpTestSuite.Require().NoError(err)

	assert.DeepEqual(mongoDumpTestSuite.T(), testData, results)
	err = mongoRestoreTarget.Container.Terminate(ctx)
	mongoDumpTestSuite.Require().NoError(err)
	err = restoreClient.Disconnect(context.TODO())
	mongoDumpTestSuite.Require().NoError(err)
}

func (mongoDumpTestSuite *MongoDumpTestSuite) TestBasicMongoDBDumpRestic() {
	ctx := context.Background()

	mongoBackupTarget, err := newTestContainerSetup(ctx, &mongoRequest, "27017/tcp")
	mongoDumpTestSuite.Require().NoError(err)

	resticContainer, err := newTestContainerSetup(ctx, &resticReq, "8000/tcp")
	mongoDumpTestSuite.Require().NoError(err)

	backupClient, err := newMongoClient(&mongoBackupTarget)
	mongoDumpTestSuite.Require().NoError(err)

	err = backupClient.Ping(context.TODO(), nil)
	mongoDumpTestSuite.Require().NoError(err)

	testData, err := newMongoCollection(&backupClient)
	mongoDumpTestSuite.Require().NoError(err)

	err = backupClient.Disconnect(context.TODO())
	mongoDumpTestSuite.Require().NoError(err)

	testMongoConfig := createMongoConfig(mongoBackupTarget, true, resticContainer.Address, resticContainer.Port)

	err = viper.ReadConfig(bytes.NewBuffer(testMongoConfig))
	mongoDumpTestSuite.Require().NoError(err)

	err = source.DoBackupForKind(ctx, "mongodump", true, true, false)
	mongoDumpTestSuite.Require().NoError(err)

	err = mongoBackupTarget.Container.Terminate(ctx)
	mongoDumpTestSuite.Require().NoError(err)

	mongoRestoreTarget, err := newTestContainerSetup(ctx, &mongoRequest, "27017/tcp")
	mongoDumpTestSuite.Require().NoError(err)

	cmd := exec.CommandContext(ctx, "restic", "restore", "-r", fmt.Sprintf("rest:http://%s:%s/",
		resticContainer.Address, resticContainer.Port),
		"--target", "data", "latest")
	_, err = cmd.CombinedOutput()
	mongoDumpTestSuite.Require().NoError(err)

	cmd = exec.CommandContext(ctx, "mongorestore", fmt.Sprintf("--host=%s", mongoRestoreTarget.Address),
		fmt.Sprintf("--port=%s", mongoRestoreTarget.Port),
		"--archive=data/tmp/dump.tar.gz", "--gzip", "--username=root", "--password=mongodbroot")
	_, err = cmd.CombinedOutput()
	mongoDumpTestSuite.Require().NoError(err)

	cmd = exec.CommandContext(ctx, "rm", "-rf", "data")
	_, err = cmd.CombinedOutput()
	mongoDumpTestSuite.Require().NoError(err)

	restoreClient, err := newMongoClient(&mongoRestoreTarget)
	mongoDumpTestSuite.Require().NoError(err)

	restoredCollection := restoreClient.Database("test").Collection("testColl")

	findOptions := options.Find()
	cur, err := restoredCollection.Find(context.TODO(), bson.D{{}}, findOptions)
	mongoDumpTestSuite.Require().NoError(err)

	results, err := getResultsFromCursor(cur)
	mongoDumpTestSuite.Require().NoError(err)

	assert.DeepEqual(mongoDumpTestSuite.T(), testData, results)

	err = resticContainer.Container.Terminate(ctx)
	mongoDumpTestSuite.Require().NoError(err)

	err = restoreClient.Disconnect(context.TODO())
	mongoDumpTestSuite.Require().NoError(err)

	err = mongoRestoreTarget.Container.Terminate(ctx)
	mongoDumpTestSuite.Require().NoError(err)
}

func TestMongoDumpTestSuite(t *testing.T) {
	suite.Run(t, new(MongoDumpTestSuite))
}
