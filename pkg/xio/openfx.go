package xio

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/the-xlang/x/pkg/x"
)

// Openfx returns X source file.
func Openfx(path string) (*File, error) {
	if filepath.Ext(path) != x.SrcExt {
		return nil, errors.New(x.Errors[`file_not_x`] + path)
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f := new(File)
	f.Path = path
	f.Content = []rune(string(bytes))
	return f, nil
}
