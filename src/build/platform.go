package build

const OS_WINDOWS = "windows"
const OS_LINUX = "linux"
const OS_DARWIN = "darwin"
const OS_UNIX = "unix"

const ARCH_ARM = "arm"
const ARCH_ARM64 = "arm64"
const ARCH_AMD64 = "amd64"
const ARCH_I386 = "i386"
const ARCH_64Bit = "64bit"
const ARCH_32Bit = "32bit"

// List of supported operating systems.
var DISTOS = []string{
	OS_WINDOWS,
	OS_LINUX,
	OS_DARWIN,
}

// List of supported architects.
var DISTARCH = []string{
	ARCH_ARM,
	ARCH_ARM64,
	ARCH_AMD64,
	ARCH_I386,
}
