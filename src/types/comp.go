package types

// Reports whether i8 is compatible with kind.
func Is_i8_compatible(k string) bool {
	k = Real_kind_of(k)
	return k == TypeKind_I8
}

// Reports whether i16 is compatible with kind.
func Is_i16_compatible(k string) bool {
	k = Real_kind_of(k)
	return k == TypeKind_I8 || k == TypeKind_I16 || k == TypeKind_U8
}

// Reports whether i32 is compatible with kind.
func Is_i32_compatible(k string) bool {
	k = Real_kind_of(k)
	return (k == TypeKind_I8 ||
		k == TypeKind_I16 ||
		k == TypeKind_I32 ||
		k == TypeKind_U8 ||
		k == TypeKind_U16)
}

// Reports whether i64 is compatible with kind.
func Is_i64_compatible(k string) bool {
	k = Real_kind_of(k)
	return (k == TypeKind_I8 ||
		k == TypeKind_I16 ||
		k == TypeKind_I32 ||
		k == TypeKind_I64 ||
		k == TypeKind_U8 ||
		k == TypeKind_U16 ||
		k == TypeKind_U32)
}

// Reports whether u8 is compatible with kind.
func Is_u8_compatible(k string) bool {
	k = Real_kind_of(k)
	return k == TypeKind_U8
}

// Reports whether u16 is compatible with kind.
func Is_u16_compatible(k string) bool {
	k = Real_kind_of(k)
	return k == TypeKind_U8 || k == TypeKind_U16
}

// Reports whether u32 is compatible with kind.
func Is_u32_compatible(k string) bool {
	k = Real_kind_of(k)
	return k == TypeKind_U8 || k == TypeKind_U16 || k == TypeKind_U32
}

// Reports whether u64 is compatible with kind.
func Is_u64_compatible(k string) bool {
	k = Real_kind_of(k)
	return (k == TypeKind_U8 ||
		k == TypeKind_U16 ||
		k == TypeKind_U32 ||
		k == TypeKind_U64)
}

// Reports whether f32 is compatible with kind.
func Is_f32_compatible(k string) bool {
	k = Real_kind_of(k)
	return (k == TypeKind_F32 ||
		k == TypeKind_I8 ||
		k == TypeKind_I16 ||
		k == TypeKind_I32 ||
		k == TypeKind_I64 ||
		k == TypeKind_U8 ||
		k == TypeKind_U16 ||
		k == TypeKind_U32 ||
		k == TypeKind_U64)
}

// Reports whether f64 is compatible with kind.
func Is_f64_compatible(k string) bool {
	k = Real_kind_of(k)
	return (k == TypeKind_F64 ||
		k == TypeKind_F32 ||
		k == TypeKind_I8 ||
		k == TypeKind_I16 ||
		k == TypeKind_I32 ||
		k == TypeKind_I64 ||
		k == TypeKind_U8 ||
		k == TypeKind_U16 ||
		k == TypeKind_U32 ||
		k == TypeKind_U64)
}

// Reports types are compatible.
// k1 is the destination type, k2 is the source type.
// Return false if k2 is unsupported kind.
func Types_are_compatible(k1 string, k2 string) bool {
	k1 = Real_kind_of(k1)
	switch k1 {
	case TypeKind_ANY:
		return true

	case TypeKind_I8:
		return Is_i8_compatible(k2)

	case TypeKind_I16:
		return Is_i16_compatible(k2)

	case TypeKind_I32:
		return Is_i32_compatible(k2)

	case TypeKind_I64:
		return Is_i64_compatible(k2)

	case TypeKind_U8:
		return Is_u8_compatible(k2)

	case TypeKind_U16:
		return Is_u16_compatible(k2)

	case TypeKind_U32:
		return Is_u32_compatible(k2)

	case TypeKind_U64:
		return Is_u64_compatible(k2)

	case TypeKind_F32:
		return Is_f32_compatible(k2)

	case TypeKind_F64:
		return Is_f64_compatible(k2)

	case TypeKind_BOOL:
		return k2 == TypeKind_BOOL

	case TypeKind_STR:
		return k2 == TypeKind_STR

	default:
		return false
	}
}

// Reports whether i16 is greater than given kind.
func Is_i16_greater(k string) bool {
	k = Real_kind_of(k)
	return k == TypeKind_U8
}

// Reports whether i32 is greater than given kind.
func Is_i32_greater(k string) bool {
	k = Real_kind_of(k)
	return k == TypeKind_I8 || k == TypeKind_I16
}

// Reports whether i64 is greater than given kind.
func Is_i64_greater(k string) bool {
	k = Real_kind_of(k)
	return k == TypeKind_I8 || k == TypeKind_I16 || k == TypeKind_I32
}

// Reports whether u8 is greater than given kind.
func Is_u8_greater(k string) bool {
	k = Real_kind_of(k)
	return k == TypeKind_I8
}

// Reports whether u16 is greater than given kind.
func Is_u16_greater(k string) bool {
	k = Real_kind_of(k)
	return k == TypeKind_U8 || k == TypeKind_I8 || k == TypeKind_I16
}

// Reports whether u32 is greater than given kind.
func Is_u32_greater(k string) bool {
	k = Real_kind_of(k)
	return (k == TypeKind_U8 ||
		k == TypeKind_U16 ||
		k == TypeKind_I8 ||
		k == TypeKind_I16 ||
		k == TypeKind_I32)
}

// Reports whether u64 is greater than given kind.
func Is_u64_greater(k string) bool {
	k = Real_kind_of(k)
	return (k == TypeKind_U8 ||
		k == TypeKind_U16 ||
		k == TypeKind_U32 ||
		k == TypeKind_I8 ||
		k == TypeKind_I16 ||
		k == TypeKind_I32 ||
		k == TypeKind_I64)
}

// Reports whether f32 is greater than given kind.
func Is_f32_greater(k string) bool {
	return k != TypeKind_F64
}

// Reports whether f64 is greater than given kind.
func Is_f64_greater(k string) bool {
	return true
}

// Reports whether k1 kind greater than k2 kind.
func Is_greater(k1 string, k2 string) bool {
	k1 = Real_kind_of(k1)

	switch k1 {
	case TypeKind_I16:
		return Is_i16_greater(k2)

	case TypeKind_I32:
		return Is_i32_greater(k2)

	case TypeKind_I64:
		return Is_i64_greater(k2)

	case TypeKind_U16:
		return Is_u16_greater(k2)

	case TypeKind_U8:
		return Is_u8_greater(k2)

	case TypeKind_U32:
		return Is_u32_greater(k2)

	case TypeKind_U64:
		return Is_u64_greater(k2)

	case TypeKind_F32:
		return Is_f32_greater(k2)

	case TypeKind_F64:
		return Is_f64_greater(k2)

	case TypeKind_ANY:
		return true

	default:
		return false
	}
}
