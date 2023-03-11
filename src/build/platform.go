package build

const OS_WINDOWS = "windows" // File annotation kind for windows operating system.
const OS_LINUX = "linux"     // File annotation kind for linux operating system.
const OS_DARWIN = "darwin"   // File annotation kind for darwin operating system.
const OS_UNIX = "unix"       // File annotation kind for unix operating systems.

const ARCH_ARM = "arm"     // File annotation kind for arm architecture.
const ARCH_ARM64 = "arm64" // File annotation kind for arm64 architecture.
const ARCH_AMD64 = "amd64" // File annotation kind for amd64 architecture.
const ARCH_I386 = "i386"   // File annotation kind for intel 386 architecture.
const ARCH_64Bit = "64bit" // File annotation kind for 64-bit architectures.
const ARCH_32Bit = "32bit" // File annotation kind for 32-bit architectures.

// List of supported operating systems.
var DISTOS = []string{
	OS_WINDOWS,
	OS_LINUX,
	OS_DARWIN,
}

// List of supported architectures.
var DISTARCH = []string{
	ARCH_ARM,
	ARCH_ARM64,
	ARCH_AMD64,
	ARCH_I386,
}

// List of all possible runtime.GOOS values:
const _RUNTIME_OS_WINDOWS = "windows"
const _RUNTIME_OS_DARWIN = "darwin"
const _RUNTIME_OS_LINUX = "linux"

// List of all possible runtime.GOARCH values:
const _RUNTIME_ARCH_I386 = "386"
const _RUNTIME_ARCH_AMD64 = "amd64"
const _RUNTIME_ARCH_ARM = "arm"
const _RUNTIME_ARCH_ARM64 = "arm64"

// Reports whether os is windows.
func Is_windows(os string) bool { return os == _RUNTIME_OS_WINDOWS }
// Reports whether os is darwin.
func Is_darwin(os string) bool { return os == _RUNTIME_OS_DARWIN }
// Reports whether os is linux.
func Is_linux(os string) bool { return os == _RUNTIME_OS_LINUX }
// Reports whether architecture is intel 386.
func Is_i386(arch string) bool { return arch == _RUNTIME_ARCH_I386 }
// Reports whether architecture is amd64.
func Is_amd64(arch string) bool { return arch == _RUNTIME_ARCH_AMD64 }
// Reports whether architecture is arm.
func Is_arm(arch string) bool { return arch == _RUNTIME_ARCH_ARM }
// Is_arm64 reports whether architecture is arm64.
func Is_arm64(arch string) bool { return arch == _RUNTIME_ARCH_ARM64 }

// Reports whether os is unix.
func Is_unix(os string) bool {
	return Is_darwin(os) || Is_linux(os)
}

// Reports whether architecture is 32-bit.
func Is_32bit(arch string) bool {
	return Is_i386(arch) || Is_arm(arch)
}

// Reports whether architecture is 64-bit.
func Is_64bit(arch string) bool {
	return Is_amd64(arch) || Is_arm64(arch)
}
