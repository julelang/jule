package models

import (
	"strings"

	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/juleapi"
)

// Expr is AST model of expression.
type Expr struct {
	Toks      []Tok
	Processes [][]Tok
	Model     IExprModel
}

func (e Expr) String() string {
	if e.Model != nil {
		return e.Model.String()
	}
	var expr strings.Builder
	for _, process := range e.Processes {
		for _, tok := range process {
			switch tok.Id {
			case tokens.Id:
				expr.WriteString(juleapi.OutId(tok.Kind, tok.File))
			default:
				expr.WriteString(tok.Kind)
			}
		}
	}
	return expr.String()
}
