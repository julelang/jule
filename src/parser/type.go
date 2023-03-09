package parser

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

type type_builder struct {
	p        *parser
	tokens   []lex.Token
	i        *int
	err      bool
	first    int
	finished bool
}

func (tb *type_builder) push_err(token lex.Token, key string) {
	if tb.err {
		tb.p.push_err(token, key)
	}
}

func (tb *type_builder) build_primitive() *ast.Type {
	return &ast.Type{
		Token: tb.tokens[*tb.i],
		Kind:  nil,
	}
}

func (tb *type_builder) step() *ast.Type {
	token := tb.tokens[*tb.i]
	switch token.Id {
	case lex.ID_DT:
		return tb.build_primitive()
	// TODO: implement other types
	default:
		tb.push_err(token, "invalid_syntax")
		return nil
	}
}

// Builds type.
// Returns nil if error occurs.
func (tb *type_builder) build() *ast.Type {
	root := tb.step()
	if root == nil {
		return nil
	}

	node := &root.Kind
	tb.first = *tb.i
	for ; *tb.i < len(tb.tokens); *tb.i++ {
		*node = tb.step()
		if *node == nil {
			return nil
		} else if tb.finished {
			break
		}
	}
	return root
}
