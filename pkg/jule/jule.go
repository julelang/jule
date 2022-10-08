package jule

import "github.com/jule-lang/jule/pkg/juleset"

// Jule constants.
const (
	Version       = `@development_channel`
	SrcExt        = `.jule`
	DocExt        = SrcExt + "doc"
	SettingsFile  = "jule.set"
	Stdlib        = "std"
	Localizations = "localization"

	EntryPoint          = "main"
	InitializerFunction = "init"

	Anonymous = "<anonymous>"

	CommentPragmaSeparator = ":"
	PragmaCommentPrefix    = "jule" + CommentPragmaSeparator

	PlatformWindows = "windows"
	PlatformLinux   = "linux"
	PlatformDarwin  = "darwin"
	PlatformUnix    = "unix"

	ArchArm   = "arm"
	ArchArm64 = "arm64"
	ArchAmd64 = "amd64"
	ArchI386  = "i386"
	Arch64Bit = "64bit"
	Arch32Bit = "32bit"

	// This attributes should be added to the attribute map.
	Attribute_CDef    = "cdef"
	Attribute_Typedef = "typedef"

	PreprocessorDirective      = "pragma"
	PreprocessorDirectiveEnofi = "enofi"

	Mark_Array = "..."

	Prefix_Slice = "[]"
	Prefix_Array = "[" + Mark_Array + "]"

	CompilerGCC   = "gcc"
	CompilerClang = "clang"
)

// Environment Variables.
var (
	LangsPath   string
	StdlibPath  string
	ExecPath    string
	WorkingPath string
	Set         *juleset.Set
)
