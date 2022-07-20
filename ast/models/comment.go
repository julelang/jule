package models

import "strings"

// Comment is the AST model of just comment lines.
type Comment struct{ Content string }

func (c Comment) String() string {
	var cxx strings.Builder
	cxx.WriteString("// ")
	cxx.WriteString(c.Content)
	return cxx.String()
}
