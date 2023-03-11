package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Type alias.
type TypeAlias struct {
	Public       bool
	Cpp_linked   bool
	Token        lex.Token
	Ident        string
	Kind         *ast.Type
	Doc_comments *ast.CommentGroup
}
