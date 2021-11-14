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

func (builder *expressionModelBuilder) setIndex(index int) {
	if builder.index != -1 {
		builder.appendBuildNode(expressionBuildNode{
			index: builder.index,
			node:  builder.current,
		})
	}
	builder.index = index
	builder.current = expressionModel{}
}

func (builder *expressionModelBuilder) appendBuildNode(node expressionBuildNode) {
	builder.nodes = append(builder.nodes, node)
}

func (builder *expressionModelBuilder) appendNode(node expressionNode) {
	builder.current.nodes = append(builder.current.nodes, node)
}

func (builder *expressionModelBuilder) build() (e expressionModel) {
	builder.setIndex(-1)
	for index := range builder.nodes {
		for _, buildNode := range builder.nodes {
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
	return node.token.Value
}

type runeExpNode struct {
	token lex.Token
}

func (run runeExpNode) String() string {
	return "L" + run.token.Value
}

type strExpNode struct {
	token lex.Token
}

func (str strExpNode) String() string {
	return "L" + str.token.Value
}

type functionPointerExp struct {
	valueDataType ast.DataTypeAST
	nodes         []expressionNode
}

func (fb functionPointerExp) String() string {
	var cxxNodes strings.Builder
	for _, node := range fb.nodes {
		cxxNodes.WriteString(node.String())
	}
	return "new " + fb.valueDataType.String() + "(" + cxxNodes.String() + ")"
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
	if len(a.expressions) == 0 {
		return "({})"
	}
	var cxx strings.Builder
	cxx.WriteString(a.dataType.String())
	cxx.WriteString("({")
	for _, exp := range a.expressions {
		cxx.WriteString(exp.String())
		cxx.WriteString(", ")
	}
	return cxx.String()[:cxx.Len()-2] + "})"
}
