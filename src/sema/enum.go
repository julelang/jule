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

// Reports whether item has auto expression.
func (ei *EnumItem) Auto_expr() bool { return ei.Value.Expr == nil }

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

// Returns item by identifier.
// Returns nil if not exist any item in this identifier.
func (e *Enum) Find_item(ident string) *EnumItem {
	for _, item := range e.Items {
		if item.Ident == ident {
			return item
		}
	}
	return nil
}
