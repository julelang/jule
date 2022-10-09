package juleapi

import "strings"

// JULEC_HEADER is the header path of "julec.hpp"
var JULEC_HEADER = ""

// CPP_IGNORE is the ignoring of cpp.
const CPP_IGNORE = "std::ignore"

// SELF is the self keyword equavalent of cpp.
const SELF = "this"

// ToDeferredCall returns cpp of deferred function call expression string.
func ToDeferredCall(expr string) string {
	var cpp strings.Builder
	cpp.WriteString("__JULEC_DEFER(")
	cpp.WriteString(expr)
	cpp.WriteString(");")
	return cpp.String()
}

// ToConcurrentCall returns cpp of concurrent function call expression string.
func ToConcurrentCall(expr string) string {
	var cpp strings.Builder
	cpp.WriteString("__JULEC_CO(")
	cpp.WriteString(expr)
	cpp.WriteString(");")
	return cpp.String()
}
