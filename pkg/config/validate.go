package config

import (
	"github.com/go-playground/validator/v10"
)

// Validate() takes a struct and n validator.StructLevelFunc
// The given struct is validated by the validator package
// If you pass nested structs to this function, keep in mind that the provided validatorFuncs are executed on the parent struct
func Validate(config interface{}, f ...validator.StructLevelFunc) error {
	validate := validator.New()

	for vF := range f {
		validate.RegisterStructValidation(f[vF], config)
	}

	return validate.Struct(config)
}
