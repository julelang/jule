package models

import "github.com/jule-lang/jule/pkg/juleapi"

// Defer is the AST model of deferred calls.
type Defer struct {
	Tok  Tok
	Expr Expr
}

func (d Defer) String() string {
	return juleapi.ToDeferredCall(d.Expr.String())
}

// ConcurrentCall is the AST model of concurrent calls.
type ConcurrentCall struct {
	Tok  Tok
	Expr Expr
}

func (cc ConcurrentCall) String() string {
	return juleapi.ToConcurrentCall(cc.Expr.String())
}
