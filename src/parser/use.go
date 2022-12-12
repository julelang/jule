package parser

import (
	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/lex"
)

type use struct {
	Defines    *models.Defmap
	Token      lex.Token
	CppLink    bool
	FullUse    bool
	Path       string
	LinkString string
	Selectors  []lex.Token
}
