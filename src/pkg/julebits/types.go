package julebits

import "github.com/julelang/jule/pkg/juletype"

// BitsizeType returns bit-size of
// data type of specified type code.
func BitsizeType(t uint8) int {
	switch t {
	case juletype.I8, juletype.U8:
		return 0b1000
	case juletype.I16, juletype.U16:
		return 0b00010000
	case juletype.I32, juletype.U32, juletype.F32:
		return 0b00100000
	case juletype.I64, juletype.U64, juletype.F64:
		return 0b01000000
	case juletype.UINT, juletype.INT:
		return juletype.BIT_SIZE
	default:
		return 0
	}
}
