// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Field.
type Field struct {
	Owner   *Struct
	Token   lex.Token
	Public  bool
	Mutable bool         // Interior mutability.
	Ident   string
	Kind    *TypeSymbol
}

func (f *Field) instance() *FieldIns {
	return &FieldIns{
		Decl: f,
		Kind: f.Kind.Kind,
	}
}

// Structure.
type Struct struct {
	// Used for type parsing.
	// Used declaration'sema sema for instance type checking.
	sema       *_Sema

	// This structure depended to these structures.
	// Only stores plain identifier references such as A, B, and MyStruct.
	// Not includes non-pain identifier references such as *A, &B, and []MyStruct.
	Depends    []*Struct

	Token      lex.Token
	Ident      string
	Fields     []*Field
	Methods    []*Fn
	Public     bool
	Cpp_linked bool
	Directives []*ast.Directive
	Doc        string
	Generics   []*ast.Generic
	Implements []*Trait

	// Structure instances for each unique type combination of structure.
	// Nil if structure is never used.
	Instances  []*StructIns
}

func (s *Struct) instance() *StructIns {
	// Returns already created instance for just one unique combination.
	if len(s.Generics) == 0 && len(s.Instances) == 1 {
		return s.Instances[0]
	}

	ins := &StructIns{
		Decl:    s,
		Fields:  make([]*FieldIns, len(s.Fields)),
		Methods: make([]*Fn, len(s.Methods)),
	}

	for i, f := range s.Fields {
		ins.Fields[i] = f.instance()
	}

	for i, f := range s.Methods {
		fins := new(Fn)
		*fins = *f
		ins.Methods[i] = fins
	}

	return ins
}

func (s *Struct) append_instance(ins *StructIns) {
	// Skip already created instance for just one unique combination.
	if len(s.Generics) == 0 && len(s.Instances) == 1 {
		return
	}

	for _, ains := range s.Instances {
		for i, ag := range ains.Generics {
			if ag.To_str() == ins.Generics[i].To_str() {
				// Instance exist.
				return
			}
		}
	}

	s.Instances = append(s.Instances, ins)
}

// Returns method by identifier.
// Returns nil if not exist any method in this identifier.
func (s *Struct) Find_method(ident string) *Fn {
	for _, f := range s.Methods {
		if f.Ident == ident {
			return f
		}
	}
	return nil
}

// Reports whether structure implements given trait.
func (s *Struct) Is_implements(t *Trait) bool {
	for _, it := range s.Implements {
		if t == it {
			return true
		}
	}
	return false
}

// Field structure.
type FieldIns struct {
	Decl *Field
	Kind *TypeKind
}

// Strucutre instance.
type StructIns struct {
	Decl     *Struct
	Generics []*TypeKind
	Fields   []*FieldIns
	Methods  []*Fn
}

// Implement: Kind
// Returns Struct's type kind as string.
func (s StructIns) To_str() string {
	kind := ""
	kind += s.Decl.Ident
	if len(s.Generics) > 0 {
		kind += "["
		for _, g := range s.Generics {
			kind += g.To_str()
			kind += ","
		}
		kind = kind[:len(kind)-1] // Remove comma.
		kind += "]"
	}
	return kind
}

// Returns method by identifier.
// Returns nil if not exist any method in this identifier.
func (s *StructIns) Find_method(ident string) *Fn {
	for _, f := range s.Methods {
		if f.Ident == ident {
			return f
		}
	}
	return nil
}

// Returns field by identifier.
// Returns nil if not exist any field in this identifier.
func (s *StructIns) Find_field(ident string) *FieldIns {
	for _, f := range s.Fields {
		if f.Decl.Ident == ident {
			return f
		}
	}
	return nil
}
