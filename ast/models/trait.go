package models

// Trait is the AST model of traits.
type Trait struct {
	Pub   bool
	Tok   Tok
	Id    string
	Desc  string
	Used  bool
	Funcs []*Func
}
