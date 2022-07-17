package xapi

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// Raw-String literal.
const RawStrMark = "R"

// ToStr returns specified literal as X string literal for cxx.
func ToStr(bytes []byte) string {
	var cxx strings.Builder
	cxx.WriteString("str_xt{\"")
	cxx.WriteString(bytesToStr(bytes))
	cxx.WriteString("\"}")
	return cxx.String()
}

// ToRawStr returns specified literal as X raw-string literal for cxx.
func ToRawStr(bytes []byte) string {
	var cxx strings.Builder
	cxx.WriteString("str_xt{")
	cxx.WriteString(RawStrMark)
	cxx.WriteString("\"(")
	cxx.WriteString(bytesToStr(bytes))
	cxx.WriteString(")\"}")
	return cxx.String()
}

// ToChar returns specified literal as X rune literal for cxx.
func ToChar(b byte) string {
	return "'" + string(b) + "'"
}

// ToRune returns specified literal as X rune literal for cxx.
func ToRune(bytes []byte) string {
	if len(bytes) == 0 {
		return ""
	} else if bytes[0] == '\\' {
		if len(bytes) > 1 && (bytes[1] == 'u' || bytes[1] == 'U') {
			bytes = bytes[2:]
			i, _ := strconv.ParseInt(string(bytes), 16, 32)
			return "0x" + strconv.FormatInt(i, 16)
		}
	}
	r, _ := utf8.DecodeRune(bytes)
	return "0x" + strconv.FormatInt(int64(r), 16)
}

func btoa(b byte) string {
	if b <= 127 { // ASCII
		return string(b)
	}
	return "\\" + strconv.FormatUint(uint64(b), 8)
}

func bytesToStr(bytes []byte) string {
	var str strings.Builder
	for _, b := range bytes {
		str.WriteString(btoa(b))
	}
	return str.String()
}
