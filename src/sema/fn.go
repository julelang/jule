// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"strings"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Return type.
type RetType struct {
	Kind   *TypeSymbol
	Idents []lex.Token
}

// Parameter.
type Param struct {
	Token    lex.Token
	Mutable  bool
	Variadic bool
	Kind     *TypeSymbol
	Ident    string
}

// Reports whether parameter is self (receiver) parameter.
func (p *Param) Is_self() bool { return strings.HasSuffix(p.Ident, lex.KND_SELF) }
// Reports whether self (receiver) parameter is reference.
func (p *Param) Is_ref() bool { return p.Ident != "" && p.Ident[0] == '&'}

// Implement: Kind
// Returns Param's type kind as string.
func (p Param) To_str() string {
	s := ""
	if p.Mutable {
		s += lex.KND_MUT + " "
	}
	if p.Variadic {
		s += lex.KND_TRIPLE_DOT
	}
	s += p.Kind.Kind.To_str()
	return s
}

// Function.
type Fn struct {
	Token      lex.Token
	Unsafety   bool
	Public     bool
	Cpp_linked bool
	Ident      string
	Directives []*ast.Directive
	Doc        string
	Scope      *ast.Scope
	Generics   []*ast.Generic
	Result     *RetType
	Params     []*Param

	// Function instances for each unique type combination of function call.
	// Nil if function is never used.
	Combines   []*FnIns
}

// Reports whether return type is void.
func (f *Fn) Is_void() bool { return f.Result == nil }
// Reports whether function is method.
func (f *Fn) Is_method() bool { return len(f.Params) > 0 && f.Params[0].Is_self() }

func (f *Fn) instance() *FnIns {
	return &FnIns{Decl: f}
}

// Function instance.
type FnIns struct {
	Decl     *Fn
	Generics []*TypeKind
	Params   []*TypeKind
	Result   *TypeKind
	Scope    *ast.Scope
}

// Implement: Kind
// Returns Fn's type kind as string.
func (f FnIns) To_str() string {
	s := ""
	if f.Decl.Unsafety {
		s += "unsafe "
	}
	s += "fn"
	if len(f.Generics) > 0 {
		s += "["
		for i, t := range f.Generics {
			s += t.To_str()
			if i+1 < len(f.Generics) {
				s += ","
			}
		}
		s += "]"
	}
	s += "("
	n := len(f.Params)
	if n > 0 {
		for _, p := range f.Params {
			s += p.To_str()
			s += ","
		}
		s = s[:len(s)-1] // Remove comma.
	}
	s += ")"
	if !f.Decl.Is_void() {
		s += f.Result.To_str()
	}
	return s
}
