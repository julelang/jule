package x

// Data type (built-in) constants.
const (
	Void     uint8 = 0
	Int8     uint8 = 1
	Int16    uint8 = 2
	Int32    uint8 = 3
	Int64    uint8 = 4
	UInt8    uint8 = 5
	UInt16   uint8 = 6
	UInt32   uint8 = 7
	UInt64   uint8 = 8
	Bool     uint8 = 9
	Str      uint8 = 10
	Float32  uint8 = 11
	Float64  uint8 = 12
	Any      uint8 = 13
	Rune     uint8 = 14
	Name     uint8 = 15
	Function uint8 = 16
	Nil      uint8 = 17
)

// TypeGreaterThan reports type one is greater than type two or not.
func TypeGreaterThan(t1, t2 uint8) bool {
	switch t1 {
	case Int16:
		return t2 == Int8
	case Int32:
		return t2 == Int8 ||
			t2 == Int16
	case Int64:
		return t2 == Int8 ||
			t2 == Int16 ||
			t2 == Int32
	case UInt16:
		return t2 == UInt8
	case UInt32:
		return t2 == UInt8 ||
			t2 == UInt16
	case UInt64:
		return t2 == UInt8 ||
			t2 == UInt16 ||
			t2 == UInt32
	case Float32:
		return t2 != Any && t2 != Float64
	case Float64:
		return t2 != Any
	}
	return false
}

// TypeAreCompatible reports type one and type two is compatible or not.
func TypesAreCompatible(t1, t2 uint8, ignoreany bool) bool {
	if !ignoreany && t2 == Any {
		return true
	}
	switch t1 {
	case Any:
		if ignoreany {
			return false
		}
		return true
	case Int8:
		return t2 == Int8 ||
			t2 == Int16 ||
			t2 == Int32 ||
			t2 == Int64 ||
			t2 == Float32 ||
			t2 == Float64
	case Int16:
		return t2 == Int16 ||
			t2 == Int32 ||
			t2 == Int64 ||
			t2 == Float32 ||
			t2 == Float64
	case Int32:
		return t2 == Int32 ||
			t2 == Int64 ||
			t2 == Float32 ||
			t2 == Float64
	case Int64:
		return t2 == Int64 ||
			t2 == Float32 ||
			t2 == Float64
	case UInt8:
		return t2 == UInt8 ||
			t2 == UInt16 ||
			t2 == UInt32 ||
			t2 == UInt64 ||
			t2 == Float32 ||
			t2 == Float64
	case UInt16:
		return t2 == UInt16 ||
			t2 == UInt32 ||
			t2 == UInt64 ||
			t2 == Float32 ||
			t2 == Float64
	case UInt32:
		return t2 == UInt32 ||
			t2 == UInt64 ||
			t2 == Float32 ||
			t2 == Float64
	case UInt64:
		return t2 == UInt64 ||
			t2 == Float32 ||
			t2 == Float64
	case Bool:
		return t2 == Bool
	case Str:
		return t2 == Str
	case Float32:
		return t2 == Float32 ||
			t2 == Float64
	case Float64:
		return t2 == Float64
	case Rune:
		return t2 == Rune ||
			t2 == Int32 ||
			t2 == Int64 ||
			t2 == UInt16 ||
			t2 == UInt32 ||
			t2 == UInt64
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
func IsNumericType(t uint8) bool {
	return IsIntegerType(t) || IsFloatType(t)
}

// IsFloatType reports type is float or not.
func IsFloatType(t uint8) bool {
	return t == Float32 || t == Float64
}

// IsSignedNumericType reports type is signed numeric or not.
func IsSignedNumericType(t uint8) bool {
	return IsSignedIntegerType(t) ||
		t == Float32 ||
		t == Float64
}

// IsSignedIntegerType reports type is signed integer or not.
func IsSignedIntegerType(t uint8) bool {
	return t == Int8 ||
		t == Int16 ||
		t == Int32 ||
		t == Int64
}

// IsUnsignedNumericType reports type is unsigned numeric or not.
func IsUnsignedNumericType(t uint8) bool {
	return t == UInt8 ||
		t == UInt16 ||
		t == UInt32 ||
		t == UInt64
}

// TypeFromName returns type name of specified type code.
func TypeFromName(name string) uint8 {
	switch name {
	case "int8":
		return Int8
	case "int16":
		return Int16
	case "int32":
		return Int32
	case "int64":
		return Int64
	case "uint8":
		return UInt8
	case "uint16":
		return UInt16
	case "uint32":
		return UInt32
	case "uint64":
		return UInt64
	case "str":
		return Str
	case "bool":
		return Bool
	case "float32":
		return Float32
	case "float64":
		return Float64
	case "any":
		return Any
	case "rune":
		return Rune
	}
	return 0
}

func CxxTypeNameFromType(typeCode uint8) string {
	switch typeCode {
	case Void:
		return "void"
	case Int8:
		return "int8"
	case Int16:
		return "int16"
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case UInt8:
		return "uint8"
	case UInt16:
		return "uint16"
	case UInt32:
		return "uint32"
	case UInt64:
		return "uint64"
	case Bool:
		return "bool"
	case Float32:
		return "float32"
	case Float64:
		return "float64"
	case Any:
		return "any"
	case Str:
		return "str"
	case Rune:
		return "rune"
	}
	return "" // Unreachable code.
}

// DefaultValueOfType returns default value of specified type.
//
// Special case is:
//  DefaultValueOfType(t) = "nil" if t is invalid
//  DefaultValueOfType(t) = "nil" if t is not have default value
func DefaultValueOfType(code uint8) string {
	if IsNumericType(code) {
		return "0"
	}
	switch code {
	case Bool:
		return "false"
	case Str:
		return `""`
	case Rune:
		return `'\0'`
	}
	return "nil"
}
