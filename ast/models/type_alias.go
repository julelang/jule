package models

import (
	"strings"

	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/pkg/juleapi"
)

// TypeAlias is type alias declaration.
type TypeAlias struct {
	Pub     bool
	Token   lex.Token
	Id      string
	Type    Type
	Desc    string
	Used    bool
	Generic bool
}

func (t TypeAlias) String() string {
	var cpp strings.Builder
	cpp.WriteString("typedef ")
	cpp.WriteString(t.Type.String())
	cpp.WriteByte(' ')
	if t.Generic {
		cpp.WriteString(juleapi.AsId(t.Id))
	} else {
		cpp.WriteString(juleapi.OutId(t.Id, t.Token.File))
	}
	cpp.WriteByte(';')
	return cpp.String()
}
