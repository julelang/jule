package models

import "github.com/jule-lang/jule/lex"

// Arg is AST model of argument.
type Arg struct {
	Token    lex.Token
	TargetId string
	Expr     Expr
}

func (a Arg) String() string {
	return a.Expr.String()
}
