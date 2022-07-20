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
	if !typeIsPure(v.data.Type) || !xtype.IsNumericType(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.MINUS)
	}
	if isConstNumeric(v.data.Value) {
		v.data.Value = tokens.MINUS + v.data.Value
	}
	return v
}

func (u *unary) plus() value {
	v := u.p.eval.process(u.toks, u.model)
	if !typeIsPure(v.data.Type) || !xtype.IsNumericType(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.PLUS)
	}
	return v
}

func (u *unary) tilde() value {
	v := u.p.eval.process(u.toks, u.model)
	if !typeIsPure(v.data.Type) || !xtype.IsIntegerType(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.TILDE)
	}
	return v
}

func (u *unary) logicalNot() value {
	v := u.p.eval.process(u.toks, u.model)
	if !isBoolExpr(v) {
		u.p.eval.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.EXCLAMATION)
	}
	v.data.Type.Id = xtype.Bool
	v.data.Type.Kind = tokens.BOOL
	return v
}

func (u *unary) star() value {
	v := u.p.eval.process(u.toks, u.model)
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
