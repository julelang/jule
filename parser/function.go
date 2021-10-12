package parser

import (
	"strings"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
)

type function struct {
	Token      lex.Token
	Name       string
	ReturnType ast.DataTypeAST
	Params     []ast.ParameterAST
	Attributes []ast.AttributeAST
	Block      ast.BlockAST
}

func (f function) String() string {
	var cxx strings.Builder
	prototype := f.Prototype()
	cxx.WriteString(prototype[:len(prototype)-1])
	cxx.WriteString(" {")
	cxx.WriteString(f.Block.String())
	cxx.WriteString("\n}")
	return cxx.String()
}

// Prototype returns prototype cxx code of function.
func (f function) Prototype() string {
	var cxx strings.Builder
	cxx.WriteString(attributesToString(f.Attributes))
	cxx.WriteString(f.ReturnType.String())
	cxx.WriteByte(' ')
	cxx.WriteString(f.Name)
	cxx.WriteByte('(')
	cxx.WriteString(paramsToCxx(f.Params))
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
