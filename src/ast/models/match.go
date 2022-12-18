package models

import (
	"strconv"
	"strings"

	"github.com/julelang/jule/lex"
)

type Fallthrough struct {
	Token lex.Token
	Case  *Case
}

// Case the AST model of case.
type Case struct {
	Token lex.Token
	Exprs []Expr
	Block *Block
	Match *Match
	Next  *Case
}

// BeginLabel returns of cpp goto label identifier of case begin.
func (c *Case) BeginLabel() string {
	var cpp strings.Builder
	cpp.WriteString("case_begin_")
	cpp.WriteString(strconv.Itoa(c.Token.Row))
	cpp.WriteString(strconv.Itoa(c.Token.Column))
	return cpp.String()
}

// EndLabel returns of cpp goto label identifier of case end.
func (c *Case) EndLabel() string {
	var cpp strings.Builder
	cpp.WriteString("case_end_")
	cpp.WriteString(strconv.Itoa(c.Token.Row))
	cpp.WriteString(strconv.Itoa(c.Token.Column))
	return cpp.String()
}

// Match the AST model of match-case.
type Match struct {
	Token    lex.Token
	Expr     Expr
	ExprType Type
	Default  *Case
	Cases    []Case
}

// EndLabel returns of cpp goto label identifier of end.
func (m *Match) EndLabel() string {
	var cpp strings.Builder
	cpp.WriteString("match_end_")
	cpp.WriteString(strconv.FormatInt(int64(m.Token.Row), 10))
	cpp.WriteString(strconv.FormatInt(int64(m.Token.Column), 10))
	return cpp.String()
}
