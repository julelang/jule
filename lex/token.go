package lex

import "github.com/the-xlang/x/pkg/xio"

// Token is lexer token.
type Token struct {
	File   *xio.File
	Row    int
	Column int
	Kind   string
	Id     uint8
}

// Token types.
const (
	NA        uint8 = 0
	DataType  uint8 = 1
	Id        uint8 = 2
	Brace     uint8 = 3
	Ret       uint8 = 4
	SemiColon uint8 = 5
	Value     uint8 = 6
	Operator  uint8 = 7
	Comma     uint8 = 8
	Const     uint8 = 9
	Type      uint8 = 10
	Colon     uint8 = 11
	At        uint8 = 12
	New       uint8 = 13
	Free      uint8 = 14
	Iter      uint8 = 15
	Break     uint8 = 16
	Continue  uint8 = 17
	In        uint8 = 18
	If        uint8 = 19
	Else      uint8 = 20
	Volatile  uint8 = 21
	Comment   uint8 = 22
	Use       uint8 = 23
	Dot       uint8 = 24
	Pub       uint8 = 25
)
