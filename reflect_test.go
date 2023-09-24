package configape

import (
	"reflect"
	"strconv"
	"testing"
)

func TestTypes(t *testing.T) {
	types := struct {
		Str     string
		Bool    bool
		Int     int
		Int16   int16
		Int32   int32
		Int64   int64
		Float64 float64
		Float32 float32
		Byte    byte
		Rune    rune
		Uint    uint
		Uint8   uint8
		Uint16  uint16
		Uint32  uint32
		Uint64  uint64
	}{}
	// For each field in types, call strToType with a valid string for that type
	// and check that the value is correct.

	var err error
	var value reflect.Value

	value, err = strToType(reflect.TypeOf(types.Str), "foo")
	if err != nil {
		t.Fatal(err)
	}
	if value.String() != "foo" {
		t.Error("Str was not foo")
	}
	// Now for Bool
	value, err = strToType(reflect.TypeOf(types.Bool), "true")
	if err != nil {
		t.Fatal(err)
	}
	if value.Type() != reflect.TypeOf(true) {
		t.Fatalf("Bool was not a bool, it's a %s", value.Type().String())
	}
	if value.Bool() != true {
		t.Error("Bool was not true")
	}
	value, err = strToType(reflect.TypeOf(types.Bool), "false")
	if err != nil {
		t.Fatal(err)
	}
	if value.Bool() != false {
		t.Error("Bool was not false")
	}
	// Now for Int
	value, err = strToType(reflect.TypeOf(types.Int), "42")
	if err != nil {
		t.Fatal(err)
	}
	if value.Int() != 42 {
		t.Error("Int was not 42")
	}
	// Now for Int16
	value, err = strToType(reflect.TypeOf(types.Int16), "42")
	if err != nil {
		t.Fatal(err)
	}
	if value.Int() != 42 {
		t.Error("Int16 was not 42")
	}
	// Now for Int32
	value, err = strToType(reflect.TypeOf(types.Int32), "42")
	if err != nil {
		t.Fatal(err)
	}
	if value.Int() != 42 {
		t.Error("Int32 was not 42")
	}
	// Now for Int64
	value, err = strToType(reflect.TypeOf(types.Int64), "42")
	if err != nil {
		t.Fatal(err)
	}
	if value.Int() != 42 {
		t.Error("Int64 was not 42")
	}
	// Now for Float64
	value, err = strToType(reflect.TypeOf(types.Float64), "42.1234")
	if err != nil {
		t.Fatal(err)
	}
	if value.Float() != 42.1234 {
		t.Error("Float64 was not 42.1234")
	}
	// Now for Float32
	f, _ := strconv.ParseFloat("42.1234", 32)
	value, err = strToType(reflect.TypeOf(types.Float32), "42.1234")
	if err != nil {
		t.Fatal(err)
	}
	if value.Float() != f {
		t.Errorf("Float32 was not %f, it is %f", f, value.Float())
	}
	// Now for Byte
	value, err = strToType(reflect.TypeOf(types.Byte), "42")
	if err != nil {
		t.Fatal(err)
	}
	if value.Uint() != 42 {
		t.Error("Byte was not 42")
	}
	// Now for Rune
	value, err = strToType(reflect.TypeOf(types.Rune), "42")
	if err != nil {
		t.Fatal(err)
	}
	if value.Int() != 42 {
		t.Error("Rune was not 42")
	}
	// Now for Uint
	value, err = strToType(reflect.TypeOf(types.Uint), "42")
	if err != nil {
		t.Fatal(err)
	}
	if value.Uint() != 42 {
		t.Error("Uint was not 42")
	}
	// Now for Uint8
	value, err = strToType(reflect.TypeOf(types.Uint8), "42")
	if err != nil {
		t.Fatal(err)
	}
	if value.Uint() != 42 {
		t.Error("Uint8 was not 42")
	}
	// Now for Uint16
	value, err = strToType(reflect.TypeOf(types.Uint16), "42")
	if err != nil {
		t.Fatal(err)
	}
	if value.Uint() != 42 {
		t.Error("Uint16 was not 42")
	}
	// Now for Uint32
	value, err = strToType(reflect.TypeOf(types.Uint32), "42")
	if err != nil {
		t.Fatal(err)
	}
	if value.Uint() != 42 {
		t.Error("Uint32 was not 42")
	}
	// Now for Uint64
	value, err = strToType(reflect.TypeOf(types.Uint64), "42")
	if err != nil {
		t.Fatal(err)
	}
	if value.Uint() != 42 {
		t.Error("Uint64 was not 42")
	}

}
