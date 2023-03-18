// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Enum item.
type EnumItem struct {
	Token lex.Token
	Ident string
	Value *Value
}

// Enum.
type Enum struct {
	Token  lex.Token
	Public bool
	Ident  string
	Kind   *TypeSymbol
	Items  []*EnumItem
	Doc    string
	Refers []*ast.IdentType // Referred identifiers.
}

// Implement: Kind
// Returns Enum's identifier.
func (e Enum) To_str() string { return e.Ident }
