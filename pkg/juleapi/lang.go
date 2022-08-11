package juleapi

import "strings"

// JuleCHeader is the header path of "julec.hpp"
var JuleCHeader = ""

// CppIgnore is the ignoring of cpp.
const CppIgnore = "std::ignore"

// CppSelf is the self keyword equavalent of cpp.
const CppSelf = "this"

// ToDeferredCall returns cpp of deferred function call expression string.
func ToDeferredCall(expr string) string {
	var cpp strings.Builder
	cpp.WriteString("DEFER(")
	cpp.WriteString(expr)
	cpp.WriteString(");")
	return cpp.String()
}

// ToConcurrentCall returns cpp of concurrent function call expression string.
func ToConcurrentCall(expr string) string {
	var cpp strings.Builder
	cpp.WriteString("CO(")
	cpp.WriteString(expr)
	cpp.WriteString(");")
	return cpp.String()
}
