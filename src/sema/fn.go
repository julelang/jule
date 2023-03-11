package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Function.
type Fn struct {
	Token        lex.Token
	Unsafety     bool
	Public       bool
	Cpp_linked   bool
	Ident        string
	Directives   []*ast.Directive
	Doc_comments *ast.CommentGroup
	Scope        *ast.Scope
	Generics     []*ast.Generic
	Result       *ast.RetType
	Params       []*ast.Param
}
