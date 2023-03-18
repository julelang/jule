// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"github.com/julelang/jule/lex"
)

// Trait.
type Trait struct {
	Token   lex.Token
	Ident   string
	Public  bool
	Doc     string
	Methods []*Fn
}

// Implement: Kind
// Returns Trait's identifier.
func (t Trait) To_str() string { return t.Ident }

// Returns method by identifier.
// Returns nil if not exist any method in this identifier.
func (t *Trait) Find_method(ident string) *Fn {
	for _, f := range t.Methods {
		if f.Ident == ident {
			return f
		}
	}
	return nil
}
