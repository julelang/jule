package xbits

import (
	"strconv"

	"github.com/the-xlang/xxc/pkg/x"
)

// BitsizeType returns bit-size of
// data type of specified type code.
func BitsizeType(t uint8) int {
	switch t {
	case x.I8, x.U8:
		return 0b1000
	case x.I16, x.U16:
		return 0b00010000
	case x.I32, x.U32, x.F32:
		return 0b00100000
	case x.I64, x.U64, x.F64:
		return 0b01000000
	case x.Size:
		return strconv.IntSize
	default:
		return 0
	}
}
