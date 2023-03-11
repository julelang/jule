package ast

import "github.com/julelang/jule/lex"

// Abstract syntax tree.
type Ast struct {
	File     *lex.File
	UseDecls []*UseDecl
	Impls    []*Impl
	Comments []*Comment

	// Possible types:
	//  *EnumDecl
	//  *FnDecl
	//  *StructDecl
	//  *TraitDecl
	//  *TypeAliasDecl
	//  *VarDecl
	Decls []Node
}
