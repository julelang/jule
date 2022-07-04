package parser

import (
	"strconv"
	"strings"

	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/pkg/xapi"
)

type xstruct struct {
	Ast         Struct
	Defs        *Defmap
	Used        bool
	Desc        string
	constructor *Func
	// Instance generics.
	generics []DataType
}

func (s *xstruct) declString() string {
	var cxx strings.Builder
	cxx.WriteString(genericsToCxx(s.Ast.Generics))
	cxx.WriteByte('\n')
	cxx.WriteString("struct ")
	cxx.WriteString(xapi.OutId(s.Ast.Id, s.Ast.Tok.File))
	cxx.WriteString(" {\n")
	models.AddIndent()
	for _, g := range s.Defs.Globals {
		cxx.WriteString(models.IndentString())
		cxx.WriteString(g.FieldString())
		cxx.WriteByte('\n')
	}
	models.DoneIndent()
	cxx.WriteString(models.IndentString())
	cxx.WriteString("};")
	return cxx.String()
}

func (s *xstruct) ostreams() string {
	var cxx strings.Builder
	var generics string
	if len(s.Ast.Generics) > 0 {
		var gb strings.Builder
		gb.WriteByte('<')
		for i := range s.Ast.Generics {
			gb.WriteByte('T')
			gb.WriteString(strconv.Itoa(i))
			gb.WriteByte(',')
		}
		generics = gb.String()[:gb.Len()-1] + ">"
		cxx.WriteString("template<typename ")
		// Starts 1 for skip "<"
		cxx.WriteString(generics[1:])
		cxx.WriteByte('\n')
	}
	cxx.WriteString("std::ostream &operator<<(std::ostream &_Stream, const ")
	cxx.WriteString(xapi.OutId(s.Ast.Id, s.Ast.Tok.File))
	cxx.WriteString(generics)
	cxx.WriteString(" &_Src) {\n")
	models.AddIndent()
	cxx.WriteString(models.IndentString())
	cxx.WriteString(`_Stream << "`)
	cxx.WriteString(s.Ast.Id)
	cxx.WriteString("{\";\n")
	for i, field := range s.Ast.Fields {
		cxx.WriteString(models.IndentString())
		cxx.WriteString(`_Stream << "`)
		cxx.WriteString(field.Id)
		cxx.WriteString(`:" << _Src.`)
		cxx.WriteString(xapi.OutId(field.Id, s.Ast.Tok.File))
		if i+1 < len(s.Ast.Fields) {
			cxx.WriteString(" << \", \"")
		}
		cxx.WriteString(";\n")
	}
	cxx.WriteString(models.IndentString())
	cxx.WriteString("_Stream << \"}\";\n")
	cxx.WriteString(models.IndentString())
	cxx.WriteString("return _Stream;\n")
	models.DoneIndent()
	cxx.WriteString(models.IndentString())
	cxx.WriteString("}")
	return cxx.String()
}

func (s xstruct) String() string {
	var cxx strings.Builder
	cxx.WriteString(s.declString())
	cxx.WriteString("\n\n")
	cxx.WriteString(models.IndentString())
	cxx.WriteString(s.ostreams())
	return cxx.String()
}

// Generics returns generics of type.
//
// This function is should be have this function
// for Genericable interface of ast package.
func (s *xstruct) Generics() []DataType {
	return s.generics
}

// SetGenerics set generics of type.
//
// This function is should be have this function
// for Genericable interface of ast package.
func (s *xstruct) SetGenerics(generics []DataType) {
	s.generics = generics
}

func (s *xstruct) dataTypeString() string {
	var dts strings.Builder
	dts.WriteString(s.Ast.Id)
	if len(s.Ast.Generics) > 0 {
		dts.WriteByte('[')
		var gs strings.Builder
		// Instance
		if len(s.generics) > 0 {
			for _, generic := range s.generics {
				gs.WriteString(generic.String())
				gs.WriteByte(',')
			}
		} else {
			for _, generic := range s.Ast.Generics {
				gs.WriteString(generic.Id)
				gs.WriteByte(',')
			}
		}
		dts.WriteString(gs.String()[:gs.Len()-1])
		dts.WriteByte(']')
	}
	return dts.String()
}
