package configape_test

import (
	"os"
	"testing"

	"github.com/zafnz/configape"
)

func TestComplexJson(t *testing.T) {
	fileContents := `
	{
		"foo": "bar",
		"bar": "baz",
		"subsection": {
			"foo": "bar",
			"subsubsection": {
				"baz": "goo"
			}
		},
		"flag": true,
		"falseflag": false,
		"list": ["foo", "bar", "baz"]
	}
	`
	fh, _ := os.CreateTemp("", "configape")
	defer os.Remove(fh.Name())
	fh.WriteString(fileContents)

	cfg := struct {
		Foo        string `name:"foo"`
		Bar        string `name:"bar"`
		Flag       bool
		Falseflag  bool `default:"false"`
		List       []string
		Subsection struct {
			Foo           string `name:"foo"`
			SubSubsection struct {
				Baz string `name:"baz"`
			} `name:"subsubsection"`
		} `name:"subsection"`
	}{}

	options := configape.Options{
		DisableEnviornment: true,
		DisableCommandLine: true,
		DisableConfigFile:  false,
		ConfigFilename:     fh.Name(),
		ConfigFileType:     "json",
	}
	err := configape.Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Foo != "bar" {
		t.Error("foo was not bar")
	}
	if cfg.Bar != "baz" {
		t.Error("bar was not baz")
	}
	if cfg.Subsection.Foo != "bar" {
		t.Error("subsection.foo was not bar")
	}
	if cfg.Subsection.SubSubsection.Baz != "goo" {
		t.Error("subsection.subsubsection.baz was not goo")
	}
	if cfg.Flag != true {
		t.Error("flag was not true")
	}
	if cfg.Falseflag != false {
		t.Error("falseflag was not false")
	}
	if len(cfg.List) != 3 {
		t.Error("list was not length 3")
	} else {
		if cfg.List[0] != "foo" {
			t.Error("list[0] was not foo")
		}
		if cfg.List[1] != "bar" {
			t.Error("list[1] was not bar")
		}
		if cfg.List[2] != "baz" {
			t.Error("list[2] was not baz")
		}
	}

}

func TestYaml(t *testing.T) {
	fileContents := `
foo: bar
bar: baz
subsection:
    foo: bar
    list:
      - foo
      - bar
`
	fh, _ := os.CreateTemp("", "configape")
	defer os.Remove(fh.Name())
	fh.WriteString(fileContents)
	cfg := struct {
		Foo        string `cfg:"foo"`
		Bar        string `cfg:"bar"`
		Subsection struct {
			Foo           string `cfg:"foo"`
			List          []string
			SubSubsection struct {
				Baz string `cfg:"baz"`
			}
		}
	}{}
	options := configape.Options{
		ConfigFileType:     "yaml",
		DisableEnviornment: true,
		DisableCommandLine: true,
		DisableConfigFile:  false,
		ConfigFilename:     fh.Name(),
	}
	err := configape.Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Foo != "bar" {
		t.Errorf("foo was not bar, was: %s", cfg.Foo)
	}
	if cfg.Bar != "baz" {
		t.Errorf("bar was not baz, was: %s", cfg.Bar)
	}
	if cfg.Subsection.Foo != "bar" {
		t.Errorf("subsection.foo was not bar, was: %s", cfg.Subsection.Foo)
	}
	if cfg.Subsection.SubSubsection.Baz != "" {
		t.Errorf("subsection.subsubsection.baz was not empty, was: %s", cfg.Subsection.SubSubsection.Baz)
	}
	if len(cfg.Subsection.List) != 2 {
		t.Error("subsection.list was not length 2")
	} else {
		if cfg.Subsection.List[0] != "foo" {
			t.Errorf("subsection.list[0] was not foo, was: %s", cfg.Subsection.List[0])
		}
		if cfg.Subsection.List[1] != "bar" {
			t.Errorf("subsection.list[1] was not bar, was: %s", cfg.Subsection.List[1])
		}
	}
}

func TestUnknownSetting(t *testing.T) {
	fileContents := `
	{
		"foo": "bar",
		"bar": "baz",
		"flag": true,
		"falseflag": false,
		"list": ["foo", "bar", "baz"]
	}`
	fh, _ := os.CreateTemp("", "configape")
	defer os.Remove(fh.Name())
	fh.WriteString(fileContents)

	cfg := struct {
		Foo string `name:"foo"`
		Bar string `name:"bar"`
	}{}

	options := configape.Options{
		DisableEnviornment: true,
		DisableCommandLine: true,
		DisableConfigFile:  false,
		ConfigFilename:     fh.Name(),
		ConfigFileType:     "json",
	}
	err := configape.Apply(&cfg, &options)
	if err == nil {
		t.Fatal("Expected error")
	}
	options.AllowUnknownConfigFileKeys = true

	err = configape.Apply(&cfg, &options)
	if err != nil {
		t.Error(err)
	}
}
