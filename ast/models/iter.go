package models

import (
	"strconv"
	"strings"
)

// Iter is the AST model of iterations.
type Iter struct {
	Tok     Tok
	Block   *Block
	Profile IterProfile
}

// BeginLabel returns of cpp goto label identifier of iteration begin.
func (i *Iter) BeginLabel() string {
	var cpp strings.Builder
	cpp.WriteString("iter_begin_")
	cpp.WriteString(strconv.Itoa(i.Tok.Row))
	cpp.WriteString(strconv.Itoa(i.Tok.Column))
	return cpp.String()
}

// EndLabel returns of cpp goto label identifier of iteration end.
// Used for "break" keword by default.
func (i *Iter) EndLabel() string {
	var cpp strings.Builder
	cpp.WriteString("iter_end_")
	cpp.WriteString(strconv.Itoa(i.Tok.Row))
	cpp.WriteString(strconv.Itoa(i.Tok.Column))
	return cpp.String()
}

// NextLabel returns of cpp goto label identifier of iteration next point.
// Used for "continue" keyword by default.
func (i *Iter) NextLabel() string {
	var cpp strings.Builder
	cpp.WriteString("iter_next_")
	cpp.WriteString(strconv.Itoa(i.Tok.Row))
	cpp.WriteString(strconv.Itoa(i.Tok.Column))
	return cpp.String()
}

func (i *Iter) infinityString() string {
	var cpp strings.Builder
	indent := IndentString()
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
	cpp.WriteString("goto ")
	cpp.WriteString(begin)
	cpp.WriteString(";\n")
	cpp.WriteString(indent)
	cpp.WriteString(i.EndLabel())
	cpp.WriteString(":;")
	return cpp.String()
}

func (i Iter) String() string {
	if i.Profile == nil {
		return i.infinityString()
	}
	return i.Profile.String(&i)
}
