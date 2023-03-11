package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Variable.
type Var struct {
	Scope        *ast.Scope
	Token        lex.Token
	Ident        string
	Cpp_linked   bool
	Constant     bool
	Mutable      bool
	Public       bool
	Doc_comments *ast.CommentGroup
	Kind         *ast.Type
}
