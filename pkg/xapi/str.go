package xapi

import (
	"encoding/hex"
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
	return strconv.Itoa(int(b))
}

// ToRune returns specified literal as X rune literal for cxx.
func ToRune(bytes []byte) string {
	r, _ := utf8.DecodeRune(bytes)
	return strconv.FormatInt(int64(r), 10)
}

func bytesToStr(bytes []byte) string {
	var str strings.Builder
	for _, b := range bytes {
		if b <= 127 { // ASCII
			str.WriteByte(b)
		} else {
			str.WriteString("\\x")
			str.WriteString(hex.EncodeToString([]byte{b}))
		}
	}
	return str.String()
}
