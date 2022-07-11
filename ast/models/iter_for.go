package models

import "strings"

// IterFor is while iteration profile.
type IterFor struct {
	Once      Statement
	Condition Expr
	Next      Statement
}

func (f IterFor) String(iter Iter) string {
	var cxx strings.Builder
	cxx.WriteString("for (")
	if f.Once.Val != nil {
		cxx.WriteString(f.Once.String())
	} else {
		cxx.WriteString("; ")
	}
	cxx.WriteString(f.Condition.String())
	cxx.WriteString("; ")
	if f.Next.Val != nil {
		s := f.Next.String()
		// Remove statement terminator
		cxx.WriteString(s[:len(s)-1])
	}
	cxx.WriteString(") ")
	cxx.WriteString(iter.Block.String())
	return cxx.String()
}
