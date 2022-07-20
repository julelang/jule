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

func (s *xstruct) cxxGenerics() (def string, serie string) {
	if len(s.Ast.Generics) == 0 {
		return "", ""
	}
	var cxxDef strings.Builder
	cxxDef.WriteString("template<typename ")
	var cxxSerie strings.Builder
	cxxSerie.WriteByte('<')
	for i := range s.Ast.Generics {
		cxxSerie.WriteByte('T')
		cxxSerie.WriteString(strconv.Itoa(i))
		cxxSerie.WriteByte(',')
	}
	serie = cxxSerie.String()[:cxxSerie.Len()-1] + ">"
	cxxDef.WriteString(serie[1:])
	cxxDef.WriteByte('\n')
	return cxxDef.String(), serie
}

func (s *xstruct) outId() string {
	return xapi.OutId(s.Ast.Id, s.Ast.Tok.File)
}

func (s *xstruct) operators() string {
	outid := s.outId()
	genericsDef, genericsSerie := s.cxxGenerics()
	var cxx strings.Builder
	cxx.WriteString(models.IndentString())
	if l, _ := cxx.WriteString(genericsDef); l > 0 {
		cxx.WriteString(models.IndentString())
	}
	cxx.WriteString("inline bool operator==(const ")
	cxx.WriteString(outid)
	cxx.WriteString(genericsSerie)
	cxx.WriteString(" &_Src) {")
	if len(s.Defs.Globals) > 0 {
		models.AddIndent()
		cxx.WriteByte('\n')
		cxx.WriteString(models.IndentString())
		var expr strings.Builder
		expr.WriteString("return ")
		models.AddIndent()
		for _, g := range s.Defs.Globals {
			expr.WriteByte('\n')
			expr.WriteString(models.IndentString())
			expr.WriteString("this->")
			gid := g.OutId()
			expr.WriteString(gid)
			expr.WriteString(" == _Src.")
			expr.WriteString(gid)
			expr.WriteString(" &&")
		}
		models.DoneIndent()
		cxx.WriteString(expr.String()[:expr.Len()-3])
		cxx.WriteString(";\n")
		models.DoneIndent()
		cxx.WriteString(models.IndentString())
		cxx.WriteByte('}')
	} else {
		cxx.WriteString(" return true; }")
	}
	cxx.WriteString("\n\n")
	cxx.WriteString(models.IndentString())
	if l, _ := cxx.WriteString(genericsDef); l > 0 {
		cxx.WriteString(models.IndentString())
	}
	cxx.WriteString("inline bool operator!=(const ")
	cxx.WriteString(outid)
	cxx.WriteString(genericsSerie)
	cxx.WriteString(" &_Src) { return !this->operator==(_Src); }")
	return cxx.String()
}

func (s *xstruct) cxxConstructor() string {
	var cxx strings.Builder
	cxx.WriteString(models.IndentString())
	cxx.WriteString(s.outId())
	cxx.WriteString(paramsToCxx(s.constructor.Params))
	cxx.WriteString(" noexcept {")
	if len(s.Defs.Globals) > 0 {
		models.AddIndent()
		for i, g := range s.Defs.Globals {
			cxx.WriteByte('\n')
			cxx.WriteString(models.IndentString())
			cxx.WriteString(g.OutId())
			cxx.WriteString(" = ")
			cxx.WriteString(s.constructor.Params[i].OutId())
			cxx.WriteByte(';')
		}
		models.DoneIndent()
		cxx.WriteByte('\n')
	}
	cxx.WriteString(models.IndentString())
	cxx.WriteByte('}')
	return cxx.String()
}

func (s *xstruct) decldefString() string {
	var cxx strings.Builder
	cxx.WriteString(genericsToCxx(s.Ast.Generics))
	cxx.WriteByte('\n')
	cxx.WriteString("struct ")
	cxx.WriteString(s.outId())
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
	cxx.WriteString(s.cxxConstructor())
	cxx.WriteString("\n\n")
	for _, f := range s.Defs.Funcs {
		if f.used {
			cxx.WriteString(models.IndentString())
			cxx.WriteString(f.String())
			cxx.WriteString("\n\n")
		}
	}
	cxx.WriteString(s.operators())
	cxx.WriteByte('\n')
	models.DoneIndent()
	cxx.WriteString(models.IndentString())
	cxx.WriteString("};")
	return cxx.String()
}

func (s *xstruct) ostream() string {
	var cxx strings.Builder
	genericsDef, genericsSerie := s.cxxGenerics()
	cxx.WriteString(models.IndentString())
	if l, _ := cxx.WriteString(genericsDef); l > 0 {
		cxx.WriteString(models.IndentString())
	}
	cxx.WriteString("std::ostream &operator<<(std::ostream &_Stream, const ")
	cxx.WriteString(s.outId())
	cxx.WriteString(genericsSerie)
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
		cxx.WriteString(field.OutId())
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
	cxx.WriteString(s.ostream())
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
		v.Expr.Model = exprNode{xapi.CxxSelf}
	} else {
		v.Expr.Model = exprNode{tokens.STAR + xapi.CxxSelf}
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
