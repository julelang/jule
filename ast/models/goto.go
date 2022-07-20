package models

import "strings"

// Labels is label slice type.
type Labels []*Label

// Gotos is goto slice type.
type Gotos []*Goto

// Label is the AST model of labels.
type Label struct {
	Tok   Tok
	Label string
	Index int
	Used  bool
	Block *Block
}

func (l Label) String() string {
	return l.Label + ":;"
}

// Goto is the AST model of goto statements.
type Goto struct {
	Tok   Tok
	Label string
	Index int
	Block *Block
}

func (gt Goto) String() string {
	var cxx strings.Builder
	cxx.WriteString("goto ")
	cxx.WriteString(gt.Label)
	cxx.WriteByte(';')
	return cxx.String()
}
