package xbits

import (
	"strconv"
	"strings"
)

// CheckBitInt reports integer is compatible this bit-size or not.
func CheckBitInt(value string, bit int) bool {
	value = strings.TrimPrefix(value, "0x")
	_, err := strconv.ParseInt(value, 16, bit)
	return err == nil
}

// CheckBitUInt reports unsigned integer is
// compatible this bit-size or not.
func CheckBitUInt(value string, bit int) bool {
	value = strings.TrimPrefix(value, "0x")
	_, err := strconv.ParseUint(value, 16, bit)
	return err == nil
}
