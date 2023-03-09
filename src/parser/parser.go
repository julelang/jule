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

type parser struct {
	file          *lex.File
	directives    []*ast.Directive
	comment_group *ast.CommentGroup
	tree          []ast.Node
	errors        []build.Log
}

func (p *parser) stop() { p.file = nil }
func (p *parser) stopped() bool { return p.file == nil }

// Appends error by specified token, key and args.
func (p *parser) push_err(token lex.Token, key string, args ...any) {
	p.errors = append(p.errors, compiler_err(token, key, args...))
}

func (p *parser) push_directive(token lex.Token) {
	d := &ast.Directive{
		Token: token,
		Tag:   token.Kind[len(lex.DIRECTIVE_COMMENT_PREFIX):], // Remove directive prefix
	}

	// Don't append if directive kind is invalid.
	ok := false
	for _, kind := range build.ATTRS {
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

func (p *parser) build_comment(token lex.Token) ast.NodeData {
	// Remove slashes and trim spaces.
	token.Kind = strings.TrimSpace(token.Kind[2:])

	if strings.HasPrefix(token.Kind, lex.DIRECTIVE_COMMENT_PREFIX) {
		p.push_directive(token)
	} else {
		if p.comment_group == nil {
			p.comment_group = &ast.CommentGroup{}
		}
		p.comment_group.Comments = append(p.comment_group.Comments, &ast.Comment{
			Token: token,
			Text:  token.Kind,
		})
	}

	return &ast.Comment{
		Token: token,
		Text:  token.Kind,
	}
}

func (p *parser) build_scope(tokens []lex.Token) *ast.Scope {
	s := new_scope()
	sp := scope_parser{
		p: p,
	}
	sp.build(tokens, s)
	return s
}

func (p *parser) __build_type(tokens []lex.Token, i *int, err bool) (*ast.Type, bool) {
	tb := type_builder{
		p:      p,
		tokens: tokens,
		i:      i,
		err:    err,
	}
	return tb.build()
}

// build_type builds AST model of data-type.
func (p *parser) build_type(tokens []lex.Token, i *int, err bool) (*ast.Type, bool) {
	token := tokens[*i]
	t, ok := p.__build_type(tokens, i, err)
	if err && !ok {
		p.push_err(token, "invalid_type")
	}
	return t, ok
}

func (p *parser) build_expr(tokens []lex.Token) *ast.Expr {
	// TODO: implement here
	return &ast.Expr{}
}

func (p *parser) build_var_type_and_expr(v *ast.VarDecl, tokens []lex.Token) {
	i := 0
	tok := tokens[i]
	if tok.Id == lex.ID_COLON {
		i++ // Skip type annotation operator (:)
		if i >= len(tokens) ||
		(tokens[i].Id == lex.ID_OP && tokens[i].Kind == lex.KND_EQ) {
			p.push_err(tok, "missing_type")
			return
		}
		t, ok := p.build_type(tokens, &i, false)
		if ok {
			v.DataType = t
			i++
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

func (p *parser) build_var_common(v *ast.VarDecl, tokens []lex.Token) {
	v.Token = tokens[0]
	if v.Token.Id != lex.ID_IDENT {
		p.push_err(v.Token, "invalid_syntax")
		return
	}
	v.Ident = v.Token.Kind
	v.DataType = get_void_type()
	if len(tokens) > 1 {
		tokens = tokens[1:] // Remove identifier.
		p.build_var_type_and_expr(v, tokens)
	}
}

func (p *parser) build_var_begin(v *ast.VarDecl, i *int, tokens []lex.Token) {
	tok := tokens[*i]
	switch tok.Id {
	case lex.ID_LET:
		// Initialize 1 for skip the let keyword
		*i++
		if tokens[*i].Id == lex.ID_MUT {
			v.IsMut = true
			// Skip the mut keyword
			*i++
		}
	case lex.ID_CONST:
		*i++
		if v.IsConst {
			p.push_err(tok, "already_const")
			break
		}
		v.IsConst = true
		if !v.IsMut {
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

func (p *parser) build_var(tokens []lex.Token) *ast.VarDecl {
	v := &ast.VarDecl{}
	i := 0
	v.Token = tokens[i]
	p.build_var_begin(v, &i, tokens)
	if i >= len(tokens) {
		return nil
	}
	tokens = tokens[i:]
	p.build_var_common(v, tokens)
	if v.DataType.IsVoid() && v.Expr == nil {
		p.push_err(v.Token, "missing_type")
	}
	return v
}

func (p *parser) build_generic(tokens []lex.Token) *ast.Generic {
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

func (p *parser) build_generics(tokens []lex.Token) []*ast.Generic {
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

func (p *parser) build_self_param(tokens []lex.Token) *ast.Param {
	if len(tokens) == 0 {
		return nil
	}

	param := &ast.Param{}

	// Detects mut keyword.
	i := 0
	if tokens[i].Id == lex.ID_MUT {
		param.IsMut = true
		i++
		if i >= len(tokens) {
			p.push_err(tokens[i-1], "invalid_syntax")
			return nil
		}
	}

	if tokens[i].Kind == lex.KND_AMPER {
		param.DataType.Kind = &ast.RefType{}
		i++
		if i >= len(tokens) {
			p.push_err(tokens[i-1], "invalid_syntax")
			return nil
		}
	}

	if tokens[i].Id == lex.ID_SELF {
		param.Ident = lex.KND_SELF
		param.Token = tokens[i]
		i++
		if i < len(tokens) {
			p.push_err(tokens[i+1], "invalid_syntax")
		}
	}

	return param
}

func (p *parser) param_type_begin(param *ast.Param, i *int, tokens []lex.Token) {
	for ; *i < len(tokens); *i++ {
		token := tokens[*i]
		if token.Id != lex.ID_OP {
			return
		} else if token.Kind != lex.KND_TRIPLE_DOT {
			return
		}

		if param.IsVariadic {
			p.push_err(token, "already_variadic")
			continue
		}
		param.IsVariadic = true
	}
}

func (p *parser) build_param_type(param *ast.Param, tokens []lex.Token, must_pure bool) {
	i := 0
	if !must_pure {
		p.param_type_begin(param, &i, tokens)
		if i >= len(tokens) {
			return
		}
	}
	param.DataType, _ = p.build_type(tokens, &i, true)
	i++
	if i < len(tokens) {
		p.push_err(tokens[i], "invalid_syntax")
	}
}

func (p *parser) param_body_id(param *ast.Param, token lex.Token) {
	if lex.IsIgnoreId(token.Kind) {
		param.Ident = lex.ANONYMOUS_ID
		return
	}
	param.Ident = token.Kind
}

func (p *parser) build_param_body(param *ast.Param, i *int, tokens []lex.Token, must_pure bool) {
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

func (p *parser) build_param(tokens []lex.Token, must_pure bool) *ast.Param {
	param := &ast.Param{
		Token: tokens[0],
	}

	// Detects mut keyword.
	if param.Token.Id == lex.ID_MUT {
		param.IsMut = true
		if len(tokens) == 1 {
			p.push_err(tokens[0], "invalid_syntax")
			return nil
		}
		tokens = tokens[1:]
		param.Token = tokens[0]
	}

	if param.Token.Id != lex.ID_IDENT {
		// Just data type
		param.Ident = lex.ANONYMOUS_ID
		p.build_param_type(param, tokens, must_pure)
	} else {
		i := 0
		p.build_param_body(param, &i, tokens, must_pure)
	}

	return param
}

func (p *parser) check_params(params []*ast.Param) {
	for _, param := range params {
		if param.Ident == lex.KND_SELF || param.DataType != nil {
			continue
		}
		if param.Token.Id == lex.ID_NA {
			p.push_err(param.Token, "missing_type")
		} else {
			param.DataType = &ast.Type{
				Token: param.Token,
				Kind:   &ast.IdentType{Ident: param.Token.Kind},
			}
			param.Ident = lex.ANONYMOUS_ID
			param.Token = lex.Token{}
		}
	}
}

func (p *parser) build_params(tokens []lex.Token, method bool, must_pure bool) []*ast.Param {
	parts, errs := lex.Parts(tokens, lex.ID_COMMA, true)
	p.errors = append(p.errors, errs...)
	if len(parts) == 0 {
		return nil
	}

	var params []*ast.Param
	if method && len(parts) > 0 {
		param := p.build_self_param(parts[0])
		if param.Ident == lex.KND_SELF {
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

func (p *parser) build_multi_ret_type(tokens []lex.Token, i *int) (t *ast.RetType, ok bool) {
	t = &ast.RetType{}
	*i++
	if *i >= len(tokens) {
		*i--
		t.Kind, ok = p.build_type(tokens, i, false)
		return
	}

	*i-- // For point to parenthses - ( -
	rang := lex.Range(i, lex.KND_LPAREN, lex.KND_RPARENT, tokens)
	params := p.build_params(rang, false, true)

	types := make([]*ast.Type, len(params))
	for i, param := range params {
		types[i] = param.DataType
		if param.Ident != lex.ANONYMOUS_ID {
			param.Token.Kind = param.Ident
		} else {
			param.Token.Kind = lex.IGNORE_ID
		}
		t.Idents = append(t.Idents, param.Token)
	}

	if len(types) > 1 {
		t.Kind.Token = tokens[0]
		t.Kind.Kind = &ast.MultiRetType{Types: types}
	} else {
		t.Kind = types[0]
	}

	// Decrament for correct block parsing
	*i--
	ok = true
	return
}

func (p *parser) build_ret_type(tokens []lex.Token, i *int) (t *ast.RetType, ok bool) {
	if *i >= len(tokens) {
		return nil, false
	}
	t = &ast.RetType{}

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

func (p *parser) build_fn_prototype(tokens []lex.Token, i *int, method bool, anon bool) *ast.FnDecl {
	f := &ast.FnDecl{
		Token: tokens[*i],
	}

	// Detect unsafe keyword.
	if f.Token.Id == lex.ID_UNSAFE {
		f.IsUnsafe = true
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
		f.Ident = lex.ANONYMOUS_ID
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

	t, ret_ok := p.build_ret_type(tokens, i)
	if ret_ok {
		f.RetType = t
		*i++
	}

	return f
}

func (p *parser) build_fn(tokens []lex.Token, method bool, anon bool, prototype bool) *ast.FnDecl {
	i := 0
	f := p.build_fn_prototype(tokens, &i, method, anon)
	if prototype {
		if i+1 < len(tokens) {
			p.push_err(tokens[i+1], "invalid_syntax")
		}
		return nil
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
		f.Scope.IsUnsafe = f.IsUnsafe
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

func (p *parser) build_node_data(st []lex.Token) ast.NodeData {
	token := st[0]
	switch token.Id {
	case lex.ID_CONST, lex.ID_LET, lex.ID_MUT:
		return p.build_var(st)

	case lex.ID_FN, lex.ID_UNSAFE:
		return p.build_fn(st, false, false, false)
	
	case lex.ID_COMMENT:
		// Push first token because this is full text comment.
		// Comments are just single-line.
		// Range comments not accepts by lexer.
		return p.build_comment(token)

	default:
		p.push_err(token, "invalid_syntax")
		return nil
	}
}

func (p *parser) check_comment_group(node ast.Node) {
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

func (p *parser) check_directive(node ast.Node) {
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

func (p *parser) apply_meta(node *ast.Node, is_pub bool) {
	switch node.Data.(type) {
	case *ast.VarDecl:
		v := node.Data.(*ast.VarDecl)
		v.IsPub = is_pub
		is_pub = false
		v.DocComments = p.comment_group
		p.comment_group = nil

	case *ast.FnDecl:
		f := node.Data.(*ast.FnDecl)
		f.IsPub = is_pub
		is_pub = false
		f.Directives = p.directives
		p.directives = nil
		f.DocComments = p.comment_group
		p.comment_group = nil
	}
	if is_pub {
		p.push_err(node.Token, "def_not_support_pub")
	}
}

func (p *parser) append_node(st []lex.Token) {
	if len(st) == 0 {
		return
	}

	token := st[0]

	// Detect pub keyword.
	is_pub := false
	if token.Id == lex.ID_PUB {
		is_pub = true
		st = st[1:]
		if len(st) == 0 {
			p.push_err(token, "invalid_syntax")
			return
		}
	}

	node := ast.Node{
		Token: st[0],
		Data:  p.build_node_data(st),
	}

	if node.Data == nil {
		return
	}

	p.apply_meta(&node, is_pub)
	p.check_comment_group(node)
	p.check_directive(node)
	p.tree = append(p.tree, node)
}

func (p *parser) parse() {
	stms := split_stms(p.file.Tokens())
	for _, st := range stms {
		p.append_node(st.tokens)

		if p.stopped() {
			break
		}
	}
}
