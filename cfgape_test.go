package configape_test

import (
	"fmt"
	"os"

	"github.com/zafnz/configape"
)

func ExampleApply() {
	cfg := struct {
		Foo       string `name:"foo" default:"baz" help:"This is the help for foo"`
		Bar       int    `default:"42"`
		Test      bool   `default:"true"`
		Verbosity int    `name:"verbose" short:"v" default:"0" cfgtype:"counter" help:"Verbosity level"`
	}{}
	// Fake the os arguments, here we use --no-test to override
	// the default true value of test, and we increment verbosity twice
	os.Args = []string{"test", "--foo", "bar", "--no-test", "-v", "-v"}

	configape.Apply(&cfg, nil)

	fmt.Printf("Foo: %s\n", cfg.Foo)
	fmt.Printf("Bar: %d\n", cfg.Bar)
	fmt.Printf("Test: %t\n", cfg.Test)
	fmt.Printf("Verbosity: %d\n", cfg.Verbosity)

	// Output:
	// Foo: bar
	// Bar: 42
	// Test: false
	// Verbosity: 2
}
