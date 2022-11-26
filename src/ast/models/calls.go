package models

import (
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/pkg/juleapi"
)

// ConcurrentCall is the AST model of concurrent calls.
type ConcurrentCall struct {
	Token  lex.Token
	Expr Expr
}

func (cc ConcurrentCall) String() string {
	return juleapi.ToConcurrentCall(cc.Expr.String())
}
