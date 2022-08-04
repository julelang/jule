package parser

import (
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xtype"
)

type unary struct {
	tok   Tok
	toks  Toks
	model *exprModel
	p     *Parser
}

func (u *unary) minus() value {
	v := u.p.eval.process(u.toks, u.model)
	if !typeIsPure(v.data.Type) || !xtype.IsNumeric(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.MINUS)
	}
	if v.constExpr {
		v.data.Value = tokens.MINUS + v.data.Value
		switch t := v.expr.(type) {
		case float64:
			v.expr = -t
		case int64:
			v.expr = -t
		case uint64:
			v.expr = -t
		}
		v.model = numericModel(v)
	}
	return v
}

func (u *unary) plus() value {
	v := u.p.eval.process(u.toks, u.model)
	if !typeIsPure(v.data.Type) || !xtype.IsNumeric(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.PLUS)
	}
	if v.constExpr {
		switch t := v.expr.(type) {
		case float64:
			v.expr = +t
		case int64:
			v.expr = +t
		case uint64:
			v.expr = +t
		}
		v.model = numericModel(v)
	}
	return v
}

func (u *unary) caret() value {
	v := u.p.eval.process(u.toks, u.model)
	if !typeIsPure(v.data.Type) || !xtype.IsInteger(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.CARET)
	}
	if v.constExpr {
		switch t := v.expr.(type) {
		case int64:
			v.expr = ^t
		case uint64:
			v.expr = ^t
		}
		v.model = numericModel(v)
	}
	return v
}

func (u *unary) logicalNot() value {
	v := u.p.eval.process(u.toks, u.model)
	if !isBoolExpr(v) {
		u.p.eval.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.EXCLAMATION)
	} else if v.constExpr {
		v.expr = !v.expr.(bool)
		v.model = boolModel(v)
	}
	v.data.Type.Id = xtype.Bool
	v.data.Type.Kind = tokens.BOOL
	return v
}

func (u *unary) star() value {
	v := u.p.eval.process(u.toks, u.model)
	v.constExpr = false
	v.lvalue = true
	if !typeIsExplicitPtr(v.data.Type) {
		u.p.eval.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.STAR)
	} else {
		v.data.Type.Kind = v.data.Type.Kind[1:]
	}
	return v
}

func (u *unary) amper() value {
	v := u.p.eval.process(u.toks, u.model)
	v.constExpr = false
	switch {
	case typeIsFunc(v.data.Type):
		mainNode := &u.model.nodes[u.model.index]
		mainNode.nodes = mainNode.nodes[1:] // Remove unary operator from model
		node := &u.model.nodes[u.model.index].nodes[0]
		switch t := (*node).(type) {
		case anonFuncExpr:
			if t.capture == xapi.LambdaByReference {
				u.p.eval.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.AMPER)
				break
			}
			t.capture = xapi.LambdaByReference
			*node = t
		default:
			u.p.eval.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.AMPER)
		}
	default:
		if !canGetPtr(v) {
			u.p.eval.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.AMPER)
		}
		v.lvalue = true
		v.data.Type.Kind = tokens.STAR + v.data.Type.Kind
	}
	return v
}
