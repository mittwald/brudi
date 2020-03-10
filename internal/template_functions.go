package internal

import (
	"html/template"
	"reflect"
)

func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"isBool": func(i interface{}) bool {
			return reflect.ValueOf(i).Kind() == reflect.Bool
		},
		"isInt": func(i interface{}) bool {
			k := reflect.ValueOf(i).Kind()
			return k == reflect.Int ||
				k == reflect.Int8 ||
				k == reflect.Int32 ||
				k == reflect.Int64 ||
				k == reflect.Uint ||
				k == reflect.Uint8 ||
				k == reflect.Uint32 ||
				k == reflect.Uint64 ||
				k == reflect.Float32 ||
				k == reflect.Float64
		},
		"isString": func(i interface{}) bool {
			return reflect.ValueOf(i).Kind() == reflect.String
		},
		"isSlice": func(i interface{}) bool {
			return reflect.ValueOf(i).Kind() == reflect.Slice
		},
		"isArray": func(i interface{}) bool {
			return reflect.ValueOf(i).Kind() == reflect.Array
		},
		"isMap": func(i interface{}) bool {
			return reflect.ValueOf(i).Kind() == reflect.Map
		},
	}
}
