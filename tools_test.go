package configape

import "testing"

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
