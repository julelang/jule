// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Variable.
type Var struct {
	Scope      *ast.ScopeTree
	Token      lex.Token
	Ident      string
	Cpp_linked bool
	Constant   bool
	Mutable    bool
	Public     bool
	Doc        string
	Kind       *TypeSymbol
	Value      *Value

	// This variable depended to these variables for initialization expression.
	// Nil if not global variable.
	Depends    []*Var
}

// Reports whether variable is auto-typed.
func (v *Var) Is_auto_typed() bool { return v.Kind == nil || v.Kind.Decl == nil }
