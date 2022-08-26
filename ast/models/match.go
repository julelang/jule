package models

import (
	"strconv"
	"strings"

	"github.com/jule-lang/jule/lex"
)

type Fallthrough struct {
	Token lex.Token
	Case  *Case
}

func (f Fallthrough) String() string {
	var cpp strings.Builder
	cpp.WriteString("goto ")
	cpp.WriteString(f.Case.Next.BeginLabel())
	cpp.WriteByte(';')
	return cpp.String()
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

func (c *Case) String(matchExpr string) string {
	endlabel := c.EndLabel()
	var cpp strings.Builder
	if len(c.Exprs) > 0 {
		cpp.WriteString("if (!(")
		for i, expr := range c.Exprs {
			cpp.WriteString(expr.String())
			if matchExpr != "" {
				cpp.WriteString(" == ")
				cpp.WriteString(matchExpr)
			}
			if i+1 < len(c.Exprs) {
				cpp.WriteString(" || ")
			}
		}
		cpp.WriteString(")) { goto ")
		cpp.WriteString(endlabel)
		cpp.WriteString("; }\n")
	}
	if len(c.Block.Tree) > 0 {
		cpp.WriteString(IndentString())
		cpp.WriteString(c.BeginLabel())
		cpp.WriteString(":;\n")
		cpp.WriteString(IndentString())
		cpp.WriteString(c.Block.String())
		cpp.WriteByte('\n')
		cpp.WriteString(IndentString())
		cpp.WriteString("goto ")
		cpp.WriteString(c.Match.EndLabel())
		cpp.WriteString(";")
		cpp.WriteByte('\n')
	}
	cpp.WriteString(IndentString())
	cpp.WriteString(endlabel)
	cpp.WriteString(":;")
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

func (m *Match) MatchExprString() string {
	if len(m.Cases) == 0 {
		if m.Default != nil {
			return m.Default.String("")
		}
		return ""
	}
	var cpp strings.Builder
	cpp.WriteString("{\n")
	AddIndent()
	cpp.WriteString(IndentString())
	cpp.WriteString(m.ExprType.String())
	cpp.WriteString(" expr{")
	cpp.WriteString(m.Expr.String())
	cpp.WriteString("};\n")
	cpp.WriteString(IndentString())
	if len(m.Cases) > 0 {
		cpp.WriteString(m.Cases[0].String("expr"))
		for _, c := range m.Cases[1:] {
			cpp.WriteByte('\n')
			cpp.WriteString(IndentString())
			cpp.WriteString(c.String("expr"))
		}
	}
	if m.Default != nil {
		cpp.WriteString(m.Default.String(""))
	}
	cpp.WriteByte('\n')
	DoneIndent()
	cpp.WriteString(IndentString())
	cpp.WriteByte('}')
	return cpp.String()
}

func (m *Match) MatchBoolString() string {
	var cpp strings.Builder
	if len(m.Cases) > 0 {
		cpp.WriteString(m.Cases[0].String(""))
		for _, c := range m.Cases[1:] {
			cpp.WriteByte('\n')
			cpp.WriteString(IndentString())
			cpp.WriteString(c.String(""))
		}
	}
	if m.Default != nil {
		cpp.WriteByte('\n')
		cpp.WriteString(m.Default.String(""))
		cpp.WriteByte('\n')
	}
	return cpp.String()
}

// EndLabel returns of cpp goto label identifier of end.
func (m *Match) EndLabel() string {
	var cpp strings.Builder
	cpp.WriteString("match_end_")
	cpp.WriteString(strconv.FormatInt(int64(m.Token.Row), 10))
	cpp.WriteString(strconv.FormatInt(int64(m.Token.Column), 10))
	return cpp.String()
}

func (m Match) String() string {
	var cpp strings.Builder
	if m.Expr.Model != nil {
		cpp.WriteString(m.MatchExprString())
	} else {
		cpp.WriteString(m.MatchBoolString())
	}
	cpp.WriteByte('\n')
	cpp.WriteString(IndentString())
	cpp.WriteString(m.EndLabel())
	cpp.WriteString(":;")
	return cpp.String()
}
