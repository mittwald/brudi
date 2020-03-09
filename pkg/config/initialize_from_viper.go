package config

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"reflect"
)

// Viper itself is sadly not capable of resolving childs env when fetching the parent key
// Therefore we have to workaround the env-resolving
// https://github.com/spf13/viper/issues/696
// InitializeStructFromViper takes a pointer to a struct
// It simply does a 'viper.Get()' on the config-key provided by the 'viper'-tag and loads the result into the struct property
// If no 'viper'-tags are provided by the target struct, the struct property name is used as config-key
// Use '-' to skip env-resolving for a struct property
// For now, there is no support for nested structs or pointer
func InitializeStructFromViper(parentKey string, target interface{}) error {
	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		return errors.WithStack(fmt.Errorf("target is not a pointer to a struct"))
	}

	v := reflect.ValueOf(target).Elem()

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("viper")
		if tag == "-" {
			log.WithFields(log.Fields{
				"field": v.Type().Field(i).Name,
			}).Debug("skipping field")
			continue
		}

		var configKey string

		if tag == "" {
			configKey = v.Type().Field(i).Name
		}

		if parentKey != "" {
			configKey = fmt.Sprintf("%s.%s", parentKey, configKey)
		}

		viperVal := viper.Get(configKey)
		reflectedVal := reflect.ValueOf(viperVal)

		log.WithFields(log.Fields{
			"viperKey":       configKey,
			"viperVal":       viperVal,
			"reflectedValue": reflectedVal,
		}).Debug("loading viper value into struct")

		v.Field(i).Set(reflectedVal)
	}

	return nil
}
