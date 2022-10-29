package models

import (
	"fmt"
	"strings"

	"github.com/julelang/jule/lex"
)

// Statement is statement.
type Statement struct {
	Token          lex.Token
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
