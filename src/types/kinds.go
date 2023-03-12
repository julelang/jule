package types

import "github.com/julelang/jule/lex"

// Reports whether kind is signed integer.
func Is_sig_int(kind string) bool {
	kind = Real_type_kind(kind)
	switch kind {
	case lex.KND_I8, lex.KND_I16, lex.KND_I32, lex.KND_I64:
		return true

	default:
		return false
	}
}

// Reports kind is unsigned integer.
func Is_unsig_int(kind string) bool {
	kind = Real_type_kind(kind)
	switch kind {
	case lex.KND_U8, lex.KND_U16, lex.KND_U32, lex.KND_U64:
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
	return kind == lex.KND_F32 || kind == lex.KND_F64
}

// Reports whether kind is numeric.
func Is_num(kind string) bool {
	return Is_int(kind) || Is_float(kind)
}

// Reports whether kind is signed numeric.
func Is_sig_num(kind string) bool {
	return Is_sig_int(kind) || Is_float(kind)
}
