package analysis

import (
	"strings"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

type expr_build_node struct {
	nodes []ast.ExprModel
}

type expr_model struct {
	index int
	nodes []expr_build_node
}

func new_expr_model(n int) *expr_model {
	m := new(expr_model)
	m.index = 0
	m.nodes = make([]expr_build_node, n)
	return m
}

func (m *expr_model) append_sub(node ast.ExprModel) {
	nodes := &m.nodes[m.index].nodes
	*nodes = append(*nodes, node)
}

func (m expr_model) String() string {
	var expr strings.Builder
	for _, node := range m.nodes {
		for _, node := range node.nodes {
			if node != nil {
				expr.WriteString(node.String())
			}
		}
	}
	return expr.String()
}

func (m *expr_model) Expr() Expr {
	return Expr{Model: m}
}

type exprNode struct {
	value string
}

func (node exprNode) String() string {
	return node.value
}

type sliceExpr struct {
	dataType Type
	expr     []ast.ExprModel
}

func (a sliceExpr) String() string {
	var cpp strings.Builder
	cpp.WriteString(a.dataType.String())
	cpp.WriteString("({")
	if len(a.expr) == 0 {
		cpp.WriteString("})")
		return cpp.String()
	}
	for _, exp := range a.expr {
		cpp.WriteString(exp.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + "})"
}

type mapExpr struct {
	dataType Type
	keyExprs []ast.ExprModel
	valExprs []ast.ExprModel
}

func (m mapExpr) String() string {
	var cpp strings.Builder
	cpp.WriteString(m.dataType.String())
	cpp.WriteByte('{')
	for i, k := range m.keyExprs {
		v := m.valExprs[i]
		cpp.WriteByte('{')
		cpp.WriteString(k.String())
		cpp.WriteByte(',')
		cpp.WriteString(v.String())
		cpp.WriteString("},")
	}
	cpp.WriteByte('}')
	return cpp.String()
}

type genericsExpr struct {
	exprs []Type
}

func (ge genericsExpr) String() string {
	if len(ge.exprs) == 0 {
		return ""
	}
	var cpp strings.Builder
	cpp.WriteByte('<')
	for _, g := range ge.exprs {
		cpp.WriteString(g.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ">"
}

type argsExpr struct {
	args []ast.Arg
}

func (a argsExpr) String() string {
	if len(a.args) == 0 {
		return ""
	}
	var cpp strings.Builder
	for _, arg := range a.args {
		cpp.WriteString(arg.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1]
}

type callExpr struct {
	f        *Fn
	generics genericsExpr
	args     argsExpr
}

func (ce callExpr) String() string {
	var cpp strings.Builder
	if !ast.HasAttribute(build.ATTR_CDEF, ce.f.Attributes) {
		cpp.WriteString(ce.generics.String())
	}
	cpp.WriteByte('(')
	cpp.WriteString(ce.args.String())
	cpp.WriteByte(')')
	return cpp.String()
}

type retExpr struct {
	vars   []*Var
	models []ast.ExprModel
}

func (re *retExpr) get_model(i int) string {
	if len(re.vars) > 0 {
		v := re.vars[i]
		if lex.IsIgnoreId(v.Id) {
			return re.models[i].String()
		}
		return v.OutId()
	}
	return re.models[i].String()
}

func (re *retExpr) required_return_expr() int {
	// vars always represents return expression count
	return len(re.vars)
}

func (re *retExpr) is_one_expr_for_multi_ret() bool {
	return re.required_return_expr() > 1 && len(re.models) == 1
}

func (re *retExpr) ready_ignored_var_to_decl(v *Var) {
	v.Id = "ret_var"
}

func (re *retExpr) setup_one_expr_to_multi_vars() string {
	var cpp strings.Builder
	for _, v := range re.vars {
		if lex.IsIgnoreId(v.Id) {
			// This assignment effects to original variable instance.
			re.ready_ignored_var_to_decl(v)
			cpp.WriteString(v.String())
			// To default
			v.Id = lex.IGNORE_ID
		}
	}
	cpp.WriteString("std::tie(")
	for _, v := range re.vars {
		if lex.IsIgnoreId(v.Id) {
			// This assignment effects to original variable instance.
			re.ready_ignored_var_to_decl(v)
			cpp.WriteString(v.OutId())
			// To default
			v.Id = lex.IGNORE_ID
		} else {
			cpp.WriteString(v.OutId())
		}
		cpp.WriteByte(',')
	}
	assign_expr := cpp.String()
	// Remove comma
	assign_expr = assign_expr[:cpp.Len()-1]
	assign_expr += ")"
	cpp.Reset()
	cpp.WriteByte('=')
	cpp.WriteString(re.models[0].String())
	cpp.WriteByte(';')
	return assign_expr + cpp.String()
}

func (re *retExpr) setup_plain_vars() string {
	var cpp strings.Builder
	for i, v := range re.vars {
		if lex.IsIgnoreId(v.Id) {
			continue
		}
		cpp.WriteString(v.OutId())
		cpp.WriteByte('=')
		cpp.WriteString(re.models[i].String())
		cpp.WriteByte(';')
	}
	return cpp.String()
}

func (re *retExpr) setup_vars() string {
	if re.is_one_expr_for_multi_ret() {
		return re.setup_one_expr_to_multi_vars()
	}
	return re.setup_plain_vars()
}

func (re *retExpr) multi_with_one_expr_str() string {
	var cpp strings.Builder
	cpp.WriteString("std::make_tuple(")
	for _, v := range re.vars {
		if lex.IsIgnoreId(v.Id) {
			// This assignment effects to original variable instance.
			re.ready_ignored_var_to_decl(v)
			cpp.WriteString(v.OutId())
			// To default
			v.Id = lex.IGNORE_ID
		} else {
			cpp.WriteString(v.OutId())
		}
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ")"
}

func (re *retExpr) multi_ret() string {
	var cpp strings.Builder
	cpp.WriteString("std::make_tuple(")
	for i := range re.models {
		cpp.WriteString(re.get_model(i))
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ")"
}

func (re *retExpr) single_ret() string {
	var cpp strings.Builder
	cpp.WriteString(re.get_model(0))
	return cpp.String()
}

func (re retExpr) String() string {
	var cpp strings.Builder
	if len(re.vars) > 0 {
		cpp.WriteString(re.setup_vars())
	}
	cpp.WriteString(" return ")
	switch {
	case re.is_one_expr_for_multi_ret():
		cpp.WriteString(re.multi_with_one_expr_str())
	case len(re.models) > 1:
		cpp.WriteString(re.multi_ret())
	default:
		cpp.WriteString(re.single_ret())
	}
	return cpp.String()
}

type serieExpr struct {
	exprs []ast.ExprModel
}

func (se serieExpr) String() string {
	var exprs strings.Builder
	for _, expr := range se.exprs {
		exprs.WriteString(expr.String())
	}
	return exprs.String()
}
