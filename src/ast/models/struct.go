package models

import (
	"strings"

	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/pkg/juleapi"
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
	v.Type.Id = struct_t
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

func (s *Struct) GetSelfRefVarType() Type {
	var t Type
	t.Id = struct_t
	t.Kind = lex.KND_AMPER + s.Id
	t.Tag = s
	t.Token = s.Token
	return t
}
