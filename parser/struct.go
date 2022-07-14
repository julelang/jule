package parser

import (
	"strconv"
	"strings"

	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xtype"
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

func (s *xstruct) decldefString() string {
	var cxx strings.Builder
	cxx.WriteString(genericsToCxx(s.Ast.Generics))
	cxx.WriteByte('\n')
	cxx.WriteString("struct ")
	cxx.WriteString(xapi.OutId(s.Ast.Id, s.Ast.Tok.File))
	cxx.WriteString(" {\n")
	models.AddIndent()
	if len(s.Defs.Globals) > 0 {
		for _, g := range s.Defs.Globals {
			cxx.WriteString(models.IndentString())
			cxx.WriteString(g.FieldString())
			cxx.WriteByte('\n')
		}
		cxx.WriteString("\n\n")
	}
	for _, f := range s.Defs.Funcs {
		if f.used {
			cxx.WriteString(models.IndentString())
			cxx.WriteString(f.String())
			cxx.WriteString("\n\n")
		}
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
	cxx.WriteString(s.decldefString())
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

func (s *xstruct) selfVar(receiver DataType) *Var {
	v := new(models.Var)
	v.IdTok = s.Ast.Tok
	v.Type = receiver
	v.Type.Id = xtype.Struct
	v.Id = tokens.SELF
	if typeIsPtr(receiver) {
		v.Val.Model = exprNode{xapi.CxxSelf}
	} else {
		v.Val.Model = exprNode{tokens.STAR + xapi.CxxSelf}
	}
	return v
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
