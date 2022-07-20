package models

import (
	"strings"

	"github.com/the-xlang/xxc/pkg/xapi"
)

// EnumItem is the AST model of enumerator items.
type EnumItem struct {
	Tok  Tok
	Id   string
	Expr Expr
}

func (ei EnumItem) String() string {
	var cxx strings.Builder
	cxx.WriteString(xapi.OutId(ei.Id, ei.Tok.File))
	cxx.WriteString(" = ")
	cxx.WriteString(ei.Expr.String())
	return cxx.String()
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
	var cxx strings.Builder
	cxx.WriteString("enum ")
	cxx.WriteString(xapi.OutId(e.Id, e.Tok.File))
	cxx.WriteByte(':')
	cxx.WriteString(e.Type.String())
	cxx.WriteString(" {\n")
	AddIndent()
	for _, item := range e.Items {
		cxx.WriteString(IndentString())
		cxx.WriteString(item.String())
		cxx.WriteString(",\n")
	}
	DoneIndent()
	cxx.WriteString("};")
	return cxx.String()
}
