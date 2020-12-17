package testconfig

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mittwald/brudi/pkg/config"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MergeConfigsTestSuite struct {
	suite.Suite
}

func (mergeConfigsTestSuite *MergeConfigsTestSuite) SetupTest() {
	viper.Reset()
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func (mergeConfigsTestSuite *MergeConfigsTestSuite) TearDownTest() {
	viper.Reset()
}

var testDBConfig = []byte(`
      mongodump:
        options:
          flags:
            host: 127.0.0.1
            port: 27017
            username: root
            password: mongodbroot
            gzip: true
            archive: /tmp/dump.tar.gz
          additionalArgs: []
`)

var testResticConfig = []byte(`
      restic:
        global:
          flags:
            repo: "s3:s3.eu-central-1.amazonaws.com/your.s3.bucket/myResticRepo"
        forget:
          flags:
            keepLast: 1
            keepHourly: 0
            keepDaily: 0
            keepWeekly: 0
            keepMonthly: 0
            keepYearly: 0
`)

var expectedConfig = []byte(`
      mongodump:
        options:
          flags:
            host: 127.0.0.1
            port: 27017
            username: root
            password: mongodbroot
            gzip: true
            archive: /tmp/dump.tar.gz
          additionalArgs: []
      restic:
        global:
          flags:
            repo: "s3:s3.eu-central-1.amazonaws.com/your.s3.bucket/myResticRepo"
        forget:
          flags:
            keepLast: 1
            keepHourly: 0
            keepDaily: 0
            keepWeekly: 0
            keepMonthly: 0
            keepYearly: 0
`)

func (mergeConfigsTestSuite *MergeConfigsTestSuite) TestMergeConfigs() {
	testData := []*bytes.Buffer{bytes.NewBuffer(testDBConfig), bytes.NewBuffer(testResticConfig)}
	config.MergeConfigs(testData)
	testResult := viper.AllSettings()

	assert.NoError(mergeConfigsTestSuite.T(), viper.ReadConfig(bytes.NewBuffer(expectedConfig)))
	shouldBeConfig := viper.AllSettings()
	assert.Equal(mergeConfigsTestSuite.T(), shouldBeConfig, testResult)
}

func TestMergeConfigsTestSuite(t *testing.T) {
	suite.Run(t, new(MergeConfigsTestSuite))
}
