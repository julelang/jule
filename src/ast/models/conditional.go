package models

import "github.com/julelang/jule/lex"

// If is the AST model of if expression.
type If struct {
	Token lex.Token
	Expr  Expr
	Block *Block
}

// Else is the AST model of else blocks.
type Else struct {
	Token lex.Token
	Block *Block
}

// Condition tree.
type Conditional struct {
	If    *If
	Elifs []*If
	Default  *Else
}
