package types

// These kinds are must match keyword form itself.

// Jule design: Kinds are enum actually.

// Kind of signed 8-bit integer.
const TypeKind_I8 = "i8"
// Kind of signed 16-bit integer.
const TypeKind_I16 = "i16"
// Kind of signed 32-bit integer.
const TypeKind_I32 = "i32"
// Kind of signed 64-bit integer.
const TypeKind_I64 = "i64"
// Kind of unsigned 8-bit integer.
const TypeKind_U8 = "u8"
// Kind of unsigned 16-bit integer.
const TypeKind_U16 = "u16"
// Kind of unsigned 32-bit integer.
const TypeKind_U32 = "u32"
// Kind of unsigned 64-bit integer.
const TypeKind_U64 = "u64"
// Kind of 32-bit floating-point.
const TypeKind_F32 = "f32"
// Kind of 64-bit floating-point.
const TypeKind_F64 = "f64"
// Kind of system specific bit-size unsigned integer.
const TypeKind_UINT = "uint"
// Kind of system specific bit-size signed integer.
const TypeKind_INT = "int"
// Kind of system specific bit-size unsigned integer.
const TypeKind_UINTPTR = "uintptr"
// Kind of boolean.
const TypeKind_BOOL = "bool"
// Kind of string.
const TypeKind_STR = "str"
// Kind of any type.
const TypeKind_ANY = "any"

// Reports whether kind is signed integer.
func Is_sig_int(kind string) bool {
	kind = Real_type_kind(kind)
	switch kind {
	case TypeKind_I8, TypeKind_I16, TypeKind_I32, TypeKind_I64:
		return true

	default:
		return false
	}
}

// Reports kind is unsigned integer.
func Is_unsig_int(kind string) bool {
	kind = Real_type_kind(kind)
	switch kind {
	case TypeKind_U8, TypeKind_U16, TypeKind_U32, TypeKind_U64:
		return true

	default:
		return false
	}
}

// Reports whether kind is signed/unsigned integer.
func Is_int(kind string) bool {
	return Is_sig_int(kind) || Is_unsig_int(kind)
}

// Reports whether kind is float.
func Is_float(kind string) bool {
	return kind == TypeKind_F32 || kind == TypeKind_F64
}

// Reports whether kind is numeric.
func Is_num(kind string) bool {
	return Is_int(kind) || Is_float(kind)
}

// Reports whether kind is signed numeric.
func Is_sig_num(kind string) bool {
	return Is_sig_int(kind) || Is_float(kind)
}
