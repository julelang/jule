package parser

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

type unary struct {
	token lex.Token
	toks  []lex.Token
	model *expr_model
	p     *Parser
}

func (u *unary) minus() value {
	v := u.p.eval.process(u.toks, u.model)
	if !types.IsPure(v.data.DataType) || !types.IsNumeric(v.data.DataType.Id) {
		u.p.eval.push_err_tok(u.token, "invalid_expr_unary_operator", lex.KND_MINUS)
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
		v.model = get_num_model(v)
	}
	return v
}

func (u *unary) plus() value {
	v := u.p.eval.process(u.toks, u.model)
	if !types.IsPure(v.data.DataType) || !types.IsNumeric(v.data.DataType.Id) {
		u.p.eval.push_err_tok(u.token, "invalid_expr_unary_operator", lex.KND_PLUS)
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
		v.model = get_num_model(v)
	}
	return v
}

func (u *unary) caret() value {
	v := u.p.eval.process(u.toks, u.model)
	if !types.IsPure(v.data.DataType) || !types.IsInteger(v.data.DataType.Id) {
		u.p.eval.push_err_tok(u.token, "invalid_expr_unary_operator", lex.KND_CARET)
	}
	if v.constant {
		switch t := v.expr.(type) {
		case int64:
			v.expr = ^t
		case uint64:
			v.expr = ^t
		}
		v.model = get_num_model(v)
	}
	return v
}

func (u *unary) logical_not() value {
	v := u.p.eval.process(u.toks, u.model)
	if !is_bool_expr(v) {
		u.p.eval.push_err_tok(u.token, "invalid_expr_unary_operator", lex.KND_EXCL)
	} else if v.constant {
		v.expr = !v.expr.(bool)
		v.model = get_bool_model(v)
	}
	v.data.DataType.Id = types.BOOL
	v.data.DataType.Kind = lex.KND_BOOL
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
	case !types.IsExplicitPtr(v.data.DataType):
		u.p.eval.push_err_tok(u.token, "invalid_expr_unary_operator", lex.KND_STAR)
		goto end
	}
	v.data.DataType.Kind = v.data.DataType.Kind[1:]
end:
	v.data.Value = " "
	return v
}

func (u *unary) amper() value {
	v := u.p.eval.process(u.toks, u.model)
	v.constant = false
	v.lvalue = true
	v.mutable = true
	nodes := &u.model.nodes[u.model.index].nodes
	switch {
	case types.IsRef(v.data.DataType):
		model := exprNode{(*nodes)[1].String() + ".__alloc"}
		*nodes = nil
		*nodes = make([]ast.ExprModel, 1)
		(*nodes)[0] = model
		v.data.DataType.Kind = lex.KND_STAR + types.Elem(v.data.DataType).Kind
		return v
	case is_struct_ins(v):
		s := v.data.DataType.Tag.(*Struct)
		// Is not struct literal
		if s.Id != v.data.Value {
			break
		}
		if s.CppLinked {
			(*nodes)[0] = exprNode{"(new( std::nothrow ) "}
			last := &(*nodes)[len(*nodes)-1]
			*last = exprNode{(*last).String() + ")"}
		} else {
			var alloc_model exprNode
			alloc_model.value = "__julec_new_structure<"
			alloc_model.value += s.OutId()
			alloc_model.value += ">(new( std::nothrow ) "
			(*nodes)[0] = alloc_model
			last := &(*nodes)[len(*nodes)-1]
			*last = exprNode{(*last).String() + ")"}
		}
		v.data.DataType.Kind = lex.KND_AMPER + v.data.DataType.Kind
		return v
	case !can_get_ptr(v):
		u.p.eval.push_err_tok(u.token, "invalid_expr_unary_operator", lex.KND_AMPER)
		return v
	}
	v.data.DataType.Kind = lex.KND_STAR + v.data.DataType.Kind
	return v
}
