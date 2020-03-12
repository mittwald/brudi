package config

import (
	"github.com/go-playground/validator/v10"
)

type fooConfig struct {
	Bar barConfig
}

type barConfig struct {
	Example        bool
	AnotherExample bool
	BrudiTest      string `validate:"min=1"`
}

func fooConfigValidation(sl validator.StructLevel) {
	c := sl.Current().Interface().(fooConfig)

	if !c.Bar.Example && !c.Bar.AnotherExample {
		sl.ReportError(c.Bar.Example, "example", "Example", "ExampleAndAnotherExampleCanNotBothBeFalse", "")
		sl.ReportError(c.Bar.AnotherExample, "anotherExample", "AnotherExample", "ExampleAndAnotherExampleCanNotBothBeFalse", "")
	}
}
