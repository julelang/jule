package models

import (
	"strings"

	"github.com/the-xlang/xxc/pkg/xapi"
)

// Var is variable declaration AST model.
type Var struct {
	Pub       bool
	DefTok    Tok
	IdTok     Tok
	SetterTok Tok
	Id        string
	Type      DataType
	Expr      Expr
	Const     bool
	New       bool
	Tag       any
	ExprTag   any
	Desc      string
	Used      bool
	IsField   bool
}

// OutId returns xapi.OutId result of var.
func (v *Var) OutId() string {
	switch {
	case v.IsField:
		return xapi.AsId(v.Id)
	default:
		return xapi.OutId(v.Id, v.IdTok.File)
	}
}

func (v Var) String() string {
	if v.Const {
		return ""
	}
	var cpp strings.Builder
	cpp.WriteString(v.Type.String())
	cpp.WriteByte(' ')
	cpp.WriteString(v.OutId())
	expr := v.Expr.String()
	if expr != "" {
		cpp.WriteString(" = ")
		cpp.WriteString(v.Expr.String())
	} else {
		cpp.WriteString(xapi.DefaultExpr)
	}
	cpp.WriteByte(';')
	return cpp.String()
}

// FieldString returns variable as cpp struct field.
func (v *Var) FieldString() string {
	var cpp strings.Builder
	if v.Const {
		cpp.WriteString("const ")
	}
	cpp.WriteString(v.Type.String())
	cpp.WriteByte(' ')
	cpp.WriteString(v.OutId())
	cpp.WriteString(xapi.DefaultExpr)
	cpp.WriteByte(';')
	return cpp.String()
}
