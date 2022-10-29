package parser

import "github.com/julelang/jule/lex"

type namespace struct {
	Id      string
	Token   lex.Token
	defines *Defmap
}
