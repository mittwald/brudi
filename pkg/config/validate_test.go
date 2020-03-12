package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ValidationTestSuite struct {
	suite.Suite
}

func (validationSuite *ValidationTestSuite) TestSucceedValidationOnTags() {
	config := fooConfig{
		Bar: barConfig{
			Example:        true,
			AnotherExample: false,
			BrudiTest:      "asdf",
		},
	}

	assert.NoError(
		validationSuite.T(),
		Validate(config),
	)
}

func (validationSuite *ValidationTestSuite) TestFailValidationOnTags() {
	config := fooConfig{
		Bar: barConfig{
			Example:        true,
			AnotherExample: false,
			BrudiTest:      "",
		},
	}

	assert.EqualError(
		validationSuite.T(),
		Validate(config),
		"Key: 'fooConfig.Bar.BrudiTest' Error:Field validation for 'BrudiTest' failed on the 'min' tag",
	)
}

func (validationSuite *ValidationTestSuite) TestSucceedValidationOnFunc() {
	config := fooConfig{
		Bar: barConfig{
			Example:        true,
			AnotherExample: false,
			BrudiTest:      "123",
		},
	}

	assert.NoError(
		validationSuite.T(),
		Validate(config, fooConfigValidation),
	)
}

func (validationSuite *ValidationTestSuite) TestFailValidationOnFunc() {
	config := fooConfig{
		Bar: barConfig{
			Example:        false,
			AnotherExample: false,
			BrudiTest:      "123",
		},
	}

	assert.EqualError(
		validationSuite.T(),
		Validate(config, fooConfigValidation),
		fmt.Sprintf(
			"%s%s\n%s%s",
			"Key: 'fooConfig.example' Error:Field validation ",
			"for 'example' failed on the 'ExampleAndAnotherExampleCanNotBothBeFalse' tag",
			"Key: 'fooConfig.anotherExample' Error:Field validation ",
			"for 'anotherExample' failed on the 'ExampleAndAnotherExampleCanNotBothBeFalse' tag",
		),
	)
}

func TestValidationTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}
