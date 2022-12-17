package types

import (
	"math"
	"strconv"
	"strings"

	"github.com/julelang/jule/pkg/juletype"
)

// BitsizeType returns bit-size of
// data type of specified type code.
func BitsizeType(t uint8) int {
	switch t {
	case juletype.I8, juletype.U8:
		return 0b1000
	case juletype.I16, juletype.U16:
		return 0b00010000
	case juletype.I32, juletype.U32, juletype.F32:
		return 0b00100000
	case juletype.I64, juletype.U64, juletype.F64:
		return 0b01000000
	case juletype.UINT, juletype.INT:
		return juletype.BIT_SIZE
	default:
		return 0
	}
}

// CheckBitFloat reports float is compatible this bit-size or not.
func CheckBitFloat(val string, bit int) bool {
	_, err := strconv.ParseFloat(val, bit)
	return err == nil
}

// BitsizeFloat returns minimum bitsize of given value.
func BitsizeFloat(x float64) uint64 {
	switch {
	case x >= -math.MaxFloat32 && x <= math.MaxFloat32:
		return 32
	default:
		return 64
	}
}

// MAX_INT is the maximum bitsize of integer types.
const MAX_INT = 64

type bitChecker = func(v string, base int, bit int) error

// CheckBitInt reports integer is compatible this bit-size or not.
func CheckBitInt(v string, bit int) bool {
	return check_bit(v, bit, func(v string, base int, bit int) error {
		_, err := strconv.ParseInt(v, base, bit)
		return err
	})
}

// CheckBitUInt reports unsigned integer is
// compatible this bit-size or not.
func CheckBitUInt(v string, bit int) bool {
	return check_bit(v, bit, func(v string, base int, bit int) error {
		_, err := strconv.ParseUint(v, base, bit)
		return err
	})
}

func check_bit(v string, bit int, checker bitChecker) bool {
	var err error
	switch {
	case v == "":
		return false
	case len(v) == 1:
		return true
	case strings.HasPrefix(v, "0x"):
		err = checker(v[2:], 16, bit)
	case strings.HasPrefix(v, "0b"):
		err = checker(v[2:], 2, bit)
	case v[0] == '0':
		err = checker(v[1:], 8, bit)
	default:
		err = checker(v, 10, bit)
	}
	return err == nil
}

// BitsizeInt returns minimum bitsize of given value.
func BitsizeInt(x int64) uint64 {
	switch {
	case x >= math.MinInt8 && x <= math.MaxInt8:
		return 8
	case x >= math.MinInt16 && x <= math.MaxInt16:
		return 16
	case x >= math.MinInt32 && x <= math.MaxInt32:
		return 32
	default:
		return MAX_INT
	}
}

// BitsizeUInt returns minimum bitsize of given value.
func BitsizeUInt(x uint64) uint64 {
	switch {
	case x <= math.MaxUint8:
		return 8
	case x <= math.MaxUint16:
		return 16
	case x <= math.MaxUint32:
		return 32
	default:
		return MAX_INT
	}
}
