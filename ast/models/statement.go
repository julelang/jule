package models

import (
	"fmt"
	"strings"
)

// Statement is statement.
type Statement struct {
	Tok            Tok
	Data           any
	WithTerminator bool
	Block          *Block
}

func (s Statement) String() string {
	return fmt.Sprint(s.Data)
}

// ExprStatement is AST model of expression statement in block.
type ExprStatement struct {
	Expr Expr
}

func (be ExprStatement) String() string {
	var cxx strings.Builder
	cxx.WriteString(be.Expr.String())
	cxx.WriteByte(';')
	return cxx.String()
}
