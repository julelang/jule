package models

import "strings"

// Ret is return statement AST model.
type Ret struct {
	Tok  Tok
	Expr Expr
}

func (r Ret) String() string {
	var cpp strings.Builder
	cpp.WriteString("return ")
	cpp.WriteString(r.Expr.String())
	cpp.WriteByte(';')
	return cpp.String()
}
