package parser

import (
	"strings"

	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

func check_value_for_indexing(v value) (err_key string) {
	switch {
	case !types.IsPure(v.data.Type):
		return "invalid_expr"
	case !types.IsInteger(v.data.Type.Id):
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
	model.WriteString(types.CppId(types.INT))
	model.WriteString(">(")
	model.WriteString(i.String())
	model.WriteByte(')')
	return exprNode{model.String()}
}

func valIsEnumType(v value) bool {
	return v.is_type && types.IsEnum(v.data.Type)
}

func isBoolExpr(v value) bool {
	return types.IsPure(v.data.Type) && v.data.Type.Id == types.BOOL
}

func canGetPtr(v value) bool {
	if !v.lvalue || v.constExpr {
		return false
	}
	switch v.data.Type.Id {
	case types.FN, types.ENUM:
		return false
	default:
		return v.data.Token.Id == lex.ID_IDENT
	}
}

func valIsStructIns(val value) bool {
	return !val.is_type && types.IsStruct(val.data.Type)
}

func valIsTraitIns(val value) bool {
	return !val.is_type && types.IsTrait(val.data.Type)
}

func isForeachIterExpr(val value) bool {
	switch {
	case types.IsSlice(val.data.Type),
		types.IsArray(val.data.Type),
		types.IsMap(val.data.Type):
		return true
	case !types.IsPure(val.data.Type):
		return false
	}
	code := val.data.Type.Id
	return code == types.STR
}

func checkFloatBit(v models.Data, bit int) bool {
	if bit == 0 {
		return false
	}
	return types.CheckBitFloat(v.Value, bit)
}

func validExprForConst(v value) bool {
	return v.constExpr
}

func okForShifting(v value) bool {
	if !types.IsPure(v.data.Type) ||
		!types.IsInteger(v.data.Type.Id) {
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
