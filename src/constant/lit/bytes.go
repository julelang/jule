package lit

import (
	"strconv"
	"unicode/utf8"
)

// Reports whether kind is byte literal and returns
// literal without quotes.
//
// Byte literal patterns:
//  - 'x': 0 <= x && x <= 255
//  - '\xhh'
//  - '\nnn'
func Is_byte_lit(kind string) (string, bool) {
	if len(kind) < 3 {
		return "", false
	}

	kind = kind[1 : len(kind)-1] // Remove quotes.
	is_byte := false
	
	// TODO: Add support for byte escape sequences.
	switch {
	case len(kind) == 1 && kind[0] <= 255:
		is_byte = true

	case kind[0] == '\\' && kind[1] == 'x':
		is_byte = true

	case kind[0] == '\\' && kind[1] >= '0' && kind[1] <= '7':
		is_byte = true
	}

	return kind, is_byte
}

// Returns rune value string from bytes, not includes quotes.
// Bytes are represents rune literal, allows escape sequences.
// Returns empty string if len(bytes) == 0
func To_rune(bytes []byte) rune {
	if len(bytes) == 0 {
		return 0
	}

	var r rune = 0
	if bytes[0] == '\\' && len(bytes) > 1 {
		i := 0
		r = rune_from_esq_seq(bytes, &i)
	} else {
		r, _ = utf8.DecodeRune(bytes)
	}

	return r
}

// Returns rune as rune value in hexadecimal.
func Rtoa(r rune) string { return "0x" + strconv.FormatInt(int64(r), 16) }

// Returns raw-string value string from bytes, not includes quotes.
// Bytes are represents string characters, allows escape sequences.
// Returns empty string if len(bytes) == 0
func To_raw_str(bytes []byte) string { return To_str(bytes) }

// Returns string value string from bytes, not includes quotes.
// Bytes are represents string characters, allows escape sequences.
// Returns empty string if len(bytes) == 0
func To_str(bytes []byte) string {
	if len(bytes) == 0 {
		return ""
	}

	s := ""
	i := 0
	for i < len(bytes) {
		b := bytes[i]
		if b == '\\' {
			s += str_esq_seq(bytes, &i)
		} else {
			s += string(b)
			i++
		}
	}
	return s
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
	*i++ // Skip escape sequence solidus.
	if ok {
		*i++ // Skip sequence specifier.
		return rune(b)
	}

	switch bytes[*i] {
	case 'u':
		const SEQ_LEN = 5
		rc, _ := strconv.ParseUint(string(bytes[*i+1:*i+SEQ_LEN]), 16, 32)
		*i += SEQ_LEN
		r := rune(rc)
		return r

	case 'U':
		const SEQ_LEN = 9
		rc, _ := strconv.ParseUint(string(bytes[*i+1:*i+SEQ_LEN]), 16, 32)
		*i += SEQ_LEN
		r := rune(rc)
		return r

	case 'x':
		const SEQ_LEN = 3
		seq := bytes[*i+1:*i+SEQ_LEN]
		*i += SEQ_LEN
		b, _ := strconv.ParseUint(string(seq), 16, 8)
		return rune(b)

	default:
		const SEQ_LEN = 3
		seq := bytes[*i : *i+SEQ_LEN]
		*i += SEQ_LEN
		b, _ := strconv.ParseUint(string(seq), 8, 8)
		return rune(b)
	}
}

func str_esq_seq(bytes []byte, i *int) string {
	r := rune_from_esq_seq(bytes, i)
	if r <= 255 {
		return string(r)
	}
	return To_str([]byte(string(r)))
}
