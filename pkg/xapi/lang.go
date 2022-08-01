package xapi

import "strings"

// XXCHeader is the header path of "xxc.hpp"
var XXCHeader = ""

// CxxIgnore is the ignoring of cxx.
const CxxIgnore = "std::ignore"

// CxxSelf is the self keyword equavalent of cxx.
const CxxSelf = "this"

// ToDeferredCall returns cxx of deferred function call expression string.
func ToDeferredCall(expr string) string {
	var cxx strings.Builder
	cxx.WriteString("DEFER(")
	cxx.WriteString(expr)
	cxx.WriteString(");")
	return cxx.String()
}

// ToConcurrentCall returns cxx of concurrent function call expression string.
func ToConcurrentCall(expr string) string {
	var cxx strings.Builder
	cxx.WriteString("CO(")
	cxx.WriteString(expr)
	cxx.WriteString(");")
	return cxx.String()
}
