package types

import "math"

// MinOfType returns minimum value of integer type.
//
// Special case is;
//  MinOfType(id) -> returns 0 if type id is not integer type.
//  MinOfType(id) -> returns 0 if type id is not supported.
func MinOfType(id uint8) int64 {
	if !IsInteger(id) {
		return 0
	}
	id = GetRealCode(id)
	switch id {
	case I8:
		return math.MinInt8
	case I16:
		return math.MinInt16
	case I32:
		return math.MinInt32
	case I64:
		return math.MinInt64
	}
	return 0
}

// MaxOfType returns maximum value of integer type.
//
// Special case is;
//  MaxOfType(id) -> returns 0 if type id is not integer type.
//  MaxOfType(id) -> returns 0 if type id is not supported.
func MaxOfType(id uint8) uint64 {
	if !IsInteger(id) {
		return 0
	}
	id = GetRealCode(id)
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
	}
	return 0
}
