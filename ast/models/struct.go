package models

import "github.com/jule-lang/jule/lex"

// Struct is the AST model of structures.
type Struct struct {
	Token    lex.Token
	Id       string
	Pub      bool
	Fields   []*Var
	Generics []*GenericType
	Owner    any
}
