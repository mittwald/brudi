package config

import "fmt"

func GenerateConfigKey(parentKey, subKey string) string {
	return fmt.Sprintf("%s.%s", parentKey, subKey)
}
