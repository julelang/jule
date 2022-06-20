package xio

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/the-xlang/xxc/pkg/x"
)

// File instance of fs.
type File struct {
	Path string
	Data []rune
}

// IsUseable returns true if file path is useable,
// returns false if not.
func IsUseable(path string) bool {
	path = filepath.Base(path)
	path = path[:len(path)-len(filepath.Ext(path))]
	index := strings.LastIndexByte(path, '_')
	if index == -1 {
		return true
	}
	path = path[index+1:]
	switch path {
	case x.PlatformWindows:
		return runtime.GOOS == "windows"
	case x.PlatformDarwin:
		return runtime.GOOS == "darwin"
	case x.PlatformLinux:
		return runtime.GOOS == "linux"
	default:
		return true
	}
}
