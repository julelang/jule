package models

import (
	"strings"

	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/pkg/juleapi"
)

// TypeAlias is type alias declaration.
type TypeAlias struct {
	Owner   *Block
	Pub     bool
	Token   lex.Token
	Id      string
	Type    Type
	Doc     string
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
		cpp.WriteString(juleapi.OutId(t.Id, t.Token.File.Addr()))
	}
	cpp.WriteByte(';')
	return cpp.String()
}
