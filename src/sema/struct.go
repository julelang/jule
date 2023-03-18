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
	Token   lex.Token
	Public  bool
	Mutable bool      // Interior mutability.
	Ident   string
	Kind    *TypeSymbol
}

// Structure.
type Struct struct {
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

func (s *Struct) instance() *StructIns {
	return &StructIns{Decl: s}
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
