package sema

import (
	"github.com/julelang/jule/lex"
)

// Trait.
type Trait struct {
	Token   lex.Token
	Ident   string
	Public  bool
	Doc     string
	Methods []*Fn
}
