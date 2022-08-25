package models

import "strings"

// IterWhile is while iteration profile.
type IterWhile struct {
	Expr Expr
}

func (w IterWhile) String(i *Iter) string {
	var cpp strings.Builder
	indent := IndentString()
	begin := i.BeginLabel()
	cpp.WriteString(begin)
	cpp.WriteString(":;\n")
	cpp.WriteString(indent)
	cpp.WriteString(i.Block.String())
	cpp.WriteByte('\n')
	cpp.WriteString(indent)
	cpp.WriteString(i.NextLabel())
	cpp.WriteString(":;\n")
	cpp.WriteString(indent)
	cpp.WriteString("if (")
	cpp.WriteString(w.Expr.String())
	cpp.WriteString(") { goto ")
	cpp.WriteString(begin)
	cpp.WriteString("; }\n")
	cpp.WriteString(indent)
	cpp.WriteString(i.EndLabel())
	cpp.WriteString(":;")
	return cpp.String()
}
