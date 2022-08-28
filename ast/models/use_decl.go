package models

import "github.com/jule-lang/jule/lex"

// UseDecl is the AST model of use declaration.
type UseDecl struct {
	Token      lex.Token
	Path       string
	Cpp        bool
	LinkString string
	FullUse    bool
	Selectors  []lex.Token
}
