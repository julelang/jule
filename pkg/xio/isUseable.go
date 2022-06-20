package xio

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/the-xlang/xxc/pkg/x"
)

func checkPlatform(path string) (ok bool, exist bool) {
	ok = false
	exist = true
	switch path {
	case x.PlatformWindows:
		ok = runtime.GOOS == "windows"
	case x.PlatformDarwin:
		ok = runtime.GOOS == "darwin"
	case x.PlatformLinux:
		ok = runtime.GOOS == "linux"
	default:
		ok = true
		exist = false
	}
	return
}

func checkArch(path string) (ok bool, exist bool) {
	ok = false
	exist = true
	switch path {
	case x.ArchI386:
		ok = runtime.GOARCH == "386"
	case x.ArchAmd64:
		ok = runtime.GOARCH == "amd64"
	case x.ArchArm:
		ok = runtime.GOARCH == "arm"
	case x.ArchArm64:
		ok = runtime.GOARCH == "arm64"
	default:
		ok = true
		exist = false
	}
	return
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
	ok, exist := checkPlatform(path)
	if exist {
		return ok
	}
	ok, _ = checkArch(path)
	return ok
}
