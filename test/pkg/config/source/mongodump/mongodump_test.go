package mongodump_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mittwald/brudi/pkg/source"
	"github.com/spf13/viper"
	"github.com/testcontainers/testcontainers-go"
	"log"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TestColl struct {
	Name string
	Age  int
}

func TestBasicMongoDBDump(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		Env: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": "root",
			"MONGO_INITDB_ROOT_PASSWORD": "mongodbroot",
		},
	}
	mongoC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}

	ip, err := mongoC.Host(ctx)
	if err != nil {
		t.Error(err)
	}

	mongoPort, err := mongoC.MappedPort(ctx, "27017/tcp")
	if err != nil {
		t.Error(err)
	}
	mongoPortStr := fmt.Sprint(mongoPort.Int())

	defer mongoC.Terminate(ctx)

	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", ip, mongoPortStr))
	clientAuth := options.Client().SetAuth(options.Credential{Username: "root", Password: "mongodbroot"})
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions, clientAuth)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fooColl := TestColl{"Foo", 10}
	barColl := TestColl{"Bar", 13}
	gopherColl := TestColl{"Gopher", 42}

	testData := []interface{}{fooColl, barColl, gopherColl}

	collection := client.Database("test").Collection("testColl")

	_, err = collection.InsertMany(context.TODO(), testData)
	if err != nil {
		log.Fatal(err)
	}

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
`, ip, fmt.Sprint(mongoPort.Int())))
	viper.Reset()
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	err = viper.ReadConfig(bytes.NewBuffer(testMongoConfig))
	if err != nil {
		t.Error(err)
	}
	err = source.DoBackupForKind(ctx, "mongodump", false, false, false)
	if err != nil {
		t.Error(err)
	}

	//ctxTest := context.Background()
	//reqTest := testcontainers.ContainerRequest{
	//	Image:        "mongo:latest",
	//	ExposedPorts: []string{"27017/tcp"},
	//	Env: map[string]string{
	//		"MONGO_INITDB_ROOT_USERNAME": "root",
	//		"MONGO_INITDB_ROOT_PASSWORD": "mongodbroot",
	//	},
	//}
	//mongoTest, err := testcontainers.GenericContainer(ctxTest, testcontainers.GenericContainerRequest{
	//	ContainerRequest: reqTest,
	//	Started:          true,
	//})
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//mongoTestIP, err := mongoTest.Host(ctx)
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//mongoTestPort, err := mongoTest.MappedPort(ctx, "27017/tcp")
	//if err != nil {
	//	t.Error(err)
	//}
	//mongoTestPortStr := fmt.Sprint(mongoPort.Int())
	//
	//defer mongoTest.Terminate(ctx)
}
