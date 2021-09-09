package parser

import "github.com/the-xlang/x/pkg/x"

func typeFromName(name string) uint {
	switch name {
	case "int8":
		return x.Int8
	case "int16":
		return x.Int16
	case "int32":
		return x.Int32
	case "int64":
		return x.Int64
	case "uint8":
		return x.UInt8
	case "uint16":
		return x.UInt16
	case "uint32":
		return x.UInt32
	case "uint64":
		return x.UInt64
	}
	return 0 // Unreachable code.
}

func cxxTypeNameFromType(typeCode uint) string {
	switch typeCode {
	case x.Void:
		return "void"
	case x.Int8:
		return "signed char"
	case x.Int16:
		return "short"
	case x.Int32:
		return "int"
	case x.Int64:
		return "long"
	case x.UInt8:
		return "unsigned char"
	case x.UInt16:
		return "unsigned short"
	case x.UInt32:
		return "unsigned int"
	case x.UInt64:
		return "unsigned long"
	}
	return "" // Unreachable code.
}
