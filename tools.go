package configape

import (
	"fmt"
	"strings"
	"unicode"
)

func debugf(format string, args ...interface{}) {
	// Print the string with "DEBUG: " prepended
	fmt.Print("DEBUG: ")
	fmt.Printf(format, args...)
}

// Split a comma separated name=value string into a map. The value can be quoted with single or double
// quotes and the comma inside the quotes will not be used as a separator, and the quotes will be
// stripped from the value.
func stringListToMap(str string) map[string]string {
	m := make(map[string]string)

	for _, s := range splitPreservingQuotes(str) {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) == 2 {
			val := parts[1]
			// Strip surrounding single or double quotes
			if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
				val = val[1 : len(val)-1]
			}
			m[parts[0]] = val
		} else {
			m[parts[0]] = ""
		}
	}
	return m
}

// Splits the string on commas, but preserves any quoted strings
func splitPreservingQuotes(str string) []string {
	var result []string
	var current string
	inQuotes := false

	for _, c := range str {
		if c == '"' || c == '\'' {
			inQuotes = !inQuotes
		} else if c == ',' && !inQuotes {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}

	return result
}

func camelCaseToUnderscore(name string) string {
	return camelCaseConvert(name, '_')
}
func camelCaseToDash(name string) string {
	return camelCaseConvert(name, '-')
}

// Take a CamelCase name and convert it to a hyphen-name, if there is a double
// uppercase, then don't split it. Eg DDoubleUppercase -> ddouble-uppercase
// and testURL -> test-url
func camelCaseConvert(input string, separator rune) string {
	var camel strings.Builder
	camel.Grow(len(input) + 5) // Preallocate some space
	prevCharIsUpper := false

	for i, char := range input {
		if i == 0 && unicode.IsUpper(char) {
			// If the first character is uppercase, add the lowercase version, but
			// don't prefix it with a hyphen
			camel.WriteRune(unicode.ToLower(char))
			prevCharIsUpper = true
		} else if i > 0 && unicode.IsUpper(char) {
			if !prevCharIsUpper {
				// If the character is uppercase and not the first character,
				// add a hyphen followed by its lowercase equivalent.
				camel.WriteRune(separator)
			}
			camel.WriteRune(unicode.ToLower(char))
			prevCharIsUpper = true
		} else {
			// Otherwise, add the character as is and reset prevCharIsUpper.
			camel.WriteRune(char)
			prevCharIsUpper = false
		}
	}

	return camel.String()
}
