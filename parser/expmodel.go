package parser

import (
	"strings"

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
