package models

// Namespace is the AST model of namespace statements.
type Namespace struct {
	Tok  Tok
	Ids  []string
	Tree []Object
}
