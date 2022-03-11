package x

import "github.com/the-xlang/x/pkg/x/xset"

// X constants.
const (
	Version         = `@developer_beta 0.0.1`
	SourceExtension = `.xx`
	DocExtension    = ".xdoc"
	SettingsFile    = "x.set"
	StdlibName      = "lib"

	EntryPoint = "main"
)

// Environment Variables.
var (
	StdlibPath     string
	ExecutablePath string
	XSet           *xset.XSet
)
