package configape

import (
	"fmt"
	"strings"
)

// Searches the osArgs looking for the specified config file argument. This
// argument is specified in the cfg structure as having the tag cfgtype of "cfgfile".
func (c *cfgApe) getCliConfigFile(osArgs []string) (string, error) {
	// Get the name of the config file argument, if any
	setting := c.settings.FindCfgFile()
	if setting == nil {
		// There is none
		return "", nil
	}

	configArgName := strings.ToLower(setting.name)
	configArgHyphen := strings.ToLower(camelCaseToDash(setting.name))
	configArgUnderscore := strings.Replace(configArgHyphen, "-", "_", -1)

	for idx := 1; idx < len(osArgs); idx++ {
		arg := osArgs[idx]
		if !strings.HasPrefix(arg, "--") {
			continue
		}
		if arg == "--" {
			// Everything after this is a remainingArg
			break
		}
		// It's an argument
		// Strip off the --
		arg = strings.ToLower(arg[2:])
		if setting.cliName != "" {
			if arg == setting.cliName {
				// Grab the next argument
				if idx >= len(osArgs) {
					return "", fmt.Errorf("missing value for argument: %s", arg)
				}
				idx++
				return osArgs[idx], nil
			}
		} else if arg == configArgName || arg == configArgHyphen || arg == configArgUnderscore {
			// Grab the next argument
			if idx >= len(osArgs) {
				return "", fmt.Errorf("missing value for argument: %s", arg)
			}
			idx++
			return osArgs[idx], nil
		}
	}
	return "", nil
}

func stringPtr(s string) *string {
	// I hate that this has to exist. I hate that I have to use it.
	return &s
}

func (c *cfgApe) parseCommandLine(osArgs []string) error {
	// Loop while osArgs has something in it
	max := 50
	// Pop the program name off the stack
	osArgs = osArgs[1:]
	for len(osArgs) > 0 {
		// DEBUG
		max -= 1
		if max == 0 {
			debugf("Something is wrong, max iterations exceeded\n")
			return fmt.Errorf("too many iterations")
		}

		var arg string
		arg, osArgs = osArgs[0], osArgs[1:]
		what := arg
		if arg == "--" {
			// Everything after this is a remainingArg
			c.remaining = append(c.remaining, osArgs...)
			break
		}
		//debugf("arg: %s\n", arg)

		if strings.HasPrefix(arg, "--") {
			// Longform argument
			arg = arg[2:]
			// We may have a value after the equals or if it's a --no- prefix
			var forceValue *string
			// If there is an equals in the name, then that's the value, so arg becomes everything
			// before the equals, and value is the rest.
			equalIdx := strings.Index(arg, "=")
			if equalIdx != -1 {
				value := arg[equalIdx+1:]
				arg = arg[:equalIdx]
				forceValue = &value
			}
			// Find the setting
			setting := c.settings.FindRecursive(arg, "cli")
			if setting == nil && strings.HasPrefix(arg, "no-") {
				// We didn't find an existing field with the name `no-foo`, so let's try `foo`
				// If it's a no- then strip that off and find the setting
				setting = c.settings.FindRecursive(arg[3:], "cli")
				if setting != nil {
					forceValue = stringPtr("false")
				}
			}
			if setting == nil {
				// if they asked for help, then spit it out
				if arg == "help" && !c.options.DisableHelp {
					c.printHelp()
					return nil
				}
				if arg == "version" && !c.options.DisableVersion {
					c.printVersion()
					return nil
				}
				return fmt.Errorf("unknown command line argument: %s", what)
			}
			err := c.setSetting(setting, forceValue, what, &osArgs)
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(arg, "-") {
			// Shortform argument
			args := arg[1:]
			for _, a := range args {
				arg := string(a)
				setting := c.settings.FindShort(arg)
				err := c.setSetting(setting, nil, what, &osArgs)
				if err != nil {
					return err
				}
			}
		} else {
			c.remaining = append(c.remaining, arg)
			continue
		}
	}
	return nil
}

func (c *cfgApe) setSetting(setting *cfgSetting, forceValue *string, whereFrom string, args *[]string) error {
	// Find the setting that has the name "arg"
	var err error

	if setting == nil {
		return fmt.Errorf("unknown command line argument: %s", whereFrom)
	}
	var value string
	setting.whereSet = whereFrom
	//debugf("Setting %s, forceValue: %v, whereFrom: %s, args: %v\n", setting.name, forceValue, whereFrom, *args)

	// If it's a boolean, then set it to true
	if setting.fieldType == fieldTypeFlag {
		if forceValue != nil {
			value = *forceValue
		} else {
			value = "true"
		}
	} else if setting.fieldType == fieldTypeCounter {
		setting.reflectValue, err = incrementNumber(setting.reflectType, setting.reflectValue, 1)
		if err != nil {
			return fmt.Errorf("failed to parse %s into cfg.%s: %s", whereFrom, setting.name, err)
		}
		setting.valueSet = true
		return nil
	} else {
		if forceValue == nil {
			if len(*args) == 0 {
				return fmt.Errorf("missing value for argument: %s", whereFrom)
			}
			value = (*args)[0]
			*args = (*args)[1:]
		} else {
			value = *forceValue
		}
	}

	if setting.fieldType == fieldTypeList {
		//setting.values = append(setting.values, value)
		setting.reflectValue, err = appendStrToListType(setting.reflectType, setting.reflectValue, value)
	} else {
		setting.reflectValue, err = strToType(setting.reflectType, value)
	}
	if err != nil {
		return fmt.Errorf("failed to parse %s=%s into cfg.%s: %s", whereFrom, value, setting.name, err)
	}
	setting.valueSet = true
	return nil
}
