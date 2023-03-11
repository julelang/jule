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
	Kind    *ast.Type
}

// Structure.
type Struct struct {
	Token      lex.Token
	Ident      string
	Fields     []*Field
	Public     bool
	Cpp_linked bool
	Directives []*ast.Directive
	Doc        string
	Generics   []*ast.Generic
	Implements []*Trait
}
