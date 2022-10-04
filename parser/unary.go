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
	t     *Parser
}

func (u *unary) minus() value {
	v := u.t.eval.process(u.toks, u.model)
	if !typeIsPure(v.data.Type) || !juletype.IsNumeric(v.data.Type.Id) {
		u.t.eval.pusherrtok(u.token, "invalid_expr_unary_operator", tokens.MINUS)
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
	v := u.t.eval.process(u.toks, u.model)
	if !typeIsPure(v.data.Type) || !juletype.IsNumeric(v.data.Type.Id) {
		u.t.eval.pusherrtok(u.token, "invalid_expr_unary_operator", tokens.PLUS)
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
	v := u.t.eval.process(u.toks, u.model)
	if !typeIsPure(v.data.Type) || !juletype.IsInteger(v.data.Type.Id) {
		u.t.eval.pusherrtok(u.token, "invalid_expr_unary_operator", tokens.CARET)
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
	v := u.t.eval.process(u.toks, u.model)
	if !isBoolExpr(v) {
		u.t.eval.pusherrtok(u.token, "invalid_expr_unary_operator", tokens.EXCLAMATION)
	} else if v.constExpr {
		v.expr = !v.expr.(bool)
		v.model = boolModel(v)
	}
	v.data.Type.Id = juletype.Bool
	v.data.Type.Kind = tokens.BOOL
	return v
}

func (u *unary) star() value {
	if !u.t.eval.unsafe_allowed() {
		u.t.pusherrtok(u.token, "unsafe_behavior_at_out_of_unsafe_scope")
	}
	v := u.t.eval.process(u.toks, u.model)
	v.constExpr = false
	v.lvalue = true
	switch {
	case !typeIsExplicitPtr(v.data.Type):
		u.t.eval.pusherrtok(u.token, "invalid_expr_unary_operator", tokens.STAR)
		goto end
	}
	v.data.Type.Kind = v.data.Type.Kind[1:]
end:
	v.data.Value = " "
	return v
}

func (u *unary) amper() value {
	v := u.t.eval.process(u.toks, u.model)
	v.constExpr = false
	v.lvalue = true
	nodes := &u.model.nodes[u.model.index].nodes
	switch {
	case valIsStructIns(v):
		s := v.data.Type.Tag.(*structure)
		// Is not struct literal
		if s.Ast.Id != v.data.Value {
			break
		}
		var alloc_model exprNode
		alloc_model.value = "__julec_new_structure<"
		alloc_model.value += s.OutId()
		alloc_model.value += ">(new( std::nothrow ) "
		(*nodes)[0] = alloc_model
		last := &(*nodes)[len(*nodes)-1]
		*last = exprNode{(*last).String() + ")"}
		v.data.Type.Kind = tokens.AMPER + v.data.Type.Kind
		v.mutable = true
		return v
	case typeIsRef(v.data.Type):
		model := exprNode{(*nodes)[1].String() + "._alloc"}
		*nodes = nil
		*nodes = make([]iExpr, 1)
		(*nodes)[0] = model
		v.data.Type.Kind = tokens.STAR + un_ptr_or_ref_type(v.data.Type).Kind
		return v
	case !canGetPtr(v):
		u.t.eval.pusherrtok(u.token, "invalid_expr_unary_operator", tokens.AMPER)
		return v
	}
	v.data.Type.Kind = tokens.STAR + v.data.Type.Kind
	return v
}
