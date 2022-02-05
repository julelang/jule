package parser

import (
	"strings"
	"sync/atomic"

	"github.com/the-xlang/x/ast"
)

type function struct {
	Ast        ast.FunctionAST
	Attributes []ast.AttributeAST
}

func (f function) String() string {
	var cxx strings.Builder
	prototype := f.Prototype()
	cxx.WriteString(prototype[:len(prototype)-1])
	cxx.WriteByte(' ')
	atomic.SwapInt32(&ast.Indent, 0)
	cxx.WriteString(f.Ast.Block.String())
	return cxx.String()
}

// Prototype returns prototype cxx code of function.
func (f function) Prototype() string {
	var cxx strings.Builder
	cxx.WriteString(attributesToString(f.Attributes))
	cxx.WriteString(f.Ast.ReturnType.String())
	cxx.WriteByte(' ')
	cxx.WriteString(f.Ast.Name)
	cxx.WriteString(paramsToCxx(f.Ast.Params))
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
