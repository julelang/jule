package models

import "github.com/jule-lang/jule/lex"

// Attribute is attribtue AST model.
type Attribute struct {
	Token lex.Token
	Tag   string
}

func (a Attribute) String() string {
	return a.Tag
}
