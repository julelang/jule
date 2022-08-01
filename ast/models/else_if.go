package models

import "strings"

// ElseIf is the AST model of else if expression.
type ElseIf struct {
	Tok   Tok
	Expr  Expr
	Block *Block
}

func (elif ElseIf) String() string {
	var cpp strings.Builder
	cpp.WriteString("else if (")
	cpp.WriteString(elif.Expr.String())
	cpp.WriteString(") ")
	cpp.WriteString(elif.Block.String())
	return cpp.String()
}
