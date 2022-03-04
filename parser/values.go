package parser

import (
	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/pkg/x"
	"github.com/the-xlang/x/pkg/xbits"
)

func isstr(value string) bool  { return value[0] == '"' }
func isrune(value string) bool { return value[0] == '\'' }
func isnil(value string) bool  { return value == "nil" }

func isbool(value string) bool {
	return value == "true" || value == "false"
}

func isBoolExpr(val value) bool {
	switch {
	case typeIsNilCompatible(val.ast.Type):
		return true
	case val.ast.Type.Code == x.Bool && typeIsSingle(val.ast.Type):
		return true
	}
	return false
}

func isForeachIterExpr(val value) bool {
	switch {
	case typeIsArray(val.ast.Type):
		return true
	case !typeIsSingle(val.ast.Type):
		return false
	}
	code := val.ast.Type.Code
	return code == x.Str
}

func isConstNum(v string) bool {
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

func checkFloatBit(v ast.ValueAST, bit int) bool {
	if bit == 0 {
		return false
	}
	return xbits.CheckBitFloat(v.Value, bit)
}
