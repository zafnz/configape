package configape_test

import (
	"fmt"

	"github.com/zafnz/configape"
)

// This example shows how to use configape to parse a complex configuration

// Define a struct to hold the configuration
type HelpExampleConfig struct {
	// This is the same as the complex example, see that for an explanation
	// of all the tags.
	SimpleString   string `help:"This is the help for simple string"`
	SimpleInt      int    `default:"42"`
	RequiredString string `required:"true"`
	Different      string `name:"something"`
	ShortFlag      bool   `name:"short" short:"s" help:"This is the help for the short flag"`
	NotOnCli       string `cli:"-"`

	ConfigFile string `cli:"config" cfgtype:"configfile" help:"Load config from this file"`

	Database struct {
		// Set with --database-host or CFG_DATABASE_HOST
		Host string `help:"Database host to connect to"`
		// Set with --database-port or CFG_DATABASE_PORT
		Port int `help:"Database port to use with the database host"`
	}
}

func Example_help() {
	cfg := HelpExampleConfig{}
	help, _ := configape.Help(cfg, &configape.Options{Name: "example"})
	fmt.Println(help)

	// Output:
	// example (v0.0.0)
	//
	//   --simple-string <SimpleString>
	//     This is the help for simple string
	//   --simple-int <SimpleInt> (default: 42)
	//   --required-string <RequiredString>
	//   --something <something>
	//   --short, -s
	//     This is the help for the short flag
	//   --config <ConfigFile>
	//     Load config from this file
	//
	// database
	//
	//   --database-host <Host>
	//     Database host to connect to
	//   --database-port <Port>
	//     Database port to use with the database host
}
