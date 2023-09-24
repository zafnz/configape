package configape

import (
	"os"
	"testing"
)

func TestEnvironment(t *testing.T) {
	cfg := struct {
		Foo        string
		Bar        string
		CamelCase  string
		Flag       bool
		FlagTrue   bool
		FlagFalse  bool `default:"true"`
		FlagOne    bool
		Subsection struct {
			Foo string
		}
	}{}
	options := Options{
		DisableEnviornment: false,
		DisableConfigFile:  true,
		DisableCommandLine: true,
	}
	os.Setenv("CFG_FOO", "bar")
	os.Setenv("CFG_BAR", "baz")
	os.Setenv("CFG_SUBSECTION_FOO", "bar")
	os.Setenv("CFG_CAMEL_CASE", "bar")
	os.Setenv("CFG_FLAG", "")
	os.Setenv("CFG_FLAG_TRUE", "true")
	os.Setenv("CFG_FLAG_FALSE", "false")
	os.Setenv("CFG_FLAG_ONE", "1")

	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Foo != "bar" {
		t.Error("Foo was not bar")
	}
	if cfg.Bar != "baz" {
		t.Error("Bar was not baz")
	}
	if cfg.CamelCase != "bar" {
		t.Error("CamelCase was not bar")
	}
	if cfg.Subsection.Foo != "bar" {
		t.Error("Subsection.Foo was not bar")
	}
	if cfg.Flag != true {
		t.Error("Flag was not true")
	}
	if cfg.FlagTrue != true {
		t.Error("FlagTrue was not true")
	}
	if cfg.FlagFalse != false {
		t.Error("FlagFalse was not false")
	}
	if cfg.FlagOne != true {
		t.Error("FlagOne was not true")
	}
	os.Unsetenv("CFG_FOO")
	os.Unsetenv("CFG_BAR")
	os.Unsetenv("CFG_SUBSECTION_FOO")
	os.Unsetenv("CFG_CAMEL_CASE")
	os.Unsetenv("CFG_SUB_CAMEL_CASE_FOO")
	os.Unsetenv("CFG_FLAG")
	os.Unsetenv("CFG_FLAG_TRUE")
	os.Unsetenv("CFG_FLAG_FALSE")
	os.Unsetenv("CFG_FLAG_ONE")

}

func TestCustomEnvironmentPrefix(t *testing.T) {
	cfg := struct {
		Foo string
	}{}
	options := Options{
		DisableEnviornment: false,
		DisableConfigFile:  true,
		DisableCommandLine: true,
		EnvironmentPrefix:  "TEST_",
	}
	os.Setenv("TEST_FOO", "bar")
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Foo != "bar" {
		t.Error("Foo was not bar")
	}
	os.Unsetenv("TEST_FOO")
}

func TestEnvironmentList(t *testing.T) {
	cfg := struct {
		List []string
	}{}
	options := Options{
		DisableEnviornment: false,
		DisableConfigFile:  true,
		DisableCommandLine: true,
	}
	os.Setenv("CFG_LIST", "foo,bar")
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.List[0] != "foo" {
		t.Error("List[0] was not foo")
	}
	if cfg.List[1] != "bar" {
		t.Error("List[1] was not bar")
	}
	os.Unsetenv("CFG_LIST")
}

func TestEnvironmentCustomType(t *testing.T) {
	cfg := struct {
		Foo    string
		Custom customUnmarshalerType
	}{}
	options := Options{
		DisableEnviornment: false,
		DisableConfigFile:  true,
		DisableCommandLine: true,
	}
	os.Setenv("CFG_FOO", "bar")
	os.Setenv("CFG_CUSTOM", "foo")
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Foo != "bar" {
		t.Error("Foo was not bar")
	}
	if cfg.Custom.Foo != "\"foo\"" {
		t.Error("Custom was not \"foo\"")
	}
	os.Unsetenv("CFG_FOO")
	os.Unsetenv("CFG_CUSTOM")
}

func TestEnvironmentPointers(t *testing.T) {
	cfg := struct {
		Foo    *string
		Bar    *string
		Custom *customUnmarshalerType
	}{}
	options := Options{
		DisableEnviornment: false,
		DisableConfigFile:  true,
		DisableCommandLine: true,
	}
	os.Setenv("CFG_FOO", "bar")
	os.Setenv("CFG_CUSTOM", "foo")
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Foo == nil {
		t.Error("Foo was nil")
	} else if *cfg.Foo != "bar" {
		t.Error("Foo was not bar")
	}
	if cfg.Bar != nil {
		t.Error("Bar was not nil")
	}
	if cfg.Custom == nil {
		t.Error("Custom was nil")
	} else if cfg.Custom.Foo != "\"foo\"" {
		t.Error("Custom was not \"foo\"")
	}
	os.Unsetenv("CFG_FOO")
	os.Unsetenv("CFG_CUSTOM")
}
