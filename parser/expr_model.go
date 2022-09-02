package parser

import (
	"fmt"
	"strings"

	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/pkg/jule"
	"github.com/jule-lang/jule/pkg/juleapi"
)

type iExpr interface {
	String() string
}

type exprBuildNode struct {
	nodes []iExpr
}

type exprModel struct {
	index int
	nodes []exprBuildNode
}

func newExprModel(processes [][]lex.Token) *exprModel {
	m := new(exprModel)
	m.index = 0
	m.nodes = make([]exprBuildNode, len(processes))
	return m
}

func (m *exprModel) appendSubNode(node iExpr) {
	nodes := &m.nodes[m.index].nodes
	*nodes = append(*nodes, node)
}

func (m exprModel) String() string {
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

func (m *exprModel) Expr() Expr {
	return Expr{Model: m}
}

type exprNode struct {
	value string
}

func (node exprNode) String() string {
	return node.value
}

type anonFuncExpr struct {
	ast  *Func
}

func (af anonFuncExpr) String() string {
	var cpp strings.Builder
	t := Type{
		Token:  af.ast.Token,
		Kind: af.ast.DataTypeString(),
		Tag:  af.ast,
	}
	cpp.WriteString(t.FuncString())
	cpp.WriteString("([=]")
	cpp.WriteString(paramsToCpp(af.ast.Params))
	cpp.WriteString(" mutable -> ")
	cpp.WriteString(af.ast.RetType.String())
	cpp.WriteByte(' ')
	cpp.WriteString(af.ast.Block.String())
	cpp.WriteByte(')')
	return cpp.String()
}

type sliceExpr struct {
	dataType Type
	expr     []iExpr
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
	keyExprs []iExpr
	valExprs []iExpr
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
	types []Type
}

func (ge genericsExpr) String() string {
	if len(ge.types) == 0 {
		return ""
	}
	var cpp strings.Builder
	cpp.WriteByte('<')
	for _, generic := range ge.types {
		cpp.WriteString(generic.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ">"
}

type argsExpr struct {
	args []models.Arg
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
	f        *Func
	generics genericsExpr
	args     argsExpr
}

func (ce callExpr) String() string {
	var cpp strings.Builder
	if ce.f.FindAttribute(jule.Attribute_CDef) == nil {
		cpp.WriteString(ce.generics.String())
	}
	cpp.WriteByte('(')
	cpp.WriteString(ce.args.String())
	cpp.WriteByte(')')
	return cpp.String()
}

type retExpr struct {
	vars   []*Var
	models []iExpr
}

func (re *retExpr) get_model(i int) string {
	if len(re.vars) > 0 {
		v := re.vars[i]
		if juleapi.IsIgnoreId(v.Id) {
			return re.models[i].String()
		}
		return v.OutId()
	}
	return re.models[0].String()
}

func (re *retExpr) setup_vars() string {
	var cpp strings.Builder
	for i, v := range re.vars {
		if juleapi.IsIgnoreId(v.Id) {
			continue
		}
		cpp.WriteString(v.OutId())
		cpp.WriteByte('=')
		cpp.WriteString(re.models[i].String())
		cpp.WriteByte(';')
	}
	return cpp.String()
}

func (re *retExpr) multiRetString() string {
	var cpp strings.Builder
	cpp.WriteString("std::make_tuple(")
	for i := range re.models {
		cpp.WriteString(re.get_model(i))
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ")"
}

func (re *retExpr) singleRetString() string {
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
	if len(re.models) > 1 {
		cpp.WriteString(re.multiRetString())
	} else {
		cpp.WriteString(re.singleRetString())
	}
	return cpp.String()
}

type serieExpr struct {
	exprs []any
}

func (se serieExpr) String() string {
	var exprs strings.Builder
	for _, expr := range se.exprs {
		exprs.WriteString(fmt.Sprint(expr))
	}
	return exprs.String()
}
