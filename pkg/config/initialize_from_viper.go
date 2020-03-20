package config

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

const (
	flagTag = "viper"
)

// Viper itself is sadly not capable of resolving childs env when fetching the parent key
// Therefore we have to workaround the env-resolving
// https://github.com/spf13/viper/issues/696
// InitializeStructFromViper takes a struct (it supports pointers as well as direct values)
// It simply does a 'viper.Get()' on the config-key provided by the 'viper'-tag and loads the result into the struct property
// If no 'viper'-tags are provided by the target struct, the struct property name is used as config-key
// Use '-' to skip env-resolving for a struct property
func InitializeStructFromViper(parentKey string, target interface{}) error {
	var structElem reflect.Value

	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		structElem = reflect.ValueOf(target)
	} else {
		structElem = reflect.ValueOf(target).Elem()
	}

	for i := 0; i < structElem.NumField(); i++ {
		field := structElem.Field(i)
		flag := structElem.Type().Field(i).Tag.Get(flagTag)

		if flag == "-" {
			log.WithFields(log.Fields{
				"field": structElem.Type().Field(i).Name,
			}).Debug("skipping field")
			continue
		}

		configKey := flag
		if configKey == "" {
			configKey = structElem.Type().Field(i).Name
		}

		if parentKey != "" {
			configKey = fmt.Sprintf("%s.%s", parentKey, configKey)
		}

		if field.Kind() == reflect.Ptr {
			if field.Elem().Kind() == reflect.Struct {
				err := InitializeStructFromViper(configKey, field.Interface())
				if err != nil {
					return errors.WithStack(err)
				}
				continue
			}
			reflectSetValueFromConfigKey(configKey, field.Elem())
			continue
		}

		if field.Kind() == reflect.Struct {
			err := InitializeStructFromViper(configKey, field.Addr().Interface())
			if err != nil {
				return errors.WithStack(err)
			}
			continue
		}

		reflectSetValueFromConfigKey(configKey, field)
	}

	return nil
}

func reflectSetValueFromConfigKey(viperConfigKey string, fieldToBeSet reflect.Value) {
	viperVal := viper.Get(viperConfigKey)
	reflectedVal := reflect.ValueOf(viperVal)

	fieldLogger := log.WithFields(log.Fields{
		"viperKey":       viperConfigKey,
		"viperVal":       viperVal,
		"reflectedValue": reflectedVal,
	})

	fieldLogger.Debug("processing config key")

	if reflectedVal == reflect.Zero(reflect.TypeOf(reflectedVal)) || viperVal == nil {
		fieldLogger.Debug("skipping viper value due to non-existence")
		return
	}

	if reflectedVal.Kind() == reflect.Slice {
		sliceType := fieldToBeSet.Type()

		switch sliceType.Elem().Kind() {
		case reflect.String:
			fieldToBeSet.Set(reflect.MakeSlice(sliceType, reflectedVal.Len(), reflectedVal.Len()))
			for i := 0; i < reflectedVal.Len(); i++ {
				fieldToBeSet.Index(i).SetString(reflectedVal.Index(i).Elem().String())
			}
		default:
			fieldLogger.Debug("skipping slice since it is not a []string")
		}

		return
	}

	fieldLogger.Debug("loading viper value into struct")

	fieldToBeSet.Set(reflectedVal)
}
