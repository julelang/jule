package lex

// Tok is lexer token.
type Tok struct {
	File   *File
	Row    int
	Column int
	Kind   string
	Id     uint8
}
