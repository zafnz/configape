package configape

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// type exampleConfig struct {
// 	Backend       string   `cfg:"backend" required:"true" default:"sqlite"`
// 	Path          string   `cfg:"path" help:"Path to the database file."`
// 	ListOfStrings []string `cfg:"list_of_strings" help:"List of strings."`
// 	Flag          bool     `cfg:"flag" help:"A flag."`
// 	Verbose       int      `cfg:"verbose" type:"counter"`
// 	CfgFile       string   `cfg:"config" type:"configfile" help:"Path to the configuration file."`
// 	Extra         []string `cfg:"*"`
// 	Section       struct {
// 		SectionString string `cfg:"section_string" help:"A string in a section."`
// 	} `cfg:"section"`
// }

// Options on how Config Ape should work.
type Options struct {
	ConfigFilename         string // Name of the config file to use
	ConfigFileType         string // The file type, defaults to json and determines the file extension.
	EnvironmentPrefix      string // Prefix for environment variables, empty string defaults to CFG_, if you really want no prefix, set to ! (not recommended)
	UseSingleDashArguments bool   // If set, then arguments are expected as "-foo bar" instead of "--foo bar" (not recommended)

	Help                         func(str string) // If set, then this function will be called when the help flag is set.
	HelpHeader                   string           // Help text that is prefixed to the help output.
	HelpFooter                   string           // Help text that is appended to the help output.
	HelpWriter                   io.Writer        // Where to write the help output, defaults to os.Stderr
	DisableHelpOnMissingRequired bool             // If set, then the help will not be printed if a required setting is missing.
	DisableHelp                  bool             // Disable the help flag
	DisableVersion               bool             // Disable the version flag

	Name    string // Name of the program, used in the help output. Defaults to os.Args[0]
	Version string // Version of the program, used in the help output. Defaults to "v0.0.0"

	DisableEnviornment bool // Disable environment variables
	DisableConfigFile  bool // Disable config file parsing
	DisableCommandLine bool // Disable command line parsing

	AllowUnknownConfigFileKeys bool // If set, then unknown keys in the config file will not cause an error.

	// For testing
	cfgFileContents string
	osArgs          []string
}

type cfgApe struct {
	options   Options
	cfg       interface{}
	settings  cfgSettings
	remaining []string // The remaining non option arguments.
}

// Apply the configuration to the provided cfg struct, using the options provided.
func Apply(cfg interface{}, options *Options) error {
	c := cfgApe{}
	return c.Apply(cfg, options)
}

func (c *cfgApe) Apply(cfg interface{}, options *Options) error {
	if options == nil {
		options = &Options{}
	}
	if reflect.TypeOf(cfg).Kind() != reflect.Ptr {
		return fmt.Errorf("cfg must be a pointer")
	}
	c.cfg = cfg
	c.options = *options

	// First we need to extract out the commandline parameters
	err := c.parseStructIntoSettings()
	if err != nil {
		return err
	}
	defaultCfgFile := c.options.ConfigFilename
	if defaultCfgFile == "" {
		defaultCfgFile = "config.json"
	}

	// See if there is a config file specified on the command line
	if !c.options.DisableCommandLine {
		args := os.Args
		if c.options.osArgs != nil {
			args = c.options.osArgs
		}
		file, err := c.getCliConfigFile(args)
		if err != nil {
			return err
		}
		if file != "" {
			defaultCfgFile = file
		}
	}

	// Set defaults
	err = c.settings.SetDefaults()
	if err != nil {
		return fmt.Errorf("error setting defaults: %s", err)
	}
	// fmt.Println("After defaults")
	// debugf("%+v\n", c.settings)

	if !c.options.DisableConfigFile {
		// read the config file first.
		err = c.parseConfigFile(defaultCfgFile)
		if err != nil {
			return err
		}
	}
	// fmt.Println("After Config File")
	// debugf("%+v\n", c.settings)

	if !c.options.DisableEnviornment {
		// Now we need to parse the environment variables
		err = c.processEnvironment()
		if err != nil {
			return err
		}
	}
	// fmt.Println("After Environment")
	// debugf("%+v\n", c.settings)

	// Now we need to parse the command line
	if !c.options.DisableCommandLine {
		if c.options.osArgs == nil {
			err = c.parseCommandLine(os.Args)
		} else {
			err = c.parseCommandLine(c.options.osArgs)
		}
		if err != nil {
			return err
		}
		if len(c.remaining) > 0 {
			setting := c.settings.FindRemaining()
			if setting != nil {
				var err error
				setting.reflectValue, err = strListToType(setting.reflectType, c.remaining)
				if err != nil {
					return fmt.Errorf("failed to parse remaining arguments into cfg.%s: %s", setting.name, err)
				}
				setting.valueSet = true
			}
		}
	}
	// fmt.Println("After Commandline")
	// debugf("%+v\n", c.settings)

	// Check if all required settings are set
	err = c.settings.CheckRequired()
	if err != nil {
		if !c.options.DisableHelpOnMissingRequired {
			c.printHelp()
		}
		return err
	}
	// Now we have parsed all the settings from file and commandline
	// so we can set the values in the cfg struct
	err = setValues(cfg, c.settings)
	return err
}

// Split a comma separated name=value string into a map. The value can be quoted with single or double
// quotes and the comma inside the quotes will not be used as a separator, and the quotes will be
// stripped from the value.
func stringListToMap(str string) map[string]string {
	m := make(map[string]string)

	for _, s := range splitPreservingQuotes(str) {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) == 2 {
			val := parts[1]
			// Strip surrounding single or double quotes
			if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
				val = val[1 : len(val)-1]
			}
			m[parts[0]] = val
		} else {
			m[parts[0]] = ""
		}
	}
	return m
}

// Splits the string on commans, but preserves any quoted strings
func splitPreservingQuotes(str string) []string {
	var result []string
	var current string
	inQuotes := false

	for _, c := range str {
		if c == '"' || c == '\'' {
			inQuotes = !inQuotes
		} else if c == ',' && !inQuotes {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

func setValues(cfg interface{}, settings cfgSettings) error {
	// Loop through the settings and set the values in the cfg struct
	typeOfCfg := reflect.TypeOf(cfg)
	valueOfCfg := reflect.ValueOf(cfg)
	if typeOfCfg.Kind() == reflect.Ptr {
		valueOfCfg = valueOfCfg.Elem()
	}
	for _, setting := range settings {
		// Find the field in the cfg struct
		value := valueOfCfg.FieldByIndex([]int{setting.idx})
		if !value.IsValid() {
			return fmt.Errorf("invalid setting: %s", setting.name)
		}
		// debugf("Setting %s to %s\n", setting.name, setting.value)
		// debugf("field type: %s\n", field.Type)
		// debugf("field kind: %s\n", field.Type.Kind())
		// debugf("field current value: %v\n", value)
		// If the setting is a subsection, then we need to recurse
		if setting.fieldType == fieldTypeSubsection {
			err := setValues(value.Addr().Interface(), setting.subsection)
			if err != nil {
				return err
			}
			continue
		}
		// If there is a reflectValue set (valueSet) then just use that.
		if setting.valueSet {
			//debugf("field %s is set to %v\n", setting.name, setting.reflectValue)
			if setting.reflectType.Kind() == reflect.Ptr && setting.reflectValue.Kind() != reflect.Ptr {
				// If the setting is a pointer, then we need to set the value to the pointer
				vPtr := reflect.New(setting.reflectType.Elem())
				vPtr.Elem().Set(setting.reflectValue)
				//value.Set(setting.reflectValue.Addr())
				value.Set(vPtr)
			} else {
				//debugf("setting %s type is %s, setting value type is %s, setting value is %+v\n", setting.name, setting.reflectType.Kind(), setting.reflectValue.Type().Kind(), setting.reflectValue)
				value.Set(setting.reflectValue)
			}
		}

	}

	return nil
}
