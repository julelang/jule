package parser

import (
	"github.com/the-xlang/xxc/ast"
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/xbits"
	"github.com/the-xlang/xxc/pkg/xtype"
)

func isstr(value string) bool    { return value[0] == '"' || israwstr(value) }
func israwstr(value string) bool { return value[0] == '`' }
func isRune(value string) bool   { return value[0] == '\'' }
func isnil(value string) bool    { return value == tokens.NIL }
func isbool(value string) bool   { return value == tokens.TRUE || value == tokens.FALSE }

func isBoolExpr(val value) bool {
	switch {
	case typeIsNilCompatible(val.ast.Type):
		return true
	case val.ast.Type.Id == xtype.Bool && typeIsSingle(val.ast.Type):
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
	return code == xtype.Str
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
	if xtype.IsSignedNumericType(v.Type.Id) {
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

func defaultValueOfType(t DataType) string {
	if typeIsNilCompatible(t) {
		return "nil"
	}
	return xtype.DefaultValOfType(t.Id)
}
