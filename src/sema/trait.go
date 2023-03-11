package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Trait.
type Trait struct {
	Token        lex.Token
	Ident        string
	Public       bool
	Doc_comments *ast.CommentGroup
	Methods      []*Fn
}
