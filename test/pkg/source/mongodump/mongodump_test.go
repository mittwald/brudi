package mongodump_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/mittwald/brudi/pkg/source"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/bson"
	"gotest.tools/assert"
	"os/exec"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func newTestContainerSetup(ctx context.Context, request testcontainers.ContainerRequest, port nat.Port) (TestContainerSetup, error) {
	result := TestContainerSetup{}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
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

func createMongoConfig(container TestContainerSetup, useRestic bool, resticIP string, resticPort string) []byte {
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
	} else {
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
}

func (mongoDumpTestSuite *MongoDumpTestSuite) TestBasicMongoDBDump() {
	ctx := context.Background()
	mongoBackupTarget, err := newTestContainerSetup(ctx, mongoRequest, "27017/tcp")
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	backupClientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", mongoBackupTarget.Address,
		mongoBackupTarget.Port))
	clientAuth := options.Client().SetAuth(options.Credential{Username: "root", Password: "mongodbroot"})
	// Connect to MongoDB
	backupClient, err := mongo.Connect(context.TODO(), backupClientOptions, clientAuth)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	// Check the connection
	err = backupClient.Ping(context.TODO(), nil)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	fooColl := TestColl{"Foo", 10}
	barColl := TestColl{"Bar", 13}
	gopherColl := TestColl{"Gopher", 42}
	testData := []interface{}{fooColl, barColl, gopherColl}

	collection := backupClient.Database("test").Collection("testColl")

	_, err = collection.InsertMany(context.TODO(), testData)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	backupClient.Disconnect(context.TODO())

	testMongoConfig := createMongoConfig(mongoBackupTarget, false, "", "")

	err = viper.ReadConfig(bytes.NewBuffer(testMongoConfig))
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}
	err = source.DoBackupForKind(ctx, "mongodump", false, false, false)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	mongoBackupTarget.Container.Terminate(ctx)

	mongoRestoreTarget, err := newTestContainerSetup(ctx, mongoRequest, "27017/tcp")
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	_, err = execCommand(ctx, "mongorestore", fmt.Sprintf("--host=%s", mongoRestoreTarget.Address),
		fmt.Sprintf("--port=%s", mongoRestoreTarget.Port), "--archive=/tmp/dump.tar.gz", "--gzip", "--username=root",
		"--password=mongodbroot")
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	var results []interface{}
	findOptions := options.Find()

	restoreClientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", mongoRestoreTarget.Address,
		mongoRestoreTarget.Port))
	// Connect to MongoDB
	restoreClient, err := mongo.Connect(context.TODO(), restoreClientOptions, clientAuth)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	restoredCollection := restoreClient.Database("test").Collection("testColl")

	cur, err := restoredCollection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	for cur.Next(context.TODO()) {
		var elem TestColl
		err := cur.Decode(&elem)
		if err != nil {
			mongoDumpTestSuite.Error(err)
		}
		results = append(results, elem)
	}
	if err := cur.Err(); err != nil {
		mongoDumpTestSuite.Error(err)
	}
	cur.Close(context.TODO())

	assert.DeepEqual(mongoDumpTestSuite.T(), testData, results)
	mongoRestoreTarget.Container.Terminate(ctx)
}

func (mongoDumpTestSuite *MongoDumpTestSuite) TestBasicMongoDBDumpRestic() {
	ctx := context.Background()

	mongoBackupTarget, err := newTestContainerSetup(ctx, mongoRequest, "27017/tcp")
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	resticContainer, err := newTestContainerSetup(ctx, resticReq, "8000/tcp")
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	backupClientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", mongoBackupTarget.Address,
		mongoBackupTarget.Port))
	clientAuth := options.Client().SetAuth(options.Credential{Username: "root", Password: "mongodbroot"})
	backupClient, err := mongo.Connect(context.TODO(), backupClientOptions, clientAuth)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	err = backupClient.Ping(context.TODO(), nil)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	fooColl := TestColl{"Foo", 10}
	barColl := TestColl{"Bar", 13}
	gopherColl := TestColl{"Gopher", 42}
	testData := []interface{}{fooColl, barColl, gopherColl}

	collection := backupClient.Database("test").Collection("testColl")

	_, err = collection.InsertMany(context.TODO(), testData)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	backupClient.Disconnect(context.TODO())

	testMongoConfig := createMongoConfig(mongoBackupTarget, true, resticContainer.Address, resticContainer.Port)

	err = viper.ReadConfig(bytes.NewBuffer(testMongoConfig))
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	err = source.DoBackupForKind(ctx, "mongodump", true, true, false)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	mongoBackupTarget.Container.Terminate(ctx)

	mongoRestoreTarget, err := newTestContainerSetup(ctx, mongoRequest, "27017/tcp")
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	defer mongoRestoreTarget.Container.Terminate(ctx)

	cmd := exec.CommandContext(ctx, "restic", "restore", "-r", fmt.Sprintf("rest:http://%s:%s/",
		resticContainer.Address, resticContainer.Port),
		"--target", "data", "latest")
	out, err := cmd.CombinedOutput()
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}
	fmt.Println(string(out))

	cmd = exec.CommandContext(ctx, "mongorestore", fmt.Sprintf("--host=%s", mongoRestoreTarget.Address),
		fmt.Sprintf("--port=%s", mongoRestoreTarget.Port),
		"--archive=data/tmp/dump.tar.gz", "--gzip", "--username=root", "--password=mongodbroot")
	out, err = cmd.CombinedOutput()
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}
	fmt.Println(string(out))

	cmd = exec.CommandContext(ctx, "rm", "-rf", "data")
	_, err = cmd.CombinedOutput()
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	var results []interface{}
	findOptions := options.Find()

	restoreClientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s",
		mongoRestoreTarget.Address, mongoRestoreTarget.Port))
	restoreClient, err := mongo.Connect(context.TODO(), restoreClientOptions, clientAuth)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	restoredCollection := restoreClient.Database("test").Collection("testColl")

	cur, err := restoredCollection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	for cur.Next(context.TODO()) {
		var elem TestColl
		err := cur.Decode(&elem)
		if err != nil {
			mongoDumpTestSuite.Error(err)
		}
		results = append(results, elem)
	}
	if err := cur.Err(); err != nil {
		mongoDumpTestSuite.Error(err)
	}
	cur.Close(context.TODO())

	assert.DeepEqual(mongoDumpTestSuite.T(), testData, results)
	resticContainer.Container.Terminate(ctx)
}

func TestMongoDumpTestSuite(t *testing.T) {
	suite.Run(t, new(MongoDumpTestSuite))
}
