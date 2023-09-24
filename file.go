package configape

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (c *cfgApe) parseConfigFile(cfgFile string) error {
	var fh io.Reader
	var err error
	fileType := "json"
	if c.options.ConfigFileType != "" {
		fileType = c.options.ConfigFileType
	}
	// Set the fileType to the file extension if there is an extension
	ext := filepath.Ext(cfgFile)
	if ext != "" {
		fileType = ext[1:]
	}

	if c.options.cfgFileContents != "" {
		// We instead use this string as the config file
		// Make an io Reader that reads from a string
		fh = strings.NewReader(c.options.cfgFileContents)
	} else {
		fh, err = os.Open(cfgFile)
		if err != nil {
			// Unable to find file is not an error
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		defer fh.(io.Closer).Close()

	}
	switch fileType {
	case "json":
		return c.parseJsonConfigFile(cfgFile, fh)
	case "yaml":
		return c.parseYamlConfigFile(cfgFile, fh)
	default:
		return fmt.Errorf("unknown config file type: %s", fileType)
	}
}
