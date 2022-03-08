package xapi

import "strings"

// Returns type as X heap-allocation expression for cxx.
func ToXAlloc(t string) string {
	var cxx strings.Builder
	cxx.WriteString("XALLOC(")
	cxx.WriteString(t)
	cxx.WriteByte(')')
	return cxx.String()
}
