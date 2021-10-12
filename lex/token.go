package lex

import "github.com/the-xlang/x/pkg/io"

// Token is lexer token.
type Token struct {
	File   *io.FILE
	Line   int
	Column int
	Value  string
	Type   uint8
}

// Token types.
const (
	NA        uint8 = 0
	DataType  uint8 = 1
	Name      uint8 = 2
	Brace     uint8 = 3
	Return    uint8 = 4
	SemiColon uint8 = 5
	Value     uint8 = 6
	Operator  uint8 = 7
	Comma     uint8 = 8
	Var       uint8 = 9
	Const     uint8 = 10
	Type      uint8 = 11
)
