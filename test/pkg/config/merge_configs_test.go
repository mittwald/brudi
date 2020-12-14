package testconfig

import (
	"bytes"
	"fmt"
	"github.com/mittwald/brudi/pkg/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
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

func (mergeConfigsTestSuite *MergeConfigsTestSuite) TestMergeConfigs() {

	testConfigs := []string{"../..//testdata/configA_1.yaml", "../../testdata/configA_2.yaml"}

	config.MergeConfigs(testConfigs)
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
	testResult := viper.AllSettings()
	fmt.Println(testResult)
	assert.NoError(mergeConfigsTestSuite.T(), viper.ReadConfig(bytes.NewBuffer(expectedConfig)))
	shouldBeConfig := viper.AllSettings()
	fmt.Println(shouldBeConfig)
	assert.Equal(mergeConfigsTestSuite.T(), shouldBeConfig, testResult)

}

func TestMergeConfigsTestSuite(t *testing.T) {
	suite.Run(t, new(MergeConfigsTestSuite))
}
