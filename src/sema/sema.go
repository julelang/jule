package sema

import (
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// Semantic analyzer for tables.
// Accepts tables as files of package.
type _Sema struct {
	errors   []build.Log
	tables   []*SymbolTable
}

func (sa *_Sema) push_err(token lex.Token, key string, args ...any) {
	sa.errors = append(sa.errors, build.Log{
		Type:   build.ERR,
		Row:    token.Row,
		Column: token.Column,
		Path:   token.File.Path(),
		Text:   build.Errorf(key, args...),
	})
}

// Checks semantic errors of tables.
func (sa *_Sema) check() {
	// TODO: implement here.
}

func (sa *_Sema) analyze(tables []*SymbolTable) {
	sa.tables = tables
	// TODO: implement here.
}
