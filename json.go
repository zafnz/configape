package configape

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

// A function that takes a string, and an offset, and returns the line number
// and column number of the offset
func offsetToLineColumn(str string, offset int64) (int, int) {
	line := 1
	column := 1
	for i := int64(0); i < offset; i++ {
		if str[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}

func (c *cfgApe) parseJsonConfigFile(cfgFile string, fh io.Reader) error {
	jsonCfg := make(map[string]json.RawMessage)
	decoder := json.NewDecoder(fh)
	err := decoder.Decode(&jsonCfg)
	if err == nil {
		err = c.parseJsonMap(c.settings, jsonCfg)
	}
	if err != nil {
		var offset int
		var msg string = err.Error()
		// If the error is of type json.SyntaxError, then we can give a better error message
		if syntaxError, ok := err.(*json.SyntaxError); ok {
			offset = int(syntaxError.Offset)
		}
		// If the error is of type json.UnmarshalTypeError, then we can give a better error message
		if typeError, ok := err.(*json.UnmarshalTypeError); ok {
			offset = int(typeError.Offset)
			msg = fmt.Sprintf("expected type %s, got %s", typeError.Type, typeError.Value)
		}
		if offset > 0 {
			line, column := offsetToLineColumn(c.options.cfgFileContents, int64(offset))
			msg = fmt.Sprintf("%s at line %d, column %d", msg, line, column)
		}
		return fmt.Errorf("error parsing config file %s: %s", cfgFile, msg)
	}
	return nil
}

func (c *cfgApe) parseJsonMap(settings cfgSettings, m map[string]json.RawMessage) error {
	for key, value := range m {
		setting := settings.Find(key, "config")
		if setting == nil {
			if c.options.AllowUnknownConfigFileKeys {
				continue
			}
			return fmt.Errorf("unknown setting in config file: %s", key)
		}
		// If we've decided it's a subsection, recurse into it.
		if setting.fieldType == fieldTypeSubsection {
			subsection := make(map[string]json.RawMessage)
			err := json.Unmarshal(value, &subsection)
			if err != nil {
				return err
			}
			err = c.parseJsonMap(setting.subsection, subsection)
			if err != nil {
				return err
			}
			continue
		}
		//debugf("key %s is type %s\n", key, setting.reflectType)
		// Just use json.Unmarshal to unmarshal the value into the reflectType
		// of the setting
		// Create a pointer to the setting.reflectType
		v := reflect.New(setting.reflectType)
		// Unmarshal the value into the pointer
		err := json.Unmarshal(value, v.Interface())
		if err != nil {
			return err
		}
		//debugf("%s: v is type %s, val %+v\n", key, reflect.TypeOf(v), v.Elem())
		setting.reflectValue = v.Elem()
		setting.valueSet = true
	}

	return nil
}
