package models

// Struct is the AST model of structures.
type Struct struct {
	Tok      Tok
	Id       string
	Pub      bool
	Fields   []*Var
	Generics []*GenericType
}
