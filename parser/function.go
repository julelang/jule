package parser

import (
	"strings"

	"github.com/the-xlang/x/ast"
)

type function struct {
	ast        ast.FunctionAST
	attributes []ast.AttributeAST
}

func (f function) String() string {
	var cxx strings.Builder
	prototype := f.Prototype()
	cxx.WriteString(prototype[:len(prototype)-1])
	cxx.WriteString(f.ast.Block.String())
	return cxx.String()
}

// Prototype returns prototype cxx code of function.
func (f function) Prototype() string {
	var cxx strings.Builder
	cxx.WriteString(attributesToString(f.attributes))
	cxx.WriteString(f.ast.ReturnType.String())
	cxx.WriteByte(' ')
	cxx.WriteString(f.ast.Name)
	cxx.WriteString(paramsToCxx(f.ast.Params))
	cxx.WriteByte(';')
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
		return "(void)"
	}
	var cxx strings.Builder
	cxx.WriteByte('(')
	for _, p := range params {
		cxx.WriteString(p.String())
		cxx.WriteString(", ")
	}
	return cxx.String()[:cxx.Len()-2] + ")"
}
