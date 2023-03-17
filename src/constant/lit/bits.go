package lit

import (
	"strconv"
	"strings"

	"github.com/julelang/jule/types"
)

type bit_checker = func(v string, base int, bit int) bool

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
//  - 32 for 32-bit
//  - 64 for 64-bit
func Bitsize_of_float(x float64) uint64 {
	switch {
	case types.MIN_F32 <= x && x <= types.MAX_F32:
		return 0b00100000

	default:
		return 0b01000000
	}
}

// Reports minimum bit-size of given signed integer.
//
// Possible values are:
//  - 8 for 8-bit
//  - 16 for 16-bit
//  - 32 for 32-bit
//  - 64 for 64-bit
func Bitsize_of_int(x int64) uint64 {
	switch {
	case types.MIN_I8 <= x && x <= types.MAX_I8:
		return 0b1000

	case types.MIN_I16 <= x && x <= types.MAX_I16:
		return 0b00010000

	case types.MIN_I32 <= x && x <= types.MAX_I32:
		return 0b00100000

	default:
		return 0b01000000
	}
}

// Reports minimum bit-size of given unsigned integer.
//
// Possible values are:
//  - 8 for 8-bit
//  - 16 for 16-bit
//  - 32 for 32-bit
//  - 64 for 64-bit
func Bitsize_of_uint(x uint64) uint64 {
	switch {
	case x <= types.MAX_U8:
		return 0b1000

	case x <= types.MAX_U16:
		return 0b00010000

	case x <= types.MAX_U32:
		return 0b00100000

	default:
		return 0b01000000
	}
}
