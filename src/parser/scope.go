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

type st struct {
	tokens     []lex.Token
	terminated bool
}

// Splits all statements.
func split_stms(tokens []lex.Token) []*st {
	var stms []*st = nil
	pos := 0
	for pos < len(tokens) {
		tokens, terminated := skip_st(&pos, tokens)
		stms = append(stms, &st{
			tokens:     tokens,
			terminated: terminated,
		})
	}
	return stms
}

type scope_parser struct {
	p    *parser
	s    *ast.Scope
	stms []*st
	pos  int
}

func (sp *scope_parser) stop() { sp.pos = -1 }
func (sp *scope_parser) stopped() bool { return sp.pos == -1 }
func (sp *scope_parser) finished() bool { return sp.pos >= len(sp.stms) }
func (sp *scope_parser) is_last_st() bool { return sp.pos+1 >= len(sp.stms) }
func (sp *scope_parser) push_err(token lex.Token, key string) { sp.p.push_err(token, key) }

func (sp *scope_parser) next() *st {
	sp.pos++
	return sp.stms[sp.pos]
}

func (sp *scope_parser) build_scope(tokens []lex.Token) *ast.Scope {
	s := new_scope()
	s.Parent = sp.s
	ssp := scope_parser{
		p: sp.p,
	}
	ssp.build(tokens, s)
	return s
}

func (sp *scope_parser) build_var_st(tokens []lex.Token) ast.NodeData {
	v := sp.p.build_var(tokens)
	v.Scope = sp.s
	return v
}

func (sp *scope_parser) build_ret_st(tokens []lex.Token) ast.NodeData {
	st := &ast.RetSt{
		Token: tokens[0],
	}
	if len(tokens) > 1 {
		tokens = tokens[1:] // Remove ret keyword.
		st.Expr = sp.p.build_expr(tokens)
	}
	return st
}

func (sp *scope_parser) build_while_next_iter(s *st) ast.NodeData {
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
	s = sp.next()
	st_tokens := get_block_expr(s.tokens)
	if len(st_tokens) > 0 {
		s := &st{
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

func (sp *scope_parser) build_while_iter_kind(tokens []lex.Token) *ast.WhileKind {
	return &ast.WhileKind{
		Expr: sp.p.build_expr(tokens),
	}
}

func (sp *scope_parser) get_range_kind_keys_tokens(toks []lex.Token) [][]lex.Token {
	vars, errs := lex.Parts(toks, lex.ID_COMMA, true)
	sp.p.errors = append(sp.p.errors, errs...)
	return vars
}

func (sp *scope_parser) build_range_kind_key(tokens []lex.Token) *ast.VarDecl {
	if len(tokens) == 0 {
		return nil
	}
	key := &ast.VarDecl{
		Token: tokens[0],
	}
	if key.Token.Id == lex.ID_MUT {
		key.IsMut = true
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

func (sp *scope_parser) build_range_kind_keys(parts [][]lex.Token) []*ast.VarDecl {
	var keys []*ast.VarDecl
	for _, tokens := range parts {
		keys = append(keys, sp.build_range_kind_key(tokens))
	}
	return keys
}

func (sp *scope_parser) setup_range_kind_keys_plain(rng *ast.RangeKind, tokens []lex.Token) {
	key_tokens := sp.get_range_kind_keys_tokens(tokens)
	if len(key_tokens) == 0 {
		return
	}
	if len(key_tokens) > 2 {
		sp.push_err(rng.InToken, "much_foreach_vars")
	}
	keys := sp.build_range_kind_keys(key_tokens)
	rng.KeyA = keys[0]
	if len(keys) > 1 {
		rng.KeyB = keys[1]
	}
}

func (sp *scope_parser) setup_range_kind_keys_explicit(rng *ast.RangeKind, tokens []lex.Token) {
	i := 0
	rang := lex.Range(&i, lex.KND_LPAREN, lex.KND_RPARENT, tokens)
	if i < len(tokens) {
		sp.push_err(rng.InToken, "invalid_syntax")
	}
	sp.setup_range_kind_keys_plain(rng, rang)
}

func (sp *scope_parser) setup_range_kind_keys(rng *ast.RangeKind, tokens []lex.Token) {
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

func (sp *scope_parser) build_range_iter_kind(var_tokens []lex.Token, expr_tokens []lex.Token, in_token lex.Token) *ast.RangeKind {
	rng := &ast.RangeKind{
		InToken: in_token,
	}
	if len(expr_tokens) == 0 {
		sp.push_err(rng.InToken, "missing_expr")
		return rng
	}
	rng.Expr = sp.p.build_expr(expr_tokens)
	if len(var_tokens) > 0 {
		sp.setup_range_kind_keys(rng, var_tokens)
	}
	return rng
}

func (sp *scope_parser) build_common_iter_kind(tokens []lex.Token, err_tok lex.Token) ast.IterKind {
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

func (sp *scope_parser) build_common_iter(tokens []lex.Token) ast.NodeData {
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

func (sp *scope_parser) buid_iter_st(st *st) ast.NodeData {
	if st.terminated {
		return sp.build_while_next_iter(st)
	}
	return sp.build_common_iter(st.tokens)
}

func (sp *scope_parser) build_break_st(tokens []lex.Token) ast.NodeData {
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

func (sp *scope_parser) build_cont_st(tokens []lex.Token) ast.NodeData {
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

func (sp *scope_parser) build_if(tokens *[]lex.Token) *ast.If {
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

func (sp *scope_parser) build_else(tokens []lex.Token) *ast.Else {
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

func (sp *scope_parser) build_if_else_chain(tokens []lex.Token) ast.NodeData {
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

func (sp *scope_parser) build_comment_st(token lex.Token) ast.NodeData {
	return build_comment(token)
}

// Tokens should include brackets.
func (sp *scope_parser) build_call_generics(tokens []lex.Token) []*ast.Type {
	if len(tokens) == 0 {
		return nil
	}

	tokens = tokens[1 : len(tokens)-1] // Remove braces
	parts, errs := lex.Parts(tokens, lex.ID_COMMA, true)
	generics := make([]*ast.Type, len(parts))
	sp.p.errors = append(sp.p.errors, errs...)
	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		j := 0
		generic, _ := sp.p.build_type(part, &j, true)
		if j+1 < len(part) {
			sp.push_err(part[j+1], "invalid_syntax")
		}
		generics[i] = generic
	}

	return generics
}

func (sp *scope_parser) build_args(tokens []lex.Token) []*ast.Expr {
	i := 0
	tokens = lex.Range(&i, lex.KND_LPAREN, lex.KND_RPARENT, tokens)
	if tokens == nil {
		tokens = make([]lex.Token, 0)
	}
	args := sp.p.build_args(tokens)
	return args
}

func (sp *scope_parser) build_call_st(tokens []lex.Token) ast.NodeData {
	cc := &ast.FnCallExpr{
		Token: tokens[0],
	}
	if len(tokens) == 0 {
		sp.push_err(cc.Token, "missing_expr")
		return nil
	}
	if is_fn_call(tokens) == nil {
		sp.push_err(cc.Token, "expr_not_func_call")
	}

	data := get_call_data(tokens)
	if len(data.expr_tokens) == 0 {
		sp.push_err(cc.Token, "missing_expr")
		return nil
	}

	cc.Expr = sp.p.build_expr(data.expr_tokens)
	cc.Generics = sp.build_call_generics(data.generics_tokens)
	cc.Args = sp.build_args(data.args_tokens)

	return cc
}

func (sp *scope_parser) build_co_call_st(tokens []lex.Token) ast.NodeData {
	cc := sp.build_call_st(tokens)
	cc.(*ast.FnCallExpr).IsCo = true
	return cc
}

func (sp *scope_parser) build_goto_st(tokens []lex.Token) ast.NodeData {
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

func (sp *scope_parser) build_st(st *st) ast.NodeData {
	token := st.tokens[0]
	switch token.Id {
	case lex.ID_CONST, lex.ID_LET, lex.ID_MUT:
		return sp.build_var_st(st.tokens)

	case lex.ID_RET:
		return sp.build_ret_st(st.tokens)

	case lex.ID_ITER:
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
	}
	sp.push_err(token, "invalid_syntax")
	return nil
}

func (sp *scope_parser) build(tokens []lex.Token, s *ast.Scope) {
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
			sp.s.Tree = append(sp.s.Tree, data)
		}

		if sp.stopped() {
			break
		}
	}
}
