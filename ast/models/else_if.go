package models

import "strings"

// ElseIf is the AST model of else if expression.
type ElseIf struct {
	Tok   Tok
	Expr  Expr
	Block Block
}

func (elif ElseIf) String() string {
	var cxx strings.Builder
	cxx.WriteString("else if (")
	cxx.WriteString(elif.Expr.String())
	cxx.WriteString(") ")
	cxx.WriteString(elif.Block.String())
	return cxx.String()
}
