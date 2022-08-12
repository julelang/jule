package jule

import "github.com/jule-lang/jule/pkg/juleset"

// X constants.
const (
	Version       = `@developer_beta 0.0.1`
	SrcExt        = `.jule`
	DocExt        = SrcExt + "doc"
	SettingsFile  = "jule.set"
	Stdlib        = "std"
	Localizations = "localization"

	EntryPoint          = "main"
	InitializerFunction = "init"

	Anonymous = "<anonymous>"

	DocPrefix = "doc:"

	PlatformWindows = "windows"
	PlatformLinux   = "linux"
	PlatformDarwin  = "darwin"

	ArchArm   = "arm"
	ArchArm64 = "arm64"
	ArchAmd64 = "amd64"
	ArchI386  = "i386"

	Attribute_Inline  = "inline"
	Attribute_TypeArg = "typearg"

	PreprocessorDirective      = "pragma"
	PreprocessorDirectiveEnofi = "enofi"

	Mark_Array = "..."

	Prefix_Slice = "[]"
	Prefix_Array = "[" + Mark_Array + "]"
)

// Environment Variables.
var (
	LangsPath  string
	StdlibPath string
	ExecPath   string
	Set        *juleset.Set
)
