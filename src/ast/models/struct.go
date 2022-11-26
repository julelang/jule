package models

import "github.com/julelang/jule/lex"

// Struct is the AST model of structures.
type Struct struct {
	Token      lex.Token
	Id         string
	Pub        bool
	Fields     []*Var
	Attributes []Attribute
	Generics   []*GenericType
	Owner      any
}
