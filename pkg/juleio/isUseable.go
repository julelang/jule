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
func IsPassFileAnnotation(p string) bool {
	p = filepath.Base(p)
	n := len(p)
	p = p[:n-len(filepath.Ext(p))]

	// a1 is the second annotation.
	// Should be architecture annotation if exist annotation 2,
	// can operating system or architecture annotation if not.
	a1 := ""
	// a2 is first filter.
	// Should be operating system filter if exist and valid annotation.
	a2 := ""

	// Annotation 1
	i := strings.LastIndexByte(p, '_')
	if i == -1 {
		return true
	}
	if i+1 >= n {
		return true
	}
	a1 = p[i+1:]

	p = p[:i]

	// Annotation 2
	i = strings.LastIndexByte(p, '_')
	if i != -1 {
		a2 = p[i+1:]
	}

	
	if a2 == "" {
		ok, exist := checkPlatform(a1)
		if exist && !ok {
			return false
		}
		ok, exist = checkArch(a1)
		return !exist || ok
	}
	
	ok, exist := checkArch(a1)
	if exist && !ok {
		return false
	}
	ok, exist = checkPlatform(a2)
	return !exist || ok
}
