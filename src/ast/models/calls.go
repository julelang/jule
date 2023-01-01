package models

import "github.com/julelang/jule/lex"

// ConcurrentCall is the AST model of concurrent calls.
type ConcurrentCall struct {
	Token lex.Token
	Expr  Expr
}
