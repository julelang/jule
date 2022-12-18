package models

import "github.com/julelang/jule/lex"

// TypeAlias is type alias declaration.
type TypeAlias struct {
	Owner   *Block
	Pub     bool
	Token   lex.Token
	Id      string
	Type    Type
	Doc     string
	Used    bool
	Generic bool
}
