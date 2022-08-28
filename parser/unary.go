package parser

import (
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/juletype"
)

type unary struct {
	token lex.Token
	toks  []lex.Token
	model *exprModel
	p     *Parser
}

func (u *unary) minus() value {
	v := u.p.eval.process(u.toks, u.model)
	if !typeIsPure(v.data.Type) || !juletype.IsNumeric(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.token, "invalid_type_unary_operator", tokens.MINUS)
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
	if !typeIsPure(v.data.Type) || !juletype.IsNumeric(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.token, "invalid_type_unary_operator", tokens.PLUS)
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
	if !typeIsPure(v.data.Type) || !juletype.IsInteger(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.token, "invalid_type_unary_operator", tokens.CARET)
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
		u.p.eval.pusherrtok(u.token, "invalid_type_unary_operator", tokens.EXCLAMATION)
	} else if v.constExpr {
		v.expr = !v.expr.(bool)
		v.model = boolModel(v)
	}
	v.data.Type.Id = juletype.Bool
	v.data.Type.Kind = tokens.BOOL
	return v
}

func (u *unary) star() value {
	v := u.p.eval.process(u.toks, u.model)
	v.constExpr = false
	v.lvalue = true
	if !typeIsExplicitPtr(v.data.Type) {
		u.p.eval.pusherrtok(u.token, "invalid_type_unary_operator", tokens.STAR)
	} else {
		v.data.Type.Kind = v.data.Type.Kind[1:]
	}
	if u.p.eval.unsafe_allowed() {
		// Use unsafe deferencing
		u.model.nodes[u.model.index].nodes[0] = exprNode{"~"}
	}
	return v
}

func (u *unary) amper() value {
	v := u.p.eval.process(u.toks, u.model)
	nodes := &u.model.nodes[u.model.index].nodes
	switch {
	case valIsStructIns(v):
		s := v.data.Type.Tag.(*structure)
		// Is not struct literal
		if s.Ast.Id != v.data.Value {
			break
		}
		(*nodes)[0] = exprNode{"__julec_guaranteed_ptr(new "}
		goto end
	case !canGetPtr(v):
		u.p.eval.pusherrtok(u.token, "invalid_expr_unary_operator", tokens.AMPER)
		return v
	}
	if u.p.eval.unsafe_allowed() {
		(*nodes)[0] = exprNode{"__julec_unsafe_ptr(&"}
	} else if v.heapMust {
		(*nodes)[0] = exprNode{"__julec_ptr(&"}
	} else {
		(*nodes)[0] = exprNode{"__julec_never_guarantee_ptr(&"}
	}
end:
	v.constExpr = false
	v.lvalue = true
	v.data.Type.Kind = tokens.STAR + v.data.Type.Kind
	*nodes = append(*nodes, exprNode{tokens.RPARENTHESES})
	return v
}
