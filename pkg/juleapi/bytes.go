package juleapi

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// ToStrLiteral returns string literas as given bytes.
func ToStrLiteral(bytes []byte) string {
	btoa := bytes_to_str(bytes)
	if btoa != "" {
		return `"` + btoa + `"`
	}
	return `""`
}

// ToStr returns specified literal as Jule string literal for cpp.
func ToStr(bytes []byte) string {
	var cpp strings.Builder
	cpp.WriteString("str_jt(")
	literal := ToStrLiteral(bytes)
	if literal != `""` {
		cpp.WriteString(literal)
	}
	cpp.WriteString(")")
	return cpp.String()
}

// ToRawStr returns specified literal as Jule raw-string literal for cpp.
func ToRawStr(bytes []byte) string { return ToStr(bytes) }

// ToRune returns specified literal as Jule rune literal for cpp.
func ToRune(bytes []byte) string {
	if len(bytes) == 0 {
		return ""
	}
	var r rune = 0
	if bytes[0] == '\\' && len(bytes) > 1 {
		i := 0
		r = rune_from_esq_seq(bytes, &i)
	} else {
		r, _ = utf8.DecodeRune(bytes)
	}
	return rtoa(r)
}

func rtoa(r rune) string { return "0x" + strconv.FormatInt(int64(r), 16) }

func sbtoa(b byte) string {
	if b == 0 {
		return "\\x00"
	}
	if b < 128 {
		seq := decompose_common_esq(b)
		if seq != "" {
			return seq
		}
		return string(b)
	}
	seq := strconv.FormatUint(uint64(b), 8)
	return "\\" + seq
}

func decompose_common_esq(b byte) string {
	switch b {
	case '\\':
		return "\\\\"
	case '\'':
		return "'"
	case '"':
		return `\"`
	case '\a':
		return `\a`
	case '\b':
		return `\b`
	case '\f':
		return `\f`
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	case '\t':
		return `\t`
	case '\v':
		return `\v`
	default:
		return ""
	}
}

func try_btoa_common_esq(bytes []byte) (seq byte, ok bool) {
	if len(bytes) < 2 || bytes[0] != '\\' {
		return
	}
	switch bytes[1] {
	case '\\':
		seq = '\\'
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
		seq = '\v'
	}
	ok = seq != 0
	return
}

func rune_from_esq_seq(bytes []byte, i *int) rune {
	b, ok := try_btoa_common_esq(bytes[*i:])
	*i++
	if ok {
		return rune(b)
	}
	switch bytes[*i] {
	case 'u':
		rc, _ := strconv.ParseUint(string(bytes[*i+1:*i+5]), 16, 32)
		*i += 4
		r := rune(rc)
		return r
	case 'U':
		rc, _ := strconv.ParseUint(string(bytes[*i+1:*i+9]), 16, 32)
		*i += 8
		r := rune(rc)
		return r
	case 'x':
		seq := bytes[*i : *i+3]
		*i += 2
		b, _ := strconv.ParseUint(string(seq), 16, 8)
		return rune(b)
	default:
		seq := bytes[*i : *i+3]
		*i += 2
		b, _ := strconv.ParseUint(string(seq), 8, 8)
		return rune(b)
	}
}

func str_esq_seq(bytes []byte, i *int) string {
	r := rune_from_esq_seq(bytes, i)
	if r <= 255 {
		return sbtoa(byte(r))
	}
	return bytes_to_str([]byte(string(r)))
}

func bytes_to_str(bytes []byte) string {
	if len(bytes) == 0 {
		return ""
	}
	var str strings.Builder
	for i := 0; i < len(bytes); i++ {
		b := bytes[i]
		if b == '\\' {
			seq := str_esq_seq(bytes, &i)
			str.WriteString(seq)
		} else {
			str.WriteString(sbtoa(b))
		}
	}
	return str.String()
}
