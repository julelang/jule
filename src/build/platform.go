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

// List of supported architectures.
var DISTARCH = []string{
	ARCH_ARM,
	ARCH_ARM64,
	ARCH_AMD64,
	ARCH_I386,
}

const goos_windows = "windows"
const goos_darwin = "darwin"
const goos_linux = "linux"

const goarch_i386 = "386"
const goarch_amd64 = "amd64"
const goarch_arm = "arm"
const goarch_arm64 = "arm64"

// IsWindows reports whether os is windows.
func IsWindows(os string) bool { return os == goos_windows }

// IsDarwin reports whether os is darwin.
func IsDarwin(os string) bool { return os == goos_darwin }

// IsLinux reports whether os is linux.
func IsLinux(os string) bool { return os == goos_linux }

// IsUnix reports whether os is unix.
func IsUnix(os string) bool {
	return IsDarwin(os) || IsLinux(os)
}

// IsI386 reports whether architecture is i386.
func IsI386(arch string) bool { return arch == goarch_i386 }

// IsAmd64 reports whether architecture is amd64.
func IsAmd64(arch string) bool { return arch == goarch_amd64 }

// IsArm reports whether architecture is arm.
func IsArm(arch string) bool { return arch == goarch_arm }

// IsArm64 reports whether architecture is arm64.
func IsArm64(arch string) bool { return arch == goarch_arm64 }

// IsX32 reports whether architecture is 32 bit.
func IsX32(arch string) bool {
	return IsI386(arch) || IsArm(arch)
}

// IsX64 reports whether architecture is 64 bit.
func IsX64(arch string) bool {
	return IsAmd64(arch) || IsArm64(arch)
}
