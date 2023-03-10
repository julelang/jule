package parser

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// Returns function expressions without call expression
// if tokens are function call, nil if not.
func is_fn_call(tokens []lex.Token) []lex.Token {
	switch tokens[0].Id {
	case lex.ID_RANGE, lex.ID_IDENT, lex.ID_DT:
		// Ignore.
	default:
		tok := tokens[len(tokens)-1]
		if tok.Id != lex.ID_RANGE && tok.Kind != lex.KND_RPARENT {
			return nil
		}
	}
	tok := tokens[len(tokens)-1]
	if tok.Id != lex.ID_RANGE || tok.Kind != lex.KND_RPARENT {
		return nil
	}
	brace_n := 0
	// Loops i >= 1 because expression must be has function expression at begin.
	// For this reason, ignore first token.
	for i := len(tokens) - 1; i >= 1; i-- {
		tok := tokens[i]
		if tok.Id == lex.ID_RANGE {
			switch tok.Kind {
			case lex.KND_RPARENT:
				brace_n++
			case lex.KND_LPAREN:
				brace_n--
			}
			if brace_n == 0 {
				return tokens[:i]
			}
		}
	}
	return nil
}

type call_data struct {
	expr_tokens     []lex.Token
	args_tokens     []lex.Token
	generics_tokens []lex.Token
}

func get_call_data(tokens []lex.Token) *call_data {
	data := &call_data{}
	data.expr_tokens, data.args_tokens = lex.RangeLast(tokens)
	if len(data.expr_tokens) == 0 {
		return data
	}

	// Below is call expression
	token := data.expr_tokens[len(data.expr_tokens)-1]
	if token.Id == lex.ID_RANGE && token.Kind == lex.KND_RBRACKET {
		data.expr_tokens, data.generics_tokens = lex.RangeLast(data.expr_tokens)
	}
	return data
}

// Returns expression tokens comes before block if exist, nil if not.
func get_block_expr(tokens []lex.Token) []lex.Token {
	brace_n := 0
	for i, tok := range tokens {
		if tok.Id == lex.ID_RANGE {
			switch tok.Kind {
			case lex.KND_LBRACE:
				if brace_n > 0 {
					brace_n++
					break
				}
				return tokens[:i]
			case lex.KND_LBRACKET, lex.KND_LPAREN:
				brace_n++
			default:
				brace_n--
			}
		}
	}
	return nil
}

// Returns colon index and range tokens.
// Returns nil slice and -1 if not found.
// Starts search at *i.
func split_colon(tokens []lex.Token, i *int) (range_tokens []lex.Token, colon int) {
	colon = -1
	range_n := 0
	start := *i
	for ; *i < len(tokens); *i++ {
		token := tokens[*i]
		if token.Id == lex.ID_RANGE {
			switch token.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				range_n++
				continue
			default:
				range_n--
			}
		}
		if range_n == 0 {
			if start+1 > *i {
				return
			}
			range_tokens = tokens[start+1 : *i]
			break
		} else if range_n != 1 {
			continue
		}
		if colon == -1 && token.Id == lex.ID_COLON {
			colon = *i - start - 1
		}
	}
	return
}

type precedencer struct {
	pairs [][]any
}

func (p *precedencer) set(level int, expr any) {
	for i, pair := range p.pairs {
		pair_level := pair[0].(int)
		if level > pair_level {
			first := p.pairs[:i]
			appended := append([][]any{{level, expr}}, p.pairs[i:]...)
			p.pairs = append(first, appended...)
			return
		}
	}
	p.pairs = append(p.pairs, []any{level, expr})
}

func (p *precedencer) get_lower() any {
	for i := len(p.pairs) - 1; i >= 0; i-- {
		data := p.pairs[i][1]
		if data != nil {
			return data
		}
	}
	return nil
}

func eliminate_comments(tokens []lex.Token) []lex.Token {
	cutted := []lex.Token{}
	for _, token := range tokens {
		if token.Id != lex.ID_COMMENT {
			cutted = append(cutted, token)
		}
	}
	return cutted
}

// Finds index of priority operator and returns index of operator
// if found, returns -1 if not.
func find_lowest_prec_op(tokens []lex.Token) int {
	prec := precedencer{}
	brace_n := 0
	for i, token := range tokens {
		switch {
		case token.Id == lex.ID_RANGE:
			switch token.Kind {
			case lex.KND_LBRACE, lex.KND_LPAREN, lex.KND_LBRACKET:
				brace_n++
			default:
				brace_n--
			}
			continue
		case i == 0:
			continue
		case token.Id != lex.ID_OP:
			continue
		case brace_n > 0:
			continue
		}
		// Skip unary operator.
		if tokens[i-1].Id == lex.ID_OP {
			continue
		}
		p := token.Prec()
		if p != -1 {
			prec.set(p, i)
		}
	}
	data := prec.get_lower()
	if data == nil {
		return -1
	}
	return data.(int)
}

func build_ident_expr(token lex.Token) *ast.IdentExpr {
	return &ast.IdentExpr{
		Token:     token,
		Ident:     token.Kind,
		CppLinked: false,
	}
}

type expr_builder struct {
	errors []build.Log
}

func (ep *expr_builder) push_err(token lex.Token, key string, args ...any) {
	ep.errors = append(ep.errors, compiler_err(token, key, args...))
}

func (ep *expr_builder) build_tuple(parts [][]lex.Token) *ast.TupleExpr {
	tuple := &ast.TupleExpr{
		Expr: make([]ast.ExprData, len(parts)),
	}
	for i, part := range parts {
		tuple.Expr[i] = ep.build(part)
	}
	return tuple
}

func (ep *expr_builder) build_lit(token lex.Token) *ast.LitExpr {
	return &ast.LitExpr{
		Token: token,
		Value: token.Kind,
	}
}

func (ep *expr_builder) build_primitive_type(token lex.Token) *ast.TypeExpr {
	return &ast.TypeExpr{Token: token}
}

func (ep *expr_builder) build_single(token lex.Token) ast.ExprData {
	switch token.Id {
	case lex.ID_LIT:
		return ep.build_lit(token)

	case lex.ID_IDENT, lex.ID_SELF:
		return build_ident_expr(token)

	case lex.ID_DT:
		return ep.build_primitive_type(token)

	default:
		ep.push_err(token, "invalid_syntax")
		return nil
	}
}

func (ep *expr_builder) build_cpp_linked_ident(tokens []lex.Token) *ast.IdentExpr {
	if tokens[0].Id != lex.ID_CPP {
		return nil
	} else if tokens[1].Id != lex.ID_DOT {
		ep.push_err(tokens[1], "invalid_syntax")
		return nil
	}
	token := tokens[2]
	if token.Id != lex.ID_IDENT {
		ep.push_err(tokens[2], "invalid_syntax")
		return nil
	}
	expr := build_ident_expr(token)
	expr.CppLinked = true
	return expr
}

func (ep *expr_builder) build_unary(tokens []lex.Token) *ast.UnaryExpr {
	op := tokens[0]
	if len(tokens) == 1 {
		ep.push_err(op, "missing_expr_for_unary")
		return nil
	} else if !lex.IsUnaryOp(op.Kind) {
		ep.push_err(op, "invalid_op_for_unary", op.Kind)
		return nil
	}

	// Length is 1 cause all length of operator tokens is 1.
	// Change "1" with length of token's value
	// if all operators length is not 1.
	tokens = tokens[1:]

	return &ast.UnaryExpr{
		Op:   op,
		Expr: ep.build(tokens),
	}
}

func (ep *expr_builder) build_obj_sub_ident(tokens []lex.Token) *ast.SubIdentExpr {
	i := len(tokens) - 1
	ident_token := tokens[i]
	i-- // Set offset to delimiter token.
	tokens = tokens[:i] // Remove dot token and selected identifier token.
	if len(tokens) == 0 {
		ep.push_err(ident_token, "invalid_syntax")
		return nil
	}
	return &ast.SubIdentExpr{
		Ident: ident_token,
		Expr:  ep.build(tokens),
	}
}

func (ep *expr_builder) build_ns_sub_ident(tokens []lex.Token) *ast.NsSelectionExpr {
	ns := &ast.NsSelectionExpr{}
	for i, token := range tokens {
		if i%2 == 0 {
			if token.Id != lex.ID_IDENT {
				ep.push_err(token, "invalid_syntax")
			}
			ns.Ns = append(ns.Ns, token)
		} else if token.Id != lex.ID_DBLCOLON {
			ep.push_err(token, "invalid_syntax")
		}
	}
	ns.Ident = ns.Ns[len(ns.Ns)-1]
	ns.Ns = ns.Ns[:len(ns.Ns)-1]
	return ns
}

func (ep *expr_builder) build_sub_ident(tokens []lex.Token) ast.ExprData {
	i := len(tokens) - 1
	i-- // Set offset to delimiter token.
	token := tokens[i]
	switch token.Id {
	case lex.ID_DOT:
		return ep.build_obj_sub_ident(tokens)
	case lex.ID_DBLCOLON:
		return ep.build_ns_sub_ident(tokens)
	default:
		ep.push_err(token, "invalid_syntax")
		return nil
	}
}

func (ep *expr_builder) build_data(tokens []lex.Token) ast.ExprData {
	switch len(tokens) {
	case 1:
		return ep.build_single(tokens[0])
	case 3:
		if tokens[0].Id == lex.ID_CPP {
			return ep.build_cpp_linked_ident(tokens)
		}
	}

	token := tokens[0]
	switch token.Id {
	case lex.ID_OP:
		return ep.build_unary(tokens)
	}

	token = tokens[len(tokens)-1]
	switch token.Id {
	case lex.ID_IDENT:
		return ep.build_sub_ident(tokens)
	// TODO: implement other nodes
	}

	ep.push_err(tokens[0], "invalid_syntax")
	return nil
}

func (ep *expr_builder) build_binop(tokens []lex.Token, i int) *ast.BinopExpr {
	return &ast.BinopExpr{
		L:  ep.build(tokens[:i]),
		R:  ep.build(tokens[i+1:]),
		Op: tokens[i],
	}
}

func (ep *expr_builder) build(tokens []lex.Token) ast.ExprData {
	i := find_lowest_prec_op(tokens)
	if i == -1 {
		return ep.build_data(tokens)
	}
	return ep.build_binop(tokens, i)
}

func (ep *expr_builder) build_kind(tokens []lex.Token) ast.ExprData {
	parts, errors := lex.Parts(tokens, lex.ID_COMMA, true)
	if errors != nil {
		ep.errors = append(ep.errors, errors...)
		return nil
	} else if len(parts) > 1 {
		return ep.build_tuple(parts)
	}
	return ep.build(tokens)
}

func (ep *expr_builder) build_from_tokens(tokens []lex.Token) *ast.Expr {
	tokens = eliminate_comments(tokens)
	if len(tokens) == 0 {
		return nil
	}
	return &ast.Expr{
		Token: tokens[0],
		Kind:  ep.build_kind(tokens),
	}
}
