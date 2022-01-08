package xbits

import "github.com/the-xlang/x/pkg/x"

// BitsizeOfType returns bit-size of
// data type of specified type code.
func BitsizeOfType(t uint8) int {
	switch t {
	case x.I8, x.U8:
		return 8
	case x.I16, x.U16:
		return 16
	case x.I32, x.U32:
		return 32
	case x.I64, x.U64:
		return 64
	}
	return 0
}
