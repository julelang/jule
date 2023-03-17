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

// Returns kind's bit-size.
// Returns -1 if kind is not numeric.
func Bitsize_of(k string) int {
	switch k {
	case lex.KND_I8, lex.KND_U8:
		return 0b1000

	case lex.KND_I16, lex.KND_U16:
		return 0b00010000

	case lex.KND_I32, lex.KND_U32, lex.KND_F32:
		return 0b00100000

	case lex.KND_I64, lex.KND_U64, lex.KND_F64:
		return 0b01000000

	case lex.KND_UINT, lex.KND_INT:
		return BIT_SIZE

	default:
		return -1
	}
}

// Returns signed integer kind by bit-size.
// Possible bit-sizes are: 8, 16, 32, and 64.
// Returns empty string if bits is invalid.
func Int_from_bits(bits uint64) string {
	switch bits {
	case 0b1000:
		return lex.KND_I8

	case 0b00010000:
		return lex.KND_I16

	case 0b00100000:
		return lex.KND_I32

	case 0b01000000:
		return lex.KND_I64

	default:
		return ""
	}
}

// Returns unsigned integer kind by bit-size.
// Possible bit-sizes are: 8, 16, 32, and 64.
// Returns empty string if bits is invalid.
func Uint_from_bits(bits uint64) string {
	switch bits {
	case 0b1000:
		return lex.KND_U8

	case 0b00010000:
		return lex.KND_U16

	case 0b00100000:
		return lex.KND_U32

	case 0b01000000:
		return lex.KND_U64

	default:
		return ""
	}
}

// Returns floating-point kind by bit-size.
// Possible bit-sizes are: 32, and 64.
// Returns empty string if bits is invalid.
func Float_from_bits(bits uint64) string {
	switch bits {
	case 0b00100000:
		return lex.KND_F32

	case 0b01000000:
		return lex.KND_F64

	default:
		return ""
	}
}

func init() {
	switch BIT_SIZE {
	case 0b00100000:
		INT_KIND = lex.KND_I32
		UINT_KIND = lex.KND_U32

	case 0b01000000:
		INT_KIND = lex.KND_I64
		UINT_KIND = lex.KND_U64
	}
}
