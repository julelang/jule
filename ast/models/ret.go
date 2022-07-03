package models

import "strings"

// Ret is return statement AST model.
type Ret struct {
	Tok  Tok
	Expr Expr
}

func (r Ret) String() string {
	var cxx strings.Builder
	cxx.WriteString("return ")
	cxx.WriteString(r.Expr.String())
	cxx.WriteByte(';')
	return cxx.String()
}
