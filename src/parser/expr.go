// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package parser

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Returns function expressions without call expression
// if tokens are function call, nil if not.
func is_fn_call(tokens []lex.Token) []lex.Token {
	switch tokens[0].Id {
	case lex.ID_RANGE, lex.ID_IDENT, lex.ID_PRIM:
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

type _CallData struct {
	expr_tokens     []lex.Token
	args_tokens     []lex.Token
	generics_tokens []lex.Token
}

func get_call_data(tokens []lex.Token) *_CallData {
	data := &_CallData{}
	data.expr_tokens, data.args_tokens = lex.Range_last(tokens)
	if len(data.expr_tokens) == 0 {
		return data
	}

	// Below is call expression
	token := data.expr_tokens[len(data.expr_tokens)-1]
	if token.Id == lex.ID_RANGE && token.Kind == lex.KND_RBRACKET {
		data.expr_tokens, data.generics_tokens = lex.Range_last(data.expr_tokens)
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

// Returns colon index, left range and right range tokens..
// Returns nil slice and -1 if not found.
// Starts search at *i.
// Increases once *i for each selection.
// *i points to close range token after selection.
func split_colon(tokens []lex.Token) ([]lex.Token, []lex.Token) {
	range_n := 0
	for i, token := range tokens {
		switch token.Id {
		case lex.ID_RANGE:
			switch token.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				range_n++
				continue

			default:
				range_n--
			}
		
		case lex.ID_COLON:
			if range_n < 1 {
				l := tokens[:i]
				r := tokens[i+1:]
				return l, r
			}
		}
	}
	return nil, nil
}

type _Precedencer struct {
	pairs [][]any
}

func (p *_Precedencer) set(level int, expr any) {
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

func (p *_Precedencer) get_lower() any {
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
	prec := _Precedencer{}
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
		Cpp_linked: false,
	}
}

func get_range_expr_tokens(tokens []lex.Token) ([]lex.Token, int) {
	range_n := 0
	i := len(tokens) - 1
	for ; i >= 0; i-- {
		tok := tokens[i]
		if tok.Id == lex.ID_RANGE {
			switch tok.Kind {
			case lex.KND_RBRACE, lex.KND_RBRACKET, lex.KND_RPARENT:
				range_n++
	
			default:
				range_n--
			}
		}

		if range_n == 0 {
			return tokens[:i], range_n
		}
	}
	return nil, range_n
}

type _ExprBuilder struct {
	p *_Parser
}

func (ep *_ExprBuilder) push_err(token lex.Token, key string, args ...any) {
	ep.p.push_err(token, key, args...)
}

func (ep *_ExprBuilder) build_tuple(parts [][]lex.Token) *ast.TupleExpr {
	tuple := &ast.TupleExpr{
		Expr: make([]ast.ExprData, len(parts)),
	}
	for i, part := range parts {
		tuple.Expr[i] = ep.build(part)
	}
	return tuple
}

func (ep *_ExprBuilder) build_lit(token lex.Token) *ast.LitExpr {
	return &ast.LitExpr{
		Token: token,
		Value: token.Kind,
	}
}

func (ep *_ExprBuilder) build_primitive_type(token lex.Token) *ast.Type {
	return build_prim_type(token)
}

func (ep *_ExprBuilder) build_single(token lex.Token) ast.ExprData {
	switch token.Id {
	case lex.ID_LIT:
		return ep.build_lit(token)

	case lex.ID_IDENT, lex.ID_SELF:
		return build_ident_expr(token)

	case lex.ID_PRIM:
		return ep.build_primitive_type(token)

	default:
		ep.push_err(token, "invalid_syntax")
		return nil
	}
}

func (ep *_ExprBuilder) build_cpp_linked_ident(tokens []lex.Token) *ast.IdentExpr {
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
	expr.Cpp_linked = true
	return expr
}

func (ep *_ExprBuilder) build_unary(tokens []lex.Token) *ast.UnaryExpr {
	op := tokens[0]
	if len(tokens) == 1 {
		ep.push_err(op, "missing_expr_for_unary")
		return nil
	} else if !lex.Is_unary_op(op.Kind) {
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

func (ep *_ExprBuilder) build_obj_sub_ident(tokens []lex.Token) *ast.SubIdentExpr {
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

func (ep *_ExprBuilder) build_ns_sub_ident(tokens []lex.Token) *ast.NsSelectionExpr {
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

func (ep *_ExprBuilder) build_type(tokens []lex.Token) *ast.Type {
	i := 0
	t, ok := ep.p.build_type(tokens, &i, false)
	if !ok {
		ep.push_err(tokens[0], "invalid_syntax")
		return nil
	}

	if i < len(tokens) {
		ep.push_err(tokens[i], "invalid_syntax")
	}
	return t
}

func (ep *_ExprBuilder) build_sub_ident(tokens []lex.Token) ast.ExprData {
	i := len(tokens) - 1
	i-- // Set offset to delimiter token.
	token := tokens[i]
	switch token.Id {
	case lex.ID_DOT:
		return ep.build_obj_sub_ident(tokens)

	case lex.ID_DBLCOLON:
		return ep.build_ns_sub_ident(tokens)

	case lex.ID_RANGE:
		// Catch slice, and array types.
		if token.Kind == lex.KND_RBRACKET {
			return ep.build_type(tokens)
		}
	}

	ep.push_err(token, "invalid_syntax")
	return nil
}

func (ep *_ExprBuilder) build_variadic(tokens []lex.Token) *ast.VariadicExpr {
	token := tokens[len(tokens)-1] // Variadic operator token.
	tokens = tokens[:len(tokens)-1] // Remove variadic operator token.
	return &ast.VariadicExpr{
		Token: token,
		Expr:  ep.build(tokens),
	}
}

func (ep *_ExprBuilder) build_op_right(tokens []lex.Token) ast.ExprData {
	token := tokens[len(tokens)-1]
	switch token.Kind {
	case lex.KND_TRIPLE_DOT:
		return ep.build_variadic(tokens)

	default:
		ep.push_err(token, "invalid_syntax")
		return nil
	}
}

func (ep *_ExprBuilder) build_between_parentheses(tokens []lex.Token) ast.ExprData {
	token := tokens[0]
	tokens = tokens[1 : len(tokens)-1] // Remove parentheses.
	if len(tokens) == 0 {
		ep.push_err(token, "missing_expr")
		return nil
	}
	return ep.build(tokens)
}

func (ep *_ExprBuilder) try_build_cast(tokens []lex.Token) *ast.CastExpr {
	range_n := 0
	error_token := tokens[0]
	for i, token := range tokens {
		if token.Id == lex.ID_RANGE {
			switch token.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				range_n++
				continue
			default:
				range_n--
			}
		}
		if range_n > 0 {
			continue
		} else if i+1 == len(tokens) {
			return nil
		}

		type_index := 0
		type_tokens := tokens[1:i]
		expr_tokens := tokens[i+1:]

		if len(expr_tokens) == 0 {
			// Expression is parentheses group.
			return nil
		}

		token = expr_tokens[0]
		if token.Id != lex.ID_RANGE || token.Kind != lex.KND_LPAREN {
			return nil
		}

		cast := &ast.CastExpr{}

		// Expression tokens just parentheses.
		if len(expr_tokens) == 2 {
			ep.push_err(error_token, "missing_expr")
		}

		t, ok := ep.p.build_type(type_tokens, &type_index, true)
		if ok && type_index < len(type_tokens) {
			ep.push_err(type_tokens[type_index], "invalid_syntax")
		} else if !ok {
			return cast
		}
		cast.Kind = t

		i = 0
		expr_tokens = lex.Range(&i, lex.KND_LPAREN, lex.KND_RPARENT, expr_tokens)
		cast.Expr = ep.build(expr_tokens)
		return cast
	}
	return nil
}

func (ep *_ExprBuilder) push_arg(args *[]*ast.Expr, tokens []lex.Token, err_token lex.Token) {
	if len(tokens) == 0 {
		ep.push_err(err_token, "invalid_syntax")
		return
	}
	*args = append(*args, ep.build_from_tokens(tokens))
}

func (ep *_ExprBuilder) build_args(tokens []lex.Token) []*ast.Expr {
	// No argument.
	if len(tokens) < 2 {
		return nil
	}

	var args []*ast.Expr
	last := 0
	range_n := 0
	tokens = tokens[1 : len(tokens)-1] // Remove parentheses.
	for i, token := range tokens {
		if token.Id == lex.ID_RANGE {
			switch token.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				range_n++
			default:
				range_n--
			}
		}
		if range_n > 0 || token.Id != lex.ID_COMMA {
			continue
		}
		ep.push_arg(&args, tokens[last:i], token)
		last = i + 1
	}

	if last < len(tokens) {
		if last == 0 {
			if len(tokens) > 0 {
				ep.push_arg(&args, tokens[last:], tokens[last])
			}
		} else {
			ep.push_arg(&args, tokens[last:], tokens[last-1])
		}
	}

	return args
}

// Tokens should include brackets.
func (ep *_ExprBuilder) build_call_generics(tokens []lex.Token) []*ast.Type {
	if len(tokens) == 0 {
		return nil
	}

	tokens = tokens[1 : len(tokens)-1] // Remove brackets.
	parts, errs := lex.Parts(tokens, lex.ID_COMMA, true)
	generics := make([]*ast.Type, len(parts))
	ep.p.errors = append(ep.p.errors, errs...)
	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		j := 0
		generic, _ := ep.p.build_type(part, &j, true)
		if j < len(part) {
			ep.push_err(part[j], "invalid_syntax")
		}
		generics[i] = generic
	}

	return generics
}

func (ep *_ExprBuilder) build_fn_call(token lex.Token, data *_CallData) *ast.FnCallExpr {
	return &ast.FnCallExpr{
		Token:    token,
		Expr:     ep.build_from_tokens(data.expr_tokens),
		Generics: ep.build_call_generics(data.generics_tokens),
		Args:     ep.build_args(data.args_tokens),
	}
}

func (ep *_ExprBuilder) build_parentheses_range(tokens []lex.Token) ast.ExprData {
	token := tokens[0]
	switch token.Id {
	case lex.ID_RANGE:
		switch token.Kind {
		case lex.KND_LPAREN:
			expr := ep.try_build_cast(tokens)
			if expr != nil {
				return expr
			}
		}
	}

	data := get_call_data(tokens)

	// Expression is parentheses group if data.expr_tokens is zero.
	// data.args_tokens holds tokens of parentheses range (include parentheses).
	if len(data.expr_tokens) == 0 {
		return ep.build_between_parentheses(data.args_tokens)
	}

	return ep.build_fn_call(token, data)
}

func (ep *_ExprBuilder) build_unsafe_expr(tokens []lex.Token) *ast.UnsafeExpr {
	token := tokens[0]
	tokens = tokens[1:] // Remove unsafe keyword.
	i := 0
	range_tokens := lex.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, tokens)
	if len(range_tokens) == 0 {
		ep.push_err(tokens[0], "missing_expr")
		return nil
	}
	return &ast.UnsafeExpr{
		Token: token,
		Expr:  ep.build_from_tokens(range_tokens).Kind,
	}
}

func (ep *_ExprBuilder) build_anon_fn(tokens []lex.Token) *ast.FnDecl {
	return ep.p.build_fn(tokens, false, true, false)
}

func (ep *_ExprBuilder) build_unsafe(tokens []lex.Token) ast.ExprData {
	if len(tokens) == 0 {
		ep.push_err(tokens[0], "invalid_syntax")
		return nil
	}
	switch tokens[1].Id {
	case lex.ID_FN:
		// Unsafe anonymous function.
		return ep.build_anon_fn(tokens)

	default:
		return ep.build_unsafe_expr(tokens)
	}
}

// Tokens should include brace tokens.
func (ep *_ExprBuilder) get_brace_range_literal_expr_parts(tokens []lex.Token) ([][]lex.Token) {
	// No part.
	if len(tokens) < 2 {
		return nil
	}

	var parts [][]lex.Token

	push := func(part []lex.Token, error_token lex.Token) {
		if len(part) == 0 {
			ep.push_err(error_token, "invalid_syntax")
			return
		}
		parts = append(parts, part)
	}

	last := 0
	range_n := 0
	tokens = tokens[1 : len(tokens)-1] // Remove parentheses.
	for i, token := range tokens {
		if token.Id == lex.ID_RANGE {
			switch token.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				range_n++
			default:
				range_n--
			}
		}
		if range_n > 0 || token.Id != lex.ID_COMMA {
			continue
		}
		push(tokens[last:i], token)
		last = i + 1
	}

	if last < len(tokens) {
		if last == 0 {
			if len(tokens) > 0 {
				push(tokens[last:], tokens[last])
			}
		} else {
			push(tokens[last:], tokens[last-1])
		}
	}

	return parts
}

func (ep *_ExprBuilder) build_field_expr_pair(tokens []lex.Token) *ast.FieldExprPair {
	pair := &ast.FieldExprPair{}
	token := tokens[0]
	if token.Id == lex.ID_IDENT {
		if len(tokens) > 1 {
			token := tokens[1]
			if token.Id == lex.ID_COLON {
				pair.Field = tokens[0]
				tokens = tokens[2:] // Remove field identifier and colon tokens.
			}
		}
	}
	pair.Expr = ep.build_from_tokens(tokens).Kind
	return pair
}

func (ep *_ExprBuilder) build_field_expr_pairs(tokens []lex.Token) []*ast.FieldExprPair {
	parts := ep.get_brace_range_literal_expr_parts(tokens)
	if len(parts) == 0 {
		return nil
	}

	pairs := make([]*ast.FieldExprPair, len(parts))
	for i, part := range parts {
		pairs[i] = ep.build_field_expr_pair(part)
	}
	return pairs
}

func (ep *_ExprBuilder) build_typed_struct_literal(tokens []lex.Token) *ast.StructLit {
	i := 0
	t, ok := ep.p.build_type(tokens, &i, true)
	if !ok {
		return nil
	} else if i >= len(tokens) {
		ep.push_err(tokens[0], "invalid_syntax")
		return nil
	}

	tokens = tokens[i:] // Remove type tokens.
	token := tokens[0]
	if token.Id != lex.ID_RANGE || token.Kind != lex.KND_LBRACE {
		ep.push_err(token, "invalid_syntax")
		return nil
	}

	return &ast.StructLit{
		Kind:  t,
		Pairs: ep.build_field_expr_pairs(tokens),
	}
}

func (ep *_ExprBuilder) build_brace_lit_part(tokens []lex.Token) ast.ExprData {
	l, r := split_colon(tokens)
	// If left is not nil, colon token found.
	if l != nil {
		println("pair")
		return &ast.KeyValPair{
			Key: ep.build_from_tokens(l).Kind,
			Val: ep.build_from_tokens(r).Kind,
		}
	}
	println("non-pair")
	return ep.build_from_tokens(tokens).Kind
}

func (ep *_ExprBuilder) build_brace_literal(tokens []lex.Token) *ast.BraceLit {
	parts := ep.get_brace_range_literal_expr_parts(tokens)
	if parts == nil {
		return &ast.BraceLit{Exprs: nil}
	}

	lit := &ast.BraceLit{
		Exprs: make([]ast.ExprData, len(parts)),
	}
	for i, part := range parts {
		lit.Exprs[i] = ep.build_brace_lit_part(part)
	}
	return lit
}

func (ep *_ExprBuilder) build_brace_range(tokens []lex.Token) ast.ExprData {
	expr_tokens, range_n := get_range_expr_tokens(tokens)

	switch {
	case len(expr_tokens) == 0:
		return ep.build_brace_literal(tokens)

	case range_n > 0:
		ep.push_err(tokens[0], "invalid_syntax")
		return nil
	}

	switch expr_tokens[0].Id {
	case lex.ID_UNSAFE:
		return ep.build_unsafe(tokens)

	case lex.ID_FN:
		return ep.build_anon_fn(tokens)

	case lex.ID_IDENT, lex.ID_CPP:
		return ep.build_typed_struct_literal(tokens)

	default:
		ep.push_err(expr_tokens[0], "invalid_syntax")
		return nil
	}
}

// Tokens is should be store enumerable range tokens.
func (ep *_ExprBuilder) get_enumerable_parts(tokens []lex.Token) [][]lex.Token {
	tokens = tokens[1 : len(tokens)-1] // Remove range tokens.
	parts, errors := lex.Parts(tokens, lex.ID_COMMA, true)
	ep.p.errors = append(ep.p.errors, errors...)
	return parts
}

func (ep *_ExprBuilder) build_slice(tokens []lex.Token) *ast.SliceExpr {
	parts := ep.get_enumerable_parts(tokens)
	if len(parts) == 0 {
		return nil
	}

	slc := &ast.SliceExpr{
		Token: tokens[0],
		Elems: make([]ast.ExprData, len(parts)),
	}
	for i, p := range parts {
		slc.Elems[i] = ep.build_from_tokens(p).Kind
	}

	return slc
}

func (ep *_ExprBuilder) build_indexing(expr_tokens []lex.Token,
	tokens []lex.Token, error_token lex.Token) *ast.IndexingExpr {
	tokens = tokens[1 : len(tokens)-1] // Remove brackets.
	return &ast.IndexingExpr{
		Token: error_token,
		Expr:  ep.build_from_tokens(expr_tokens).Kind,
		Index: ep.build_from_tokens(tokens).Kind,
	}
}

func (ep *_ExprBuilder) build_slicing(expr_tokens []lex.Token,
	slicing_tokens []lex.Token, colon int, error_token lex.Token) *ast.SlicingExpr {
	slc := &ast.SlicingExpr{
		Token: error_token,
		Expr:  ep.build_from_tokens(expr_tokens).Kind,
	}

	start_expr_tokens := slicing_tokens[:colon]
	if len(start_expr_tokens) > 0 {
		slc.Start = ep.build_from_tokens(start_expr_tokens).Kind
	}

	to_expr_tokens := slicing_tokens[colon+1:]
	if len(to_expr_tokens) > 0 {
		slc.To = ep.build_from_tokens(to_expr_tokens).Kind
	}

	return slc
}

func (ep *_ExprBuilder) build_bracket_range(tokens []lex.Token) ast.ExprData {
	error_token := tokens[0]
	expr_tokens, range_n := get_range_expr_tokens(tokens)

	switch {
	case len(expr_tokens) == 0:
		return ep.build_slice(tokens)
	case len(expr_tokens) == 0 || range_n > 0:
		ep.push_err(error_token, "invalid_syntax")
		return nil
	}

	// Remove expression tokens.
	// Holds only indexing tokens.
	// Includes brackets.
	tokens = tokens[len(expr_tokens):]

	// Use split_map_range because same thing.
	// Map types like: [KEY:VALUE]
	// Slicing expressions like: [START:TO]
	i := 0
	slicing_tokens, colon := split_map_range(tokens, &i)
	if colon != -1 {
		return ep.build_slicing(expr_tokens, slicing_tokens, colon, error_token)
	}
	return ep.build_indexing(expr_tokens, tokens, error_token)
}

func (ep *_ExprBuilder) build_data(tokens []lex.Token) ast.ExprData {
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

	case lex.ID_PRIM:
		// Catch slice, and array types.
		return ep.build_type(tokens)
	
	case lex.ID_OP:
		return ep.build_op_right(tokens)

	case lex.ID_RANGE:
		switch token.Kind {
		case lex.KND_RPARENT:
			return ep.build_parentheses_range(tokens)

		case lex.KND_RBRACE:
			return ep.build_brace_range(tokens)
		
		case lex.KND_RBRACKET:
			return ep.build_bracket_range(tokens)
		}
	}

	ep.push_err(tokens[0], "invalid_syntax")
	return nil
}

func (ep *_ExprBuilder) build_binop(tokens []lex.Token, i int) *ast.BinopExpr {
	return &ast.BinopExpr{
		L:  ep.build(tokens[:i]),
		R:  ep.build(tokens[i+1:]),
		Op: tokens[i],
	}
}

func (ep *_ExprBuilder) build(tokens []lex.Token) ast.ExprData {
	i := find_lowest_prec_op(tokens)
	if i == -1 {
		return ep.build_data(tokens)
	}
	return ep.build_binop(tokens, i)
}

func (ep *_ExprBuilder) build_kind(tokens []lex.Token) ast.ExprData {
	parts, errors := lex.Parts(tokens, lex.ID_COMMA, true)
	if errors != nil {
		ep.p.errors = append(ep.p.errors, errors...)
		return nil
	} else if len(parts) > 1 {
		return ep.build_tuple(parts)
	}
	return ep.build(tokens)
}

func (ep *_ExprBuilder) build_from_tokens(tokens []lex.Token) *ast.Expr {
	tokens = eliminate_comments(tokens)
	if len(tokens) == 0 {
		return nil
	}
	return &ast.Expr{
		Token: tokens[0],
		Kind:  ep.build_kind(tokens),
	}
}
