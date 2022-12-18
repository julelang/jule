package models

import "github.com/julelang/jule/lex"

// AssignLeft is selector for assignment operation.
type AssignLeft struct {
	Var    Var
	Expr   Expr
	Ignore bool
}

// Assign is assignment AST model.
type Assign struct {
	Setter      lex.Token
	Left        []AssignLeft
	Right       []Expr
	IsExpr      bool
	MultipleRet bool
}
