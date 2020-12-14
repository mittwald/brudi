package testconfig

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/mittwald/brudi/pkg/config"

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
    brudiNumber: 2222
`)

func (initializeFromViperTestSuite *InitializeFromViperTestSuite) TestInitializeStructFromViperWithoutTags() {
	assert.NoError(initializeFromViperTestSuite.T(), viper.ReadConfig(bytes.NewBuffer(exampleConfig)))

	isConfig := untaggedFooConfig{}
	shouldBeConfig := untaggedFooConfig{
		Bar: untaggedBarConfig{
			Example:        true,
			AnotherExample: false,
			BrudiTest:      "foobar",
			BrudiNumber:    2222,
		},
	}

	assert.NoError(
		initializeFromViperTestSuite.T(),
		config.InitializeStructFromViper(
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
			CustomBrudiNumber:    2222,
		},
	}

	assert.NoError(
		initializeFromViperTestSuite.T(),
		config.InitializeStructFromViper(
			"foo",
			&isConfig,
		),
	)

	assert.Equal(initializeFromViperTestSuite.T(), shouldBeConfig, isConfig)
}

func (initializeFromViperTestSuite *InitializeFromViperTestSuite) TestInitializeStructFromViperOverwriteEnv() {
	assert.NoError(initializeFromViperTestSuite.T(), viper.ReadConfig(bytes.NewBuffer(exampleConfig)))

	os.Setenv("FOO_BAR_BRUDINUMBER", "2223")
	os.Setenv("FOO_BAR_BRUDITEST", "foobar1")
	os.Setenv("FOO_BAR_ANOTHEREXAMPLE", "true")

	defer os.Unsetenv("FOO_BAR_BRUDITEST")
	defer os.Unsetenv("FOO_BAR_BRUDINUMBER")
	defer os.Unsetenv("FOO_BAR_ANOTHEREXAMPLE")

	isConfig := taggedFooConfig{
		CustomBar: &taggedBarConfig{},
	}
	shouldBeConfig := taggedFooConfig{
		CustomBar: &taggedBarConfig{
			CustomExample:        true,
			CustomAnotherExample: true,
			CustomBrudiTest:      "foobar1",
			CustomBrudiNumber:    2223,
		},
	}

	assert.NoError(
		initializeFromViperTestSuite.T(),
		config.InitializeStructFromViper(
			"foo",
			&isConfig,
		),
	)

	assert.Equal(initializeFromViperTestSuite.T(), shouldBeConfig, isConfig)
}

func TestInitializeFromViperTestSuite(t *testing.T) {
	suite.Run(t, new(InitializeFromViperTestSuite))
}
