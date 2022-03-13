package parser

import (
	"strings"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/xapi"
)

type IExprNode interface {
	String() string
}

type exprBuildNode struct {
	index int
	nodes []IExprNode
}

type exprBuilder struct {
	index int
	nodes []exprBuildNode
}

func newExprBuilder(processes [][]lex.Token) *exprBuilder {
	b := new(exprBuilder)
	b.index = 0
	for i := range processes {
		b.nodes = append(b.nodes, exprBuildNode{index: i})
	}
	return b
}

func (b *exprBuilder) appendNodeToSubNodes(node IExprNode) {
	nodes := &b.nodes[b.index].nodes
	*nodes = append(*nodes, node)
}

func (b *exprBuilder) build() (e exprModel) {
	for _, node := range b.nodes {
		e.nodes = append(e.nodes, node.nodes...)
	}
	return
}

type exprModel struct {
	nodes []IExprNode
}

func (model exprModel) String() string {
	var expr strings.Builder
	for _, node := range model.nodes {
		expr.WriteString(node.String())
	}
	return expr.String()
}

func (model exprModel) ExprAST() ast.Expr {
	return ast.Expr{Model: model}
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
	expr     []exprModel
}

func (a arrayExpr) String() string {
	var cxx strings.Builder
	cxx.WriteString(a.dataType.String())
	cxx.WriteString("({")
	if len(a.expr) == 0 {
		return cxx.String() + "})"
	}
	for _, exp := range a.expr {
		cxx.WriteString(exp.String())
		cxx.WriteString(", ")
	}
	return cxx.String()[:cxx.Len()-2] + "})"
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
	models []exprModel
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
