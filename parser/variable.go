package parser

import (
	"github.com/the-xlang/x/lex"
)

type variable struct {
	Name  string
	Token lex.Token
	Type  uint8
}
