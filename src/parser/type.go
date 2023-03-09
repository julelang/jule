package parser

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

func build_primitive_type(kind string) *ast.Type {
	return &ast.Type{
		Token: lex.Token{
			Id:   lex.ID_DT,
			Kind: kind,
		},
	}
}
func build_void_type() *ast.Type { return &ast.Type{} }
func build_u32_type() *ast.Type { return build_primitive_type(lex.KND_U32) }

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
	t := &ast.Type{
		Token: tb.tokens[*tb.i],
		Kind:  nil,
	}
	*tb.i++
	tb.finished = true
	return t
}

func (tb *type_builder) step() *ast.Type {
	token := tb.tokens[*tb.i]
	switch token.Id {
	case lex.ID_DT:
		return tb.build_primitive()
	// TODO: implement other types
	default:
		*tb.i++
		tb.push_err(token, "invalid_syntax")
		return nil
	}
}

// Builds type.
// Returns void if error occurs.
func (tb *type_builder) build() (*ast.Type, bool) {
	root := tb.step()
	if root == nil {
		return build_void_type(), false
	}
	if tb.finished {
		return root, true
	}

	node := &root.Kind
	tb.first = *tb.i
	for ; *tb.i < len(tb.tokens); *tb.i++ {
		*node = tb.step()
		if *node == nil {
			return build_void_type(), false
		} else if tb.finished {
			break
		}
	}
	return root, true
}
