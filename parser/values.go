package parser

import (
	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/pkg/x"
	"github.com/the-xlang/x/pkg/xbits"
)

// IsString reports value is string or not.
func IsString(value string) bool { return value[0] == '"' }

// IsRune reports value is rune of not.
func IsRune(value string) bool { return value[0] == '\'' }

// IsBoolean reports value is boolean or not.
func IsBoolean(value string) bool {
	return value == "true" || value == "false"
}

// IsNull reports value is null or not.
func IsNull(value string) bool { return value == "null" }

func isConstantNumeric(v string) bool {
	if v == "" {
		return false
	}
	return v[0] >= '0' && v[0] <= '9'
}

func checkIntBit(v ast.ValueAST, bit int) bool {
	if bit == 0 {
		return false
	}
	if x.IsSignedNumericType(v.Type.Code) {
		return xbits.CheckBitInt(v.Value, bit)
	}
	return xbits.CheckBitUInt(v.Value, bit)
}
