package xbits

import "strconv"

// CheckBitFloat reports float is compatible this bit-size or not.
func CheckBitFloat(value string, bit int) bool {
	_, err := strconv.ParseFloat(value, bit)
	return err == nil
}
