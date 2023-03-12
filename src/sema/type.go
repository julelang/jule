// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Type's kind's type.
type TypeKind = any

// Type.
type Type struct {
	Ast  *ast.Type // Never changed by semantic analyzer.
	Kind TypeKind
}

// Reports whether type is parsed already.
func (t *Type) parsed() bool { return t.Kind != nil }

// Type alias.
type TypeAlias struct {
	Public     bool
	Cpp_linked bool
	Token      lex.Token
	Ident      string
	Kind       *Type
	Doc        string
	Refers     []*ast.IdentType
}
