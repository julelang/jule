package x

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
	XSettings      *XSet
)

// IsIgnoreName reports name is ignore or not.
func IsIgnoreName(name string) bool {
	return name == "__"
}
