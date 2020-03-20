package config

import (
	"bytes"
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

var exampleConfig = []byte(`
foo:
  bar:
    example: true
    newExample: false
    brudiTest: "foobar"
`)

func (initializeFromViperTestSuite *InitializeFromViperTestSuite) TestInitializeStructFromViperWithoutTags() {
	assert.NoError(initializeFromViperTestSuite.T(), viper.ReadConfig(bytes.NewBuffer(exampleConfig)))

	isConfig := untaggedFooConfig{}
	shouldBeConfig := untaggedFooConfig{
		Bar: untaggedBarConfig{
			Example:        true,
			AnotherExample: false,
			BrudiTest:      "foobar",
		},
	}

	assert.NoError(
		initializeFromViperTestSuite.T(),
		InitializeStructFromViper(
			"foo",
			&isConfig,
		),
	)

	assert.Equal(initializeFromViperTestSuite.T(), shouldBeConfig, isConfig)
}

func (initializeFromViperTestSuite *InitializeFromViperTestSuite) TestInitializeStructFromViperWithTags() {
	assert.NoError(initializeFromViperTestSuite.T(), viper.ReadConfig(bytes.NewBuffer(exampleConfig)))

	isConfig := taggedFooConfig{
		CustomBar: &taggedBarConfig{},
	}
	shouldBeConfig := taggedFooConfig{
		CustomBar: &taggedBarConfig{
			CustomExample:        true,
			CustomAnotherExample: false,
			CustomBrudiTest:      "foobar",
		},
	}

	assert.NoError(
		initializeFromViperTestSuite.T(),
		InitializeStructFromViper(
			"foo",
			&isConfig,
		),
	)

	assert.Equal(initializeFromViperTestSuite.T(), shouldBeConfig, isConfig)
}

func TestInitializeFromViperTestSuite(t *testing.T) {
	suite.Run(t, new(InitializeFromViperTestSuite))
}
