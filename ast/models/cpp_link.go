package models

import "github.com/jule-lang/jule/lex"

// CppLink is attribtue AST model.
type CppLink struct {
	Token lex.Token
	Link  *Fn
}
