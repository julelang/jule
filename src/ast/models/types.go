package models

import (
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// Necessary defines to remove "types" package dependency

const nil_type_str = "<nil>"
const void_type_str = "<void>"

// Data type (built-in) constants.
const void_t = 0
const i8_t = 1
const i16_t = 2
const i32_t = 3
const i64_t = 4
const u8_t = 5
const u16_t = 6
const u32_t = 7
const u64_t = 8
const bool_t = 9
const str_t = 10
const f32_t = 11
const f64_t = 12
const any_t = 13
const id_t = 14
const fn_t = 15
const nil_t = 16
const uint_t = 17
const int_t = 18
const map_t = 19
const uintptr_t = 20
const enum_t = 21
const struct_t = 22
const trait_t = 23
const slice_t = 24
const array_t = 25
const unsafe_t = 26

// type_map keeps data type codes and kinds.
var type_map = map[uint8]string{
	void_t:    void_type_str,
	nil_t:     nil_type_str,
	i8_t:      lex.KND_I8,
	i16_t:     lex.KND_I16,
	i32_t:     lex.KND_I32,
	i64_t:     lex.KND_I64,
	u8_t:      lex.KND_U8,
	u16_t:     lex.KND_U16,
	u32_t:     lex.KND_U32,
	u64_t:     lex.KND_U64,
	str_t:     lex.KND_STR,
	bool_t:    lex.KND_BOOL,
	f32_t:     lex.KND_F32,
	f64_t:     lex.KND_F64,
	any_t:     lex.KND_ANY,
	uint_t:    lex.KND_UINT,
	int_t:     lex.KND_INT,
	uintptr_t: lex.KND_UINTPTR,
	unsafe_t:  lex.KND_UNSAFE,
}

// cpp_id returns cpp output identifier of data-type.
func cpp_id(t uint8) string {
	if t == void_t || t == unsafe_t {
		return "void"
	}
	id := type_map[t]
	if id == "" {
		return id
	}
	id = build.AsTypeId(id)
	return id
}
