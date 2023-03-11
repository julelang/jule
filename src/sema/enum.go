package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Enum item.
type EnumItem struct {
	Token lex.Token
	Ident string
	Expr *ast.Expr
}

// Enum.
type Enum struct {
	Token  lex.Token
	Public bool
	Ident  string
	Kind   *ast.Type
	Items  []*EnumItem
	Doc    string
}
