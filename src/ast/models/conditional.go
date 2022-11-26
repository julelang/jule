package models

import (
	"strings"

	"github.com/julelang/jule/lex"
)

// If is the AST model of if expression.
type If struct {
	Token lex.Token
	Expr  Expr
	Block *Block
}

func (i If) String() string {
	var cpp strings.Builder
	cpp.WriteString("if (")
	cpp.WriteString(i.Expr.String())
	cpp.WriteString(") ")
	cpp.WriteString(i.Block.String())
	return cpp.String()
}

// Else is the AST model of else blocks.
type Else struct {
	Token lex.Token
	Block *Block
}

func (e Else) String() string {
	var cpp strings.Builder
	cpp.WriteString("else ")
	cpp.WriteString(e.Block.String())
	return cpp.String()
}

// Condition tree.
type Conditional struct {
	If    *If
	Elifs []*If
	Default  *Else
}

func (c Conditional) String() string {
	var cpp strings.Builder
	cpp.WriteString(c.If.String())
	for _, elif := range c.Elifs {
		cpp.WriteString(" else ")
		cpp.WriteString(elif.String())
	}
	if c.Default != nil {
		cpp.WriteByte(' ')
		cpp.WriteString(c.Default.String())
	}
	return cpp.String()
}
