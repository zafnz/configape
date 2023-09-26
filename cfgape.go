// A unified configuration/environment/cli parser for Go, with a focus on simplicity and ease of use.
package configape

import (
	"fmt"
	"io"
	"os"
	"reflect"
)

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

// Internal state holder.
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
