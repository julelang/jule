package models

import "strings"

// If is the AST model of if expression.
type If struct {
	Tok   Tok
	Expr  Expr
	Block *Block
}

func (ifast If) String() string {
	var cxx strings.Builder
	cxx.WriteString("if (")
	cxx.WriteString(ifast.Expr.String())
	cxx.WriteString(") ")
	cxx.WriteString(ifast.Block.String())
	return cxx.String()
}
