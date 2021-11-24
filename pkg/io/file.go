package io

import "os"

// FILE is source file.
type FILE struct {
	Path    string
	Content []rune
}

// WriteFileTruncate writes file truncate
// by specified path and content.
func WriteFileTruncate(path string, content []byte) error {
	if _, err := os.Open(path); err == nil {
		os.Remove(path)
	}
	return os.WriteFile(path, content, 0x025E)
}
