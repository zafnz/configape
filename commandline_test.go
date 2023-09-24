package configape

import (
	"testing"
)

func TestParseCommandline(t *testing.T) {
	args := []string{"test", "--foo", "bar", "--baz-test=thing"}
	cfg := struct {
		Foo     string
		BazTest string
		Flag    bool
	}{}
	options := Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		DisableCommandLine: false,
		osArgs:             args,
	}
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Foo != "bar" {
		t.Error("Foo was not bar")
	}
	if cfg.BazTest != "thing" {
		t.Error("BazTest was not thing")
	}
}

func TestShortArguments(t *testing.T) {
	args := []string{"test", "--foo", "bar", "-b", "thing"}
	cfg := struct {
		Foo     string
		BazTest string `short:"b"`
		Flag    bool   `short:"f"`
	}{}
	options := Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		DisableCommandLine: false,
		osArgs:             args,
	}
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Foo != "bar" {
		t.Error("Foo was not bar")
	}
	if cfg.BazTest != "thing" {
		t.Error("BazTest was not thing")
	}

}

func TestSetFlagNo(t *testing.T) {
	args := []string{"cfgape", "--flag", "--no-truth", "--set-truth=false"}
	cfg := struct {
		Flag     bool
		Truth    bool `default:"true"`
		SetTruth bool `default:"true"`
	}{}
	options := Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		DisableCommandLine: false,
		osArgs:             args,
	}
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Flag != true {
		t.Error("Flag was not true")
	}
	if cfg.Truth != false {
		t.Error("Truth was not false")
	}
	if cfg.SetTruth != false {
		t.Error("SetTruth was not false")
	}
}

func TestDoubleHyphen(t *testing.T) {
	args := []string{"cfgape", "blah", "--", "--foo", "bar"}
	cfg := struct {
		Foo       string
		Remaining []string `name:"*"`
	}{}
	options := Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		DisableCommandLine: false,
		osArgs:             args,
	}
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Foo != "" {
		t.Error("Foo was not empty")
	}
	if len(cfg.Remaining) != 3 {
		t.Errorf("Remaining was not 3, is %d", len(cfg.Remaining))
	} else if cfg.Remaining[0] != "blah" {
		t.Error("Remaining[0] was not blah")
	} else if cfg.Remaining[1] != "--foo" {
		t.Error("Remaining[1] was not --foo")
	} else if cfg.Remaining[2] != "bar" {
		t.Error("Remaining[2] was not bar")
	}
}

func TestCase(t *testing.T) {
	args := []string{"cfgape", "--bar-name", "bar", "--FooName", "foo"}
	cfg := struct {
		FooName string
		BarName string
	}{}
	options := Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		DisableCommandLine: false,
		osArgs:             args,
	}
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FooName != "foo" {
		t.Error("FooName was not foo")
	}
	if cfg.BarName != "bar" {
		t.Error("BarName was not bar")
	}
}

func TestCommandLineSections(t *testing.T) {
	args := []string{"cfgape", "--foo", "bar", "--section-test", "thing", "--camelsection-another-test", "another-thing"}
	cfg := struct {
		Foo     string
		Section struct {
			Test string
		}
		CamelSection struct {
			AnotherTest string
		}
	}{}
	options := Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		DisableCommandLine: false,
		osArgs:             args,
	}
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Foo != "bar" {
		t.Error("Foo was not bar")
	}
	if cfg.Section.Test != "thing" {
		t.Error("Section.Test was not thing")
	}
	if cfg.CamelSection.AnotherTest != "another-thing" {
		t.Error("CamelSection.AnotherTest was not another-thing")
	}

}

func TestCommandLineConfigFile(t *testing.T) {
	args := []string{"cfgape", "--config", "test_config.json"}
	cfg := struct {
		Foo    string
		Config string `cfgtype:"configfile"`
	}{}
	options := Options{
		DisableEnviornment: true,
		DisableConfigFile:  false,
		DisableCommandLine: false,
		osArgs:             args,
	}
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Config != "test_config.json" {
		t.Error("Config was not test_config.json")
	}

	if cfg.Foo != "bar" {
		t.Error("Foo was not bar")
	}

}
