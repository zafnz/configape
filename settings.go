package configape

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// enum of field types
type cfgFieldType int

const (
	fieldTypeString cfgFieldType = iota
	fieldTypeFlag
	fieldTypeList
	fieldTypeCounter
	fieldTypeConfigFile
	fieldTypeSubsection
	fieldTypeCustomMarshaler
)

// Each field in the struct is a setting (except ones that are skipped).
// Yes, this is bad in that it contains the definition and value, but we aren't
// reusing these multiple times.
type cfgSetting struct {
	idx          int    // The index of the setting in the cfg struct
	name         string // The name of the setting
	envName      string // Override for environment variable name
	cliName      string // Override the name for the cli
	shortName    string // If it is a cli, then the short name for the setting
	required     bool   // Is this required
	defaultValue string // default value
	help         string
	fieldType    cfgFieldType
	reflectType  reflect.Type // The reflect type of the setting
	subsection   cfgSettings

	// These are the results after all the parsing.
	reflectValue reflect.Value // The raw reflect value of the setting
	valueSet     bool          // Set true when the value is set.
	whereSet     string        // Describe where this value came from
}

type cfgSettings []cfgSetting

func (s cfgSettings) Find(name, what string) *cfgSetting {
	return s.doFind(name, what, false)
}

func (s cfgSettings) FindRecursive(name string, what string) *cfgSetting {
	return s.doFind(name, what, true)
}

// Find the setting with the specified name, using the what parameter to determine
// which name to use (eg if what=cli, then searches for any setting with a matching cli tag
// or otherwise the name)
func (s cfgSettings) doFind(name string, what string, doRecursive bool) *cfgSetting {
	// The name to search for could be in the form of CamelName, hyphen-name, or underscore_name
	// and hyphen and underscore could be either case.

	// First check if there is a specific cli or env name for this setting
	// If so, then use that with an explicit match -- don't lowercase or camelcase compare.
	for i := 0; i < len(s); i++ {
		//debugf("Finding %s, checking %s\n", name, s[i].name)
		var strictMatch string
		if what == "env" && s[i].envName != "" {
			strictMatch = s[i].envName
		} else if what == "cli" && s[i].cliName != "" {
			strictMatch = s[i].cliName
		}
		if strictMatch != "" {
			if name == strictMatch {
				return &s[i]
			}
			if doRecursive {
				// It could be that the first part of the name is actually a subsection.
				// If the part before a dash or hyphen matches the envName and the setting
				// is of type subsection, then recurse into the subsection.
				parts := strings.SplitN(name, "-", 2)
				if len(parts) == 1 {
					// Maybe it was on an underscore
					parts = strings.SplitN(name, "_", 2)
				}
				if len(parts) == 2 {
					section := s.doFind(parts[0], what, false)
					if section != nil && section.fieldType == fieldTypeSubsection {
						// It's a subsection, so find the setting in the subsection
						return section.subsection.doFind(parts[1], what, true)
					}
				}
			}
		}
	}
	name = strings.ToLower(name)
	// Now check just the name, checking all the possible forms
	for i := 0; i < len(s); i++ {
		underscore := strings.ToLower(camelCaseToUnderscore(s[i].name))
		underscore = strings.Replace(underscore, "-", "_", -1)
		dash := strings.Replace(underscore, "_", "-", -1)
		lowercase := strings.ToLower(s[i].name)
		//debugf("Checking name %s matches %s, %s, %s\n", name, underscore, dash, lowercase)
		if name == underscore || name == dash || name == lowercase {
			// if that field has a cliName then it should have matched that earlier.
			if what == "cli" && s[i].cliName != "" {
				continue
			} else if what == "env" && s[i].envName != "" {
				continue
			}
			return &s[i]
		}
		if doRecursive {
			// It could be that the first part of the name is actually a subsection.
			// If the part before a dash or hyphen matches the envName and the setting
			// is of type subsection, then recurse into the subsection.
			parts := strings.SplitN(name, "-", 2)
			if len(parts) == 1 {
				// Maybe it was on an underscore
				parts = strings.SplitN(name, "_", 2)
			}
			if len(parts) == 2 {
				section := s.doFind(parts[0], what, false)
				if section != nil && section.fieldType == fieldTypeSubsection {
					// It's a subsection, so find the setting in the subsection
					return section.subsection.doFind(parts[1], what, true)
				}
			}
		}
	}
	return nil
}
func (s cfgSettings) FindShort(name string) *cfgSetting {
	for i := 0; i < len(s); i++ {
		if s[i].shortName == name {
			return &s[i]
		}
	}
	return nil
}

func (s cfgSettings) FindCfgFile() *cfgSetting {
	for i := 0; i < len(s); i++ {
		if s[i].fieldType == fieldTypeConfigFile {
			return &s[i]
		}
	}
	return nil
}
func (s cfgSettings) FindRemaining() *cfgSetting {
	for i := 0; i < len(s); i++ {
		if s[i].name == "*" {
			return &s[i]
		}
	}
	return nil
}

func (s *cfgSettings) SetDefaults() error {
	var err error
	for i := 0; i < len(*s); i++ {
		setting := &(*s)[i]
		if setting.defaultValue != "" {
			setting.whereSet = fmt.Sprintf("%s default value", setting.name)
			setting.reflectValue, err = strToType(setting.reflectType, setting.defaultValue)
			setting.valueSet = true
			if err != nil {
				return fmt.Errorf("failed to parse default value for %s: %s", setting.name, err)
			}
		}
		if setting.fieldType == fieldTypeSubsection {
			err := setting.subsection.SetDefaults()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (s cfgSettings) CheckRequired() error {
	for i := 0; i < len(s); i++ {
		setting := &s[i]
		if setting.required && !setting.valueSet {
			return fmt.Errorf("required setting %s not set", setting.name)
		}
		if setting.fieldType == fieldTypeSubsection {
			err := setting.subsection.CheckRequired()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Reads the cfg struct and creates the settings that represents the struct
func (c *cfgApe) parseStructIntoSettings() error {
	typeOfCfg := reflect.TypeOf(c.cfg)
	if typeOfCfg.Kind() == reflect.Ptr {
		typeOfCfg = typeOfCfg.Elem()
	}
	if typeOfCfg.Kind() != reflect.Struct {
		return fmt.Errorf("cfg must be a struct")
	}
	var err error
	c.settings, err = structToSettings(typeOfCfg)
	if err != nil {
		return err
	}
	return nil
}

func structToSettings(cfgType reflect.Type) (cfgSettings, error) {
	// Loop through all the fields in the cfg struct
	// and extract out the commandline parameters
	// and their values.
	//
	var settings cfgSettings
	jsonUnmarshaler := reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	yamlUnmarshaler := reflect.TypeOf((*yaml.Unmarshaler)(nil)).Elem()
	textUnmarshaler := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

	for i := 0; i < cfgType.NumField(); i++ {
		field := cfgType.Field(i)
		var setting cfgSetting
		setting.idx = i

		// If the field is a private field and inaccessible, skip it
		if field.PkgPath != "" {
			continue
		}

		if cfgTag := field.Tag.Get("name"); cfgTag != "" {
			if cfgTag == "-" {
				continue // Skip this field
			}
			setting.name = cfgTag
		} else {
			setting.name = field.Name
		}
		setting.reflectType = field.Type

		if help := field.Tag.Get("help"); help != "" {
			setting.help = help
		}
		if required := field.Tag.Get("required"); required != "" {
			setting.required = true
		}
		if defaultVal := field.Tag.Get("default"); defaultVal != "" {
			setting.defaultValue = defaultVal
		}
		if envName := field.Tag.Get("env"); envName != "" {
			setting.envName = envName
		}
		if shortName := field.Tag.Get("short"); shortName != "" {
			if len(shortName) != 1 {
				return nil, fmt.Errorf("struct field %s, short name must be a single character", field.Name)
			}
			setting.shortName = shortName
		}
		if cliName := field.Tag.Get("cli"); cliName != "" {
			setting.cliName = cliName
		}

		// make ptrType the type of a pointer to fieldType
		// (As Unmarshal needs a pointer to the type)
		ptrType := reflect.PtrTo(field.Type)

		// If field implements any of the unmarshallers, then it's not a subsection.
		if field.Tag.Get("cfgtype") != "subsection" && (ptrType.Implements(jsonUnmarshaler) || ptrType.Implements(yamlUnmarshaler) || ptrType.Implements(textUnmarshaler)) {
			// We need to detect and flag if the field has a custom unmarshaler, as we can't recurse into it, like we do
			// for the structs.
			setting.fieldType = fieldTypeCustomMarshaler
		} else if field.Type.Kind() == reflect.Struct {
			// if the field type is a struct
			// this is a subsection
			subsettings, err := structToSettings(field.Type)
			if err != nil {
				return nil, err
			}
			// If the setting.name has an uppercase letter other than the first letter,
			// print a warning
			if len(setting.name) > 1 && setting.name[1:] != strings.ToLower(setting.name[1:]) {
				fmt.Printf("WARNING: struct field %s, subsection name %s should be all lowercase, use name tag to rename\n", field.Name, setting.name)
			}
			// If the setting.name has a hyphen or underscore, then strip it out.
			if strings.Contains(setting.name, "-") || strings.Contains(setting.name, "_") {
				fmt.Printf("WARNING: struct field %s, subsection name %s should not contain hyphens or underscores, use name tag to rename\n", field.Name, setting.name)
				// Remove all hyphens and underscores
				setting.name = strings.ReplaceAll(setting.name, "-", "")
				setting.name = strings.ReplaceAll(setting.name, "_", "")
			}
			// Remove camel casing, because subsections can't have hyphens or underscores
			setting.name = strings.ToLower(setting.name)

			setting.subsection = subsettings
			setting.fieldType = fieldTypeSubsection
		} else if field.Type.Kind() == reflect.Bool {
			setting.fieldType = fieldTypeFlag
		} else if field.Type.Kind() == reflect.Slice {
			setting.fieldType = fieldTypeList
		} else if fieldType := field.Tag.Get("cfgtype"); fieldType != "" {
			switch fieldType {
			case "configfile":
				setting.fieldType = fieldTypeConfigFile
			case "counter":
				setting.fieldType = fieldTypeCounter
			default:
				return nil, fmt.Errorf("struct field %s, unknown field type: %s", field.Name, fieldType)
			}
		}
		settings = append(settings, setting)
	}

	return settings, nil
}
