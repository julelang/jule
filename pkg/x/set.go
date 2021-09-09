package x

import (
	"errors"
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

func splitLines(content string) []string {
	if runtime.GOOS == "windows" {
		return strings.SplitN(string(content), "\n", -1)
	}
	return strings.SplitN(string(content), "\n\r", -1)
}

func (xs *XSet) checkUnset() []error {
	var errs []error
	for key := range xs.Fields {
		if xs.Fields[key] == "" {
			errs = append(errs, errors.New("\""+key+"\" is not defined!"))
		}
	}
	return errs
}

// Parse is parse x.set file content to fields.
func (xs *XSet) Parse(content []byte) []error {
	lines := splitLines(string(content))
	for index, line := range lines {
		line = strings.TrimFunc(line, unicode.IsSpace)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", -1)
		if len(parts) < 2 {
			return []error{fmt.Errorf("invalid syntax at line %d", index+1)}
		}
		key, value := parts[0], parts[1]
		_, ok := xs.Fields[key]
		if !ok {
			return []error{fmt.Errorf("invalid field at line %d", index+1)}
		}
		switch key {
		case "out_name":
			if len(parts) > 2 {
				return []error{fmt.Errorf("invalid value at line %d", index+1)}
			}
		}
		xs.Fields[key] = value
	}
	return xs.checkUnset()
}
