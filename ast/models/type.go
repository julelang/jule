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
	var cxx strings.Builder
	cxx.WriteString("typedef ")
	cxx.WriteString(t.Type.String())
	cxx.WriteByte(' ')
	cxx.WriteString(xapi.OutId(t.Id, t.Tok.File))
	cxx.WriteByte(';')
	return cxx.String()
}
