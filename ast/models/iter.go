package models

import "strings"

// Iter is the AST model of iterations.
type Iter struct {
	Tok     Tok
	Block   *Block
	Profile IterProfile
}

func (iter Iter) String() string {
	if iter.Profile == nil {
		var cpp strings.Builder
		cpp.WriteString("while (true) ")
		cpp.WriteString(iter.Block.String())
		return cpp.String()
	}
	return iter.Profile.String(iter)
}
