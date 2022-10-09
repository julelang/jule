package juleio

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/jule-lang/jule/pkg/jule"
)

// OpenJuleF returns Jule source file.
func OpenJuleF(path string) (*File, error) {
	path, _ = filepath.Abs(path)
	if filepath.Ext(path) != jule.SRC_EXT {
		return nil, errors.New(jule.GetError("file_not_jule", path))
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
