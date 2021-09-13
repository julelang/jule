package xbits

import "github.com/the-xlang/x/pkg/x"

// BitsizeOfType returns bit-size of
// data type of specified type code.
func BitsizeOfType(t uint8) int {
	switch t {
	case x.Int8, x.UInt8:
		return 8
	case x.Int16, x.UInt16:
		return 16
	case x.Int32, x.UInt32:
		return 32
	case x.Int64, x.UInt64:
		return 64
	}
	return 0
}
