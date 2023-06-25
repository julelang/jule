// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"strconv"

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

// Reports whether Trait is built-in.
func (t *Trait) Is_builtin() bool { return t.Token.Id == lex.ID_NA }

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

// Returns function declaration text for trait checking.
// Returns special representation of function.
func to_trait_kind_str(f *FnIns) string {
	s := ""
	if f.Decl.Public {
		s += "p"
	}
	if f.Decl.Unsafety {
		s += "u"
	}
	s += "f&"
	s += f.Decl.Ident
	s += "#"

	if len(f.Generics) > 0 {
		for i := range f.Generics {
			s += strconv.Itoa(i)
		}
	} else if len(f.Decl.Generics) > 0 { // Use Decl's generic if not parsed yet.
		for i := range f.Decl.Generics {
			s += strconv.Itoa(i)
		}
	}

	s += "?"
	n := len(f.Params)
	if n > 0 {
		for _, p := range f.Params {
			s += p.To_str()
		}
	}
	s += "="
	if !f.Decl.Is_void() {
		s += f.Result.To_str()
	}
	return s
}
