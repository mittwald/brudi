package testconfig

import (
	"fmt"
	config2 "github.com/mittwald/brudi/pkg/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ValidationTestSuite struct {
	suite.Suite
}

func (validationSuite *ValidationTestSuite) TestSucceedValidationOnTags() {
	config := untaggedFooConfig{
		Bar: untaggedBarConfig{
			Example:        true,
			AnotherExample: false,
			BrudiTest:      "asdf",
		},
	}

	assert.NoError(
		validationSuite.T(),
		config2.Validate(config),
	)
}

func (validationSuite *ValidationTestSuite) TestFailValidationOnTags() {
	config := untaggedFooConfig{
		Bar: untaggedBarConfig{
			Example:        true,
			AnotherExample: false,
			BrudiTest:      "",
		},
	}

	assert.EqualError(
		validationSuite.T(),
		config2.Validate(config),
		"Key: 'untaggedFooConfig.Bar.BrudiTest' Error:Field validation for 'BrudiTest' failed on the 'min' tag",
	)
}

func (validationSuite *ValidationTestSuite) TestSucceedValidationOnFunc() {
	config := untaggedFooConfig{
		Bar: untaggedBarConfig{
			Example:        true,
			AnotherExample: false,
			BrudiTest:      "123",
		},
	}

	assert.NoError(
		validationSuite.T(),
		config2.Validate(config, fooConfigValidation),
	)
}

func (validationSuite *ValidationTestSuite) TestFailValidationOnFunc() {
	config := untaggedFooConfig{
		Bar: untaggedBarConfig{
			Example:        false,
			AnotherExample: false,
			BrudiTest:      "123",
		},
	}

	assert.EqualError(
		validationSuite.T(),
		config2.Validate(config, fooConfigValidation),
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
