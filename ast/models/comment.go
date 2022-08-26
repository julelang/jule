package models

import (
	"strings"

	"github.com/jule-lang/jule/lex"
)

// Comment is the AST model of just comment lines.
type Comment struct {
	Token   lex.Token
	Content string
}

func (c Comment) String() string {
	var cpp strings.Builder
	cpp.WriteString("// ")
	cpp.WriteString(c.Content)
	return cpp.String()
}
