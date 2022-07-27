package parser

import (
	"fmt"
	"strings"

	"github.com/the-xlang/xxc/ast/models"
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

func newExprModel(processes []Toks) *exprModel {
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
			expr.WriteString(node.String())
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
	ast     Func
	capture byte
}

func (af anonFuncExpr) String() string {
	var cxx strings.Builder
	cxx.WriteByte('[')
	cxx.WriteByte(af.capture)
	cxx.WriteByte(']')
	cxx.WriteString(paramsToCxx(af.ast.Params))
	cxx.WriteString(" mutable -> ")
	cxx.WriteString(af.ast.RetType.String())
	cxx.WriteByte(' ')
	cxx.WriteString(af.ast.Block.String())
	return cxx.String()
}

type sliceExpr struct {
	dataType DataType
	expr     []iExpr
}

func (a sliceExpr) String() string {
	var cxx strings.Builder
	cxx.WriteString(a.dataType.String())
	cxx.WriteString("({")
	if len(a.expr) == 0 {
		cxx.WriteString("})")
		return cxx.String()
	}
	for _, exp := range a.expr {
		cxx.WriteString(exp.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + "})"
}

type mapExpr struct {
	dataType DataType
	keyExprs []iExpr
	valExprs []iExpr
}

func (m mapExpr) String() string {
	var cxx strings.Builder
	cxx.WriteString(m.dataType.String())
	cxx.WriteByte('{')
	for i, k := range m.keyExprs {
		v := m.valExprs[i]
		cxx.WriteByte('{')
		cxx.WriteString(k.String())
		cxx.WriteByte(',')
		cxx.WriteString(v.String())
		cxx.WriteString("},")
	}
	cxx.WriteByte('}')
	return cxx.String()
}

type argsExpr struct {
	args []models.Arg
}

func (a argsExpr) String() string {
	if len(a.args) == 0 {
		return ""
	}
	var cxx strings.Builder
	for _, arg := range a.args {
		cxx.WriteString(arg.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1]
}

type multiRetExpr struct {
	models []iExpr
}

func (mre multiRetExpr) String() string {
	var cxx strings.Builder
	cxx.WriteString("std::make_tuple(")
	for _, model := range mre.models {
		cxx.WriteString(model.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ")"
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
