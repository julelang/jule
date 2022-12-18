package models

import "github.com/julelang/jule/lex"

// IterForeach is foreach iteration profile.
type IterForeach struct {
	KeyA     Var
	KeyB     Var
	InToken    lex.Token
	Expr     Expr
	ExprType Type
}
