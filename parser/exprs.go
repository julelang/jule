package parser

import (
	"strings"

	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/pkg/julebits"
	"github.com/julelang/jule/pkg/juletype"
)

func check_value_for_indexing(v value) (err_key string) {
	switch {
	case !type_is_pure(v.data.Type):
		return "invalid_expr"
	case !juletype.IsInteger(v.data.Type.Id):
		return "invalid_expr"
	case v.constExpr && tonums(v.expr) < 0:
		return "overflow_limits"
	default:
		return ""
	}
}

func indexingExprModel(i iExpr) iExpr {
	if i == nil {
		return i
	}
	var model strings.Builder
	model.WriteString("static_cast<")
	model.WriteString(juletype.CppId(juletype.INT))
	model.WriteString(">(")
	model.WriteString(i.String())
	model.WriteByte(')')
	return exprNode{model.String()}
}

func isstr(s string) bool {
	return s != "" && (s[0] == '"' || israwstr(s))
}

func israwstr(s string) bool {
	return s != "" && s[0] == '`'
}

func ischar(s string) bool {
	return s != "" && s[0] == '\''
}

func isnil(s string) bool {
	return s == lex.KND_NIL
}

func isbool(s string) bool {
	return s == lex.KND_TRUE || s == lex.KND_FALSE
}

func valIsEnumType(v value) bool {
	return v.is_type && type_is_enum(v.data.Type)
}

func isBoolExpr(v value) bool {
	return type_is_pure(v.data.Type) && v.data.Type.Id == juletype.BOOL
}

func isfloat(s string) bool {
	if strings.HasPrefix(s, "0x") {
		return strings.ContainsAny(s, ".pP")
	}
	return strings.ContainsAny(s, ".eE")
}

func canGetPtr(v value) bool {
	if !v.lvalue || v.constExpr {
		return false
	}
	switch v.data.Type.Id {
	case juletype.FN, juletype.ENUM:
		return false
	default:
		return v.data.Token.Id == lex.ID_IDENT
	}
}

func valIsStructIns(val value) bool {
	return !val.is_type && type_is_struct(val.data.Type)
}

func valIsTraitIns(val value) bool {
	return !val.is_type && type_is_trait(val.data.Type)
}

func isForeachIterExpr(val value) bool {
	switch {
	case type_is_slc(val.data.Type),
		type_is_array(val.data.Type),
		type_is_map(val.data.Type):
		return true
	case !type_is_pure(val.data.Type):
		return false
	}
	code := val.data.Type.Id
	return code == juletype.STR
}

func isConstNumeric(v string) bool {
	if v == "" {
		return false
	}
	return v[0] == '-' || (v[0] >= '0' && v[0] <= '9')
}

func isConstExpression(v string) bool {
	return isConstNumeric(v) || isstr(v) || ischar(v) || isnil(v) || isbool(v)
}

func checkFloatBit(v models.Data, bit int) bool {
	if bit == 0 {
		return false
	}
	return julebits.CheckBitFloat(v.Value, bit)
}

func validExprForConst(v value) bool {
	return v.constExpr
}

func okForShifting(v value) bool {
	if !type_is_pure(v.data.Type) ||
		!juletype.IsInteger(v.data.Type.Id) {
		return false
	}
	if !v.constExpr {
		return true
	}
	switch t := v.expr.(type) {
	case int64:
		return t >= 0
	case uint64:
		return true
	}
	return false
}
