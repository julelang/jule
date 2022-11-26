package models

import "github.com/julelang/jule/lex"

// Expression AST model for binop.
type BinopExpr struct {
	Tokens []lex.Token
}

// Binop is AST model of the binary operation.
type Binop struct {
	L  any
	R  any
	Op lex.Token
}

// Expr is AST model of expression.
type Expr struct {
	Tokens []lex.Token
	Op     any
	Model  IExprModel
}

func (e *Expr) IsNotBinop() bool {
	switch e.Op.(type) {
	case BinopExpr:
		return true
	default:
		return false
	}
}

func (e *Expr) IsEmpty() bool { return e.Op == nil }

func (e Expr) String() string {
	if e.Model != nil {
		return e.Model.String()
	}
	return ""
}
