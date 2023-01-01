package parser

// builder is the AST (Abstract Syntax Tree) builder of Parser.

import (
	"os"
	"strings"
	"sync"

	"github.com/julelang/jule"
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

func compilerErr(t lex.Token, key string, args ...any) build.Log {
	return build.Log{
		Type:    build.ERR,
		Row:     t.Row,
		Column:  t.Column,
		Path:    t.File.Path(),
		Text: build.Errorf(key, args...),
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

type builder struct {
	wg  sync.WaitGroup
	pub bool

	Tree   []ast.Node
	Errors []build.Log
	Tokens []lex.Token
	Pos    int
}

func new_builder(t []lex.Token) *builder {
	b := new(builder)
	b.Tokens = t
	b.Pos = 0
	return b
}

// pusherr appends error by specified token.
func (b *builder) pusherr(t lex.Token, key string, args ...any) {
	b.Errors = append(b.Errors, compilerErr(t, key, args...))
}

// Ended reports position is at end of tokens or not.
func (b *builder) Ended() bool {
	return b.Pos >= len(b.Tokens)
}

func (b *builder) buildNode(toks []lex.Token) {
	t := toks[0]
	switch t.Id {
	case lex.ID_USE:
		b.Use(toks)
	case lex.ID_FN, lex.ID_UNSAFE:
		s := ast.Statement{Token: t}
		s.Data = b.Fn(toks, false, false, false)
		b.Tree = append(b.Tree, ast.Node{Token: s.Token, Data: s})
	case lex.ID_CONST, lex.ID_LET, lex.ID_MUT:
		b.GlobalVar(toks)
	case lex.ID_TYPE:
		b.Tree = append(b.Tree, b.TypeOrGenerics(toks))
	case lex.ID_ENUM:
		b.Enum(toks)
	case lex.ID_STRUCT:
		b.Struct(toks)
	case lex.ID_TRAIT:
		b.Trait(toks)
	case lex.ID_IMPL:
		b.Impl(toks)
	case lex.ID_CPP:
		b.CppLink(toks)
	case lex.ID_COMMENT:
		b.Tree = append(b.Tree, b.Comment(toks[0]))
	default:
		b.pusherr(t, "invalid_syntax")
		return
	}
	if b.pub {
		b.pusherr(t, "def_not_support_pub")
	}
}

// Build builds AST tree.
func (b *builder) Build() {
	for b.Pos != -1 && !b.Ended() {
		toks := b.nextBuilderSt()
		b.pub = toks[0].Id == lex.ID_PUB
		if b.pub {
			if len(toks) == 1 {
				if b.Ended() {
					b.pusherr(toks[0], "invalid_syntax")
					continue
				}
				toks = b.nextBuilderSt()
			} else {
				toks = toks[1:]
			}
		}
		b.buildNode(toks)
	}
	b.Wait()
}

// Wait waits for concurrency.
func (b *builder) Wait() { b.wg.Wait() }

// TypeAlias builds AST model of type definition statement.
func (b *builder) TypeAlias(toks []lex.Token) (t ast.TypeAlias) {
	i := 1 // Initialize value is 1 for skip keyword.
	if i >= len(toks) {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	t.Token = toks[1]
	t.Id = t.Token.Kind
	token := toks[i]
	if token.Id != lex.ID_IDENT {
		b.pusherr(token, "invalid_syntax")
	}
	i++
	if i >= len(toks) {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	token = toks[i]
	if token.Id != lex.ID_COLON {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	i++
	if i >= len(toks) {
		b.pusherr(toks[i-1], "missing_type")
		return
	}
	destType, ok := b.DataType(toks, &i, true)
	t.Type = destType
	if ok && i+1 < len(toks) {
		b.pusherr(toks[i+1], "invalid_syntax")
	}
	return
}

func (b *builder) buildEnumItemExpr(i *int, toks []lex.Token) ast.Expr {
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
			return b.Expr(exprToks)
		}
	}
	return ast.Expr{}
}

func (b *builder) buildEnumItems(toks []lex.Token) []*ast.EnumItem {
	items := make([]*ast.EnumItem, 0)
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		if t.Id == lex.ID_COMMENT {
			continue
		}
		item := new(ast.EnumItem)
		item.Token = t
		if item.Token.Id != lex.ID_IDENT {
			b.pusherr(item.Token, "invalid_syntax")
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
			b.pusherr(toks[0], "invalid_syntax")
		}
		i++
		if i >= len(toks) || toks[i].Id == lex.ID_COMMA {
			b.pusherr(toks[0], "missing_expr")
			continue
		}
		item.Expr = b.buildEnumItemExpr(&i, toks)
		items = append(items, item)
	}
	return items
}

// Enum builds AST model of enumerator statement.
func (b *builder) Enum(toks []lex.Token) {
	var e ast.Enum
	if len(toks) < 2 || len(toks) < 3 {
		b.pusherr(toks[0], "invalid_syntax")
		return
	}
	e.Token = toks[1]
	if e.Token.Id != lex.ID_IDENT {
		b.pusherr(e.Token, "invalid_syntax")
	}
	e.Id = e.Token.Kind
	i := 2
	if toks[i].Id == lex.ID_COLON {
		i++
		if i >= len(toks) {
			b.pusherr(toks[i-1], "invalid_syntax")
			return
		}
		e.Type, _ = b.DataType(toks, &i, true)
		i++
		if i >= len(toks) {
			b.pusherr(e.Token, "body_not_exist")
			return
		}
	} else {
		e.Type = ast.Type{Id: types.U32, Kind: types.TYPE_MAP[types.U32]}
	}
	itemToks := b.getrange(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	if itemToks == nil {
		b.pusherr(e.Token, "body_not_exist")
		return
	} else if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	e.Pub = b.pub
	b.pub = false
	e.Items = b.buildEnumItems(itemToks)
	b.Tree = append(b.Tree, ast.Node{Token: e.Token, Data: e})
}

// Comment builds AST model of comment.
func (b *builder) Comment(t lex.Token) ast.Node {
	t.Kind = strings.TrimSpace(t.Kind[2:])
	return ast.Node{
		Token: t,
		Data: ast.Comment{
			Token:   t,
			Content: t.Kind,
		},
	}
}

func (b *builder) structFields(toks []lex.Token, cpp_linked bool) []*ast.Var {
	var fields []*ast.Var
	i := 0
	for i < len(toks) {
		var_tokens := b.skipSt(&i, &toks)
		if var_tokens[0].Id == lex.ID_COMMENT {
			continue
		}
		is_pub := var_tokens[0].Id == lex.ID_PUB
		if is_pub {
			if len(var_tokens) == 1 {
				b.pusherr(var_tokens[0], "invalid_syntax")
				continue
			}
			var_tokens = var_tokens[1:]
		}
		is_mut := var_tokens[0].Id == lex.ID_MUT
		if is_mut {
			if len(var_tokens) == 1 {
				b.pusherr(var_tokens[0], "invalid_syntax")
				continue
			}
			var_tokens = var_tokens[1:]
		}
		v := b.Var(var_tokens, false, false)
		v.Pub = is_pub
		v.Mutable = is_mut
		v.IsField = true
		v.CppLinked = cpp_linked
		fields = append(fields, &v)
	}
	return fields
}

func (b *builder) parse_struct(toks []lex.Token, cpp_linked bool) ast.Struct {
	var s ast.Struct
	s.Pub = b.pub
	b.pub = false
	if len(toks) < 3 {
		b.pusherr(toks[0], "invalid_syntax")
		return s
	}
	s.Token = toks[1]
	if s.Token.Id != lex.ID_IDENT {
		b.pusherr(s.Token, "invalid_syntax")
	}
	s.Id = s.Token.Kind
	i := 2
	bodyToks := b.getrange(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	if bodyToks == nil {
		b.pusherr(s.Token, "body_not_exist")
		return s
	}
	if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	s.Fields = b.structFields(bodyToks, cpp_linked)
	return s
}

// Struct builds AST model of structure.
func (b *builder) Struct(toks []lex.Token) {
	s := b.parse_struct(toks, false)
	b.Tree = append(b.Tree, ast.Node{Token: s.Token, Data: s})
}

func (b *builder) traitFuncs(toks []lex.Token, trait_id string) []*ast.Fn {
	var funcs []*ast.Fn
	i := 0
	for i < len(toks) {
		fnToks := b.skipSt(&i, &toks)
		f := b.Fn(fnToks, true, false, true)
		b.setup_receiver(&f, trait_id)
		f.Pub = true
		funcs = append(funcs, &f)
	}
	return funcs
}

// Trait builds AST model of trait.
func (b *builder) Trait(toks []lex.Token) {
	var t ast.Trait
	t.Pub = b.pub
	b.pub = false
	if len(toks) < 3 {
		b.pusherr(toks[0], "invalid_syntax")
		return
	}
	t.Token = toks[1]
	if t.Token.Id != lex.ID_IDENT {
		b.pusherr(t.Token, "invalid_syntax")
	}
	t.Id = t.Token.Kind
	i := 2
	bodyToks := b.getrange(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	if bodyToks == nil {
		b.pusherr(t.Token, "body_not_exist")
		return
	}
	if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	t.Funcs = b.traitFuncs(bodyToks, t.Id)
	b.Tree = append(b.Tree, ast.Node{Token: t.Token, Data: t})
}

func (b *builder) implTraitFuncs(impl *ast.Impl, toks []lex.Token) {
	pos, btoks := b.Pos, make([]lex.Token, len(b.Tokens))
	copy(btoks, b.Tokens)
	b.Pos = 0
	b.Tokens = toks
	for b.Pos != -1 && !b.Ended() {
		fnToks := b.nextBuilderSt()
		tok := fnToks[0]
		switch tok.Id {
		case lex.ID_COMMENT:
			impl.Tree = append(impl.Tree, b.Comment(tok))
			continue
		case lex.ID_FN, lex.ID_UNSAFE:
			f := b.get_method(fnToks)
			f.Pub = true
			b.setup_receiver(f, impl.Target.Kind)
			impl.Tree = append(impl.Tree, ast.Node{Token: f.Token, Data: f})
		default:
			b.pusherr(tok, "invalid_syntax")
			continue
		}
	}
	b.Pos, b.Tokens = pos, btoks
}

func (b *builder) implStruct(impl *ast.Impl, toks []lex.Token) {
	pos, btoks := b.Pos, make([]lex.Token, len(b.Tokens))
	copy(btoks, b.Tokens)
	b.Pos = 0
	b.Tokens = toks
	for b.Pos != -1 && !b.Ended() {
		fnToks := b.nextBuilderSt()
		tok := fnToks[0]
		pub := false
		switch tok.Id {
		case lex.ID_COMMENT:
			impl.Tree = append(impl.Tree, b.Comment(tok))
			continue
		case lex.ID_TYPE:
			impl.Tree = append(impl.Tree, ast.Node{
				Token: tok,
				Data:  b.Generics(fnToks),
			})
			continue
		}
		if tok.Id == lex.ID_PUB {
			pub = true
			if len(fnToks) == 1 {
				b.pusherr(fnToks[0], "invalid_syntax")
				continue
			}
			fnToks = fnToks[1:]
			if len(fnToks) > 0 {
				tok = fnToks[0]
			}
		}
		switch tok.Id {
		case lex.ID_FN, lex.ID_UNSAFE:
			f := b.get_method(fnToks)
			f.Pub = pub
			b.setup_receiver(f, impl.Base.Kind)
			impl.Tree = append(impl.Tree, ast.Node{Token: f.Token, Data: f})
		default:
			b.pusherr(tok, "invalid_syntax")
			continue
		}
	}
	b.Pos, b.Tokens = pos, btoks
}

func (b *builder) get_method(toks []lex.Token) *ast.Fn {
	tok := toks[0]
	if tok.Id == lex.ID_UNSAFE {
		toks = toks[1:]
		if len(toks) == 0 || toks[0].Id != lex.ID_FN {
			b.pusherr(tok, "invalid_syntax")
			return nil
		}
	} else if toks[0].Id != lex.ID_FN {
		b.pusherr(tok, "invalid_syntax")
		return nil
	}
	f := new(ast.Fn)
	*f = b.Fn(toks, true, false, false)
	f.IsUnsafe = tok.Id == lex.ID_UNSAFE
	if f.Block != nil {
		f.Block.IsUnsafe = f.IsUnsafe
	}
	return f
}

func (b *builder) implFuncs(impl *ast.Impl, toks []lex.Token) {
	if impl.Target.Id != types.VOID {
		b.implTraitFuncs(impl, toks)
		return
	}
	b.implStruct(impl, toks)
}

// Impl builds AST model of impl statement.
func (b *builder) Impl(toks []lex.Token) {
	tok := toks[0]
	if len(toks) < 2 {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	tok = toks[1]
	if tok.Id != lex.ID_IDENT {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	var impl ast.Impl
	if len(toks) < 3 {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	impl.Base = tok
	tok = toks[2]
	if tok.Id != lex.ID_ITER {
		if tok.Id == lex.ID_BRACE && tok.Kind == lex.KND_LBRACE {
			toks = toks[2:]
			goto body
		}
		b.pusherr(tok, "invalid_syntax")
		return
	}
	if len(toks) < 4 {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	tok = toks[3]
	if tok.Id != lex.ID_IDENT {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	{
		i := 0
		impl.Target, _ = b.DataType(toks[3:4], &i, true)
		toks = toks[4:]
	}
body:
	i := 0
	bodyToks := b.getrange(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	if bodyToks == nil {
		b.pusherr(impl.Base, "body_not_exist")
		return
	}
	if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	b.implFuncs(&impl, bodyToks)
	b.Tree = append(b.Tree, ast.Node{Token: impl.Base, Data: impl})
}

// link_fn builds AST model of cpp function link.
func (b *builder) link_fn(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := b.pub
	b.pub = false

	var link ast.CppLinkFn
	link.Token = tok
	link.Link = new(ast.Fn)
	*link.Link = b.Fn(toks[1:], false, false, true)
	b.Tree = append(b.Tree, ast.Node{Token: tok, Data: link})

	b.pub = bpub
}

// link_var builds AST model of cpp variable link.
func (b *builder) link_var(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := b.pub
	b.pub = false

	var link ast.CppLinkVar
	link.Token = tok
	link.Link = new(ast.Var)
	*link.Link = b.Var(toks[1:], true, false)
	b.Tree = append(b.Tree, ast.Node{Token: tok, Data: link})

	b.pub = bpub
}

// link_struct builds AST model of cpp structure link.
func (b *builder) link_struct(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := b.pub
	b.pub = false

	var link ast.CppLinkStruct
	link.Token = tok
	link.Link = b.parse_struct(toks[1:], true)
	b.Tree = append(b.Tree, ast.Node{Token: tok, Data: link})

	b.pub = bpub
}

// link_type_alias builds AST model of cpp type alias link.
func (b *builder) link_type_alias(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := b.pub
	b.pub = false

	var link ast.CppLinkAlias
	link.Token = tok
	link.Link = b.TypeAlias(toks[1:])
	b.Tree = append(b.Tree, ast.Node{Token: tok, Data: link})

	b.pub = bpub
}

// CppLinks builds AST model of cpp link statement.
func (b *builder) CppLink(toks []lex.Token) {
	tok := toks[0]
	if len(toks) == 1 {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	tok = toks[1]
	switch tok.Id {
	case lex.ID_FN, lex.ID_UNSAFE:
		b.link_fn(toks)
	case lex.ID_LET:
		b.link_var(toks)
	case lex.ID_STRUCT:
		b.link_struct(toks)
	case lex.ID_TYPE:
		b.link_type_alias(toks)
	default:
		b.pusherr(tok, "invalid_syntax")
	}
}

func tokstoa(toks []lex.Token) string {
	var str strings.Builder
	for _, tok := range toks {
		str.WriteString(tok.Kind)
	}
	return str.String()
}

// Use builds AST model of use declaration.
func (b *builder) Use(toks []lex.Token) {
	var use ast.UseDecl
	use.Token = toks[0]
	if len(toks) < 2 {
		b.pusherr(use.Token, "missing_use_path")
		return
	}
	toks = toks[1:]
	b.buildUseDecl(&use, toks)
	b.Tree = append(b.Tree, ast.Node{Token: use.Token, Data: use})
}

func (b *builder) getSelectors(toks []lex.Token) []lex.Token {
	i := 0
	toks = b.getrange(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	parts, errs := ast.Parts(toks, lex.ID_COMMA, true)
	if len(errs) > 0 {
		b.Errors = append(b.Errors, errs...)
		return nil
	}
	selectors := make([]lex.Token, len(parts))
	for i, part := range parts {
		if len(part) > 1 {
			b.pusherr(part[1], "invalid_syntax")
		}
		tok := part[0]
		if tok.Id != lex.ID_IDENT && tok.Id != lex.ID_SELF {
			b.pusherr(tok, "invalid_syntax")
			continue
		}
		selectors[i] = tok
	}
	return selectors
}

func (b *builder) buildUseCppDecl(use *ast.UseDecl, toks []lex.Token) {
	if len(toks) > 2 {
		b.pusherr(toks[2], "invalid_syntax")
	}
	tok := toks[1]
	if tok.Id != lex.ID_LITERAL || (tok.Kind[0] != '`' && tok.Kind[0] != '"') {
		b.pusherr(tok, "invalid_expr")
		return
	}
	use.Cpp = true
	use.Path = tok.Kind[1 : len(tok.Kind)-1]
}

func (b *builder) buildUseDecl(use *ast.UseDecl, toks []lex.Token) {
	var path strings.Builder
	path.WriteString(jule.STDLIB_PATH)
	path.WriteRune(os.PathSeparator)
	tok := toks[0]
	isStd := false
	if tok.Id == lex.ID_CPP {
		b.buildUseCppDecl(use, toks)
		return
	}
	if tok.Id != lex.ID_IDENT || tok.Kind != "std" {
		b.pusherr(toks[0], "invalid_syntax")
	}
	isStd = true
	if len(toks) < 3 {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	toks = toks[2:]
	tok = toks[len(toks)-1]
	switch tok.Id {
	case lex.ID_DBLCOLON:
		b.pusherr(tok, "invalid_syntax")
		return
	case lex.ID_BRACE:
		if tok.Kind != lex.KND_RBRACE {
			b.pusherr(tok, "invalid_syntax")
			return
		}
		var selectors []lex.Token
		toks, selectors = ast.RangeLast(toks)
		use.Selectors = b.getSelectors(selectors)
		if len(toks) == 0 {
			b.pusherr(tok, "invalid_syntax")
			return
		}
		tok = toks[len(toks)-1]
		if tok.Id != lex.ID_DBLCOLON {
			b.pusherr(tok, "invalid_syntax")
			return
		}
		toks = toks[:len(toks)-1]
		if len(toks) == 0 {
			b.pusherr(tok, "invalid_syntax")
			return
		}
	case lex.ID_OP:
		if tok.Kind != lex.KND_STAR {
			b.pusherr(tok, "invalid_syntax")
			return
		}
		toks = toks[:len(toks)-1]
		if len(toks) == 0 {
			b.pusherr(tok, "invalid_syntax")
			return
		}
		tok = toks[len(toks)-1]
		if tok.Id != lex.ID_DBLCOLON {
			b.pusherr(tok, "invalid_syntax")
			return
		}
		toks = toks[:len(toks)-1]
		if len(toks) == 0 {
			b.pusherr(tok, "invalid_syntax")
			return
		}
		use.FullUse = true
	}
	for i, tok := range toks {
		if i%2 != 0 {
			if tok.Id != lex.ID_DBLCOLON {
				b.pusherr(tok, "invalid_syntax")
			}
			path.WriteRune(os.PathSeparator)
			continue
		}
		if tok.Id != lex.ID_IDENT {
			b.pusherr(tok, "invalid_syntax")
		}
		path.WriteString(tok.Kind)
	}
	use.LinkString = tokstoa(toks)
	if isStd {
		use.LinkString = "std::" + use.LinkString
	}
	use.Path = path.String()
}

func (b *builder) setup_receiver(f *ast.Fn, owner_id string) {
	if len(f.Params) == 0 {
		b.pusherr(f.Token, "missing_receiver")
		return
	}
	param := f.Params[0]
	if param.Id != lex.KND_SELF {
		b.pusherr(f.Token, "missing_receiver")
		return
	}
	f.Receiver = new(ast.Var)
	f.Receiver.Type = ast.Type{
		Id:   types.STRUCT,
		Kind: owner_id,
	}
	f.Receiver.Mutable = param.Mutable
	if param.Type.Kind != "" && param.Type.Kind[0] == '&' {
		f.Receiver.Type.Kind = lex.KND_AMPER + f.Receiver.Type.Kind
	}
	f.Params = f.Params[1:]
}

func (b *builder) fn_prototype(toks []lex.Token, i *int, method, anon bool) (f ast.Fn, ok bool) {
	ok = true
	f.Token = toks[*i]
	if f.Token.Id == lex.ID_UNSAFE {
		f.IsUnsafe = true
		*i++
		if *i >= len(toks) {
			b.pusherr(f.Token, "invalid_syntax")
			ok = false
			return
		}
		f.Token = toks[*i]
	}
	// Skips fn tok
	*i++
	if *i >= len(toks) {
		b.pusherr(f.Token, "invalid_syntax")
		ok = false
		return
	}
	f.Pub = b.pub
	b.pub = false
	if anon {
		f.Id = lex.ANONYMOUS_ID
	} else {
		tok := toks[*i]
		if tok.Id != lex.ID_IDENT {
			b.pusherr(tok, "invalid_syntax")
			ok = false
		}
		f.Id = tok.Kind
		*i++
	}
	f.RetType.Type.Id = types.VOID
	f.RetType.Type.Kind = types.TYPE_MAP[f.RetType.Type.Id]
	if *i >= len(toks) {
		b.pusherr(f.Token, "invalid_syntax")
		return
	} else if toks[*i].Kind != lex.KND_LPAREN {
		b.pusherr(toks[*i], "missing_function_parentheses")
		return
	}
	paramToks := b.getrange(i, lex.KND_LPAREN, lex.KND_RPARENT, &toks)
	if len(paramToks) > 0 {
		f.Params = b.Params(paramToks, method, false)
	}
	t, ret_ok := b.FnRetDataType(toks, i)
	if ret_ok {
		f.RetType = t
		*i++
	}
	return
}

// Fn builds AST model of function.
func (b *builder) Fn(toks []lex.Token, method, anon, prototype bool) (f ast.Fn) {
	var ok bool
	i := 0
	f, ok = b.fn_prototype(toks, &i, method, anon)
	if prototype {
		if i+1 < len(toks) {
			b.pusherr(toks[i+1], "invalid_syntax")
		}
		return
	} else if !ok {
		return
	}
	if i >= len(toks) {
		b.pusherr(f.Token, "body_not_exist")
		return
	}
	blockToks := b.getrange(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	if blockToks != nil {
		f.Block = b.Block(blockToks)
		f.Block.IsUnsafe = f.IsUnsafe
		if i < len(toks) {
			b.pusherr(toks[i], "invalid_syntax")
		}
	} else {
		b.pusherr(f.Token, "body_not_exist")
		b.Tokens = append(toks, b.Tokens...)
	}
	return
}

func (b *builder) generic(toks []lex.Token) ast.GenericType {
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
	}
	var gt ast.GenericType
	gt.Token = toks[0]
	if gt.Token.Id != lex.ID_IDENT {
		b.pusherr(gt.Token, "invalid_syntax")
	}
	gt.Id = gt.Token.Kind
	return gt
}

// Generic builds generic type.
func (b *builder) Generics(toks []lex.Token) []ast.GenericType {
	tok := toks[0]
	i := 1
	genericsToks := ast.Range(&i, lex.KND_LBRACKET, lex.KND_RBRACKET, toks)
	if len(genericsToks) == 0 {
		b.pusherr(tok, "missing_expr")
		return make([]ast.GenericType, 0)
	} else if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	parts, errs := ast.Parts(genericsToks, lex.ID_COMMA, true)
	b.Errors = append(b.Errors, errs...)
	generics := make([]ast.GenericType, len(parts))
	for i, part := range parts {
		if len(parts) == 0 {
			continue
		}
		generics[i] = b.generic(part)
	}
	return generics
}

// TypeOrGenerics builds type alias or generics type declaration.
func (b *builder) TypeOrGenerics(toks []lex.Token) ast.Node {
	if len(toks) > 1 {
		tok := toks[1]
		if tok.Id == lex.ID_BRACE && tok.Kind == lex.KND_LBRACKET {
			generics := b.Generics(toks)
			return ast.Node{
				Token: tok,
				Data:  generics,
			}
		}
	}
	t := b.TypeAlias(toks)
	t.Pub = b.pub
	b.pub = false
	return ast.Node{
		Token: t.Token,
		Data:  t,
	}
}

// GlobalVar builds AST model of global variable.
func (b *builder) GlobalVar(toks []lex.Token) {
	if toks == nil {
		return
	}
	bs := block_st{toks: toks}
	s := b.VarSt(&bs, true)
	b.Tree = append(b.Tree, ast.Node{
		Token: s.Token,
		Data:  s,
	})
}

func (b *builder) build_self(toks []lex.Token) (model ast.Param) {
	if len(toks) == 0 {
		return
	}
	i := 0
	if toks[i].Id == lex.ID_MUT {
		model.Mutable = true
		i++
		if i >= len(toks) {
			b.pusherr(toks[i-1], "invalid_syntax")
			return
		}
	}
	if toks[i].Kind == lex.KND_AMPER {
		model.Type.Kind = "&"
		i++
		if i >= len(toks) {
			b.pusherr(toks[i-1], "invalid_syntax")
			return
		}
	}
	if toks[i].Id == lex.ID_SELF {
		model.Id = lex.KND_SELF
		model.Token = toks[i]
		i++
		if i < len(toks) {
			b.pusherr(toks[i+1], "invalid_syntax")
		}
	}
	return
}

// Params builds AST model of function parameters.
func (b *builder) Params(toks []lex.Token, method, mustPure bool) []ast.Param {
	parts, errs := ast.Parts(toks, lex.ID_COMMA, true)
	b.Errors = append(b.Errors, errs...)
	if len(parts) == 0 {
		return nil
	}
	var params []ast.Param
	if method && len(parts) > 0 {
		param := b.build_self(parts[0])
		if param.Id == lex.KND_SELF {
			params = append(params, param)
			parts = parts[1:]
		}
	}
	for _, part := range parts {
		b.pushParam(&params, part, mustPure)
	}
	b.checkParams(&params)
	return params
}

func (b *builder) checkParams(params *[]ast.Param) {
	for i := range *params {
		param := &(*params)[i]
		if param.Id == lex.KND_SELF || param.Type.Token.Id != lex.ID_NA {
			continue
		}
		if param.Token.Id == lex.ID_NA {
			b.pusherr(param.Token, "missing_type")
		} else {
			param.Type.Token = param.Token
			param.Type.Id = types.ID
			param.Type.Kind = param.Type.Token.Kind
			param.Type.Original = param.Type
			param.Id = lex.ANONYMOUS_ID
			param.Token = lex.Token{}
		}
	}
}

func (b *builder) paramTypeBegin(param *ast.Param, i *int, toks []lex.Token) {
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		switch tok.Id {
		case lex.ID_OP:
			switch tok.Kind {
			case lex.KND_TRIPLE_DOT:
				if param.Variadic {
					b.pusherr(tok, "already_variadic")
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

func (b *builder) paramBodyId(param *ast.Param, tok lex.Token) {
	if lex.IsIgnoreId(tok.Kind) {
		param.Id = lex.ANONYMOUS_ID
		return
	}
	param.Id = tok.Kind
}

func (b *builder) paramBody(param *ast.Param, i *int, toks []lex.Token, mustPure bool) {
	b.paramBodyId(param, toks[*i])
	// +1 for skip identifier token
	tok := toks[*i]
	toks = toks[*i+1:]
	if len(toks) == 0 {
		return
	} else if len(toks) < 2 {
		b.pusherr(tok, "missing_type")
		return
	}
	tok = toks[*i]
	if tok.Id != lex.ID_COLON {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	toks = toks[*i+1:] // Skip colon
	b.paramType(param, toks, mustPure)
}

func (b *builder) paramType(param *ast.Param, toks []lex.Token, mustPure bool) {
	i := 0
	if !mustPure {
		b.paramTypeBegin(param, &i, toks)
		if i >= len(toks) {
			return
		}
	}
	param.Type, _ = b.DataType(toks, &i, true)
	i++
	if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
}

func (b *builder) pushParam(params *[]ast.Param, toks []lex.Token, mustPure bool) {
	var param ast.Param
	param.Token = toks[0]
	if param.Token.Id == lex.ID_MUT {
		param.Mutable = true
		if len(toks) == 1 {
			b.pusherr(toks[0], "invalid_syntax")
			return
		}
		toks = toks[1:]
		param.Token = toks[0]
	}
	// Just data type
	if param.Token.Id != lex.ID_IDENT {
		param.Id = lex.ANONYMOUS_ID
		b.paramType(&param, toks, mustPure)
	} else {
		i := 0
		b.paramBody(&param, &i, toks, mustPure)
	}
	*params = append(*params, param)
}

func (b *builder) datatype(t *ast.Type, toks []lex.Token, i *int, err bool) (ok bool) {
	tb := type_builder{
		r:      b,
		t:      t,
		tokens: toks,
		i:      i,
		err:    err,
	}
	return tb.build()
}

// DataType builds AST model of data-type.
func (b *builder) DataType(toks []lex.Token, i *int, err bool) (t ast.Type, ok bool) {
	tok := toks[*i]
	ok = b.datatype(&t, toks, i, err)
	if err && t.Token.Id == lex.ID_NA {
		b.pusherr(tok, "invalid_type")
	}
	return
}

func (b *builder) fnMultiTypeRet(toks []lex.Token, i *int) (t ast.RetType, ok bool) {
	tok := toks[*i]
	t.Type.Kind += tok.Kind
	*i++
	if *i >= len(toks) {
		*i--
		t.Type, ok = b.DataType(toks, i, false)
		return
	}
	tok = toks[*i]
	*i-- // For point to parenthses - ( -
	rang := ast.Range(i, lex.KND_LPAREN, lex.KND_RPARENT, toks)
	params := b.Params(rang, false, true)
	types := make([]ast.Type, len(params))
	for i, param := range params {
		types[i] = param.Type
		if param.Id != lex.ANONYMOUS_ID {
			param.Token.Kind = param.Id
		} else {
			param.Token.Kind = lex.IGNORE_ID
		}
		t.Identifiers = append(t.Identifiers, param.Token)
	}
	if len(types) > 1 {
		t.Type.MultiTyped = true
		t.Type.Tag = types
	} else {
		t.Type = types[0]
	}
	// Decrament for correct block parsing
	*i--
	ok = true
	return
}

// FnRetDataType builds ret data-type of function.
func (b *builder) FnRetDataType(toks []lex.Token, i *int) (t ast.RetType, ok bool) {
	t.Type.Id = types.VOID
	t.Type.Kind = types.TYPE_MAP[t.Type.Id]
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
			b.pusherr(tok, "missing_type")
			return
		}
		*i++
		tok = toks[*i]
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_LPAREN:
				return b.fnMultiTypeRet(toks, i)
			case lex.KND_LBRACE:
				return
			}
		}
		t.Type, ok = b.DataType(toks, i, true)
		return
	}
	*i++
	b.pusherr(tok, "invalid_syntax")
	return
}

func (b *builder) pushStToBlock(bs *block_st) {
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
	s := b.St(bs)
	if s.Data == nil {
		return
	}
	s.WithTerminator = bs.terminated
	bs.block.Tree = append(bs.block.Tree, s)
}

func setToNextSt(bs *block_st) {
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

func blockStFinished(bs *block_st) bool {
	return bs.nextToks == nil && bs.pos >= len(*bs.srcToks)
}

// Block builds AST model of statements of code block.
func (b *builder) Block(toks []lex.Token) (block *ast.Block) {
	block = new(ast.Block)
	var bs block_st
	bs.block = block
	bs.srcToks = &toks
	for {
		setToNextSt(&bs)
		b.pushStToBlock(&bs)
		if blockStFinished(&bs) {
			break
		}
	}
	return
}

// St builds AST model of statement.
func (b *builder) St(bs *block_st) (s ast.Statement) {
	tok := bs.toks[0]
	if tok.Id == lex.ID_IDENT {
		s, ok := b.IdSt(bs)
		if ok {
			return s
		}
	}
	s, ok := b.AssignSt(bs.toks)
	if ok {
		return s
	}
	switch tok.Id {
	case lex.ID_CONST, lex.ID_LET, lex.ID_MUT:
		return b.VarSt(bs, true)
	case lex.ID_RET:
		return b.RetSt(bs.toks)
	case lex.ID_ITER:
		return b.IterExpr(bs)
	case lex.ID_BREAK:
		return b.BreakSt(bs.toks)
	case lex.ID_CONTINUE:
		return b.ContinueSt(bs.toks)
	case lex.ID_IF:
		return b.Conditional(bs)
	case lex.ID_COMMENT:
		return b.CommentSt(bs.toks[0])
	case lex.ID_CO:
		return b.ConcurrentCallSt(bs.toks)
	case lex.ID_GOTO:
		return b.GotoSt(bs.toks)
	case lex.ID_FALLTHROUGH:
		return b.Fallthrough(bs.toks)
	case lex.ID_TYPE:
		t := b.TypeAlias(bs.toks)
		s.Token = t.Token
		s.Data = t
		return
	case lex.ID_MATCH:
		return b.MatchCase(bs.toks)
	case lex.ID_UNSAFE, lex.ID_DEFER:
		return b.blockSt(bs.toks)
	case lex.ID_BRACE:
		if tok.Kind == lex.KND_LBRACE {
			return b.blockSt(bs.toks)
		}
	}
	if ast.IsFnCall(bs.toks) != nil {
		return b.ExprSt(bs)
	}
	b.pusherr(tok, "invalid_syntax")
	return
}

func (b *builder) blockSt(toks []lex.Token) ast.Statement {
	is_unsafe := false
	is_deferred := false
	tok := toks[0]
	if tok.Id == lex.ID_UNSAFE {
		is_unsafe = true
		toks = toks[1:]
		if len(toks) == 0 {
			b.pusherr(tok, "invalid_syntax")
			return ast.Statement{}
		}
		tok = toks[0]
		if tok.Id == lex.ID_DEFER {
			is_deferred = true
			toks = toks[1:]
			if len(toks) == 0 {
				b.pusherr(tok, "invalid_syntax")
				return ast.Statement{}
			}
		}
	} else if tok.Id == lex.ID_DEFER {
		is_deferred = true
		toks = toks[1:]
		if len(toks) == 0 {
			b.pusherr(tok, "invalid_syntax")
			return ast.Statement{}
		}
	}

	i := 0
	toks = ast.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, toks)
	if len(toks) == 0 {
		b.pusherr(tok, "invalid_syntax")
		return ast.Statement{}
	} else if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	block := b.Block(toks)
	block.IsUnsafe = is_unsafe
	block.Deferred = is_deferred
	return ast.Statement{Token: tok, Data: block}
}

func (b *builder) assignInfo(toks []lex.Token) (info ast.AssignInfo) {
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
			b.pusherr(tok, "invalid_syntax")
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
				b.pusherr(info.Right[0], "invalid_syntax")
				info.Right = nil
			}
		}
		break
	}
	return
}

func (b *builder) build_assign_left(toks []lex.Token) (l ast.AssignLeft) {
	l.Expr.Tokens = toks
	if l.Expr.Tokens[0].Id == lex.ID_IDENT {
		l.Var.Token = l.Expr.Tokens[0]
		l.Var.Id = l.Var.Token.Kind
	}
	l.Expr = b.Expr(l.Expr.Tokens)
	return
}

func (b *builder) assignLefts(parts [][]lex.Token) []ast.AssignLeft {
	var lefts []ast.AssignLeft
	for _, p := range parts {
		l := b.build_assign_left(p)
		lefts = append(lefts, l)
	}
	return lefts
}

func (b *builder) assignExprs(toks []lex.Token) []ast.Expr {
	parts, errs := ast.Parts(toks, lex.ID_COMMA, true)
	if len(errs) > 0 {
		b.Errors = append(b.Errors, errs...)
		return nil
	}
	exprs := make([]ast.Expr, len(parts))
	for i, p := range parts {
		exprs[i] = b.Expr(p)
	}
	return exprs
}

// AssignSt builds AST model of assignment statement.
func (b *builder) AssignSt(toks []lex.Token) (s ast.Statement, _ bool) {
	assign, ok := b.AssignExpr(toks)
	if !ok {
		return
	}
	s.Token = toks[0]
	s.Data = assign
	return s, true
}

// AssignExpr builds AST model of assignment expression.
func (b *builder) AssignExpr(toks []lex.Token) (assign ast.Assign, ok bool) {
	if !ast.CheckAssignTokens(toks) {
		return
	}
	switch toks[0].Id {
	case lex.ID_LET:
		return b.letDeclAssign(toks)
	default:
		return b.plainAssign(toks)
	}
}

func (b *builder) letDeclAssign(toks []lex.Token) (assign ast.Assign, ok bool) {
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
		b.pusherr(tok, "invalid_syntax")
		return
	} else if i+1 < len(toks) {
		assign.Setter = toks[i]
		i++
		assign.Right = b.assignExprs(toks[i:])
	}
	parts, errs := ast.Parts(rang, lex.ID_COMMA, true)
	if len(errs) > 0 {
		b.Errors = append(b.Errors, errs...)
		return
	}
	for _, p := range parts {
		if len(p) > 2 {
			b.pusherr(p[2], "invalid_syntax")
		}
		mutable := false
		tok := p[0]
		if tok.Id == lex.ID_MUT {
			mutable = true
			p = p[1:]
			if len(p) == 0 {
				b.pusherr(tok, "invalid_syntax")
				continue
			}
		}
		l := b.build_assign_left(p)
		l.Var.Mutable = mutable
		l.Var.New = !lex.IsIgnoreId(l.Var.Id)
		l.Var.SetterTok = assign.Setter
		assign.Left = append(assign.Left, l)
	}
	return
}

func (b *builder) plainAssign(toks []lex.Token) (assign ast.Assign, ok bool) {
	info := b.assignInfo(toks)
	if !info.Ok {
		return
	}
	ok = true
	assign.Setter = info.Setter
	parts, errs := ast.Parts(info.Left, lex.ID_COMMA, true)
	if len(errs) > 0 {
		b.Errors = append(b.Errors, errs...)
		return
	}
	assign.Left = b.assignLefts(parts)
	if info.Right != nil {
		assign.Right = b.assignExprs(info.Right)
	}
	return
}

// BuildReturnStatement builds AST model of return statement.
func (b *builder) IdSt(bs *block_st) (s ast.Statement, ok bool) {
	if len(bs.toks) == 1 {
		return
	}
	tok := bs.toks[1]
	switch tok.Id {
	case lex.ID_COLON:
		return b.LabelSt(bs), true
	}
	return
}

// LabelSt builds AST model of label.
func (b *builder) LabelSt(bs *block_st) ast.Statement {
	var l ast.Label
	l.Token = bs.toks[0]
	l.Label = l.Token.Kind
	if len(bs.toks) > 2 {
		bs.nextToks = bs.toks[2:]
	}
	return ast.Statement{
		Token: l.Token,
		Data:  l,
	}
}

// ExprSt builds AST model of expression.
func (b *builder) ExprSt(bs *block_st) ast.Statement {
	expr := ast.ExprStatement{
		Expr: b.Expr(bs.toks),
	}
	return ast.Statement{
		Token: bs.toks[0],
		Data:  expr,
	}
}

// Args builds AST model of arguments.
func (b *builder) Args(toks []lex.Token, targeting bool) *ast.Args {
	args := new(ast.Args)
	last := 0
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
		if brace_n > 0 || tok.Id != lex.ID_COMMA {
			continue
		}
		b.pushArg(args, targeting, toks[last:i], tok)
		last = i + 1
	}
	if last < len(toks) {
		if last == 0 {
			if len(toks) > 0 {
				b.pushArg(args, targeting, toks[last:], toks[last])
			}
		} else {
			b.pushArg(args, targeting, toks[last:], toks[last-1])
		}
	}
	return args
}

func (b *builder) pushArg(args *ast.Args, targeting bool, toks []lex.Token, err lex.Token) {
	if len(toks) == 0 {
		b.pusherr(err, "invalid_syntax")
		return
	}
	var arg ast.Arg
	arg.Token = toks[0]
	if targeting && arg.Token.Id == lex.ID_IDENT {
		if len(toks) > 1 {
			tok := toks[1]
			if tok.Id == lex.ID_COLON {
				args.Targeted = true
				arg.TargetId = arg.Token.Kind
				toks = toks[2:]
			}
		}
	}
	arg.Expr = b.Expr(toks)
	args.Src = append(args.Src, arg)
}

func (b *builder) varBegin(v *ast.Var, i *int, toks []lex.Token) {
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
		if v.Const {
			b.pusherr(tok, "already_const")
			break
		}
		v.Const = true
		if !v.Mutable {
			break
		}
		fallthrough
	default:
		b.pusherr(tok, "invalid_syntax")
		return
	}
	if *i >= len(toks) {
		b.pusherr(tok, "invalid_syntax")
	}
}

func (b *builder) varTypeNExpr(v *ast.Var, toks []lex.Token, i int, expr bool) {
	tok := toks[i]
	if tok.Id == lex.ID_COLON {
		i++ // Skip type annotation operator (:)
		if i >= len(toks) ||
			(toks[i].Id == lex.ID_OP && toks[i].Kind == lex.KND_EQ) {
			b.pusherr(tok, "missing_type")
			return
		}
		t, ok := b.DataType(toks, &i, false)
		if ok {
			v.Type = t
			i++
			if i >= len(toks) {
				return
			}
			tok = toks[i]
		}
	}
	if expr && tok.Id == lex.ID_OP {
		if tok.Kind != lex.KND_EQ {
			b.pusherr(tok, "invalid_syntax")
			return
		}
		valueToks := toks[i+1:]
		if len(valueToks) == 0 {
			b.pusherr(tok, "missing_expr")
			return
		}
		v.Expr = b.Expr(valueToks)
		v.SetterTok = tok
	} else {
		b.pusherr(tok, "invalid_syntax")
	}
}

// Var builds AST model of variable statement.
func (b *builder) Var(toks []lex.Token, begin, expr bool) (v ast.Var) {
	v.Pub = b.pub
	b.pub = false
	i := 0
	v.Token = toks[i]
	if begin {
		b.varBegin(&v, &i, toks)
		if i >= len(toks) {
			return
		}
	}
	v.Token = toks[i]
	if v.Token.Id != lex.ID_IDENT {
		b.pusherr(v.Token, "invalid_syntax")
		return
	}
	v.Id = v.Token.Kind
	v.Type.Id = types.VOID
	v.Type.Kind = types.TYPE_MAP[v.Type.Id]
	if i >= len(toks) {
		return
	}
	i++
	if i < len(toks) {
		b.varTypeNExpr(&v, toks, i, expr)
	} else if !expr {
		b.pusherr(v.Token, "missing_type")
	}
	return
}

// VarSt builds AST model of variable declaration statement.
func (b *builder) VarSt(bs *block_st, expr bool) ast.Statement {
	v := b.Var(bs.toks, true, expr)
	v.Owner = bs.block
	return ast.Statement{Token: v.Token, Data: v}
}

// CommentSt builds AST model of comment statement.
func (b *builder) CommentSt(tok lex.Token) (s ast.Statement) {
	s.Token = tok
	tok.Kind = strings.TrimSpace(tok.Kind[2:])
	s.Data = ast.Comment{Content: tok.Kind}
	return
}

func (b *builder) ConcurrentCallSt(toks []lex.Token) (s ast.Statement) {
	var cc ast.ConcurrentCall
	cc.Token = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		b.pusherr(cc.Token, "missing_expr")
		return
	}
	if ast.IsFnCall(toks) == nil {
		b.pusherr(cc.Token, "expr_not_func_call")
	}
	cc.Expr = b.Expr(toks)
	s.Token = cc.Token
	s.Data = cc
	return
}

func (b *builder) Fallthrough(toks []lex.Token) (s ast.Statement) {
	s.Token = toks[0]
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
	}
	s.Data = ast.Fallthrough{
		Token: s.Token,
	}
	return
}

func (b *builder) GotoSt(toks []lex.Token) (s ast.Statement) {
	s.Token = toks[0]
	if len(toks) == 1 {
		b.pusherr(s.Token, "missing_goto_label")
		return
	} else if len(toks) > 2 {
		b.pusherr(toks[2], "invalid_syntax")
	}
	idTok := toks[1]
	if idTok.Id != lex.ID_IDENT {
		b.pusherr(idTok, "invalid_syntax")
		return
	}
	var gt ast.Goto
	gt.Token = s.Token
	gt.Label = idTok.Kind
	s.Data = gt
	return
}

// RetSt builds AST model of return statement.
func (b *builder) RetSt(toks []lex.Token) ast.Statement {
	var ret ast.Ret
	ret.Token = toks[0]
	if len(toks) > 1 {
		ret.Expr = b.Expr(toks[1:])
	}
	return ast.Statement{
		Token: ret.Token,
		Data:  ret,
	}
}

func (b *builder) getWhileIterProfile(toks []lex.Token) ast.IterWhile {
	return ast.IterWhile{
		Expr: b.Expr(toks),
	}
}

func (b *builder) getForeachVarsToks(toks []lex.Token) [][]lex.Token {
	vars, errs := ast.Parts(toks, lex.ID_COMMA, true)
	b.Errors = append(b.Errors, errs...)
	return vars
}

func (b *builder) getVarProfile(toks []lex.Token) (v ast.Var) {
	if len(toks) == 0 {
		return
	}
	v.Token = toks[0]
	if v.Token.Id == lex.ID_MUT {
		v.Mutable = true
		if len(toks) == 1 {
			b.pusherr(v.Token, "invalid_syntax")
		}
		v.Token = toks[1]
	} else if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
	}
	if v.Token.Id != lex.ID_IDENT {
		b.pusherr(v.Token, "invalid_syntax")
		return
	}
	v.Id = v.Token.Kind
	v.New = true
	return
}

func (b *builder) getForeachIterVars(varsToks [][]lex.Token) []ast.Var {
	var vars []ast.Var
	for _, toks := range varsToks {
		vars = append(vars, b.getVarProfile(toks))
	}
	return vars
}

func (b *builder) setup_foreach_explicit_vars(f *ast.IterForeach, toks []lex.Token) {
	i := 0
	rang := ast.Range(&i, lex.KND_LPAREN, lex.KND_RPARENT, toks)
	if i < len(toks) {
		b.pusherr(f.InToken, "invalid_syntax")
	}
	b.setup_foreach_plain_vars(f, rang)
}

func (b *builder) setup_foreach_plain_vars(f *ast.IterForeach, toks []lex.Token) {
	varsToks := b.getForeachVarsToks(toks)
	if len(varsToks) == 0 {
		return
	}
	if len(varsToks) > 2 {
		b.pusherr(f.InToken, "much_foreach_vars")
	}
	vars := b.getForeachIterVars(varsToks)
	f.KeyA = vars[0]
	if len(vars) > 1 {
		f.KeyB = vars[1]
	} else {
		f.KeyB.Id = lex.IGNORE_ID
	}
}

func (b *builder) setup_foreach_vars(f *ast.IterForeach, toks []lex.Token) {
	if toks[0].Id == lex.ID_BRACE {
		if toks[0].Kind != lex.KND_LPAREN {
			b.pusherr(toks[0], "invalid_syntax")
			return
		}
		b.setup_foreach_explicit_vars(f, toks)
		return
	}
	b.setup_foreach_plain_vars(f, toks)
}

func (b *builder) getForeachIterProfile(varToks, exprToks []lex.Token, inTok lex.Token) ast.IterForeach {
	var foreach ast.IterForeach
	foreach.InToken = inTok
	if len(exprToks) == 0 {
		b.pusherr(inTok, "missing_expr")
		return foreach
	}
	foreach.Expr = b.Expr(exprToks)
	if len(varToks) == 0 {
		foreach.KeyA.Id = lex.IGNORE_ID
		foreach.KeyB.Id = lex.IGNORE_ID
	} else {
		b.setup_foreach_vars(&foreach, varToks)
	}
	return foreach
}

func (b *builder) getIterProfile(toks []lex.Token, errtok lex.Token) any {
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
			return b.getForeachIterProfile(varToks, exprToks, tok)
		}
	}
	return b.getWhileIterProfile(toks)
}

func (b *builder) next_st(toks []lex.Token) ast.Statement {
	s := b.St(&block_st{toks: toks})
	switch s.Data.(type) {
	case ast.ExprStatement, ast.Assign, ast.Var:
	default:
		b.pusherr(toks[0], "invalid_syntax")
	}
	return s
}

func (b *builder) getWhileNextIterProfile(bs *block_st) (s ast.Statement) {
	var iter ast.Iter
	iter.Token = bs.toks[0]
	bs.toks = bs.toks[1:]
	profile := ast.IterWhile{}
	if len(bs.toks) > 0 {
		profile.Expr = b.Expr(bs.toks)
	}
	if blockStFinished(bs) {
		b.pusherr(iter.Token, "invalid_syntax")
		return
	}
	setToNextSt(bs)
	st_toks := ast.BlockExpr(bs.toks)
	if len(st_toks) > 0 {
		profile.Next = b.next_st(st_toks)
	}
	i := len(st_toks)
	blockToks := b.getrange(&i, lex.KND_LBRACE, lex.KND_RBRACE, &bs.toks)
	if blockToks == nil {
		b.pusherr(iter.Token, "body_not_exist")
		return
	}
	if i < len(bs.toks) {
		b.pusherr(bs.toks[i], "invalid_syntax")
	}
	iter.Block = b.Block(blockToks)
	iter.Profile = profile
	return ast.Statement{Token: iter.Token, Data: iter}
}

func (b *builder) commonIterProfile(toks []lex.Token) (s ast.Statement) {
	var iter ast.Iter
	iter.Token = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		b.pusherr(iter.Token, "body_not_exist")
		return
	}
	exprToks := ast.BlockExpr(toks)
	if len(exprToks) > 0 {
		iter.Profile = b.getIterProfile(exprToks, iter.Token)
	}
	i := len(exprToks)
	blockToks := b.getrange(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	if blockToks == nil {
		b.pusherr(iter.Token, "body_not_exist")
		return
	}
	if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	iter.Block = b.Block(blockToks)
	return ast.Statement{Token: iter.Token, Data: iter}
}

func (b *builder) IterExpr(bs *block_st) ast.Statement {
	if bs.terminated {
		return b.getWhileNextIterProfile(bs)
	}
	return b.commonIterProfile(bs.toks)
}

func (b *builder) caseexprs(toks *[]lex.Token, caseIsDefault bool) []ast.Expr {
	var exprs []ast.Expr
	pushExpr := func(toks []lex.Token, tok lex.Token) {
		if caseIsDefault {
			if len(toks) > 0 {
				b.pusherr(tok, "invalid_syntax")
			}
			return
		}
		if len(toks) > 0 {
			exprs = append(exprs, b.Expr(toks))
			return
		}
		b.pusherr(tok, "missing_expr")
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
		switch tok.Id {
		case lex.ID_COMMA:
			pushExpr((*toks)[j:i], tok)
			j = i + 1
		case lex.ID_COLON:
			pushExpr((*toks)[j:i], tok)
			*toks = (*toks)[i+1:]
			return exprs
		}
	}
	b.pusherr((*toks)[0], "invalid_syntax")
	*toks = nil
	return nil
}

func (b *builder) caseblock(toks *[]lex.Token) *ast.Block {
	brace_n := 0
	for i, tok := range *toks {
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
		switch tok.Id {
		case lex.ID_CASE, lex.ID_DEFAULT:
			blockToks := (*toks)[:i]
			*toks = (*toks)[i:]
			return b.Block(blockToks)
		}
	}
	block := b.Block(*toks)
	*toks = nil
	return block
}

func (b *builder) getcase(toks *[]lex.Token) ast.Case {
	var c ast.Case
	c.Token = (*toks)[0]
	*toks = (*toks)[1:]
	c.Exprs = b.caseexprs(toks, c.Token.Id == lex.ID_DEFAULT)
	c.Block = b.caseblock(toks)
	return c
}

func (b *builder) cases(toks []lex.Token) ([]ast.Case, *ast.Case) {
	var cases []ast.Case
	var def *ast.Case
	for len(toks) > 0 {
		tok := toks[0]
		switch tok.Id {
		case lex.ID_CASE:
			cases = append(cases, b.getcase(&toks))
		case lex.ID_DEFAULT:
			c := b.getcase(&toks)
			c.Token = tok
			if def == nil {
				def = new(ast.Case)
				*def = c
				break
			}
			fallthrough
		default:
			b.pusherr(tok, "invalid_syntax")
		}
	}
	return cases, def
}

// MatchCase builds AST model of match-case.
func (b *builder) MatchCase(toks []lex.Token) (s ast.Statement) {
	m := new(ast.Match)
	m.Token = toks[0]
	s.Token = m.Token
	toks = toks[1:]
	exprToks := ast.BlockExpr(toks)
	if len(exprToks) > 0 {
		m.Expr = b.Expr(exprToks)
	}
	i := len(exprToks)
	blockToks := b.getrange(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	if blockToks == nil {
		b.pusherr(m.Token, "body_not_exist")
		return
	}
	m.Cases, m.Default = b.cases(blockToks)
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

func (b *builder) if_expr(bs *block_st) *ast.If {
	model := new(ast.If)
	model.Token = bs.toks[0]
	bs.toks = bs.toks[1:]
	exprToks := ast.BlockExpr(bs.toks)
	i := 0
	if len(exprToks) == 0 {
		b.pusherr(model.Token, "missing_expr")
	} else {
		i = len(exprToks)
	}
	blockToks := b.getrange(&i, lex.KND_LBRACE, lex.KND_RBRACE, &bs.toks)
	if blockToks == nil {
		b.pusherr(model.Token, "body_not_exist")
		return nil
	}
	if i < len(bs.toks) {
		if bs.toks[i].Id == lex.ID_ELSE {
			bs.nextToks = bs.toks[i:]
		} else {
			b.pusherr(bs.toks[i], "invalid_syntax")
		}
	}
	model.Expr = b.Expr(exprToks)
	model.Block = b.Block(blockToks)
	return model
}

func (b *builder) conditional_default(bs *block_st) *ast.Else {
	model := new(ast.Else)
	model.Token = bs.toks[0]
	bs.toks = bs.toks[1:]
	i := 0
	blockToks := b.getrange(&i, lex.KND_LBRACE, lex.KND_RBRACE, &bs.toks)
	if blockToks == nil {
		if i < len(bs.toks) {
			b.pusherr(model.Token, "else_have_expr")
		} else {
			b.pusherr(model.Token, "body_not_exist")
		}
		return nil
	}
	if i < len(bs.toks) {
		b.pusherr(bs.toks[i], "invalid_syntax")
	}
	model.Block = b.Block(blockToks)
	return model
}

// IfExpr builds condition tree AST model.
func (b *builder) Conditional(bs *block_st) (s ast.Statement) {
	s.Token = bs.toks[0]
	var c ast.Conditional
	c.If = b.if_expr(bs)
	if c.If == nil {
		return
	}

node:
	if bs.terminated {
		goto end
	}
	if blockStFinished(bs) {
		goto end
	}
	setToNextSt(bs)
	if bs.toks[0].Id == lex.ID_ELSE {
		if len(bs.toks) > 1 && bs.toks[1].Id == lex.ID_IF {
			bs.toks = bs.toks[1:] // Remove else token
			elif := b.if_expr(bs)
			c.Elifs = append(c.Elifs, elif)
			goto node
		}
		c.Default = b.conditional_default(bs)
	} else {
		// Save statement
		bs.nextToks = bs.toks
	}

end:
	s.Data = c
	return
}

// BreakSt builds AST model of break statement.
func (b *builder) BreakSt(toks []lex.Token) ast.Statement {
	var breakAST ast.Break
	breakAST.Token = toks[0]
	if len(toks) > 1 {
		if toks[1].Id != lex.ID_IDENT {
			b.pusherr(toks[1], "invalid_syntax")
		} else {
			breakAST.LabelToken = toks[1]
			if len(toks) > 2 {
				b.pusherr(toks[1], "invalid_syntax")
			}
		}
	}
	return ast.Statement{
		Token: breakAST.Token,
		Data:  breakAST,
	}
}

// ContinueSt builds AST model of continue statement.
func (b *builder) ContinueSt(toks []lex.Token) ast.Statement {
	var continueAST ast.Continue
	continueAST.Token = toks[0]
	if len(toks) > 1 {
		if toks[1].Id != lex.ID_IDENT {
			b.pusherr(toks[1], "invalid_syntax")
		} else {
			continueAST.LoopLabel = toks[1]
			if len(toks) > 2 {
				b.pusherr(toks[1], "invalid_syntax")
			}
		}
	}
	return ast.Statement{Token: continueAST.Token, Data: continueAST}
}

// Expr builds AST model of expression.
func (b *builder) Expr(toks []lex.Token) (e ast.Expr) {
	e.Op = b.build_expr_op(toks)
	e.Tokens = toks
	return
}

func (b *builder) build_binop_expr(toks []lex.Token) any {
	i := b.find_lowest_precedenced_operator(toks)
	if i != -1 {
		return b.build_binop(toks)
	}
	return ast.BinopExpr{Tokens: toks}
}

func (b *builder) build_binop(toks []lex.Token) ast.Binop {
	op := ast.Binop{}
	i := b.find_lowest_precedenced_operator(toks)
	op.L = b.build_binop_expr(toks[:i])
	op.R = b.build_binop_expr(toks[i+1:])
	op.Op = toks[i]
	return op
}

func eliminate_comments(toks []lex.Token) []lex.Token {
	cutted := []lex.Token{}
	for _, token := range toks {
		if token.Id != lex.ID_COMMENT {
			cutted = append(cutted, token)
		}
	}
	return cutted
}

// Returns BinopExpr or Binop instance for expression Op.
func (b *builder) build_expr_op(toks []lex.Token) any {
	toks = eliminate_comments(toks)
	i := b.find_lowest_precedenced_operator(toks)
	if i == -1 {
		return b.build_binop_expr(toks)
	}
	return b.build_binop(toks)
}

// nextOperator find index of priority operator and returns index of operator
// if found, returns -1 if not.
func (b *builder) find_lowest_precedenced_operator(toks []lex.Token) int {
	prec := precedencer{}
	brace_n := 0
	for i, tok := range toks {
		switch {
		case tok.Id == lex.ID_BRACE:
			switch tok.Kind {
			case lex.KND_LBRACE, lex.KND_LPAREN, lex.KND_LBRACKET:
				brace_n++
			default:
				brace_n--
			}
			continue
		case i == 0:
			continue
		case tok.Id != lex.ID_OP:
			continue
		case brace_n > 0:
			continue
		}
		// Skip unary operator.
		if toks[i-1].Id == lex.ID_OP {
			continue
		}
		switch tok.Kind {
		case lex.KND_STAR, lex.KND_PERCENT, lex.KND_SOLIDUS,
			lex.KND_RSHIFT, lex.KND_LSHIFT, lex.KND_AMPER:
			prec.set(5, i)
		case lex.KND_PLUS, lex.KND_MINUS, lex.KND_VLINE, lex.KND_CARET:
			prec.set(4, i)
		case lex.KND_EQS, lex.KND_NOT_EQ, lex.KND_LT,
			lex.KND_LESS_EQ, lex.KND_GT, lex.KND_GREAT_EQ:
			prec.set(3, i)
		case lex.KND_DBL_AMPER:
			prec.set(2, i)
		case lex.KND_DBL_VLINE:
			prec.set(1, i)
		}
	}
	data := prec.get_lower()
	if data == nil {
		return -1
	}
	return data.(int)
}

func (b *builder) getrange(i *int, open, close string, toks *[]lex.Token) []lex.Token {
	rang := ast.Range(i, open, close, *toks)
	if rang != nil {
		return rang
	}
	if b.Ended() {
		return nil
	}
	*i = 0
	*toks = b.nextBuilderSt()
	rang = ast.Range(i, open, close, *toks)
	return rang
}

func (b *builder) skipSt(i *int, toks *[]lex.Token) []lex.Token {
	start := *i
	*i, _ = ast.NextStPos(*toks, start)
	stoks := (*toks)[start:*i]
	if stoks[len(stoks)-1].Id == lex.ID_SEMICOLON {
		if len(stoks) == 1 {
			return b.skipSt(i, toks)
		}
		stoks = stoks[:len(stoks)-1]
	}
	return stoks
}

func (b *builder) nextBuilderSt() []lex.Token {
	return b.skipSt(&b.Pos, &b.Tokens)
}
