package models

import (
	"fmt"
	"strings"
)

// Statement is statement.
type Statement struct {
	Tok            Tok
	Val            any
	WithTerminator bool
}

func (s Statement) String() string { return fmt.Sprint(s.Val) }

// ExprStatement is AST model of expression statement in block.
type ExprStatement struct{ Expr Expr }

func (be ExprStatement) String() string {
	var cxx strings.Builder
	cxx.WriteString(be.Expr.String())
	cxx.WriteByte(';')
	return cxx.String()
}
