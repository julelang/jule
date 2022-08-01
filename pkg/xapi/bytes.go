package xapi

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// ToStr returns specified literal as X string literal for cpp.
func ToStr(bytes []byte) string {
	var cpp strings.Builder
	cpp.WriteString("str_xt{")
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
	return "0x" + strconv.FormatUint(uint64(b), 16)
}

func bytesToStr(bytes []byte) string {
	if len(bytes) == 0 {
		return ""
	}
	var str strings.Builder
	for i := 0; i < len(bytes); i++ {
		b := bytes[i]
		if b == '\\' {
			i++
			switch bytes[i] {
			case 'u':
				rc, _ := strconv.ParseUint(string(bytes[i+1:i+5]), 16, 32)
				r := rune(rc)
				str.WriteString(bytesToStr([]byte(string(r))))
				i += 4
			case 'U':
				rc, _ := strconv.ParseUint(string(bytes[i+1:i+9]), 16, 32)
				r := rune(rc)
				str.WriteString(bytesToStr([]byte(string(r))))
				i += 8
			case 'x':
				str.WriteByte('0')
				str.Write(bytes[i : i+3])
				i += 2
			default:
				str.Write(bytes[i : i+3])
				i += 2
			}
		} else {
			str.WriteString(btoa(b))
		}
		str.WriteByte(',')
	}
	return str.String()[:str.Len()-1]
}
