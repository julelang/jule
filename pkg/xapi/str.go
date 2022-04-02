package xapi

import "strings"

// Raw-String literal.
const RawStrMark = "R"

// ToStr returns specified literal as X string literal for cxx.
func ToStr(literal string) string {
	var cxx strings.Builder
	cxx.WriteString("str(")
	cxx.WriteString(literal)
	cxx.WriteByte(')')
	return cxx.String()
}

// ToRawStr returns specified literal as X raw-string literal for cxx.
func ToRawStr(literal string) string {
	var cxx strings.Builder
	cxx.WriteString("str(")
	cxx.WriteString(RawStrMark)
	cxx.WriteString(literal)
	cxx.WriteByte(')')
	return cxx.String()
}

// ToRune returns specified literal as X rune literal for cxx.
func ToRune(literal string) string {
	var cxx strings.Builder
	cxx.WriteString(literal)
	return cxx.String()
}
