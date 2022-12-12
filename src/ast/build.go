package ast

import (
	"os"
	"strings"
	"sync"

	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/pkg/jule"
	"github.com/julelang/jule/pkg/juleapi"
	"github.com/julelang/jule/pkg/julelog"
	"github.com/julelang/jule/pkg/juletype"
)

// Builder is builds AST tree.
type Builder struct {
	wg  sync.WaitGroup
	pub bool

	Tree   []models.Object
	Errors []julelog.CompilerLog
	Tokens []lex.Token
	Pos    int
}

// NewBuilder instance.
func NewBuilder(t []lex.Token) *Builder {
	b := new(Builder)
	b.Tokens = t
	b.Pos = 0
	return b
}

func compilerErr(t lex.Token, key string, args ...any) julelog.CompilerLog {
	return julelog.CompilerLog{
		Type:    julelog.ERR,
		Row:     t.Row,
		Column:  t.Column,
		Path:    t.File.Path(),
		Message: jule.GetError(key, args...),
	}
}

// pusherr appends error by specified token.
func (b *Builder) pusherr(t lex.Token, key string, args ...any) {
	b.Errors = append(b.Errors, compilerErr(t, key, args...))
}

// Ended reports position is at end of tokens or not.
func (b *Builder) Ended() bool {
	return b.Pos >= len(b.Tokens)
}

func (b *Builder) buildNode(toks []lex.Token) {
	t := toks[0]
	switch t.Id {
	case lex.ID_USE:
		b.Use(toks)
	case lex.ID_FN, lex.ID_UNSAFE:
		s := models.Statement{Token: t}
		s.Data = b.Fn(toks, false, false, false)
		b.Tree = append(b.Tree, models.Object{Token: s.Token, Data: s})
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
func (b *Builder) Build() {
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
func (b *Builder) Wait() { b.wg.Wait() }

// TypeAlias builds AST model of type definition statement.
func (b *Builder) TypeAlias(toks []lex.Token) (t models.TypeAlias) {
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

func (b *Builder) buildEnumItemExpr(i *int, toks []lex.Token) models.Expr {
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
	return models.Expr{}
}

func (b *Builder) buildEnumItems(toks []lex.Token) []*models.EnumItem {
	items := make([]*models.EnumItem, 0)
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		if t.Id == lex.ID_COMMENT {
			continue
		}
		item := new(models.EnumItem)
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
func (b *Builder) Enum(toks []lex.Token) {
	var e models.Enum
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
		e.Type = models.Type{Id: juletype.U32, Kind: juletype.TYPE_MAP[juletype.U32]}
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
	b.Tree = append(b.Tree, models.Object{Token: e.Token, Data: e})
}

// Comment builds AST model of comment.
func (b *Builder) Comment(t lex.Token) models.Object {
	t.Kind = strings.TrimSpace(t.Kind[2:])
	return models.Object{
		Token: t,
		Data: models.Comment{
			Token:   t,
			Content: t.Kind,
		},
	}
}

func (b *Builder) structFields(toks []lex.Token, cpp_linked bool) []*models.Var {
	var fields []*models.Var
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

func (b *Builder) parse_struct(toks []lex.Token, cpp_linked bool) models.Struct {
	var s models.Struct
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
func (b *Builder) Struct(toks []lex.Token) {
	s := b.parse_struct(toks, false)
	b.Tree = append(b.Tree, models.Object{Token: s.Token, Data: s})
}

func (b *Builder) traitFuncs(toks []lex.Token, trait_id string) []*models.Fn {
	var funcs []*models.Fn
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
func (b *Builder) Trait(toks []lex.Token) {
	var t models.Trait
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
	b.Tree = append(b.Tree, models.Object{Token: t.Token, Data: t})
}

func (b *Builder) implTraitFuncs(impl *models.Impl, toks []lex.Token) {
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
			impl.Tree = append(impl.Tree, models.Object{Token: f.Token, Data: f})
		default:
			b.pusherr(tok, "invalid_syntax")
			continue
		}
	}
	b.Pos, b.Tokens = pos, btoks
}

func (b *Builder) implStruct(impl *models.Impl, toks []lex.Token) {
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
			impl.Tree = append(impl.Tree, models.Object{
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
			impl.Tree = append(impl.Tree, models.Object{Token: f.Token, Data: f})
		default:
			b.pusherr(tok, "invalid_syntax")
			continue
		}
	}
	b.Pos, b.Tokens = pos, btoks
}

func (b *Builder) get_method(toks []lex.Token) *models.Fn {
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
	f := new(models.Fn)
	*f = b.Fn(toks, true, false, false)
	f.IsUnsafe = tok.Id == lex.ID_UNSAFE
	if f.Block != nil {
		f.Block.IsUnsafe = f.IsUnsafe
	}
	return f
}

func (b *Builder) implFuncs(impl *models.Impl, toks []lex.Token) {
	if impl.Target.Id != juletype.VOID {
		b.implTraitFuncs(impl, toks)
		return
	}
	b.implStruct(impl, toks)
}

// Impl builds AST model of impl statement.
func (b *Builder) Impl(toks []lex.Token) {
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
	var impl models.Impl
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
	b.Tree = append(b.Tree, models.Object{Token: impl.Base, Data: impl})
}

// link_fn builds AST model of cpp function link.
func (b *Builder) link_fn(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := b.pub
	b.pub = false

	var link models.CppLinkFn
	link.Token = tok
	link.Link = new(models.Fn)
	*link.Link = b.Fn(toks[1:], false, false, true)
	b.Tree = append(b.Tree, models.Object{Token: tok, Data: link})

	b.pub = bpub
}

// link_var builds AST model of cpp variable link.
func (b *Builder) link_var(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := b.pub
	b.pub = false

	var link models.CppLinkVar
	link.Token = tok
	link.Link = new(models.Var)
	*link.Link = b.Var(toks[1:], true, false)
	b.Tree = append(b.Tree, models.Object{Token: tok, Data: link})

	b.pub = bpub
}

// link_struct builds AST model of cpp structure link.
func (b *Builder) link_struct(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := b.pub
	b.pub = false

	var link models.CppLinkStruct
	link.Token = tok
	link.Link = b.parse_struct(toks[1:], true)
	b.Tree = append(b.Tree, models.Object{Token: tok, Data: link})

	b.pub = bpub
}

// link_type_alias builds AST model of cpp type alias link.
func (b *Builder) link_type_alias(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := b.pub
	b.pub = false

	var link models.CppLinkAlias
	link.Token = tok
	link.Link = b.TypeAlias(toks[1:])
	b.Tree = append(b.Tree, models.Object{Token: tok, Data: link})

	b.pub = bpub
}

// CppLinks builds AST model of cpp link statement.
func (b *Builder) CppLink(toks []lex.Token) {
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
func (b *Builder) Use(toks []lex.Token) {
	var use models.UseDecl
	use.Token = toks[0]
	if len(toks) < 2 {
		b.pusherr(use.Token, "missing_use_path")
		return
	}
	toks = toks[1:]
	b.buildUseDecl(&use, toks)
	b.Tree = append(b.Tree, models.Object{Token: use.Token, Data: use})
}

func (b *Builder) getSelectors(toks []lex.Token) []lex.Token {
	i := 0
	toks = b.getrange(&i, lex.KND_LBRACE, lex.KND_RBRACE, &toks)
	parts, errs := Parts(toks, lex.ID_COMMA, true)
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

func (b *Builder) buildUseCppDecl(use *models.UseDecl, toks []lex.Token) {
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

func (b *Builder) buildUseDecl(use *models.UseDecl, toks []lex.Token) {
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
		toks, selectors = RangeLast(toks)
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

func (b *Builder) setup_receiver(f *models.Fn, owner_id string) {
	if len(f.Params) == 0 {
		b.pusherr(f.Token, "missing_receiver")
		return
	}
	param := f.Params[0]
	if param.Id != lex.KND_SELF {
		b.pusherr(f.Token, "missing_receiver")
		return
	}
	f.Receiver = new(models.Var)
	f.Receiver.Type = models.Type{
		Id:   juletype.STRUCT,
		Kind: owner_id,
	}
	f.Receiver.Mutable = param.Mutable
	if param.Type.Kind != "" && param.Type.Kind[0] == '&' {
		f.Receiver.Type.Kind = lex.KND_AMPER + f.Receiver.Type.Kind
	}
	f.Params = f.Params[1:]
}

func (b *Builder) fn_prototype(toks []lex.Token, i *int, method, anon bool) (f models.Fn, ok bool) {
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
		f.Id = jule.ANONYMOUS
	} else {
		tok := toks[*i]
		if tok.Id != lex.ID_IDENT {
			b.pusherr(tok, "invalid_syntax")
			ok = false
		}
		f.Id = tok.Kind
		*i++
	}
	f.RetType.Type.Id = juletype.VOID
	f.RetType.Type.Kind = juletype.TYPE_MAP[f.RetType.Type.Id]
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
func (b *Builder) Fn(toks []lex.Token, method, anon, prototype bool) (f models.Fn) {
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

func (b *Builder) generic(toks []lex.Token) models.GenericType {
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
	}
	var gt models.GenericType
	gt.Token = toks[0]
	if gt.Token.Id != lex.ID_IDENT {
		b.pusherr(gt.Token, "invalid_syntax")
	}
	gt.Id = gt.Token.Kind
	return gt
}

// Generic builds generic type.
func (b *Builder) Generics(toks []lex.Token) []models.GenericType {
	tok := toks[0]
	i := 1
	genericsToks := Range(&i, lex.KND_LBRACKET, lex.KND_RBRACKET, toks)
	if len(genericsToks) == 0 {
		b.pusherr(tok, "missing_expr")
		return make([]models.GenericType, 0)
	} else if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	parts, errs := Parts(genericsToks, lex.ID_COMMA, true)
	b.Errors = append(b.Errors, errs...)
	generics := make([]models.GenericType, len(parts))
	for i, part := range parts {
		if len(parts) == 0 {
			continue
		}
		generics[i] = b.generic(part)
	}
	return generics
}

// TypeOrGenerics builds type alias or generics type declaration.
func (b *Builder) TypeOrGenerics(toks []lex.Token) models.Object {
	if len(toks) > 1 {
		tok := toks[1]
		if tok.Id == lex.ID_BRACE && tok.Kind == lex.KND_LBRACKET {
			generics := b.Generics(toks)
			return models.Object{
				Token: tok,
				Data:  generics,
			}
		}
	}
	t := b.TypeAlias(toks)
	t.Pub = b.pub
	b.pub = false
	return models.Object{
		Token: t.Token,
		Data:  t,
	}
}

// GlobalVar builds AST model of global variable.
func (b *Builder) GlobalVar(toks []lex.Token) {
	if toks == nil {
		return
	}
	bs := block_st{toks: toks}
	s := b.VarSt(&bs, true)
	b.Tree = append(b.Tree, models.Object{
		Token: s.Token,
		Data:  s,
	})
}

func (b *Builder) build_self(toks []lex.Token) (model models.Param) {
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
func (b *Builder) Params(toks []lex.Token, method, mustPure bool) []models.Param {
	parts, errs := Parts(toks, lex.ID_COMMA, true)
	b.Errors = append(b.Errors, errs...)
	if len(parts) == 0 {
		return nil
	}
	var params []models.Param
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

func (b *Builder) checkParams(params *[]models.Param) {
	for i := range *params {
		param := &(*params)[i]
		if param.Id == lex.KND_SELF || param.Type.Token.Id != lex.ID_NA {
			continue
		}
		if param.Token.Id == lex.ID_NA {
			b.pusherr(param.Token, "missing_type")
		} else {
			param.Type.Token = param.Token
			param.Type.Id = juletype.ID
			param.Type.Kind = param.Type.Token.Kind
			param.Type.Original = param.Type
			param.Id = jule.ANONYMOUS
			param.Token = lex.Token{}
		}
	}
}

func (b *Builder) paramTypeBegin(param *models.Param, i *int, toks []lex.Token) {
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

func (b *Builder) paramBodyId(param *models.Param, tok lex.Token) {
	if juleapi.IsIgnoreId(tok.Kind) {
		param.Id = jule.ANONYMOUS
		return
	}
	param.Id = tok.Kind
}

func (b *Builder) paramBody(param *models.Param, i *int, toks []lex.Token, mustPure bool) {
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

func (b *Builder) paramType(param *models.Param, toks []lex.Token, mustPure bool) {
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

func (b *Builder) pushParam(params *[]models.Param, toks []lex.Token, mustPure bool) {
	var param models.Param
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
		param.Id = jule.ANONYMOUS
		b.paramType(&param, toks, mustPure)
	} else {
		i := 0
		b.paramBody(&param, &i, toks, mustPure)
	}
	*params = append(*params, param)
}

func (b *Builder) datatype(t *models.Type, toks []lex.Token, i *int, err bool) (ok bool) {
	tb := type_builder{
		b:      b,
		t:      t,
		tokens: toks,
		i:      i,
		err:    err,
	}
	return tb.build()
}

// DataType builds AST model of data-type.
func (b *Builder) DataType(toks []lex.Token, i *int, err bool) (t models.Type, ok bool) {
	tok := toks[*i]
	ok = b.datatype(&t, toks, i, err)
	if err && t.Token.Id == lex.ID_NA {
		b.pusherr(tok, "invalid_type")
	}
	return
}

func (b *Builder) fnMultiTypeRet(toks []lex.Token, i *int) (t models.RetType, ok bool) {
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
	rang := Range(i, lex.KND_LPAREN, lex.KND_RPARENT, toks)
	params := b.Params(rang, false, true)
	types := make([]models.Type, len(params))
	for i, param := range params {
		types[i] = param.Type
		if param.Id != jule.ANONYMOUS {
			param.Token.Kind = param.Id
		} else {
			param.Token.Kind = juleapi.IGNORE
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
func (b *Builder) FnRetDataType(toks []lex.Token, i *int) (t models.RetType, ok bool) {
	t.Type.Id = juletype.VOID
	t.Type.Kind = juletype.TYPE_MAP[t.Type.Id]
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

func (b *Builder) pushStToBlock(bs *block_st) {
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
	bs.pos, bs.terminated = NextStPos(*bs.srcToks, 0)
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
func (b *Builder) Block(toks []lex.Token) (block *models.Block) {
	block = new(models.Block)
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
func (b *Builder) St(bs *block_st) (s models.Statement) {
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
	if IsFnCall(bs.toks) != nil {
		return b.ExprSt(bs)
	}
	b.pusherr(tok, "invalid_syntax")
	return
}

func (b *Builder) blockSt(toks []lex.Token) models.Statement {
	is_unsafe := false
	is_deferred := false
	tok := toks[0]
	if tok.Id == lex.ID_UNSAFE {
		is_unsafe = true
		toks = toks[1:]
		if len(toks) == 0 {
			b.pusherr(tok, "invalid_syntax")
			return models.Statement{}
		}
		tok = toks[0]
		if tok.Id == lex.ID_DEFER {
			is_deferred = true
			toks = toks[1:]
			if len(toks) == 0 {
				b.pusherr(tok, "invalid_syntax")
				return models.Statement{}
			}
		}
	} else if tok.Id == lex.ID_DEFER {
		is_deferred = true
		toks = toks[1:]
		if len(toks) == 0 {
			b.pusherr(tok, "invalid_syntax")
			return models.Statement{}
		}
	}

	i := 0
	toks = Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, toks)
	if len(toks) == 0 {
		b.pusherr(tok, "invalid_syntax")
		return models.Statement{}
	} else if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	block := b.Block(toks)
	block.IsUnsafe = is_unsafe
	block.Deferred = is_deferred
	return models.Statement{Token: tok, Data: block}
}

func (b *Builder) assignInfo(toks []lex.Token) (info AssignInfo) {
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
		} else if !IsAssignOp(tok.Kind) {
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
			info.Ok = IsPostfixOp(info.Setter.Kind)
			break
		}
		info.Right = toks[i+1:]
		if IsPostfixOp(info.Setter.Kind) {
			if info.Right != nil {
				b.pusherr(info.Right[0], "invalid_syntax")
				info.Right = nil
			}
		}
		break
	}
	return
}

func (b *Builder) build_assign_left(toks []lex.Token) (l models.AssignLeft) {
	l.Expr.Tokens = toks
	if l.Expr.Tokens[0].Id == lex.ID_IDENT {
		l.Var.Token = l.Expr.Tokens[0]
		l.Var.Id = l.Var.Token.Kind
	}
	l.Expr = b.Expr(l.Expr.Tokens)
	return
}

func (b *Builder) assignLefts(parts [][]lex.Token) []models.AssignLeft {
	var lefts []models.AssignLeft
	for _, p := range parts {
		l := b.build_assign_left(p)
		lefts = append(lefts, l)
	}
	return lefts
}

func (b *Builder) assignExprs(toks []lex.Token) []models.Expr {
	parts, errs := Parts(toks, lex.ID_COMMA, true)
	if len(errs) > 0 {
		b.Errors = append(b.Errors, errs...)
		return nil
	}
	exprs := make([]models.Expr, len(parts))
	for i, p := range parts {
		exprs[i] = b.Expr(p)
	}
	return exprs
}

// AssignSt builds AST model of assignment statement.
func (b *Builder) AssignSt(toks []lex.Token) (s models.Statement, _ bool) {
	assign, ok := b.AssignExpr(toks)
	if !ok {
		return
	}
	s.Token = toks[0]
	s.Data = assign
	return s, true
}

// AssignExpr builds AST model of assignment expression.
func (b *Builder) AssignExpr(toks []lex.Token) (assign models.Assign, ok bool) {
	if !CheckAssignTokens(toks) {
		return
	}
	switch toks[0].Id {
	case lex.ID_LET:
		return b.letDeclAssign(toks)
	default:
		return b.plainAssign(toks)
	}
}

func (b *Builder) letDeclAssign(toks []lex.Token) (assign models.Assign, ok bool) {
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
	rang := Range(&i, lex.KND_LPAREN, lex.KND_RPARENT, toks)
	if rang == nil {
		b.pusherr(tok, "invalid_syntax")
		return
	} else if i+1 < len(toks) {
		assign.Setter = toks[i]
		i++
		assign.Right = b.assignExprs(toks[i:])
	}
	parts, errs := Parts(rang, lex.ID_COMMA, true)
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
		l.Var.New = !juleapi.IsIgnoreId(l.Var.Id)
		l.Var.SetterTok = assign.Setter
		assign.Left = append(assign.Left, l)
	}
	return
}

func (b *Builder) plainAssign(toks []lex.Token) (assign models.Assign, ok bool) {
	info := b.assignInfo(toks)
	if !info.Ok {
		return
	}
	ok = true
	assign.Setter = info.Setter
	parts, errs := Parts(info.Left, lex.ID_COMMA, true)
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
func (b *Builder) IdSt(bs *block_st) (s models.Statement, ok bool) {
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
func (b *Builder) LabelSt(bs *block_st) models.Statement {
	var l models.Label
	l.Token = bs.toks[0]
	l.Label = l.Token.Kind
	if len(bs.toks) > 2 {
		bs.nextToks = bs.toks[2:]
	}
	return models.Statement{
		Token: l.Token,
		Data: l,
	}
}

// ExprSt builds AST model of expression.
func (b *Builder) ExprSt(bs *block_st) models.Statement {
	expr := models.ExprStatement{
		Expr: b.Expr(bs.toks),
	}
	return models.Statement{
		Token: bs.toks[0],
		Data:  expr,
	}
}

// Args builds AST model of arguments.
func (b *Builder) Args(toks []lex.Token, targeting bool) *models.Args {
	args := new(models.Args)
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

func (b *Builder) pushArg(args *models.Args, targeting bool, toks []lex.Token, err lex.Token) {
	if len(toks) == 0 {
		b.pusherr(err, "invalid_syntax")
		return
	}
	var arg models.Arg
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

func (b *Builder) varBegin(v *models.Var, i *int, toks []lex.Token) {
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

func (b *Builder) varTypeNExpr(v *models.Var, toks []lex.Token, i int, expr bool) {
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
func (b *Builder) Var(toks []lex.Token, begin, expr bool) (v models.Var) {
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
	v.Type.Id = juletype.VOID
	v.Type.Kind = juletype.TYPE_MAP[v.Type.Id]
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
func (b *Builder) VarSt(bs *block_st, expr bool) models.Statement {
	v := b.Var(bs.toks, true, expr)
	v.Owner = bs.block
	return models.Statement{Token: v.Token, Data: v}
}

// CommentSt builds AST model of comment statement.
func (b *Builder) CommentSt(tok lex.Token) (s models.Statement) {
	s.Token = tok
	tok.Kind = strings.TrimSpace(tok.Kind[2:])
	s.Data = models.Comment{Content: tok.Kind}
	return
}

func (b *Builder) ConcurrentCallSt(toks []lex.Token) (s models.Statement) {
	var cc models.ConcurrentCall
	cc.Token = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		b.pusherr(cc.Token, "missing_expr")
		return
	}
	if IsFnCall(toks) == nil {
		b.pusherr(cc.Token, "expr_not_func_call")
	}
	cc.Expr = b.Expr(toks)
	s.Token = cc.Token
	s.Data = cc
	return
}

func (b *Builder) Fallthrough(toks []lex.Token) (s models.Statement) {
	s.Token = toks[0]
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
	}
	s.Data = models.Fallthrough{
		Token: s.Token,
	}
	return
}

func (b *Builder) GotoSt(toks []lex.Token) (s models.Statement) {
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
	var gt models.Goto
	gt.Token = s.Token
	gt.Label = idTok.Kind
	s.Data = gt
	return
}

// RetSt builds AST model of return statement.
func (b *Builder) RetSt(toks []lex.Token) models.Statement {
	var ret models.Ret
	ret.Token = toks[0]
	if len(toks) > 1 {
		ret.Expr = b.Expr(toks[1:])
	}
	return models.Statement{
		Token: ret.Token,
		Data:  ret,
	}
}

func (b *Builder) getWhileIterProfile(toks []lex.Token) models.IterWhile {
	return models.IterWhile{
		Expr: b.Expr(toks),
	}
}

func (b *Builder) getForeachVarsToks(toks []lex.Token) [][]lex.Token {
	vars, errs := Parts(toks, lex.ID_COMMA, true)
	b.Errors = append(b.Errors, errs...)
	return vars
}

func (b *Builder) getVarProfile(toks []lex.Token) (v models.Var) {
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

func (b *Builder) getForeachIterVars(varsToks [][]lex.Token) []models.Var {
	var vars []models.Var
	for _, toks := range varsToks {
		vars = append(vars, b.getVarProfile(toks))
	}
	return vars
}

func (b *Builder) setup_foreach_explicit_vars(f *models.IterForeach, toks []lex.Token) {
	i := 0
	rang := Range(&i, lex.KND_LPAREN, lex.KND_RPARENT, toks)
	if i < len(toks) {
		b.pusherr(f.InToken, "invalid_syntax")
	}
	b.setup_foreach_plain_vars(f, rang)
}

func (b *Builder) setup_foreach_plain_vars(f *models.IterForeach, toks []lex.Token) {
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
		f.KeyB.Id = juleapi.IGNORE
	}
}

func (b *Builder) setup_foreach_vars(f *models.IterForeach, toks []lex.Token) {
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

func (b *Builder) getForeachIterProfile(varToks, exprToks []lex.Token, inTok lex.Token) models.IterForeach {
	var foreach models.IterForeach
	foreach.InToken = inTok
	if len(exprToks) == 0 {
		b.pusherr(inTok, "missing_expr")
		return foreach
	}
	foreach.Expr = b.Expr(exprToks)
	if len(varToks) == 0 {
		foreach.KeyA.Id = juleapi.IGNORE
		foreach.KeyB.Id = juleapi.IGNORE
	} else {
		b.setup_foreach_vars(&foreach, varToks)
	}
	return foreach
}

func (b *Builder) getIterProfile(toks []lex.Token, errtok lex.Token) models.IterProfile {
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

func (b *Builder) next_st(toks []lex.Token) models.Statement {
	s := b.St(&block_st{toks: toks})
	switch s.Data.(type) {
	case models.ExprStatement, models.Assign, models.Var:
	default:
		b.pusherr(toks[0], "invalid_syntax")
	}
	return s
}

func (b *Builder) getWhileNextIterProfile(bs *block_st) (s models.Statement) {
	var iter models.Iter
	iter.Token = bs.toks[0]
	bs.toks = bs.toks[1:]
	profile := models.IterWhile{}
	if len(bs.toks) > 0 {
		profile.Expr = b.Expr(bs.toks)
	}
	if blockStFinished(bs) {
		b.pusherr(iter.Token, "invalid_syntax")
		return
	}
	setToNextSt(bs)
	st_toks := BlockExpr(bs.toks)
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
	return models.Statement{Token: iter.Token, Data: iter}
}

func (b *Builder) commonIterProfile(toks []lex.Token) (s models.Statement) {
	var iter models.Iter
	iter.Token = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		b.pusherr(iter.Token, "body_not_exist")
		return
	}
	exprToks := BlockExpr(toks)
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
	return models.Statement{Token: iter.Token, Data: iter}
}

func (b *Builder) IterExpr(bs *block_st) models.Statement {
	if bs.terminated {
		return b.getWhileNextIterProfile(bs)
	}
	return b.commonIterProfile(bs.toks)
}

func (b *Builder) caseexprs(toks *[]lex.Token, caseIsDefault bool) []models.Expr {
	var exprs []models.Expr
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

func (b *Builder) caseblock(toks *[]lex.Token) *models.Block {
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

func (b *Builder) getcase(toks *[]lex.Token) models.Case {
	var c models.Case
	c.Token = (*toks)[0]
	*toks = (*toks)[1:]
	c.Exprs = b.caseexprs(toks, c.Token.Id == lex.ID_DEFAULT)
	c.Block = b.caseblock(toks)
	return c
}

func (b *Builder) cases(toks []lex.Token) ([]models.Case, *models.Case) {
	var cases []models.Case
	var def *models.Case
	for len(toks) > 0 {
		tok := toks[0]
		switch tok.Id {
		case lex.ID_CASE:
			cases = append(cases, b.getcase(&toks))
		case lex.ID_DEFAULT:
			c := b.getcase(&toks)
			c.Token = tok
			if def == nil {
				def = new(models.Case)
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
func (b *Builder) MatchCase(toks []lex.Token) (s models.Statement) {
	m := new(models.Match)
	m.Token = toks[0]
	s.Token = m.Token
	toks = toks[1:]
	exprToks := BlockExpr(toks)
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

func (b *Builder) if_expr(bs *block_st) *models.If {
	model := new(models.If)
	model.Token = bs.toks[0]
	bs.toks = bs.toks[1:]
	exprToks := BlockExpr(bs.toks)
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

func (b *Builder) conditional_default(bs *block_st) *models.Else {
	model := new(models.Else)
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
func (b *Builder) Conditional(bs *block_st) (s models.Statement) {
	s.Token = bs.toks[0]
	var c models.Conditional
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
func (b *Builder) BreakSt(toks []lex.Token) models.Statement {
	var breakAST models.Break
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
	return models.Statement{
		Token: breakAST.Token,
		Data:  breakAST,
	}
}

// ContinueSt builds AST model of continue statement.
func (b *Builder) ContinueSt(toks []lex.Token) models.Statement {
	var continueAST models.Continue
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
	return models.Statement{Token: continueAST.Token, Data:  continueAST}
}

// Expr builds AST model of expression.
func (b *Builder) Expr(toks []lex.Token) (e models.Expr) {
	e.Op = b.build_expr_op(toks)
	e.Tokens = toks
	return
}

func (b *Builder) build_binop_expr(toks []lex.Token) any {
	i := b.find_lowest_precedenced_operator(toks)
	if i != -1 {
		return b.build_binop(toks)
	}
	return models.BinopExpr{Tokens: toks}
}

func (b *Builder) build_binop(toks []lex.Token) models.Binop {
	op := models.Binop{}
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
func (b *Builder) build_expr_op(toks []lex.Token) any {
	toks = eliminate_comments(toks)
	i := b.find_lowest_precedenced_operator(toks)
	if i == -1 {
		return b.build_binop_expr(toks)
	}
	return b.build_binop(toks)
}

// nextOperator find index of priority operator and returns index of operator
// if found, returns -1 if not.
func (b *Builder) find_lowest_precedenced_operator(toks []lex.Token) int {
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

func (b *Builder) getrange(i *int, open, close string, toks *[]lex.Token) []lex.Token {
	rang := Range(i, open, close, *toks)
	if rang != nil {
		return rang
	}
	if b.Ended() {
		return nil
	}
	*i = 0
	*toks = b.nextBuilderSt()
	rang = Range(i, open, close, *toks)
	return rang
}

func (b *Builder) skipSt(i *int, toks *[]lex.Token) []lex.Token {
	start := *i
	*i, _ = NextStPos(*toks, start)
	stoks := (*toks)[start:*i]
	if stoks[len(stoks)-1].Id == lex.ID_SEMICOLON {
		if len(stoks) == 1 {
			return b.skipSt(i, toks)
		}
		stoks = stoks[:len(stoks)-1]
	}
	return stoks
}

func (b *Builder) nextBuilderSt() []lex.Token {
	return b.skipSt(&b.Pos, &b.Tokens)
}
