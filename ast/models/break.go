package models

import "strings"

// Break is the AST model of break statement.
type Break struct {
	Tok  Tok
	Case *Case
}

func (b Break) String() string {
	if b.Case != nil {
		var cpp strings.Builder
		cpp.WriteString("goto ")
		cpp.WriteString(b.Case.Match.EndLabel())
		return cpp.String()
	}
	return "break;"
}
