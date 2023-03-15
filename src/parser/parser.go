// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package parser

import (
	"strings"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

func compiler_err(token lex.Token, key string, args ...any) build.Log {
	return build.Log{
		Type:   build.ERR,
		Row:    token.Row,
		Column: token.Column,
		Path:   token.File.Path(),
		Text:   build.Errorf(key, args...),
	}
}

func build_comment(token lex.Token) *ast.Comment {
	// Remove slashes and trim spaces.
	token.Kind = strings.TrimSpace(token.Kind[2:])
	return &ast.Comment{
		Token: token,
		Text:  token.Kind,
	}
}

func tokstoa(tokens []lex.Token) string {
	s := ""
	for _, token := range tokens {
		s += token.Kind
	}
	return s
}

type _Parser struct {
	ast           *ast.Ast
	directives    []*ast.Directive
	comment_group *ast.CommentGroup
	errors        []build.Log
}

func (p *_Parser) stop() { p.ast = nil }
func (p *_Parser) stopped() bool { return p.ast == nil }

// Appends error by specified token, key and args.
func (p *_Parser) push_err(token lex.Token, key string, args ...any) {
	p.errors = append(p.errors, compiler_err(token, key, args...))
}

func (p *_Parser) build_expr(tokens []lex.Token) *ast.Expr {
	ep := _ExprBuilder{p: p}
	expr := ep.build_from_tokens(tokens)
	return expr
}

func (p *_Parser) push_directive(c *ast.Comment) {
	d := &ast.Directive{
		Token: c.Token,
		Tag:   c.Token.Kind[len(lex.DIRECTIVE_PREFIX):], // Remove directive prefix
	}

	// Don't append if directive kind is invalid.
	ok := false
	for _, kind := range build.DIRECTIVES {
		if d.Tag == kind {
			ok = true
			break
		}
	}
	if !ok {
		return
	}

	// Don't append if already added this directive.
	for _, pd := range p.directives {
		if d.Tag == pd.Tag {
			return
		}
	}

	p.directives = append(p.directives, d)
}

func (p *_Parser) process_comment(c *ast.Comment) {
	if c.Is_directive() {
		p.push_directive(c)
		return
	}
	if p.comment_group == nil {
		p.comment_group = &ast.CommentGroup{}
	}
	p.comment_group.Comments = append(p.comment_group.Comments, c)
}

func (p *_Parser) build_scope(tokens []lex.Token) *ast.Scope {
	s := new_scope()
	sp := _ScopeParser{
		p: p,
	}
	sp.build(tokens, s)
	return s
}

func (p *_Parser) __build_type(tokens []lex.Token, i *int, err bool) (*ast.TypeDecl, bool) {
	tb := _TypeBuilder{
		p:      p,
		tokens: tokens,
		i:      i,
		err:    err,
	}
	return tb.build()
}

// build_type builds AST model of data-type.
func (p *_Parser) build_type(tokens []lex.Token, i *int, err bool) (*ast.TypeDecl, bool) {
	token := tokens[*i]
	t, ok := p.__build_type(tokens, i, err)
	if err && !ok {
		p.push_err(token, "invalid_type")
	}
	return t, ok
}

func (p *_Parser) build_type_alias_decl(tokens []lex.Token) *ast.TypeAliasDecl {
	i := 1 // Skip "type" keyword.
	if i >= len(tokens) {
		p.push_err(tokens[i-1], "invalid_syntax")
		return nil
	}
	tad := &ast.TypeAliasDecl{
		Token: tokens[1],
		Ident: tokens[1].Kind,
	}
	token := tokens[i]
	if token.Id != lex.ID_IDENT {
		p.push_err(token, "invalid_syntax")
	}
	i++
	if i >= len(tokens) {
		p.push_err(tokens[i-1], "invalid_syntax")
		return tad
	}
	token = tokens[i]
	if token.Id != lex.ID_COLON {
		p.push_err(tokens[i-1], "invalid_syntax")
		return tad
	}
	i++
	if i >= len(tokens) {
		p.push_err(tokens[i-1], "missing_type")
		return tad
	}
	t, ok := p.build_type(tokens, &i, true)
	tad.Kind = t
	if ok && i < len(tokens) {
		p.push_err(tokens[i], "invalid_syntax")
	}
	return tad
}

func (p *_Parser) build_var_type_and_expr(v *ast.VarDecl, tokens []lex.Token) {
	i := 0
	tok := tokens[i]
	if tok.Id == lex.ID_COLON {
		i++ // Skip type annotation operator (:)
		if i >= len(tokens) ||
		(tokens[i].Id == lex.ID_OP && tokens[i].Kind == lex.KND_EQ) {
			p.push_err(tok, "missing_type")
			return
		}
		t, ok := p.build_type(tokens, &i, true)
		if ok {
			v.Kind = t
			if i >= len(tokens) {
				return
			}
			tok = tokens[i]
		}
	}

	if tok.Id == lex.ID_OP {
		if tok.Kind != lex.KND_EQ {
			p.push_err(tok, "invalid_syntax")
			return
		}
		expr_tokens := tokens[i+1:]
		if len(expr_tokens) == 0 {
			p.push_err(tok, "missing_expr")
			return
		}
		v.Expr = p.build_expr(expr_tokens)
	} else {
		p.push_err(tok, "invalid_syntax")
	}
}

func (p *_Parser) build_var_common(v *ast.VarDecl, tokens []lex.Token) {
	v.Token = tokens[0]
	if v.Token.Id != lex.ID_IDENT {
		p.push_err(v.Token, "invalid_syntax")
		return
	}
	v.Ident = v.Token.Kind
	v.Kind = build_void_type()
	if len(tokens) > 1 {
		tokens = tokens[1:] // Remove identifier.
		p.build_var_type_and_expr(v, tokens)
	}
}

func (p *_Parser) build_var_begin(v *ast.VarDecl, i *int, tokens []lex.Token) {
	tok := tokens[*i]
	switch tok.Id {
	case lex.ID_LET:
		// Initialize 1 for skip the let keyword
		*i++
		if tokens[*i].Id == lex.ID_MUT {
			v.Mutable = true
			// Skip the mut keyword
			*i++
		}
	case lex.ID_CONST:
		*i++
		if v.Constant {
			p.push_err(tok, "already_const")
			break
		}
		v.Constant = true
		if !v.Mutable {
			break
		}
		fallthrough
	default:
		p.push_err(tok, "invalid_syntax")
		return
	}
	if *i >= len(tokens) {
		p.push_err(tok, "invalid_syntax")
	}
}

func (p *_Parser) build_var(tokens []lex.Token) *ast.VarDecl {
	i := 0
	v := &ast.VarDecl{
		Token: tokens[i],
	}
	p.build_var_begin(v, &i, tokens)
	if i >= len(tokens) {
		return nil
	}
	tokens = tokens[i:]
	p.build_var_common(v, tokens)
	return v
}

func (p *_Parser) build_generic(tokens []lex.Token) *ast.Generic {
	if len(tokens) > 1 {
		p.push_err(tokens[1], "invalid_syntax")
	}
	g := &ast.Generic{
		Token: tokens[0],
	}
	if g.Token.Id != lex.ID_IDENT {
		p.push_err(g.Token, "invalid_syntax")
	}
	g.Ident = g.Token.Kind
	return g
}

func (p *_Parser) build_generics(tokens []lex.Token) []*ast.Generic {
	token := tokens[0]
	if len(tokens) == 0 {
		p.push_err(token, "missing_expr")
		return nil
	}

	parts, errors := lex.Parts(tokens, lex.ID_COMMA, true)
	p.errors = append(p.errors, errors...)

	generics := make([]*ast.Generic, len(parts))
	for i, part := range parts {
		if len(parts) > 0 {
			generics[i] = p.build_generic(part)
		}
	}

	return generics
}

func (p *_Parser) build_self_param(tokens []lex.Token) *ast.Param {
	if len(tokens) == 0 {
		return nil
	}

	param := &ast.Param{}

	// Detects mut keyword.
	i := 0
	if tokens[i].Id == lex.ID_MUT {
		param.Mutable = true
		i++
		if i >= len(tokens) {
			p.push_err(tokens[i-1], "invalid_syntax")
			return nil
		}
	}

	if tokens[i].Kind == lex.KND_AMPER {
		param.Ident = lex.KND_AMPER
		i++
		if i >= len(tokens) {
			p.push_err(tokens[i-1], "invalid_syntax")
			return nil
		}
	}

	if tokens[i].Id == lex.ID_SELF {
		param.Ident += lex.KND_SELF
		param.Token = tokens[i]
		i++
		if i < len(tokens) {
			p.push_err(tokens[i+1], "invalid_syntax")
		}
	}

	return param
}

func (p *_Parser) param_type_begin(param *ast.Param, i *int, tokens []lex.Token) {
	for ; *i < len(tokens); *i++ {
		token := tokens[*i]
		if token.Id != lex.ID_OP {
			return
		} else if token.Kind != lex.KND_TRIPLE_DOT {
			return
		}

		if param.Variadic {
			p.push_err(token, "already_variadic")
			continue
		}
		param.Variadic = true
	}
}

func (p *_Parser) build_param_type(param *ast.Param, tokens []lex.Token, must_pure bool) {
	i := 0
	if !must_pure {
		p.param_type_begin(param, &i, tokens)
		if i >= len(tokens) {
			return
		}
	}
	param.Kind, _ = p.build_type(tokens, &i, true)
	if i < len(tokens) {
		p.push_err(tokens[i], "invalid_syntax")
	}
}

func (p *_Parser) param_body_id(param *ast.Param, token lex.Token) {
	if lex.Is_ignore_ident(token.Kind) {
		param.Ident = lex.ANON_IDENT
		return
	}
	param.Ident = token.Kind
}

func (p *_Parser) build_param_body(param *ast.Param, i *int, tokens []lex.Token, must_pure bool) {
	p.param_body_id(param, tokens[*i])
	tok := tokens[*i]
	// +1 for skip identifier token
	tokens = tokens[*i+1:]
	if len(tokens) == 0 {
		return
	} else if len(tokens) < 2 {
		p.push_err(tok, "missing_type")
		return
	}

	tok = tokens[*i]
	if tok.Id != lex.ID_COLON {
		p.push_err(tok, "invalid_syntax")
		return
	}

	tokens = tokens[*i+1:] // Skip colon
	p.build_param_type(param, tokens, must_pure)
}

func (p *_Parser) build_param(tokens []lex.Token, must_pure bool) *ast.Param {
	param := &ast.Param{
		Token: tokens[0],
	}

	// Detects mut keyword.
	if param.Token.Id == lex.ID_MUT {
		param.Mutable = true
		if len(tokens) == 1 {
			p.push_err(tokens[0], "invalid_syntax")
			return nil
		}
		tokens = tokens[1:]
		param.Token = tokens[0]
	}

	if param.Token.Id != lex.ID_IDENT {
		// Just data type
		param.Ident = lex.ANON_IDENT
		p.build_param_type(param, tokens, must_pure)
	} else {
		i := 0
		p.build_param_body(param, &i, tokens, must_pure)
	}

	return param
}

func (p *_Parser) check_params(params []*ast.Param) {
	for _, param := range params {
		if param.Is_self() || param.Kind != nil {
			continue
		}
		if param.Token.Id == lex.ID_NA {
			p.push_err(param.Token, "missing_type")
		} else {
			param.Kind = &ast.TypeDecl{
				Token: param.Token,
				Kind:   &ast.IdentType{
					Token: param.Token,
					Ident: param.Token.Kind,
				},
			}
			param.Ident = lex.ANON_IDENT
			param.Token = lex.Token{}
		}
	}
}

func (p *_Parser) build_params(tokens []lex.Token, method bool, must_pure bool) []*ast.Param {
	parts, errs := lex.Parts(tokens, lex.ID_COMMA, true)
	p.errors = append(p.errors, errs...)
	if len(parts) == 0 {
		return nil
	}

	var params []*ast.Param
	if method && len(parts) > 0 {
		param := p.build_self_param(parts[0])
		if param != nil && param.Is_self() {
			params = append(params, param)
			parts = parts[1:]
		}
	}

	for _, part := range parts {
		param := p.build_param(part, must_pure)
		if param != nil {
			params = append(params, param)
		}
	}

	p.check_params(params)
	return params
}

func (p *_Parser) build_multi_ret_type(tokens []lex.Token, i *int) (t *ast.RetType, ok bool) {
	t = &ast.RetType{}
	*i++
	if *i >= len(tokens) {
		*i--
		t.Kind, ok = p.build_type(tokens, i, false)
		return
	}

	*i-- // For point to parenthses - ( -
	range_tokens := lex.Range(i, lex.KND_LPAREN, lex.KND_RPARENT, tokens)
	params := p.build_params(range_tokens, false, true)

	types := make([]*ast.TypeDecl, len(params))
	for i, param := range params {
		types[i] = param.Kind
		if param.Ident != lex.ANON_IDENT {
			param.Token.Kind = param.Ident
		} else {
			param.Token.Kind = lex.IGNORE_IDENT
		}
		t.Idents = append(t.Idents, param.Token)
	}

	if len(types) > 1 {
		t.Kind = &ast.TypeDecl{
			Token: tokens[0],
			Kind:  &ast.TupleType{
				Types: types,
			},
		}
	} else {
		t.Kind = types[0]
	}

	ok = true
	return
}

func (p *_Parser) build_ret_type(tokens []lex.Token, i *int) (t *ast.RetType, ok bool) {
	t = &ast.RetType{}
	if *i >= len(tokens) {
		return
	}

	token := tokens[*i]
	switch token.Id {
	case lex.ID_RANGE:
		if token.Kind == lex.KND_LBRACE {
			return
		}
	case lex.ID_OP:
		if token.Kind == lex.KND_EQ {
			return
		}
	case lex.ID_COLON:
		if *i+1 >= len(tokens) {
			p.push_err(token, "missing_type")
			return
		}
		*i++
		token = tokens[*i]
		if token.Id == lex.ID_RANGE {
			switch token.Kind {
			case lex.KND_LPAREN:
				return p.build_multi_ret_type(tokens, i)
			case lex.KND_LBRACE:
				return
			}
		}
		t.Kind, ok = p.build_type(tokens, i, true)
		return
	}
	*i++
	p.push_err(token, "invalid_syntax")
	return
}

func (p *_Parser) build_fn_prototype(tokens []lex.Token, i *int, method bool, anon bool) *ast.FnDecl {
	f := &ast.FnDecl{
		Token: tokens[*i],
	}

	// Detect unsafe keyword.
	if f.Token.Id == lex.ID_UNSAFE {
		f.Unsafety = true
		*i++
		if *i >= len(tokens) {
			p.push_err(f.Token, "invalid_syntax")
			return nil
		}
		f.Token = tokens[*i]
	}

	// Skips fn tok
	*i++
	if *i >= len(tokens) {
		p.push_err(f.Token, "invalid_syntax")
		return nil
	}

	if anon {
		f.Ident = lex.ANON_IDENT
	} else {
		tok := tokens[*i]
		if tok.Id != lex.ID_IDENT {
			p.push_err(tok, "invalid_syntax")
			return nil
		}
		f.Ident = tok.Kind
		*i++
	}

	if *i >= len(tokens) {
		p.push_err(f.Token, "invalid_syntax")
		return nil
	}

	generics_tokens := lex.Range(i, lex.KND_LBRACKET, lex.KND_RBRACKET, tokens)
	if generics_tokens != nil {
		f.Generics = p.build_generics(generics_tokens)
	}

	if tokens[*i].Kind != lex.KND_LPAREN {
		p.push_err(tokens[*i], "missing_function_parentheses")
		return nil
	}

	params_toks := lex.Range(i, lex.KND_LPAREN, lex.KND_RPARENT, tokens)
	if len(params_toks) > 0 {
		f.Params = p.build_params(params_toks, method, false)
	}

	f.Result, _ = p.build_ret_type(tokens, i)
	return f
}

func (p *_Parser) build_fn(tokens []lex.Token, method bool, anon bool, prototype bool) *ast.FnDecl {
	i := 0
	f := p.build_fn_prototype(tokens, &i, method, anon)
	if prototype {
		if i < len(tokens) {
			p.push_err(tokens[i+1], "invalid_syntax")
		}
		return f
	} else if f == nil {
		return nil
	}

	if i >= len(tokens) {
		p.stop()
		p.push_err(f.Token, "body_not_exist")
		return nil
	}
	block_tokens := lex.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, tokens)
	if block_tokens != nil {
		f.Scope = p.build_scope(block_tokens)
		f.Scope.Unsafety = f.Unsafety
		if i < len(tokens) {
			p.push_err(tokens[i], "invalid_syntax")
		}
	} else {
		p.stop()
		p.push_err(f.Token, "body_not_exist")
		return nil
	}
	return f
}

func (p *_Parser) get_use_decl_selectors(tokens []lex.Token) []lex.Token {
	i := 0
	tokens = lex.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, tokens)
	parts, errs := lex.Parts(tokens, lex.ID_COMMA, true)
	if len(errs) > 0 {
		p.errors = append(p.errors, errs...)
		return nil
	}
	selectors := make([]lex.Token, len(parts))
	for i, part := range parts {
		if len(part) > 1 {
			p.push_err(part[1], "invalid_syntax")
		}
		tok := part[0]
		if tok.Id != lex.ID_IDENT && tok.Id != lex.ID_SELF {
			p.push_err(tok, "invalid_syntax")
			continue
		}
		selectors[i] = tok
	}
	return selectors
}

func (p *_Parser) build_cpp_use_decl(decl *ast.UseDecl, tokens []lex.Token) {
	if len(tokens) > 2 {
		p.push_err(tokens[2], "invalid_syntax")
	}
	token := tokens[1]
	if token.Id != lex.ID_LIT || (token.Kind[0] != '`' && token.Kind[0] != '"') {
		p.push_err(token, "invalid_expr")
		return
	}
	decl.Cpp = true
	decl.Link_path = token.Kind[1 : len(token.Kind)-1]
}

func (p *_Parser) build_std_use_decl(decl *ast.UseDecl, tokens []lex.Token) {
	decl.Std = true

	token := tokens[0]
	if len(tokens) < 3 {
		p.push_err(token, "invalid_syntax")
		return
	}

	tokens = tokens[2:]
	token = tokens[len(tokens)-1]
	switch token.Id {
	case lex.ID_DBLCOLON:
		p.push_err(token, "invalid_syntax")
		return

	case lex.ID_RANGE:
		if token.Kind != lex.KND_RBRACE {
			p.push_err(token, "invalid_syntax")
			return
		}
		var selectors []lex.Token
		tokens, selectors = lex.Range_last(tokens)
		decl.Selected = p.get_use_decl_selectors(selectors)
		if len(tokens) == 0 {
			p.push_err(token, "invalid_syntax")
			return
		}
		token = tokens[len(tokens)-1]
		if token.Id != lex.ID_DBLCOLON {
			p.push_err(token, "invalid_syntax")
			return
		}
		tokens = tokens[:len(tokens)-1]
		if len(tokens) == 0 {
			p.push_err(token, "invalid_syntax")
			return
		}

	case lex.ID_OP:
		if token.Kind != lex.KND_STAR {
			p.push_err(token, "invalid_syntax")
			return
		}
		tokens = tokens[:len(tokens)-1]
		if len(tokens) == 0 {
			p.push_err(token, "invalid_syntax")
			return
		}
		token = tokens[len(tokens)-1]
		if token.Id != lex.ID_DBLCOLON {
			p.push_err(token, "invalid_syntax")
			return
		}
		tokens = tokens[:len(tokens)-1]
		if len(tokens) == 0 {
			p.push_err(token, "invalid_syntax")
			return
		}
		decl.Full = true
	}
	decl.Link_path = "std::" + tokstoa(tokens)
}

func (p *_Parser) parse_use_decl(decl *ast.UseDecl, tokens []lex.Token) {
	token := tokens[0]
	if token.Id == lex.ID_CPP {
		p.build_cpp_use_decl(decl, tokens)
		return
	}
	if token.Id != lex.ID_IDENT || token.Kind != "std" {
		p.push_err(token, "invalid_syntax")
		return
	}
	p.build_std_use_decl(decl, tokens)
}

func (p *_Parser) build_use_decl(tokens []lex.Token) *ast.UseDecl {
	decl := &ast.UseDecl{
		Token: tokens[0],
	}
	if len(tokens) < 2 {
		p.push_err(decl.Token, "missing_use_path")
		return nil
	}
	tokens = tokens[1:] // Skip "use" keyword.
	p.parse_use_decl(decl, tokens)
	return decl
}

func (p *_Parser) build_enum_item_expr(i *int, tokens []lex.Token) *ast.Expr {
	brace_n := 0
	expr_start := *i
	for ; *i < len(tokens); *i++ {
		t := tokens[*i]
		if t.Id == lex.ID_RANGE {
			switch t.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				brace_n++
				continue
			default:
				brace_n--
			}
		}
		if brace_n > 0 {
			continue
		}
		if t.Id == lex.ID_COMMA || *i+1 >= len(tokens) {
			var expr_tokens []lex.Token
			if t.Id == lex.ID_COMMA {
				expr_tokens = tokens[expr_start:*i]
			} else {
				expr_tokens = tokens[expr_start:]
			}
			return p.build_expr(expr_tokens)
		}
	}
	return nil
}

func (p *_Parser) build_enum_items(tokens []lex.Token) []*ast.EnumItem {
	items := make([]*ast.EnumItem, 0)
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		if t.Id == lex.ID_COMMENT {
			continue
		}
		item := new(ast.EnumItem)
		item.Token = t
		if item.Token.Id != lex.ID_IDENT {
			p.push_err(item.Token, "invalid_syntax")
		}
		item.Ident = item.Token.Kind
		if i+1 >= len(tokens) || tokens[i+1].Id == lex.ID_COMMA {
			if i+1 < len(tokens) {
				i++
			}
			items = append(items, item)
			continue
		}
		i++
		t = tokens[i]
		if t.Id != lex.ID_OP && t.Kind != lex.KND_EQ {
			p.push_err(tokens[0], "invalid_syntax")
		}
		i++
		if i >= len(tokens) || tokens[i].Id == lex.ID_COMMA {
			p.push_err(tokens[0], "missing_expr")
			continue
		}
		item.Expr = p.build_enum_item_expr(&i, tokens)
		items = append(items, item)
	}
	return items
}

func (p *_Parser) build_enum_decl(tokens []lex.Token) *ast.EnumDecl {
	if len(tokens) < 2 || len(tokens) < 3 {
		p.push_err(tokens[0], "invalid_syntax")
		return nil
	}
	e := &ast.EnumDecl{
		Token: tokens[1],
	}
	if e.Token.Id != lex.ID_IDENT {
		p.push_err(e.Token, "invalid_syntax")
	}
	e.Ident = e.Token.Kind
	i := 2
	if tokens[i].Id == lex.ID_COLON {
		i++
		if i >= len(tokens) {
			p.push_err(tokens[i-1], "invalid_syntax")
			return e
		}
		e.Kind, _ = p.build_type(tokens, &i, true)
		if i >= len(tokens) {
			p.stop()
			p.push_err(e.Token, "body_not_exist")
			return e
		}
	} else {
		e.Kind = nil
	}
	item_tokens := lex.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, tokens)
	if item_tokens == nil {
		p.stop()
		p.push_err(e.Token, "body_not_exist")
		return e
	} else if i < len(tokens) {
		p.push_err(tokens[i], "invalid_syntax")
	}
	e.Items = p.build_enum_items(item_tokens)
	return e
}

func (p *_Parser) build_field(tokens []lex.Token) *ast.FieldDecl {
	f := &ast.FieldDecl{}

	f.Public = tokens[0].Id == lex.ID_PUB
	if f.Public {
		if len(tokens) == 1 {
			p.push_err(tokens[0], "invalid_syntax")
			return nil
		}
		tokens = tokens[1:]
	}

	f.Mutable = tokens[0].Id == lex.ID_MUT
	if f.Mutable {
		if len(tokens) == 1 {
			p.push_err(tokens[0], "invalid_syntax")
			return nil
		}
		tokens = tokens[1:]
	}

	f.Token = tokens[0]
	if f.Token.Id != lex.ID_IDENT {
		p.push_err(f.Token, "invalid_syntax")
		return nil
	}
	f.Ident = f.Token.Kind

	if len(tokens) == 1 {
		p.push_err(tokens[0], "missing_type")
		return nil
	} else if tokens[1].Id != lex.ID_COLON {
		p.push_err(tokens[1], "missing_type")
		return nil
	}

	tokens = tokens[2:] // Remove identifier and colon tokens.
	i := 0
	f.Kind, _ = p.build_type(tokens, &i, true)
	if i < len(tokens) {
		p.push_err(tokens[i], "invalid_syntax")
		return nil
	}

	return f
}

func (p *_Parser) build_struct_decl_fields(tokens []lex.Token) []*ast.FieldDecl {
	var fields []*ast.FieldDecl
	stms := split_stms(tokens)
	for _, st := range stms {
		tokens := st.tokens
		if tokens[0].Id == lex.ID_COMMENT {
			continue
		}
		f := p.build_field(tokens)
		fields = append(fields, f)
	}
	return fields
}

func (p *_Parser) build_struct_decl(tokens []lex.Token) *ast.StructDecl {
	if len(tokens) < 3 {
		p.push_err(tokens[0], "invalid_syntax")
		return nil
	}
	
	i := 1
	s := &ast.StructDecl{
		Token: tokens[i],
	}
	if s.Token.Id != lex.ID_IDENT {
		p.push_err(s.Token, "invalid_syntax")
	}
	i++
	if i >= len(tokens) {
		p.push_err(tokens[i], "invalid_syntax")
		return s
	}
	s.Ident = s.Token.Kind

	generics_tokens := lex.Range(&i, lex.KND_LBRACKET, lex.KND_RBRACKET, tokens)
	if generics_tokens != nil {
		s.Generics = p.build_generics(generics_tokens)
	}
	if i >= len(tokens) {
		p.push_err(tokens[i], "invalid_syntax")
		return s
	}

	body_tokens := lex.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, tokens)
	if body_tokens == nil {
		p.stop()
		p.push_err(s.Token, "body_not_exist")
		return s
	}
	if i < len(tokens) {
		p.push_err(tokens[i], "invalid_syntax")
	}
	s.Fields = p.build_struct_decl_fields(body_tokens)
	return s
}

func (p *_Parser) check_method_receiver(f *ast.FnDecl) {
	if len(f.Params) == 0 {
		p.push_err(f.Token, "missing_receiver")
		return
	}
	param := f.Params[0]
	if !param.Is_self() {
		p.push_err(f.Token, "missing_receiver")
		return
	}
}

func (p *_Parser) build_trait_methods(tokens []lex.Token) []*ast.FnDecl {
	var methods []*ast.FnDecl
	stms := split_stms(tokens)
	for _, st := range stms {
		tokens := st.tokens
		f := p.build_fn(tokens, true, false, true)
		if f != nil {
			p.check_method_receiver(f)
			f.Public = true
			methods = append(methods, f)
		}
	}
	return methods
}

func (p *_Parser) build_trait_decl(tokens []lex.Token) *ast.TraitDecl {
	if len(tokens) < 3 {
		p.push_err(tokens[0], "invalid_syntax")
		return nil
	}
	t := &ast.TraitDecl{
		Token: tokens[1],
	}
	if t.Token.Id != lex.ID_IDENT {
		p.push_err(t.Token, "invalid_syntax")
	}
	t.Ident = t.Token.Kind
	i := 2
	body_tokens := lex.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, tokens)
	if body_tokens == nil {
		p.stop()
		p.push_err(t.Token, "body_not_exist")
		return nil
	}
	if i < len(tokens) {
		p.push_err(tokens[i], "invalid_syntax")
	}
	t.Methods = p.build_trait_methods(body_tokens)
	return t
}

func (p *_Parser) build_cpp_link_fn(tokens []lex.Token) *ast.FnDecl {
	tokens = tokens[1:] // Remove "cpp" keyword.
	f := p.build_fn(tokens, false, false, true)
	if f != nil {
		f.Cpp_linked = true
	}
	return f
}

func (p *_Parser) build_cpp_link_var(tokens []lex.Token) *ast.VarDecl {
	tokens = tokens[1:] // Remove "cpp" keyword.
	v := p.build_var(tokens)
	if v != nil {
		v.Cpp_linked = true
		if v.Expr != nil {
			p.push_err(v.Token, "cpp_linked_variable_has_expr")
		}
		if v.Constant {
			p.push_err(v.Token, "cpp_linked_variable_is_const")
		}
	}
	return v
}

func (p *_Parser) build_cpp_link_struct(tokens []lex.Token) *ast.StructDecl {
	tokens = tokens[1:] // Remove "cpp" keyword.
	s := p.build_struct_decl(tokens)
	if s != nil {
		s.Cpp_linked = true
	}
	return s
}

func (p *_Parser) build_cpp_link_type_alias(tokens []lex.Token) *ast.TypeAliasDecl {
	tokens = tokens[1:] // Remove "cpp" keyword.
	t := p.build_type_alias_decl(tokens)
	if t != nil {
		t.Cpp_linked = true
	}
	return t
}

func (p *_Parser) build_cpp_link(tokens []lex.Token) ast.NodeData {
	token := tokens[0]
	if len(tokens) == 1 {
		p.push_err(token, "invalid_syntax")
		return nil
	}
	token = tokens[1]
	switch token.Id {
	case lex.ID_FN, lex.ID_UNSAFE:
		return p.build_cpp_link_fn(tokens)
	case lex.ID_CONST, lex.ID_LET:
		return p.build_cpp_link_var(tokens)
	case lex.ID_STRUCT:
		return p.build_cpp_link_struct(tokens)
	case lex.ID_TYPE:
		return p.build_cpp_link_type_alias(tokens)
	default:
		p.push_err(token, "invalid_syntax")
	}
	return nil
}

func (p *_Parser) get_method(tokens []lex.Token) *ast.FnDecl {
	token := tokens[0]
	if token.Id == lex.ID_UNSAFE {
		if len(tokens) == 1 || tokens[1].Id != lex.ID_FN {
			p.push_err(token, "invalid_syntax")
			return nil
		}
	} else if tokens[0].Id != lex.ID_FN {
		p.push_err(token, "invalid_syntax")
		return nil
	}
	return p.build_fn(tokens, true, false, false)
}

func (p *_Parser) parse_impl_trait(ipl *ast.Impl, tokens []lex.Token) {
	stms := split_stms(tokens)
	for _, st := range stms {
		tokens := st.tokens
		token := tokens[0]
		switch token.Id {
		case lex.ID_COMMENT:
			// Ignore.
			continue

		case lex.ID_FN, lex.ID_UNSAFE:
			f := p.get_method(tokens)
			if f != nil {
				f.Public = true
				p.check_method_receiver(f)
				ipl.Methods = append(ipl.Methods, f)
			}

		default:
			p.push_err(token, "invalid_syntax")
			continue
		}
	}
}

func (p *_Parser) parse_impl_struct(ipl *ast.Impl, tokens []lex.Token) {
	stms := split_stms(tokens)
	for _, st := range stms {
		tokens := st.tokens
		token := tokens[0]
		is_pub := false
		switch token.Id {
		case lex.ID_COMMENT:
			// Ignore.
			continue

		case lex.ID_PUB:
			is_pub = true
			if len(tokens) == 1 {
				p.push_err(tokens[0], "invalid_syntax")
				continue
			}
			tokens = tokens[1:]
			if len(tokens) > 0 {
				token = tokens[0]
			}
		}

		switch token.Id {
		case lex.ID_FN, lex.ID_UNSAFE:
			f := p.get_method(tokens)
			if f != nil {
				f.Public = is_pub
				p.check_method_receiver(f)
				ipl.Methods = append(ipl.Methods, f)
			}

		default:
			p.push_err(token, "invalid_syntax")
			continue
		}
	}
}

func (p *_Parser) parse_impl_body(ipl *ast.Impl, tokens []lex.Token) {
	if ipl.Is_trait_impl() {
		p.parse_impl_trait(ipl, tokens)
		return
	}
	p.parse_impl_struct(ipl, tokens)
}

func (p *_Parser) build_impl(tokens []lex.Token) *ast.Impl {
	token := tokens[0]
	if len(tokens) < 2 {
		p.push_err(token, "invalid_syntax")
		return nil
	}
	token = tokens[1]
	if token.Id != lex.ID_IDENT {
		p.push_err(token, "invalid_syntax")
		return nil
	}
	if len(tokens) < 3 {
		p.push_err(token, "invalid_syntax")
		return nil
	}
	ipl := &ast.Impl{
		Base: token,
	}
	token = tokens[2]
	if token.Id != lex.ID_FOR {
		if token.Id == lex.ID_RANGE && token.Kind == lex.KND_LBRACE {
			// This implementation is single.
			// Just implements to destination.
			// Therefore, swap Base and Dest tokens.
			ipl.Base, ipl.Dest = ipl.Dest, ipl.Base

			tokens = tokens[2:]  // Remove prefix tokens.
			goto body
		}
		p.push_err(token, "invalid_syntax")
		return nil
	}
	if len(tokens) < 4 {
		p.push_err(token, "invalid_syntax")
		return nil
	}
	token = tokens[3]
	if token.Id != lex.ID_IDENT {
		p.push_err(token, "invalid_syntax")
		return nil
	}
	ipl.Dest = token
	tokens = tokens[4:] // Remove prefix tokens.
body:
	i := 0
	body_tokens := lex.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, tokens)
	if body_tokens == nil {
		p.stop()
		p.push_err(ipl.Base, "body_not_exist")
		return nil
	}
	if i < len(tokens) {
		p.push_err(tokens[i], "invalid_syntax")
	}
	p.parse_impl_body(ipl, body_tokens)
	return ipl
}

func (p *_Parser) build_node_data(tokens []lex.Token) ast.NodeData {
	token := tokens[0]
	switch token.Id {
	case lex.ID_USE:
		return p.build_use_decl(tokens)
		
	case lex.ID_FN, lex.ID_UNSAFE:
		return p.build_fn(tokens, false, false, false)

	case lex.ID_CONST, lex.ID_LET, lex.ID_MUT:
		return p.build_var(tokens)
	
	case lex.ID_TYPE:
		return p.build_type_alias_decl(tokens)

	case lex.ID_ENUM:
		return p.build_enum_decl(tokens)

	case lex.ID_STRUCT:
		return p.build_struct_decl(tokens)
	
	case lex.ID_TRAIT:
		return p.build_trait_decl(tokens)

	case lex.ID_IMPL:
		return p.build_impl(tokens)

	case lex.ID_CPP:
		return p.build_cpp_link(tokens)

	case lex.ID_COMMENT:
		// Push first token because this is full text comment.
		// Comments are just single-line.
		// Range comments not accepts by lexer.
		c := build_comment(token)
		p.process_comment(c)
		return c

	default:
		p.push_err(token, "invalid_syntax")
		return nil
	}
}

func (p *_Parser) check_comment_group(node ast.Node) {
	if p.comment_group == nil {
		return
	}
	switch node.Data.(type) {
	case ast.Comment, ast.Directive:
		// Ignore
	default:
		p.comment_group = nil
	}
}

func (p *_Parser) check_directive(node ast.Node) {
	if p.directives == nil {
		return
	}
	switch node.Data.(type) {
	case ast.Directive, ast.Comment:
		// Ignore
	default:
		p.directives = nil
	}
}

func (p *_Parser) apply_meta(node ast.Node, is_pub bool) {
	switch node.Data.(type) {
	case *ast.VarDecl:
		v := node.Data.(*ast.VarDecl)
		if v == nil {
			return
		}
		v.Public = is_pub
		v.Doc_comments = p.comment_group
		is_pub = false
		p.comment_group = nil

	case *ast.FnDecl:
		f := node.Data.(*ast.FnDecl)
		if f == nil {
			return
		}
		f.Public = is_pub
		is_pub = false
		f.Directives = p.directives
		p.directives = nil
		f.Doc_comments = p.comment_group
		p.comment_group = nil

	case *ast.TypeAliasDecl:
		tad := node.Data.(*ast.TypeAliasDecl)
		if tad == nil {
			return
		}
		tad.Public = is_pub
		is_pub = false
		tad.Doc_comments = p.comment_group
		p.comment_group = nil

	case *ast.EnumDecl:
		ed := node.Data.(*ast.EnumDecl)
		if ed == nil {
			return
		}
		ed.Doc_comments = p.comment_group
		p.comment_group = nil
		ed.Public = is_pub
		is_pub = false

	case *ast.StructDecl:
		sd := node.Data.(*ast.StructDecl)
		if sd == nil {
			return
		}
		sd.Directives = p.directives
		p.directives = nil
		sd.Doc_comments = p.comment_group
		p.comment_group = nil
		sd.Public = is_pub
		is_pub = false

	case *ast.TraitDecl:
		td := node.Data.(*ast.TraitDecl)
		if td == nil {
			return
		}
		td.Doc_comments = p.comment_group
		p.comment_group = nil
		td.Public = is_pub
		is_pub = false
	}
	if is_pub {
		p.push_err(node.Token, "def_not_support_pub")
	}
}

func (p *_Parser) check_use_decl(node ast.Node) {
	switch node.Data.(type) {
	case *ast.UseDecl:
		// Ignore.
	default:
		return
	}

	if len(p.ast.Decls) > 0 {
		p.push_err(node.Token, "use_decl_at_body")
	}
}

func (p *_Parser) parse_node(st []lex.Token) ast.Node {
	node := ast.Node{
		Token: st[0],
	}

	// Detect pub keyword.
	is_pub := false
	if node.Token.Id == lex.ID_PUB {
		is_pub = true
		st = st[1:]
		if len(st) == 0 {
			p.push_err(node.Token, "invalid_syntax")
			return node
		}
	}

	node.Data = p.build_node_data(st)
	if node.Data == nil {
		return node
	}

	p.apply_meta(node, is_pub)
	p.check_comment_group(node)
	p.check_directive(node)
	p.check_use_decl(node)
	return node
}

func (p *_Parser) append_node(st []lex.Token) {
	if len(st) == 0 {
		return
	}

	node := p.parse_node(st)
	if node.Data == nil || p.stopped() {
		return
	}

	switch {
	case node.Is_use_decl():
		p.ast.UseDecls = append(p.ast.UseDecls, node.Data.(*ast.UseDecl))

	case node.Is_decl():
		// Use declarations eliminated.
		p.ast.Decls = append(p.ast.Decls, node)

	case node.Is_comment():
		// Global scope is not appends *CommentGroup.
		c := node.Data.(*ast.Comment)
		p.ast.Comments = append(p.ast.Comments, c)

	case node.Is_impl():
		p.ast.Impls = append(p.ast.Impls, node.Data.(*ast.Impl))
	
	default:
		p.push_err(node.Token, "invalid_syntax")
	}
}

func (p *_Parser) parse(f *lex.File) {
	p.ast = &ast.Ast{
		File: f,
	}
	stms := split_stms(f.Tokens())
	for _, st := range stms {
		p.append_node(st.tokens)

		if p.stopped() {
			break
		}
	}
}
