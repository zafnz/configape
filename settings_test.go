package configape

import (
	"reflect"
	"testing"
)

type simpleCustomUnmarshalerType int

func (c *simpleCustomUnmarshalerType) UnmarshalText(b []byte) error {
	*c = 42
	return nil
}

func TestBadSubsectionName(t *testing.T) {
	// Make sure that subsections are not camelcased
	cfg := struct {
		Foo    string
		BarBar struct {
			Baz string
		}
	}{}
	typeOfCfg := reflect.TypeOf(cfg)
	settings, err := structToSettings(typeOfCfg)
	if err != nil {
		t.Error(err)
	}
	if len(settings) != 2 {
		t.Error("Expected 2 settings")
	}
	setting := settings.Find("foo", "")
	if setting == nil {
		t.Error("Did not find foo")
	}
	setting = settings.Find("bar-bar", "")
	if setting != nil {
		t.Error("Found bar-bar when it should have had it's camcelcase removed")
	}
}

func TestSubsections(t *testing.T) {
	s := struct {
		Foo     string
		Section struct {
			Test string
		}
		CamelSection struct {
			AnotherTest string
		} `name:"camelsection"`
	}{}
	typeOfCfg := reflect.TypeOf(s)

	settings, err := structToSettings(typeOfCfg)
	if err != nil {
		t.Error(err)
	}
	if len(settings) != 3 {
		t.Error("Expected 3 settings")
	}
	setting := settings.Find("foo", "")
	if setting == nil {
		t.Error("Did not find foo")
	}
	setting = settings.FindRecursive("section-test", "")
	if setting == nil {
		t.Error("Did not find section-test")
	}
	setting = settings.FindRecursive("camelsection-another-test", "")
	if setting == nil {
		t.Error("Did not find camelsection-another-test")
	}
}

func TestCustomUnmarshaler(t *testing.T) {
	s := struct {
		Foo    customUnmarshalerType
		Bar    simpleCustomUnmarshalerType
		Struct struct {
			Foo string
		}
	}{}
	typeOfCfg := reflect.TypeOf(s)

	settings, err := structToSettings(typeOfCfg)
	if err != nil {
		t.Error(err)
	}
	if len(settings) != 3 {
		t.Error("Expected 3 settings")
	}
	setting := settings.Find("foo", "")
	if setting == nil {
		t.Error("Did not find foo")
	} else if setting.fieldType != fieldTypeCustomMarshaler {
		t.Error("foo was not a custom unmarshaler")
	}
	setting = settings.Find("bar", "")
	if setting == nil {
		t.Error("Did not find bar")
	} else if setting.fieldType != fieldTypeCustomMarshaler {
		t.Error("bar was not a custom unmarshaler")
	}
	setting = settings.Find("struct", "")
	if setting == nil {
		t.Error("Did not find struct")
	} else if setting.fieldType != fieldTypeSubsection {
		t.Error("struct was not a struct")
	}
}

func TestParseIntoSettings(t *testing.T) {
	cfg := struct {
		Foo     string `default:"baz"`
		Bar     string `required:"true"`
		Flag    bool
		Empty   string
		Number  int
		List    []string
		Counter int `cfgtype:"counter"`
	}{}
	typeOfCfg := reflect.TypeOf(cfg)

	settings, err := structToSettings(typeOfCfg)
	if err != nil {
		t.Error(err)
	}
	if len(settings) != 7 {
		t.Error("Expected 6 settings")
	}
	setting := settings.Find("foo", "")
	if setting == nil {
		t.Error("Did not find foo")
	} else if setting.defaultValue != "baz" {
		t.Error("foo default value was not baz")
	}
	setting = settings.Find("bar", "")
	if setting == nil {
		t.Error("Did not find bar")
	} else if setting.required != true {
		t.Error("bar was not required")
	}
	setting = settings.Find("flag", "")
	if setting == nil {
		t.Error("Did not find flag")
	} else if setting.fieldType != fieldTypeFlag {
		t.Error("flag was not a flag")
	}
	setting = settings.Find("empty", "")
	if setting == nil {
		t.Error("Did not find empty")
	}
	setting = settings.Find("number", "")
	if setting == nil {
		t.Error("Did not find number")
	}
	setting = settings.Find("list", "")
	if setting == nil {
		t.Error("Did not find list")
	} else if setting.fieldType != fieldTypeList {
		t.Error("list was not a list")
	}

	setting = settings.Find("counter", "")
	if setting == nil {
		t.Error("Did not find counter")
	} else if setting.fieldType != fieldTypeCounter {
		t.Error("counter was not a counter")
	}
}

func TestCamelCaseConvert(t *testing.T) {
	tests := []struct {
		Input string
		Match string
	}{
		{"foo", "foo"},
		{"foo-bar", "foo-bar"},
		{"FooBar", "foo-bar"},
		{"fooBar", "foo-bar"},
		{"FFooBar", "ffoo-bar"},
		{"fooBarBaz", "foo-bar-baz"},
		{"testURL", "test-url"},
		{"endDD", "end-dd"},
		{"MiddleDDMiddle", "middle-ddmiddle"},
	}
	for _, test := range tests {
		result := camelCaseConvert(test.Input, '-')
		if result != test.Match {
			t.Errorf("camelCaseConfig(%s) was not %s, was %s", test.Input, test.Match, result)
		}
	}
}

func TestFindSetting(t *testing.T) {
	// Create sample cfgSettings
	cfgSettings := cfgSettings{
		cfgSetting{name: "lowercase"},
		cfgSetting{name: "Uppercase"},
		cfgSetting{name: "CamelCase"},
		cfgSetting{name: "hyphen-field"},
		cfgSetting{name: "DDoubleStart"},
		cfgSetting{name: "Number1Case"},
		cfgSetting{name: "UInt32List"},
	}

	// Test a few cases
	result := cfgSettings.Find("lowercase", "")
	if result == nil {
		t.Error("Did not find plain lowercase")
	}
	result = cfgSettings.Find("LOWERCASE", "")
	if result == nil {
		t.Error("Did not find lowercase when given all uppercase")
	}
	result = cfgSettings.Find("CamelCase", "")
	if result == nil {
		t.Error("Did not find CamcelCase as plain")
	}
	result = cfgSettings.Find("camel-case", "")
	if result == nil {
		t.Error("Did not find camel-case with dash")
	}
	result = cfgSettings.Find("camel_case", "")
	if result == nil {
		t.Error("Did not find camel-case with underscore")
	}
	result = cfgSettings.Find("uppercase", "")
	if result == nil {
		t.Error("Did not find uppercase as plain")
	}
	result = cfgSettings.Find("foo", "")
	if result != nil {
		t.Error("Found foo when it should not have been found")
	}
	result = cfgSettings.Find("hyphen_field", "")
	if result == nil {
		t.Error("Did not find hyphen_field as plain")
	}
	result = cfgSettings.Find("hyphen-field", "")
	if result == nil {
		t.Error("Did not find hyphen-field as plain")
	}
	result = cfgSettings.Find("ddouble-start", "")
	if result == nil {
		t.Error("Did not find ddouble-start as plain")
	}

	result = cfgSettings.Find("number1-case", "")
	if result == nil {
		t.Error("Did not find number1-case as plain")
	}
	result = cfgSettings.Find("uint32-list", "")
	if result == nil {
		t.Error("Did not find uint32-list as plain")
	}
}

func TestSkipField(t *testing.T) {
	cfg := struct {
		Foo string `name:"-"`
		Bar string
	}{}
	typeOfCfg := reflect.TypeOf(cfg)

	settings, err := structToSettings(typeOfCfg)
	if err != nil {
		t.Error(err)
	}
	if len(settings) != 1 {
		t.Error("Expected 1 setting")
	}
	setting := settings.Find("foo", "")
	if setting != nil {
		t.Error("Found foo when it should have been skipped")
	}
	setting = settings.Find("bar", "")
	if setting == nil {
		t.Error("Did not find bar")
	}
}

func TestNamedField(t *testing.T) {
	cfg := struct {
		Foo string `name:"bar"`
	}{}
	typeOfCfg := reflect.TypeOf(cfg)

	settings, err := structToSettings(typeOfCfg)
	if err != nil {
		t.Error(err)
	}
	if len(settings) != 1 {
		t.Error("Expected 1 setting")
	}
	setting := settings.Find("foo", "")
	if setting != nil {
		t.Error("Found foo when it should have been named bar")
	}
	setting = settings.Find("bar", "")
	if setting == nil {
		t.Error("Did not find bar")
	}
}

func TestSpecificName(t *testing.T) {
	cfg := struct {
		Foo string `name:"bar" cli:"Fizz"`
	}{}
	typeOfCfg := reflect.TypeOf(cfg)

	settings, err := structToSettings(typeOfCfg)
	if err != nil {
		t.Error(err)
	}
	if len(settings) != 1 {
		t.Error("Expected 1 setting")
	}
	setting := settings.Find("foo", "cli")
	if setting != nil {
		t.Error("Found foo when it should have been named Fizz")
	}
	setting = settings.Find("bar", "cli")
	if setting != nil {
		t.Error("Found bar when it should have been named Fizz")
	}
	setting = settings.Find("Fizz", "cli")
	if setting == nil {
		t.Error("Did not find Fizz")
	}
}

func TestDefault(t *testing.T) {
	cfg := struct {
		Foo string `default:"bar"`
		Num int    `default:"42"`
	}{}
	typeOfCfg := reflect.TypeOf(cfg)
	settings, err := structToSettings(typeOfCfg)
	if err != nil {
		t.Error(err)
	}
	settings.SetDefaults()
	if setting := settings.Find("foo", ""); setting != nil {
		if setting.reflectValue.String() != "bar" {
			t.Error("foo was not bar")
		}
	} else {
		t.Error("Did not find foo")
	}
	if setting := settings.Find("num", ""); setting != nil {
		if setting.reflectValue.Int() != 42 {
			t.Error("num was not 42")
		}
	} else {
		t.Error("Did not find num")
	}
}

func TestPrivateFields(t *testing.T) {
	cfg := struct {
		Foo string
		Bar string
		baz string
	}{}
	typeOfCfg := reflect.TypeOf(cfg)
	settings, err := structToSettings(typeOfCfg)
	if err != nil {
		t.Error(err)
	}
	if len(settings) != 2 {
		t.Error("Expected 2 settings")
	}
	setting := settings.Find("foo", "")
	if setting == nil {
		t.Error("Did not find foo")
	}
	setting = settings.Find("bar", "")
	if setting == nil {
		t.Error("Did not find bar")
	}
	setting = settings.Find("baz", "")
	if setting != nil {
		t.Error("Found baz when it should have been skipped")
	}
}

func TestStrToTypes(t *testing.T) {
	cfg := struct {
		NumList    []int
		UInt32List []uint32
		FloatList  []float64
		BoolList   []bool
		StrList    []string
	}{}
	typeOfCfg := reflect.TypeOf(cfg)
	settings, err := structToSettings(typeOfCfg)
	if err != nil {
		t.Fatal(err)
	}
	setting := settings.Find("num-list", "")
	if setting == nil {
		t.Error("Did not find num-list")
	} else {
		val, err := strToType(setting.reflectType, "1,2,3")
		if err != nil {
			t.Error(err)
		}
		// Check if the value is a slice of ints
		if val.Kind() != reflect.Slice || val.Type().Elem().Kind() != reflect.Int {
			t.Errorf("num-list was not a slice of ints ")
		} else if val.Interface().([]int)[0] != 1 {
			t.Error("num-list[0] was not 1")
		}

	}
	setting = settings.Find("uint32-list", "")
	if setting == nil {
		t.Error("Did not find uint32-list")
	} else {
		val, err := strToType(setting.reflectType, "1,2,3")
		if err != nil {
			t.Error(err)
		}
		if val.Kind() != reflect.Slice || val.Type().Elem().Kind() != reflect.Uint32 {
			t.Error("uint32-list was not a slice of uint32")
		} else if val.Interface().([]uint32)[0] != 1 {
			t.Error("uint32-list[0] was not 1")
		}
	}
}
