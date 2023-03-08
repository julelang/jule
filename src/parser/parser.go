package parser

// builder is the AST (Abstract Syntax Tree) builder of Parser.

import (
	"os"
	"strings"

	"github.com/julelang/jule"
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

func compiler_err(t lex.Token, key string, args ...any) build.Log {
	return build.Log{
		Type:   build.ERR,
		Row:    t.Row,
		Column: t.Column,
		Path:   t.File.Path(),
		Text:   build.Errorf(key, args...),
	}
}

type block_st struct {
	pos        int
	block      *ast.Block
	srcToks    *[]lex.Token
	toks       []lex.Token
	nextToks   []lex.Token
	terminated bool
}

type parser struct {
	attributes    []ast.Attribute
	comment_group *ast.CommentGroup
	public        bool
	pos           int
	tree          []ast.Node
	errors        []build.Log
	tokens        []lex.Token
}

func new_parser(tokens []lex.Token) *parser {
	return &parser{
		tokens: tokens,
		pos:    0,
	}
}

// push_err appends error by specified token.
func (p *parser) push_err(t lex.Token, key string, args ...any) {
	p.errors = append(p.errors, compiler_err(t, key, args...))
}

// ended reports position is at end of tokens or not.
func (p *parser) ended() bool {
	return p.pos >= len(p.tokens)
}

func (p *parser) build_node(toks []lex.Token) {
	t := toks[0]
	switch t.Id {
	case lex.ID_USE:
		p.build_use(toks)
	case lex.ID_FN, lex.ID_UNSAFE:
		s := ast.St{Token: t}
		f := p.build_fn(toks, false, false, false)
		f.Attributes = p.attributes
		p.attributes = nil
		f.Doc = p.comment_group
		p.comment_group = nil
		s.Data = f
		p.tree = append(p.tree, ast.Node{Token: s.Token, Data: s})
	case lex.ID_CONST, lex.ID_LET, lex.ID_MUT:
		p.build_global_var(toks)
	case lex.ID_TYPE:
		p.tree = append(p.tree, p.build_global_type_alias(toks))
	case lex.ID_ENUM:
		p.build_enum(toks)
	case lex.ID_STRUCT:
		p.build_structure_node(toks)
	case lex.ID_TRAIT:
		p.build_trait(toks)
	case lex.ID_IMPL:
		p.build_impl(toks)
	case lex.ID_CPP:
		p.cpp_link(toks)
	case lex.ID_COMMENT:
		p.tree = append(p.tree, p.comment(toks[0]))
	default:
		p.push_err(t, "invalid_syntax")
		return
	}
	if p.public {
		p.push_err(t, "def_not_support_pub")
	}
	last_node := p.tree[len(p.tree)-1]
	p.check_doc(last_node)
	p.check_attribute(last_node)
}

func (p *parser) check_doc(node ast.Node) {
	if p.comment_group == nil {
		return
	}
	switch node.Data.(type) {
	case ast.Comment, ast.Attribute, []ast.GenericType:
		return
	}
	p.comment_group = nil
}

func (p *parser) check_attribute(node ast.Node) {
	if p.attributes == nil {
		return
	}
	switch node.Data.(type) {
	case ast.Attribute, ast.Comment, []ast.GenericType:
		return
	}
	p.attributes = nil
}

// build builds AST tree.
func (p *parser) build() {
	for p.pos != -1 && !p.ended() {
		toks := p.next_builder_st()
		p.public = toks[0].Id == lex.ID_PUB
		if p.public {
			if len(toks) == 1 {
				if p.ended() {
					p.push_err(toks[0], "invalid_syntax")
					continue
				}
				toks = p.next_builder_st()
			} else {
				toks = toks[1:]
			}
		}
		p.build_node(toks)
	}
}

// type_alias builds AST model of type definition statement.
func (p *parser) type_alias(toks []lex.Token) (t ast.TypeAlias) {
	t.Doc = p.comment_group
	p.comment_group = nil
	
	i := 1 // Initialize value is 1 for skip keyword.
	if i >= len(toks) {
		p.push_err(toks[i-1], "invalid_syntax")
		return
	}
	t.Token = toks[1]
	t.Id = t.Token.Kind
	token := toks[i]
	if token.Id != lex.ID_IDENT {
		p.push_err(token, "invalid_syntax")
	}
	i++
	if i >= len(toks) {
		p.push_err(toks[i-1], "invalid_syntax")
		return
	}
	token = toks[i]
	if token.Id != lex.ID_COLON {
		p.push_err(toks[i-1], "invalid_syntax")
		return
	}
	i++
	if i >= len(toks) {
		p.push_err(toks[i-1], "missing_type")
		return
	}
	destType, ok := p.build_type(toks, &i, true)
	t.TargetType = destType
	if ok && i+1 < len(toks) {
		p.push_err(toks[i+1], "invalid_syntax")
	}
	return
}

func (p *parser) build_enum_item_expr(i *int, toks []lex.Token) ast.Expr {
	brace_n := 0
	exprStart := *i
	for ; *i < len(toks); *i++ {
		t := toks[*i]
		if t.Id == lex.ID_BRACE {
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
		if t.Id == lex.ID_COMMA || *i+1 >= len(toks) {
			var exprToks []lex.Token
			if t.Id == lex.ID_COMMA {
				exprToks = toks[exprStart:*i]
			} else {
				exprToks = toks[exprStart:]
			}
			return BuildExpr(exprToks)
		}
	}
	return ast.Expr{}
}

func (p *parser) build_enum_items(toks []lex.Token) []*ast.EnumItem {
	items := make([]*ast.EnumItem, 0)
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		if t.Id == lex.ID_COMMENT {
			continue
		}
		item := new(ast.EnumItem)
		item.Token = t
		if item.Token.Id != lex.ID_IDENT {
			p.push_err(item.Token, "invalid_syntax")
		}
		item.Id = item.Token.Kind
		if i+1 >= len(toks) || toks[i+1].Id == lex.ID_COMMA {
			if i+1 < len(toks) {
				i++
			}
			items = append(items, item)
			continue
		}
		i++
		t = toks[i]
		if t.Id != lex.ID_OP && t.Kind != lex.KND_EQ {
			p.push_err(toks[0], "invalid_syntax")
		}
		i++
		if i >= len(toks) || toks[i].Id == lex.ID_COMMA {
			p.push_err(toks[0], "missing_expr")
			continue
		}
		item.Expr = p.build_enum_item_expr(&i, toks)
		items = append(items, item)
	}
	return items
}

// build_enum builds AST model of enumerator statement.
func (p *parser) build_enum(toks []lex.Token) {
	if len(toks) < 2 || len(toks) < 3 {
		p.push_err(toks[0], "invalid_syntax")
		return
	}
	e := &ast.Enum{}
	e.Token = toks[1]
	e.Doc = p.comment_group
	p.comment_group = nil
	if e.Token.Id != lex.ID_IDENT {
		p.push_err(e.Token, "invalid_syntax")
	}
	e.Id = e.Token.Kind
	i := 2
	if toks[i].Id == lex.ID_COLON {
		i++
		if i >= len(toks) {
			p.push_err(toks[i-1], "invalid_syntax")
			return
		}
		e.DataType, _ = p.build_type(toks, &i, true)
		i++
		if i >= len(toks) {
			p.stop()
			p.push_err(e.Token, "body_not_exist")
			return
		}
	} else {
		e.DataType = ast.Type{Id: types.U32, Kind: types.TYPE_MAP[types.U32]}
	}
	itemToks := p.get_range(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	if itemToks == nil {
		p.stop()
		p.push_err(e.Token, "body_not_exist")
		return
	} else if i < len(toks) {
		p.push_err(toks[i], "invalid_syntax")
	}
	e.Public = p.public
	p.public = false
	e.Items = p.build_enum_items(itemToks)
	p.tree = append(p.tree, ast.Node{Token: e.Token, Data: e})
}

// comment builds AST model of comment.
func (p *parser) comment(t lex.Token) ast.Node {	
	t.Kind = strings.TrimSpace(t.Kind[2:])

	if strings.HasPrefix(t.Kind, lex.PRAGMA_COMMENT_PREFIX) {
		p.push_attribute(t)
	} else {
		if p.comment_group == nil {
			p.comment_group = &ast.CommentGroup{}
		}
		p.comment_group.Comments = append(p.comment_group.Comments, &ast.Comment{
			Token:   t,
			Content: t.Kind,
		})
	}

	return ast.Node{
		Token: t,
		Data: ast.Comment{
			Token:   t,
			Content: t.Kind,
		},
	}
}

func (p *parser) push_attribute(t lex.Token) {
	var attr ast.Attribute
	// Skip attribute prefix
	attr.Tag = t.Kind[len(lex.PRAGMA_COMMENT_PREFIX):]
	attr.Token = t
	ok := false
	for _, kind := range build.ATTRS {
		if attr.Tag == kind {
			ok = true
			break
		}
	}
	if !ok {
		return
	}
	for _, attr2 := range p.attributes {
		if attr.Tag == attr2.Tag {
			return
		}
	}
	p.attributes = append(p.attributes, attr)
}

func (p *parser) struct_fields(toks []lex.Token, cpp_linked bool) []*ast.Var {
	var fields []*ast.Var
	i := 0
	for i < len(toks) {
		var_tokens := p.skip_st(&i, &toks)
		if var_tokens[0].Id == lex.ID_COMMENT {
			continue
		}
		is_pub := var_tokens[0].Id == lex.ID_PUB
		if is_pub {
			if len(var_tokens) == 1 {
				p.push_err(var_tokens[0], "invalid_syntax")
				continue
			}
			var_tokens = var_tokens[1:]
		}
		is_mut := var_tokens[0].Id == lex.ID_MUT
		if is_mut {
			if len(var_tokens) == 1 {
				p.push_err(var_tokens[0], "invalid_syntax")
				continue
			}
			var_tokens = var_tokens[1:]
		}
		v := p.build_var(var_tokens, false, false)
		v.Public = is_pub
		v.Mutable = is_mut
		v.IsField = true
		v.CppLinked = cpp_linked
		fields = append(fields, &v)
	}
	return fields
}

func (p *parser) build_struct(toks []lex.Token, cpp_linked bool) ast.Struct {
	var s ast.Struct
	s.Public = p.public
	p.public = false
	s.Doc = p.comment_group
	p.comment_group = nil
	s.Attributes = p.attributes
	p.attributes = nil
	if len(toks) < 3 {
		p.push_err(toks[0], "invalid_syntax")
		return s
	}

	i := 1
	s.Token = toks[i]
	if s.Token.Id != lex.ID_IDENT {
		p.push_err(s.Token, "invalid_syntax")
	}
	i++
	if i >= len(toks) {
		p.push_err(toks[i], "invalid_syntax")
		return s
	}
	s.Id = s.Token.Kind

	generics_toks := ast.Range(&i, lex.KND_LBRACKET, lex.KND_RBRACKET, toks)
	if generics_toks != nil {
		s.Generics = p.build_generics(generics_toks)
	}
	if i >= len(toks) {
		p.push_err(toks[i], "invalid_syntax")
		return s
	}

	body_toks := p.get_range(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	if body_toks == nil {
		p.stop()
		p.push_err(s.Token, "body_not_exist")
		return s
	}
	if i < len(toks) {
		p.push_err(toks[i], "invalid_syntax")
	}
	s.Fields = p.struct_fields(body_toks, cpp_linked)
	return s
}

func (p *parser) build_structure_node(toks []lex.Token) {
	s := p.build_struct(toks, false)
	p.tree = append(p.tree, ast.Node{Token: s.Token, Data: s})
}

func (p *parser) trait_fns(toks []lex.Token, trait_id string) []*ast.Fn {
	var fns []*ast.Fn
	i := 0
	for i < len(toks) {
		fnToks := p.skip_st(&i, &toks)
		f := p.build_fn(fnToks, true, false, true)
		p.setup_receiver(&f, trait_id)
		f.Public = true
		fns = append(fns, &f)
	}
	return fns
}

// build_trait builds AST model of trait.
func (p *parser) build_trait(toks []lex.Token) {
	var t ast.Trait
	t.Public = p.public
	p.public = false
	t.Doc = p.comment_group
	p.comment_group = nil
	if len(toks) < 3 {
		p.push_err(toks[0], "invalid_syntax")
		return
	}
	t.Token = toks[1]
	if t.Token.Id != lex.ID_IDENT {
		p.push_err(t.Token, "invalid_syntax")
	}
	t.Id = t.Token.Kind
	i := 2
	bodyToks := p.get_range(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	if bodyToks == nil {
		p.stop()
		p.push_err(t.Token, "body_not_exist")
		return
	}
	if i < len(toks) {
		p.push_err(toks[i], "invalid_syntax")
	}
	t.Fns = p.trait_fns(bodyToks, t.Id)
	p.tree = append(p.tree, ast.Node{Token: t.Token, Data: t})
}

func (p *parser) impl_trait_fns(impl *ast.Impl, toks []lex.Token) {
	pos, btoks := p.pos, make([]lex.Token, len(p.tokens))
	copy(btoks, p.tokens)
	p.pos = 0
	p.tokens = toks
	for p.pos != -1 && !p.ended() {
		fnToks := p.next_builder_st()
		tok := fnToks[0]
		switch tok.Id {
		case lex.ID_COMMENT:
			impl.Tree = append(impl.Tree, p.comment(tok))
			continue
		case lex.ID_FN, lex.ID_UNSAFE:
			f := p.get_method(fnToks)
			f.Public = true
			f.Doc = p.comment_group
			p.comment_group = nil
			f.Attributes = p.attributes
			p.attributes = nil
			p.setup_receiver(f, impl.Target.Kind)
			impl.Tree = append(impl.Tree, ast.Node{Token: f.Token, Data: f})
		default:
			p.push_err(tok, "invalid_syntax")
			continue
		}
	}
	p.pos, p.tokens = pos, btoks
}

func (p *parser) impl_struct(impl *ast.Impl, toks []lex.Token) {
	pos, btoks := p.pos, make([]lex.Token, len(p.tokens))
	copy(btoks, p.tokens)
	p.pos = 0
	p.tokens = toks
	for p.pos != -1 && !p.ended() {
		fnToks := p.next_builder_st()
		tok := fnToks[0]
		pub := false
		switch tok.Id {
		case lex.ID_COMMENT:
			impl.Tree = append(impl.Tree, p.comment(tok))
			continue
		}
		if tok.Id == lex.ID_PUB {
			pub = true
			if len(fnToks) == 1 {
				p.push_err(fnToks[0], "invalid_syntax")
				continue
			}
			fnToks = fnToks[1:]
			if len(fnToks) > 0 {
				tok = fnToks[0]
			}
		}
		switch tok.Id {
		case lex.ID_FN, lex.ID_UNSAFE:
			f := p.get_method(fnToks)
			f.Public = pub
			f.Attributes = p.attributes
			p.attributes = nil
			f.Doc = p.comment_group
			p.comment_group = nil
			p.setup_receiver(f, impl.Base.Kind)
			impl.Tree = append(impl.Tree, ast.Node{Token: f.Token, Data: f})
		default:
			p.push_err(tok, "invalid_syntax")
			continue
		}
	}
	p.pos, p.tokens = pos, btoks
}

func (p *parser) get_method(toks []lex.Token) *ast.Fn {
	tok := toks[0]
	if tok.Id == lex.ID_UNSAFE {
		toks = toks[1:]
		if len(toks) == 0 || toks[0].Id != lex.ID_FN {
			p.push_err(tok, "invalid_syntax")
			return nil
		}
	} else if toks[0].Id != lex.ID_FN {
		p.push_err(tok, "invalid_syntax")
		return nil
	}
	f := new(ast.Fn)
	*f = p.build_fn(toks, true, false, false)
	f.IsUnsafe = tok.Id == lex.ID_UNSAFE
	if f.Block != nil {
		f.Block.IsUnsafe = f.IsUnsafe
	}
	return f
}

func (p *parser) impl_fns(impl *ast.Impl, toks []lex.Token) {
	if impl.Target.Id != types.VOID {
		p.impl_trait_fns(impl, toks)
		return
	}
	p.impl_struct(impl, toks)
}

// build_impl builds AST model of impl statement.
func (p *parser) build_impl(toks []lex.Token) {
	tok := toks[0]
	if len(toks) < 2 {
		p.push_err(tok, "invalid_syntax")
		return
	}
	tok = toks[1]
	if tok.Id != lex.ID_IDENT {
		p.push_err(tok, "invalid_syntax")
		return
	}
	var impl ast.Impl
	if len(toks) < 3 {
		p.push_err(tok, "invalid_syntax")
		return
	}
	impl.Base = tok
	tok = toks[2]
	if tok.Id != lex.ID_ITER {
		if tok.Id == lex.ID_BRACE && tok.Kind == lex.KND_LBRACE {
			toks = toks[2:]
			goto body
		}
		p.push_err(tok, "invalid_syntax")
		return
	}
	if len(toks) < 4 {
		p.push_err(tok, "invalid_syntax")
		return
	}
	tok = toks[3]
	if tok.Id != lex.ID_IDENT {
		p.push_err(tok, "invalid_syntax")
		return
	}
	{
		i := 0
		impl.Target, _ = p.build_type(toks[3:4], &i, true)
		toks = toks[4:]
	}
body:
	i := 0
	bodyToks := p.get_range(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	if bodyToks == nil {
		p.stop()
		p.push_err(impl.Base, "body_not_exist")
		return
	}
	if i < len(toks) {
		p.push_err(toks[i], "invalid_syntax")
	}
	p.impl_fns(&impl, bodyToks)
	p.tree = append(p.tree, ast.Node{Token: impl.Base, Data: impl})
}

// link_fn builds AST model of cpp function link.
func (p *parser) link_fn(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := p.public
	p.public = false

	var link ast.CppLinkFn
	link.Token = tok
	link.Link = new(ast.Fn)
	*link.Link = p.build_fn(toks[1:], false, false, true)
	link.Link.Attributes = p.attributes
	p.attributes = nil
	p.tree = append(p.tree, ast.Node{Token: tok, Data: link})

	p.public = bpub
}

// link_var builds AST model of cpp variable link.
func (p *parser) link_var(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := p.public
	p.public = false

	var link ast.CppLinkVar
	link.Token = tok
	link.Link = new(ast.Var)
	*link.Link = p.build_var(toks[1:], true, false)
	p.tree = append(p.tree, ast.Node{Token: tok, Data: link})

	p.public = bpub
}

// link_struct builds AST model of cpp structure link.
func (p *parser) link_struct(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := p.public
	p.public = false

	var link ast.CppLinkStruct
	link.Token = tok
	link.Link = p.build_struct(toks[1:], true)
	p.tree = append(p.tree, ast.Node{Token: tok, Data: link})

	p.public = bpub
}

// link_type_alias builds AST model of cpp type alias link.
func (p *parser) link_type_alias(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := p.public
	p.public = false

	var link ast.CppLinkAlias
	link.Token = tok
	link.Link = p.type_alias(toks[1:])
	p.tree = append(p.tree, ast.Node{Token: tok, Data: link})

	p.public = bpub
}

// CppLinks builds AST model of cpp link statement.
func (p *parser) cpp_link(toks []lex.Token) {
	tok := toks[0]
	if len(toks) == 1 {
		p.push_err(tok, "invalid_syntax")
		return
	}
	tok = toks[1]
	switch tok.Id {
	case lex.ID_FN, lex.ID_UNSAFE:
		p.link_fn(toks)
	case lex.ID_LET:
		p.link_var(toks)
	case lex.ID_STRUCT:
		p.link_struct(toks)
	case lex.ID_TYPE:
		p.link_type_alias(toks)
	default:
		p.push_err(tok, "invalid_syntax")
	}
}

func tokstoa(toks []lex.Token) string {
	var str strings.Builder
	for _, tok := range toks {
		str.WriteString(tok.Kind)
	}
	return str.String()
}

// build_use builds AST model of use declaration.
func (p *parser) build_use(toks []lex.Token) {
	var use ast.UseDecl
	use.Token = toks[0]
	if len(toks) < 2 {
		p.push_err(use.Token, "missing_use_path")
		return
	}
	toks = toks[1:]
	p.build_use_decl(&use, toks)
	p.tree = append(p.tree, ast.Node{Token: use.Token, Data: use})
}

func (p *parser) get_selectors(toks []lex.Token) []lex.Token {
	i := 0
	toks = p.get_range(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	parts, errs := ast.Parts(toks, lex.ID_COMMA, true)
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

func (p *parser) build_use_cpp_decl(use *ast.UseDecl, toks []lex.Token) {
	if len(toks) > 2 {
		p.push_err(toks[2], "invalid_syntax")
	}
	tok := toks[1]
	if tok.Id != lex.ID_LITERAL || (tok.Kind[0] != '`' && tok.Kind[0] != '"') {
		p.push_err(tok, "invalid_expr")
		return
	}
	use.Cpp = true
	use.Path = tok.Kind[1 : len(tok.Kind)-1]
}

func (p *parser) build_use_decl(use *ast.UseDecl, toks []lex.Token) {
	var path strings.Builder
	path.WriteString(jule.STDLIB_PATH)
	path.WriteRune(os.PathSeparator)
	tok := toks[0]
	isStd := false
	if tok.Id == lex.ID_CPP {
		p.build_use_cpp_decl(use, toks)
		return
	}
	if tok.Id != lex.ID_IDENT || tok.Kind != "std" {
		p.push_err(toks[0], "invalid_syntax")
	}
	isStd = true
	if len(toks) < 3 {
		p.push_err(tok, "invalid_syntax")
		return
	}
	toks = toks[2:]
	tok = toks[len(toks)-1]
	switch tok.Id {
	case lex.ID_DBLCOLON:
		p.push_err(tok, "invalid_syntax")
		return
	case lex.ID_BRACE:
		if tok.Kind != lex.KND_RBRACE {
			p.push_err(tok, "invalid_syntax")
			return
		}
		var selectors []lex.Token
		toks, selectors = ast.RangeLast(toks)
		use.Selectors = p.get_selectors(selectors)
		if len(toks) == 0 {
			p.push_err(tok, "invalid_syntax")
			return
		}
		tok = toks[len(toks)-1]
		if tok.Id != lex.ID_DBLCOLON {
			p.push_err(tok, "invalid_syntax")
			return
		}
		toks = toks[:len(toks)-1]
		if len(toks) == 0 {
			p.push_err(tok, "invalid_syntax")
			return
		}
	case lex.ID_OP:
		if tok.Kind != lex.KND_STAR {
			p.push_err(tok, "invalid_syntax")
			return
		}
		toks = toks[:len(toks)-1]
		if len(toks) == 0 {
			p.push_err(tok, "invalid_syntax")
			return
		}
		tok = toks[len(toks)-1]
		if tok.Id != lex.ID_DBLCOLON {
			p.push_err(tok, "invalid_syntax")
			return
		}
		toks = toks[:len(toks)-1]
		if len(toks) == 0 {
			p.push_err(tok, "invalid_syntax")
			return
		}
		use.FullUse = true
	}
	for i, tok := range toks {
		if i%2 != 0 {
			if tok.Id != lex.ID_DBLCOLON {
				p.push_err(tok, "invalid_syntax")
			}
			path.WriteRune(os.PathSeparator)
			continue
		}
		if tok.Id != lex.ID_IDENT {
			p.push_err(tok, "invalid_syntax")
		}
		path.WriteString(tok.Kind)
	}
	use.LinkString = tokstoa(toks)
	if isStd {
		use.LinkString = "std::" + use.LinkString
	}
	use.Path = path.String()
}

func (p *parser) setup_receiver(f *ast.Fn, owner_id string) {
	if len(f.Params) == 0 {
		p.push_err(f.Token, "missing_receiver")
		return
	}
	param := f.Params[0]
	if param.Id != lex.KND_SELF {
		p.push_err(f.Token, "missing_receiver")
		return
	}
	f.Receiver = new(ast.Var)
	f.Receiver.DataType = ast.Type{
		Id:   types.STRUCT,
		Kind: owner_id,
	}
	f.Receiver.Mutable = param.Mutable
	if param.DataType.Kind != "" && param.DataType.Kind[0] == '&' {
		f.Receiver.DataType.Kind = lex.KND_AMPER + f.Receiver.DataType.Kind
	}
	f.Params = f.Params[1:]
}

func (p *parser) fn_prototype(toks []lex.Token, i *int, method bool, anon bool) (f ast.Fn, ok bool) {
	ok = true
	f.Token = toks[*i]
	if f.Token.Id == lex.ID_UNSAFE {
		f.IsUnsafe = true
		*i++
		if *i >= len(toks) {
			p.push_err(f.Token, "invalid_syntax")
			ok = false
			return
		}
		f.Token = toks[*i]
	}
	// Skips fn tok
	*i++
	if *i >= len(toks) {
		p.push_err(f.Token, "invalid_syntax")
		ok = false
		return
	}
	f.Public = p.public
	p.public = false
	if anon {
		f.Id = lex.ANONYMOUS_ID
	} else {
		tok := toks[*i]
		if tok.Id != lex.ID_IDENT {
			p.push_err(tok, "invalid_syntax")
			ok = false
		}
		f.Id = tok.Kind
		*i++
	}

	f.RetType.DataType.Id = types.VOID
	f.RetType.DataType.Kind = types.TYPE_MAP[f.RetType.DataType.Id]
	if *i >= len(toks) {
		p.push_err(f.Token, "invalid_syntax")
		return
	}

	generics_toks := ast.Range(i, lex.KND_LBRACKET, lex.KND_RBRACKET, toks)
	if generics_toks != nil {
		f.Generics = p.build_generics(generics_toks)
		if len(f.Generics) > 0 {
			f.Combines = new([][]ast.Type)
		}
	}

	if toks[*i].Kind != lex.KND_LPAREN {
		p.push_err(toks[*i], "missing_function_parentheses")
		return
	}
	params_toks := p.get_range(i, lex.KND_LPAREN, lex.KND_RPARENT, &toks)
	if len(params_toks) > 0 {
		f.Params = p.build_params(params_toks, method, false)
	}

	t, ret_ok := p.fn_ret_type(toks, i)
	if ret_ok {
		f.RetType = t
		*i++
	}
	return
}

// stop stops ast building at next iteration.
func (p *parser) stop() { p.pos = -1 }

// build_fn builds AST model of function.
func (p *parser) build_fn(toks []lex.Token, method, anon, prototype bool) (f ast.Fn) {
	var ok bool
	i := 0
	f, ok = p.fn_prototype(toks, &i, method, anon)
	if prototype {
		if i+1 < len(toks) {
			p.push_err(toks[i+1], "invalid_syntax")
		}
		return
	} else if !ok {
		return
	}
	if i >= len(toks) {
		p.stop()
		p.push_err(f.Token, "body_not_exist")
		return
	}
	block_toks := p.get_range(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	if block_toks != nil {
		f.Block = p.build_block(block_toks)
		f.Block.IsUnsafe = f.IsUnsafe
		if i < len(toks) {
			p.push_err(toks[i], "invalid_syntax")
		}
	} else {
		p.stop()
		p.push_err(f.Token, "body_not_exist")
		p.tokens = append(toks, p.tokens...)
	}
	return
}

func (p *parser) build_generic(toks []lex.Token) *ast.GenericType {
	if len(toks) > 1 {
		p.push_err(toks[1], "invalid_syntax")
	}
	gt := new(ast.GenericType)
	gt.Token = toks[0]
	if gt.Token.Id != lex.ID_IDENT {
		p.push_err(gt.Token, "invalid_syntax")
	}
	gt.Id = gt.Token.Kind
	return gt
}

func (p *parser) build_generics(toks []lex.Token) []*ast.GenericType {
	tok := toks[0]
	if len(toks) == 0 {
		p.push_err(tok, "missing_expr")
		return nil
	}
	parts, errs := ast.Parts(toks, lex.ID_COMMA, true)
	p.errors = append(p.errors, errs...)
	generics := make([]*ast.GenericType, len(parts))
	for i, part := range parts {
		if len(parts) == 0 {
			continue
		}
		generics[i] = p.build_generic(part)
	}
	return generics
}

// build_global_type_alias builds global type alias.
func (p *parser) build_global_type_alias(toks []lex.Token) ast.Node {
	t := p.type_alias(toks)
	t.Public = p.public
	p.public = false
	return ast.Node{Token: t.Token, Data: t}
}

// build_global_var builds AST model of global variable.
func (p *parser) build_global_var(toks []lex.Token) {
	if toks == nil {
		return
	}
	bs := block_st{toks: toks}
	s := p.build_var_st(&bs, true)
	v := s.Data.(ast.Var)
	v.Doc = p.comment_group
	p.comment_group = nil
	s.Data = v
	p.tree = append(p.tree, ast.Node{
		Token: s.Token,
		Data:  s,
	})
}

func (p *parser) build_self(toks []lex.Token) (model ast.Param) {
	if len(toks) == 0 {
		return
	}
	i := 0
	if toks[i].Id == lex.ID_MUT {
		model.Mutable = true
		i++
		if i >= len(toks) {
			p.push_err(toks[i-1], "invalid_syntax")
			return
		}
	}
	if toks[i].Kind == lex.KND_AMPER {
		model.DataType.Kind = "&"
		i++
		if i >= len(toks) {
			p.push_err(toks[i-1], "invalid_syntax")
			return
		}
	}
	if toks[i].Id == lex.ID_SELF {
		model.Id = lex.KND_SELF
		model.Token = toks[i]
		i++
		if i < len(toks) {
			p.push_err(toks[i+1], "invalid_syntax")
		}
	}
	return
}

// build_params builds AST model of function parameters.
func (p *parser) build_params(toks []lex.Token, method, mustPure bool) []ast.Param {
	parts, errs := ast.Parts(toks, lex.ID_COMMA, true)
	p.errors = append(p.errors, errs...)
	if len(parts) == 0 {
		return nil
	}
	var params []ast.Param
	if method && len(parts) > 0 {
		param := p.build_self(parts[0])
		if param.Id == lex.KND_SELF {
			params = append(params, param)
			parts = parts[1:]
		}
	}
	for _, part := range parts {
		p.push_param(&params, part, mustPure)
	}
	p.check_params(&params)
	return params
}

func (p *parser) check_params(params *[]ast.Param) {
	for i := range *params {
		param := &(*params)[i]
		if param.Id == lex.KND_SELF || param.DataType.Token.Id != lex.ID_NA {
			continue
		}
		if param.Token.Id == lex.ID_NA {
			p.push_err(param.Token, "missing_type")
		} else {
			param.DataType.Token = param.Token
			param.DataType.Id = types.ID
			param.DataType.Kind = param.DataType.Token.Kind
			param.DataType.Original = param.DataType
			param.Id = lex.ANONYMOUS_ID
			param.Token = lex.Token{}
		}
	}
}

func (p *parser) param_type_begin(param *ast.Param, i *int, toks []lex.Token) {
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		switch tok.Id {
		case lex.ID_OP:
			switch tok.Kind {
			case lex.KND_TRIPLE_DOT:
				if param.Variadic {
					p.push_err(tok, "already_variadic")
					continue
				}
				param.Variadic = true
			default:
				return
			}
		default:
			return
		}
	}
}

func (p *parser) param_body_id(param *ast.Param, tok lex.Token) {
	if lex.IsIgnoreId(tok.Kind) {
		param.Id = lex.ANONYMOUS_ID
		return
	}
	param.Id = tok.Kind
}

func (p *parser) param_body(param *ast.Param, i *int, toks []lex.Token, mustPure bool) {
	p.param_body_id(param, toks[*i])
	// +1 for skip identifier token
	tok := toks[*i]
	toks = toks[*i+1:]
	if len(toks) == 0 {
		return
	} else if len(toks) < 2 {
		p.push_err(tok, "missing_type")
		return
	}
	tok = toks[*i]
	if tok.Id != lex.ID_COLON {
		p.push_err(tok, "invalid_syntax")
		return
	}
	toks = toks[*i+1:] // Skip colon
	p.param_type(param, toks, mustPure)
}

func (p *parser) param_type(param *ast.Param, toks []lex.Token, mustPure bool) {
	i := 0
	if !mustPure {
		p.param_type_begin(param, &i, toks)
		if i >= len(toks) {
			return
		}
	}
	param.DataType, _ = p.build_type(toks, &i, true)
	i++
	if i < len(toks) {
		p.push_err(toks[i], "invalid_syntax")
	}
}

func (p *parser) push_param(params *[]ast.Param, toks []lex.Token, mustPure bool) {
	var param ast.Param
	param.Token = toks[0]
	if param.Token.Id == lex.ID_MUT {
		param.Mutable = true
		if len(toks) == 1 {
			p.push_err(toks[0], "invalid_syntax")
			return
		}
		toks = toks[1:]
		param.Token = toks[0]
	}
	// Just data type
	if param.Token.Id != lex.ID_IDENT {
		param.Id = lex.ANONYMOUS_ID
		p.param_type(&param, toks, mustPure)
	} else {
		i := 0
		p.param_body(&param, &i, toks, mustPure)
	}
	*params = append(*params, param)
}

func (p *parser) __build_type(t *ast.Type, toks []lex.Token, i *int, err bool) (ok bool) {
	tb := type_builder{
		p:      p,
		t:      t,
		tokens: toks,
		i:      i,
		err:    err,
	}
	return tb.build()
}

// build_type builds AST model of data-type.
func (p *parser) build_type(toks []lex.Token, i *int, err bool) (t ast.Type, ok bool) {
	tok := toks[*i]
	ok = p.__build_type(&t, toks, i, err)
	if err && t.Token.Id == lex.ID_NA {
		p.push_err(tok, "invalid_type")
	}
	return
}

func (p *parser) fn_multi_type_tet(toks []lex.Token, i *int) (t ast.RetType, ok bool) {
	tok := toks[*i]
	t.DataType.Kind += tok.Kind
	*i++
	if *i >= len(toks) {
		*i--
		t.DataType, ok = p.build_type(toks, i, false)
		return
	}
	tok = toks[*i]
	*i-- // For point to parenthses - ( -
	rang := ast.Range(i, lex.KND_LPAREN, lex.KND_RPARENT, toks)
	params := p.build_params(rang, false, true)
	types := make([]ast.Type, len(params))
	for i, param := range params {
		types[i] = param.DataType
		if param.Id != lex.ANONYMOUS_ID {
			param.Token.Kind = param.Id
		} else {
			param.Token.Kind = lex.IGNORE_ID
		}
		t.Identifiers = append(t.Identifiers, param.Token)
	}
	if len(types) > 1 {
		t.DataType.MultiTyped = true
		t.DataType.Tag = types
	} else {
		t.DataType = types[0]
	}
	// Decrament for correct block parsing
	*i--
	ok = true
	return
}

// fn_ret_type builds ret data-type of function.
func (p *parser) fn_ret_type(toks []lex.Token, i *int) (t ast.RetType, ok bool) {
	t.DataType.Id = types.VOID
	t.DataType.Kind = types.TYPE_MAP[t.DataType.Id]
	if *i >= len(toks) {
		return
	}
	tok := toks[*i]
	switch tok.Id {
	case lex.ID_BRACE:
		if tok.Kind == lex.KND_LBRACE {
			return
		}
	case lex.ID_OP:
		if tok.Kind == lex.KND_EQ {
			return
		}
	case lex.ID_COLON:
		if *i+1 >= len(toks) {
			p.push_err(tok, "missing_type")
			return
		}
		*i++
		tok = toks[*i]
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_LPAREN:
				return p.fn_multi_type_tet(toks, i)
			case lex.KND_LBRACE:
				return
			}
		}
		t.DataType, ok = p.build_type(toks, i, true)
		return
	}
	*i++
	p.push_err(tok, "invalid_syntax")
	return
}

func (p *parser) push_st_to_block(bs *block_st) {
	if len(bs.toks) == 0 {
		return
	}
	lastTok := bs.toks[len(bs.toks)-1]
	if lastTok.Id == lex.ID_SEMICOLON {
		if len(bs.toks) == 1 {
			return
		}
		bs.toks = bs.toks[:len(bs.toks)-1]
	}
	s := p.build_st(bs)
	if s.Data == nil {
		return
	}
	s.WithTerminator = bs.terminated
	bs.block.Tree = append(bs.block.Tree, s)
}

func get_next_st(toks []lex.Token) []lex.Token {
	pos, terminated := ast.NextStPos(toks, 0)
	if terminated {
		return toks[:pos-1]
	}
	return toks[:pos]
}

func set_to_next_st(bs *block_st) {
	if bs.nextToks != nil {
		bs.toks = bs.nextToks
		bs.nextToks = nil
		return
	}
	*bs.srcToks = (*bs.srcToks)[bs.pos:]
	bs.pos, bs.terminated = ast.NextStPos(*bs.srcToks, 0)
	if bs.terminated {
		bs.toks = (*bs.srcToks)[:bs.pos-1]
	} else {
		bs.toks = (*bs.srcToks)[:bs.pos]
	}
}

func block_st_finished(bs *block_st) bool {
	return bs.nextToks == nil && bs.pos >= len(*bs.srcToks)
}

// build_block builds AST model of statements of code build_block.
func (p *parser) build_block(toks []lex.Token) (block *ast.Block) {
	block = new(ast.Block)
	var bs block_st
	bs.block = block
	bs.srcToks = &toks
	for {
		set_to_next_st(&bs)
		p.push_st_to_block(&bs)
		if block_st_finished(&bs) {
			break
		}
	}
	return
}

// build_st builds AST model of statement.
func (p *parser) build_st(bs *block_st) (s ast.St) {
	tok := bs.toks[0]
	if tok.Id == lex.ID_IDENT {
		s, ok := p.build_id_st(bs)
		if ok {
			return s
		}
	}
	s, ok := p.build_assign_st(bs.toks)
	if ok {
		return s
	}
	switch tok.Id {
	case lex.ID_CONST, lex.ID_LET, lex.ID_MUT:
		return p.build_var_st(bs, true)
	case lex.ID_RET:
		return p.build_ret_st(bs.toks)
	case lex.ID_ITER:
		return p.buid_iter(bs)
	case lex.ID_BREAK:
		return p.build_break_st(bs.toks)
	case lex.ID_CONTINUE:
		return p.build_continue_st(bs.toks)
	case lex.ID_IF:
		return p.build_if_else_chain(bs)
	case lex.ID_COMMENT:
		return p.comment_st(bs.toks[0])
	case lex.ID_CO:
		return p.build_concurrent_call_st(bs.toks)
	case lex.ID_GOTO:
		return p.build_goto_st(bs.toks)
	case lex.ID_FALL:
		return p.build_fall_st(bs.toks)
	case lex.ID_TYPE:
		t := p.type_alias(bs.toks)
		s.Token = t.Token
		s.Data = t
		return
	case lex.ID_MATCH:
		return p.build_match(bs.toks)
	case lex.ID_UNSAFE, lex.ID_DEFER:
		return p.block_st(bs.toks)
	case lex.ID_BRACE:
		if tok.Kind == lex.KND_LBRACE {
			return p.block_st(bs.toks)
		}
	}
	if ast.IsFnCall(bs.toks) != nil {
		return p.build_expr_st(bs)
	}
	p.push_err(tok, "invalid_syntax")
	return
}

func (p *parser) block_st(toks []lex.Token) ast.St {
	is_unsafe := false
	is_deferred := false
	tok := toks[0]
	if tok.Id == lex.ID_UNSAFE {
		is_unsafe = true
		toks = toks[1:]
		if len(toks) == 0 {
			p.push_err(tok, "invalid_syntax")
			return ast.St{}
		}
		tok = toks[0]
		if tok.Id == lex.ID_DEFER {
			is_deferred = true
			toks = toks[1:]
			if len(toks) == 0 {
				p.push_err(tok, "invalid_syntax")
				return ast.St{}
			}
		}
	} else if tok.Id == lex.ID_DEFER {
		is_deferred = true
		toks = toks[1:]
		if len(toks) == 0 {
			p.push_err(tok, "invalid_syntax")
			return ast.St{}
		}
	}

	i := 0
	toks = ast.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, toks)
	if len(toks) == 0 {
		p.push_err(tok, "invalid_syntax")
		return ast.St{}
	} else if i < len(toks) {
		p.push_err(toks[i], "invalid_syntax")
	}
	block := p.build_block(toks)
	block.IsUnsafe = is_unsafe
	block.Deferred = is_deferred
	return ast.St{Token: tok, Data: block}
}

func (p *parser) build_assign_info(toks []lex.Token) (info ast.AssignInfo) {
	info.Ok = true
	brace_n := 0
	for i, tok := range toks {
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				brace_n++
			default:
				brace_n--
			}
		}
		if brace_n > 0 {
			continue
		} else if tok.Id != lex.ID_OP {
			continue
		} else if !ast.IsAssignOp(tok.Kind) {
			continue
		}
		info.Left = toks[:i]
		if info.Left == nil {
			p.push_err(tok, "invalid_syntax")
			info.Ok = false
		}
		info.Setter = tok
		if i+1 >= len(toks) {
			info.Right = nil
			info.Ok = ast.IsPostfixOp(info.Setter.Kind)
			break
		}
		info.Right = toks[i+1:]
		if ast.IsPostfixOp(info.Setter.Kind) {
			if info.Right != nil {
				p.push_err(info.Right[0], "invalid_syntax")
				info.Right = nil
			}
		}
		break
	}
	return
}

func (p *parser) build_assign_left(toks []lex.Token) (l ast.AssignLeft) {
	l.Expr.Tokens = toks
	if l.Expr.Tokens[0].Id == lex.ID_IDENT {
		l.Var.Token = l.Expr.Tokens[0]
		l.Var.Id = l.Var.Token.Kind
	}
	l.Expr = BuildExpr(l.Expr.Tokens)
	return
}

func (p *parser) build_assign_lefts(parts [][]lex.Token) []ast.AssignLeft {
	var lefts []ast.AssignLeft
	for _, part := range parts {
		l := p.build_assign_left(part)
		lefts = append(lefts, l)
	}
	return lefts
}

func (p *parser) build_assign_exprs(toks []lex.Token) []ast.Expr {
	parts, errs := ast.Parts(toks, lex.ID_COMMA, true)
	if len(errs) > 0 {
		p.errors = append(p.errors, errs...)
		return nil
	}
	exprs := make([]ast.Expr, len(parts))
	for i, p := range parts {
		exprs[i] = BuildExpr(p)
	}
	return exprs
}

// build_assign_st builds AST model of assignment statement.
func (p *parser) build_assign_st(toks []lex.Token) (s ast.St, _ bool) {
	assign, ok := p.build_assign_expr(toks)
	if !ok {
		return
	}
	s.Token = toks[0]
	s.Data = assign
	return s, true
}

// build_assign_expr builds AST model of assignment expression.
func (p *parser) build_assign_expr(toks []lex.Token) (assign ast.Assign, ok bool) {
	if !ast.CheckAssignTokens(toks) {
		return
	}
	switch toks[0].Id {
	case lex.ID_LET:
		return p.let_decl_assign(toks)
	default:
		return p.plain_assign(toks)
	}
}

func (p *parser) let_decl_assign(toks []lex.Token) (assign ast.Assign, ok bool) {
	if len(toks) < 1 {
		return
	}
	// Skip "let" keyword
	toks = toks[1:]
	tok := toks[0]
	if tok.Id != lex.ID_BRACE || tok.Kind != lex.KND_LPAREN {
		return
	}
	ok = true
	var i int
	rang := ast.Range(&i, lex.KND_LPAREN, lex.KND_RPARENT, toks)
	if rang == nil {
		p.push_err(tok, "invalid_syntax")
		return
	} else if i+1 < len(toks) {
		assign.Setter = toks[i]
		i++
		assign.Right = p.build_assign_exprs(toks[i:])
	}
	parts, errs := ast.Parts(rang, lex.ID_COMMA, true)
	if len(errs) > 0 {
		p.errors = append(p.errors, errs...)
		return
	}
	for _, part := range parts {
		mutable := false
		tok := part[0]
		if tok.Id == lex.ID_MUT {
			mutable = true
			part = part[1:]
			if len(part) != 1 {
				p.push_err(tok, "invalid_syntax")
				continue
			}
		}
		if part[0].Id != lex.ID_IDENT && part[0].Id != lex.ID_BRACE && part[0].Kind != lex.KND_LPAREN {
			p.push_err(tok, "invalid_syntax")
			continue
		}
		l := p.build_assign_left(part)
		l.Var.Mutable = mutable
		l.Var.New = l.Var.Id != "" && !lex.IsIgnoreId(l.Var.Id)
		l.Var.SetterTok = assign.Setter
		assign.Left = append(assign.Left, l)
	}
	return
}

func (p *parser) plain_assign(toks []lex.Token) (assign ast.Assign, ok bool) {
	info := p.build_assign_info(toks)
	if !info.Ok {
		return
	}
	ok = true
	assign.Setter = info.Setter
	parts, errs := ast.Parts(info.Left, lex.ID_COMMA, true)
	if len(errs) > 0 {
		p.errors = append(p.errors, errs...)
		return
	}
	assign.Left = p.build_assign_lefts(parts)
	if info.Right != nil {
		assign.Right = p.build_assign_exprs(info.Right)
	}
	return
}

// BuildReturnStatement builds AST model of return statement.
func (p *parser) build_id_st(bs *block_st) (s ast.St, ok bool) {
	if len(bs.toks) == 1 {
		return
	}
	tok := bs.toks[1]
	switch tok.Id {
	case lex.ID_COLON:
		return p.build_label_st(bs), true
	}
	return
}

// build_label_st builds AST model of label.
func (p *parser) build_label_st(bs *block_st) ast.St {
	var l ast.Label
	l.Token = bs.toks[0]
	l.Label = l.Token.Kind
	if len(bs.toks) > 2 {
		bs.nextToks = bs.toks[2:]
	}
	return ast.St{
		Token: l.Token,
		Data:  l,
	}
}

// build_expr_st builds AST model of expression.
func (p *parser) build_expr_st(bs *block_st) ast.St {
	st := ast.ExprSt{Expr: BuildExpr(bs.toks)}
	return ast.St{Token: bs.toks[0], Data: st}
}

func (p *parser) build_var_begin(v *ast.Var, i *int, toks []lex.Token) {
	tok := toks[*i]
	switch tok.Id {
	case lex.ID_LET:
		// Initialize 1 for skip the let keyword
		*i++
		if toks[*i].Id == lex.ID_MUT {
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
	if *i >= len(toks) {
		p.push_err(tok, "invalid_syntax")
	}
}

func (p *parser) build_var_type_and_expr(v *ast.Var, toks []lex.Token, i int, expr bool) {
	tok := toks[i]
	if tok.Id == lex.ID_COLON {
		i++ // Skip type annotation operator (:)
		if i >= len(toks) ||
			(toks[i].Id == lex.ID_OP && toks[i].Kind == lex.KND_EQ) {
			p.push_err(tok, "missing_type")
			return
		}
		t, ok := p.build_type(toks, &i, false)
		if ok {
			v.DataType = t
			i++
			if i >= len(toks) {
				return
			}
			tok = toks[i]
		}
	}
	if expr && tok.Id == lex.ID_OP {
		if tok.Kind != lex.KND_EQ {
			p.push_err(tok, "invalid_syntax")
			return
		}
		valueToks := toks[i+1:]
		if len(valueToks) == 0 {
			p.push_err(tok, "missing_expr")
			return
		}
		v.Expr = BuildExpr(valueToks)
		v.SetterTok = tok
	} else {
		p.push_err(tok, "invalid_syntax")
	}
}

// build_var builds AST model of variable statement.
func (p *parser) build_var(toks []lex.Token, begin, expr bool) (v ast.Var) {
	v.Public = p.public
	p.public = false
	i := 0
	v.Token = toks[i]
	if begin {
		p.build_var_begin(&v, &i, toks)
		if i >= len(toks) {
			return
		}
	}
	v.Token = toks[i]
	if v.Token.Id != lex.ID_IDENT {
		p.push_err(v.Token, "invalid_syntax")
		return
	}
	v.Id = v.Token.Kind
	v.DataType.Id = types.VOID
	v.DataType.Kind = types.TYPE_MAP[v.DataType.Id]
	if i >= len(toks) {
		return
	}
	i++
	if i < len(toks) {
		p.build_var_type_and_expr(&v, toks, i, expr)
	} else if !expr {
		p.push_err(v.Token, "missing_type")
	}
	return
}

// build_var_st builds AST model of variable declaration statement.
func (p *parser) build_var_st(bs *block_st, expr bool) ast.St {
	v := p.build_var(bs.toks, true, expr)
	v.Owner = bs.block
	return ast.St{Token: v.Token, Data: v}
}

// comment_st builds AST model of comment statement.
func (p *parser) comment_st(tok lex.Token) (s ast.St) {
	s.Token = tok
	tok.Kind = strings.TrimSpace(tok.Kind[2:])
	s.Data = ast.Comment{Content: tok.Kind}
	return
}

func (p *parser) build_concurrent_call_st(toks []lex.Token) (s ast.St) {
	var cc ast.ConcurrentCall
	cc.Token = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		p.push_err(cc.Token, "missing_expr")
		return
	}
	if ast.IsFnCall(toks) == nil {
		p.push_err(cc.Token, "expr_not_func_call")
	}
	cc.Expr = BuildExpr(toks)
	s.Token = cc.Token
	s.Data = cc
	return
}

func (p *parser) build_fall_st(toks []lex.Token) (s ast.St) {
	s.Token = toks[0]
	if len(toks) > 1 {
		p.push_err(toks[1], "invalid_syntax")
	}
	s.Data = ast.Fall{
		Token: s.Token,
	}
	return
}

func (p *parser) build_goto_st(toks []lex.Token) (s ast.St) {
	s.Token = toks[0]
	if len(toks) == 1 {
		p.push_err(s.Token, "missing_goto_label")
		return
	} else if len(toks) > 2 {
		p.push_err(toks[2], "invalid_syntax")
	}
	idTok := toks[1]
	if idTok.Id != lex.ID_IDENT {
		p.push_err(idTok, "invalid_syntax")
		return
	}
	var gt ast.Goto
	gt.Token = s.Token
	gt.Label = idTok.Kind
	s.Data = gt
	return
}

// build_ret_st builds AST model of return statement.
func (p *parser) build_ret_st(toks []lex.Token) ast.St {
	var ret ast.Ret
	ret.Token = toks[0]
	if len(toks) > 1 {
		ret.Expr = BuildExpr(toks[1:])
	}
	return ast.St{
		Token: ret.Token,
		Data:  ret,
	}
}

func (p *parser) get_while_iter_profile(toks []lex.Token) ast.IterWhile {
	return ast.IterWhile{
		Expr: BuildExpr(toks),
	}
}

func (p *parser) get_foreach_vars_tokens(toks []lex.Token) [][]lex.Token {
	vars, errs := ast.Parts(toks, lex.ID_COMMA, true)
	p.errors = append(p.errors, errs...)
	return vars
}

func (p *parser) get_var_profile(toks []lex.Token) (v ast.Var) {
	if len(toks) == 0 {
		return
	}
	v.Token = toks[0]
	if v.Token.Id == lex.ID_MUT {
		v.Mutable = true
		if len(toks) == 1 {
			p.push_err(v.Token, "invalid_syntax")
		}
		v.Token = toks[1]
	} else if len(toks) > 1 {
		p.push_err(toks[1], "invalid_syntax")
	}
	if v.Token.Id != lex.ID_IDENT {
		p.push_err(v.Token, "invalid_syntax")
		return
	}
	v.Id = v.Token.Kind
	v.New = true
	return
}

func (p *parser) get_foreach_iter_vars(varsToks [][]lex.Token) []ast.Var {
	var vars []ast.Var
	for _, toks := range varsToks {
		vars = append(vars, p.get_var_profile(toks))
	}
	return vars
}

func (p *parser) setup_foreach_explicit_vars(f *ast.IterForeach, toks []lex.Token) {
	i := 0
	rang := ast.Range(&i, lex.KND_LPAREN, lex.KND_RPARENT, toks)
	if i < len(toks) {
		p.push_err(f.InToken, "invalid_syntax")
	}
	p.setup_foreach_plain_vars(f, rang)
}

func (p *parser) setup_foreach_plain_vars(f *ast.IterForeach, toks []lex.Token) {
	varsToks := p.get_foreach_vars_tokens(toks)
	if len(varsToks) == 0 {
		return
	}
	if len(varsToks) > 2 {
		p.push_err(f.InToken, "much_foreach_vars")
	}
	vars := p.get_foreach_iter_vars(varsToks)
	f.KeyA = vars[0]
	if len(vars) > 1 {
		f.KeyB = vars[1]
	} else {
		f.KeyB.Id = lex.IGNORE_ID
	}
}

func (p *parser) setup_foreach_vars(f *ast.IterForeach, toks []lex.Token) {
	if toks[0].Id == lex.ID_BRACE {
		if toks[0].Kind != lex.KND_LPAREN {
			p.push_err(toks[0], "invalid_syntax")
			return
		}
		p.setup_foreach_explicit_vars(f, toks)
		return
	}
	p.setup_foreach_plain_vars(f, toks)
}

func (p *parser) get_foreach_iter_profile(varToks, exprToks []lex.Token, inTok lex.Token) ast.IterForeach {
	var foreach ast.IterForeach
	foreach.InToken = inTok
	if len(exprToks) == 0 {
		p.push_err(inTok, "missing_expr")
		return foreach
	}
	foreach.Expr = BuildExpr(exprToks)
	if len(varToks) == 0 {
		foreach.KeyA.Id = lex.IGNORE_ID
		foreach.KeyB.Id = lex.IGNORE_ID
	} else {
		p.setup_foreach_vars(&foreach, varToks)
	}
	return foreach
}

func (p *parser) get_iter_profile(toks []lex.Token, errtok lex.Token) any {
	brace_n := 0
	for i, tok := range toks {
		if tok.Id == lex.ID_BRACE {
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
			varToks := toks[:i]
			exprToks := toks[i+1:]
			return p.get_foreach_iter_profile(varToks, exprToks, tok)
		}
	}
	return p.get_while_iter_profile(toks)
}

func (p *parser) build_next_st(toks []lex.Token) ast.St {
	s := p.build_st(&block_st{toks: toks})
	switch s.Data.(type) {
	case ast.ExprSt, ast.Assign, ast.Var:
	default:
		p.push_err(toks[0], "invalid_syntax")
	}
	return s
}

func (p *parser) get_while_next_iter_profile(bs *block_st) (s ast.St) {
	var iter ast.Iter
	iter.Token = bs.toks[0]
	bs.toks = bs.toks[1:]
	profile := ast.IterWhile{}
	if len(bs.toks) > 0 {
		profile.Expr = BuildExpr(bs.toks)
	}
	if block_st_finished(bs) {
		p.push_err(iter.Token, "invalid_syntax")
		return
	}
	set_to_next_st(bs)
	st_toks := ast.GetBlockExpr(bs.toks)
	if len(st_toks) > 0 {
		profile.Next = p.build_next_st(st_toks)
	}
	i := len(st_toks)
	blockToks := p.get_range(&i, lex.KND_LBRACE, lex.KND_RBRACE, &bs.toks)
	if blockToks == nil {
		p.stop()
		p.push_err(iter.Token, "body_not_exist")
		return
	}
	if i < len(bs.toks) {
		p.push_err(bs.toks[i], "invalid_syntax")
	}
	iter.Block = p.build_block(blockToks)
	iter.Profile = profile
	return ast.St{Token: iter.Token, Data: iter}
}

func (p *parser) build_common_iter_profile(toks []lex.Token) (s ast.St) {
	var iter ast.Iter
	iter.Token = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		p.stop()
		p.push_err(iter.Token, "body_not_exist")
		return
	}
	exprToks := ast.GetBlockExpr(toks)
	if len(exprToks) > 0 {
		iter.Profile = p.get_iter_profile(exprToks, iter.Token)
	}
	i := len(exprToks)
	blockToks := p.get_range(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	if blockToks == nil {
		p.stop()
		p.push_err(iter.Token, "body_not_exist")
		return
	}
	if i < len(toks) {
		p.push_err(toks[i], "invalid_syntax")
	}
	iter.Block = p.build_block(blockToks)
	return ast.St{Token: iter.Token, Data: iter}
}

func (p *parser) buid_iter(bs *block_st) ast.St {
	if bs.terminated {
		return p.get_while_next_iter_profile(bs)
	}
	return p.build_common_iter_profile(bs.toks)
}

func (p *parser) build_case_expressions(toks *[]lex.Token, type_match bool) []ast.Expr {
	var exprs []ast.Expr
	push_expr := func(toks []lex.Token, tok lex.Token) {
		if len(toks) > 0 {
			if type_match {
				i := 0
				t, ok := p.build_type(toks, &i, true)
				if ok {
					exprs = append(exprs, ast.Expr{
						Tokens: toks,
						Op:     t,
					})
				}
				i++
				if i < len(toks) {
					p.push_err(toks[i], "invalid_syntax")
				}
				return
			}
			exprs = append(exprs, BuildExpr(toks))
		}
	}
	brace_n := 0
	j := 0
	var i int
	var tok lex.Token
	for i, tok = range *toks {
		if tok.Id == lex.ID_BRACE {
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
			push_expr((*toks)[j:i], tok)
			j = i + 1
		case tok.Id == lex.ID_COLON:
			push_expr((*toks)[j:i], tok)
			*toks = (*toks)[i+1:]
			return exprs
		}
	}
	p.push_err((*toks)[0], "invalid_syntax")
	*toks = nil
	return nil
}

func (p *parser) build_case_block(toks *[]lex.Token) *ast.Block {
	n := 0
	for {
		next := get_next_st((*toks)[n:])
		if len(next) == 0 {
			break
		}
		tok := next[0]
		if tok.Id != lex.ID_OP || tok.Kind != lex.KND_VLINE {
			n += len(next)
			continue
		}
		block := p.build_block((*toks)[:n])
		*toks = (*toks)[n:]
		return block
	}
	block := p.build_block(*toks)
	*toks = nil
	return block
}

func (p *parser) get_case(toks *[]lex.Token, type_match bool) (ast.Case, bool) {
	var c ast.Case
	c.Token = (*toks)[0]
	*toks = (*toks)[1:]
	c.Exprs = p.build_case_expressions(toks, type_match)
	c.Block = p.build_case_block(toks)
	is_default := len(c.Exprs) == 0
	return c, is_default
}

func (p *parser) build_cases(toks []lex.Token, type_match bool) ([]ast.Case, *ast.Case) {
	var cases []ast.Case
	var def *ast.Case
	for len(toks) > 0 {
		tok := toks[0]
		if tok.Id != lex.ID_OP || tok.Kind != lex.KND_VLINE {
			p.push_err(tok, "invalid_syntax")
			break
		}
		c, is_default := p.get_case(&toks, type_match)
		if is_default {
			c.Token = tok
			if def == nil {
				def = new(ast.Case)
				*def = c
			} else {
				p.push_err(tok, "invalid_syntax")
			}
		} else {
			cases = append(cases, c)
		}
	}
	return cases, def
}

// build_match builds AST model of match-case.
func (p *parser) build_match(toks []lex.Token) (s ast.St) {
	m := new(ast.Match)
	m.Token = toks[0]
	s.Token = m.Token
	toks = toks[1:]
	
	if len(toks) > 0 && toks[0].Id == lex.ID_TYPE {
		m.TypeMatch = true
		toks = toks[1:] // Skip "type" keyword
	}

	exprToks := ast.GetBlockExpr(toks)
	if len(exprToks) > 0 {
		m.Expr = BuildExpr(exprToks)
	} else if m.TypeMatch {
		p.push_err(m.Token, "missing_expr")
	}
	
	i := len(exprToks)
	block_toks := p.get_range(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	if block_toks == nil {
		p.stop()
		p.push_err(m.Token, "body_not_exist")
		return
	}
	
	m.Cases, m.Default = p.build_cases(block_toks, m.TypeMatch)

	for i := range m.Cases {
		c := &m.Cases[i]
		c.Match = m
		if i > 0 {
			m.Cases[i-1].Next = c
		}
	}
	if m.Default != nil {
		if len(m.Cases) > 0 {
			m.Cases[len(m.Cases)-1].Next = m.Default
		}
		m.Default.Match = m
	}

	s.Data = m
	return
}

func (p *parser) build_if(bs *block_st) *ast.If {
	model := new(ast.If)
	model.Token = bs.toks[0]
	bs.toks = bs.toks[1:]
	exprToks := ast.GetBlockExpr(bs.toks)
	i := 0
	if len(exprToks) == 0 {
		p.push_err(model.Token, "missing_expr")
	} else {
		i = len(exprToks)
	}
	blockToks := p.get_range(&i, lex.KND_LBRACE, lex.KND_RBRACE, &bs.toks)
	if blockToks == nil {
		p.stop()
		p.push_err(model.Token, "body_not_exist")
		return nil
	}
	if i < len(bs.toks) {
		if bs.toks[i].Id == lex.ID_ELSE {
			bs.nextToks = bs.toks[i:]
		} else {
			p.push_err(bs.toks[i], "invalid_syntax")
		}
	}
	model.Expr = BuildExpr(exprToks)
	model.Block = p.build_block(blockToks)
	return model
}

func (p *parser) build_else(bs *block_st) *ast.Else {
	model := new(ast.Else)
	model.Token = bs.toks[0]
	bs.toks = bs.toks[1:]
	i := 0
	blockToks := p.get_range(&i, lex.KND_LBRACE, lex.KND_RBRACE, &bs.toks)
	if blockToks == nil {
		if i < len(bs.toks) {
			p.push_err(model.Token, "else_have_expr")
		} else {
			p.stop()
			p.push_err(model.Token, "body_not_exist")
		}
		return nil
	}
	if i < len(bs.toks) {
		p.push_err(bs.toks[i], "invalid_syntax")
	}
	model.Block = p.build_block(blockToks)
	return model
}

// IfExpr builds condition tree AST model.
func (p *parser) build_if_else_chain(bs *block_st) (s ast.St) {
	s.Token = bs.toks[0]
	var c ast.Conditional
	c.If = p.build_if(bs)
	if c.If == nil {
		return
	}
node:
	if bs.terminated || block_st_finished(bs) {
		goto end
	}
	set_to_next_st(bs)
	if bs.toks[0].Id == lex.ID_ELSE {
		if len(bs.toks) > 1 && bs.toks[1].Id == lex.ID_IF {
			bs.toks = bs.toks[1:] // Remove else token
			elif := p.build_if(bs)
			c.Elifs = append(c.Elifs, elif)
			goto node
		}
		c.Default = p.build_else(bs)
	} else {
		// Save statement
		bs.nextToks = bs.toks
	}
end:
	s.Data = c
	return
}

// build_break_st builds AST model of break statement.
func (p *parser) build_break_st(toks []lex.Token) ast.St {
	var breakAST ast.Break
	breakAST.Token = toks[0]
	if len(toks) > 1 {
		if toks[1].Id != lex.ID_IDENT {
			p.push_err(toks[1], "invalid_syntax")
		} else {
			breakAST.LabelToken = toks[1]
			if len(toks) > 2 {
				p.push_err(toks[1], "invalid_syntax")
			}
		}
	}
	return ast.St{
		Token: breakAST.Token,
		Data:  breakAST,
	}
}

// build_continue_st builds AST model of continue statement.
func (p *parser) build_continue_st(toks []lex.Token) ast.St {
	var continueAST ast.Continue
	continueAST.Token = toks[0]
	if len(toks) > 1 {
		if toks[1].Id != lex.ID_IDENT {
			p.push_err(toks[1], "invalid_syntax")
		} else {
			continueAST.LoopLabel = toks[1]
			if len(toks) > 2 {
				p.push_err(toks[1], "invalid_syntax")
			}
		}
	}
	return ast.St{Token: continueAST.Token, Data: continueAST}
}

func (p *parser) get_range(i *int, open, close string, toks *[]lex.Token) []lex.Token {
	rang := ast.Range(i, open, close, *toks)
	if rang != nil {
		return rang
	}
	if p.ended() {
		return nil
	}
	*i = 0
	*toks = p.next_builder_st()
	rang = ast.Range(i, open, close, *toks)
	return rang
}

func (p *parser) skip_st(i *int, toks *[]lex.Token) []lex.Token {
	start := *i
	*i, _ = ast.NextStPos(*toks, start)
	stoks := (*toks)[start:*i]
	if stoks[len(stoks)-1].Id == lex.ID_SEMICOLON {
		if len(stoks) == 1 {
			return p.skip_st(i, toks)
		}
		stoks = stoks[:len(stoks)-1]
	}
	return stoks
}

func (p *parser) next_builder_st() []lex.Token {
	return p.skip_st(&p.pos, &p.tokens)
}
