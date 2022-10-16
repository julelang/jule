package parser

import "github.com/jule-lang/jule/lex"

type namespace struct {
	Id      string
	Token   lex.Token
	defines *Defmap
}
