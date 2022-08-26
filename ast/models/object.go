package models

import "github.com/jule-lang/jule/lex"

// Object is an element of AST.
type Object struct {
	Token lex.Token
	Data  any
}
