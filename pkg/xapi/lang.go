package xapi

import "strings"

// CxxIgnore is the ignoring of cxx.
const CxxIgnore = "std::ignore"

// ToDeferredCall returns cxx of deferred function call expression string.
func ToDeferredCall(expr string) string {
	var cxx strings.Builder
	cxx.WriteString("DEFER(")
	cxx.WriteString(expr)
	cxx.WriteString(");")
	return cxx.String()
}

func ToConcurrentCall(expr string) string {
	var cxx strings.Builder
	cxx.WriteString("CO(")
	cxx.WriteString(expr)
	cxx.WriteString(");")
	return cxx.String()
}
