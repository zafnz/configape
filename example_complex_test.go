package configape_test

import (
	"fmt"
	"os"

	"github.com/zafnz/configape"
)

// This example shows how to use configape to parse a complex configuration

// Define a struct to hold the configuration
type ComplexConfig struct {
	// The name tag is used to map the configuration to the struct
	// field. If the name tag is not specified, the field name is
	// used instead.
	// The default tag is used to specify the default value for the
	// field.
	// The help tag is used to specify the help text for the field.
	// The cfgtype tag is used to specify the type of the field.
	// The short tag is used to specify the short name for the field.
	// The env tag is used to specify the environment variable name
	// for the field.
	// The required tag is used to specify that the field must be
	// specified.

	// This can be set with --simple-string or CFG_SIMPLE_STRING
	// or "SimpleString" in the config file.
	SimpleString   string `help:"This is the help for simple string"`
	SimpleInt      int    `default:"42"`     // This will default to 42
	RequiredString string `required:"true"`  // This must be specified
	Different      string `name:"something"` // This can be set with --something or CFG_SOMETHING
	// Here this can be set with --short or -s.
	ShortFlag bool   `name:"short" short:"s" help:"This is the help for the short flag"`
	NotOnCli  string `cli:"-"` // This will not be on the cli, but is settable in the config file

	ConfigFile string `cli:"config" cfgtype:"configfile" help:"Load config from this file"`

	// Example of a sub-configuration. Here all of these settings
	// can be set with a prefix of "--database--" on the cli, or
	// as an object in the config file.
	Database struct {
		// Set with --database-host or CFG_DATABASE_HOST
		Host string `help:"Database host"`
		// Set with --database-port or CFG_DATABASE_PORT
		Port int `help:"Database port"`
	}
}

// Here we can see how to apply the configuration to the struct.
// Note that because we haven't specified a specific name, we can
// use camel case, or underscores, or dashes, for those fields.
var cfgFile = `
{
	"simple_string": "foo",
	"SimpleInt": 69,
	"required_string": "bar",
	"not_on_cli": "set in config file"
}`

func Example_complex() {
	// Write out the config as a temporary file to use
	fh, _ := os.CreateTemp("", "configape")
	defer os.Remove(fh.Name())
	fh.WriteString(cfgFile)

	os.Args = []string{"test", "--something", "set on cli", "--database-host", "dbhost", "-s", "--config", fh.Name()}

	cfg := ComplexConfig{}
	err := configape.Apply(&cfg, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("SimpleString: %s\n", cfg.SimpleString)
	fmt.Printf("SimpleInt: %d\n", cfg.SimpleInt)
	fmt.Printf("RequiredString: %s\n", cfg.RequiredString)
	fmt.Printf("Different: %s\n", cfg.Different)
	fmt.Printf("ShortFlag: %t\n", cfg.ShortFlag)
	fmt.Printf("NotOnCli: %s\n", cfg.NotOnCli)
	fmt.Printf("Database.Host: %s\n", cfg.Database.Host)

	// Output:
	// SimpleString: foo
	// SimpleInt: 69
	// RequiredString: bar
	// Different: set on cli
	// ShortFlag: true
	// NotOnCli: set in config file
	// Database.Host: dbhost

}
