package lex

import (
	"path/filepath"
	"unsafe"

	"github.com/julelang/jule"
	"github.com/julelang/jule/build"
)

// File instance of fs.
type File struct {
	_path string
}

// NewFile returns new File points to Jule file.
func NewFile(path string) *File {
	abs, _ := filepath.Abs(path)
	if filepath.Ext(abs) != jule.SRC_EXT {
		panic(build.GetError("file_not_jule", path))
	}
	return &File{path}
}

// Path returns full path.
func (f *File) Path() string { return f._path }

// Dir returns directory.
func (f *File) Dir() string { return filepath.Dir(f._path) }

// Name returns filename.
func (f *File) Name() string { return filepath.Base(f._path) }

// Addr returns uintptr(unsafe.Pointer(f)).
func (f *File) Addr() uintptr { return uintptr(unsafe.Pointer(f)) }
