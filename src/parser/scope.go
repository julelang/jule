// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package parser

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

func new_scope() *ast.Scope {
	return &ast.Scope{}
}

// Reports whether token is statement finish point.
func is_st(current lex.Token, prev lex.Token) (ok bool, terminated bool) {
	ok = current.Id == lex.ID_SEMICOLON || prev.Row < current.Row
	terminated = current.Id == lex.ID_SEMICOLON
	return
}

// Reports position of the next statement if exist, len(toks) if not.
func next_st_pos(tokens []lex.Token, start int) (int, bool) {
	brace_n := 0
	i := start
	for ; i < len(tokens); i++ {
		var ok, terminated bool
		tok := tokens[i]
		if tok.Id == lex.ID_RANGE {
			switch tok.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				if brace_n == 0 && i > start {
					ok, terminated = is_st(tok, tokens[i-1])
					if ok {
						goto ret
					}
				}
				brace_n++
				continue
			default:
				brace_n--
				if brace_n == 0 && i+1 < len(tokens) {
					ok, terminated = is_st(tokens[i+1], tok)
					if ok {
						i++
						goto ret
					}
				}
				continue
			}
		}
		if brace_n != 0 {
			continue
		} else if i > start {
			ok, terminated = is_st(tok, tokens[i-1])
		} else {
			ok, terminated = is_st(tok, tok)
		}
		if !ok {
			continue
		}
	ret:
		if terminated {
			i++
		}
		return i, terminated
	}
	return i, false
}

// Returns current statement tokens.
// Starts selection at *i.
func skip_st(i *int, tokens []lex.Token) ([]lex.Token, bool) {
	start := *i
	terminated := false
	*i, terminated = next_st_pos(tokens, start)
	st_tokens := tokens[start:*i]
	if terminated {
		if len(st_tokens) == 1 {
			return skip_st(i, tokens)
		}
		// -1 for eliminate statement terminator.
		st_tokens = st_tokens[:len(st_tokens)-1]
	}
	return st_tokens, terminated
}

type _Stmt struct {
	tokens     []lex.Token
	terminated bool
}

// Splits all statements.
func split_stms(tokens []lex.Token) []*_Stmt {
	var stms []*_Stmt = nil
	pos := 0
	for pos < len(tokens) {
		tokens, terminated := skip_st(&pos, tokens)
		stms = append(stms, &_Stmt{
			tokens:     tokens,
			terminated: terminated,
		})
	}
	return stms
}

type _ScopeParser struct {
	p    *_Parser
	s    *ast.Scope
	stms []*_Stmt
	pos  int
}

func (sp *_ScopeParser) stop() { sp.pos = -1 }
func (sp *_ScopeParser) stopped() bool { return sp.pos == -1 }
func (sp *_ScopeParser) finished() bool { return sp.pos >= len(sp.stms) }
func (sp *_ScopeParser) is_last_st() bool { return sp.pos+1 >= len(sp.stms) }
func (sp *_ScopeParser) push_err(token lex.Token, key string) { sp.p.push_err(token, key) }

func (sp *_ScopeParser) insert_as_next(tokens []lex.Token) {
    sp.stms = append(sp.stms[:sp.pos+1], sp.stms[sp.pos:]...)
    sp.stms[sp.pos+1] = &_Stmt{tokens: tokens}
}

func (sp *_ScopeParser) next() *_Stmt {
	sp.pos++
	return sp.stms[sp.pos]
}

func (sp *_ScopeParser) build_scope(tokens []lex.Token) *ast.Scope {
	s := new_scope()
	s.Parent = sp.s
	ssp := _ScopeParser{
		p: sp.p,
	}
	ssp.build(tokens, s)
	return s
}

func (sp *_ScopeParser) build_var_st(tokens []lex.Token) ast.NodeData {
	v := sp.p.build_var(tokens)
	v.Scope = sp.s
	return v
}

func (sp *_ScopeParser) build_ret_st(tokens []lex.Token) ast.NodeData {
	st := &ast.RetSt{
		Token: tokens[0],
	}
	if len(tokens) > 1 {
		tokens = tokens[1:] // Remove ret keyword.
		st.Expr = sp.p.build_expr(tokens)
	}
	return st
}

func (sp *_ScopeParser) build_while_next_iter(s *_Stmt) ast.NodeData {
	it := &ast.Iter{
		Token: s.tokens[0],
	}
	tokens := s.tokens[1:] // Skip "iter" keyword.
	kind := &ast.WhileNextKind{}
	if len(tokens) > 0 {
		kind.Expr = sp.p.build_expr(tokens)
	}
	if sp.is_last_st() {
		sp.push_err(it.Token, "invalid_syntax")
		return nil
	}
	tokens = sp.next().tokens
	st_tokens := get_block_expr(tokens)
	if len(st_tokens) > 0 {
		s := &_Stmt{
			terminated: s.terminated,
			tokens:     st_tokens,
		}
		kind.Next = sp.build_st(s)
	}
	i := len(st_tokens)
	block_tokens := lex.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, tokens)
	if block_tokens == nil {
		sp.stop()
		sp.push_err(it.Token, "body_not_exist")
		return nil
	}
	if i < len(tokens) {
		sp.push_err(tokens[i], "invalid_syntax")
	}
	it.Scope = sp.build_scope(block_tokens)
	it.Kind = kind
	return it
}

func (sp *_ScopeParser) build_while_iter_kind(tokens []lex.Token) *ast.WhileKind {
	return &ast.WhileKind{
		Expr: sp.p.build_expr(tokens),
	}
}

func (sp *_ScopeParser) get_range_kind_keys_tokens(toks []lex.Token) [][]lex.Token {
	vars, errs := lex.Parts(toks, lex.ID_COMMA, true)
	sp.p.errors = append(sp.p.errors, errs...)
	return vars
}

func (sp *_ScopeParser) build_range_kind_key(tokens []lex.Token) *ast.VarDecl {
	if len(tokens) == 0 {
		return nil
	}
	key := &ast.VarDecl{
		Token: tokens[0],
	}
	if key.Token.Id == lex.ID_MUT {
		key.Mutable = true
		if len(tokens) == 1 {
			sp.push_err(key.Token, "invalid_syntax")
		}
		key.Token = tokens[1]
	} else if len(tokens) > 1 {
		sp.push_err(tokens[1], "invalid_syntax")
	}
	if key.Token.Id != lex.ID_IDENT {
		sp.push_err(key.Token, "invalid_syntax")
		return nil
	}
	key.Ident = key.Token.Kind
	return key
}

func (sp *_ScopeParser) build_range_kind_keys(parts [][]lex.Token) []*ast.VarDecl {
	var keys []*ast.VarDecl
	for _, tokens := range parts {
		keys = append(keys, sp.build_range_kind_key(tokens))
	}
	return keys
}

func (sp *_ScopeParser) setup_range_kind_keys_plain(rng *ast.RangeKind, tokens []lex.Token) {
	key_tokens := sp.get_range_kind_keys_tokens(tokens)
	if len(key_tokens) == 0 {
		return
	}
	if len(key_tokens) > 2 {
		sp.push_err(rng.In_token, "much_foreach_vars")
	}
	keys := sp.build_range_kind_keys(key_tokens)
	rng.Key_a = keys[0]
	if len(keys) > 1 {
		rng.Key_b = keys[1]
	}
}

func (sp *_ScopeParser) setup_range_kind_keys_explicit(rng *ast.RangeKind, tokens []lex.Token) {
	i := 0
	rang := lex.Range(&i, lex.KND_LPAREN, lex.KND_RPARENT, tokens)
	if i < len(tokens) {
		sp.push_err(rng.In_token, "invalid_syntax")
	}
	sp.setup_range_kind_keys_plain(rng, rang)
}

func (sp *_ScopeParser) setup_range_kind_keys(rng *ast.RangeKind, tokens []lex.Token) {
	if tokens[0].Id == lex.ID_RANGE {
		if tokens[0].Kind != lex.KND_LPAREN {
			sp.push_err(tokens[0], "invalid_syntax")
			return
		}
		sp.setup_range_kind_keys_explicit(rng, tokens)
		return
	}
	sp.setup_range_kind_keys_plain(rng, tokens)
}

func (sp *_ScopeParser) build_range_iter_kind(var_tokens []lex.Token, expr_tokens []lex.Token, in_token lex.Token) *ast.RangeKind {
	rng := &ast.RangeKind{
		In_token: in_token,
	}
	if len(expr_tokens) == 0 {
		sp.push_err(rng.In_token, "missing_expr")
		return rng
	}
	rng.Expr = sp.p.build_expr(expr_tokens)
	if len(var_tokens) > 0 {
		sp.setup_range_kind_keys(rng, var_tokens)
	}
	return rng
}

func (sp *_ScopeParser) build_common_iter_kind(tokens []lex.Token, err_tok lex.Token) ast.IterKind {
	brace_n := 0
	for i, tok := range tokens {
		if tok.Id == lex.ID_RANGE {
			switch tok.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				brace_n++
				continue
			default:
				brace_n--
			}
		}
		if brace_n != 0 {
			continue
		}
		switch tok.Id {
		case lex.ID_IN:
			decl_tokens := tokens[:i]
			expr_tokens := tokens[i+1:]
			return sp.build_range_iter_kind(decl_tokens, expr_tokens, tok)
		}
	}
	return sp.build_while_iter_kind(tokens)
}

func (sp *_ScopeParser) build_common_iter(tokens []lex.Token) ast.NodeData {
	it := &ast.Iter{
		Token: tokens[0],
	}
	tokens = tokens[1:] // Skip "iter" keyword.
	if len(tokens) == 0 {
		sp.stop()
		sp.push_err(it.Token, "body_not_exist")
		return nil
	}
	expr_tokens := get_block_expr(tokens)
	if len(expr_tokens) > 0 {
		it.Kind = sp.build_common_iter_kind(expr_tokens, it.Token)
	}
	i := len(expr_tokens)
	scope_tokens := lex.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, tokens)
	if scope_tokens == nil {
		sp.stop()
		sp.push_err(it.Token, "body_not_exist")
		return nil
	}
	if i < len(tokens) {
		sp.push_err(tokens[i], "invalid_syntax")
	}
	it.Scope = sp.build_scope(scope_tokens)
	return it
}

func (sp *_ScopeParser) buid_iter_st(st *_Stmt) ast.NodeData {
	if st.terminated {
		return sp.build_while_next_iter(st)
	}
	return sp.build_common_iter(st.tokens)
}

func (sp *_ScopeParser) build_break_st(tokens []lex.Token) ast.NodeData {
	brk := &ast.BreakSt{
		Token: tokens[0],
	}
	if len(tokens) > 1 {
		if tokens[1].Id != lex.ID_IDENT {
			sp.push_err(tokens[1], "invalid_syntax")
		} else {
			brk.Label = tokens[1]
			if len(tokens) > 2 {
				sp.push_err(tokens[1], "invalid_syntax")
			}
		}
	}
	return brk
}

func (sp *_ScopeParser) build_cont_st(tokens []lex.Token) ast.NodeData {
	cont := &ast.ContSt{
		Token: tokens[0],
	}
	if len(tokens) > 1 {
		if tokens[1].Id != lex.ID_IDENT {
			sp.push_err(tokens[1], "invalid_syntax")
		} else {
			cont.Label = tokens[1]
			if len(tokens) > 2 {
				sp.push_err(tokens[1], "invalid_syntax")
			}
		}
	}
	return cont
}

func (sp *_ScopeParser) build_if(tokens *[]lex.Token) *ast.If {
	model := &ast.If{
		Token: (*tokens)[0],
	}
	*tokens = (*tokens)[1:]
	expr_tokens := get_block_expr(*tokens)
	i := 0
	if len(expr_tokens) == 0 {
		sp.push_err(model.Token, "missing_expr")
	} else {
		i = len(expr_tokens)
	}
	scope_tokens := lex.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, *tokens)
	if scope_tokens == nil {
		sp.stop()
		sp.push_err(model.Token, "body_not_exist")
		return nil
	}
	if i < len(*tokens) {
		if (*tokens)[i].Id == lex.ID_ELSE {
			*tokens = (*tokens)[i:]
		} else {
			sp.push_err((*tokens)[i], "invalid_syntax")
			*tokens = nil
		}
	}
	model.Expr = sp.p.build_expr(expr_tokens)
	model.Scope = sp.build_scope(scope_tokens)
	return model
}

func (sp *_ScopeParser) build_else(tokens []lex.Token) *ast.Else {
	els := &ast.Else{
		Token: tokens[0],
	}
	tokens = tokens[1:] // Remove "else" keyword.
	i := 0
	scope_tokens := lex.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, tokens)
	if scope_tokens == nil {
		if i < len(tokens) {
			sp.push_err(els.Token, "else_have_expr")
		} else {
			sp.stop()
			sp.push_err(els.Token, "body_not_exist")
		}
		return nil
	}
	if i < len(tokens) {
		sp.push_err(tokens[i], "invalid_syntax")
	}
	els.Scope = sp.build_scope(scope_tokens)
	return els
}

func (sp *_ScopeParser) build_if_else_chain(tokens []lex.Token) ast.NodeData {
	chain := &ast.Conditional{
		If: sp.build_if(&tokens),
	}
	if chain.If == nil {
		return nil
	}
	for tokens != nil {
		if tokens[0].Id != lex.ID_ELSE {
			break
		}
		if len(tokens) > 1 && tokens[1].Id == lex.ID_IF {
			tokens = tokens[1:] // Remove else token
			elif := sp.build_if(&tokens)
			chain.Elifs = append(chain.Elifs, elif)
			continue
		}
		chain.Default = sp.build_else(tokens)
		break
	}
	return chain
}

func (sp *_ScopeParser) build_comment_st(token lex.Token) ast.NodeData {
	return build_comment(token)
}

func (sp *_ScopeParser) build_call_st(tokens []lex.Token) ast.NodeData {
	token := tokens[0]
	if is_fn_call(tokens) == nil {
		sp.push_err(token, "expr_not_func_call")
	}
	expr := sp.p.build_expr(tokens)
	if !expr.Is_fn_call() {
		sp.push_err(token, "invalid_syntax")
	}
	return expr
}

func (sp *_ScopeParser) build_co_call_st(tokens []lex.Token) ast.NodeData {
	cc := sp.build_call_st(tokens)
	cc.(*ast.FnCallExpr).Concurrent = true
	return cc
}

func (sp *_ScopeParser) build_goto_st(tokens []lex.Token) ast.NodeData {
	gt := &ast.GotoSt{
		Token: tokens[0],
	}
	if len(tokens) == 1 {
		sp.push_err(gt.Token, "missing_goto_label")
		return nil
	} else if len(tokens) > 2 {
		sp.push_err(tokens[2], "invalid_syntax")
	}
	ident_token := tokens[1]
	if ident_token.Id != lex.ID_IDENT {
		sp.push_err(ident_token, "invalid_syntax")
		return gt
	}
	gt.Label = ident_token
	return gt
}

func (sp *_ScopeParser) build_fall_st(tokens []lex.Token) ast.NodeData {
	fll := &ast.FallSt{
		Token: tokens[0],
	}
	if len(tokens) > 1 {
		sp.push_err(tokens[1], "invalid_syntax")
	}
	return fll
}

func (sp *_ScopeParser) build_type_alias_st(tokens []lex.Token) ast.NodeData {
	tad := sp.p.build_type_alias_decl(tokens)
	return tad
}

func (sp *_ScopeParser) build_case_exprs(tokens *[]lex.Token, type_match bool) []*ast.Expr {
	var exprs []*ast.Expr
	push_expr := func(tokens []lex.Token, token lex.Token) {
		if len(tokens) > 0 {
			if type_match {
				i := 0
				t, ok := sp.p.build_type(tokens, &i, true)
				if ok {
					exprs = append(exprs, &ast.Expr{
						Token: token,
						Kind:  t,
					})
				}
				if i < len(tokens) {
					sp.push_err(tokens[i], "invalid_syntax")
				}
				return
			}
			exprs = append(exprs, sp.p.build_expr(tokens))
		}
	}

	brace_n := 0
	j := 0
	var i int
	var tok lex.Token
	for i, tok = range *tokens {
		if tok.Id == lex.ID_RANGE {
			switch tok.Kind {
			case lex.KND_LPAREN, lex.KND_LBRACE, lex.KND_LBRACKET:
				brace_n++
			default:
				brace_n--
			}
			continue
		} else if brace_n != 0 {
			continue
		}
		switch {
		case tok.Id == lex.ID_OP && tok.Kind == lex.KND_VLINE:
			push_expr((*tokens)[j:i], tok)
			j = i + 1
		case tok.Id == lex.ID_COLON:
			push_expr((*tokens)[j:i], tok)
			*tokens = (*tokens)[i+1:]
			return exprs
		}
	}
	sp.push_err((*tokens)[0], "invalid_syntax")
	*tokens = nil
	return nil
}

func (sp *_ScopeParser) build_case_scope(tokens *[]lex.Token) *ast.Scope {
	n := 0
	for {
		i := 0
		next, _ := skip_st(&i, (*tokens)[n:])
		if len( next) == 0 {
			break
		}
		tok := next[0]
		if tok.Id != lex.ID_OP || tok.Kind != lex.KND_VLINE {
			n += len(next)
			continue
		}
		scope := sp.build_scope((*tokens)[:n])
		*tokens = (*tokens)[n:]
		return scope
	}
	scope := sp.build_scope(*tokens)
	*tokens = nil
	return scope
}

func (sp *_ScopeParser) build_case(tokens *[]lex.Token, type_match bool) (*ast.Case, bool) {
	c := &ast.Case{
		Token: (*tokens)[0], 
	}
	*tokens = (*tokens)[1:] // Remove case prefix.
	c.Exprs = sp.build_case_exprs(tokens, type_match)
	c.Scope = sp.build_case_scope(tokens)
	is_default := len(c.Exprs) == 0
	return c, is_default
}

func (sp *_ScopeParser) build_cases(tokens []lex.Token, type_match bool) ([]*ast.Case, *ast.Else) {
	var cases []*ast.Case
	var def *ast.Else
	for len(tokens) > 0 {
		tok := tokens[0]
		if tok.Id != lex.ID_OP || tok.Kind != lex.KND_VLINE {
			sp.push_err(tok, "invalid_syntax")
			break
		}
		c, is_default := sp.build_case(&tokens, type_match)
		if is_default {
			c.Token = tok
			if def == nil {
				def = &ast.Else{
					Token: c.Token,
					Scope: c.Scope,
				}
			} else {
				sp.push_err(tok, "invalid_syntax")
			}
		} else {
			cases = append(cases, c)
		}
	}
	return cases, def
}

func (sp *_ScopeParser) build_match_case(tokens []lex.Token) *ast.MatchCase {
	m := &ast.MatchCase{
		Token: tokens[0],
	}
	tokens = tokens[1:] // Remove "match" keyword.
	
	if len(tokens) > 0 && tokens[0].Id == lex.ID_TYPE {
		m.Type_match = true
		tokens = tokens[1:] // Skip "type" keyword
	}

	expr_tokens := get_block_expr(tokens)
	if len(expr_tokens) > 0 {
		m.Expr = sp.p.build_expr(expr_tokens)
	} else if m.Type_match {
		sp.push_err(m.Token, "missing_expr")
	}
	
	i := len(expr_tokens)
	block_toks := lex.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, tokens)
	if block_toks == nil {
		sp.stop()
		sp.push_err(m.Token, "body_not_exist")
		return nil
	}
	
	m.Cases, m.Default = sp.build_cases(block_toks, m.Type_match)
	return m
}

func (sp *_ScopeParser) build_scope_st(tokens []lex.Token) *ast.Scope {
	is_unsafe := false
	is_deferred := false
	token := tokens[0]
	if token.Id == lex.ID_UNSAFE {
		is_unsafe = true
		tokens = tokens[1:]
		if len(tokens) == 0 {
			sp.push_err(token, "invalid_syntax")
			return nil
		}
		token = tokens[0]
		if token.Id == lex.ID_DEFER {
			is_deferred = true
			tokens = tokens[1:]
			if len(tokens) == 0 {
				sp.push_err(token, "invalid_syntax")
				return nil
			}
		}
	} else if token.Id == lex.ID_DEFER {
		is_deferred = true
		tokens = tokens[1:]
		if len(tokens) == 0 {
			sp.push_err(token, "invalid_syntax")
			return nil
		}
	}

	i := 0
	tokens = lex.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, tokens)
	if len(tokens) == 0 {
		sp.push_err(token, "invalid_syntax")
		return nil
	} else if i < len(tokens) {
		sp.push_err(tokens[i], "invalid_syntax")
	}
	scope := sp.build_scope(tokens)
	scope.Unsafety = is_unsafe
	scope.Deferred = is_deferred
	return scope
}

func (sp *_ScopeParser) build_label_st(tokens []lex.Token) *ast.LabelSt {
	lbl := &ast.LabelSt{
		Token: tokens[0],
		Ident: tokens[0].Kind,
	}

	// Save followed statement
	if len(tokens) > 2 {
		tokens = tokens[2:] // Remove goto keyword and label
		sp.insert_as_next(tokens)
	}

	return lbl
}

func (sp *_ScopeParser) build_id_st(tokens []lex.Token) (_ ast.NodeData, ok bool) {
	if len(tokens) == 1 {
		return
	}
	token := tokens[1]
	switch token.Id {
	case lex.ID_COLON:
		return sp.build_label_st(tokens), true
	}
	return
}

func (sp *_ScopeParser) build_assign_info(tokens []lex.Token) *_AssignInfo {
	info:= &_AssignInfo{
		ok: true,
	}
	brace_n := 0
	for i, token := range tokens {
		if token.Id == lex.ID_RANGE {
			switch token.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				brace_n++
			default:
				brace_n--
			}
		}
		if brace_n > 0 {
			continue
		} else if token.Id != lex.ID_OP {
			continue
		} else if !is_assign_op(token.Kind) {
			continue
		}
		info.l = tokens[:i]
		if len(info.l) == 0 {
			info.ok = false
		}
		info.setter = token
		if i+1 >= len(tokens) {
			info.r = nil
			info.ok = is_postfix_op(info.setter.Kind)
			break
		}
		info.r = tokens[i+1:]
		if is_postfix_op(info.setter.Kind) {
			if len(info.r) > 0 {
				sp.push_err(info.r[0], "invalid_syntax")
				info.r = nil
			}
		}
		break
	}
	return info
}

func (sp *_ScopeParser) build_assign_l(tokens []lex.Token) *ast.AssignLeft {
	l := &ast.AssignLeft{
		Token: tokens[0],
	}
	if tokens[0].Id == lex.ID_IDENT {
		l.Ident = l.Token.Kind
	}
	l.Expr = sp.p.build_expr(tokens)
	return l
}

func (sp *_ScopeParser) build_assign_ls(parts [][]lex.Token) []*ast.AssignLeft {
	var lefts []*ast.AssignLeft
	for _, part := range parts {
		l := sp.build_assign_l(part)
		lefts = append(lefts, l)
	}
	return lefts
}

func (sp *_ScopeParser) build_plain_assign(tokens []lex.Token) (_ *ast.AssignSt, ok bool) {
	info := sp.build_assign_info(tokens)
	if !info.ok {
		return
	}
	ok = true
	assign := &ast.AssignSt{
		Setter: info.setter,
	}
	parts, errs := lex.Parts(info.l, lex.ID_COMMA, true)
	if len(errs) > 0 {
		sp.p.errors = append(sp.p.errors, errs...)
		return nil, false
	}
	assign.L = sp.build_assign_ls(parts)
	if info.r != nil {
		assign.R = sp.p.build_expr(info.r)
	}
	return
}

func (sp *_ScopeParser) build_decl_assign(tokens []lex.Token) (_ *ast.AssignSt, ok bool) {
	if len(tokens) < 1 {
		return
	}

	tokens = tokens[1:] // Skip "let" keyword
	token := tokens[0]
	if token.Id != lex.ID_RANGE || token.Kind != lex.KND_LPAREN {
		return
	}
	ok = true

	assign := &ast.AssignSt{}

	var i int
	rang := lex.Range(&i, lex.KND_LPAREN, lex.KND_RPARENT, tokens)
	if rang == nil {
		sp.push_err(token, "invalid_syntax")
		return
	} else if i+1 < len(tokens) {
		assign.Setter = tokens[i]
		i++
		assign.R = sp.p.build_expr(tokens[i:])
	}

	// Lefts
	parts, errs := lex.Parts(rang, lex.ID_COMMA, true)
	if len(errs) > 0 {
		sp.p.errors = append(sp.p.errors, errs...)
		return
	}
	for _, part := range parts {
		is_mut := false
		token := part[0]
		if token.Id == lex.ID_MUT {
			is_mut = true
			part = part[1:]
			if len(part) != 1 {
				sp.push_err(token, "invalid_syntax")
				continue
			}
		}
		if part[0].Id != lex.ID_IDENT && part[0].Id != lex.ID_RANGE && part[0].Kind != lex.KND_LPAREN {
			sp.push_err(token, "invalid_syntax")
			continue
		}
		l := sp.build_assign_l(part)
		l.Mutable = is_mut
		assign.L = append(assign.L, l)
	}
	return
}

func (sp *_ScopeParser) build_assign_st(tokens []lex.Token) (*ast.AssignSt, bool) {
	if !check_assign_tokens(tokens) {
		return nil, false
	}
	switch tokens[0].Id {
	case lex.ID_LET:
		return sp.build_decl_assign(tokens)
	default:
		return sp.build_plain_assign(tokens)
	}
}

func (sp *_ScopeParser) build_st(st *_Stmt) ast.NodeData {
	token := st.tokens[0]
	if token.Id == lex.ID_IDENT {
		s, ok := sp.build_id_st(st.tokens)
		if ok {
			return s
		}
	}

	s, ok := sp.build_assign_st(st.tokens)
	if ok {
		return s
	}

	switch token.Id {
	case lex.ID_CONST, lex.ID_LET, lex.ID_MUT:
		return sp.build_var_st(st.tokens)

	case lex.ID_RET:
		return sp.build_ret_st(st.tokens)

	case lex.ID_FOR:
		return sp.buid_iter_st(st)

	case lex.ID_BREAK:
		return sp.build_break_st(st.tokens)

	case lex.ID_CONTINUE:
		return sp.build_cont_st(st.tokens)

	case lex.ID_IF:
		return sp.build_if_else_chain(st.tokens)

	case lex.ID_COMMENT:
		// Push first token because this is full text comment.
		// Comments are just single-line.
		// Range comments not accepts by lexer.
		return sp.build_comment_st(token)

	case lex.ID_CO:
		return sp.build_co_call_st(st.tokens)

	case lex.ID_GOTO:
		return sp.build_goto_st(st.tokens)

	case lex.ID_FALL:
		return sp.build_fall_st(st.tokens)

	case lex.ID_TYPE:
		return sp.build_type_alias_st(st.tokens)

	case lex.ID_MATCH:
		return sp.build_match_case(st.tokens)

	case lex.ID_UNSAFE, lex.ID_DEFER:
		return sp.build_scope_st(st.tokens)
	
	case lex.ID_RANGE:
		if token.Kind == lex.KND_LBRACE {
			return sp.build_scope_st(st.tokens)
		}
	
	default:
		if is_fn_call(st.tokens) != nil {
			return sp.build_call_st(st.tokens)
		}
	}
	sp.push_err(token, "invalid_syntax")
	return nil
}

func (sp *_ScopeParser) build(tokens []lex.Token, s *ast.Scope) {
	if s == nil {
		return
	}

	sp.stms = split_stms(tokens)
	sp.pos = -1 // sp.next() first increase position
	sp.s = s
	for !sp.is_last_st() && !sp.finished() {
		st := sp.next()
		data := sp.build_st(st)
		if data != nil {
			sp.s.Stmts = append(sp.s.Stmts, data)
		}

		if sp.stopped() {
			break
		}
	}
}
