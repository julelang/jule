package models

import "github.com/jule-lang/jule/lex"

// Use is the AST model of use declaration.
type Use struct {
	Token      lex.Token
	Path       string
	Cpp        bool
	LinkString string
	FullUse    bool
	Selectors  []lex.Token
}
