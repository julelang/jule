package xbits

import (
	"strconv"
	"strings"
)

// MaxInt is the maximum bitsize of integer types.
const MaxInt = 64

type bitChecker = func(val string, base, bit int) error

// CheckBitInt reports integer is compatible this bit-size or not.
func CheckBitInt(val string, bit int) bool {
	return checkBit(val, bit, func(val string, base, bit int) error {
		_, err := strconv.ParseInt(val, base, bit)
		return err
	})
}

// CheckBitUInt reports unsigned integer is
// compatible this bit-size or not.
func CheckBitUInt(val string, bit int) bool {
	return checkBit(val, bit, func(val string, base, bit int) error {
		_, err := strconv.ParseUint(val, base, bit)
		return err
	})
}

func checkBit(val string, bit int, checker bitChecker) bool {
	var err error
	switch {
	case val == "":
		return false
	case len(val) == 1:
		return true
	case strings.HasPrefix(val, "0x"):
		err = checker(val[2:], 16, bit)
	case strings.HasPrefix(val, "0b"):
		err = checker(val[2:], 2, bit)
	case val[0] == '0':
		err = checker(val[1:], 8, bit)
	default:
		err = checker(val, 10, bit)
	}
	return err == nil
}
