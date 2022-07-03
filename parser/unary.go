package parser

import (
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xtype"
)

type unary struct {
	tok    Tok
	toks   Toks
	model  *exprModel
	parser *Parser
}

func (u *unary) minus() value {
	v := u.parser.evalExprPart(u.toks, u.model)
	if !typeIsSingle(v.ast.Type) || !xtype.IsNumericType(v.ast.Type.Id) {
		u.parser.pusherrtok(u.tok, "invalid_type_unary_operator", '-')
	}
	if isConstNum(v.ast.Data) {
		v.ast.Data = tokens.MINUS + v.ast.Data
	}
	return v
}

func (u *unary) plus() value {
	v := u.parser.evalExprPart(u.toks, u.model)
	if !typeIsSingle(v.ast.Type) || !xtype.IsNumericType(v.ast.Type.Id) {
		u.parser.pusherrtok(u.tok, "invalid_type_unary_operator", '+')
	}
	return v
}

func (u *unary) tilde() value {
	v := u.parser.evalExprPart(u.toks, u.model)
	if !typeIsSingle(v.ast.Type) || !xtype.IsIntegerType(v.ast.Type.Id) {
		u.parser.pusherrtok(u.tok, "invalid_type_unary_operator", '~')
	}
	return v
}

func (u *unary) logicalNot() value {
	v := u.parser.evalExprPart(u.toks, u.model)
	if !isBoolExpr(v) {
		u.parser.pusherrtok(u.tok, "invalid_type_unary_operator", '!')
	}
	v.ast.Type.Id = xtype.Bool
	v.ast.Type.Val = tokens.BOOL
	return v
}

func (u *unary) star() value {
	v := u.parser.evalExprPart(u.toks, u.model)
	v.lvalue = true
	if !typeIsExplicitPtr(v.ast.Type) {
		u.parser.pusherrtok(u.tok, "invalid_type_unary_operator", '*')
	} else {
		v.ast.Type.Val = v.ast.Type.Val[1:]
	}
	return v
}

func (u *unary) amper() value {
	v := u.parser.evalExprPart(u.toks, u.model)
	switch {
	case typeIsFunc(v.ast.Type):
		mainNode := &u.model.nodes[u.model.index]
		mainNode.nodes = mainNode.nodes[1:] // Remove unary operator from model
		node := &u.model.nodes[u.model.index].nodes[0]
		switch t := (*node).(type) {
		case anonFuncExpr:
			if t.capture == xapi.LambdaByReference {
				u.parser.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.AMPER)
				break
			}
			t.capture = xapi.LambdaByReference
			*node = t
		default:
			u.parser.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.AMPER)
		}
	default:
		if !canGetPtr(v) {
			u.parser.pusherrtok(u.tok, "invalid_type_unary_operator", tokens.AMPER)
		}
		v.lvalue = true
		v.ast.Type.Val = tokens.STAR + v.ast.Type.Val
	}
	return v
}
