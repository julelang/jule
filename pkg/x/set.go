package x

import (
	"fmt"
	"runtime"
	"strings"
	"unicode"
)

// XSet is parser and decomposer of x.set files.
type XSet struct {
	Fields map[string]string
}

// NewXSet XSet instance.
func NewXSet() *XSet {
	xs := new(XSet)
	xs.Fields = make(map[string]string)
	xs.Fields["out_dir"] = ""
	xs.Fields["out_name"] = ""
	return xs
}

// Parse is parse x.set file content to fields.
func (xs *XSet) Parse(content []byte) error {
	var lines []string
	if runtime.GOOS == "windows" {
		lines = strings.SplitN(string(content), "\n", -1)
	} else {
		lines = strings.SplitN(string(content), "\n\r", -1)
	}
	for index, line := range lines {
		line = strings.TrimFunc(line, unicode.IsSpace)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", -1)
		if len(parts) < 2 {
			return fmt.Errorf("invalid syntax at line %d", index+1)
		}
		key, value := parts[0], parts[1]
		_, ok := xs.Fields[key]
		if !ok {
			return fmt.Errorf("invalid field at line %d", index+1)
		}
		switch key {
		case "out_name":
			if len(parts) > 2 {
				return fmt.Errorf("invalid value at line %d", index+1)
			}
		}
		xs.Fields[value] = value
	}
	return nil
}
