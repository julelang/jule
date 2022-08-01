package models

import "strings"

// Else is the AST model of else blocks.
type Else struct {
	Tok   Tok
	Block *Block
}

func (elseast Else) String() string {
	var cpp strings.Builder
	cpp.WriteString("else ")
	cpp.WriteString(elseast.Block.String())
	return cpp.String()
}
