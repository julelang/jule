package models

import (
	"strconv"
	"strings"

	"github.com/julelang/jule/lex"
)

// Break is the AST model of break statement.
type Break struct {
	Token      lex.Token
	LabelToken lex.Token
	Label      string
}

func (b Break) String() string {
	return "goto " + b.Label + ";"
}

// Continue is the AST model of break statement.
type Continue struct {
	Token     lex.Token
	LoopLabel lex.Token
	Label     string
}

func (c Continue) String() string {
	return "goto " + c.Label + ";"
}

// Iter is the AST model of iterations.
type Iter struct {
	Token   lex.Token
	Block   *Block
	Parent  *Block
	Profile any
}

// BeginLabel returns of cpp goto label identifier of iteration begin.
func (i *Iter) BeginLabel() string {
	var cpp strings.Builder
	cpp.WriteString("iter_begin_")
	cpp.WriteString(strconv.Itoa(i.Token.Row))
	cpp.WriteString(strconv.Itoa(i.Token.Column))
	return cpp.String()
}

// EndLabel returns of cpp goto label identifier of iteration end.
// Used for "break" keword by default.
func (i *Iter) EndLabel() string {
	var cpp strings.Builder
	cpp.WriteString("iter_end_")
	cpp.WriteString(strconv.Itoa(i.Token.Row))
	cpp.WriteString(strconv.Itoa(i.Token.Column))
	return cpp.String()
}

// NextLabel returns of cpp goto label identifier of iteration next point.
// Used for "continue" keyword by default.
func (i *Iter) NextLabel() string {
	var cpp strings.Builder
	cpp.WriteString("iter_next_")
	cpp.WriteString(strconv.Itoa(i.Token.Row))
	cpp.WriteString(strconv.Itoa(i.Token.Column))
	return cpp.String()
}
