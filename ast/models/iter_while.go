package models

import "strings"

// IterWhile is while iteration profile.
type IterWhile struct {
	Expr Expr
}

func (w IterWhile) String(iter Iter) string {
	var cxx strings.Builder
	cxx.WriteString("while (")
	cxx.WriteString(w.Expr.String())
	cxx.WriteString(") ")
	cxx.WriteString(iter.Block.String())
	return cxx.String()
}
