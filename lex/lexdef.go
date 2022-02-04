package lex

import "github.com/the-xlang/x/pkg/io"

// Lex is lexer of Fract.
type Lex struct {
	File     *io.FILE
	Position int
	Column   int
	Line     int
	Errors   []string
}

// New Lex instance.
func NewLex(f *io.FILE) *Lex {
	l := new(Lex)
	l.File = f
	l.Line = 1
	l.Column = 1
	l.Position = 0
	return l
}
