package parser

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/pkg/juleapi"
	"github.com/jule-lang/jule/pkg/juletype"
)

type valueEvaluator struct {
	token lex.Token
	model *exprModel
	p     *Parser
}

func strModel(v value) iExpr {
	content := v.expr.(string)
	if israwstr(content) {
		return exprNode{juleapi.ToRawStr([]byte(content))}
	}
	return exprNode{juleapi.ToStr([]byte(content))}
}

func boolModel(v value) iExpr {
	if v.expr.(bool) {
		return exprNode{lex.KND_TRUE}
	}
	return exprNode{lex.KND_FALSE}
}

func getModel(v value) iExpr {
	switch v.expr.(type) {
	case string:
		return strModel(v)
	case bool:
		return boolModel(v)
	default:
		return numericModel(v)
	}
}

func numericModel(v value) iExpr {
	switch t := v.expr.(type) {
	case uint64:
		fmt := strconv.FormatUint(t, 10)
		return exprNode{fmt + "LLU"}
	case int64:
		fmt := strconv.FormatInt(t, 10)
		return exprNode{fmt + "LL"}
	case float64:
		switch {
		case normalize(&v):
			return numericModel(v)
		case v.data.Type.Id == juletype.F32:
			return exprNode{fmt.Sprint(t) + "f"}
		case v.data.Type.Id == juletype.F64:
			return exprNode{fmt.Sprint(t)}
		}
	}
	return nil
}

func (ve *valueEvaluator) str() value {
	var v value
	v.constExpr = true
	v.data.Value = ve.token.Kind
	v.data.Type.Id = juletype.STR
	v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	content := ve.token.Kind[1 : len(ve.token.Kind)-1]
	v.expr = content
	v.model = strModel(v)
	ve.model.appendSubNode(v.model)
	return v
}

func toCharLiteral(kind string) (string, bool) {
	kind = kind[1 : len(kind)-1]
	isByte := false
	switch {
	case len(kind) == 1 && kind[0] <= 255:
		isByte = true
	case kind[0] == '\\' && kind[1] == 'x':
		isByte = true
	case kind[0] == '\\' && kind[1] >= '0' && kind[1] <= '7':
		isByte = true
	}
	return kind, isByte
}

func (ve *valueEvaluator) char() value {
	var v value
	v.constExpr = true
	v.data.Value = ve.token.Kind
	content, isByte := toCharLiteral(ve.token.Kind)
	if isByte {
		v.data.Type.Id = juletype.U8
	} else { // rune
		v.data.Type.Id = juletype.I32
	}
	content = juleapi.ToRune([]byte(content))
	v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	v.expr, _ = strconv.ParseInt(content[2:], 16, 64)
	v.model = exprNode{content}
	ve.model.appendSubNode(v.model)
	return v
}

func (ve *valueEvaluator) bool() value {
	var v value
	v.constExpr = true
	v.data.Value = ve.token.Kind
	v.data.Type.Id = juletype.BOOL
	v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	v.expr = ve.token.Kind == lex.KND_TRUE
	v.model = boolModel(v)
	ve.model.appendSubNode(v.model)
	return v
}

func (ve *valueEvaluator) nil() value {
	var v value
	v.constExpr = true
	v.data.Value = ve.token.Kind
	v.data.Type.Id = juletype.NIL
	v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	v.expr = nil
	v.model = exprNode{ve.token.Kind}
	ve.model.appendSubNode(v.model)
	return v
}

func normalize(v *value) (normalized bool) {
	switch {
	case !v.constExpr:
		return
	case integerAssignable(juletype.U64, *v):
		v.data.Type.Id = juletype.U64
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		v.expr = tonumu(v.expr)
		bitize(v)
		return true
	case integerAssignable(juletype.I64, *v):
		v.data.Type.Id = juletype.I64
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		v.expr = tonums(v.expr)
		bitize(v)
		return true
	}
	return
}

func (ve *valueEvaluator) float() value {
	var v value
	v.data.Value = ve.token.Kind
	v.data.Type.Id = juletype.F64
	v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	v.expr, _ = strconv.ParseFloat(v.data.Value, 64)
	return v
}

func (ve *valueEvaluator) integer() value {
	var v value
	v.data.Value = ve.token.Kind
	var bigint big.Int
	switch {
	case strings.HasPrefix(ve.token.Kind, "0x"):
		_, _ = bigint.SetString(ve.token.Kind[2:], 16)
	case strings.HasPrefix(ve.token.Kind, "0b"):
		_, _ = bigint.SetString(ve.token.Kind[2:], 2)
	case ve.token.Kind[0] == '0':
		_, _ = bigint.SetString(ve.token.Kind[1:], 8)
	default:
		_, _ = bigint.SetString(ve.token.Kind, 10)
	}
	if bigint.IsInt64() {
		v.expr = bigint.Int64()
	} else {
		v.expr = bigint.Uint64()
	}
	bitize(&v)
	return v
}

func (ve *valueEvaluator) numeric() value {
	var v value
	if isfloat(ve.token.Kind) {
		v = ve.float()
	} else {
		v = ve.integer()
	}
	v.constExpr = true
	v.model = numericModel(v)
	ve.model.appendSubNode(v.model)
	return v
}

func make_value_from_var(v *Var) (val value) {
	val.data.Value = v.Id
	val.data.Type = v.Type
	val.constExpr = v.Const
	val.data.Token = v.Token
	val.lvalue = !val.constExpr
	val.mutable = v.Mutable
	if val.constExpr {
		val.expr = v.ExprTag
		val.model = v.Expr.Model
	}
	return
}

func (ve *valueEvaluator) varId(id string, variable *Var, global bool) (v value) {
	variable.Used = true
	v = make_value_from_var(variable)
	if v.constExpr {
		ve.model.appendSubNode(v.model)
	} else {
		if variable.Id == lex.KND_SELF && !typeIsRef(variable.Type) {
			ve.model.appendSubNode(exprNode{"(*this)"})
		} else {
			ve.model.appendSubNode(exprNode{variable.OutId()})
		}
		ve.p.eval.has_error = ve.p.eval.has_error || typeIsVoid(v.data.Type)
	}
	return
}

func make_value_from_fn(f *models.Fn) (v value) {
	v.data.Value = f.Id
	v.data.Type.Id = juletype.FN
	v.data.Type.Tag = f
	v.data.Type.Kind = f.DataTypeString()
	v.data.Token = f.Token
	return
}

func (ve *valueEvaluator) funcId(id string, f *Fn) (v value) {
	f.used = true
	v = make_value_from_fn(f.Ast)
	ve.model.appendSubNode(exprNode{f.outId()})
	return
}

func (ve *valueEvaluator) enumId(id string, e *Enum) (v value) {
	e.Used = true
	v.data.Value = id
	v.data.Type.Id = juletype.ENUM
	v.data.Type.Kind = e.Id
	v.data.Type.Tag = e
	v.data.Token = e.Token
	v.constExpr = true
	v.is_type = true
	// If built-in.
	if e.Token.Id == lex.ID_NA {
		ve.model.appendSubNode(exprNode{juleapi.OutId(id, nil)})
	} else {
		ve.model.appendSubNode(exprNode{juleapi.OutId(id, e.Token.File)})
	}
	return
}

func make_value_from_struct(s *structure) (v value) {
	v.data.Value = s.Ast.Id
	v.data.Type.Id = juletype.STRUCT
	v.data.Type.Tag = s
	v.data.Type.Kind = s.Ast.Id
	v.data.Type.Token = s.Ast.Token
	v.data.Token = s.Ast.Token
	v.is_type = true
	return
}

func (ve *valueEvaluator) structId(id string, s *structure) (v value) {
	s.Used = true
	v = make_value_from_struct(s)
	// If builtin.
	if s.Ast.Token.Id == lex.ID_NA {
		ve.model.appendSubNode(exprNode{juleapi.OutId(id, nil)})
	} else {
		ve.model.appendSubNode(exprNode{juleapi.OutId(id, s.Ast.Token.File)})
	}
	return
}

func (ve *valueEvaluator) typeId(id string, t *TypeAlias) (_ value, _ bool) {
	dt, ok := ve.p.realType(t.Type, true)
	if !ok {
		return
	}
	if typeIsStruct(dt) {
		return ve.structId(id, dt.Tag.(*structure)), true
	}
	return
}

func (ve *valueEvaluator) id() (_ value, ok bool) {
	id := ve.token.Kind

	v, _ := ve.p.blockVarById(id)
	if v != nil {
		return ve.varId(id, v, false), true
	} else {
		v, _, _ := ve.p.globalById(id)
		if v != nil {
			return ve.varId(id, v, true), true
		}
	}

	f, _, _ := ve.p.FuncById(id)
	if f != nil {
		return ve.funcId(id, f), true
	}

	e, _, _ := ve.p.enumById(id)
	if e != nil {
		return ve.enumId(id, e), true
	}

	s, _, _ := ve.p.structById(id)
	if s != nil {
		return ve.structId(id, s), true
	}

	t, _, _ := ve.p.typeById(id)
	if t != nil {
		return ve.typeId(id, t)
	}

	ve.p.eval.pusherrtok(ve.token, "id_not_exist", id)
	return
}
