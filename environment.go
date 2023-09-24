package configape

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

func (c *cfgApe) processEnvironment() error {
	prefix := c.options.EnvironmentPrefix
	if prefix == "" {
		prefix = "CFG_"
	}
	if prefix == "!" {
		prefix = ""
	}

	env := os.Environ()
	for _, envVar := range env {
		parts := strings.SplitN(envVar, "=", 2)
		val := ""
		originalName := parts[0]
		if !strings.HasPrefix(originalName, prefix) {
			continue
		}
		if len(parts) == 2 {
			val = parts[1]
		}

		name := strings.TrimPrefix(originalName, prefix)
		name = strings.ToLower(name)

		setting := c.settings.FindRecursive(name, "env")
		if setting == nil {
			continue
		}
		// Boolean here is handled as if the environment variable is set to empty, or
		// if its value is "true" or "1", then it's true, otherwise it's false
		if setting.reflectType.Kind() == reflect.Bool || setting.fieldType == fieldTypeFlag {
			if val == "true" || val == "1" || val == "" {
				val = "true"
			} else {
				val = "false"
			}
		}

		var err error
		if setting.fieldType == fieldTypeList {
			values := strings.Split(val, ",")
			setting.reflectValue, err = strListToType(setting.reflectType, values)
		} else {
			// debugf("Setting %s to %s\n", setting.name, val)
			setting.reflectValue, err = strToType(setting.reflectType, val)
		}
		setting.valueSet = true
		if err != nil {
			return fmt.Errorf("failed to parse environment %s into cfg.%s: %s", originalName, setting.name, err)
		}
	}
	return nil
}
