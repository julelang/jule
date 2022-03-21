package xbits

import (
	"strconv"
	"strings"
)

// CheckBitInt reports integer is compatible this bit-size or not.
func CheckBitInt(val string, bit int) bool {
	var err error
	if strings.HasPrefix(val, "0x") {
		_, err = strconv.ParseInt(val[2:], 16, bit)
	} else {
		_, err = strconv.ParseInt(val, 10, bit)
	}
	return err == nil
}

// CheckBitUInt reports unsigned integer is
// compatible this bit-size or not.
func CheckBitUInt(val string, bit int) bool {
	var err error
	if strings.HasPrefix(val, "0x") {
		_, err = strconv.ParseUint(val[2:], 16, bit)
	} else {
		_, err = strconv.ParseUint(val, 10, bit)
	}
	return err == nil
}
