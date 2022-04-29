package parser

import (
	"strings"

	"github.com/the-xlang/xxc/ast"
	"github.com/the-xlang/xxc/pkg/xapi"
)

type xstruct struct {
	Ast         Struct
	Defs        *Defmap
	Used        bool
	Desc        string
	constructor *Func
}

func (s *xstruct) declString() string {
	var cxx strings.Builder
	cxx.WriteString("struct ")
	cxx.WriteString(xapi.AsId(s.Ast.Id))
	cxx.WriteString(" {\n")
	ast.AddIndent()
	for _, g := range s.Defs.Globals {
		cxx.WriteString(ast.IndentString())
		cxx.WriteString(g.FieldString())
		cxx.WriteByte('\n')
	}
	ast.DoneIndent()
	cxx.WriteString(ast.IndentString())
	cxx.WriteString("};")
	return cxx.String()
}

func (s *xstruct) ostream() string {
	var cxx strings.Builder
	cxx.WriteString("std::ostream &operator<<(std::ostream &_Stream, const ")
	cxx.WriteString(xapi.AsId(s.Ast.Id))
	cxx.WriteString(" &_Src) {\n")
	ast.AddIndent()
	cxx.WriteString(ast.IndentString())
	cxx.WriteString(`_Stream << "`)
	cxx.WriteString(s.Ast.Id)
	cxx.WriteString("{\";\n")
	for i, field := range s.Ast.Fields {
		cxx.WriteString(ast.IndentString())
		cxx.WriteString(`_Stream << "`)
		cxx.WriteString(field.Id)
		cxx.WriteString(`: " << _Src.`)
		cxx.WriteString(xapi.AsId(field.Id))
		if i+1 < len(s.Ast.Fields) {
			cxx.WriteString(" << \", \"")
		}
		cxx.WriteString(";\n")
	}
	cxx.WriteString(ast.IndentString())
	cxx.WriteString("_Stream << \"}\";\n")
	cxx.WriteString(ast.IndentString())
	cxx.WriteString("return _Stream;\n")
	ast.DoneIndent()
	cxx.WriteString(ast.IndentString())
	cxx.WriteString("}")
	return cxx.String()
}

func (s xstruct) String() string {
	var cxx strings.Builder
	cxx.WriteString(s.declString())
	cxx.WriteString("\n\n")
	cxx.WriteString(ast.IndentString())
	cxx.WriteString(s.ostream())
	return cxx.String()
}
