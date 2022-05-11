package xtype

import (
	"strconv"

	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/xapi"
)

// Data type (built-in) constants.
const (
	Void    uint8 = 0
	I8      uint8 = 1
	I16     uint8 = 2
	I32     uint8 = 3
	I64     uint8 = 4
	U8      uint8 = 5
	U16     uint8 = 6
	U32     uint8 = 7
	U64     uint8 = 8
	Bool    uint8 = 9
	Str     uint8 = 10
	F32     uint8 = 11
	F64     uint8 = 12
	Any     uint8 = 13
	Char    uint8 = 14
	Id      uint8 = 15
	Func    uint8 = 16
	Nil     uint8 = 17
	UInt    uint8 = 18
	Int     uint8 = 19
	Map     uint8 = 20
	Voidptr uint8 = 21
	Intptr  uint8 = 22
	UIntptr uint8 = 23
	Enum    uint8 = 24
	Struct  uint8 = 25
)

// CodeMap keep data-type codes and identifiers.
var CodeMap = map[uint8]string{
	I8:      tokens.I8,
	I16:     tokens.I16,
	I32:     tokens.I32,
	I64:     tokens.I64,
	U8:      tokens.U8,
	U16:     tokens.U16,
	U32:     tokens.U32,
	U64:     tokens.U64,
	Str:     tokens.STR,
	Bool:    tokens.BOOL,
	F32:     tokens.F32,
	F64:     tokens.F64,
	Any:     "any",
	Char:    tokens.CHAR,
	UInt:    tokens.UINT,
	Int:     tokens.INT,
	Voidptr: tokens.VOIDPTR,
	Intptr:  tokens.INTPTR,
	UIntptr: tokens.UINTPTR,
}

// IntCode is integer type code of current platform architecture.
// Is equavalent to "int", but specific bit-sized integer type code.
var IntCode uint8

// UIntCode is integer type code of current platform architecture.
// Is equavalent to "uint", but specific bit-sized integer type code.
var UIntCode uint8

// BitSize is bit size of architecture.
var BitSize int

const (
	NumericTypeStr = "<numeric>"
	NilTypeStr     = "<nil>"
	VoidTypeStr    = "<void>"
)

// GetRealCode returns real type code of code.
// If types is "int" or "uint", set to bit-specific type code.
func GetRealCode(t uint8) uint8 {
	switch t {
	case Char:
		t = U8
	case Int, Intptr:
		t = IntCode
	case UInt, UIntptr:
		t = UIntCode
	}
	return t
}

// I16GreaterThan reports I16 is greater or not data-type than specifed type.
func I16GreaterThan(t uint8) bool {
	t = GetRealCode(t)
	return t == U8
}

// I32GreaterThan reports I32 is greater or not data-type than specifed type.
func I32GreaterThan(t uint8) bool {
	t = GetRealCode(t)
	return t == I8 || t == I16
}

// I64GreaterThan reports I64 is greater or not data-type than specifed type.
func I64GreaterThan(t uint8) bool {
	t = GetRealCode(t)
	return t == I8 || t == I16 || t == I32
}

// U16GreaterThan reports U16 is greater or not data-type than specifed type.
func U16GreaterThan(t uint8) bool {
	t = GetRealCode(t)
	return t == U8
}

// U32GreaterThan reports U32 is greater or not data-type than specifed type.
func U32GreaterThan(t uint8) bool {
	t = GetRealCode(t)
	return t == U8 || t == U16
}

// U64GreaterThan reports U64 is greater or not data-type than specifed type.
func U64GreaterThan(t uint8) bool {
	t = GetRealCode(t)
	return t == U8 || t == U16 || t == U32
}

// F32GreaterThan reports F32 is greater or not data-type than specifed type.
func F32GreaterThan(t uint8) bool { return t != Any && t != F64 }

// F64GreaterThan reports F64 is greater or not data-type than specifed type.
func F64GreaterThan(t uint8) bool { return t != Any }

// TypeGreaterThan reports type one is greater than type two or not.
func TypeGreaterThan(t1, t2 uint8) bool {
	t1 = GetRealCode(t1)
	switch t1 {
	case I16:
		return I16GreaterThan(t2)
	case I32:
		return I32GreaterThan(t2)
	case I64:
		return I64GreaterThan(t2)
	case U16:
		return U16GreaterThan(t2)
	case U32:
		return U32GreaterThan(t2)
	case U64:
		return U64GreaterThan(t2)
	case F32:
		return F32GreaterThan(t2)
	case F64:
		return F64GreaterThan(t2)
	case Enum:
		return true
	}
	return false
}

// I8CompatibleWith reports i8 is compatible or not with data-type specifed type.
func I8CompatibleWith(t uint8) bool {
	t = GetRealCode(t)
	return t == I8
}

// I16CompatibleWith reports i16 is compatible or not with data-type specifed type.
func I16CompatibleWith(t uint8) bool {
	t = GetRealCode(t)
	return t == I8 || t == I16 || t == U8
}

// I32CompatibleWith reports i32 is compatible or not with data-type specifed type.
func I32CompatibleWith(t uint8) bool {
	t = GetRealCode(t)
	return t == I8 || t == I16 || t == I32 || t == U8 || t == U16
}

// I64CompatibleWith reports i64 is compatible or not with data-type specifed type.
func I64CompatibleWith(t uint8) bool {
	t = GetRealCode(t)
	switch t {
	case I8, I16, I32, I64, U8, U16, U32:
		return true
	default:
		return false
	}
}

// U8CompatibleWith reports u8 is compatible or not with data-type specifed type.
func U8CompatibleWith(t uint8) bool {
	t = GetRealCode(t)
	return t == U8
}

// U16CompatibleWith reports u16 is compatible or not with data-type specifed type.
func U16CompatibleWith(t uint8) bool {
	t = GetRealCode(t)
	return t == U8 || t == U16
}

// U32CompatibleWith reports u32 is compatible or not with data-type specifed type.
func U32CompatibleWith(t uint8) bool {
	t = GetRealCode(t)
	return t == U8 || t == U16 || t == U32
}

// U16CompatibleWith reports u64 is compatible or not with data-type specifed type.
func U64CompatibleWith(t uint8) bool {
	t = GetRealCode(t)
	return t == U8 || t == U16 || t == U32 || t == U64
}

// F32CompatibleWith reports f32 is compatible or not with data-type specifed type.
func F32CompatibleWith(t uint8) bool {
	t = GetRealCode(t)
	switch t {
	case F32, I8, I16, I32, I64, U8, U16, U32, U64:
		return true
	default:
		return false
	}
}

// F64CompatibleWith reports f64 is compatible or not with data-type specifed type.
func F64CompatibleWith(t uint8) bool {
	t = GetRealCode(t)
	switch t {
	case F64, F32, I8, I16, I32, I64, U8, U16, U32, U64:
		return true
	default:
		return false
	}
}

// TypeAreCompatible reports type one and type two is compatible or not.
func TypesAreCompatible(t1, t2 uint8, ignoreany bool) bool {
	if !ignoreany && t1 == Any {
		return true
	}
	t1 = GetRealCode(t1)
	switch t1 {
	case I8:
		return I8CompatibleWith(t2)
	case I16:
		return I16CompatibleWith(t2)
	case I32:
		return I32CompatibleWith(t2)
	case I64:
		return I64CompatibleWith(t2)
	case U8:
		return U8CompatibleWith(t2)
	case U16:
		return U16CompatibleWith(t2)
	case U32:
		return U32CompatibleWith(t2)
	case U64:
		return U64CompatibleWith(t2)
	case Bool:
		return t2 == Bool
	case Str:
		return t2 == Str
	case F32:
		return F32CompatibleWith(t2)
	case F64:
		return F64CompatibleWith(t2)
	case Nil:
		return t2 == Nil
	}
	return false
}

// IsIntegerType reports type is signed/unsigned integer or not.
func IsIntegerType(t uint8) bool {
	return IsSignedIntegerType(t) || IsUnsignedNumericType(t)
}

// IsNumericType reports type is numeric or not.
func IsNumericType(t uint8) bool { return IsIntegerType(t) || IsFloatType(t) }

// IsFloatType reports type is float or not.
func IsFloatType(t uint8) bool { return t == F32 || t == F64 }

// IsSignedNumericType reports type is signed numeric or not.
func IsSignedNumericType(t uint8) bool {
	return IsSignedIntegerType(t) || IsFloatType(t)
}

// IsSignedIntegerType reports type is signed integer or not.
func IsSignedIntegerType(t uint8) bool {
	switch t {
	case I8, I16, I32, I64, Int, Intptr:
		return true
	default:
		return false
	}
}

// IsUnsignedNumericType reports type is unsigned numeric or not.
func IsUnsignedNumericType(t uint8) bool {
	switch t {
	case U8, U16, U32, U64, UInt, UIntptr:
		return true
	default:
		return false
	}
}

// TypeFromId returns type id of specified type code.
func TypeFromId(id string) uint8 {
	for t, tid := range CodeMap {
		if id == tid {
			return t
		}
	}
	return 0
}

// CxxTypeIdFromType returns cxx output identifier of data-type.
func CxxTypeIdFromType(t uint8) string {
	if t == Void {
		return "void"
	}
	id := CodeMap[t]
	if id == "" {
		return id
	}
	id = xapi.AsTypeId(id)
	return id
}

// DefaultValOfType returns default value of specified type.
//
// Special case is:
//  DefaultValOfType(t) = "nil" if t is invalid
//  DefaultValOfType(t) = "nil" if t is not have default value
func DefaultValOfType(code uint8) string {
	if IsNumericType(code) || code == Enum {
		return "0"
	}
	switch code {
	case Bool:
		return "false"
	case Str:
		return `""`
	case Char:
		return `'\0'`
	}
	return "nil"
}

func init() {
	BitSize = strconv.IntSize
	switch BitSize {
	case 8:
		IntCode = I8
		UIntCode = U8
	case 16:
		IntCode = I16
		UIntCode = U16
	case 32:
		IntCode = I32
		UIntCode = U32
	case 64:
		IntCode = I64
		UIntCode = U64
	}
}
