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
	case Int, Intptr:
		t = IntCode
	case UInt, UIntptr:
		t = UIntCode
	}
	return t
}

// TypeGreaterThan reports type one is greater than type two or not.
func TypeGreaterThan(t1, t2 uint8) bool {
	t1 = GetRealCode(t1)
	t2 = GetRealCode(t2)

	switch t1 {
	case I16:
		return t2 == I8
	case I32:
		return t2 == I8 ||
			t2 == I16
	case I64:
		return t2 == I8 ||
			t2 == I16 ||
			t2 == I32
	case U16:
		return t2 == U8
	case U32:
		return t2 == U8 ||
			t2 == U16
	case U64:
		return t2 == U8 ||
			t2 == U16 ||
			t2 == U32
	case F32:
		return t2 != Any && t2 != F64
	case F64:
		return t2 != Any
	case Enum:
		return true
	}
	return false
}

// TypeAreCompatible reports type one and type two is compatible or not.
func TypesAreCompatible(t1, t2 uint8, ignoreany bool) bool {
	if !ignoreany && t1 == Any {
		return true
	}

	t1 = GetRealCode(t1)
	t2 = GetRealCode(t2)

	// Check.
	switch t1 {
	case I8:
		return t2 == I8
	case I16:
		return t2 == I8 ||
			t2 == I16
	case I32:
		return t2 == I8 ||
			t2 == I16 ||
			t2 == I32 ||
			t2 == Char
	case I64:
		return t2 == I8 ||
			t2 == I16 ||
			t2 == I32 ||
			t2 == I64 ||
			t2 == Char ||
			t2 == Int ||
			t2 == Intptr
	case U8:
		return t2 == U8 ||
			t2 == Char
	case U16:
		return t2 == U8 ||
			t2 == U16 ||
			t2 == Char
	case U32:
		return t2 == U8 ||
			t2 == U16 ||
			t2 == U32 ||
			t2 == Char
	case U64:
		return t2 == U8 ||
			t2 == U16 ||
			t2 == U32 ||
			t2 == U64 ||
			t2 == UInt ||
			t2 == UIntptr ||
			t2 == Char
	case Bool:
		return t2 == Bool
	case Str:
		return t2 == Str
	case F32:
		return t2 == F32 ||
			t2 == I8 ||
			t2 == I16 ||
			t2 == I32 ||
			t2 == U8 ||
			t2 == U16 ||
			t2 == U32 ||
			t2 == Char
	case F64:
		return t2 == F64 ||
			t2 == F32 ||
			t2 == I8 ||
			t2 == I16 ||
			t2 == I32 ||
			t2 == U8 ||
			t2 == U16 ||
			t2 == U32 ||
			t2 == Char
	case Char:
		return t2 == Char ||
			t2 == U8
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
	return t == I8 ||
		t == I16 ||
		t == I32 ||
		t == I64 ||
		t == Int ||
		t == Intptr
}

// IsUnsignedNumericType reports type is unsigned numeric or not.
func IsUnsignedNumericType(t uint8) bool {
	return t == U8 ||
		t == U16 ||
		t == U32 ||
		t == U64 ||
		t == UInt ||
		t == UIntptr
}

// TypeFromId returns type id of specified type code.
func TypeFromId(id string) uint8 {
	switch id {
	case tokens.I8:
		return I8
	case tokens.I16:
		return I16
	case tokens.I32:
		return I32
	case tokens.I64:
		return I64
	case tokens.U8:
		return U8
	case tokens.U16:
		return U16
	case tokens.U32:
		return U32
	case tokens.U64:
		return U64
	case tokens.STR:
		return Str
	case tokens.BOOL:
		return Bool
	case tokens.F32:
		return F32
	case tokens.F64:
		return F64
	case "any":
		return Any
	case tokens.CHAR:
		return Char
	case tokens.UINT:
		return UInt
	case tokens.INT:
		return Int
	case tokens.VOIDPTR:
		return Voidptr
	case tokens.INTPTR:
		return Intptr
	case tokens.UINTPTR:
		return UIntptr
	}
	return 0
}

func CxxTypeIdFromType(typeCode uint8) string {
	switch typeCode {
	case Void:
		return "void"
	case I8:
		return xapi.AsTypeId(tokens.I8)
	case I16:
		return xapi.AsTypeId(tokens.I16)
	case I32:
		return xapi.AsTypeId(tokens.I32)
	case I64:
		return xapi.AsTypeId(tokens.I64)
	case U8:
		return xapi.AsTypeId(tokens.U8)
	case U16:
		return xapi.AsTypeId(tokens.U16)
	case U32:
		return xapi.AsTypeId(tokens.U32)
	case U64:
		return xapi.AsTypeId(tokens.U64)
	case Bool:
		return xapi.AsTypeId(tokens.BOOL)
	case F32:
		return xapi.AsTypeId(tokens.F32)
	case F64:
		return xapi.AsTypeId(tokens.F64)
	case Any:
		return xapi.AsTypeId("any")
	case Str:
		return xapi.AsTypeId(tokens.STR)
	case Char:
		return xapi.AsTypeId(tokens.CHAR)
	case UInt:
		return xapi.AsTypeId(tokens.UINT)
	case Int:
		return xapi.AsTypeId(tokens.INT)
	case Voidptr:
		return xapi.AsTypeId(tokens.VOIDPTR)
	case Intptr:
		return xapi.AsTypeId(tokens.INTPTR)
	case UIntptr:
		return xapi.AsTypeId(tokens.UINTPTR)
	}
	return ""
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
