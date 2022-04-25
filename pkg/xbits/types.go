package xbits

import (
	"strconv"

	"github.com/the-xlang/xxc/pkg/xtype"
)

// BitsizeType returns bit-size of
// data type of specified type code.
func BitsizeType(t uint8) int {
	switch t {
	case xtype.I8, xtype.U8:
		return 0b1000
	case xtype.I16, xtype.U16:
		return 0b00010000
	case xtype.I32, xtype.U32, xtype.F32:
		return 0b00100000
	case xtype.I64, xtype.U64, xtype.F64:
		return 0b01000000
	case xtype.Size:
		return strconv.IntSize
	default:
		return 0
	}
}
