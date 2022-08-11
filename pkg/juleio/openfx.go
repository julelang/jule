package juleio

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/jule-lang/jule/pkg/jule"
)

// Openfx returns X source file.
func Openfx(path string) (*File, error) {
	path, _ = filepath.Abs(path)
	if filepath.Ext(path) != jule.SrcExt {
		return nil, errors.New(jule.GetError("file_not_x", path))
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f := new(File)
	f.Dir, f.Name = filepath.Split(path)
	f.Data = []rune(string(bytes))
	return f, nil
}
