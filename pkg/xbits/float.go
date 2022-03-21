package xbits

import "strconv"

// CheckBitFloat reports float is compatible this bit-size or not.
func CheckBitFloat(val string, bit int) bool {
	_, err := strconv.ParseFloat(val, bit)
	return err == nil
}
