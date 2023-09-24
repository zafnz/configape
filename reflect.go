package configape

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Most of the stuff to do with reflection

// Takes a reflect.Value and a string, and assigns the string to the value, using sensible
// conversions (eg "true" to true, "1" to 1, etc)
func strToValue(value reflect.Value, str string) error {
	// If the type implements TextUnmarshaler, then we can use that
	if unmarshal, ok := value.Interface().(encoding.TextUnmarshaler); ok {
		err := unmarshal.UnmarshalText([]byte(str))
		if err != nil {
			return err
		}
		return nil
	}
	// If it's a pointer, we need to muck with what it points to.
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Interface:
		// Interface gets the string
		value.Set(reflect.ValueOf(str))
	case reflect.Slice:
		// Split the string on commas (really caller should call strListToType)
		list := strings.Split(str, ",")
		slice, err := strListToType(value.Type(), list)
		if err != nil {
			return err
		}
		value.Set(slice)
	case reflect.String:
		value.SetString(str)
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		// Because json and yaml unmarshal numbers as float64, we need to handle that here
		f, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("invalid integer value: %s", str)
		}
		value.SetInt(int64(f))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Because json and yaml unmarshal numbers as float64, we need to handle that here
		f, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("invalid integer value: %s", str)
		}
		value.SetUint(uint64(f))
	case reflect.Bool:
		switch strings.ToLower(str) {
		case "true", "t", "1":
			value.SetBool(true)
		case "false", "f", "0":
			value.SetBool(false)
		default:
			return fmt.Errorf("invalid boolean value: %s", str)
		}
	case reflect.Float64, reflect.Float32:
		f, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %s", str)
		}
		value.SetFloat(f)
	case reflect.Struct:
		return fmt.Errorf("cannot cast string (%s) into struct, implement TextUnmarshaler", str)
	default:
		return fmt.Errorf("strToValue(%s, %s) unknown type", value.Kind(), str)
	}
	return nil
}

// Takes a reflection type, and a string list, and returns a list of that type,
// containing the strings converted to that type
func strListToType(valType reflect.Type, list []string) (reflect.Value, error) {
	elemType := valType
	if valType.Kind() == reflect.Pointer {
		elemType = valType.Elem()
	}
	slice := reflect.MakeSlice(elemType, len(list), len(list))
	for idx, v := range list {
		value, err := strToType(elemType.Elem(), v)
		if err != nil {
			return reflect.Zero(valType), err
		}
		slice.Index(idx).Set(value)
	}
	if valType.Kind() == reflect.Pointer {
		// Set the pointer to point to a new value
		v := reflect.New(valType)
		v.Set(slice)
		return v, nil
	}
	return slice, nil
}

// Appends the string to the currentList, whose type is valType. If currentList doesn't exist, it creates it.
func appendStrToListType(valType reflect.Type, currentList reflect.Value, value string) (reflect.Value, error) {
	elemType := valType
	if valType.Kind() == reflect.Ptr {
		elemType = valType.Elem()
	}
	if currentList.Kind() == reflect.Ptr {
		currentList = currentList.Elem()
	}
	if currentList.Kind() == reflect.Invalid {
		// Create a slice of the correct type
		slice := reflect.MakeSlice(elemType, 1, 1)
		err := strToValue(slice.Index(0), value)
		return slice, err
	}
	if currentList.Kind() != reflect.Slice {
		return reflect.Zero(valType), fmt.Errorf("list type is not a slice: %s", currentList.Kind())
	}
	if currentList.Type().Elem() != elemType.Elem() {
		return reflect.Zero(valType), fmt.Errorf("list type is not a slice of %s: %s", elemType.Elem(), currentList.Type().Elem())
	}
	// Create a slice of the correct type
	slice := reflect.MakeSlice(elemType, currentList.Len()+1, currentList.Len()+1)
	for idx := 0; idx < currentList.Len(); idx++ {
		slice.Index(idx).Set(currentList.Index(idx))
	}
	err := strToValue(slice.Index(currentList.Len()), value)
	if err != nil {
		return reflect.Zero(valType), err
	}
	return slice, nil
}

// Increment the value provided by the amount provided, starting from zero if the value is an invalid type.
func incrementNumber(valType reflect.Type, currentValue reflect.Value, amount float64) (reflect.Value, error) {
	if currentValue.Kind() == reflect.Ptr {
		currentValue = currentValue.Elem()
	}
	if currentValue.Kind() == reflect.Invalid {
		currentValue = reflect.New(valType).Elem()
	}
	switch currentValue.Kind() {
	case reflect.Int:
		currentValue = reflect.ValueOf(int(currentValue.Int() + int64(amount)))
	case reflect.Float64:
		currentValue.SetFloat(currentValue.Float() + amount)
	default:
		return currentValue, fmt.Errorf("cannot increment type %s", currentValue.Kind())
	}
	return currentValue, nil
}

// Old  method, takes a type not a value
func strToType(valType reflect.Type, str string) (reflect.Value, error) {

	elemType := valType
	if valType.Kind() == reflect.Pointer {
		elemType = valType.Elem()
	}
	value := reflect.New(elemType)
	err := strToValue(value, str)
	if err != nil {
		return reflect.Zero(valType), err
	}
	// If it's a pointer, then we need to dereference it.
	if valType.Kind() == reflect.Ptr {
		return value, nil
	} else {
		return value.Elem(), nil
	}
}
