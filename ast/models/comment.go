package models

import "strings"

// Comment is the AST model of just comment lines.
type Comment struct{ Content string }

func (c Comment) String() string {
	var cpp strings.Builder
	cpp.WriteString("// ")
	cpp.WriteString(c.Content)
	return cpp.String()
}
