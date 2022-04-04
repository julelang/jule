package xapi

// Ignore operator.
const Ignore = "_"

// IsIgnoreId reports identifier is ignore or not.
func IsIgnoreId(id string) bool { return id == Ignore }

// Returns specified identifer as X identifer.
func AsId(id string) string { return "XID(" + id + ")" }
