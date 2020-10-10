package config

import (
	"github.com/go-playground/validator/v10"

	"github.com/mittwald/brudi/internal"
)

func init() {
	internal.InitLogger()
}

type untaggedFooConfig struct {
	Bar untaggedBarConfig
}

type untaggedBarConfig struct {
	Example        bool
	AnotherExample bool
	BrudiTest      string `validate:"min=1"`
	BrudiNumber    int
}

type taggedFooConfig struct {
	CustomBar *taggedBarConfig `viper:"bar"`
}

type taggedBarConfig struct {
	CustomExample        bool   `viper:"example"`
	CustomAnotherExample bool   `viper:"anotherExample"`
	CustomBrudiTest      string `viper:"brudiTest" validate:"min=1"`
	CustomBrudiNumber    int    `viper:"brudiNumber"`
}

func fooConfigValidation(sl validator.StructLevel) {
	c := sl.Current().Interface().(untaggedFooConfig)

	if !c.Bar.Example && !c.Bar.AnotherExample {
		sl.ReportError(c.Bar.Example, "example", "Example", "ExampleAndAnotherExampleCanNotBothBeFalse", "")
		sl.ReportError(c.Bar.AnotherExample, "anotherExample", "AnotherExample", "ExampleAndAnotherExampleCanNotBothBeFalse", "")
	}
}
