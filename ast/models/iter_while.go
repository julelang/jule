package models

import "strings"

// IterWhile is while iteration profile.
type IterWhile struct {
	Expr Expr
}

func (w IterWhile) String(i *Iter) string {
	var cpp strings.Builder
	indent := IndentString()
	end := i.EndLabel()
	cpp.WriteString(i.NextLabel())
	cpp.WriteString(":;\n")
	cpp.WriteString(indent)
	cpp.WriteString("if (")
	cpp.WriteString(w.Expr.String())
	cpp.WriteString(") ")
	cpp.WriteString(i.Block.String())
	cpp.WriteByte('\n')
	cpp.WriteString(indent)
	cpp.WriteString(end)
	cpp.WriteString(":;")
	return cpp.String()
}
