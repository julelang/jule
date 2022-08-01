package models

import (
	"strings"

	"github.com/the-xlang/xxc/lex"
	"github.com/the-xlang/xxc/pkg/xapi"
)

type Tok = lex.Tok
type Toks = []Tok

// Type is type declaration.
type Type struct {
	Pub  bool
	Tok  Tok
	Id   string
	Type DataType
	Desc string
	Used bool
}

func (t Type) String() string {
	var cpp strings.Builder
	cpp.WriteString("typedef ")
	cpp.WriteString(t.Type.String())
	cpp.WriteByte(' ')
	cpp.WriteString(xapi.OutId(t.Id, t.Tok.File))
	cpp.WriteByte(';')
	return cpp.String()
}
