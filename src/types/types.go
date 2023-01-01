package types

import (
	"strconv"
	"strings"

	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// INT_CODE is integer type code of current platform architecture.
// Is equavalent to "int", but specific bit-sized integer type code.
var INT_CODE uint8

// UINT_CODE is integer type code of current platform architecture.
// Is equavalent to "uint", but specific bit-sized integer type code.
var UINT_CODE uint8

// BIT_SIZE is bit size of architecture.
var BIT_SIZE int

const NUM_TYPE_STR = "<numeric>"
const NIL_TYPE_STR = "<nil>"
const VOID_TYPE_STR = "<void>"

type Type = models.Type
type GenericType = models.GenericType
type Fn = models.Fn

// GetRealCode returns real type code of code.
// If types is "int" or "uint", set to bit-specific type code.
func GetRealCode(t uint8) uint8 {
	switch t {
	case INT:
		t = INT_CODE
	case UINT, UINTPTR:
		t = UINT_CODE
	}
	return t
}

// I16GreaterThan reports I16 is greater or not data-type than specified type.
func I16GreaterThan(t uint8) bool {
	t = GetRealCode(t)
	return t == U8
}

// I32GreaterThan reports I32 is greater or not data-type than specified type.
func I32GreaterThan(t uint8) bool {
	t = GetRealCode(t)
	return t == I8 || t == I16
}

// I64GreaterThan reports I64 is greater or not data-type than specified type.
func I64GreaterThan(t uint8) bool {
	t = GetRealCode(t)
	return t == I8 || t == I16 || t == I32
}

// U16GreaterThan reports U16 is greater or not data-type than specified type.
func U16GreaterThan(t uint8) bool {
	t = GetRealCode(t)
	return t == U8
}

// U32GreaterThan reports U32 is greater or not data-type than specified type.
func U32GreaterThan(t uint8) bool {
	t = GetRealCode(t)
	return t == U8 || t == U16
}

// U64GreaterThan reports U64 is greater or not data-type than specified type.
func U64GreaterThan(t uint8) bool {
	t = GetRealCode(t)
	return t == U8 || t == U16 || t == U32
}

// F32GreaterThan reports F32 is greater or not data-type than specified type.
func F32GreaterThan(t uint8) bool {
	return t != ANY && t != F64
}

// F64GreaterThan reports F64 is greater or not data-type than specified type.
func F64GreaterThan(t uint8) bool {
	return t != ANY
}

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
	case ENUM, ANY:
		return true
	}
	return false
}

// IsInteger reports type is signed/unsigned integer or not.
func IsInteger(t uint8) bool {
	return IsSignedInteger(t) || IsUnsignedInteger(t)
}

// IsNumeric reports type is numeric or not.
func IsNumeric(t uint8) bool {
	return IsInteger(t) || IsFloat(t)
}

// IsFloat reports type is float or not.
func IsFloat(t uint8) bool {
	return t == F32 || t == F64
}

// IsSignedNumeric reports type is signed numeric or not.
func IsSignedNumeric(t uint8) bool {
	return IsSignedInteger(t) || IsFloat(t)
}

// IsSignedInteger reports type is signed integer or not.
func IsSignedInteger(t uint8) bool {
	t = GetRealCode(t)
	switch t {
	case I8, I16, I32, I64, INT:
		return true
	default:
		return false
	}
}

// IsUnsignedInteger reports type is unsigned integer or not.
func IsUnsignedInteger(t uint8) bool {
	t = GetRealCode(t)
	switch t {
	case U8, U16, U32, U64, UINT, UINTPTR:
		return true
	default:
		return false
	}
}

// TypeFromId returns type id of specified type code.
func TypeFromId(id string) uint8 {
	for t, tid := range TYPE_MAP {
		if id == tid {
			return t
		}
	}
	return 0
}

// CppId returns cpp output identifier of data-type.
func CppId(t uint8) string {
	if t == VOID || t == UNSAFE {
		return "void"
	}
	id := TYPE_MAP[t]
	if id == "" {
		return id
	}
	id = build.AsTypeId(id)
	return id
}

// DefaultValOfType returns default value of specified type.
//
// Special case is:
//  DefaultValOfType(t) = "nil" if t is invalid
//  DefaultValOfType(t) = "nil" if t is not have default value
func DefaultValOfType(t uint8) string {
	t = GetRealCode(t)
	if IsNumeric(t) || t == ENUM {
		return "0"
	}
	switch t {
	case BOOL:
		return "false"
	case STR:
		return `""`
	}
	return "nil"
}

// IntFromBits returns type code by bits.
func IntFromBits(bits uint64) uint8 {
	switch bits {
	case 8:
		return I8
	case 16:
		return I16
	case 32:
		return I32
	default:
		return I64
	}
}

// UIntFromBits returns type code by bits.
func UIntFromBits(bits uint64) uint8 {
	switch bits {
	case 8:
		return U8
	case 16:
		return U16
	case 32:
		return U32
	default:
		return U64
	}
}

// FloatFromBits returns type code by bits.
func FloatFromBits(bits uint64) uint8 {
	switch bits {
	case 32:
		return F32
	default:
		return F64
	}
}

func init() {
	BIT_SIZE = strconv.IntSize
	switch BIT_SIZE {
	case 32:
		INT_CODE = I32
		UINT_CODE = U32
	case 64:
		INT_CODE = I64
		UINT_CODE = U64
	}
}

func VariadicToSlice(t Type) Type {
	t.Original = nil
	t.ComponentType = new(Type)
	*t.ComponentType = t
	t.Id = SLICE
	t.Kind = lex.PREFIX_SLICE + t.ComponentType.Kind
	return t
}

func FindGeneric(id string, generics []*GenericType) *GenericType {
	for _, g := range generics {
		if g.Id == id {
			return g
		}
	}
	return nil
}

func IsVoid(t Type) bool {
	return t.Id == VOID && !t.MultiTyped
}

func IsAllowForConst(t Type) bool {
	if !IsPure(t) {
		return false
	}
	switch t.Id {
	case STR, BOOL:
		return true
	default:
		return IsNumeric(t.Id)
	}
}

func IsVariadicable(t Type) bool { return IsSlice(t) }
func IsStruct(t Type) bool       { return t.Id == STRUCT }
func IsTrait(t Type) bool        { return t.Id == TRAIT }
func IsEnum(t Type) bool         { return t.Id == ENUM }

func DerefPtrOrRef(t Type) Type {
	t.Kind = t.Kind[1:]
	return t
}

func HasThisGeneric(generic *GenericType, t Type) bool {
	switch {
	case IsFn(t):
		f := t.Tag.(*Fn)
		for _, p := range f.Params {
			if HasThisGeneric(generic, p.Type) {
				return true
			}
		}
		return HasThisGeneric(generic, f.RetType.Type)
	case t.MultiTyped, IsMap(t):
		types := t.Tag.([]Type)
		for _, t := range types {
			if HasThisGeneric(generic, t) {
				return true
			}
		}
		return false
	case IsSlice(t), IsArray(t):
		return HasThisGeneric(generic, *t.ComponentType)
	}
	return IsThisGeneric(generic, t)
}

func HasGenerics(generics []*GenericType, t Type) bool {
	for _, g := range generics {
		if HasThisGeneric(g, t) {
			return true
		}
	}
	return false
}

func IsThisGeneric(generic *GenericType, t Type) bool {
	id, _ := t.KindId()
	return id == generic.Id
}

func IsGeneric(generics []*GenericType, t Type) bool {
	if t.Id != ID {
		return false
	}
	for _, generic := range generics {
		if IsThisGeneric(generic, t) {
			return true
		}
	}
	return false
}

func IsExplicitPtr(t Type) bool {
	if t.Kind == "" {
		return false
	}
	return t.Kind[0] == '*' && !IsUnsafePtr(t)
}

func IsUnsafePtr(t Type) bool {
	if t.Id != UNSAFE {
		return false
	}
	return len(t.Kind)-len(lex.KND_UNSAFE) == 1
}

func IsPtr(t Type) bool {
	return IsExplicitPtr(t) || IsUnsafePtr(t)
}

func IsRef(t Type) bool {
	return t.Kind != "" && t.Kind[0] == '&'
}

func IsSlice(t Type) bool {
	return t.Id == SLICE && strings.HasPrefix(t.Kind, lex.PREFIX_SLICE)
}

func IsArray(t Type) bool {
	return t.Id == ARRAY && strings.HasPrefix(t.Kind, lex.PREFIX_ARRAY)
}

func IsMap(t Type) bool {
	if t.Kind == "" || t.Id != MAP {
		return false
	}
	return t.Kind[0] == '[' && t.Kind[len(t.Kind)-1] == ']'
}

func IsFn(t Type) bool {
	return t.Id == FN &&
		(strings.HasPrefix(t.Kind, lex.KND_FN) ||
			strings.HasPrefix(t.Kind, lex.KND_UNSAFE+" "+lex.KND_FN))
}

// Includes single ptr types.
func IsPure(t Type) bool {
	return !IsPtr(t) &&
		!IsRef(t) &&
		!IsSlice(t) &&
		!IsArray(t) &&
		!IsMap(t) &&
		!IsFn(t)
}

func ValidForRef(t Type) bool {
	return !(IsTrait(t) ||
		IsEnum(t) ||
		IsPtr(t) ||
		IsRef(t) ||
		IsSlice(t) ||
		IsArray(t))
}

func IsMut(t Type) bool {
	return IsSlice(t) || IsPtr(t) || IsRef(t)
}

func GetAccessor(t Type) string {
	if IsRef(t) || IsPtr(t) {
		return "->"
	}
	return lex.KND_DOT
}

func IsNilCompatible(t Type) bool {
	return t.Id == NIL ||
		IsFn(t) ||
		IsPtr(t) ||
		IsSlice(t) ||
		IsTrait(t) ||
		IsMap(t)
}

func IsLvalue(t Type) bool {
	return IsRef(t) || IsPtr(t) || IsSlice(t) || IsMap(t)
}

func Equals(t1, t2 Type) bool {
	return t1.Id == t2.Id && t1.Kind == t2.Kind
}
