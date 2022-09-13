package juleio

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jule-lang/jule/pkg/jule"
)

const os_windows = "windows"
const os_darwin = "darwin"
const os_linux = "linux"

const arch_i386 = "386"
const arch_amd64 = "amd64"
const arch_arm = "arm"
const arch_arm64 = "arm64"

func checkPlatform(path string) (ok bool, exist bool) {
	ok = false
	exist = true
	switch path {
	case jule.PlatformWindows:
		ok = runtime.GOOS == os_windows
	case jule.PlatformDarwin:
		ok = runtime.GOOS == os_darwin
	case jule.PlatformLinux:
		ok = runtime.GOOS == os_linux
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
	case jule.ArchI386:
		ok = runtime.GOARCH == arch_i386
	case jule.ArchAmd64:
		ok = runtime.GOARCH == arch_amd64
	case jule.ArchArm:
		ok = runtime.GOARCH == arch_arm
	case jule.ArchArm64:
		ok = runtime.GOARCH == arch_arm64
	default:
		ok = true
		exist = false
	}
	return
}

// IsPassFileAnnotation returns true
// if file path is passes file annotation,
// returns false if not.
func IsPassFileAnnotation(path string) bool {
	path = filepath.Base(path)
	path = path[:len(path)-len(filepath.Ext(path))]

	// Filter 1
	i := strings.LastIndexByte(path, '_')
	if i == -1 {
		return true
	}
	length := len(path)
	if i+1 >= length {
		return true
	}
	filter := path[i+1:]
	ok, exist := checkPlatform(filter)
	if exist && !ok {
		return false
	}
	ok, exist = checkArch(filter)
	if !exist {
		return true
	} else if !ok {
		return false
	}

	// Filter 2
	path = path[:i]
	i = strings.LastIndexByte(path, '_')
	if i == -1 {
		return true
	}
	if i+1 >= length {
		return true
	}
	filter = path[i+1:]
	ok, exist = checkPlatform(filter)
	if exist && !ok {
		return false
	}
	ok, exist = checkArch(filter)
	return !exist || ok
}
