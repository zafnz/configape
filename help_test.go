package configape_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/zafnz/configape"
)

type testConfig struct {
	Foo       string  `name:"foo" default:"baz" help:"This is the help for foo"`
	Bar       *string `help:"This is the help for bar"`
	Flag      bool    `help:"This is the help for flag"`
	Empty     string  `cfg:"-"`
	Number    int     `help:"This is the help for number" default:"42"`
	List      []string
	Counter   int `cfgtype:"counter" help:"This is the help for counter"`
	Interface interface{}
	CamelTest string `help:"CamelCase test"`
}

func TestHelp(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()
	// Create an io.Writer that we can write to
	buffer := bytes.NewBuffer([]byte{})

	cfg := testConfig{}
	os.Args = []string{"test", "--help"}

	err := configape.Apply(&cfg, &configape.Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		DisableCommandLine: false,
		HelpWriter:         buffer,
	})
	if err != nil {
		t.Fatal(err)
	}

	output := buffer.String()
	// The output should contain things like:
	// --camel-test
	// --counter
	// --number (default: 42)
	// --flag
	if !strings.Contains(output, "--camel-test") {
		t.Error("output did not contain --camel-test")
	}
	if !strings.Contains(output, "--counter") {
		t.Error("output did not contain --counter")
	}
	if !strings.Contains(output, "--number <Number> (default: 42)") {
		fmt.Println(output)

		t.Error("output did not contain --number <Number> (default: 42)")
	}
	if !strings.Contains(output, "--flag") {
		t.Error("output did not contain --flag")
	}
}

func ExampleHelp() {
	cfg := struct {
		Foo string `name:"foo" default:"baz" help:"This is the help for foo"`
	}{}
	options := configape.Options{
		HelpWriter: os.Stdout, // By default goes to Stderr
		Name:       "my-prog",
		Version:    "1.2.3",
	}
	os.Args = []string{"test", "--help"}

	configape.Apply(&cfg, &options)
	// Output:
	// my-prog (v1.2.3)
	//
	//   --foo <foo> (default: baz)
	//     This is the help for foo
}

func TestHelpComplex(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()
	// Create an io.Writer that we can write to
	buffer := bytes.NewBuffer([]byte{})

	cfg := struct {
		Foo     string
		Section struct {
			Test        string
			AnotherTest string `help:"This is another test"`
		}
		AnotherSection struct {
			MoreTests  int `default:"42"`
			SubSection struct {
				SubTest       string `help:"This is a sub test"`
				MoreMoreTests int    `default:"42"`
			} `name:"namedsubsection" help:"This is a named section"`
		} `name:"anothersection"`
	}{}

	os.Args = []string{"test", "--help"}

	err := configape.Apply(&cfg, &configape.Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		DisableCommandLine: false,
		HelpWriter:         buffer,
	})
	if err != nil {
		t.Fatal(err)
	}

	output := buffer.String()
	// The output should contain things like:
	// --section-test
	// --section-another-test
	// --anothersection-more-tests (default: 42)
	if !strings.Contains(output, "--section-test") {
		fmt.Println(output)
		t.Error("output did not contain --section-test")
	}
	if !strings.Contains(output, "--section-another-test") {
		fmt.Println(output)
		t.Error("output did not contain --section-another-test")
	}
	if !strings.Contains(output, "--anothersection-more-tests <MoreTests> (default: 42)") {
		fmt.Println(output)
		t.Error("output did not contain --anothersection-more-tests <MoreMoreTests> (default: 42)")
	}
	if !strings.Contains(output, "--anothersection-namedsubsection-sub-test <SubTest>") {
		fmt.Println(output)
		t.Error("output did not contain --anothersection-namedsubsection-sub-test <SubTest>")
	}

}

func TestCommandLineHelpSubsections(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()
	// Create an io.Writer that we can write to
	buffer := bytes.NewBuffer([]byte{})

	cfg := struct {
		Foo     string
		Section struct {
			Test        string
			AnotherTest string `help:"This is another test"`
		}
		AnotherSection struct {
			MoreTests  int `default:"42"`
			SubSection struct {
				SubTest       string `help:"This is a sub test"`
				MoreMoreTests int    `default:"42"`
			} `name:"named-subsection" help:"This is a named section"`
		}
	}{}

	os.Args = []string{"test", "--help"}

	err := configape.Apply(&cfg, &configape.Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		DisableCommandLine: false,
		HelpWriter:         buffer,
	})
	if err != nil {
		t.Fatal(err)
	}

	output := buffer.String()

	// The output should contain things like:
	// --section--test
	// --section--another-test
	// --another-section--more-tests (default: 42)
	if !strings.Contains(output, "--section-test") {
		t.Error("output did not contain --section-test")
	}
	if !strings.Contains(output, "--section-another-test") {
		t.Error("output did not contain --section--another-test")
	}
	if !strings.Contains(output, "--anothersection-more-tests <MoreTests> (default: 42)") {
		fmt.Println(output)
		t.Error("output did not contain --anothersection-more-tests <MoreTests> (default: 42)")
	}
}

func TestVersion(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()
	// Create an io.Writer that we can write to
	buffer := bytes.NewBuffer([]byte{})

	cfg := testConfig{}
	os.Args = []string{"test", "--version"}

	options := configape.Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		DisableCommandLine: false,
		HelpWriter:         buffer,
	}

	err := configape.Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}

	output := buffer.String()
	if output != "test\n" {
		fmt.Println(output)
		t.Error("output did not contain version")
	}
}
func TestFullVersion(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()
	// Create an io.Writer that we can write to
	buffer := bytes.NewBuffer([]byte{})

	cfg := testConfig{}
	os.Args = []string{"test", "--version"}

	options := configape.Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		DisableCommandLine: false,
		HelpWriter:         buffer,
		Version:            "1.2.3",
		Name:               "myprog",
	}

	buffer.Reset()
	err := configape.Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	output := buffer.String()
	if output != "myprog (v1.2.3)\n" {
		fmt.Println(output)
		t.Error("output did not contain myprog (v1.2.3)")
	}
}

func TestNoVersion(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	os.Args = []string{"test", "--version"}
	cfg := testConfig{}
	buffer := bytes.NewBuffer([]byte{})

	options := configape.Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		DisableCommandLine: false,
		Version:            "1.2.3",
		Name:               "myprog",
		DisableVersion:     true,
		HelpWriter:         buffer,
	}
	err := configape.Apply(&cfg, &options)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--version") {
		t.Errorf("expected error to contain --version: %s", err)
	}

	output := buffer.String()
	if output != "" {
		fmt.Println(output)
		t.Error("output was not empty")
	}
}
