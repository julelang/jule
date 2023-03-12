package types

import "github.com/julelang/jule/lex"

// Bit-size of runtime architecture.
// Possible values are: 32, and 64.
const BIT_SIZE = 32 << (^uint(0) >> 63)

// Signed integer kind of runtime architecture.
// Is equavalent to "int", but specific bit-sized integer kind.
// Accept as constant.
var INT_KIND string

// Unsigned integer kind of runtime architecture.
// Is equavalent to "uint" and "uintptr", but specific bit-sized integer kind.
// Accept as constant.
var UINT_KIND string

// Returns kind's bit-specific kind if bit-specific like int, uint, and uintptr.
// Returns kind if not bit-specific.
// Bit-size is determined by runtime.
func Real_type_kind(kind string) string {
	switch kind {
	case lex.KND_INT:
		return INT_KIND

	case lex.KND_UINT, lex.KND_UINTPTR:
		return UINT_KIND

	default:
		return kind
	}
}

func init() {
	switch BIT_SIZE {
	case 32:
		INT_KIND = lex.KND_I32
		UINT_KIND = lex.KND_U32

	case 64:
		INT_KIND = lex.KND_I64
		UINT_KIND = lex.KND_U64
	}
}
