package config

import (
	"github.com/go-playground/validator/v10"
)

func Validate(config interface{}, f ...validator.StructLevelFunc) error {
	validate := validator.New()

	for vF := range f {
		validate.RegisterStructValidation(f[vF], config)
	}

	return validate.Struct(config)
}
