package parser

import (
	"strings"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
)

type variable struct {
	Name  string
	Token lex.Token
	Value ast.ExpressionAST
	Type  uint8
}

func (v variable) String() string {
	var sb strings.Builder
	sb.WriteString(x.CxxTypeNameFromType(v.Type))
	sb.WriteByte(' ')
	sb.WriteString(v.Name)
	sb.WriteByte('=')
	sb.WriteString(v.Value.String())
	sb.WriteByte(';')
	return sb.String()
}
