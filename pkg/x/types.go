package x

// Data type (built-in) constants.
const (
	Void    uint8 = 0
	Int8    uint8 = 1
	Int16   uint8 = 2
	Int32   uint8 = 3
	Int64   uint8 = 4
	UInt8   uint8 = 5
	UInt16  uint8 = 6
	UInt32  uint8 = 7
	UInt64  uint8 = 8
	Boolean uint8 = 9
	String  uint8 = 10
	Float32 uint8 = 11
	Float64 uint8 = 12
)

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
	case "string":
		return String
	case "bool":
		return Boolean
	case "float32":
		return Float32
	case "float64":
		return Float64
	}
	return 0 // Unreachable code.
}
