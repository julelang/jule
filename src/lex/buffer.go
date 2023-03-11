package lex

import (
	"os"
	"path/filepath"
	"unsafe"
)

// Fileset for lexing.
type File struct {
	_path  string
	tokens []Token
}

// Reports whether file path is exist and accessible.
func (f *File) Is_ok() bool {
	_, err := os.Stat(f._path)
	return err == nil
}

// Returns path.
func (f *File) Path() string { return f._path }

// Returns directory of file's path.
func (f *File) Dir() string { return filepath.Dir(f._path) }

// Returns filename.
func (f *File) Name() string { return filepath.Base(f._path) }

// Returns self as uintptr.
func (f *File) Addr() uintptr { return uintptr(unsafe.Pointer(f)) }

// Returns tokens.
// Copies into new slice.
func (f *File) Tokens() []Token {
	if f.tokens == nil {
		return nil
	}
	tokens := make([]Token, len(f.tokens))
	_ = copy(tokens, f.tokens)
	return tokens
}

// Returns new File points to Jule file.
func New_file_set(path string) *File {
	return &File{
		_path:  path,
		tokens: nil,
	}
}
