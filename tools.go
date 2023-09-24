package configape

import "fmt"

func debugf(format string, args ...interface{}) {
	// Print the string with "DEBUG: " prepended
	fmt.Print("DEBUG: ")
	fmt.Printf(format, args...)
}
