package x

import "github.com/the-xlang/x/pkg/x/xset"

// X constants.
const (
	Version      = `@developer_beta 0.0.1`
	Extension    = `.xx`
	SettingsFile = "x.set"

	EntryPoint = "main"
)

// Environment Variables.
var (
	ExecutablePath string
	XSet           *xset.XSet
)

// IsIgnoreName reports name is ignore or not.
func IsIgnoreName(name string) bool {
	return name == "__"
}
