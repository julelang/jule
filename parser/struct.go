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
	traits      []*trait
	// Instance generics.
	generics []DataType
}

func (s *xstruct) hasTrait(t *trait) bool {
	for _, st := range s.traits {
		if t == st {
			return true
		}
	}
	return false
}

func (s *xstruct) cppGenerics() (def string, serie string) {
	if len(s.Ast.Generics) == 0 {
		return "", ""
	}
	var cppDef strings.Builder
	cppDef.WriteString("template<typename ")
	var cppSerie strings.Builder
	cppSerie.WriteByte('<')
	for i := range s.Ast.Generics {
		cppSerie.WriteByte('T')
		cppSerie.WriteString(strconv.Itoa(i))
		cppSerie.WriteByte(',')
	}
	serie = cppSerie.String()[:cppSerie.Len()-1] + ">"
	cppDef.WriteString(serie[1:])
	cppDef.WriteByte('\n')
	return cppDef.String(), serie
}

// OutId returns xapi.OutId of struct.
//
// This function is should be have this function
// for CompiledStruct interface of ast package.
func (s *xstruct) OutId() string {
	return xapi.OutId(s.Ast.Id, s.Ast.Tok.File)
}

func (s *xstruct) operators() string {
	outid := s.OutId()
	genericsDef, genericsSerie := s.cppGenerics()
	var cpp strings.Builder
	cpp.WriteString(models.IndentString())
	if l, _ := cpp.WriteString(genericsDef); l > 0 {
		cpp.WriteString(models.IndentString())
	}
	cpp.WriteString("inline bool operator==(const ")
	cpp.WriteString(outid)
	cpp.WriteString(genericsSerie)
	cpp.WriteString(" &_Src) {")
	if len(s.Defs.Globals) > 0 {
		models.AddIndent()
		cpp.WriteByte('\n')
		cpp.WriteString(models.IndentString())
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
		cpp.WriteString(expr.String()[:expr.Len()-3])
		cpp.WriteString(";\n")
		models.DoneIndent()
		cpp.WriteString(models.IndentString())
		cpp.WriteByte('}')
	} else {
		cpp.WriteString(" return true; }")
	}
	cpp.WriteString("\n\n")
	cpp.WriteString(models.IndentString())
	if l, _ := cpp.WriteString(genericsDef); l > 0 {
		cpp.WriteString(models.IndentString())
	}
	cpp.WriteString("inline bool operator!=(const ")
	cpp.WriteString(outid)
	cpp.WriteString(genericsSerie)
	cpp.WriteString(" &_Src) { return !this->operator==(_Src); }")
	return cpp.String()
}

func (s *xstruct) cppConstructor() string {
	var cpp strings.Builder
	cpp.WriteString(models.IndentString())
	cpp.WriteString(s.OutId())
	cpp.WriteString(paramsToCpp(s.constructor.Params))
	cpp.WriteString(" noexcept {")
	if len(s.Defs.Globals) > 0 {
		models.AddIndent()
		for i, g := range s.Defs.Globals {
			cpp.WriteByte('\n')
			cpp.WriteString(models.IndentString())
			cpp.WriteString("this->")
			cpp.WriteString(g.OutId())
			cpp.WriteString(" = ")
			cpp.WriteString(s.constructor.Params[i].OutId())
			cpp.WriteByte(';')
		}
		models.DoneIndent()
		cpp.WriteByte('\n')
	}
	cpp.WriteString(models.IndentString())
	cpp.WriteByte('}')
	return cpp.String()
}

func (s *xstruct) cppTraits() string {
	if len(s.traits) == 0 {
		return ""
	}
	var cpp strings.Builder
	cpp.WriteString(": ")
	for _, t := range s.traits {
		cpp.WriteString("public ")
		cpp.WriteString(t.OutId())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1]
}

func (s *xstruct) prototype() string {
	var cpp strings.Builder
	cpp.WriteString(genericsToCpp(s.Ast.Generics))
	cpp.WriteString(" struct ")
	cpp.WriteString(s.OutId())
	cpp.WriteByte(';')
	return cpp.String()
}

func (s *xstruct) decldefString() string {
	var cpp strings.Builder
	cpp.WriteString(genericsToCpp(s.Ast.Generics))
	cpp.WriteByte('\n')
	cpp.WriteString("struct ")
	cpp.WriteString(s.OutId())
	cpp.WriteString(s.cppTraits())
	cpp.WriteString(" {\n")
	models.AddIndent()
	if len(s.Defs.Globals) > 0 {
		for _, g := range s.Defs.Globals {
			cpp.WriteString(models.IndentString())
			cpp.WriteString(g.FieldString())
			cpp.WriteByte('\n')
		}
		cpp.WriteString("\n\n")
		cpp.WriteString(s.cppConstructor())
		cpp.WriteString("\n\n")
	}
	cpp.WriteString(models.IndentString())
	cpp.WriteString(s.OutId())
	cpp.WriteString("(void) noexcept {}\n\n")
	for _, f := range s.Defs.Funcs {
		if f.used {
			cpp.WriteString(models.IndentString())
			cpp.WriteString(f.String())
			cpp.WriteString("\n\n")
		}
	}
	cpp.WriteString(s.operators())
	cpp.WriteByte('\n')
	models.DoneIndent()
	cpp.WriteString(models.IndentString())
	cpp.WriteString("};")
	return cpp.String()
}

func (s *xstruct) ostream() string {
	var cpp strings.Builder
	genericsDef, genericsSerie := s.cppGenerics()
	cpp.WriteString(models.IndentString())
	if l, _ := cpp.WriteString(genericsDef); l > 0 {
		cpp.WriteString(models.IndentString())
	}
	cpp.WriteString("std::ostream &operator<<(std::ostream &_Stream, const ")
	cpp.WriteString(s.OutId())
	cpp.WriteString(genericsSerie)
	cpp.WriteString(" &_Src) {\n")
	models.AddIndent()
	cpp.WriteString(models.IndentString())
	cpp.WriteString(`_Stream << "`)
	cpp.WriteString(s.Ast.Id)
	cpp.WriteString("{\";\n")
	for i, field := range s.Ast.Fields {
		cpp.WriteString(models.IndentString())
		cpp.WriteString(`_Stream << "`)
		cpp.WriteString(field.Id)
		cpp.WriteString(`:" << _Src.`)
		cpp.WriteString(field.OutId())
		if i+1 < len(s.Ast.Fields) {
			cpp.WriteString(" << \", \"")
		}
		cpp.WriteString(";\n")
	}
	cpp.WriteString(models.IndentString())
	cpp.WriteString("_Stream << \"}\";\n")
	cpp.WriteString(models.IndentString())
	cpp.WriteString("return _Stream;\n")
	models.DoneIndent()
	cpp.WriteString(models.IndentString())
	cpp.WriteString("}")
	return cpp.String()
}

func (s xstruct) String() string {
	var cpp strings.Builder
	cpp.WriteString(s.decldefString())
	cpp.WriteString("\n\n")
	cpp.WriteString(s.ostream())
	return cpp.String()
}

// Generics returns generics of type.
//
// This function is should be have this function
// for Genericable & CompiledStruct interface of ast package.
func (s *xstruct) Generics() []DataType {
	return s.generics
}

// SetGenerics set generics of type.
//
// This function is should be have this function
// for Genericable & CompiledStruct interface of ast package.
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
		v.Expr.Model = exprNode{xapi.CppSelf}
	} else {
		v.Expr.Model = exprNode{tokens.STAR + xapi.CppSelf}
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
				gs.WriteString(generic.Kind)
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
