package types

import (
	"strconv"
	"strings"
)

type bit_checker = func(v string, base int, bit int) bool

// Bit-size of runtime architecture.
// Possible values are: 32, and 64.
const BIT_SIZE = 32 << (^uint(0) >> 63)

// Signed integer kind of runtime architecture.
// Is equavalent to "int", but specific bit-sized integer kind.
// Accept as constant.
var SYS_INT string

// Unsigned integer kind of runtime architecture.
// Is equavalent to "uint" and "uintptr", but specific bit-sized integer kind.
// Accept as constant.
var SYS_UINT string

func check_bit(v string, bit int, checker bit_checker) bool {
	switch {
	case v == "":
		return false

	case len(v) == 1:
		return true

	case strings.HasPrefix(v, "0x"): // hexadecimal
		return checker(v[2:], 0b00010000, bit)

	case strings.HasPrefix(v, "0b"): // binary
		return checker(v[2:], 0b10, bit)

	case v[0] == '0': // octal
		return checker(v[1:], 0b1000, bit)

	default: // decimal
		return checker(v, 0b1010, bit)
	}
}

// Returns kind's bit-specific kind if bit-specific like int, uint, and uintptr.
// Returns kind if not bit-specific.
// Bit-size is determined by runtime.
func Real_kind_of(kind string) string {
	switch kind {
	case TypeKind_INT:
		return SYS_INT

	case TypeKind_UINT, TypeKind_UINTPTR:
		return SYS_UINT

	default:
		return kind
	}
}

// Returns kind's bit-size.
// Returns -1 if kind is not numeric.
func Bitsize_of(k string) int {
	switch k {
	case TypeKind_I8, TypeKind_U8:
		return 0b1000

	case TypeKind_I16, TypeKind_U16:
		return 0b00010000

	case TypeKind_I32, TypeKind_U32, TypeKind_F32:
		return 0b00100000

	case TypeKind_I64, TypeKind_U64, TypeKind_F64:
		return 0b01000000

	case TypeKind_UINT, TypeKind_INT:
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
		return TypeKind_I8

	case 0b00010000:
		return TypeKind_I16

	case 0b00100000:
		return TypeKind_I32

	case 0b01000000:
		return TypeKind_I64

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
		return TypeKind_U8

	case 0b00010000:
		return TypeKind_U16

	case 0b00100000:
		return TypeKind_U32

	case 0b01000000:
		return TypeKind_U64

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
		return TypeKind_F32

	case 0b01000000:
		return TypeKind_F64

	default:
		return ""
	}
}

// Reports whether signed integer literal is compatible given bit-size.
func Check_bit_int(v string, bit int) bool {
	return check_bit(v, bit, func(v string, base int, bit int) bool {
		_, err := strconv.ParseInt(v, base, bit)
		return err == nil
	})
}

// Reports whether unsigned integer literal is compatible given bit-size.
func Check_bit_uint(v string, bit int) bool {
	return check_bit(v, bit, func(v string, base int, bit int) bool {
		_, err := strconv.ParseUint(v, base, bit)
		return err == nil
	})
}

// Reports whether float literal is compatible given bit-size.
func Check_bit_float(val string, bit int) bool {
	_, err := strconv.ParseFloat(val, bit)
	return err == nil
}

// Reports minimum bit-size of given floating-point.
//
// Possible values are:
//   - 32 for 32-bit
//   - 64 for 64-bit
func Bitsize_of_float(x float64) uint64 {
	switch {
	case MIN_F32 <= x && x <= MAX_F32:
		return 0b00100000

	default:
		return 0b01000000
	}
}

// Reports minimum bit-size of given signed integer.
//
// Possible values are:
//   - 8 for 8-bit
//   - 16 for 16-bit
//   - 32 for 32-bit
//   - 64 for 64-bit
func Bitsize_of_int(x int64) uint64 {
	switch {
	case MIN_I8 <= x && x <= MAX_I8:
		return 0b1000

	case MIN_I16 <= x && x <= MAX_I16:
		return 0b00010000

	case MIN_I32 <= x && x <= MAX_I32:
		return 0b00100000

	default:
		return 0b01000000
	}
}

// Reports minimum bit-size of given unsigned integer.
//
// Possible values are:
//   - 8 for 8-bit
//   - 16 for 16-bit
//   - 32 for 32-bit
//   - 64 for 64-bit
func Bitsize_of_uint(x uint64) uint64 {
	switch {
	case x <= MAX_U8:
		return 0b1000

	case x <= MAX_U16:
		return 0b00010000

	case x <= MAX_U32:
		return 0b00100000

	default:
		return 0b01000000
	}
}

func init() {
	switch BIT_SIZE {
	case 0b00100000:
		SYS_INT = TypeKind_I32
		SYS_UINT = TypeKind_U32

	case 0b01000000:
		SYS_INT = TypeKind_I64
		SYS_UINT = TypeKind_U64
	}
}
