package models

import (
	"strconv"
	"strings"

	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/pkg/juleapi"
	"github.com/julelang/jule/pkg/juletype"
)

// Struct is the AST model of structures.
type Struct struct {
	Token       lex.Token
	Id          string
	Pub         bool
	Fields      []*Var
	Attributes  []Attribute
	Generics    []*GenericType
	Owner       any
	Origin      *Struct
	Traits      []*Trait // Implemented traits
	Defines     *Defmap
	Used        bool
	Doc         string
	CppLinked   bool
	Constructor *Fn
	Depends     []*Struct
	Order       int

	generics    []Type // Instance generics.
}

func (s *Struct) IsSameBase(s2 *Struct) bool {
	return s.Origin == s2.Origin
}

func (s *Struct) IsDependedTo(s2 *Struct) bool {
	for _, d := range s.Origin.Depends {
		if s2.IsSameBase(d) {
			return true
		}
	}
	return false
}

// OutId returns juleapi.OutId of struct.
func (s *Struct) OutId() string {
	if s.CppLinked {
		return s.Id
	}
	return juleapi.OutId(s.Id, s.Token.File.Addr())
}

// Generics returns generics of instance.
//
// This function is should be have this function
// for Genericable interface of ast package.
func (s *Struct) GetGenerics() []Type { return s.generics }

// SetGenerics set generics of instance.
//
// This function is should be have this function
// for Genericable interface of ast package.
func (s *Struct) SetGenerics(generics []Type) { s.generics = generics }

func (s *Struct) SelfVar(receiver *Var) *Var {
	v := new(Var)
	v.Token = s.Token
	v.Type = receiver.Type
	v.Type.Tag = s
	v.Type.Id = juletype.STRUCT
	v.Mutable = receiver.Mutable
	v.Id = lex.KND_SELF
	return v
}

func (s *Struct) AsTypeKind() string {
	var dts strings.Builder
	dts.WriteString(s.Id)
	if len(s.Generics) > 0 {
		dts.WriteByte('[')
		var gs strings.Builder
		// Instance
		if len(s.generics) > 0 {
			for _, generic := range s.GetGenerics() {
				gs.WriteString(generic.Kind)
				gs.WriteByte(',')
			}
		} else {
			for _, generic := range s.Generics {
				gs.WriteString(generic.Id)
				gs.WriteByte(',')
			}
		}
		dts.WriteString(gs.String()[:gs.Len()-1])
		dts.WriteByte(']')
	}
	return dts.String()
}

func (s *Struct) HasTrait(t *Trait) bool {
	for _, st := range s.Origin.Traits {
		if t == st {
			return true
		}
	}
	return false
}

func (s *Struct) DeclDefString() string {
	var cpp strings.Builder
	for _, f := range s.Defines.Funcs {
		if f.Used {
			cpp.WriteString(IndentString())
			cpp.WriteString(f.StringOwner(s.OutId()))
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

func (s *Struct) GetSelfRefVarType() Type {
	var t Type
	t.Id = juletype.STRUCT
	t.Kind = lex.KND_AMPER + s.Id
	t.Tag = s
	t.Token = s.Token
	return t
}

func (s *Struct) SelfRefVarStr() string {
	var cpp strings.Builder
	cpp.WriteString(s.GetSelfRefVarType().String())
	cpp.WriteString(" self{ nil };")
	return cpp.String()
}

func (s *Struct) cppGenerics() (def string, serie string) {
	if len(s.Generics) == 0 {
		return "", ""
	}
	var cppDef strings.Builder
	cppDef.WriteString("template<typename ")
	var cppSerie strings.Builder
	cppSerie.WriteByte('<')
	for i := range s.Generics {
		cppSerie.WriteByte('T')
		cppSerie.WriteString(strconv.Itoa(i))
		cppSerie.WriteByte(',')
	}
	serie = cppSerie.String()[:cppSerie.Len()-1] + ">"
	cppDef.WriteString(serie[1:])
	cppDef.WriteByte('\n')
	return cppDef.String(), serie
}

func (s *Struct) OStream() string {
	var cpp strings.Builder
	genericsDef, genericsSerie := s.cppGenerics()
	cpp.WriteString(IndentString())
	if l, _ := cpp.WriteString(genericsDef); l > 0 {
		cpp.WriteString(IndentString())
	}
	cpp.WriteString("std::ostream &operator<<(std::ostream &_Stream, const ")
	cpp.WriteString(s.OutId())
	cpp.WriteString(genericsSerie)
	cpp.WriteString(" &_Src) {\n")
	AddIndent()
	cpp.WriteString(IndentString())
	cpp.WriteString(`_Stream << "`)
	cpp.WriteString(s.Id)
	cpp.WriteString("{\";\n")
	for i, field := range s.Fields {
		cpp.WriteString(IndentString())
		cpp.WriteString(`_Stream << "`)
		cpp.WriteString(field.Id)
		cpp.WriteString(`:" << _Src.`)
		cpp.WriteString(field.OutId())
		if i+1 < len(s.Fields) {
			cpp.WriteString(" << \", \"")
		}
		cpp.WriteString(";\n")
	}
	cpp.WriteString(IndentString())
	cpp.WriteString("_Stream << \"}\";\n")
	cpp.WriteString(IndentString())
	cpp.WriteString("return _Stream;\n")
	DoneIndent()
	cpp.WriteString(IndentString())
	cpp.WriteString("}")
	return cpp.String()
}

func (s *Struct) Operators() string {
	outid := s.OutId()
	genericsDef, genericsSerie := s.cppGenerics()
	var cpp strings.Builder
	cpp.WriteString(IndentString())
	if l, _ := cpp.WriteString(genericsDef); l > 0 {
		cpp.WriteString(IndentString())
	}
	cpp.WriteString("inline bool operator==(const ")
	cpp.WriteString(outid)
	cpp.WriteString(genericsSerie)
	cpp.WriteString(" &_Src) {")
	if len(s.Defines.Globals) > 0 {
		AddIndent()
		cpp.WriteByte('\n')
		cpp.WriteString(IndentString())
		var expr strings.Builder
		expr.WriteString("return ")
		AddIndent()
		for _, g := range s.Defines.Globals {
			expr.WriteByte('\n')
			expr.WriteString(IndentString())
			expr.WriteString("this->")
			gid := g.OutId()
			expr.WriteString(gid)
			expr.WriteString(" == _Src.")
			expr.WriteString(gid)
			expr.WriteString(" &&")
		}
		DoneIndent()
		cpp.WriteString(expr.String()[:expr.Len()-3])
		cpp.WriteString(";\n")
		DoneIndent()
		cpp.WriteString(IndentString())
		cpp.WriteByte('}')
	} else {
		cpp.WriteString(" return true; }")
	}
	cpp.WriteString("\n\n")
	cpp.WriteString(IndentString())
	if l, _ := cpp.WriteString(genericsDef); l > 0 {
		cpp.WriteString(IndentString())
	}
	cpp.WriteString("inline bool operator!=(const ")
	cpp.WriteString(outid)
	cpp.WriteString(genericsSerie)
	cpp.WriteString(" &_Src) { return !this->operator==(_Src); }")
	return cpp.String()
}

func (s *Struct) SelfVarInitStatementStr() string {
	var cpp strings.Builder
	cpp.WriteString("this->self = ")
	cpp.WriteString(s.GetSelfRefVarType().String())
	cpp.WriteString("(this, nil);")
	return cpp.String()
}

func (s *Struct) CppConstructor() string {
	var cpp strings.Builder
	cpp.WriteString(IndentString())
	cpp.WriteString(s.OutId())
	cpp.WriteString(ParamsToCpp(s.Constructor.Params))
	cpp.WriteString(" noexcept {\n")
	AddIndent()
	cpp.WriteString(IndentString())
	cpp.WriteString(s.SelfVarInitStatementStr())
	cpp.WriteByte('\n')
	if len(s.Defines.Globals) > 0 {
		for i, g := range s.Defines.Globals {
			cpp.WriteByte('\n')
			cpp.WriteString(IndentString())
			cpp.WriteString("this->")
			cpp.WriteString(g.OutId())
			cpp.WriteString(" = ")
			cpp.WriteString(s.Constructor.Params[i].OutId())
			cpp.WriteByte(';')
		}
	}
	DoneIndent()
	cpp.WriteByte('\n')
	cpp.WriteString(IndentString())
	cpp.WriteByte('}')
	return cpp.String()
}

func (s *Struct) CppDestructor() string {
	var cpp strings.Builder
	cpp.WriteByte('~')
	cpp.WriteString(s.OutId())
	cpp.WriteString("(void) noexcept { /* heap allocations managed by traits or references */ this->self._ref = nil; }")
	return cpp.String()
}

func (s Struct) String() string {
	var cpp strings.Builder
	cpp.WriteString(s.DeclDefString())
	cpp.WriteString("\n\n")
	cpp.WriteString(s.OStream())
	return cpp.String()
}

func (s *Struct) CppTraits() string {
	if len(s.Traits) == 0 {
		return ""
	}
	var cpp strings.Builder
	cpp.WriteString(": ")
	for _, t := range s.Traits {
		cpp.WriteString("public ")
		cpp.WriteString(t.OutId())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1]
}

func (s *Struct) PlainPrototype() string {
	var cpp strings.Builder
	cpp.WriteString(GenericsToCpp(s.Generics))
	cpp.WriteByte('\n')
	cpp.WriteString("struct ")
	cpp.WriteString(s.OutId())
	cpp.WriteByte(';')
	return cpp.String()
}

func (s *Struct) Prototype() string {
	var cpp strings.Builder
	cpp.WriteString(GenericsToCpp(s.Generics))
	cpp.WriteByte('\n')
	cpp.WriteString("struct ")
	outid := s.OutId()
	cpp.WriteString(outid)
	cpp.WriteString(s.CppTraits())
	cpp.WriteString(" {\n")
	AddIndent()
	cpp.WriteString(IndentString())
	cpp.WriteString(s.SelfRefVarStr())
	cpp.WriteString("\n\n")
	if len(s.Defines.Globals) > 0 {
		for _, g := range s.Defines.Globals {
			cpp.WriteString(IndentString())
			cpp.WriteString(g.FieldString())
			cpp.WriteByte('\n')
		}
		cpp.WriteString("\n\n")
		cpp.WriteString(IndentString())
		cpp.WriteString(s.CppConstructor())
		cpp.WriteString("\n\n")
	}
	cpp.WriteString(IndentString())
	cpp.WriteString(s.CppDestructor())
	cpp.WriteString("\n\n")
	cpp.WriteString(IndentString())
	cpp.WriteString(outid)
	cpp.WriteString("(void) noexcept { ")
	cpp.WriteString(s.SelfVarInitStatementStr())
	cpp.WriteString(" }\n\n")
	for _, f := range s.Defines.Funcs {
		if f.Used {
			cpp.WriteString(IndentString())
			cpp.WriteString(f.Prototype(""))
			cpp.WriteString("\n\n")
		}
	}
	cpp.WriteString(s.Operators())
	cpp.WriteByte('\n')
	DoneIndent()
	cpp.WriteString(IndentString())
	cpp.WriteString("};")
	return cpp.String()
}
