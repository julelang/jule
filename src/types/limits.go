package types

// Maximum positive value of 32-bit floating-points.
const MAX_F32 = 0x1p127 * (1 + (1 - 0x1p-23))

// Maximum negative value of 32-bit floating-points.
const MIN_F32 = -0x1p127 * (1 + (1 - 0x1p-23))

// Maximum positive value of 64-bit floating-points.
const MAX_F64 = 0x1p1023 * (1 + (1 - 0x1p-52))

// Maximum negative value of 64-bit floating-points.
const MIN_F64 = -0x1p1023 * (1 + (1 - 0x1p-52))

// Maximum positive value of 8-bit signed integers.
const MAX_I8 = 127

// Maximum negative value of 8-bit signed integers.
const MIN_I8 = -128

// Maximum positive value of 16-bit signed integers.
const MAX_I16 = 32767

// Maximum negative value of 16-bit signed integers.
const MIN_I16 = -32768

// Maximum positive value of 32-bit signed integers.
const MAX_I32 = 2147483647

// Maximum negative value of 32-bit signed integers.
const MIN_I32 = -2147483648

// Maximum positive value of 64-bit signed integers.
const MAX_I64 = 9223372036854775807

// Maximum negative value of 64-bit signed integers.
const MIN_I64 = -9223372036854775808

// Maximum value of 8-bit unsigned integers.
const MAX_U8 = 255

// Maximum value of 16-bit unsigned integers.
const MAX_U16 = 65535

// Maximum value of 32-bit unsigned integers.
const MAX_U32 = 4294967295

// Maximum value of 64-bit unsigned integers.
const MAX_U64 = 18446744073709551615

// Returns minimum value of signed/unsigned integer and floating-point kinds.
// Returns 0 if kind is invalid.
func Min_of(k string) float64 {
	k = Real_kind_of(k)
	switch k {
	case TypeKind_I8:
		return MIN_I8

	case TypeKind_I16:
		return MIN_I16

	case TypeKind_I32:
		return MIN_I32

	case TypeKind_I64:
		return MIN_I64

	case TypeKind_F32:
		return MIN_F32

	case TypeKind_F64:
		return MIN_F64

	default:
		return 0
	}
}

// Returns minimum value of signed/unsigned integer and floating-point kinds.
// Returns 0 if kind is invalid.
func Max_of(k string) float64 {
	k = Real_kind_of(k)
	switch k {
	case TypeKind_I8:
		return MAX_I8

	case TypeKind_I16:
		return MAX_I16

	case TypeKind_I32:
		return MAX_I32

	case TypeKind_I64:
		return MAX_I64

	case TypeKind_U8:
		return MAX_U8

	case TypeKind_U16:
		return MAX_U16

	case TypeKind_U32:
		return MAX_U32

	case TypeKind_U64:
		return MAX_U64

	case TypeKind_F32:
		return MAX_F32

	case TypeKind_F64:
		return MAX_F64

	default:
		return 0
	}
}
