package models

import (
	"strings"

	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/pkg/juleapi"
)

type Tok = lex.Tok
type Toks = []Tok

// Type is type declaration.
type Type struct {
	Pub     bool
	Tok     Tok
	Id      string
	Type    DataType
	Desc    string
	Used    bool
	Generic bool
}

func (t Type) String() string {
	var cpp strings.Builder
	cpp.WriteString("typedef ")
	cpp.WriteString(t.Type.String())
	cpp.WriteByte(' ')
	if t.Generic {
		cpp.WriteString(juleapi.AsId(t.Id))
	} else {
		cpp.WriteString(juleapi.OutId(t.Id, t.Tok.File))
	}
	cpp.WriteByte(';')
	return cpp.String()
}
