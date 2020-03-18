package config

import (
	"fmt"
	"os"
	"reflect"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func EnsureEnv(config interface{}) error {
	if reflect.ValueOf(config).Kind() == reflect.Ptr {
		return errors.WithStack(fmt.Errorf("can not reflect value of pointer"))
	}

	v := reflect.ValueOf(config)

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("env")

		if tag == "" {
			continue
		}

		if os.Getenv(tag) == "" {
			err := os.Setenv(tag, v.Field(i).String())
			if err != nil {
				return err
			}
			log.WithFields(log.Fields{
				"variable": tag,
				"value":    v.Field(i).String(),
			}).Debug("setting environment variable")
		}
	}

	return nil
}
