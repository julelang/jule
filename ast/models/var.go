package models

import (
	"strconv"
	"strings"

	"github.com/jule-lang/jule/lex"
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
	CppLinked bool
}

//IsLocal returns variable is into the scope or not.
func (v *Var) IsLocal() bool { return v.Owner != nil }

func as_local_id(row, column int, id string) string {
	id = strconv.Itoa(row) + strconv.Itoa(column) + "_" + id
	return juleapi.AsId(id)
}

// OutId returns juleapi.OutId result of var.
func (v *Var) OutId() string {
	switch {
	case v.CppLinked:
		return v.Id
	case v.Id == lex.KND_SELF:
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
	if juleapi.IsIgnoreId(v.Id) {
		return ""
	}
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
		cpp.WriteString(juleapi.DEFAULT_EXPR)
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
	cpp.WriteString(juleapi.DEFAULT_EXPR)
	cpp.WriteByte(';')
	return cpp.String()
}

// ReeiverTypeString returns receiver declaration string.
func (v *Var) ReceiverTypeString() string {
	var s strings.Builder
	if v.Mutable {
		s.WriteString("mut ")
	}
	if v.Type.Kind != "" && v.Type.Kind[0] == '&' {
		s.WriteByte('&')
	}
	s.WriteString("self")
	return s.String()
}
