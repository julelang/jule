package parser

import (
	"github.com/the-xlang/x/lex"
)

// Variable is variable define representation.
type Variable struct {
	Name  string
	Token lex.Token
	Type  uint8
}
