package x

import "github.com/the-xlang/xxc/pkg/xset"

// X constants.
const (
	Version      = `@developer_beta 0.0.1`
	SrcExt       = `.xx`
	DocExt       = ".xdoc"
	SettingsFile = "x.set"
	StdlibName   = "lib"

	EntryPoint = "main"
)

// Environment Variables.
var (
	StdlibPath string
	ExecPath   string
	XSet       *xset.XSet
)
