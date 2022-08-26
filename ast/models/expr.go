package models

import (
	"strings"

	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/juleapi"
)

// Expr is AST model of expression.
type Expr struct {
	Tokens      []lex.Token
	Processes [][]lex.Token
	Model     IExprModel
}

func (e Expr) String() string {
	if e.Model != nil {
		return e.Model.String()
	}
	var expr strings.Builder
	for _, process := range e.Processes {
		for _, t := range process {
			switch t.Id {
			case tokens.Id:
				expr.WriteString(juleapi.OutId(t.Kind, t.File))
			default:
				expr.WriteString(t.Kind)
			}
		}
	}
	return expr.String()
}
