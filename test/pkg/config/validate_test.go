package testconfig

import (
	"fmt"
	"testing"

	"github.com/mittwald/brudi/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ValidationTestSuite struct {
	suite.Suite
}

func (validationSuite *ValidationTestSuite) TestSucceedValidationOnTags() {
	testConfig := untaggedFooConfig{
		Bar: untaggedBarConfig{
			Example:        true,
			AnotherExample: false,
			BrudiTest:      "asdf",
		},
	}

	assert.NoError(
		validationSuite.T(),
		config.Validate(testConfig),
	)
}

func (validationSuite *ValidationTestSuite) TestFailValidationOnTags() {
	testConfig := untaggedFooConfig{
		Bar: untaggedBarConfig{
			Example:        true,
			AnotherExample: false,
			BrudiTest:      "",
		},
	}

	assert.EqualError(
		validationSuite.T(),
		config.Validate(testConfig),
		"Key: 'untaggedFooConfig.Bar.BrudiTest' Error:Field validation for 'BrudiTest' failed on the 'min' tag",
	)
}

func (validationSuite *ValidationTestSuite) TestSucceedValidationOnFunc() {
	testConfig := untaggedFooConfig{
		Bar: untaggedBarConfig{
			Example:        true,
			AnotherExample: false,
			BrudiTest:      "123",
		},
	}

	assert.NoError(
		validationSuite.T(),
		config.Validate(testConfig, fooConfigValidation),
	)
}

func (validationSuite *ValidationTestSuite) TestFailValidationOnFunc() {
	testConfig := untaggedFooConfig{
		Bar: untaggedBarConfig{
			Example:        false,
			AnotherExample: false,
			BrudiTest:      "123",
		},
	}

	assert.EqualError(
		validationSuite.T(),
		config.Validate(testConfig, fooConfigValidation),
		fmt.Sprintf(
			"%s%s\n%s%s",
			"Key: 'untaggedFooConfig.example' Error:Field validation ",
			"for 'example' failed on the 'ExampleAndAnotherExampleCanNotBothBeFalse' tag",
			"Key: 'untaggedFooConfig.anotherExample' Error:Field validation ",
			"for 'anotherExample' failed on the 'ExampleAndAnotherExampleCanNotBothBeFalse' tag",
		),
	)
}

func TestValidationTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}
