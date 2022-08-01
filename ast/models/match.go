package models

import "strings"

// Case the AST model of case.
type Case struct {
	Exprs []Expr
	Block *Block
}

func (c *Case) String(matchExpr string) string {
	var cpp strings.Builder
	if len(c.Exprs) > 0 {
		cpp.WriteString("if (")
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
		cpp.WriteByte(')')
	}
	cpp.WriteString(" { do ")
	cpp.WriteString(c.Block.String())
	cpp.WriteString("while(false);")
	if len(c.Exprs) > 0 {
		cpp.WriteByte('}')
	}
	return cpp.String()
}

// Match the AST model of match-case.
type Match struct {
	Tok      Tok
	Expr     Expr
	ExprType DataType
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
	cpp.WriteString(m.Cases[0].String("expr"))
	for _, c := range m.Cases[1:] {
		cpp.WriteString("else ")
		cpp.WriteString(c.String("expr"))
	}
	if m.Default != nil {
		cpp.WriteString("else ")
		cpp.WriteString(m.Default.String(""))
		cpp.WriteByte('}')
	}
	cpp.WriteByte('\n')
	DoneIndent()
	cpp.WriteString(IndentString())
	cpp.WriteByte('}')
	return cpp.String()
}

func (m *Match) MatchBoolString() string {
	var cpp strings.Builder
	cpp.WriteString(m.Cases[0].String(""))
	for _, c := range m.Cases[1:] {
		cpp.WriteString("else ")
		cpp.WriteString(c.String(""))
	}
	if m.Default != nil {
		cpp.WriteString("else ")
		cpp.WriteString(m.Default.String(""))
		cpp.WriteByte('}')
	}
	return cpp.String()
}

func (m Match) String() string {
	if m.Expr.Model != nil {
		return m.MatchExprString()
	}
	return m.MatchBoolString()
}
