package models

import "strings"

// If is the AST model of if expression.
type If struct {
	Tok   Tok
	Expr  Expr
	Block *Block
}

func (ifast If) String() string {
	var cpp strings.Builder
	cpp.WriteString("if (")
	cpp.WriteString(ifast.Expr.String())
	cpp.WriteString(") ")
	cpp.WriteString(ifast.Block.String())
	return cpp.String()
}

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

// Else is the AST model of else blocks.
type Else struct {
	Tok   Tok
	Block *Block
}

func (elseast Else) String() string {
	var cpp strings.Builder
	cpp.WriteString("else ")
	cpp.WriteString(elseast.Block.String())
	return cpp.String()
}
