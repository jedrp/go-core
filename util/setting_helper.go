package util

import (
	"errors"
	"fmt"
	"strings"
)

// GetConfig parse string value to map[string]string object
// split config by ";" and value ":"
func GetConfig(str string) (map[string]string, error) {
	if str == "" {
		return nil, nil
	}
	config := make(map[string]string)
	splittedItemByComma := strings.Split(str, ";")
	for _, item := range splittedItemByComma {
		itemTrimmed := strings.Trim(item, " ")
		if itemTrimmed == "" {
			continue
		}
		setting := strings.Split(itemTrimmed, "=")
		key := strings.Trim(setting[0], " ")
		if key == "" {
			return nil, errors.New("key must have a value")
		}
		var value string
		if len(setting) > 1 {
			value = setting[1]
		}
		if _, found := config[key]; found {
			return nil, fmt.Errorf("duplicated key %v", key)
		}
		config[key] = value
	}
	return config, nil
}
