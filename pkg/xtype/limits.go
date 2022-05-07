package xtype

import (
	"math"
)

// MaxOfType returns maximum value of numeric type.
//
// Special case is;
//  MaxOfType(id) -> returns 0 if type id is not numeric.
//  MaxOfType(id) -> returns 0 if type id is not supported.
func MaxOfType(id uint8) uint64 {
	if !IsNumericType(id) {
		return 0
	}
	switch id {
	case I8:
		return math.MaxInt8
	case I16:
		return math.MaxInt16
	case I32:
		return math.MaxInt32
	case I64:
		return math.MaxInt64
	case U8:
		return math.MaxUint8
	case U16:
		return math.MaxUint16
	case U32:
		return math.MaxUint32
	case U64:
		return math.MaxUint64
	case UInt:
		return MaxOfType(UIntCode)
	case Int:
		return MaxOfType(IntCode)
	}
	return 0
}
