package models

import "strings"

// IterFor is while iteration profile.
type IterFor struct {
	Once      Statement
	Condition Expr
	Next      Statement
}

func (f IterFor) String(iter Iter) string {
	var cpp strings.Builder
	cpp.WriteString("for (")
	if f.Once.Data != nil {
		cpp.WriteString(f.Once.String())
	} else {
		cpp.WriteString("; ")
	}
	cpp.WriteString(f.Condition.String())
	cpp.WriteString("; ")
	if f.Next.Data != nil {
		s := f.Next.String()
		// Remove statement terminator
		cpp.WriteString(s[:len(s)-1])
	}
	cpp.WriteString(") ")
	cpp.WriteString(iter.Block.String())
	return cpp.String()
}
