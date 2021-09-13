package xbits

import "strconv"

// CheckBitInt reports integer is compatible this bit-size or not.
func CheckBitInt(value string, bit int) bool {
	_, err := strconv.ParseInt(value, 10, bit)
	return err == nil
}

// CheckBitUInt reports unsigned integer is
// compatible this bit-size or not.
func CheckBitUInt(value string, bit int) bool {
	_, err := strconv.ParseUint(value, 10, bit)
	return err == nil
}
