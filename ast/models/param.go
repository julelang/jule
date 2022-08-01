package models

import (
	"strings"

	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xapi"
)

// Param is function parameter AST model.
type Param struct {
	Tok       Tok
	Id        string
	Variadic  bool
	Reference bool
	Type      DataType
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

// OutId returns xapi.OutId result of param.
func (p *Param) OutId() string {
	return xapi.OutId(p.Id, p.Tok.File)
}

func (p Param) String() string {
	var cpp strings.Builder
	cpp.WriteString(p.Prototype())
	if p.Id != "" && !xapi.IsIgnoreId(p.Id) && p.Id != x.Anonymous {
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
