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
}

func (s Statement) String() string {
	return fmt.Sprint(s.Data)
}

// ExprStatement is AST model of expression statement in block.
type ExprStatement struct {
	Expr Expr
}

func (be ExprStatement) String() string {
	var cpp strings.Builder
	cpp.WriteString(be.Expr.String())
	cpp.WriteByte(';')
	return cpp.String()
}
