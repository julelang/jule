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

// IsIgnoreId reports identifier is ignore or not.
func IsIgnoreId(name string) bool {
	return name == "__"
}

// Returns specified identifer as X identifer.
func AsId(name string) string {
	return "_" + name
}

// Returns normalized identifer of specified X identifer.
//
// Special case is:
//  NormalizeId(name) -> "" if name is empty
func NormalizeId(name string) string {
	if name == "" {
		return ""
	}
	return name[1:]
}
