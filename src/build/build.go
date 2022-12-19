package build

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/julelang/jule"
)

const os_windows = "windows"
const os_darwin = "darwin"
const os_linux = "linux"

const arch_i386 = "386"
const arch_amd64 = "amd64"
const arch_arm = "arm"
const arch_arm64 = "arm64"

func check_os(path string) (ok bool, exist bool) {
	ok = false
	exist = true
	switch path {
	case jule.OS_WINDOWS:
		ok = runtime.GOOS == os_windows
	case jule.OS_DARWIN:
		ok = runtime.GOOS == os_darwin
	case jule.OS_LINUX:
		ok = runtime.GOOS == os_linux
	case jule.OS_UNIX:
		switch runtime.GOOS {
		case os_darwin, os_linux:
			ok = true
		}
	default:
		ok = true
		exist = false
	}
	return
}

func check_arch(path string) (ok bool, exist bool) {
	ok = false
	exist = true
	switch path {
	case jule.ARCH_I386:
		ok = runtime.GOARCH == arch_i386
	case jule.ARCH_AMD64:
		ok = runtime.GOARCH == arch_amd64
	case jule.ARCH_ARM:
		ok = runtime.GOARCH == arch_arm
	case jule.ARCH_ARM64:
		ok = runtime.GOARCH == arch_arm64
	case jule.ARCH_64Bit:
		switch runtime.GOARCH {
		case arch_amd64, arch_arm64:
			ok = true
		}
	case jule.ARCH_32Bit:
		switch runtime.GOARCH {
		case arch_i386, arch_arm:
			ok = true
		}
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
	// Should be architecture annotation if exist annotation 2 (aka a2),
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
		ok, exist := check_os(a1)
		if exist {
			return ok
		}
		ok, exist = check_arch(a1)
		return !exist || ok
	}
	
	ok, exist := check_arch(a1)
	if exist {
		if !ok {
			return false
		}
		ok, exist = check_os(a2)
		return !exist || ok
	}

	// a1 is not architecture, for this reason bad couple pattern.
	// Accept as one pattern, so a1 can be platform.
	ok, exist = check_os(a1)
	return !exist || ok
}

// IsStdHeaderPath reports path is C++ std library path.
func IsStdHeaderPath(p string) bool {
	return p[0] == '<' && p[len(p)-1] == '>'
}
