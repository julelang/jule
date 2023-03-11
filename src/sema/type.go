// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Type alias.
type TypeAlias struct {
	Public     bool
	Cpp_linked bool
	Token      lex.Token
	Ident      string
	Kind       *ast.Type
	Doc        string
}
