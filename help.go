package configape

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Returns a string suitable as the output for the help command.
// You can override the default output using the HelpWriter or Help function in the Options struct.
// Or you can disable help entirely with DisableHelp
func Help(cfg interface{}, options *Options) (string, error) {
	c := cfgApe{}
	c.cfg = cfg
	if options == nil {
		options = &Options{}
	}
	c.options = *options

	// First we need to extract out the commandline parameters
	err := c.parseStructIntoSettings()
	if err != nil {
		return "", err
	}
	return c.makeHelp(), nil
}

func (c *cfgApe) printHelp() {
	if c.options.Help != nil {
		c.options.Help(c.makeHelp())
		return
	}
	var fh io.Writer
	fh = os.Stderr
	if c.options.HelpWriter != nil {
		fh = c.options.HelpWriter
	}
	fmt.Fprint(fh, c.makeHelp())
}

func (c *cfgApe) printVersion() {
	var fh io.Writer
	fh = os.Stderr
	if c.options.HelpWriter != nil {
		fh = c.options.HelpWriter
	}

	if c.options.Version != "" {
		fmt.Fprintf(fh, "%s (v%s)\n", c.options.Name, c.options.Version)
	} else if c.options.Name != "" {
		fmt.Fprintf(fh, "%s\n", c.options.Name)
	} else {
		fmt.Fprintf(fh, "%s\n", os.Args[0])
	}
}

func (c *cfgApe) makeHelp() string {
	result := ""
	if c.options.Name == "" {
		c.options.Name = os.Args[0]
	}
	if c.options.Version == "" {
		c.options.Version = "0.0.0"
	}
	result += fmt.Sprintf("%s (v%s)\n\n", c.options.Name, c.options.Version)

	if c.options.HelpHeader != "" {
		result += fmt.Sprintf("%s\n\n", c.options.HelpHeader)
	}

	result += makeHelp(c.settings, "")

	if c.options.HelpFooter != "" {
		result += fmt.Sprintf("\n%s\n", c.options.HelpFooter)
	}

	return result
}

func makeHelp(settings cfgSettings, prefix string) string {
	result := ""

	subsections := cfgSettings{}
	for _, setting := range settings {
		if setting.fieldType == fieldTypeSubsection {
			subsections = append(subsections, setting)
			continue
		}
		name := strings.ToLower(camelCaseToDash(setting.name))
		if setting.cliName == "-" {
			continue
		} else if setting.cliName != "" {
			name = setting.cliName
		}

		result += fmt.Sprintf("  --%s%s", prefix, name)
		if setting.fieldType != fieldTypeFlag && setting.fieldType != fieldTypeCounter {
			result += fmt.Sprintf(" <%s>", setting.name)
		}
		if setting.shortName != "" {
			result += fmt.Sprintf(", -%s", setting.shortName)
			if setting.fieldType != fieldTypeFlag && setting.fieldType != fieldTypeCounter {
				result += fmt.Sprintf(" <%s>", setting.name)
			}
		}
		if setting.defaultValue != "" {
			result += fmt.Sprintf(" (default: %s)", setting.defaultValue)
		}
		if setting.help != "" {
			result += fmt.Sprintf("\n    %s\n", setting.help)
		} else {
			result += "\n"
		}
	}
	// Now do the subsections
	for _, setting := range subsections {
		result += fmt.Sprintf("\n%s\n", setting.name)
		if setting.help != "" {
			result += fmt.Sprintf("  %s\n", setting.help)
		}
		result += "\n"
		name := strings.ToLower(camelCaseToDash(setting.name))
		result += makeHelp(setting.subsection, fmt.Sprintf("%s%s-", prefix, name))
	}
	return result
}
