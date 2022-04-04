package xapi

import "strings"

// CxxIgnore is the ignoring of cxx.
const CxxIgnore = "std::ignore"

// ToDefer returns cxx of deferred function call expression string.
func ToDefer(expr string) string {
	var cxx strings.Builder
	cxx.WriteString("DEFER(")
	cxx.WriteString(expr)
	cxx.WriteString(");")
	return cxx.String()
}
