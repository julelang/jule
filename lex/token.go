package lex

import "github.com/the-xlang/x/pkg/io"

// Token is lexer token.
type Token struct {
	File   *io.FILE
	Line   int
	Column int
	Value  string
	Type   uint
}

// Token types.
const (
	NA    = 0
	Type  = 1
	Name  = 2
	Brace = 3
)
