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
