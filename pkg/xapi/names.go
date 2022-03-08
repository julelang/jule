package xapi

// Ignore operator.
const Ignore = "_"

// IsIgnoreId reports identifier is ignore or not.
func IsIgnoreId(name string) bool {
	return name == Ignore
}

// Returns specified identifer as X identifer.
func AsId(name string) string {
	return Ignore + name
}
