package xbits

import (
	"strconv"
	"strings"
)

// CheckBitInt reports integer is compatible this bit-size or not.
func CheckBitInt(val string, bit int) bool {
	var err error
	switch {
	case val == "":
		return false
	case len(val) == 1:
		return true
	case strings.HasPrefix(val, "0x"):
		_, err = strconv.ParseInt(val[2:], 16, bit)
	case strings.HasPrefix(val, "0b"):
		_, err = strconv.ParseInt(val[2:], 2, bit)
	case val[0] == '0':
		_, err = strconv.ParseInt(val[1:], 8, bit)
	default:
		_, err = strconv.ParseInt(val, 10, bit)
	}
	return err == nil
}

// CheckBitUInt reports unsigned integer is
// compatible this bit-size or not.
func CheckBitUInt(val string, bit int) bool {
	var err error
	switch {
	case val == "":
		return false
	case len(val) == 1:
		return true
	case strings.HasPrefix(val, "0x"):
		_, err = strconv.ParseUint(val[2:], 16, bit)
	case strings.HasPrefix(val, "0b"):
		_, err = strconv.ParseUint(val[2:], 2, bit)
	case val[0] == '0':
		_, err = strconv.ParseUint(val[1:], 8, bit)
	default:
		_, err = strconv.ParseUint(val, 10, bit)
	}
	return err == nil
}
