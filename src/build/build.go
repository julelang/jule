package build

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/julelang/jule"
)

// This attributes should be added to the attribute map.
const ATTR_CDEF = "cdef"
const ATTR_TYPEDEF = "typedef"

// ATTRS is list of all attributes.
var ATTRS = [...]string{
	ATTR_CDEF,
	ATTR_TYPEDEF,
}

func check_os(arg string) (ok bool, exist bool) {
	ok = false
	exist = true
	switch arg {
	case OS_WINDOWS:
		ok = IsWindows(runtime.GOOS)
	case OS_DARWIN:
		ok = IsDarwin(runtime.GOOS)
	case OS_LINUX:
		ok = IsLinux(runtime.GOOS)
	case OS_UNIX:
		ok = IsUnix(runtime.GOOS)
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
		ok = IsI386(runtime.GOARCH)
	case ARCH_AMD64:
		ok = IsAmd64(runtime.GOARCH)
	case ARCH_ARM:
		ok = IsArm(runtime.GOARCH)
	case ARCH_ARM64:
		ok = IsArm64(runtime.GOARCH)
	case ARCH_64Bit:
		ok = IsX64(runtime.GOARCH)
	case ARCH_32Bit:
		ok = IsX32(runtime.GOARCH)
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

// IsStdHeaderPath reports path is C++ std library path.
func IsStdHeaderPath(p string) bool {
	return p[0] == '<' && p[len(p)-1] == '>'
}

// CPP_HEADER_EXTS are valid extensions of cpp headers.
var CPP_HEADER_EXTS = []string{
	".h",
	".hpp",
	".hxx",
	".hh",
}

// IsValidHeader returns true if given extension is valid, false if not.
func IsValidHeader(ext string) bool {
	for _, validExt := range CPP_HEADER_EXTS {
		if ext == validExt {
			return true
		}
	}
	return false
}

// IsJule reports whether file path is Jule source code.
// Returns false if error occur.
func IsJule(path string) bool {
	abs, err := filepath.Abs(path)
	return err == nil && filepath.Ext(abs) == jule.EXT
}
