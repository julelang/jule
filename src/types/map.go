package types

import "github.com/julelang/jule/lex"

// Data type (built-in) constants.
const VOID uint8 = 0
const I8 uint8 = 1
const I16 uint8 = 2
const I32 uint8 = 3
const I64 uint8 = 4
const U8 uint8 = 5
const U16 uint8 = 6
const U32 uint8 = 7
const U64 uint8 = 8
const BOOL uint8 = 9
const STR uint8 = 10
const F32 uint8 = 11
const F64 uint8 = 12
const ANY uint8 = 13
const ID uint8 = 14
const FN uint8 = 15
const NIL uint8 = 16
const UINT uint8 = 17
const INT uint8 = 18
const MAP uint8 = 19
const UINTPTR uint8 = 20
const ENUM uint8 = 21
const STRUCT uint8 = 22
const TRAIT uint8 = 23
const SLICE uint8 = 24
const ARRAY uint8 = 25
const UNSAFE uint8 = 26

// TYPE_MAP keep data type codes and kinds.
var TYPE_MAP = map[uint8]string{
	VOID:    VOID_TYPE_STR,
	NIL:     NIL_TYPE_STR,
	I8:      lex.KND_I8,
	I16:     lex.KND_I16,
	I32:     lex.KND_I32,
	I64:     lex.KND_I64,
	U8:      lex.KND_U8,
	U16:     lex.KND_U16,
	U32:     lex.KND_U32,
	U64:     lex.KND_U64,
	STR:     lex.KND_STR,
	BOOL:    lex.KND_BOOL,
	F32:     lex.KND_F32,
	F64:     lex.KND_F64,
	ANY:     lex.KND_ANY,
	UINT:    lex.KND_UINT,
	INT:     lex.KND_INT,
	UINTPTR: lex.KND_UINTPTR,
	UNSAFE:  lex.KND_UNSAFE,
}
