package models

import "github.com/jule-lang/jule/lex"

// Arg is AST model of argument.
type Arg struct {
	Token    lex.Token
	TargetId string
	Expr     Expr
	CastType *Type
}

func (a Arg) String() string {
	if a.CastType != nil {
		return "static_cast<" + a.CastType.String() + ">(" + a.Expr.String() + ")"
	}
	return a.Expr.String()
}
