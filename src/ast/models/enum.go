package models

import (
	"strings"

	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/pkg/juleapi"
)

// EnumItem is the AST model of enumerator items.
type EnumItem struct {
	Token   lex.Token
	Id      string
	Expr    Expr
	ExprTag any
}

func (ei EnumItem) String() string {
	var cpp strings.Builder
	cpp.WriteString(juleapi.OutId(ei.Id, ei.Token.File.Addr()))
	cpp.WriteString(" = ")
	cpp.WriteString(ei.Expr.String())
	return cpp.String()
}

// Enum is the AST model of enumerator statements.
type Enum struct {
	Pub   bool
	Token   lex.Token
	Id    string
	Type  Type
	Items []*EnumItem
	Used  bool
	Desc  string
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

func (e Enum) String() string {
	var cpp strings.Builder
	cpp.WriteString("enum ")
	cpp.WriteString(juleapi.OutId(e.Id, e.Token.File.Addr()))
	cpp.WriteByte(':')
	cpp.WriteString(e.Type.String())
	cpp.WriteString(" {\n")
	AddIndent()
	for _, item := range e.Items {
		cpp.WriteString(IndentString())
		cpp.WriteString(item.String())
		cpp.WriteString(",\n")
	}
	DoneIndent()
	cpp.WriteString("};")
	return cpp.String()
}
