package parser

import (
	"strconv"
	"strings"

	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/juleapi"
	"github.com/jule-lang/jule/pkg/juletype"
)

type structure struct {
	Ast         Struct
	origin      *structure
	Traits      []*trait // Implemented traits
	Defines     *DefineMap
	Used        bool
	Description string
	
	cpp_linked  bool
	constructor *Func
	depends     []*structure
	order       int
	generics    []Type // Instance generics.
}

func structure_instances_is_uses_same_base(s1, s2 *structure) bool {
	return s1.origin == s2.origin
}

func (s *structure) depended_to(st *structure) bool {
	for _, d := range s.origin.depends {
		if structure_instances_is_uses_same_base(st, d) {
			return true
		}
	}
	return false
}

func (s *structure) hasTrait(t *trait) bool {
	for _, st := range s.origin.Traits {
		if t == st {
			return true
		}
	}
	return false
}

func (s *structure) cppGenerics() (def string, serie string) {
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

// OutId returns juleapi.OutId of struct.
//
// This function is should be have this function
// for CompiledStruct interface of ast package.
func (s *structure) OutId() string {
	if s.cpp_linked {
		return s.Ast.Id
	}
	return juleapi.OutId(s.Ast.Id, s.Ast.Token.File)
}

// CppLinked returns cpp_linked.
//
// This function is should be have this function
// for CompiledStruct interface of ast package.
func (s *structure) CppLinked() bool { return s.cpp_linked }

func (s *structure) operators() string {
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
	if len(s.Defines.Globals) > 0 {
		models.AddIndent()
		cpp.WriteByte('\n')
		cpp.WriteString(models.IndentString())
		var expr strings.Builder
		expr.WriteString("return ")
		models.AddIndent()
		for _, g := range s.Defines.Globals {
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

func (s *structure) self_var_init_statement_str() string {
	var cpp strings.Builder
	cpp.WriteString("this->self = ")
	cpp.WriteString(s.self_ref_var_type().String())
	cpp.WriteString("(this, nil);")
	return cpp.String()
}

func (s *structure) cppConstructor() string {
	var cpp strings.Builder
	cpp.WriteString(models.IndentString())
	cpp.WriteString(s.OutId())
	cpp.WriteString(paramsToCpp(s.constructor.Params))
	cpp.WriteString(" noexcept {\n")
	models.AddIndent()
	cpp.WriteString(models.IndentString())
	cpp.WriteString(s.self_var_init_statement_str())
	cpp.WriteByte('\n')
	if len(s.Defines.Globals) > 0 {
		for i, g := range s.Defines.Globals {
			cpp.WriteByte('\n')
			cpp.WriteString(models.IndentString())
			cpp.WriteString("this->")
			cpp.WriteString(g.OutId())
			cpp.WriteString(" = ")
			cpp.WriteString(s.constructor.Params[i].OutId())
			cpp.WriteByte(';')
		}
	}
	models.DoneIndent()
	cpp.WriteByte('\n')
	cpp.WriteString(models.IndentString())
	cpp.WriteByte('}')
	return cpp.String()
}

func (s *structure) cpp_destructor() string {
	var cpp strings.Builder
	cpp.WriteByte('~')
	cpp.WriteString(s.OutId())
	cpp.WriteString("(void) noexcept { /* heap allocations managed by traits or references */ this->self._ref = nil; }")
	return cpp.String()
}

func (s *structure) cppTraits() string {
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

func (s *structure) plainPrototype() string {
	var cpp strings.Builder
	cpp.WriteString(genericsToCpp(s.Ast.Generics))
	cpp.WriteByte('\n')
	cpp.WriteString("struct ")
	cpp.WriteString(s.OutId())
	cpp.WriteByte(';')
	return cpp.String()
}

func (s *structure) self_ref_var_type() Type {
	var t Type
	t.Id = juletype.Struct
	t.Kind = tokens.AMPER + s.Ast.Id
	t.Tag = s
	t.Token = s.Ast.Token
	return t
}

func (s *structure) self_ref_var_str() string {
	var cpp strings.Builder
	cpp.WriteString(s.self_ref_var_type().String())
	cpp.WriteString(" self{ nil };")
	return cpp.String()
}

func (s *structure) prototype() string {
	var cpp strings.Builder
	cpp.WriteString(genericsToCpp(s.Ast.Generics))
	cpp.WriteByte('\n')
	cpp.WriteString("struct ")
	outid := s.OutId()
	cpp.WriteString(outid)
	cpp.WriteString(s.cppTraits())
	cpp.WriteString(" {\n")
	models.AddIndent()
	cpp.WriteString(models.IndentString())
	cpp.WriteString(s.self_ref_var_str())
	cpp.WriteString("\n\n")
	if len(s.Defines.Globals) > 0 {
		for _, g := range s.Defines.Globals {
			cpp.WriteString(models.IndentString())
			cpp.WriteString(g.FieldString())
			cpp.WriteByte('\n')
		}
		cpp.WriteString("\n\n")
		cpp.WriteString(models.IndentString())
		cpp.WriteString(s.cppConstructor())
		cpp.WriteString("\n\n")
	}
	cpp.WriteString(models.IndentString())
	cpp.WriteString(s.cpp_destructor())
	cpp.WriteString("\n\n")
	cpp.WriteString(models.IndentString())
	cpp.WriteString(outid)
	cpp.WriteString("(void) noexcept { ")
	cpp.WriteString(s.self_var_init_statement_str())
	cpp.WriteString(" }\n\n")
	for _, f := range s.Defines.Funcs {
		if f.used {
			cpp.WriteString(models.IndentString())
			cpp.WriteString(f.Prototype(""))
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

func (s *structure) decldefString() string {
	var cpp strings.Builder
	for _, f := range s.Defines.Funcs {
		if f.used {
			cpp.WriteString(models.IndentString())
			cpp.WriteString(f.stringOwner(s.OutId()))
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

func (s *structure) ostream() string {
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

func (s structure) String() string {
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
func (s *structure) Generics() []Type {
	return s.generics
}

// SetGenerics set generics of type.
//
// This function is should be have this function
// for Genericable & CompiledStruct interface of ast package.
func (s *structure) SetGenerics(generics []Type) {
	s.generics = generics
}

func (s *structure) selfVar(receiver *Var) *Var {
	v := new(models.Var)
	v.Token = s.Ast.Token
	v.Type = receiver.Type
	v.Type.Tag = s
	v.Type.Id = juletype.Struct
	v.Mutable = receiver.Mutable
	v.Id = tokens.SELF
	return v
}

func (s *structure) dataTypeString() string {
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

func can_be_order(s *structure) bool {
	for _, d := range s.origin.depends {
		if d.origin.order < s.origin.order {
			return false
		}
	}
	return true
}

func order_structures(structures []*structure) {
	// As far as has been tested, there is no need for this indexing.
	/*for i, s := range structures {
		s.order = i
	}*/

	swapped := true
	for swapped {
		swapped = false
		i := 1
		for i < len(structures) {
			curr := &structures[i]
			if can_be_order(*curr) {
				(*curr).origin.order = i-1
				prev := &structures[i-1]
				(*prev).origin.order = i
				*curr, *prev = *prev, *curr
				swapped = true
			}
			i++
		}
	}
}
