package build

import (
	"runtime"
	"strings"
	"path/filepath"
)

func check_os(arg string) (ok bool, exist bool) {
	ok = false
	exist = true
	switch arg {
	case OS_WINDOWS:
		ok = Is_windows(runtime.GOOS)
	case OS_DARWIN:
		ok = Is_darwin(runtime.GOOS)
	case OS_LINUX:
		ok = Is_linux(runtime.GOOS)
	case OS_UNIX:
		ok = Is_unix(runtime.GOOS)
	default:
		ok = true
		exist = false
	}
	return
}

func check_arch(arg string) (ok bool, exist bool) {
	ok = false
	exist = true
	switch arg {
	case ARCH_I386:
		ok = Is_i386(runtime.GOARCH)
	case ARCH_AMD64:
		ok = Is_amd64(runtime.GOARCH)
	case ARCH_ARM:
		ok = Is_arm(runtime.GOARCH)
	case ARCH_ARM64:
		ok = Is_arm64(runtime.GOARCH)
	case ARCH_64Bit:
		ok = Is_64bit(runtime.GOARCH)
	case ARCH_32Bit:
		ok = Is_32bit(runtime.GOARCH)
	default:
		ok = true
		exist = false
	}
	return
}

// Reports whether file path passes file annotation by current system.
func Is_pass_file_annotation(p string) bool {
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
		// Check file name directly if not exist any _ character.
		ok, exist := check_os(p)
		if exist {
			return ok
		}
		ok, exist = check_arch(p)
		return !exist || ok
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
