package models

import (
	"strings"

	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/jule"
	"github.com/jule-lang/jule/pkg/juleapi"
)

// Param is function parameter AST model.
type Param struct {
	Token     lex.Token
	Id        string
	Variadic  bool
	Reference bool
	Type      Type
	Default   Expr
}

// TypeString returns data type string of parameter.
func (p *Param) TypeString() string {
	var ts strings.Builder
	if p.Variadic {
		ts.WriteString(tokens.TRIPLE_DOT)
	}
	if p.Reference {
		ts.WriteString(tokens.AMPER)
	}
	ts.WriteString(p.Type.Kind)
	return ts.String()
}

// OutId returns juleapi.OutId result of param.
func (p *Param) OutId() string {
	return juleapi.AsId(p.Id)
}

func (p Param) String() string {
	var cpp strings.Builder
	cpp.WriteString(p.Prototype())
	if p.Id != "" && !juleapi.IsIgnoreId(p.Id) && p.Id != jule.Anonymous {
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
	if p.Reference {
		cpp.WriteByte('&')
	}
	return cpp.String()
}
