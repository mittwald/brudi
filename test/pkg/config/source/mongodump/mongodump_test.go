package mongodump_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mittwald/brudi/pkg/source"
	"github.com/spf13/viper"
	"github.com/testcontainers/testcontainers-go"
	"strings"
	"testing"
)

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
	defer mongoC.Terminate(ctx)

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
}
