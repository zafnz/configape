package configape

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestPassByValue(t *testing.T) {
	cfg := struct {
		Foo string
	}{}
	err := Apply(cfg, nil)
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "cfg must be a pointer" {
		t.Error("Expected different error")
	}
}

func TestStringListToMap(t *testing.T) {
	str := "foo=bar,bar=baz"
	m := stringListToMap(str)
	if m["foo"] != "bar" {
		t.Error("foo was not bar")
	}
	if m["bar"] != "baz" {
		t.Error("bar was not baz")
	}
	if len(m) != 2 {
		t.Error("Map length was not 2")
	}

	str = "goo,blah=thing"
	m = stringListToMap(str)
	if m["goo"] != "" {
		t.Error("goo was not empty")
	}
	if m["blah"] != "thing" {
		t.Error("blah was not thing")
	}
	if len(m) != 2 {
		t.Error("Map length was not 2")
	}

	str = "blah=thing,str=\"foo,bar\",other='goo,baz'"
	m = stringListToMap(str)
	if m["blah"] != "thing" {
		t.Error("blah was not thing")
	}
	if m["str"] != "foo,bar" {
		t.Error("str was not foo,bar")
	}
	if m["other"] != "goo,baz" {
		t.Error("other was not goo,baz")
	}
	if len(m) != 3 {
		t.Error("Map length was not 3")
	}
}

type testConfig struct {
	Foo       string  `name:"foo" default:"baz" help:"This is the help for foo"`
	Bar       *string `help:"This is the help for bar"`
	Flag      bool    `help:"This is the help for flag"`
	Empty     string  `cfg:"-"`
	Number    int     `help:"This is the help for number" default:"42"`
	Float     float64 `help:"This is the help for float"`
	List      []string
	Counter   int `cfgtype:"counter" help:"This is the help for counter"`
	Interface interface{}
	CamelTest string `help:"CamelCase test"`
}

func TestApply(t *testing.T) {
	cfg := testConfig{}
	options := Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		osArgs:             []string{"cfgape", "--foo", "bar", "--bar=baz", "--flag", "--number=42", "--list=foo", "--list", "bar", "--counter", "--interface=boo", "--float=42.1234"},
	}
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Foo != "bar" {
		t.Error("foo was not bar")
	}
	if cfg.Bar == nil {
		t.Error("bar was nil")
	} else {
		if *cfg.Bar != "baz" {
			t.Error("bar was not baz")
		}
	}
	if cfg.Number != 42 {
		t.Error("number was not 42")
	}
	if cfg.Flag != true {
		t.Error("flag was not true")
	}
	if cfg.List[0] != "foo" {
		t.Error("list[0] was not foo")
	}
	if cfg.List[1] != "bar" {
		t.Error("list[1] was not bar")
	}
	if cfg.Interface != "boo" {
		t.Error("interface was not boo")
	}
	if cfg.Float != 42.1234 {
		t.Error("float was not 42.1234")
	}
}

func TestDefaults(t *testing.T) {
	cfg := testConfig{}
	options := Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		osArgs:             []string{"cfgape"},
	}
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Foo != "baz" {
		t.Error("foo was not the default value of baz")
	}
}

func TestNumberedList(t *testing.T) {
	cfg := struct {
		List []int
	}{}
	options := Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		osArgs:             []string{"cfgape", "--list=1", "--list", "2", "--list", "3"},
	}
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.List[0] != 1 {
		t.Error("list[0] was not 1")
	}
	if cfg.List[1] != 2 {
		t.Error("list[1] was not 2")
	}
	if cfg.List[2] != 3 {
		t.Error("list[2] was not 3")
	}
}

func TestRequired(t *testing.T) {
	cfg := struct {
		Foo string `required:"true"`
	}{}
	options := Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		osArgs:             []string{"cfgape"},
	}
	err := Apply(&cfg, &options)
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "Foo") {
		t.Errorf("Expected different error: %s", err.Error())
	}
}

func TestAllSettings(t *testing.T) {
	cfg := struct {
		Default string `cfg:"foo" default:"baz"`
		File    string
		Env     string
		CmdLine string
	}{}
	args := []string{"cfgape", "--cmd-line", "cliset"}
	fileContents := `
	{
		"file": "fileset"
	}`
	os.Setenv("CFG_ENV", "envset")
	options := Options{
		DisableEnviornment: false,
		DisableConfigFile:  false,
		DisableCommandLine: false,
		osArgs:             args,
		cfgFileContents:    fileContents,
	}
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Default != "baz" {
		t.Error("Default was not baz")
	}
	if cfg.File != "fileset" {
		t.Error("File was not fileset")
	}
	if cfg.Env != "envset" {
		t.Error("Env was not envset")
	}
	if cfg.CmdLine != "cliset" {
		t.Error("CmdLine was not cliset")
	}
	os.Unsetenv("CFG_ENV")
}

func TestAllPrescidence(t *testing.T) {
	cfg := struct {
		Default string `name:"foo" default:"baz"`
		File    string
		Env     string
		CmdLine string
	}{}
	args := []string{"cfgape", "--cmd-line", "cliset"}
	fileContents := `
	{
		"file": "fileset",
		"env": "fileset",
		"cmd-line": "fileset"
	}`
	os.Setenv("CFG_ENV", "envset")
	os.Setenv("CFG_CMD_LINE", "envset")
	options := Options{
		DisableEnviornment: false,
		DisableConfigFile:  false,
		DisableCommandLine: false,
		osArgs:             args,
		cfgFileContents:    fileContents,
	}
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Default != "baz" {
		t.Error("Default was not baz")
	}
	if cfg.File != "fileset" {
		t.Error("File was not fileset")
	}
	if cfg.Env != "envset" {
		t.Error("Env was not envset")
	}
	if cfg.CmdLine != "cliset" {
		t.Error("CmdLine was not cliset")
	}
	os.Unsetenv("CFG_ENV")
	os.Unsetenv("CFG_CMD_LINE")
}

func TestInvalidTypes(t *testing.T) {
	cfg := struct {
		Foo int
		Bar bool
		Baz float64
	}{}
	osArgs := []string{"cfgape", "--foo", "bar"}
	options := Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		osArgs:             osArgs,
	}
	err := Apply(&cfg, &options)
	if err == nil {
		t.Error("Expected error")
	} else if !strings.Contains(err.Error(), "--foo") {
		t.Errorf("Expected different error %s", err.Error())
	}
	osArgs = []string{"cfgape", "--bar=bar"}
	options.osArgs = osArgs
	err = Apply(&cfg, &options)
	if err == nil {
		t.Error("Expected error")
	} else if !strings.Contains(err.Error(), "--bar") {
		t.Errorf("Expected different error %s", err.Error())
	}
	osArgs = []string{"cfgape", "--baz", "bar"}
	options.osArgs = osArgs
	err = Apply(&cfg, &options)
	if err == nil {
		t.Error("Expected error")
	} else if !strings.Contains(err.Error(), "--baz") {
		t.Errorf("Expected different error: %s", err.Error())
	}

	// This should error because default is an int, not a string
	cfgDefault := struct {
		Default int `default:"baz"`
	}{}

	osArgs = []string{"cfgape"}
	options.osArgs = osArgs
	err = Apply(&cfgDefault, &options)
	if err == nil {
		t.Error("expected error")
	} else if !strings.Contains(err.Error(), "Default") {
		t.Errorf("Expected different error: %s", err.Error())
	}

}

type marshalTest struct {
	Foo string
}

func (m *marshalTest) UnmarshalText(text []byte) error {
	m.Foo = string(text) + "baz"
	return nil
}
func (m *marshalTest) MarshalText() ([]byte, error) {
	// strip a trailing baz from the end if it is there.
	str := m.Foo
	str = strings.TrimSuffix(str, "baz")
	return []byte(str), nil
}

func TestMarshalling(t *testing.T) {
	cfg := struct {
		Foo marshalTest
	}{}
	typeOfFoo := reflect.TypeOf(cfg.Foo)
	// is Foo a struct
	if typeOfFoo.Kind() == reflect.Struct {
		fmt.Println("Foo is a struct")
	}
	osArgs := []string{"cfgape", "--foo", "bar"}
	options := Options{
		DisableEnviornment: true,
		DisableConfigFile:  true,
		osArgs:             osArgs,
	}
	err := Apply(&cfg, &options)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Foo.Foo != "barbaz" {
		t.Error("Foo was not barbaz")
	}
}
