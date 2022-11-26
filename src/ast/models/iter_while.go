package models

import "strings"

// IterWhile is while iteration profile.
type IterWhile struct {
	Expr Expr
	Next Statement
}

func (w IterWhile) String(i *Iter) string {
	var cpp strings.Builder
	indent := IndentString()
	begin := i.BeginLabel()
	next := i.NextLabel()
	end := i.EndLabel()
	cpp.WriteString(begin)
	cpp.WriteString(":;\n")
	cpp.WriteString(indent)
	if !w.Expr.IsEmpty() {
		cpp.WriteString("if (!(")
		cpp.WriteString(w.Expr.String())
		cpp.WriteString(")) { goto ")
		cpp.WriteString(end)
		cpp.WriteString("; }\n")
		cpp.WriteString(indent)
	}
	cpp.WriteString(i.Block.String())
	cpp.WriteByte('\n')
	cpp.WriteString(indent)
	cpp.WriteString(next)
	cpp.WriteString(":;\n")
	cpp.WriteString(indent)
	if w.Next.Data != nil {
		cpp.WriteString(w.Next.String())
		cpp.WriteByte('\n')
		cpp.WriteString(indent)
	}
	cpp.WriteString("goto ")
	cpp.WriteString(begin)
	cpp.WriteString(";\n")
	cpp.WriteString(indent)
	cpp.WriteString(end)
	cpp.WriteString(":;")
	return cpp.String()
}
