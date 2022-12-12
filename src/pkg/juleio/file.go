package juleio

import "path/filepath"

// File instance of fs.
type File struct {
	Dir  string
	Name string
	Data []rune
}

// Path returns full path of file.
func (f *File) Path() string {
	return filepath.Join(f.Dir, f.Name)
}

// IsStdHeaderPath reports path is C++ std library path.
func IsStdHeaderPath(p string) bool {
	return p[0] == '<' && p[len(p)-1] == '>'
}