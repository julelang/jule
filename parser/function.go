package parser

import (
	"strings"

	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xapi"
)

type function struct {
	Ast        Func
	Attributes []Attribute
	Desc       string
	used       bool
	checked    bool
}

func (f *function) outId() string {
	if f.Ast.Id == x.EntryPoint {
		return xapi.OutId(f.Ast.Id, nil)
	}
	return xapi.OutId(f.Ast.Id, f.Ast.Tok.File)
}

func (f function) String() string {
	var cxx strings.Builder
	cxx.WriteString(f.Head())
	cxx.WriteByte(' ')
	cxx.WriteString(f.Ast.Block.String())
	return cxx.String()
}

// Head returns declaration head of function.
func (f *function) Head() string {
	var cxx strings.Builder
	cxx.WriteString(attributesToString(f.Attributes))
	cxx.WriteString(f.Ast.RetType.String())
	cxx.WriteByte(' ')
	cxx.WriteString(f.outId())
	cxx.WriteString(paramsToCxx(f.Ast.Params))
	return cxx.String()
}

// Prototype returns prototype cxx code of function.
func (f *function) Prototype() string {
	var cxx strings.Builder
	cxx.WriteString(attributesToString(f.Attributes))
	cxx.WriteString(f.Ast.RetType.String())
	cxx.WriteByte(' ')
	cxx.WriteString(f.outId())
	cxx.WriteString(f.PrototypeParams())
	cxx.WriteByte(';')
	return cxx.String()
}

// PrototypeParams returns prototype cxx code of function parameters.
func (f *function) PrototypeParams() string {
	if len(f.Ast.Params) == 0 {
		return "(void)"
	}
	var cxx strings.Builder
	cxx.WriteByte('(')
	for _, p := range f.Ast.Params {
		cxx.WriteString(p.Prototype())
		cxx.WriteString(", ")
	}
	return cxx.String()[:cxx.Len()-2] + ")"
}

func attributesToString(attributes []Attribute) string {
	var cxx strings.Builder
	for _, attr := range attributes {
		cxx.WriteString(attr.String())
		cxx.WriteByte(' ')
	}
	return cxx.String()
}

func paramsToCxx(params []Param) string {
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
