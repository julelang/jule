// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package ast

import "github.com/julelang/jule/lex"

// Abstract syntax tree.
type Ast struct {
	File           *lex.File
	Top_directives []*Directive
	Use_decls      []*UseDecl
	Impls          []*Impl
	Comments       []*Comment

	// Possible types:
	//  *EnumDecl
	//  *FnDecl
	//  *StructDecl
	//  *TraitDecl
	//  *TypeAliasDecl
	//  *VarDecl
	Decls []Node
}
