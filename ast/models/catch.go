package models

import (
	"strings"

	"github.com/the-xlang/xxc/pkg/xapi"
)

// Catch is the AST model of catch blocks.
type Catch struct {
	Tok   Tok
	Var   Var
	Block Block
}

func (c Catch) String() string {
	var cxx strings.Builder
	cxx.WriteString("catch (")
	if c.Var.Id == "" {
		cxx.WriteString("...")
	} else {
		cxx.WriteString(c.Var.Type.String())
		cxx.WriteByte(' ')
		cxx.WriteString(xapi.OutId(c.Var.Id, c.Tok.File))
	}
	cxx.WriteString(") ")
	cxx.WriteString(c.Block.String())
	return cxx.String()
}
