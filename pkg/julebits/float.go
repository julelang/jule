package julebits

import (
	"math"
	"strconv"
)

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
