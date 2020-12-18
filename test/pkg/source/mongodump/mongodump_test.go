package mongodump_test

import (
	"bytes"
	"context"
	"fmt"
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

type MongoDumpTestSuite struct {
	suite.Suite
}

type TestLogConsumer struct {
	Msgs []string
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

func (mongoDumpTestSuite *MongoDumpTestSuite) TestBasicMongoDBDump() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		Env: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": "root",
			"MONGO_INITDB_ROOT_PASSWORD": "mongodbroot",
		},
	}
	mongoBackupTarget, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	mongoBackupIP, err := mongoBackupTarget.Host(ctx)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	mongoBackupPort, err := mongoBackupTarget.MappedPort(ctx, "27017/tcp")
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}
	mongoBackupPortStr := fmt.Sprint(mongoBackupPort.Int())

	defer mongoBackupTarget.Terminate(ctx)

	backupClientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", mongoBackupIP, mongoBackupPortStr))
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

	var testMongoConfig = []byte(fmt.Sprintf(`
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
`, mongoBackupIP, fmt.Sprint(mongoBackupPort.Int())))
	err = viper.ReadConfig(bytes.NewBuffer(testMongoConfig))
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}
	err = source.DoBackupForKind(ctx, "mongodump", false, false, false)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	mongoBackupTarget.Terminate(ctx)

	reqRestore := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		Env: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": "root",
			"MONGO_INITDB_ROOT_PASSWORD": "mongodbroot",
		},
	}
	mongoRestoreTarget, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqRestore,
		Started:          true,
	})
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	mongoRestoreIP, err := mongoRestoreTarget.Host(ctx)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	mongoRestorePort, err := mongoRestoreTarget.MappedPort(ctx, "27017/tcp")
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}
	mongoRestorePortStr := fmt.Sprint(mongoRestorePort.Int())

	defer mongoRestoreTarget.Terminate(ctx)

	cmd := exec.CommandContext(ctx, "mongorestore", fmt.Sprintf("--host=%s", mongoRestoreIP), fmt.Sprintf("--port=%s", mongoRestorePortStr),
		"--archive=/tmp/dump.tar.gz", "--gzip", "--username=root", "--password=mongodbroot")
	_, err = cmd.CombinedOutput()
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	var results []interface{}
	findOptions := options.Find()

	restoreClientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", mongoRestoreIP, mongoRestorePortStr))
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
}

func (mongoDumpTestSuite *MongoDumpTestSuite) TestBasicMongoDBDumpRestic() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		Env: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": "root",
			"MONGO_INITDB_ROOT_PASSWORD": "mongodbroot",
		},
	}
	mongoBackupTarget, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	mongoBackupIP, err := mongoBackupTarget.Host(ctx)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	mongoBackupPort, err := mongoBackupTarget.MappedPort(ctx, "27017/tcp")
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}
	mongoBackupPortStr := fmt.Sprint(mongoBackupPort.Int())

	resticReq := testcontainers.ContainerRequest{
		Image:        "restic/rest-server:latest",
		ExposedPorts: []string{"8000/tcp"},
		Env: map[string]string{
			"OPTIONS":         "--no-auth",
			"RESTIC_PASSWORD": "mongorepo",
		},
	}
	resticContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: resticReq,
		Started:          true,
	})
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	//resticIP, err := resticContainer.Host(ctx)
	//if err != nil {
	//	mongoDumpTestSuite.Error(err)
	//}

	resticPort, err := resticContainer.MappedPort(ctx, "8000/tcp")
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}
	resticPortStr := fmt.Sprint(resticPort.Int())

	defer resticContainer.Terminate(ctx)

	backupClientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", mongoBackupIP, mongoBackupPortStr))
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

	var testMongoConfig = []byte(fmt.Sprintf(`
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
            repo: rest:http://127.0.0.1:%s/
        forget:
          flags:
            keepLast: 1
            keepHourly: 0
            keepDaily: 0
            keepWeekly: 0
            keepMonthly: 0
            keepYearly: 0
`, mongoBackupIP, mongoBackupPortStr, resticPortStr))

	err = viper.ReadConfig(bytes.NewBuffer(testMongoConfig))
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	err = source.DoBackupForKind(ctx, "mongodump", true, true, false)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	mongoBackupTarget.Terminate(ctx)

	reqRestore := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		Env: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": "root",
			"MONGO_INITDB_ROOT_PASSWORD": "mongodbroot",
		},
	}
	mongoRestoreTarget, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqRestore,
		Started:          true,
	})
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	mongoRestoreIP, err := mongoRestoreTarget.Host(ctx)
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	mongoRestorePort, err := mongoRestoreTarget.MappedPort(ctx, "27017/tcp")
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}
	mongoRestorePortStr := fmt.Sprint(mongoRestorePort.Int())

	defer mongoRestoreTarget.Terminate(ctx)

	cmd := exec.CommandContext(ctx, "restic", "restore", "-r", fmt.Sprintf("rest:http://127.0.0.1:%s/", resticPortStr),
		"--target", "data", "latest")
	_, err = cmd.CombinedOutput()
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	cmd = exec.CommandContext(ctx, "mongorestore", fmt.Sprintf("--host=%s", mongoRestoreIP), fmt.Sprintf("--port=%s", mongoRestorePortStr),
		"--archive=data/tmp/dump.tar.gz", "--gzip", "--username=root", "--password=mongodbroot")
	_, err = cmd.CombinedOutput()
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	cmd = exec.CommandContext(ctx, "rm", "-rf", "data")
	_, err = cmd.CombinedOutput()
	if err != nil {
		mongoDumpTestSuite.Error(err)
	}

	var results []interface{}
	findOptions := options.Find()

	restoreClientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", mongoRestoreIP, mongoRestorePortStr))
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
}

func TestMongoDumpTestSuite(t *testing.T) {
	suite.Run(t, new(MongoDumpTestSuite))
}
