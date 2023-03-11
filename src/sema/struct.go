package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Field.
type Field struct {
	Token   lex.Token
	Public  bool
	Mutable bool      // Interior mutability.
	Ident   string
	Kind    *ast.Type
}

// Structure.
type Struct struct {
	Token      lex.Token
	Ident      string
	Fields     []*Field
	Public     bool
	Cpp_linked bool
	Directives []*ast.Directive
	Doc        string
	Generics   []*ast.Generic
}
