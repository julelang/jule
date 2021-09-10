package parser

import "github.com/the-xlang/x/pkg/x"

func cxxTypeNameFromType(typeCode uint8) string {
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
