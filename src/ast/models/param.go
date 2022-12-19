package models

import (
	"strings"

	"github.com/julelang/jule"
	"github.com/julelang/jule/lex"
)

// Param is function parameter AST model.
type Param struct {
	Token    lex.Token
	Id       string
	Variadic bool
	Mutable  bool
	Type     Type
	Default  Expr
}

// TypeString returns data type string of parameter.
func (p *Param) TypeString() string {
	var ts strings.Builder
	if p.Mutable {
		ts.WriteString(lex.KND_MUT + " ")
	}
	if p.Variadic {
		ts.WriteString(lex.KND_TRIPLE_DOT)
	}
	ts.WriteString(p.Type.Kind)
	return ts.String()
}

// OutId returns juleapi.OutId result of param.
func (p *Param) OutId() string {
	return as_local_id(p.Token.Row, p.Token.Column, p.Id)
}

func (p Param) String() string {
	var cpp strings.Builder
	cpp.WriteString(p.Prototype())
	if p.Id != "" && !lex.IsIgnoreId(p.Id) && p.Id != jule.ANONYMOUS {
		cpp.WriteByte(' ')
		cpp.WriteString(p.OutId())
	}
	return cpp.String()
}

// Prototype returns prototype cpp of parameter.
func (p *Param) Prototype() string {
	var cpp strings.Builder
	if p.Variadic {
		cpp.WriteString("slice<")
		cpp.WriteString(p.Type.String())
		cpp.WriteByte('>')
	} else {
		cpp.WriteString(p.Type.String())
	}
	return cpp.String()
}
