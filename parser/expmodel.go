package parser

import (
	"strings"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
)

type IExprNode interface {
	String() string
}

type exprBuildNode struct {
	index int
	node  exprModel
}

type exprBuilder struct {
	index   int
	current exprModel
	nodes   []exprBuildNode
}

func newExprBuilder() *exprBuilder {
	builder := new(exprBuilder)
	builder.index = -1
	return builder
}

func (b *exprBuilder) setIndex(index int) {
	if b.index != -1 {
		b.appendBuildNode(exprBuildNode{b.index, b.current})
	}
	b.index = index
	b.current = exprModel{}
}

func (b *exprBuilder) appendBuildNode(node exprBuildNode) {
	b.nodes = append(b.nodes, node)
}

func (b *exprBuilder) appendNode(node IExprNode) {
	b.current.nodes = append(b.current.nodes, node)
}

func (b *exprBuilder) build() (e exprModel) {
	b.setIndex(-1)
	for index := range b.nodes {
		for _, buildNode := range b.nodes {
			if buildNode.index != index {
				continue
			}
			e.nodes = append(e.nodes, buildNode.node)
		}
	}
	return
}

type exprModel struct {
	nodes []IExprNode
}

func (model exprModel) String() string {
	var sb strings.Builder
	for _, node := range model.nodes {
		sb.WriteString(node.String())
	}
	return sb.String()
}

func (model exprModel) ExprAST() ast.ExprAST {
	return ast.ExprAST{Model: model}
}

type tokenExprNode struct {
	token lex.Token
}

func (node tokenExprNode) String() string {
	return node.token.Kind
}

type runeExprNode struct {
	token lex.Token
}

func (run runeExprNode) String() string {
	return "L" + run.token.Kind
}

type strExprNode struct {
	token lex.Token
}

func (str strExprNode) String() string {
	return "str(L" + str.token.Kind + ")"
}

type anonymousFunctionExpr struct {
	ast ast.FunctionAST
}

func (af anonymousFunctionExpr) String() string {
	var cxx strings.Builder
	cxx.WriteString("[=]")
	cxx.WriteString(paramsToCxx(af.ast.Params))
	cxx.WriteString(" mutable -> ")
	cxx.WriteString(af.ast.ReturnType.String())
	cxx.WriteByte(' ')
	cxx.WriteString(af.ast.Block.String())
	return cxx.String()
}

type arrayExpr struct {
	dataType ast.DataTypeAST
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
	args []ast.ArgAST
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

type multiReturnExprModel struct {
	models []exprModel
}

func (mre multiReturnExprModel) String() string {
	var cxx strings.Builder
	cxx.WriteString("std::make_tuple(")
	for _, model := range mre.models {
		cxx.WriteString(model.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ")"
}

type newHeapAllocationExprModel struct {
	typeAST ast.DataTypeAST
}

func (nha newHeapAllocationExprModel) String() string {
	return "new(std::nothrow) " + nha.typeAST.String()
}
