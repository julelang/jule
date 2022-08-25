package models

import "strings"

// IterFor is while iteration profile.
type IterFor struct {
	Once      Statement
	Condition Expr
	Next      Statement
}

func (f IterFor) String(i *Iter) string {
	var cpp strings.Builder
	var indent string
	if f.Once.Data != nil {
		cpp.WriteString("{\n")
		AddIndent()
		indent = IndentString()
		cpp.WriteString(indent)
		cpp.WriteString(f.Once.String())
		cpp.WriteByte('\n')
		cpp.WriteString(indent)
	} else {
		indent = IndentString()
	}
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
	cpp.WriteString(f.Next.String())
	cpp.WriteByte('\n')
	cpp.WriteString(indent)
	cpp.WriteString("if (")
	cpp.WriteString(f.Condition.String())
	cpp.WriteString(") { goto ")
	cpp.WriteString(begin)
	cpp.WriteString("; }\n")
	cpp.WriteString(indent)
	cpp.WriteString(i.EndLabel())
	cpp.WriteString(":;")
	if f.Once.Data != nil {
		cpp.WriteByte('\n')
		DoneIndent()
		cpp.WriteString(IndentString())
		cpp.WriteByte('}')
	}
	return cpp.String()
}
