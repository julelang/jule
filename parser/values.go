package parser

import (
	"github.com/the-xlang/xxc/ast"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xbits"
)

func isstr(value string) bool    { return value[0] == '"' || israwstr(value) }
func israwstr(value string) bool { return value[0] == '`' }
func isRune(value string) bool   { return value[0] == '\'' }
func isnil(value string) bool    { return value == "nil" }
func isbool(value string) bool   { return value == "true" || value == "false" }

func isBoolExpr(val value) bool {
	switch {
	case typeIsNilCompatible(val.ast.Type):
		return true
	case val.ast.Type.Id == x.Bool && typeIsSingle(val.ast.Type):
		return true
	}
	return false
}

func isForeachIterExpr(val value) bool {
	switch {
	case typeIsArray(val.ast.Type),
		typeIsMap(val.ast.Type):
		return true
	case !typeIsSingle(val.ast.Type):
		return false
	}
	code := val.ast.Type.Id
	return code == x.Str
}

func isConstNum(v string) bool {
	if v == "" {
		return false
	}
	return v[0] >= '0' && v[0] <= '9'
}

func checkIntBit(v ast.Value, bit int) bool {
	if bit == 0 {
		return false
	}
	if x.IsSignedNumericType(v.Type.Id) {
		return xbits.CheckBitInt(v.Data, bit)
	}
	return xbits.CheckBitUInt(v.Data, bit)
}

func checkFloatBit(v ast.Value, bit int) bool {
	if bit == 0 {
		return false
	}
	return xbits.CheckBitFloat(v.Data, bit)
}
