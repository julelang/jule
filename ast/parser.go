package ast

import (
	"os"
	"strings"
	"sync"

	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/jule"
	"github.com/jule-lang/jule/pkg/juleapi"
	"github.com/jule-lang/jule/pkg/julelog"
	"github.com/jule-lang/jule/pkg/juletype"
)

// Parser is builds AST tree.
type Parser struct {
	wg  sync.WaitGroup
	pub bool

	Tree   []models.Object
	Errors []julelog.CompilerLog
	Tokens []lex.Token
	Pos    int
}

// NewBuilder instance.
func NewBuilder(t []lex.Token) *Parser {
	b := new(Parser)
	b.Tokens = t
	b.Pos = 0
	return b
}

func compilerErr(t lex.Token, key string, args ...any) julelog.CompilerLog {
	return julelog.CompilerLog{
		Type:    julelog.Error,
		Row:     t.Row,
		Column:  t.Column,
		Path:    t.File.Path(),
		Message: jule.GetError(key, args...),
	}
}

// pusherr appends error by specified token.
func (p *Parser) pusherr(t lex.Token, key string, args ...any) {
	p.Errors = append(p.Errors, compilerErr(t, key, args...))
}

// Ended reports position is at end of tokens or not.
func (ast *Parser) Ended() bool {
	return ast.Pos >= len(ast.Tokens)
}

func (p *Parser) buildNode(toks []lex.Token) {
	t := toks[0]
	switch t.Id {
	case tokens.Use:
		p.Use(toks)
	case tokens.Fn, tokens.Unsafe:
		s := models.Statement{Token: t}
		s.Data = p.Func(toks, false, false, false)
		p.Tree = append(p.Tree, models.Object{Token: s.Token, Data: s})
	case tokens.Const, tokens.Let, tokens.Mut:
		p.GlobalVar(toks)
	case tokens.Type:
		p.Tree = append(p.Tree, p.TypeOrGenerics(toks))
	case tokens.Enum:
		p.Enum(toks)
	case tokens.Struct:
		p.Struct(toks)
	case tokens.Trait:
		p.Trait(toks)
	case tokens.Impl:
		p.Impl(toks)
	case tokens.Cpp:
		p.CppLink(toks)
	case tokens.Comment:
		p.Tree = append(p.Tree, p.Comment(toks[0]))
	default:
		p.pusherr(t, "invalid_syntax")
		return
	}
	if p.pub {
		p.pusherr(t, "def_not_support_pub")
	}
}

// Build builds AST tree.
func (p *Parser) Build() {
	for p.Pos != -1 && !p.Ended() {
		toks := p.nextBuilderStatement()
		p.pub = toks[0].Id == tokens.Pub
		if p.pub {
			if len(toks) == 1 {
				if p.Ended() {
					p.pusherr(toks[0], "invalid_syntax")
					continue
				}
				toks = p.nextBuilderStatement()
			} else {
				toks = toks[1:]
			}
		}
		p.buildNode(toks)
	}
	p.Wait()
}

// Wait waits for concurrency.
func (p *Parser) Wait() { p.wg.Wait() }

// Type builds AST model of type definition statement.
func (p *Parser) Type(toks []lex.Token) (t models.TypeAlias) {
	i := 1 // Initialize value is 1 for skip keyword.
	if i >= len(toks) {
		p.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	t.Token = toks[1]
	t.Id = t.Token.Kind
	token := toks[i]
	if token.Id != tokens.Id {
		p.pusherr(token, "invalid_syntax")
	}
	i++
	if i >= len(toks) {
		p.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	token = toks[i]
	if token.Id != tokens.Colon {
		p.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	i++
	if i >= len(toks) {
		p.pusherr(toks[i-1], "missing_type")
		return
	}
	destType, ok := p.DataType(toks, &i, true, true)
	t.Type = destType
	if ok && i+1 < len(toks) {
		p.pusherr(toks[i+1], "invalid_syntax")
	}
	return
}

func (p *Parser) buildEnumItemExpr(i *int, toks []lex.Token) models.Expr {
	brace_n := 0
	exprStart := *i
	for ; *i < len(toks); *i++ {
		t := toks[*i]
		if t.Id == tokens.Brace {
			switch t.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				brace_n++
				continue
			default:
				brace_n--
			}
		}
		if brace_n > 0 {
			continue
		}
		if t.Id == tokens.Comma || *i+1 >= len(toks) {
			var exprToks []lex.Token
			if t.Id == tokens.Comma {
				exprToks = toks[exprStart:*i]
			} else {
				exprToks = toks[exprStart:]
			}
			return p.Expr(exprToks)
		}
	}
	return models.Expr{}
}

func (p *Parser) buildEnumItems(toks []lex.Token) []*models.EnumItem {
	items := make([]*models.EnumItem, 0)
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		if t.Id == tokens.Comment {
			continue
		}
		item := new(models.EnumItem)
		item.Token = t
		if item.Token.Id != tokens.Id {
			p.pusherr(item.Token, "invalid_syntax")
		}
		item.Id = item.Token.Kind
		if i+1 >= len(toks) || toks[i+1].Id == tokens.Comma {
			if i+1 < len(toks) {
				i++
			}
			items = append(items, item)
			continue
		}
		i++
		t = toks[i]
		if t.Id != tokens.Operator && t.Kind != tokens.EQUAL {
			p.pusherr(toks[0], "invalid_syntax")
		}
		i++
		if i >= len(toks) || toks[i].Id == tokens.Comma {
			p.pusherr(toks[0], "missing_expr")
			continue
		}
		item.Expr = p.buildEnumItemExpr(&i, toks)
		items = append(items, item)
	}
	return items
}

// Enum builds AST model of enumerator statement.
func (p *Parser) Enum(toks []lex.Token) {
	var e models.Enum
	if len(toks) < 2 || len(toks) < 3 {
		p.pusherr(toks[0], "invalid_syntax")
		return
	}
	e.Token = toks[1]
	if e.Token.Id != tokens.Id {
		p.pusherr(e.Token, "invalid_syntax")
	}
	e.Id = e.Token.Kind
	i := 2
	if toks[i].Id == tokens.Colon {
		i++
		if i >= len(toks) {
			p.pusherr(toks[i-1], "invalid_syntax")
			return
		}
		e.Type, _ = p.DataType(toks, &i, false, true)
		i++
		if i >= len(toks) {
			p.pusherr(e.Token, "body_not_exist")
			return
		}
	} else {
		e.Type = models.Type{Id: juletype.U32, Kind: tokens.U32}
	}
	itemToks := p.getrange(&i, tokens.LBRACE, tokens.RBRACE, &toks)
	if itemToks == nil {
		p.pusherr(e.Token, "body_not_exist")
		return
	} else if i < len(toks) {
		p.pusherr(toks[i], "invalid_syntax")
	}
	e.Pub = p.pub
	p.pub = false
	e.Items = p.buildEnumItems(itemToks)
	p.Tree = append(p.Tree, models.Object{Token: e.Token, Data: e})
}

// Comment builds AST model of comment.
func (p *Parser) Comment(t lex.Token) models.Object {
	t.Kind = strings.TrimSpace(t.Kind[2:])
	return models.Object{
		Token: t,
		Data: models.Comment{
			Token:   t,
			Content: t.Kind,
		},
	}
}

func (p *Parser) structFields(toks []lex.Token, cpp_linked bool) []*models.Var {
	var fields []*models.Var
	i := 0
	for i < len(toks) {
		var_tokens := p.skipStatement(&i, &toks)
		if var_tokens[0].Id == tokens.Comment {
			continue
		}
		is_pub := var_tokens[0].Id == tokens.Pub
		if is_pub {
			if len(var_tokens) == 1 {
				p.pusherr(var_tokens[0], "invalid_syntax")
				continue
			}
			var_tokens = var_tokens[1:]
		}
		is_mut := var_tokens[0].Id == tokens.Mut
		if is_mut {
			if len(var_tokens) == 1 {
				p.pusherr(var_tokens[0], "invalid_syntax")
				continue
			}
			var_tokens = var_tokens[1:]
		}
		v := p.Var(var_tokens, false, false)
		v.Pub = is_pub
		v.Mutable = is_mut
		v.IsField = true
		v.CppLinked = cpp_linked
		fields = append(fields, &v)
	}
	return fields
}

func (p *Parser) parse_struct(toks []lex.Token, cpp_linked bool) models.Struct {
	var s models.Struct
	s.Pub = p.pub
	p.pub = false
	if len(toks) < 3 {
		p.pusherr(toks[0], "invalid_syntax")
		return s
	}
	s.Token = toks[1]
	if s.Token.Id != tokens.Id {
		p.pusherr(s.Token, "invalid_syntax")
	}
	s.Id = s.Token.Kind
	i := 2
	bodyToks := p.getrange(&i, tokens.LBRACE, tokens.RBRACE, &toks)
	if bodyToks == nil {
		p.pusherr(s.Token, "body_not_exist")
		return s
	}
	if i < len(toks) {
		p.pusherr(toks[i], "invalid_syntax")
	}
	s.Fields = p.structFields(bodyToks, cpp_linked)
	return s
}

// Struct builds AST model of structure.
func (p *Parser) Struct(toks []lex.Token) {
	s := p.parse_struct(toks, false)
	p.Tree = append(p.Tree, models.Object{Token: s.Token, Data: s})
}

func (p *Parser) traitFuncs(toks []lex.Token, trait_id string) []*models.Fn {
	var funcs []*models.Fn
	i := 0
	for i < len(toks) {
		fnToks := p.skipStatement(&i, &toks)
		f := p.Func(fnToks, true, false, true)
		p.setup_receiver(&f, trait_id)
		f.Pub = true
		funcs = append(funcs, &f)
	}
	return funcs
}

// Trait builds AST model of trait.
func (p *Parser) Trait(toks []lex.Token) {
	var t models.Trait
	t.Pub = p.pub
	p.pub = false
	if len(toks) < 3 {
		p.pusherr(toks[0], "invalid_syntax")
		return
	}
	t.Token = toks[1]
	if t.Token.Id != tokens.Id {
		p.pusherr(t.Token, "invalid_syntax")
	}
	t.Id = t.Token.Kind
	i := 2
	bodyToks := p.getrange(&i, tokens.LBRACE, tokens.RBRACE, &toks)
	if bodyToks == nil {
		p.pusherr(t.Token, "body_not_exist")
		return
	}
	if i < len(toks) {
		p.pusherr(toks[i], "invalid_syntax")
	}
	t.Funcs = p.traitFuncs(bodyToks, t.Id)
	p.Tree = append(p.Tree, models.Object{Token: t.Token, Data: t})
}

func (p *Parser) implTraitFuncs(impl *models.Impl, toks []lex.Token) {
	pos, btoks := p.Pos, make([]lex.Token, len(p.Tokens))
	copy(btoks, p.Tokens)
	defer func() { p.Pos, p.Tokens = pos, btoks }()
	p.Pos = 0
	p.Tokens = toks
	for p.Pos != -1 && !p.Ended() {
		fnToks := p.nextBuilderStatement()
		tok := fnToks[0]
		switch tok.Id {
		case tokens.Comment:
			impl.Tree = append(impl.Tree, p.Comment(tok))
			continue
		case tokens.Fn, tokens.Unsafe:
			f := p.get_method(fnToks)
			f.Pub = true
			p.setup_receiver(f, impl.Target.Kind)
			impl.Tree = append(impl.Tree, models.Object{Token: f.Token, Data: f})
		default:
			p.pusherr(tok, "invalid_syntax")
			continue
		}
	}
}

func (p *Parser) implStruct(impl *models.Impl, toks []lex.Token) {
	pos, btoks := p.Pos, make([]lex.Token, len(p.Tokens))
	copy(btoks, p.Tokens)
	defer func() { p.Pos, p.Tokens = pos, btoks }()
	p.Pos = 0
	p.Tokens = toks
	for p.Pos != -1 && !p.Ended() {
		fnToks := p.nextBuilderStatement()
		tok := fnToks[0]
		pub := false
		switch tok.Id {
		case tokens.Comment:
			impl.Tree = append(impl.Tree, p.Comment(tok))
			continue
		case tokens.Type:
			impl.Tree = append(impl.Tree, models.Object{
				Token: tok,
				Data:  p.Generics(fnToks),
			})
			continue
		}
		if tok.Id == tokens.Pub {
			pub = true
			if len(fnToks) == 1 {
				p.pusherr(fnToks[0], "invalid_syntax")
				continue
			}
			fnToks = fnToks[1:]
			if len(fnToks) > 0 {
				tok = fnToks[0]
			}
		}
		switch tok.Id {
		case tokens.Fn, tokens.Unsafe:
			f := p.get_method(fnToks)
			f.Pub = pub
			p.setup_receiver(f, impl.Base.Kind)
			impl.Tree = append(impl.Tree, models.Object{Token: f.Token, Data: f})
		default:
			p.pusherr(tok, "invalid_syntax")
			continue
		}
	}
}

func (p *Parser) get_method(toks []lex.Token) *models.Fn {
	tok := toks[0]
	if tok.Id == tokens.Unsafe {
		toks = toks[1:]
		if len(toks) == 0 || toks[0].Id != tokens.Fn {
			p.pusherr(tok, "invalid_syntax")
			return nil
		}
	} else if toks[0].Id != tokens.Fn {
		p.pusherr(tok, "invalid_syntax")
		return nil
	}
	f := new(models.Fn)
	*f = p.Func(toks, true, false, false)
	f.IsUnsafe = tok.Id == tokens.Unsafe
	if f.Block != nil {
		f.Block.IsUnsafe = f.IsUnsafe
	}
	return f
}

func (p *Parser) implFuncs(impl *models.Impl, toks []lex.Token) {
	if impl.Target.Id != juletype.Void {
		p.implTraitFuncs(impl, toks)
		return
	}
	p.implStruct(impl, toks)
}

// Impl builds AST model of impl statement.
func (p *Parser) Impl(toks []lex.Token) {
	tok := toks[0]
	if len(toks) < 2 {
		p.pusherr(tok, "invalid_syntax")
		return
	}
	tok = toks[1]
	if tok.Id != tokens.Id {
		p.pusherr(tok, "invalid_syntax")
		return
	}
	var impl models.Impl
	if len(toks) < 3 {
		p.pusherr(tok, "invalid_syntax")
		return
	}
	impl.Base = tok
	tok = toks[2]
	if tok.Id != tokens.For {
		if tok.Id == tokens.Brace && tok.Kind == tokens.LBRACE {
			toks = toks[2:]
			goto body
		}
		p.pusherr(tok, "invalid_syntax")
		return
	}
	if len(toks) < 4 {
		p.pusherr(tok, "invalid_syntax")
		return
	}
	tok = toks[3]
	if tok.Id != tokens.Id {
		p.pusherr(tok, "invalid_syntax")
		return
	}
	{
		i := 0
		impl.Target, _ = p.DataType(toks[3:4], &i, false, true)
		toks = toks[4:]
	}
body:
	i := 0
	bodyToks := p.getrange(&i, tokens.LBRACE, tokens.RBRACE, &toks)
	if bodyToks == nil {
		p.pusherr(impl.Base, "body_not_exist")
		return
	}
	if i < len(toks) {
		p.pusherr(toks[i], "invalid_syntax")
	}
	p.implFuncs(&impl, bodyToks)
	p.Tree = append(p.Tree, models.Object{Token: impl.Base, Data: impl})
}

// link_fn builds AST model of cpp function link.
func (p *Parser) link_fn(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := p.pub

	var link models.CppLinkFn
	link.Token = tok
	link.Link = new(models.Fn)
	*link.Link = p.Func(toks[1:], false, false, true)
	p.Tree = append(p.Tree, models.Object{Token: tok, Data: link})

	p.pub = bpub
}

// link_var builds AST model of cpp variable link.
func (p *Parser) link_var(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := p.pub

	var link models.CppLinkVar
	link.Token = tok
	link.Link = new(models.Var)
	*link.Link = p.Var(toks[1:], true, false)
	p.Tree = append(p.Tree, models.Object{Token: tok, Data: link})

	p.pub = bpub
}

// link_struct builds AST model of cpp structure link.
func (p *Parser) link_struct(toks []lex.Token) {
	tok := toks[0]

	// Catch pub not supported
	bpub := p.pub

	var link models.CppLinkStruct
	link.Token = tok
	link.Link = p.parse_struct(toks[1:], true)
	p.Tree = append(p.Tree, models.Object{Token: tok, Data: link})

	p.pub = bpub
}

// CppLinks builds AST model of cpp link statement.
func (p *Parser) CppLink(toks []lex.Token) {
	tok := toks[0]
	if len(toks) == 1 {
		p.pusherr(tok, "invalid_syntax")
		return
	}
	tok = toks[1]
	switch tok.Id {
	case tokens.Fn:
		p.link_fn(toks)
	case tokens.Let:
		p.link_var(toks)
	case tokens.Struct:
		p.link_struct(toks)
	default:
		p.pusherr(tok, "invalid_syntax")
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
func (p *Parser) Use(toks []lex.Token) {
	var use models.UseDecl
	use.Token = toks[0]
	if len(toks) < 2 {
		p.pusherr(use.Token, "missing_use_path")
		return
	}
	toks = toks[1:]
	p.buildUseDecl(&use, toks)
	p.Tree = append(p.Tree, models.Object{Token: use.Token, Data: use})
}

func (p *Parser) getSelectors(toks []lex.Token) []lex.Token {
	i := 0
	toks = p.getrange(&i, tokens.LBRACE, tokens.RBRACE, &toks)
	parts, errs := Parts(toks, tokens.Comma, true)
	if len(errs) > 0 {
		p.Errors = append(p.Errors, errs...)
		return nil
	}
	selectors := make([]lex.Token, len(parts))
	for i, part := range parts {
		if len(part) > 1 {
			p.pusherr(part[1], "invalid_syntax")
		}
		tok := part[0]
		if tok.Id != tokens.Id && tok.Id != tokens.Self {
			p.pusherr(tok, "invalid_syntax")
			continue
		}
		selectors[i] = tok
	}
	return selectors
}

func (p *Parser) buildUseCppDecl(use *models.UseDecl, toks []lex.Token) {
	if len(toks) > 2 {
		p.pusherr(toks[2], "invalid_syntax")
	}
	tok := toks[1]
	if tok.Id != tokens.Value || (tok.Kind[0] != '`' && tok.Kind[0] != '"') {
		p.pusherr(tok, "invalid_expr")
		return
	}
	use.Cpp = true
	use.Path = tok.Kind[1 : len(tok.Kind)-1]
}

func (p *Parser) buildUseDecl(use *models.UseDecl, toks []lex.Token) {
	var path strings.Builder
	path.WriteString(jule.StdlibPath)
	path.WriteRune(os.PathSeparator)
	tok := toks[0]
	isStd := false
	if tok.Id == tokens.Cpp {
		p.buildUseCppDecl(use, toks)
		return
	}
	if tok.Id != tokens.Id || tok.Kind != "std" {
		p.pusherr(toks[0], "invalid_syntax")
	}
	isStd = true
	if len(toks) < 3 {
		p.pusherr(tok, "invalid_syntax")
		return
	}
	toks = toks[2:]
	tok = toks[len(toks)-1]
	switch tok.Id {
	case tokens.DoubleColon:
		p.pusherr(tok, "invalid_syntax")
		return
	case tokens.Brace:
		if tok.Kind != tokens.RBRACE {
			p.pusherr(tok, "invalid_syntax")
			return
		}
		var selectors []lex.Token
		toks, selectors = RangeLast(toks)
		use.Selectors = p.getSelectors(selectors)
		if len(toks) == 0 {
			p.pusherr(tok, "invalid_syntax")
			return
		}
		tok = toks[len(toks)-1]
		if tok.Id != tokens.DoubleColon {
			p.pusherr(tok, "invalid_syntax")
			return
		}
		toks = toks[:len(toks)-1]
		if len(toks) == 0 {
			p.pusherr(tok, "invalid_syntax")
			return
		}
	case tokens.Operator:
		if tok.Kind != tokens.STAR {
			p.pusherr(tok, "invalid_syntax")
			return
		}
		toks = toks[:len(toks)-1]
		if len(toks) == 0 {
			p.pusherr(tok, "invalid_syntax")
			return
		}
		tok = toks[len(toks)-1]
		if tok.Id != tokens.DoubleColon {
			p.pusherr(tok, "invalid_syntax")
			return
		}
		toks = toks[:len(toks)-1]
		if len(toks) == 0 {
			p.pusherr(tok, "invalid_syntax")
			return
		}
		use.FullUse = true
	}
	for i, tok := range toks {
		if i%2 != 0 {
			if tok.Id != tokens.DoubleColon {
				p.pusherr(tok, "invalid_syntax")
			}
			path.WriteRune(os.PathSeparator)
			continue
		}
		if tok.Id != tokens.Id {
			p.pusherr(tok, "invalid_syntax")
		}
		path.WriteString(tok.Kind)
	}
	use.LinkString = tokstoa(toks)
	if isStd {
		use.LinkString = "std::" + use.LinkString
	}
	use.Path = path.String()
}

// Attribute builds AST model of attribute.
func (p *Parser) Attribute(toks []lex.Token) (a models.Attribute) {
	i := 0
	a.Token = toks[i]
	i++
	tag := toks[i]
	if tag.Id != tokens.Id || a.Token.Column+1 != tag.Column {
		p.pusherr(tag, "invalid_syntax")
		return
	}
	a.Tag = tag.Kind
	toks = toks[i+1:]
	if len(toks) > 0 {
		tok := toks[0]
		if a.Token.Column+len(a.Tag)+1 == tok.Column {
			p.pusherr(tok, "invalid_syntax")
		}
		p.Tokens = append(toks, p.Tokens...)
	}
	return
}

func (p *Parser) setup_receiver(f *models.Fn, owner_id string) {
	if len(f.Params) == 0 {
		p.pusherr(f.Token, "missing_receiver")
		return
	}
	param := f.Params[0]
	if param.Id != tokens.SELF {
		p.pusherr(f.Token, "missing_receiver")
		return
	}
	f.Receiver = new(models.Var)
	f.Receiver.Type = models.Type{
		Id:   juletype.Struct,
		Kind: owner_id,
	}
	f.Receiver.Mutable = param.Mutable
	if param.Type.Kind != "" && param.Type.Kind[0] == '&' {
		f.Receiver.Type.Kind = tokens.AMPER + f.Receiver.Type.Kind
	}
	f.Params = f.Params[1:]
}

func (p *Parser) funcPrototype(toks []lex.Token, i *int, method, anon bool) (f models.Fn, ok bool) {
	ok = true
	f.Token = toks[*i]
	if f.Token.Id == tokens.Unsafe {
		f.IsUnsafe = true
		*i++
		if *i >= len(toks) {
			p.pusherr(f.Token, "invalid_syntax")
			ok = false
			return
		}
		f.Token = toks[*i]
	}
	// Skips fn tok
	*i++
	if *i >= len(toks) {
		p.pusherr(f.Token, "invalid_syntax")
		ok = false
		return
	}
	f.Pub = p.pub
	p.pub = false
	if anon {
		f.Id = jule.Anonymous
	} else {
		tok := toks[*i]
		if tok.Id != tokens.Id {
			p.pusherr(tok, "invalid_syntax")
			ok = false
		}
		f.Id = tok.Kind
		*i++
	}
	f.RetType.Type.Id = juletype.Void
	f.RetType.Type.Kind = juletype.TypeMap[f.RetType.Type.Id]
	if *i >= len(toks) {
		p.pusherr(f.Token, "invalid_syntax")
		return
	} else if toks[*i].Kind != tokens.LPARENTHESES {
		p.pusherr(toks[*i], "missing_function_parentheses")
		return
	}
	paramToks := p.getrange(i, tokens.LPARENTHESES, tokens.RPARENTHESES, &toks)
	if len(paramToks) > 0 {
		f.Params = p.Params(paramToks, method, false)
	}
	t, retok := p.FuncRetDataType(toks, i)
	if retok {
		f.RetType = t
		*i++
	}
	return
}

// Func builds AST model of function.
func (p *Parser) Func(toks []lex.Token, method, anon, prototype bool) (f models.Fn) {
	var ok bool
	i := 0
	f, ok = p.funcPrototype(toks, &i, method, anon)
	if prototype {
		if i+1 < len(toks) {
			p.pusherr(toks[i+1], "invalid_syntax")
		}
		return
	} else if !ok {
		return
	}
	if i >= len(toks) {
		if p.Ended() {
			p.pusherr(f.Token, "body_not_exist")
			return
		}
		toks = p.nextBuilderStatement()
		i = 0
	}
	blockToks := p.getrange(&i, tokens.LBRACE, tokens.RBRACE, &toks)
	if blockToks != nil {
		f.Block = p.Block(blockToks)
		f.Block.IsUnsafe = f.IsUnsafe
		if i < len(toks) {
			p.pusherr(toks[i], "invalid_syntax")
		}
	} else {
		p.pusherr(f.Token, "body_not_exist")
		p.Tokens = append(toks, p.Tokens...)
	}
	return
}

func (p *Parser) generic(toks []lex.Token) models.GenericType {
	if len(toks) > 1 {
		p.pusherr(toks[1], "invalid_syntax")
	}
	var gt models.GenericType
	gt.Token = toks[0]
	if gt.Token.Id != tokens.Id {
		p.pusherr(gt.Token, "invalid_syntax")
	}
	gt.Id = gt.Token.Kind
	return gt
}

// Generic builds generic type.
func (p *Parser) Generics(toks []lex.Token) []models.GenericType {
	tok := toks[0]
	i := 1
	genericsToks := Range(&i, tokens.LBRACKET, tokens.RBRACKET, toks)
	if len(genericsToks) == 0 {
		p.pusherr(tok, "missing_expr")
		return make([]models.GenericType, 0)
	} else if i < len(toks) {
		p.pusherr(toks[i], "invalid_syntax")
	}
	parts, errs := Parts(genericsToks, tokens.Comma, true)
	p.Errors = append(p.Errors, errs...)
	generics := make([]models.GenericType, len(parts))
	for i, part := range parts {
		if len(parts) == 0 {
			continue
		}
		generics[i] = p.generic(part)
	}
	return generics
}

// TypeOrGenerics builds type alias or generics type declaration.
func (p *Parser) TypeOrGenerics(toks []lex.Token) models.Object {
	if len(toks) > 1 {
		tok := toks[1]
		if tok.Id == tokens.Brace && tok.Kind == tokens.LBRACKET {
			generics := p.Generics(toks)
			return models.Object{
				Token: tok,
				Data:  generics,
			}
		}
	}
	t := p.Type(toks)
	t.Pub = p.pub
	p.pub = false
	return models.Object{
		Token: t.Token,
		Data:  t,
	}
}

// GlobalVar builds AST model of global variable.
func (p *Parser) GlobalVar(toks []lex.Token) {
	if toks == nil {
		return
	}
	bs := blockStatement{toks: toks}
	s := p.VarStatement(&bs, true)
	p.Tree = append(p.Tree, models.Object{
		Token: s.Token,
		Data:  s,
	})
}

func (p *Parser) build_self(toks []lex.Token) (model models.Param) {
	if len(toks) == 0 {
		return
	}
	i := 0
	if toks[i].Id == tokens.Mut {
		model.Mutable = true
		i++
		if i >= len(toks) {
			p.pusherr(toks[i-1], "invalid_syntax")
			return
		}
	}
	if toks[i].Kind == tokens.AMPER {
		model.Type.Kind = "&"
		i++
		if i >= len(toks) {
			p.pusherr(toks[i-1], "invalid_syntax")
			return
		}
	}
	if toks[i].Id == tokens.Self {
		model.Id = tokens.SELF
		model.Token = toks[i]
		i++
		if i < len(toks) {
			p.pusherr(toks[i+1], "invalid_syntax")
		}
	}
	return
}

// Params builds AST model of function parameters.
func (p *Parser) Params(toks []lex.Token, method, mustPure bool) []models.Param {
	parts, errs := Parts(toks, tokens.Comma, true)
	p.Errors = append(p.Errors, errs...)
	if len(parts) == 0 {
		return nil
	}
	var params []models.Param
	if method && len(parts) > 0 {
		param := p.build_self(parts[0])
		if param.Id == tokens.SELF {
			params = append(params, param)
			parts = parts[1:]
		}
	}
	for _, part := range parts {
		p.pushParam(&params, part, mustPure)
	}
	p.checkParams(&params)
	return params
}

func (p *Parser) checkParams(params *[]models.Param) {
	for i := range *params {
		param := &(*params)[i]
		if param.Id == tokens.SELF || param.Type.Token.Id != tokens.NA {
			continue
		}
		if param.Token.Id == tokens.NA {
			p.pusherr(param.Token, "missing_type")
		} else {
			param.Type.Token = param.Token
			param.Type.Id = juletype.Id
			param.Type.Kind = param.Type.Token.Kind
			param.Type.Original = param.Type
			param.Id = jule.Anonymous
			param.Token = lex.Token{}
		}
	}
}

func (p *Parser) paramTypeBegin(param *models.Param, i *int, toks []lex.Token) {
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		switch tok.Id {
		case tokens.Operator:
			switch tok.Kind {
			case tokens.TRIPLE_DOT:
				if param.Variadic {
					p.pusherr(tok, "already_variadic")
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

func (p *Parser) paramBodyId(param *models.Param, tok lex.Token) {
	if juleapi.IsIgnoreId(tok.Kind) {
		param.Id = jule.Anonymous
		return
	}
	param.Id = tok.Kind
}

func (p *Parser) paramBody(param *models.Param, i *int, toks []lex.Token, mustPure bool) {
	p.paramBodyId(param, toks[*i])
	// +1 for skip identifier token
	tok := toks[*i]
	toks = toks[*i+1:]
	if len(toks) == 0 {
		return
	} else if len(toks) < 2 {
		p.pusherr(tok, "missing_type")
		return
	}
	tok = toks[*i]
	if tok.Id != tokens.Colon {
		p.pusherr(tok, "invalid_syntax")
		return
	}
	toks = toks[*i+1:] // Skip colon
	p.paramType(param, toks, mustPure)
}

func (p *Parser) paramType(param *models.Param, toks []lex.Token, mustPure bool) {
	i := 0
	if !mustPure {
		p.paramTypeBegin(param, &i, toks)
		if i >= len(toks) {
			return
		}
	}
	param.Type, _ = p.DataType(toks, &i, false, true)
	i++
	if i < len(toks) {
		p.pusherr(toks[i], "invalid_syntax")
	}
}

func (p *Parser) pushParam(params *[]models.Param, toks []lex.Token, mustPure bool) {
	var param models.Param
	param.Token = toks[0]
	if param.Token.Id == tokens.Mut {
		param.Mutable = true
		if len(toks) == 1 {
			p.pusherr(toks[0], "invalid_syntax")
			return
		}
		toks = toks[1:]
		param.Token = toks[0]
	}
	// Just data type
	if param.Token.Id != tokens.Id {
		param.Id = jule.Anonymous
		p.paramType(&param, toks, mustPure)
	} else {
		i := 0
		p.paramBody(&param, &i, toks, mustPure)
	}
	*params = append(*params, param)
}

func (p *Parser) idGenericsParts(toks []lex.Token, i *int) [][]lex.Token {
	first := *i
	brace_n := 0
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACKET:
				brace_n++
			case tokens.RBRACKET:
				brace_n--
			}
		}
		if brace_n == 0 {
			break
		}
	}
	toks = toks[first+1 : *i]
	parts, errs := Parts(toks, tokens.Comma, true)
	p.Errors = append(p.Errors, errs...)
	return parts
}

func (p *Parser) idDataTypePartEnd(t *models.Type, dtv *strings.Builder, toks []lex.Token, i *int) {
	if *i+1 >= len(toks) {
		return
	}
	*i++
	tok := toks[*i]
	if tok.Id != tokens.Brace || tok.Kind != tokens.LBRACKET {
		*i--
		return
	}
	dtv.WriteByte('[')
	var genericsStr strings.Builder
	parts := p.idGenericsParts(toks, i)
	generics := make([]models.Type, len(parts))
	for i, part := range parts {
		index := 0
		t, _ := p.DataType(part, &index, false, true)
		if index+1 < len(part) {
			p.pusherr(part[index+1], "invalid_syntax")
		}
		genericsStr.WriteString(t.String())
		genericsStr.WriteByte(',')
		generics[i] = t
	}
	dtv.WriteString(genericsStr.String()[:genericsStr.Len()-1])
	dtv.WriteByte(']')
	t.Tag = generics
}

func (p *Parser) datatype(t *models.Type, toks []lex.Token, i *int, arrays, err bool) (ok bool) {
	defer func() { t.Original = *t }()
	first := *i
	var dtv strings.Builder
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		switch tok.Id {
		case tokens.DataType:
			t.Token = tok
			t.Id = juletype.TypeFromId(t.Token.Kind)
			dtv.WriteString(t.Token.Kind)
			ok = true
			goto ret
		case tokens.Id:
			dtv.WriteString(tok.Kind)
			if *i+1 < len(toks) && toks[*i+1].Id == tokens.DoubleColon {
				break
			}
			t.Id = juletype.Id
			t.Token = tok
			p.idDataTypePartEnd(t, &dtv, toks, i)
			ok = true
			goto ret
		case tokens.Cpp:
			if *i+1 >= len(toks) {
				if err {
					p.pusherr(tok, "invalid_syntax")
				}
				return
			}
			*i++
			if toks[*i].Id != tokens.Dot {
				if err {
					p.pusherr(toks[*i], "invalid_syntax")
				}
			}
			if *i+1 >= len(toks) {
				if err {
					p.pusherr(tok, "invalid_syntax")
				}
				return
			}
			*i++
			if toks[*i].Id != tokens.Id {
				if err {
					p.pusherr(toks[*i], "invalid_syntax")
				}
			}
			t.CppLinked = true
			t.Id = juletype.Id
			t.Token = toks[*i]
			dtv.WriteString(t.Token.Kind)
			p.idDataTypePartEnd(t, &dtv, toks, i)
			ok = true
			goto ret
		case tokens.DoubleColon:
			dtv.WriteString(tok.Kind)
		case tokens.Unsafe:
			if *i+1 >= len(toks) || toks[*i+1].Id != tokens.Fn {
				t.Id = juletype.Unsafe
				t.Token = tok
				dtv.WriteString(tok.Kind)
				ok = true
				goto ret
			}
			fallthrough
		case tokens.Fn:
			t.Token = tok
			t.Id = juletype.Fn
			f, proto_ok := p.funcPrototype(toks, i, false, true)
			if !proto_ok {
				p.pusherr(tok, "invalid_type")
				return false
			}
			*i--
			t.Tag = &f
			dtv.WriteString(f.DataTypeString())
			ok = true
			goto ret
		case tokens.Operator:
			switch tok.Kind {
			case tokens.STAR, tokens.AMPER, tokens.DOUBLE_AMPER:
				dtv.WriteString(tok.Kind)
			default:
				if err {
					p.pusherr(tok, "invalid_syntax")
				}
				return
			}
		case tokens.Brace:
			switch tok.Kind {
			case tokens.LBRACKET:
				*i++
				if *i >= len(toks) {
					if err {
						p.pusherr(tok, "invalid_syntax")
					}
					return
				}
				tok = toks[*i]
				if tok.Id == tokens.Brace && tok.Kind == tokens.RBRACKET {
					arrays = false
					dtv.WriteString(jule.Prefix_Slice)
					t.ComponentType = new(models.Type)
					t.Id = juletype.Slice
					t.Token = tok
					*i++
					ok = p.datatype(t.ComponentType, toks, i, arrays, err)
					dtv.WriteString(t.ComponentType.Kind)
					goto ret
				}
				*i-- // Start from bracket
				if arrays {
					ok = p.MapOrArrayDataType(t, toks, i, err)
				} else {
					ok = p.MapDataType(t, toks, i, err)
				}
				if t.Id == juletype.Void {
					if err {
						p.pusherr(tok, "invalid_syntax")
					}
					return
				}
				t.Token = tok
				t.Kind = dtv.String() + t.Kind
				return
			}
			return
		default:
			if err {
				p.pusherr(tok, "invalid_syntax")
			}
			return
		}
	}
	if err {
		p.pusherr(toks[first], "invalid_type")
	}
ret:
	t.Kind = dtv.String()
	return
}

// DataType builds AST model of data-type.
func (p *Parser) DataType(toks []lex.Token, i *int, arrays, err bool) (t models.Type, ok bool) {
	tok := toks[*i]
	ok = p.datatype(&t, toks, i, arrays, err)
	if err && t.Token.Id == tokens.NA {
		p.pusherr(tok, "invalid_type")
	}
	return
}

func (p *Parser) arrayDataType(t *models.Type, toks []lex.Token, i *int, err bool) (ok bool) {
	defer func() { t.Original = *t }()
	if *i+1 >= len(toks) {
		return
	}
	t.Id = juletype.Array
	*i++
	exprI := *i
	t.ComponentType = new(models.Type)
	ok = p.datatype(t.ComponentType, toks, i, true, err)
	if !ok {
		return
	}
	if t.ComponentType.Size.AutoSized {
		p.pusherr(t.ComponentType.Size.Expr.Tokens[0], "invalid_syntax")
		ok = false
	}
	_, exprToks := RangeLast(toks[:exprI])
	exprToks = exprToks[1 : len(exprToks)-1]
	tok := exprToks[0]
	if len(exprToks) == 1 && tok.Id == tokens.Operator && tok.Kind == tokens.TRIPLE_DOT {
		t.Size.AutoSized = true
		t.Size.Expr.Tokens = exprToks
	} else {
		t.Size.Expr = p.Expr(exprToks)
	}
	t.Kind = jule.Prefix_Array + t.ComponentType.Kind
	return
}

func (p *Parser) MapOrArrayDataType(t *models.Type, toks []lex.Token, i *int, err bool) (ok bool) {
	ok = p.MapDataType(t, toks, i, err)
	if !ok {
		ok = p.arrayDataType(t, toks, i, err)
	}
	return
}

// MapDataType builds map data-type.
func (p *Parser) MapDataType(t *models.Type, toks []lex.Token, i *int, err bool) (ok bool) {
	typeToks, colon := SplitColon(toks, i)
	if typeToks == nil || colon == -1 {
		return
	}
	return p.mapDataType(t, toks, typeToks, colon, err)
}

func (p *Parser) mapDataType(t *models.Type, toks, typeToks []lex.Token,
	colon int, err bool) (ok bool) {
	defer func() { t.Original = *t }()
	t.Id = juletype.Map
	t.Token = toks[0]
	colonTok := toks[colon]
	if colon == 0 || colon+1 >= len(typeToks) {
		if err {
			p.pusherr(colonTok, "missing_expr")
		}
		return
	}
	keyTypeToks := typeToks[:colon]
	valueTypeToks := typeToks[colon+1:]
	types := make([]models.Type, 2)
	j := 0
	types[0], _ = p.DataType(keyTypeToks, &j, true, err)
	j = 0
	types[1], _ = p.DataType(valueTypeToks, &j, true, err)
	t.Tag = types
	t.Kind = t.MapKind()
	ok = true
	return
}

func (p *Parser) funcMultiTypeRet(toks []lex.Token, i *int) (t models.RetType, ok bool) {
	tok := toks[*i]
	t.Type.Kind += tok.Kind
	*i++
	if *i >= len(toks) {
		*i--
		t.Type, ok = p.DataType(toks, i, false, false)
		return
	}
	tok = toks[*i]
	*i-- // For point to parenthses - ( -
	rang := Range(i, tokens.LPARENTHESES, tokens.RPARENTHESES, toks)
	params := p.Params(rang, false, true)
	types := make([]models.Type, len(params))
	for i, param := range params {
		types[i] = param.Type
		if param.Id != jule.Anonymous {
			param.Token.Kind = param.Id
		} else {
			param.Token.Kind = juleapi.Ignore
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

// FuncRetDataType builds ret data-type of function.
func (p *Parser) FuncRetDataType(toks []lex.Token, i *int) (t models.RetType, ok bool) {
	t.Type.Id = juletype.Void
	t.Type.Kind = juletype.TypeMap[t.Type.Id]
	if *i >= len(toks) {
		return
	}
	tok := toks[*i]
	switch tok.Id {
	case tokens.Brace:
		switch tok.Kind {
		case tokens.LPARENTHESES:
			return p.funcMultiTypeRet(toks, i)
		case tokens.LBRACE:
			return
		}
	case tokens.Operator:
		if tok.Kind == tokens.EQUAL {
			return
		}
	}
	t.Type, ok = p.DataType(toks, i, false, true)
	return
}

func (p *Parser) pushStatementToBlock(bs *blockStatement) {
	if len(bs.toks) == 0 {
		return
	}
	lastTok := bs.toks[len(bs.toks)-1]
	if lastTok.Id == tokens.SemiColon {
		if len(bs.toks) == 1 {
			return
		}
		bs.toks = bs.toks[:len(bs.toks)-1]
	}
	s := p.Statement(bs)
	if s.Data == nil {
		return
	}
	s.WithTerminator = bs.withTerminator
	bs.block.Tree = append(bs.block.Tree, s)
}

func setToNextStatement(bs *blockStatement) {
	*bs.srcToks = (*bs.srcToks)[bs.pos:]
	bs.pos, bs.withTerminator = NextStatementPos(*bs.srcToks, 0)
	if bs.withTerminator {
		bs.toks = (*bs.srcToks)[:bs.pos-1]
	} else {
		bs.toks = (*bs.srcToks)[:bs.pos]
	}
}

func blockStatementFinished(bs *blockStatement) bool {
	return bs.pos >= len(*bs.srcToks)
}

// Block builds AST model of statements of code block.
func (p *Parser) Block(toks []lex.Token) (block *models.Block) {
	block = new(models.Block)
	var bs blockStatement
	bs.block = block
	bs.srcToks = &toks
	for {
		setToNextStatement(&bs)
		p.pushStatementToBlock(&bs)
	next:
		if len(bs.nextToks) > 0 {
			bs.toks = bs.nextToks
			bs.nextToks = nil
			p.pushStatementToBlock(&bs)
			goto next
		}
		if blockStatementFinished(&bs) {
			break
		}
	}
	return
}

// Statement builds AST model of statement.
func (p *Parser) Statement(bs *blockStatement) (s models.Statement) {
	tok := bs.toks[0]
	if tok.Id == tokens.Id {
		s, ok := p.IdStatement(bs)
		if ok {
			return s
		}
	}
	s, ok := p.AssignStatement(bs.toks)
	if ok {
		return s
	}
	switch tok.Id {
	case tokens.Const, tokens.Let, tokens.Mut:
		return p.VarStatement(bs, true)
	case tokens.Ret:
		return p.RetStatement(bs.toks)
	case tokens.For:
		return p.IterExpr(bs)
	case tokens.Break:
		return p.BreakStatement(bs.toks)
	case tokens.Continue:
		return p.ContinueStatement(bs.toks)
	case tokens.If:
		return p.IfExpr(bs)
	case tokens.Else:
		return p.ElseBlock(bs)
	case tokens.Comment:
		return p.CommentStatement(bs.toks[0])
	case tokens.Defer:
		return p.DeferStatement(bs.toks)
	case tokens.Co:
		return p.ConcurrentCallStatement(bs.toks)
	case tokens.Goto:
		return p.GotoStatement(bs.toks)
	case tokens.Fallthrough:
		return p.Fallthrough(bs.toks)
	case tokens.Type:
		t := p.Type(bs.toks)
		s.Token = t.Token
		s.Data = t
		return
	case tokens.Match:
		return p.MatchCase(bs.toks)
	case tokens.Unsafe:
		if len(bs.toks) == 1 || bs.toks[1].Kind != tokens.LBRACE {
			break
		}
		return p.blockStatement(bs.toks[1:], true)
	case tokens.Brace:
		if tok.Kind == tokens.LBRACE {
			return p.blockStatement(bs.toks, false)
		}
	}
	if IsFuncCall(bs.toks) != nil {
		return p.ExprStatement(bs)
	}
	p.pusherr(tok, "invalid_syntax")
	return
}

func (p *Parser) blockStatement(toks []lex.Token, is_unsafe bool) models.Statement {
	i := 0
	tok := toks[0]
	toks = Range(&i, tokens.LBRACE, tokens.RBRACE, toks)
	if i < len(toks) {
		p.pusherr(toks[i], "invalid_syntax")
	}
	block := p.Block(toks)
	block.IsUnsafe = is_unsafe
	return models.Statement{Token: tok, Data: block}
}

func (p *Parser) assignInfo(toks []lex.Token) (info AssignInfo) {
	info.Ok = true
	brace_n := 0
	for i, tok := range toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				brace_n++
			default:
				brace_n--
			}
		}
		if brace_n > 0 {
			continue
		} else if tok.Id != tokens.Operator {
			continue
		} else if !IsAssignOperator(tok.Kind) {
			continue
		}
		info.Left = toks[:i]
		if info.Left == nil {
			p.pusherr(tok, "invalid_syntax")
			info.Ok = false
		}
		info.Setter = tok
		if i+1 >= len(toks) {
			info.Right = nil
			info.Ok = IsPostfixOperator(info.Setter.Kind)
			break
		}
		info.Right = toks[i+1:]
		if IsPostfixOperator(info.Setter.Kind) {
			if info.Right != nil {
				p.pusherr(info.Right[0], "invalid_syntax")
				info.Right = nil
			}
		}
		break
	}
	return
}

func (p *Parser) buildAssignLeft(toks []lex.Token) (left models.AssignLeft) {
	left.Expr.Tokens = toks
	if left.Expr.Tokens[0].Id == tokens.Id {
		left.Var.Token = left.Expr.Tokens[0]
		left.Var.Id = left.Var.Token.Kind
	}
	left.Expr = p.Expr(left.Expr.Tokens)
	return
}

func (p *Parser) assignLefts(parts [][]lex.Token) []models.AssignLeft {
	var lefts []models.AssignLeft
	for _, part := range parts {
		left := p.buildAssignLeft(part)
		lefts = append(lefts, left)
	}
	return lefts
}

func (p *Parser) assignExprs(toks []lex.Token) []models.Expr {
	parts, errs := Parts(toks, tokens.Comma, true)
	if len(errs) > 0 {
		p.Errors = append(p.Errors, errs...)
		return nil
	}
	exprs := make([]models.Expr, len(parts))
	for i, part := range parts {
		exprs[i] = p.Expr(part)
	}
	return exprs
}

// AssignStatement builds AST model of assignment statement.
func (p *Parser) AssignStatement(toks []lex.Token) (s models.Statement, _ bool) {
	assign, ok := p.AssignExpr(toks)
	if !ok {
		return
	}
	s.Token = toks[0]
	s.Data = assign
	return s, true
}

// AssignExpr builds AST model of assignment expression.
func (p *Parser) AssignExpr(toks []lex.Token) (assign models.Assign, ok bool) {
	if !CheckAssignTokens(toks) {
		return
	}
	switch toks[0].Id {
	case tokens.Let:
		return p.letDeclAssign(toks)
	default:
		return p.plainAssign(toks)
	}
}

func (p *Parser) letDeclAssign(toks []lex.Token) (assign models.Assign, ok bool) {
	if len(toks) < 1 {
		return
	}
	// Skip "let" keyword
	toks = toks[1:]
	tok := toks[0]
	if tok.Id != tokens.Brace || tok.Kind != tokens.LPARENTHESES {
		return
	}
	ok = true
	var i int
	rang := Range(&i, tokens.LPARENTHESES, tokens.RPARENTHESES, toks)
	if rang == nil {
		p.pusherr(tok, "invalid_syntax")
		return
	} else if i+1 < len(toks) {
		assign.Setter = toks[i]
		i++
		assign.Right = p.assignExprs(toks[i:])
	}
	parts, errs := Parts(rang, tokens.Comma, true)
	if len(errs) > 0 {
		p.Errors = append(p.Errors, errs...)
		return
	}
	for _, part := range parts {
		if len(part) > 2 {
			p.pusherr(part[2], "invalid_syntax")
		}
		mutable := false
		tok := part[0]
		if tok.Id == tokens.Mut {
			mutable = true
			part = part[1:]
			if len(part) == 0 {
				p.pusherr(tok, "invalid_syntax")
				continue
			}
		}
		left := p.buildAssignLeft(part)
		left.Var.Mutable = mutable
		left.Var.New = !juleapi.IsIgnoreId(left.Var.Id)
		left.Var.SetterTok = assign.Setter
		assign.Left = append(assign.Left, left)
	}
	return
}

func (p *Parser) plainAssign(toks []lex.Token) (assign models.Assign, ok bool) {
	info := p.assignInfo(toks)
	if !info.Ok {
		return
	}
	ok = true
	assign.Setter = info.Setter
	parts, errs := Parts(info.Left, tokens.Comma, true)
	if len(errs) > 0 {
		p.Errors = append(p.Errors, errs...)
		return
	}
	assign.Left = p.assignLefts(parts)
	if info.Right != nil {
		assign.Right = p.assignExprs(info.Right)
	}
	return
}

// BuildReturnStatement builds AST model of return statement.
func (p *Parser) IdStatement(bs *blockStatement) (s models.Statement, ok bool) {
	if len(bs.toks) == 1 {
		return
	}
	tok := bs.toks[1]
	switch tok.Id {
	case tokens.Colon:
		return p.LabelStatement(bs), true
	}
	return
}

// LabelStatement builds AST model of label.
func (p *Parser) LabelStatement(bs *blockStatement) models.Statement {
	var l models.Label
	l.Token = bs.toks[0]
	l.Label = l.Token.Kind
	if len(bs.toks) > 2 {
		bs.nextToks = bs.toks[2:]
	}
	return models.Statement{Token: l.Token, Data: l}
}

// ExprStatement builds AST model of expression.
func (p *Parser) ExprStatement(bs *blockStatement) models.Statement {
	expr := models.ExprStatement{
		Expr: p.Expr(bs.toks),
	}
	return models.Statement{
		Token: bs.toks[0],
		Data:  expr,
	}
}

// Args builds AST model of arguments.
func (p *Parser) Args(toks []lex.Token, targeting bool) *models.Args {
	args := new(models.Args)
	last := 0
	brace_n := 0
	for i, tok := range toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				brace_n++
			default:
				brace_n--
			}
		}
		if brace_n > 0 || tok.Id != tokens.Comma {
			continue
		}
		p.pushArg(args, targeting, toks[last:i], tok)
		last = i + 1
	}
	if last < len(toks) {
		if last == 0 {
			if len(toks) > 0 {
				p.pushArg(args, targeting, toks[last:], toks[last])
			}
		} else {
			p.pushArg(args, targeting, toks[last:], toks[last-1])
		}
	}
	return args
}

func (p *Parser) pushArg(args *models.Args, targeting bool, toks []lex.Token, err lex.Token) {
	if len(toks) == 0 {
		p.pusherr(err, "invalid_syntax")
		return
	}
	var arg models.Arg
	arg.Token = toks[0]
	if targeting && arg.Token.Id == tokens.Id {
		if len(toks) > 1 {
			tok := toks[1]
			if tok.Id == tokens.Colon {
				args.Targeted = true
				arg.TargetId = arg.Token.Kind
				toks = toks[2:]
			}
		}
	}
	arg.Expr = p.Expr(toks)
	args.Src = append(args.Src, arg)
}

func (p *Parser) varBegin(v *models.Var, i *int, toks []lex.Token) {
	tok := toks[*i]
	switch tok.Id {
	case tokens.Let:
		// Initialize 1 for skip the let keyword
		*i++
		if toks[*i].Id == tokens.Mut {
			v.Mutable = true
			// Skip the mut keyword
			*i++
		}
	case tokens.Const:
		*i++
		if v.Const {
			p.pusherr(tok, "already_const")
			break
		}
		v.Const = true
		if !v.Mutable {
			break
		}
		fallthrough
	default:
		p.pusherr(tok, "invalid_syntax")
		return
	}
	if *i >= len(toks) {
		p.pusherr(tok, "invalid_syntax")
	}
}

func (p *Parser) varTypeNExpr(v *models.Var, toks []lex.Token, i int, expr bool) {
	tok := toks[i]
	if tok.Id == tokens.Colon {
		i++ // Skip type annotation operator (:)
		if i >= len(toks) ||
			(toks[i].Id == tokens.Operator && toks[i].Kind == tokens.EQUAL) {
			p.pusherr(tok, "missing_type")
			return
		}
		t, ok := p.DataType(toks, &i, true, false)
		if ok {
			v.Type = t
			i++
			if i >= len(toks) {
				return
			}
			tok = toks[i]
		}
	}
	if expr && tok.Id == tokens.Operator {
		if tok.Kind != tokens.EQUAL {
			p.pusherr(tok, "invalid_syntax")
			return
		}
		valueToks := toks[i+1:]
		if len(valueToks) == 0 {
			p.pusherr(tok, "missing_expr")
			return
		}
		v.Expr = p.Expr(valueToks)
		v.SetterTok = tok
	} else {
		p.pusherr(tok, "invalid_syntax")
	}
}

// Var builds AST model of variable statement.
func (p *Parser) Var(toks []lex.Token, begin, expr bool) (v models.Var) {
	v.Pub = p.pub
	p.pub = false
	i := 0
	v.Token = toks[i]
	if begin {
		p.varBegin(&v, &i, toks)
		if i >= len(toks) {
			return
		}
	}
	v.Token = toks[i]
	if v.Token.Id != tokens.Id {
		p.pusherr(v.Token, "invalid_syntax")
		return
	}
	v.Id = v.Token.Kind
	v.Type.Id = juletype.Void
	v.Type.Kind = juletype.TypeMap[v.Type.Id]
	if i >= len(toks) {
		return
	}
	i++
	if i < len(toks) {
		p.varTypeNExpr(&v, toks, i, expr)
	} else if !expr {
		p.pusherr(v.Token, "missing_type")
	}
	return
}

// VarStatement builds AST model of variable declaration statement.
func (p *Parser) VarStatement(bs *blockStatement, expr bool) models.Statement {
	v := p.Var(bs.toks, true, expr)
	v.Owner = bs.block
	return models.Statement{Token: v.Token, Data: v}
}

// CommentStatement builds AST model of comment statement.
func (p *Parser) CommentStatement(tok lex.Token) (s models.Statement) {
	s.Token = tok
	tok.Kind = strings.TrimSpace(tok.Kind[2:])
	s.Data = models.Comment{Content: tok.Kind}
	return
}

// DeferStatement builds AST model of deferred call statement.
func (p *Parser) DeferStatement(toks []lex.Token) (s models.Statement) {
	var d models.Defer
	d.Token = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		p.pusherr(d.Token, "missing_expr")
		return
	}
	if IsFuncCall(toks) == nil {
		p.pusherr(d.Token, "expr_not_func_call")
	}
	d.Expr = p.Expr(toks)
	s.Token = d.Token
	s.Data = d
	return
}

func (p *Parser) ConcurrentCallStatement(toks []lex.Token) (s models.Statement) {
	var cc models.ConcurrentCall
	cc.Token = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		p.pusherr(cc.Token, "missing_expr")
		return
	}
	if IsFuncCall(toks) == nil {
		p.pusherr(cc.Token, "expr_not_func_call")
	}
	cc.Expr = p.Expr(toks)
	s.Token = cc.Token
	s.Data = cc
	return
}

func (p *Parser) Fallthrough(toks []lex.Token) (s models.Statement) {
	s.Token = toks[0]
	if len(toks) > 1 {
		p.pusherr(toks[1], "invalid_syntax")
	}
	s.Data = models.Fallthrough{
		Token: s.Token,
	}
	return
}

func (p *Parser) GotoStatement(toks []lex.Token) (s models.Statement) {
	s.Token = toks[0]
	if len(toks) == 1 {
		p.pusherr(s.Token, "missing_goto_label")
		return
	} else if len(toks) > 2 {
		p.pusherr(toks[2], "invalid_syntax")
	}
	idTok := toks[1]
	if idTok.Id != tokens.Id {
		p.pusherr(idTok, "invalid_syntax")
		return
	}
	var gt models.Goto
	gt.Token = s.Token
	gt.Label = idTok.Kind
	s.Data = gt
	return
}

// RetStatement builds AST model of return statement.
func (p *Parser) RetStatement(toks []lex.Token) models.Statement {
	var ret models.Ret
	ret.Token = toks[0]
	if len(toks) > 1 {
		ret.Expr = p.Expr(toks[1:])
	}
	return models.Statement{
		Token: ret.Token,
		Data:  ret,
	}
}

func (p *Parser) getWhileIterProfile(toks []lex.Token) models.IterWhile {
	return models.IterWhile{
		Expr: p.Expr(toks),
	}
}

func (p *Parser) getForeachVarsToks(toks []lex.Token) [][]lex.Token {
	vars, errs := Parts(toks, tokens.Comma, true)
	p.Errors = append(p.Errors, errs...)
	return vars
}

func (p *Parser) getVarProfile(toks []lex.Token) (v models.Var) {
	if len(toks) == 0 {
		return
	}
	v.Token = toks[0]
	if v.Token.Id == tokens.Mut {
		v.Mutable = true
		if len(toks) == 1 {
			p.pusherr(v.Token, "invalid_syntax")
		}
		v.Token = toks[1]
	} else if len(toks) > 1 {
		p.pusherr(toks[1], "invalid_syntax")
	}
	if v.Token.Id != tokens.Id {
		p.pusherr(v.Token, "invalid_syntax")
		return
	}
	v.Id = v.Token.Kind
	v.New = true
	return
}

func (p *Parser) getForeachIterVars(varsToks [][]lex.Token) []models.Var {
	var vars []models.Var
	for _, toks := range varsToks {
		vars = append(vars, p.getVarProfile(toks))
	}
	return vars
}

func (p *Parser) setup_foreach_explicit_vars(f *models.IterForeach, toks []lex.Token) {
	i := 0
	rang := Range(&i, tokens.LPARENTHESES, tokens.RPARENTHESES, toks)
	if i < len(toks) {
		p.pusherr(f.InToken, "invalid_syntax")
	}
	p.setup_foreach_plain_vars(f, rang)
}

func (p *Parser) setup_foreach_plain_vars(f *models.IterForeach, toks []lex.Token) {
	varsToks := p.getForeachVarsToks(toks)
	if len(varsToks) == 0 {
		return
	}
	if len(varsToks) > 2 {
		p.pusherr(f.InToken, "much_foreach_vars")
	}
	vars := p.getForeachIterVars(varsToks)
	f.KeyA = vars[0]
	if len(vars) > 1 {
		f.KeyB = vars[1]
	} else {
		f.KeyB.Id = juleapi.Ignore
	}
}

func (p *Parser) setup_foreach_vars(f *models.IterForeach, toks []lex.Token) {
	if toks[0].Id == tokens.Brace {
		if toks[0].Kind != tokens.LPARENTHESES {
			p.pusherr(toks[0], "invalid_syntax")
			return
		}
		p.setup_foreach_explicit_vars(f, toks)
		return
	}
	p.setup_foreach_plain_vars(f, toks)
}

func (p *Parser) getForeachIterProfile(varToks, exprToks []lex.Token, inTok lex.Token) models.IterForeach {
	var foreach models.IterForeach
	foreach.InToken = inTok
	if len(exprToks) == 0 {
		p.pusherr(inTok, "missing_expr")
		return foreach
	}
	foreach.Expr = p.Expr(exprToks)
	if len(varToks) == 0 {
		foreach.KeyA.Id = juleapi.Ignore
		foreach.KeyB.Id = juleapi.Ignore
	} else {
		p.setup_foreach_vars(&foreach, varToks)
	}
	return foreach
}

func (p *Parser) getIterProfile(toks []lex.Token, errtok lex.Token) models.IterProfile {
	brace_n := 0
	for i, tok := range toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
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
		case tokens.In:
			varToks := toks[:i]
			exprToks := toks[i+1:]
			return p.getForeachIterProfile(varToks, exprToks, tok)
		}
	}
	return p.getWhileIterProfile(toks)
}

func (p *Parser) forStatement(toks []lex.Token) models.Statement {
	s := p.Statement(&blockStatement{toks: toks})
	switch s.Data.(type) {
	case models.ExprStatement, models.Assign, models.Var:
	default:
		p.pusherr(toks[0], "invalid_syntax")
	}
	return s
}

func (p *Parser) forIterProfile(bs *blockStatement) (s models.Statement) {
	var iter models.Iter
	iter.Token = bs.toks[0]
	bs.toks = bs.toks[1:]
	var profile models.IterFor
	if len(bs.toks) > 0 {
		profile.Once = p.forStatement(bs.toks)
	}
	if blockStatementFinished(bs) {
		p.pusherr(iter.Token, "invalid_syntax")
		return
	}
	setToNextStatement(bs)
	if len(bs.toks) > 0 {
		profile.Condition = p.Expr(bs.toks)
	}
	if blockStatementFinished(bs) {
		p.pusherr(iter.Token, "invalid_syntax")
		return
	}
	setToNextStatement(bs)
	exprToks := BlockExpr(bs.toks)
	if len(exprToks) > 0 {
		profile.Next = p.forStatement(exprToks)
	}
	i := len(exprToks)
	blockToks := p.getrange(&i, tokens.LBRACE, tokens.RBRACE, &bs.toks)
	if blockToks == nil {
		p.pusherr(iter.Token, "body_not_exist")
		return
	}
	if i < len(bs.toks) {
		p.pusherr(bs.toks[i], "invalid_syntax")
	}
	iter.Block = p.Block(blockToks)
	iter.Profile = profile
	return models.Statement{Token: iter.Token, Data: iter}
}

func (p *Parser) commonIterProfile(toks []lex.Token) (s models.Statement) {
	var iter models.Iter
	iter.Token = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		p.pusherr(iter.Token, "body_not_exist")
		return
	}
	exprToks := BlockExpr(toks)
	if len(exprToks) > 0 {
		iter.Profile = p.getIterProfile(exprToks, iter.Token)
	}
	i := len(exprToks)
	blockToks := p.getrange(&i, tokens.LBRACE, tokens.RBRACE, &toks)
	if blockToks == nil {
		p.pusherr(iter.Token, "body_not_exist")
		return
	}
	if i < len(toks) {
		p.pusherr(toks[i], "invalid_syntax")
	}
	iter.Block = p.Block(blockToks)
	return models.Statement{Token: iter.Token, Data: iter}
}

func (p *Parser) IterExpr(bs *blockStatement) models.Statement {
	if bs.withTerminator {
		return p.forIterProfile(bs)
	}
	return p.commonIterProfile(bs.toks)
}

func (p *Parser) caseexprs(toks *[]lex.Token, caseIsDefault bool) []models.Expr {
	var exprs []models.Expr
	pushExpr := func(toks []lex.Token, tok lex.Token) {
		if caseIsDefault {
			if len(toks) > 0 {
				p.pusherr(tok, "invalid_syntax")
			}
			return
		}
		if len(toks) > 0 {
			exprs = append(exprs, p.Expr(toks))
			return
		}
		p.pusherr(tok, "missing_expr")
	}
	brace_n := 0
	j := 0
	var i int
	var tok lex.Token
	for i, tok = range *toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LPARENTHESES, tokens.LBRACE, tokens.LBRACKET:
				brace_n++
			default:
				brace_n--
			}
			continue
		} else if brace_n != 0 {
			continue
		}
		switch tok.Id {
		case tokens.Comma:
			pushExpr((*toks)[j:i], tok)
			j = i + 1
		case tokens.Colon:
			pushExpr((*toks)[j:i], tok)
			*toks = (*toks)[i+1:]
			return exprs
		}
	}
	p.pusherr((*toks)[0], "invalid_syntax")
	*toks = nil
	return nil
}

func (p *Parser) caseblock(toks *[]lex.Token) *models.Block {
	brace_n := 0
	for i, tok := range *toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LPARENTHESES, tokens.LBRACE, tokens.LBRACKET:
				brace_n++
			default:
				brace_n--
			}
			continue
		} else if brace_n != 0 {
			continue
		}
		switch tok.Id {
		case tokens.Case, tokens.Default:
			blockToks := (*toks)[:i]
			*toks = (*toks)[i:]
			return p.Block(blockToks)
		}
	}
	block := p.Block(*toks)
	*toks = nil
	return block
}

func (p *Parser) getcase(toks *[]lex.Token) models.Case {
	var c models.Case
	c.Token = (*toks)[0]
	*toks = (*toks)[1:]
	c.Exprs = p.caseexprs(toks, c.Token.Id == tokens.Default)
	c.Block = p.caseblock(toks)
	return c
}

func (p *Parser) cases(toks []lex.Token) ([]models.Case, *models.Case) {
	var cases []models.Case
	var def *models.Case
	for len(toks) > 0 {
		tok := toks[0]
		switch tok.Id {
		case tokens.Case:
			cases = append(cases, p.getcase(&toks))
		case tokens.Default:
			c := p.getcase(&toks)
			c.Token = tok
			if def == nil {
				def = new(models.Case)
				*def = c
				break
			}
			fallthrough
		default:
			p.pusherr(tok, "invalid_syntax")
		}
	}
	return cases, def
}

// MatchCase builds AST model of match-case.
func (p *Parser) MatchCase(toks []lex.Token) (s models.Statement) {
	match := new(models.Match)
	match.Token = toks[0]
	s.Token = match.Token
	toks = toks[1:]
	exprToks := BlockExpr(toks)
	if len(exprToks) > 0 {
		match.Expr = p.Expr(exprToks)
	}
	i := len(exprToks)
	blockToks := p.getrange(&i, tokens.LBRACE, tokens.RBRACE, &toks)
	if blockToks == nil {
		p.pusherr(match.Token, "body_not_exist")
		return
	}
	match.Cases, match.Default = p.cases(blockToks)
	for i := range match.Cases {
		c := &match.Cases[i]
		c.Match = match
		if i > 0 {
			match.Cases[i-1].Next = c
		}
	}
	if match.Default != nil {
		if len(match.Cases) > 0 {
			match.Cases[len(match.Cases)-1].Next = match.Default
		}
		match.Default.Match = match
	}
	s.Data = match
	return
}

// IfExpr builds AST model of if expression.
func (p *Parser) IfExpr(bs *blockStatement) (s models.Statement) {
	var ifast models.If
	ifast.Token = bs.toks[0]
	bs.toks = bs.toks[1:]
	exprToks := BlockExpr(bs.toks)
	i := 0
	if len(exprToks) == 0 {
		if len(bs.toks) == 0 || bs.pos >= len(*bs.srcToks) {
			p.pusherr(ifast.Token, "missing_expr")
			return
		}
		exprToks = bs.toks
		setToNextStatement(bs)
	} else {
		i = len(exprToks)
	}
	blockToks := p.getrange(&i, tokens.LBRACE, tokens.RBRACE, &bs.toks)
	if blockToks == nil {
		p.pusherr(ifast.Token, "body_not_exist")
		return
	}
	if i < len(bs.toks) {
		if bs.toks[i].Id == tokens.Else {
			bs.nextToks = bs.toks[i:]
		} else {
			p.pusherr(bs.toks[i], "invalid_syntax")
		}
	}
	ifast.Expr = p.Expr(exprToks)
	ifast.Block = p.Block(blockToks)
	return models.Statement{Token: ifast.Token, Data: ifast}
}

// ElseIfEpxr builds AST model of else if expression.
func (p *Parser) ElseIfExpr(bs *blockStatement) (s models.Statement) {
	var elif models.ElseIf
	elif.Token = bs.toks[1]
	bs.toks = bs.toks[2:]
	exprToks := BlockExpr(bs.toks)
	i := 0
	if len(exprToks) == 0 {
		if len(bs.toks) == 0 || bs.pos >= len(*bs.srcToks) {
			p.pusherr(elif.Token, "missing_expr")
			return
		}
		exprToks = bs.toks
		setToNextStatement(bs)
	} else {
		i = len(exprToks)
	}
	blockToks := p.getrange(&i, tokens.LBRACE, tokens.RBRACE, &bs.toks)
	if blockToks == nil {
		p.pusherr(elif.Token, "body_not_exist")
		return
	}
	if i < len(bs.toks) {
		if bs.toks[i].Id == tokens.Else {
			bs.nextToks = bs.toks[i:]
		} else {
			p.pusherr(bs.toks[i], "invalid_syntax")
		}
	}
	elif.Expr = p.Expr(exprToks)
	elif.Block = p.Block(blockToks)
	return models.Statement{Token: elif.Token, Data: elif}
}

// ElseBlock builds AST model of else block.
func (p *Parser) ElseBlock(bs *blockStatement) (s models.Statement) {
	if len(bs.toks) > 1 && bs.toks[1].Id == tokens.If {
		return p.ElseIfExpr(bs)
	}
	var elseast models.Else
	elseast.Token = bs.toks[0]
	bs.toks = bs.toks[1:]
	i := 0
	blockToks := p.getrange(&i, tokens.LBRACE, tokens.RBRACE, &bs.toks)
	if blockToks == nil {
		if i < len(bs.toks) {
			p.pusherr(elseast.Token, "else_have_expr")
		} else {
			p.pusherr(elseast.Token, "body_not_exist")
		}
		return
	}
	if i < len(bs.toks) {
		p.pusherr(bs.toks[i], "invalid_syntax")
	}
	elseast.Block = p.Block(blockToks)
	return models.Statement{Token: elseast.Token, Data: elseast}
}

// BreakStatement builds AST model of break statement.
func (p *Parser) BreakStatement(toks []lex.Token) models.Statement {
	var breakAST models.Break
	breakAST.Token = toks[0]
	if len(toks) > 1 {
		if toks[1].Id != tokens.Id {
			p.pusherr(toks[1], "invalid_syntax")
		} else {
			breakAST.LabelToken = toks[1]
			if len(toks) > 2 {
				p.pusherr(toks[1], "invalid_syntax")
			}
		}
	}
	return models.Statement{
		Token: breakAST.Token,
		Data:  breakAST,
	}
}

// ContinueStatement builds AST model of continue statement.
func (p *Parser) ContinueStatement(toks []lex.Token) models.Statement {
	var continueAST models.Continue
	continueAST.Token = toks[0]
	if len(toks) > 1 {
		if toks[1].Id != tokens.Id {
			p.pusherr(toks[1], "invalid_syntax")
		} else {
			continueAST.LoopLabel = toks[1]
			if len(toks) > 2 {
				p.pusherr(toks[1], "invalid_syntax")
			}
		}
	}
	return models.Statement{Token: continueAST.Token, Data:  continueAST}
}

// Expr builds AST model of expression.
func (p *Parser) Expr(toks []lex.Token) (e models.Expr) {
	e.Op = p.build_expr_op(toks)
	e.Tokens = toks
	return
}

func (p *Parser) build_binop_expr(toks []lex.Token) any {
	i := p.find_lowest_precedenced_operator(toks)
	if i != -1 {
		return p.build_binop(toks)
	}
	return models.BinopExpr{Tokens: toks}
}

func (p *Parser) build_binop(toks []lex.Token) models.Binop {
	op := models.Binop{}
	i := p.find_lowest_precedenced_operator(toks)
	op.L = p.build_binop_expr(toks[:i])
	op.R = p.build_binop_expr(toks[i+1:])
	op.Op = toks[i]
	return op
}

func eliminate_comments(toks []lex.Token) []lex.Token {
	cutted := []lex.Token{}
	for _, token := range toks {
		if token.Id != tokens.Comment {
			cutted = append(cutted, token)
		}
	}
	return cutted
}

// Returns BinopExpr or Binop instance for expression Op.
func (p *Parser) build_expr_op(toks []lex.Token) any {
	toks = eliminate_comments(toks)
	i := p.find_lowest_precedenced_operator(toks)
	if i == -1 {
		return p.build_binop_expr(toks)
	}
	return p.build_binop(toks)
}

// nextOperator find index of priority operator and returns index of operator
// if found, returns -1 if not.
func (p *Parser) find_lowest_precedenced_operator(toks []lex.Token) int {
	prec := precedencer{}
	brace_n := 0
	for i, token := range toks {
		switch {
		case token.Id == tokens.Brace:
			switch token.Kind {
			case tokens.LBRACE, tokens.LPARENTHESES, tokens.LBRACKET:
				brace_n++
			default:
				brace_n--
			}
			continue
		case i == 0:
			continue
		case token.Id != tokens.Operator:
			continue
		case brace_n > 0:
			continue
		}
		// Skip unary operator.
		if toks[i-1].Id == tokens.Operator {
			continue
		}
		switch token.Kind {
		case tokens.STAR, tokens.PERCENT, tokens.SOLIDUS,
			tokens.RSHIFT, tokens.LSHIFT, tokens.AMPER:
			prec.set(5, i)
		case tokens.PLUS, tokens.MINUS, tokens.VLINE, tokens.CARET:
			prec.set(4, i)
		case tokens.EQUALS, tokens.NOT_EQUALS, tokens.LESS,
			tokens.LESS_EQUAL, tokens.GREAT, tokens.GREAT_EQUAL:
			prec.set(3, i)
		case tokens.DOUBLE_AMPER:
			prec.set(2, i)
		case tokens.DOUBLE_VLINE:
			prec.set(1, i)
		}
	}
	data := prec.get_lower()
	if data == nil {
		return -1
	}
	return data.(int)
}

func (p *Parser) getrange(i *int, open, close string, toks *[]lex.Token) []lex.Token {
	rang := Range(i, open, close, *toks)
	if rang != nil {
		return rang
	}
	if p.Ended() {
		return nil
	}
	*i = 0
	*toks = p.nextBuilderStatement()
	rang = Range(i, open, close, *toks)
	return rang
}

func (p *Parser) skipStatement(i *int, toks *[]lex.Token) []lex.Token {
	start := *i
	*i, _ = NextStatementPos(*toks, start)
	stoks := (*toks)[start:*i]
	if stoks[len(stoks)-1].Id == tokens.SemiColon {
		if len(stoks) == 1 {
			return p.skipStatement(i, toks)
		}
		stoks = stoks[:len(stoks)-1]
	}
	return stoks
}

func (p *Parser) nextBuilderStatement() []lex.Token {
	return p.skipStatement(&p.Pos, &p.Tokens)
}
