package configape

import (
	"bytes"
	"reflect"
	"testing"
)

type customUnmarshalerType struct {
	Foo string
}

func (c *customUnmarshalerType) UnmarshalJSON(b []byte) error {
	c.Foo = string(b)
	return nil
}
func (c *customUnmarshalerType) UnmarshalText(b []byte) error {
	c.Foo = "\"" + string(b) + "\""
	return nil
}

func TestJsonParse(t *testing.T) {

	settings := cfgSettings{
		cfgSetting{idx: 0, name: "foo", reflectType: reflect.TypeOf("")},
		cfgSetting{idx: 1, name: "bar", reflectType: reflect.TypeOf("")},
		cfgSetting{idx: 2, name: "flag", reflectType: reflect.TypeOf(true)},
		// This custom is a pointer to a customUnmarshalerType
		cfgSetting{idx: 3, name: "custom", reflectType: reflect.TypeOf(customUnmarshalerType{})},
	}
	fileContents := `
	{
		"foo": "bar",
		"bar": "baz",
		"flag": true,
		"custom": "wibble"
	}`
	options := Options{
		cfgFileContents:    fileContents,
		DisableEnviornment: true,
		DisableCommandLine: true,
	}
	c := cfgApe{
		settings: settings,
		options:  options,
	}
	fh := bytes.NewBufferString(fileContents)
	err := c.parseJsonConfigFile("fake.json", fh)
	if err != nil {
		t.Error(err)
	}
	// Check the setting.reflectValue matches
	if settings[0].reflectValue.String() != "bar" {
		t.Error("foo was not bar")
	}
	if settings[1].reflectValue.String() != "baz" {
		t.Error("bar was not baz")
	}
	if settings[2].reflectValue.Bool() != true {
		t.Errorf("flag was not true: %+v", settings[2].reflectValue)
	}
	// Check if custom's reflectValue has a type of customUnmarshalerType
	if settings[3].reflectValue.Type() != reflect.TypeOf(customUnmarshalerType{}) {
		t.Errorf("custom was not customUnmarshalerType: %+v", settings[3].reflectValue.Type())
	} else {
		// Cast the reflectValue to a customUnmarshalerType and check the value
		custom := settings[3].reflectValue.Interface().(customUnmarshalerType)
		if custom.Foo != "\"wibble\"" {
			t.Errorf("custom was not \"wibble\": %+v", custom)
		}
	}

}
