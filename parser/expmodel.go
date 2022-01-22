package parser

import (
	"strings"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
)

type expressionBuildNode struct {
	index int
	node  expressionModel
}

type expressionModelBuilder struct {
	index   int
	current expressionModel
	nodes   []expressionBuildNode
}

func newExpBuilder() *expressionModelBuilder {
	builder := new(expressionModelBuilder)
	builder.index = -1
	return builder
}

func (b *expressionModelBuilder) setIndex(index int) {
	if b.index != -1 {
		b.appendBuildNode(expressionBuildNode{
			index: b.index,
			node:  b.current,
		})
	}
	b.index = index
	b.current = expressionModel{}
}

func (b *expressionModelBuilder) appendBuildNode(node expressionBuildNode) {
	b.nodes = append(b.nodes, node)
}

func (b *expressionModelBuilder) appendNode(node expressionNode) {
	b.current.nodes = append(b.current.nodes, node)
}

func (b *expressionModelBuilder) build() (e expressionModel) {
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

type expressionNode interface {
	String() string
}

type expressionModel struct {
	nodes []expressionNode
}

func (model expressionModel) String() string {
	var sb strings.Builder
	for _, node := range model.nodes {
		sb.WriteString(node.String())
	}
	return sb.String()
}

func (model expressionModel) ExpressionAST() ast.ExpressionAST {
	return ast.ExpressionAST{Model: model}
}

type tokenExpNode struct {
	token lex.Token
}

func (node tokenExpNode) String() string {
	return node.token.Kind
}

type runeExpNode struct {
	token lex.Token
}

func (run runeExpNode) String() string {
	return "L" + run.token.Kind
}

type strExpNode struct {
	token lex.Token
}

func (str strExpNode) String() string {
	return "str(L" + str.token.Kind + ")"
}

type arrayPointerExp struct {
	nodes []expressionNode
}

func (ap arrayPointerExp) String() string {
	var cxxNodes strings.Builder
	for _, node := range ap.nodes {
		cxxNodes.WriteString(node.String())
	}
	return "new " + cxxNodes.String()[:cxxNodes.Len()-1] + ", true)"
}

type anonymousFunctionExp struct {
	ast ast.FunctionAST
}

func (af anonymousFunctionExp) String() string {
	var cxx strings.Builder
	cxx.WriteString("[&]")
	cxx.WriteString(paramsToCxx(af.ast.Params))
	ast.Indent++
	cxx.WriteString(ast.ParseBlock(af.ast.Block, ast.Indent))
	ast.Indent--
	return cxx.String()
}

type arrayExp struct {
	dataType    ast.DataTypeAST
	expressions []expressionModel
}

func (a arrayExp) String() string {
	var cxx strings.Builder
	cxx.WriteString(a.dataType.String())
	cxx.WriteString("({")
	if len(a.expressions) == 0 {
		return cxx.String() + "})"
	}
	for _, exp := range a.expressions {
		cxx.WriteString(exp.String())
		cxx.WriteString(", ")
	}
	return cxx.String()[:cxx.Len()-2] + "})"
}

type argsExp struct {
	args []ast.ArgAST
}

func (a argsExp) String() string {
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

type multiReturnExpModel struct {
	models []expressionModel
}

func (mre multiReturnExpModel) String() string {
	var cxx strings.Builder
	cxx.WriteString("std::make_tuple(")
	for _, model := range mre.models {
		cxx.WriteString(model.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ")"
}
