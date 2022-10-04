package parser

import "github.com/jule-lang/jule/lex"

type use struct {
	defines    *DefineMap
	token   lex.Token
	cppLink bool
	
	FullUse    bool
	Path       string
	LinkString string
	Selectors  []lex.Token
}
