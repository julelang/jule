package models

import "github.com/jule-lang/jule/lex"

// Impl is the AST model of impl statement.
type Impl struct {
	Base  lex.Token
	Target Type
	Tree   []Object
}
