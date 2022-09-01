package models

import (
	"strconv"
	"strings"

	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/juleapi"
)

// Var is variable declaration AST model.
type Var struct {
	Owner     *Block
	Pub       bool
	Mutable   bool
	Token     lex.Token
	SetterTok lex.Token
	Id        string
	Type      Type
	Expr      Expr
	Const     bool
	New       bool
	Tag       any
	ExprTag   any
	Desc      string
	Used      bool
	IsField   bool
}

func (v *Var) IsLocal() bool { return v.Owner != nil }

func as_local_id(row, column int, id string) string {
	id = strconv.Itoa(row) + strconv.Itoa(column) + "_" + id
	return juleapi.AsId(id)
}

// OutId returns juleapi.OutId result of var.
func (v *Var) OutId() string {
	switch {
	case v.Id == tokens.SELF:
		return "self"
	case v.IsLocal():
		return as_local_id(v.Token.Row, v.Token.Column, v.Id)
	case v.IsField:
		return "__julec_field_" + juleapi.AsId(v.Id)
	default:
		return juleapi.OutId(v.Id, v.Token.File)
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
		cpp.WriteString(juleapi.DefaultExpr)
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
	cpp.WriteString(juleapi.DefaultExpr)
	cpp.WriteByte(';')
	return cpp.String()
}
