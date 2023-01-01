package models

import "github.com/julelang/jule/lex"

// EnumItem is the AST model of enumerator items.
type EnumItem struct {
	Token   lex.Token
	Id      string
	Expr    Expr
	ExprTag any
}

// Enum is the AST model of enumerator statements.
type Enum struct {
	Pub   bool
	Token lex.Token
	Id    string
	Type  Type
	Items []*EnumItem
	Used  bool
	Doc   string
}

// ItemById returns item by id if exist, nil if not.
func (e *Enum) ItemById(id string) *EnumItem {
	for _, item := range e.Items {
		if item.Id == id {
			return item
		}
	}
	return nil
}
