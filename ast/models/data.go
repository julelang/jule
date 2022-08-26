package models

import "github.com/jule-lang/jule/lex"

// Data is AST model of data.
type Data struct {
	Token lex.Token
	Value string
	Type  Type
}

func (d Data) String() string {
	return d.Value
}
