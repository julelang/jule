package models

import (
	"strings"

	"github.com/jule-lang/jule/pkg/juleapi"
)

// EnumItem is the AST model of enumerator items.
type EnumItem struct {
	Tok     Tok
	Id      string
	Expr    Expr
	ExprTag any
}

func (ei EnumItem) String() string {
	var cpp strings.Builder
	cpp.WriteString(juleapi.OutId(ei.Id, ei.Tok.File))
	cpp.WriteString(" = ")
	cpp.WriteString(ei.Expr.String())
	return cpp.String()
}

// Enum is the AST model of enumerator statements.
type Enum struct {
	Pub   bool
	Tok   Tok
	Id    string
	Type  DataType
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
	cpp.WriteString(juleapi.OutId(e.Id, e.Tok.File))
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
