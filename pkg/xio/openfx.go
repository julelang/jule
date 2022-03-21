package xio

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/the-xlang/xxc/pkg/x"
)

// Openfx returns X source file.
func Openfx(path string) (*File, error) {
	if filepath.Ext(path) != x.SrcExt {
		return nil, errors.New(x.Errs[`file_not_x`] + path)
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f := new(File)
	f.Path = path
	f.Text = []rune(string(bytes))
	return f, nil
}
