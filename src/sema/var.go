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
	Scope      *ast.Scope
	Token      lex.Token
	Ident      string
	Cpp_linked bool
	Constant   bool
	Mutable    bool
	Public     bool
	Doc        string
	Kind       *Type
	Expr       *ast.Expr
}
