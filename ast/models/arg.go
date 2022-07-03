package models

// Arg is AST model of argument.
type Arg struct {
	Tok      Tok
	TargetId string
	Expr     Expr
}

func (a Arg) String() string { return a.Expr.String() }
