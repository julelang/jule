package lex

import "github.com/jule-lang/jule/pkg/juleio"

// Token is lexer token.
type Token struct {
	File   *juleio.File
	Row    int
	Column int
	Kind   string
	Id     uint8
}
