package configape

import (
	"fmt"
	"io"
	"reflect"

	"gopkg.in/yaml.v3"
)

func (c *cfgApe) parseYamlConfigFile(cfgFile string, fh io.Reader) error {
	yamlCfg := make(map[string]yaml.Node)
	decoder := yaml.NewDecoder(fh)
	err := decoder.Decode(&yamlCfg)
	if err != nil {
		return fmt.Errorf("error parsing config file %s: %s", cfgFile, err)
	}
	err = c.parseYamlMap(c.settings, yamlCfg)
	if err != nil {
		return fmt.Errorf("error parsing config file %s: %s", cfgFile, err)
	}
	return nil
}

func (c *cfgApe) parseYamlMap(settings cfgSettings, m map[string]yaml.Node) error {
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
			subsection := make(map[string]yaml.Node)
			err := value.Decode(&subsection)
			if err != nil {
				return err
			}
			err = c.parseYamlMap(setting.subsection, subsection)
			if err != nil {
				return err
			}
			continue
		}
		v := reflect.New(setting.reflectType)
		// Unmarshal the value into the pointer
		err := value.Decode(v.Interface())
		if err != nil {
			return err
		}
		//debugf("%s: v is type %s, val %+v\n", key, reflect.TypeOf(v), v.Elem())
		setting.reflectValue = v.Elem()
		setting.valueSet = true
	}
	return nil
}
