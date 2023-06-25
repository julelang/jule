// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import "github.com/julelang/jule/lex"

// Implementation.
type Impl struct {
	// Equavalent to ast.Impl's Base field.
	Base lex.Token

	// Equavalent to ast.Impl's Dest field.
	Dest lex.Token

	// Equavalent to ast.Impl's Methods field.
	Methods []*Fn
}

// Reports whether implementation type is trait to structure.
func (ipl *Impl) Is_trait_impl() bool { return ipl.Base.Id != lex.ID_NA }

// Reports whether implementation type is append to destination structure.
func (ipl *Impl) Is_struct_impl() bool { return ipl.Base.Id == lex.ID_NA }
