package io

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/the-xlang/x/pkg/x"
)

// GetX returns X source file.
func GetX(path string) (*FILE, error) {
	if filepath.Ext(path) != x.Extension {
		return nil, errors.New(x.Errors[`file_not_x`] + path)
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f := new(FILE)
	f.Path = path
	f.Content = []rune(string(bytes))
	return f, nil
}
