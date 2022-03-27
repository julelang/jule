package parser

import (
	"strings"

	"github.com/the-xlang/xxc/ast"
	"github.com/the-xlang/xxc/lex"
	"github.com/the-xlang/xxc/pkg/xapi"
)

type iExpr interface {
	String() string
}

type exprBuildNode struct {
	index int
	nodes []iExpr
}

type exprModel struct {
	index int
	nodes []exprBuildNode
}

func newExprModel(processes [][]lex.Tok) *exprModel {
	m := new(exprModel)
	m.index = 0
	for i := range processes {
		m.nodes = append(m.nodes, exprBuildNode{index: i})
	}
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

func (m *exprModel) Expr() ast.Expr {
	return ast.Expr{Model: m}
}

type exprNode struct {
	value string
}

func (node exprNode) String() string {
	return node.value
}

type anonFunc struct {
	ast ast.Func
}

func (af anonFunc) String() string {
	var cxx strings.Builder
	cxx.WriteString("[=]")
	cxx.WriteString(paramsToCxx(af.ast.Params))
	cxx.WriteString(" mutable -> ")
	cxx.WriteString(af.ast.RetType.String())
	cxx.WriteByte(' ')
	cxx.WriteString(af.ast.Block.String())
	return cxx.String()
}

type arrayExpr struct {
	dataType ast.DataType
	expr     []iExpr
}

func (a arrayExpr) String() string {
	var cxx strings.Builder
	cxx.WriteByte('{')
	if len(a.expr) == 0 {
		cxx.WriteByte('}')
		return cxx.String()
	}
	for _, exp := range a.expr {
		cxx.WriteString(exp.String())
		cxx.WriteString(", ")
	}
	return cxx.String()[:cxx.Len()-2] + "}"
}

type mapExpr struct {
	dataType ast.DataType
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
	args []ast.Arg
}

func (a argsExpr) String() string {
	if a.args == nil {
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

type newHeapAllocExpr struct {
	typeAST ast.DataType
}

func (nha newHeapAllocExpr) String() string {
	return xapi.ToXAlloc(nha.typeAST.String())
}

type assignExpr struct {
	assign ast.Assign
}

func (a assignExpr) String() string {
	var cxx strings.Builder
	cxx.WriteByte('(')
	cxx.WriteString(a.assign.String())
	cxx.WriteByte(')')
	return cxx.String()
}
