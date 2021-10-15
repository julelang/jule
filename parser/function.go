package parser

import (
	"strings"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
)

type function struct {
	ast        ast.FunctionAST
	token      lex.Token
	name       string
	returnType ast.DataTypeAST
	params     []ast.ParameterAST
	attributes []ast.AttributeAST
	block      ast.BlockAST
}

func (f function) String() string {
	var cxx strings.Builder
	prototype := f.Prototype()
	cxx.WriteString(prototype[:len(prototype)-1])
	cxx.WriteString(" {")
	cxx.WriteString(f.block.String())
	cxx.WriteString("\n}")
	return cxx.String()
}

// Prototype returns prototype cxx code of function.
func (f function) Prototype() string {
	var cxx strings.Builder
	cxx.WriteString(attributesToString(f.attributes))
	cxx.WriteString(f.returnType.String())
	cxx.WriteByte(' ')
	cxx.WriteString(f.name)
	cxx.WriteByte('(')
	cxx.WriteString(paramsToCxx(f.params))
	cxx.WriteString(");")
	return cxx.String()
}

func attributesToString(attributes []ast.AttributeAST) string {
	var cxx strings.Builder
	for _, attribute := range attributes {
		cxx.WriteString(attribute.String())
		cxx.WriteByte(' ')
	}
	return cxx.String()
}

func paramsToCxx(params []ast.ParameterAST) string {
	if len(params) == 0 {
		return ""
	}
	var cxx strings.Builder
	for _, p := range params {
		cxx.WriteString(p.String())
		cxx.WriteString(", ")
	}
	return cxx.String()[:cxx.Len()-2]
}
