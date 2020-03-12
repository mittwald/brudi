package config

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/stretchr/testify/assert"

	"github.com/spf13/viper"
)

type InitializeFromViperTestSuite struct {
	suite.Suite
}

func (initializeFromViperTestSuite *InitializeFromViperTestSuite) SetupTest() {
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func (initializeFromViperTestSuite *InitializeFromViperTestSuite) TestInitializeStructFromViper() {
	var exampleConfig = []byte(`
foo:
  bar:
    example: true
    newExample: false
    brudiTest: "foobar"
`)

	assert.NoError(initializeFromViperTestSuite.T(), viper.ReadConfig(bytes.NewBuffer(exampleConfig)))

	isConfig := fooConfig{}
	shouldBeConfig := fooConfig{
		Bar: barConfig{
			Example:        true,
			AnotherExample: false,
			BrudiTest:      "foobar",
		},
	}

	assert.NoError(
		initializeFromViperTestSuite.T(),
		InitializeStructFromViper(
			fmt.Sprintf(
				"%s.%s", "foo", "bar"),
			&isConfig.Bar,
		),
	)

	assert.Equal(initializeFromViperTestSuite.T(), shouldBeConfig, isConfig)
}

func TestInitializeFromViperTestSuite(t *testing.T) {
	suite.Run(t, new(InitializeFromViperTestSuite))
}
