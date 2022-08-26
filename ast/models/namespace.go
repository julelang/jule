package models

import "github.com/jule-lang/jule/lex"

// Namespace is the AST model of namespace statements.
type Namespace struct {
	Token       lex.Token
	Identifiers []string
	Tree        []Object
}
