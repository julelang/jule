package models

import "github.com/julelang/jule/lex"

// Statement is statement.
type Statement struct {
	Token          lex.Token
	Data           any
	WithTerminator bool
}

// ExprStatement is AST model of expression statement in block.
type ExprStatement struct {
	Expr Expr
}
