package juleapi

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/jule-lang/jule/lex"
)

// String are generated as clean byte encoded, not string literal.
// Because X's strings are UTF-8 byte encoded and some
// C++ compilers compiles wrong C++ string literals.

// ToStr returns specified literal as X string literal for cpp.
func ToStr(bytes []byte) string {
	var cpp strings.Builder
	cpp.WriteString("str_julet{")
	btoa := bytesToStr(bytes)
	if btoa != "" {
		cpp.WriteByte('{')
		cpp.WriteString(btoa)
		cpp.WriteByte('}')
	}
	cpp.WriteString("}")
	return cpp.String()
}

// ToRawStr returns specified literal as X raw-string literal for cpp.
func ToRawStr(bytes []byte) string { return ToStr(bytes) }

// ToChar returns specified literal as X rune literal for cpp.
func ToChar(b byte) string { return btoa(b) }

// ToRune returns specified literal as X rune literal for cpp.
func ToRune(bytes []byte) string {
	if len(bytes) == 0 {
		return ""
	} else if bytes[0] == '\\' && len(bytes) > 1 {
		seq, ok := tryBtoaCommonEsq(bytes)
		if ok {
			return btoa(seq)
		}
		switch bytes[1] {
		case 'u', 'U':
			bytes = bytes[2:]
			i, _ := strconv.ParseInt(string(bytes), 16, 32)
			return "0x" + strconv.FormatInt(i, 16)
		}
	}
	r, _ := utf8.DecodeRune(bytes)
	return "0x" + strconv.FormatInt(int64(r), 16)
}

func btoa(b byte) string {
	return "0x" + strconv.FormatUint(uint64(b), 16)
}

func tryBtoaCommonEsq(bytes []byte) (seq byte, ok bool) {
	if len(bytes) < 2 || bytes[0] != '\\' {
		return
	}
	switch bytes[1] {
	case '\'':
		seq = '\''
	case '"':
		seq = '"'
	case 'a':
		seq = '\a'
	case 'b':
		seq = '\b'
	case 'f':
		seq = '\f'
	case 'n':
		seq = '\n'
	case 'r':
		seq = '\r'
	case 't':
		seq = '\t'
	case 'v':
		seq = 'v'
	}
	ok = seq != 0
	return
}

func byteSeq(bytes []byte, i int) (seq []byte, n int) {
	byten := len(bytes) - i
	switch {
	case byten == 1:
		n = 1
	case !lex.IsOctal(bytes[i+1]):
		n = 1
	case byten == 2:
		n = 2
	case !lex.IsOctal(bytes[i+2]):
		n = 2
	default:
		n = 3
	}
	seq = bytes[i : i+n]
	return
}

func strEsqSeq(bytes []byte, i *int) string {
	seq, ok := tryBtoaCommonEsq(bytes[*i:])
	*i++
	if ok {
		return btoa(seq)
	}
	switch bytes[*i] {
	case 'u':
		rc, _ := strconv.ParseUint(string(bytes[*i+1:*i+5]), 16, 32)
		r := rune(rc)
		*i += 4
		return bytesToStr([]byte(string(r)))
	case 'U':
		rc, _ := strconv.ParseUint(string(bytes[*i+1:*i+9]), 16, 32)
		r := rune(rc)
		*i += 8
		return bytesToStr([]byte(string(r)))
	case 'x':
		seq := "0"
		seq += string(bytes[*i : *i+3])
		*i += 2
		return seq
	default:
		seq, n := byteSeq(bytes, *i)
		*i += n - 1
		b, _ := strconv.ParseUint(string(seq), 8, 8)
		return btoa(byte(b))
	}
}

func bytesToStr(bytes []byte) string {
	if len(bytes) == 0 {
		return ""
	}
	var str strings.Builder
	for i := 0; i < len(bytes); i++ {
		b := bytes[i]
		if b == '\\' {
			seq := strEsqSeq(bytes, &i)
			str.WriteString(seq)
		} else {
			str.WriteString(btoa(b))
		}
		str.WriteByte(',')
	}
	return str.String()[:str.Len()-1]
}
