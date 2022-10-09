package models

import (
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/pkg/juleapi"
)

// ConcurrentCall is the AST model of concurrent calls.
type ConcurrentCall struct {
	Token  lex.Token
	Expr Expr
}

func (cc ConcurrentCall) String() string {
	return juleapi.ToConcurrentCall(cc.Expr.String())
}
