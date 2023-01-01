package parser

import (
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

type unary struct {
	token lex.Token
	toks  []lex.Token
	model *exprModel
	p     *Parser
}

func (u *unary) minus() value {
	v := u.p.eval.process(u.toks, u.model)
	if !types.IsPure(v.data.Type) || !types.IsNumeric(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.token, "invalid_expr_unary_operator", lex.KND_MINUS)
	}
	if v.constant {
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
	if !types.IsPure(v.data.Type) || !types.IsNumeric(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.token, "invalid_expr_unary_operator", lex.KND_PLUS)
	}
	if v.constant {
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
	if !types.IsPure(v.data.Type) || !types.IsInteger(v.data.Type.Id) {
		u.p.eval.pusherrtok(u.token, "invalid_expr_unary_operator", lex.KND_CARET)
	}
	if v.constant {
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
	} else if v.constant {
		v.expr = !v.expr.(bool)
		v.model = boolModel(v)
	}
	v.data.Type.Id = types.BOOL
	v.data.Type.Kind = lex.KND_BOOL
	return v
}

func (u *unary) star() value {
	if !u.p.eval.unsafe_allowed() {
		u.p.pusherrtok(u.token, "unsafe_behavior_at_out_of_unsafe_scope")
	}
	v := u.p.eval.process(u.toks, u.model)
	v.constant = false
	v.lvalue = true
	switch {
	case !types.IsExplicitPtr(v.data.Type):
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
	v.constant = false
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
	case types.IsRef(v.data.Type):
		model := exprNode{(*nodes)[1].String() + "._alloc"}
		*nodes = nil
		*nodes = make([]iExpr, 1)
		(*nodes)[0] = model
		v.data.Type.Kind = lex.KND_STAR + types.DerefPtrOrRef(v.data.Type).Kind
		return v
	case !canGetPtr(v):
		u.p.eval.pusherrtok(u.token, "invalid_expr_unary_operator", lex.KND_AMPER)
		return v
	}
	v.data.Type.Kind = lex.KND_STAR + v.data.Type.Kind
	return v
}
