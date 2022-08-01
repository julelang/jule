package models

import "strings"

// IterWhile is while iteration profile.
type IterWhile struct {
	Expr Expr
}

func (w IterWhile) String(iter Iter) string {
	var cpp strings.Builder
	cpp.WriteString("while (")
	cpp.WriteString(w.Expr.String())
	cpp.WriteString(") ")
	cpp.WriteString(iter.Block.String())
	return cpp.String()
}
