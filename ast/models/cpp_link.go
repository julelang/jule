package models

import "github.com/jule-lang/jule/lex"

// CppLinkFn is linked function AST model.
type CppLinkFn struct {
	Token lex.Token
	Link  *Fn
}

// CppLinkVar is linked variable AST model.
type CppLinkVar struct {
	Token lex.Token
	Link  *Var
}

// CppLinkStruct is linked structure AST model.
type CppLinkStruct struct {
	Token lex.Token
	Link  Struct
}
