package parser

import (
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/pkg/juletype"
)

type unary struct {
	token lex.Token
	toks  []lex.Token
	model *exprModel
	p     *Parser
}

func (u *unary) minus() value {
	v := u.p.eval.process(u.toks, u.model)
	if !type_is_pure(v.data.Type) || !juletype.IsNumeric(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.token, "invalid_expr_unary_operator", lex.KND_MINUS)
	}
	if v.constExpr {
		v.data.Value = lex.KND_MINUS + v.data.Value
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
	if !type_is_pure(v.data.Type) || !juletype.IsNumeric(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.token, "invalid_expr_unary_operator", lex.KND_PLUS)
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
	if !type_is_pure(v.data.Type) || !juletype.IsInteger(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.token, "invalid_expr_unary_operator", lex.KND_CARET)
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
		u.p.eval.pusherrtok(u.token, "invalid_expr_unary_operator", lex.KND_EXCL)
	} else if v.constExpr {
		v.expr = !v.expr.(bool)
		v.model = boolModel(v)
	}
	v.data.Type.Id = juletype.BOOL
	v.data.Type.Kind = lex.KND_BOOL
	return v
}

func (u *unary) star() value {
	if !u.p.eval.unsafe_allowed() {
		u.p.pusherrtok(u.token, "unsafe_behavior_at_out_of_unsafe_scope")
	}
	v := u.p.eval.process(u.toks, u.model)
	v.constExpr = false
	v.lvalue = true
	switch {
	case !type_is_explicit_ptr(v.data.Type):
		u.p.eval.pusherrtok(u.token, "invalid_expr_unary_operator", lex.KND_STAR)
		goto end
	}
	v.data.Type.Kind = v.data.Type.Kind[1:]
end:
	v.data.Value = " "
	return v
}

func (u *unary) amper() value {
	v := u.p.eval.process(u.toks, u.model)
	v.constExpr = false
	v.lvalue = true
	nodes := &u.model.nodes[u.model.index].nodes
	switch {
	case valIsStructIns(v):
		s := v.data.Type.Tag.(*Struct)
		// Is not struct literal
		if s.Id != v.data.Value {
			break
		}
		var alloc_model exprNode
		alloc_model.value = "__julec_new_structure<"
		alloc_model.value += s.OutId()
		alloc_model.value += ">(new( std::nothrow ) "
		(*nodes)[0] = alloc_model
		last := &(*nodes)[len(*nodes)-1]
		*last = exprNode{(*last).String() + ")"}
		v.data.Type.Kind = lex.KND_AMPER + v.data.Type.Kind
		v.mutable = true
		return v
	case type_is_ref(v.data.Type):
		model := exprNode{(*nodes)[1].String() + "._alloc"}
		*nodes = nil
		*nodes = make([]iExpr, 1)
		(*nodes)[0] = model
		v.data.Type.Kind = lex.KND_STAR + un_ptr_or_ref_type(v.data.Type).Kind
		return v
	case !canGetPtr(v):
		u.p.eval.pusherrtok(u.token, "invalid_expr_unary_operator", lex.KND_AMPER)
		return v
	}
	v.data.Type.Kind = lex.KND_STAR + v.data.Type.Kind
	return v
}
