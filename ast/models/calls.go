package models

import "github.com/the-xlang/xxc/pkg/xapi"

// Defer is the AST model of deferred calls.
type Defer struct {
	Tok  Tok
	Expr Expr
}

func (d Defer) String() string { return xapi.ToDeferredCall(d.Expr.String()) }

// ConcurrentCall is the AST model of concurrent calls.
type ConcurrentCall struct {
	Tok  Tok
	Expr Expr
}

func (cc ConcurrentCall) String() string {
	return xapi.ToConcurrentCall(cc.Expr.String())
}
