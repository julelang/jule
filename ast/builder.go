package ast

import (
	"os"
	"strings"
	"sync"

	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/lex"
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xbits"
	"github.com/the-xlang/xxc/pkg/xlog"
	"github.com/the-xlang/xxc/pkg/xtype"
)

type Tok = lex.Tok
type Toks = []Tok

// Builder is builds AST tree.
type Builder struct {
	wg  sync.WaitGroup
	pub bool

	Tree   []models.Object
	Errors []xlog.CompilerLog
	Toks   Toks
	Pos    int
}

// NewBuilder instance.
func NewBuilder(toks Toks) *Builder {
	b := new(Builder)
	b.Toks = toks
	b.Pos = 0
	return b
}

func compilerErr(tok Tok, key string, args ...any) xlog.CompilerLog {
	return xlog.CompilerLog{
		Type:    xlog.Error,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    tok.File.Path(),
		Message: x.GetError(key, args...),
	}
}

// pusherr appends error by specified token.
func (b *Builder) pusherr(tok Tok, key string, args ...any) {
	b.Errors = append(b.Errors, compilerErr(tok, key, args...))
}

// Ended reports position is at end of tokens or not.
func (ast *Builder) Ended() bool {
	return ast.Pos >= len(ast.Toks)
}

func (b *Builder) buildNode(toks Toks) {
	tok := toks[0]
	if tok.Id == tokens.Id || tok.Id == tokens.Operator {
		if b.GlobalFunc(toks) {
			return
		}
	}
	switch tok.Id {
	case tokens.Use:
		b.Use(toks)
	case tokens.At:
		b.Attribute(toks)
	case tokens.Id:
		b.Id(toks)
	case tokens.Const, tokens.Volatile:
		b.GlobalVar(toks)
	case tokens.Type:
		b.Tree = append(b.Tree, b.TypeOrGenerics(toks))
	case tokens.Enum:
		b.Enum(toks)
	case tokens.Struct:
		b.Struct(toks)
	case tokens.Trait:
		b.Trait(toks)
	case tokens.Impl:
		b.Impl(toks)
	case tokens.Comment:
		b.Comment(toks[0])
	case tokens.Preprocessor:
		b.Preprocessor(toks)
	default:
		b.pusherr(tok, "invalid_syntax")
		return
	}
	if b.pub {
		b.pusherr(tok, "def_not_support_pub")
	}
}

// Build builds AST tree.
func (b *Builder) Build() {
	for b.Pos != -1 && !b.Ended() {
		toks := b.nextBuilderStatement()
		b.pub = toks[0].Id == tokens.Pub
		if b.pub {
			if len(toks) == 1 {
				if b.Ended() {
					b.pusherr(toks[0], "invalid_syntax")
					continue
				}
				toks = b.nextBuilderStatement()
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

// Type builds AST model of type definition statement.
func (b *Builder) Type(toks Toks) (t models.Type) {
	i := 1 // Initialize value is 1 for skip keyword.
	if i >= len(toks) {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	t.Tok = toks[1]
	t.Id = t.Tok.Kind
	tok := toks[i]
	if tok.Id != tokens.Id {
		b.pusherr(tok, "invalid_syntax")
	}
	i++
	if i >= len(toks) {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	destType, ok := b.DataType(toks, &i, true, true)
	t.Type = destType
	if ok && i+1 < len(toks) {
		b.pusherr(toks[i+1], "invalid_syntax")
	}
	return
}

func (b *Builder) buildEnumItemExpr(i *int, toks Toks) models.Expr {
	braceCount := 0
	exprStart := *i
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
				continue
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		}
		if tok.Id == tokens.Comma || *i+1 >= len(toks) {
			var exprToks Toks
			if tok.Id == tokens.Comma {
				exprToks = toks[exprStart:*i]
			} else {
				exprToks = toks[exprStart:]
			}
			return b.Expr(exprToks)
		}
	}
	return models.Expr{}
}

func (b *Builder) buildEnumItems(toks Toks) []*models.EnumItem {
	items := make([]*models.EnumItem, 0)
	for i := 0; i < len(toks); i++ {
		tok := toks[i]
		item := new(models.EnumItem)
		item.Tok = tok
		if item.Tok.Id != tokens.Id {
			b.pusherr(item.Tok, "invalid_syntax")
		}
		item.Id = item.Tok.Kind
		if i+1 >= len(toks) || toks[i+1].Id == tokens.Comma {
			if i+1 < len(toks) {
				i++
			}
			items = append(items, item)
			continue
		}
		i++
		tok = toks[i]
		if tok.Id != tokens.Operator && tok.Kind != tokens.EQUAL {
			b.pusherr(toks[0], "invalid_syntax")
		}
		i++
		if i >= len(toks) || toks[i].Id == tokens.Comma {
			b.pusherr(toks[0], "missing_expr")
			continue
		}
		item.Expr = b.buildEnumItemExpr(&i, toks)
		items = append(items, item)
	}
	return items
}

// Enum builds AST model of enumerator statement.
func (b *Builder) Enum(toks Toks) {
	var enum models.Enum
	if len(toks) < 2 || len(toks) < 3 {
		b.pusherr(toks[0], "invalid_syntax")
		return
	}
	enum.Tok = toks[1]
	if enum.Tok.Id != tokens.Id {
		b.pusherr(enum.Tok, "invalid_syntax")
	}
	enum.Id = enum.Tok.Kind
	i := 2
	if toks[i].Id == tokens.Colon {
		i++
		if i >= len(toks) {
			b.pusherr(toks[i-1], "invalid_syntax")
			return
		}
		enum.Type, _ = b.DataType(toks, &i, false, true)
		i++
		if i >= len(toks) {
			b.pusherr(enum.Tok, "body_not_exist")
			return
		}
	} else {
		enum.Type = models.DataType{Id: xtype.U32, Kind: tokens.U32}
	}
	itemToks := b.getrange(&i, tokens.LBRACE, tokens.RBRACE, &toks)
	if itemToks == nil {
		b.pusherr(enum.Tok, "body_not_exist")
		return
	} else if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	enum.Pub = b.pub
	b.pub = false
	enum.Items = b.buildEnumItems(itemToks)
	b.Tree = append(b.Tree, models.Object{
		Tok:  enum.Tok,
		Data: enum,
	})
}

// Comment builds AST model of comment.
func (b *Builder) Comment(tok Tok) {
	tok.Kind = strings.TrimSpace(tok.Kind[2:])
	if strings.HasPrefix(tok.Kind, "cxx:") {
		b.Tree = append(b.Tree, models.Object{
			Tok: tok,
			Data: models.CxxEmbed{
				Tok:     tok,
				Content: tok.Kind[4:]},
		})
		return
	}
	b.Tree = append(b.Tree, models.Object{
		Tok: tok,
		Data: models.Comment{
			Content: tok.Kind,
		},
	})
}

// Preprocessor builds AST model of preprocessor directives.
func (b *Builder) Preprocessor(toks Toks) {
	if len(toks) == 1 {
		b.pusherr(toks[0], "invalid_syntax")
		return
	}
	var pp models.Preprocessor
	toks = toks[1:] // Remove directive mark
	tok := toks[0]
	if tok.Id != tokens.Id {
		b.pusherr(pp.Tok, "invalid_syntax")
		return
	}
	ok := false
	switch tok.Kind {
	case x.PreprocessorDirective:
		ok = b.PreprocessorDirective(&pp, toks)
	default:
		b.pusherr(tok, "invalid_preprocessor")
		return
	}
	if ok {
		b.Tree = append(b.Tree, models.Object{
			Tok:  pp.Tok,
			Data: pp,
		})
	}
}

// PreprocessorDirective builds AST model of preprocessor pragma directive.
// Returns true if success, returns false if not.
func (b *Builder) PreprocessorDirective(pp *models.Preprocessor, toks Toks) bool {
	if len(toks) == 1 {
		b.pusherr(toks[0], "missing_pragma_directive")
		return false
	}
	toks = toks[1:] // Remove pragma identifier
	tok := toks[0]
	if tok.Id != tokens.Id {
		b.pusherr(tok, "invalid_syntax")
		return false
	}
	var d models.Directive
	ok := false
	switch tok.Kind {
	case x.PreprocessorDirectiveEnofi:
		ok = b.directiveEnofi(&d, toks)
	default:
		b.pusherr(tok, "invalid_pragma_directive")
	}
	pp.Command = d
	return ok
}

func (b *Builder) directiveEnofi(d *models.Directive, toks Toks) bool {
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
		return false
	}
	d.Command = models.DirectiveEnofi{}
	return true
}

// Id builds AST model of global id statement.
func (b *Builder) Id(toks Toks) {
	if len(toks) == 1 {
		b.pusherr(toks[0], "invalid_syntax")
		return
	}
	tok := toks[1]
	switch tok.Id {
	case tokens.Colon:
		b.GlobalVar(toks)
		return
	case tokens.DoubleColon:
		b.Namespace(toks)
		return
	case tokens.Brace:
		switch tok.Kind {
		case tokens.LBRACE: // Namespace.
			b.Namespace(toks)
			return
		}
	}
	b.pusherr(tok, "invalid_syntax")
}

func (b *Builder) nsIds(toks Toks, i *int) []string {
	var ids []string
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		if (*i+1)%2 != 0 {
			if tok.Id != tokens.Id {
				b.pusherr(tok, "invalid_syntax")
				continue
			}
			ids = append(ids, tok.Kind)
			continue
		}
		switch tok.Id {
		case tokens.DoubleColon:
			continue
		default:
			goto ret
		}
	}
ret:
	return ids
}

// Namespace builds AST model of namespace statement.
func (b *Builder) Namespace(toks Toks) {
	var ns models.Namespace
	ns.Tok = toks[0]
	i := new(int)
	ns.Ids = b.nsIds(toks, i)
	treeToks := b.getrange(i, tokens.LBRACE, tokens.RBRACE, &toks)
	if treeToks == nil {
		b.pusherr(ns.Tok, "body_not_exist")
		return
	}
	if *i < len(toks) {
		b.pusherr(toks[*i], "invalid_syntax")
	}
	tree := b.Tree
	b.Tree = nil
	btoks := b.Toks
	pos := b.Pos
	b.Toks = treeToks
	b.Pos = 0
	b.Build()
	b.Toks = btoks
	b.Pos = pos
	ns.Tree = b.Tree
	b.Tree = tree
	b.Tree = append(b.Tree, models.Object{
		Tok:  ns.Tok,
		Data: ns,
	})
}

func (b *Builder) structFields(toks Toks) []*models.Var {
	fields := make([]*models.Var, 0)
	i := new(int)
	for *i < len(toks) {
		varToks := b.skipStatement(i, &toks)
		pub := varToks[0].Id == tokens.Pub
		if pub {
			if len(varToks) == 1 {
				b.pusherr(varToks[0], "invalid_syntax")
				continue
			}
			varToks = varToks[1:]
		}
		vast := b.Var(varToks, false)
		vast.Pub = pub
		fields = append(fields, &vast)
	}
	return fields
}

// Struct builds AST model of structure.
func (b *Builder) Struct(toks Toks) {
	var s models.Struct
	s.Pub = b.pub
	b.pub = false
	if len(toks) < 3 {
		b.pusherr(toks[0], "invalid_syntax")
		return
	}
	s.Tok = toks[1]
	if s.Tok.Id != tokens.Id {
		b.pusherr(s.Tok, "invalid_syntax")
	}
	s.Id = s.Tok.Kind
	i := 2
	bodyToks := b.getrange(&i, tokens.LBRACE, tokens.RBRACE, &toks)
	if bodyToks == nil {
		b.pusherr(s.Tok, "body_not_exist")
		return
	}
	if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	s.Fields = b.structFields(bodyToks)
	b.Tree = append(b.Tree, models.Object{
		Tok:  s.Tok,
		Data: s,
	})
}

func (b *Builder) traitFuncs(toks Toks) []*models.Func {
	var funcs []*models.Func
	i := 0
	for i < len(toks) {
		funcToks := b.skipStatement(&i, &toks)
		f := b.Func(funcToks, false, true)
		f.Pub = true
		funcs = append(funcs, &f)
	}
	return funcs
}

// Trait builds AST model of trait.
func (b *Builder) Trait(toks Toks) {
	var t models.Trait
	t.Pub = b.pub
	b.pub = false
	if len(toks) < 3 {
		b.pusherr(toks[0], "invalid_syntax")
		return
	}
	t.Tok = toks[1]
	if t.Tok.Id != tokens.Id {
		b.pusherr(t.Tok, "invalid_syntax")
	}
	t.Id = t.Tok.Kind
	i := 2
	bodyToks := b.getrange(&i, tokens.LBRACE, tokens.RBRACE, &toks)
	if bodyToks == nil {
		b.pusherr(t.Tok, "body_not_exist")
		return
	}
	if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	t.Funcs = b.traitFuncs(bodyToks)
	b.Tree = append(b.Tree, models.Object{Tok: t.Tok, Data: t})
}

func (b *Builder) implFuncs(toks Toks) []*models.Func {
	var funcs []*models.Func
	i := 0
	for i < len(toks) {
		funcToks := b.skipStatement(&i, &toks)
		f := b.Func(funcToks, false, false)
		f.Pub = true
		funcs = append(funcs, &f)
	}
	return funcs
}

// Impl builds AST model of impl statement.
func (b *Builder) Impl(toks Toks) {
	tok := toks[0]
	if len(toks) < 2 {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	tok = toks[1]
	if tok.Id != tokens.Id {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	if len(toks) < 3 {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	var impl models.Impl
	impl.Trait = tok
	tok = toks[2]
	if tok.Id != tokens.In {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	if len(toks) < 4 {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	i := 3
	tok = toks[i]
	if tok.Id == tokens.Operator {
		if tok.Kind != tokens.STAR {
			b.pusherr(tok, "invalid_syntax")
			return
		}
		i++
		tok = toks[i]
	}
	if tok.Id != tokens.Id {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	i++
	receiverToks := toks[3:i]
	toks = toks[i:]
	i = 0
	impl.Target, _ = b.DataType(receiverToks, &i, false, true)
	if i+1 < len(receiverToks) {
		b.pusherr(receiverToks[i+1], "invalid_syntax")
	}
	i = 0
	bodyToks := b.getrange(&i, tokens.LBRACE, tokens.RBRACE, &toks)
	if bodyToks == nil {
		b.pusherr(impl.Trait, "body_not_exist")
		return
	}
	if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	impl.Funcs = b.implFuncs(bodyToks)
	b.Tree = append(b.Tree, models.Object{Tok: impl.Trait, Data: impl})
}

func tokstoa(toks Toks) string {
	var str strings.Builder
	for _, tok := range toks {
		str.WriteString(tok.Kind)
	}
	return str.String()
}

// Use builds AST model of use declaration.
func (b *Builder) Use(toks Toks) {
	var use models.Use
	use.Tok = toks[0]
	if len(toks) < 2 {
		b.pusherr(use.Tok, "missing_use_path")
		return
	}
	toks = toks[1:]
	use.LinkString = tokstoa(toks)
	use.Path = b.usePath(toks)
	b.Tree = append(b.Tree, models.Object{
		Tok:  use.Tok,
		Data: use,
	})
}

func (b *Builder) usePath(toks Toks) string {
	var path strings.Builder
	path.WriteString(x.StdlibPath)
	path.WriteRune(os.PathSeparator)
	for i, tok := range toks {
		if i%2 != 0 {
			if tok.Id != tokens.Dot {
				b.pusherr(tok, "invalid_syntax")
			}
			path.WriteRune(os.PathSeparator)
			continue
		}
		if tok.Id != tokens.Id {
			b.pusherr(tok, "invalid_syntax")
		}
		path.WriteString(tok.Kind)
	}
	return path.String()
}

// Attribute builds AST model of attribute.
func (b *Builder) Attribute(toks Toks) {
	var a models.Attribute
	i := 0
	a.Tok = toks[i]
	i++
	if b.Ended() {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	a.Tag = toks[i]
	if a.Tag.Id != tokens.Id || a.Tok.Column+1 != a.Tag.Column {
		b.pusherr(a.Tag, "invalid_syntax")
		return
	}
	toks = toks[i+1:]
	if len(toks) > 0 {
		tok := toks[0]
		if a.Tok.Column+len(a.Tag.Kind)+1 == tok.Column {
			b.pusherr(tok, "invalid_syntax")
		}
		b.Toks = append(toks, b.Toks...)
	}
	b.Tree = append(b.Tree, models.Object{
		Tok:  a.Tok,
		Data: a,
	})
}

// GlobalFunc builds AST model of global scope function.
func (b *Builder) GlobalFunc(toks Toks) bool {
	if len(toks) == 0 {
		return false
	}
	i := 0
	tok := toks[i]
	if tok.Id == tokens.Operator {
		if tok.Kind != tokens.STAR {
			return false
		}
		i++
		tok = toks[i]
	}
	if tok.Id != tokens.Id {
		return false
	}
	if i+1 >= len(toks) {
		return false
	}
	var receiver *models.DataType
	funcToks := toks
	i++
	tok = toks[i]
	if tok.Id == tokens.Dot {
		typeToks := toks[:i]
		funcToks = toks[i+1:]
		receiver = new(models.DataType)
		*receiver, _ = b.DataType(typeToks, new(int), true, true)
	}
	if len(funcToks) < 2 {
		return false
	}
	tok = funcToks[1]
	if tok.Id != tokens.Brace || tok.Kind != tokens.LPARENTHESES {
		return false
	}
	f := b.Func(funcToks, false, false)
	f.Receiver = receiver
	s := models.Statement{Tok: f.Tok, Data: f}
	b.Tree = append(b.Tree, models.Object{Tok: f.Tok, Data: s})
	return true
}

func (b *Builder) funcPrototype(toks *Toks, anon bool) (f models.Func, ok bool) {
	ok = true
	f.Tok = (*toks)[0]
	i := 0
	f.Pub = b.pub
	b.pub = false
	if anon {
		f.Id = x.Anonymous
	} else {
		if f.Tok.Id != tokens.Id {
			b.pusherr(f.Tok, "invalid_syntax")
			ok = false
		}
		f.Id = f.Tok.Kind
		i++
	}
	f.RetType.Type.Id = xtype.Void
	f.RetType.Type.Kind = xtype.TypeMap[f.RetType.Type.Id]
	paramToks := b.getrange(&i, tokens.LPARENTHESES, tokens.RPARENTHESES, toks)
	if len(paramToks) > 0 {
		b.Params(&f, paramToks)
	}
	t, retok := b.FuncRetDataType(*toks, &i)
	if retok {
		f.RetType = t
		i++
	}
	*toks = (*toks)[i:]
	return
}

// Func builds AST model of function.
func (b *Builder) Func(toks Toks, anon, prototype bool) (f models.Func) {
	f, ok := b.funcPrototype(&toks, anon)
	if !ok {
		return
	}
	if len(toks) == 0 {
		if prototype {
			return
		} else if b.Ended() {
			b.pusherr(f.Tok, "body_not_exist")
			return
		}
		toks = b.nextBuilderStatement()
	} else if prototype {
		b.pusherr(f.Tok, "invalid_syntax")
		return
	}
	i := 0
	blockToks := b.getrange(&i, tokens.LBRACE, tokens.RBRACE, &toks)
	if blockToks == nil {
		b.pusherr(f.Tok, "body_not_exist")
		return
	} else if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	f.Block = b.Block(blockToks)
	return
}

func (b *Builder) generic(toks Toks) models.GenericType {
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
	}
	var gt models.GenericType
	gt.Tok = toks[0]
	if gt.Tok.Id != tokens.Id {
		b.pusherr(gt.Tok, "invalid_syntax")
	}
	gt.Id = gt.Tok.Kind
	return gt
}

// Generic builds generic type.
func (b *Builder) Generics(toks Toks) []models.GenericType {
	tok := toks[0]
	i := 1
	genericsToks := Range(&i, tokens.LBRACKET, tokens.RBRACKET, toks)
	if len(genericsToks) == 0 {
		b.pusherr(tok, "missing_expr")
		return make([]models.GenericType, 0)
	} else if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	parts, errs := Parts(genericsToks, tokens.Comma, true)
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
func (b *Builder) TypeOrGenerics(toks Toks) models.Object {
	if len(toks) > 1 {
		tok := toks[1]
		if tok.Id == tokens.Brace && tok.Kind == tokens.LBRACKET {
			generics := b.Generics(toks)
			return models.Object{
				Tok:  tok,
				Data: generics,
			}
		}
	}
	t := b.Type(toks)
	t.Pub = b.pub
	b.pub = false
	return models.Object{
		Tok:  t.Tok,
		Data: t,
	}
}

// GlobalVar builds AST model of global variable.
func (b *Builder) GlobalVar(toks Toks) {
	if toks == nil {
		return
	}
	s := b.VarStatement(toks)
	b.Tree = append(b.Tree, models.Object{
		Tok:  s.Tok,
		Data: s,
	})
}

// Params builds AST model of function parameters.
func (b *Builder) Params(f *models.Func, toks Toks) {
	parts, errs := Parts(toks, tokens.Comma, true)
	b.Errors = append(b.Errors, errs...)
	for _, part := range parts {
		b.pushParam(f, part)
	}
	b.wg.Add(1)
	go b.checkParams(f)
}

func (b *Builder) checkParams(f *models.Func) {
	defer b.wg.Done()
	for i := range f.Params {
		p := &f.Params[i]
		if p.Type.Tok.Id == tokens.NA {
			if p.Tok.Id == tokens.NA {
				b.pusherr(p.Tok, "missing_type")
			} else {
				p.Type.Tok = p.Tok
				p.Type.Id = xtype.Id
				p.Type.Kind = p.Type.Tok.Kind
				p.Type.Original = p.Type
				p.Id = x.Anonymous
				p.Tok = lex.Tok{}
			}
		}
	}
}

func (b *Builder) paramBegin(p *models.Param, i *int, toks Toks) {
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		switch tok.Id {
		case tokens.Operator:
			switch tok.Kind {
			case tokens.TRIPLE_DOT:
				if p.Variadic {
					b.pusherr(tok, "already_variadic")
					continue
				}
				p.Variadic = true
			case tokens.AMPER:
				if p.Reference {
					b.pusherr(tok, "already_reference")
					continue
				}
				p.Reference = true
			default:
				b.pusherr(tok, "invalid_syntax")
			}
		default:
			return
		}
	}
}

func (b *Builder) paramBodyId(f *models.Func, p *models.Param, tok Tok) {
	if xapi.IsIgnoreId(tok.Kind) {
		p.Id = x.Anonymous
		return
	}
	for _, param := range f.Params {
		if param.Id == tok.Kind {
			b.pusherr(tok, "parameter_exist", tok.Kind)
			break
		}
	}
	p.Id = tok.Kind
}

func (b *Builder) paramBodyDataType(f *models.Func, p *models.Param, toks Toks) {
	i := 0
	p.Type, _ = b.DataType(toks, &i, false, true)
	i++
	if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	// Set param data types to this data type
	// if parameter has not any data type.
	i = len(f.Params) - 1
	for ; i >= 0; i-- {
		param := &f.Params[i]
		if param.Type.Tok.Id != tokens.NA {
			break
		}
		param.Type = p.Type
	}
}

func (b *Builder) paramBody(f *models.Func, p *models.Param, i *int, toks Toks) {
	b.paramBodyId(f, p, toks[*i])
	// +1 for skip identifier token
	toks = toks[*i+1:]
	if len(toks) == 0 {
		return
	}
	if len(toks) > 0 {
		b.paramBodyDataType(f, p, toks)
	}
}

func (b *Builder) pushParam(f *models.Func, toks Toks) {
	var p models.Param
	i := 0
	b.paramBegin(&p, &i, toks)
	if i >= len(toks) {
		return
	}
	tok := toks[i]
	p.Tok = tok
	// Just given data-type.
	if tok.Id != tokens.Id {
		if t, ok := b.DataType(toks, &i, false, true); ok {
			if i+1 == len(toks) {
				p.Type = t
			}
		}
		goto end
	}
	b.paramBody(f, &p, &i, toks)
end:
	f.Params = append(f.Params, p)
}

func (b *Builder) idGenericsParts(toks Toks, i *int) []Toks {
	first := *i
	braceCount := 0
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACKET:
				braceCount++
			case tokens.RBRACKET:
				braceCount--
			}
		}
		if braceCount == 0 {
			break
		}
	}
	toks = toks[first+1 : *i]
	parts, errs := Parts(toks, tokens.Comma, true)
	b.Errors = append(b.Errors, errs...)
	return parts
}

func (b *Builder) idDataTypePartEnd(t *models.DataType, dtv *strings.Builder, toks Toks, i *int) {
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
	parts := b.idGenericsParts(toks, i)
	generics := make([]models.DataType, len(parts))
	for i, part := range parts {
		index := 0
		t, _ := b.DataType(part, &index, false, true)
		if index+1 < len(part) {
			b.pusherr(part[index+1], "invalid_syntax")
		}
		genericsStr.WriteString(t.String())
		genericsStr.WriteByte(',')
		generics[i] = t
	}
	dtv.WriteString(genericsStr.String()[:genericsStr.Len()-1])
	dtv.WriteByte(']')
	t.Tag = generics
}

// DataType builds AST model of data-type.
func (b *Builder) DataType(toks Toks, i *int, arrays, err bool) (t models.DataType, ok bool) {
	defer func() { t.Original = t }()
	first := *i
	var dtv strings.Builder
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		switch tok.Id {
		case tokens.DataType:
			t.Tok = tok
			t.Id = xtype.TypeFromId(t.Tok.Kind)
			dtv.WriteString(t.Tok.Kind)
			ok = true
			goto ret
		case tokens.Id:
			dtv.WriteString(tok.Kind)
			if *i+1 < len(toks) && toks[*i+1].Id == tokens.DoubleColon {
				break
			}
			t.Id = xtype.Id
			t.Tok = tok
			b.idDataTypePartEnd(&t, &dtv, toks, i)
			ok = true
			goto ret
		case tokens.DoubleColon:
			dtv.WriteString(tok.Kind)
		case tokens.Operator:
			if tok.Kind == tokens.STAR {
				dtv.WriteString(tok.Kind)
				break
			}
			if err {
				b.pusherr(tok, "invalid_syntax")
			}
			return
		case tokens.Brace:
			switch tok.Kind {
			case tokens.LPARENTHESES:
				t.Tok = tok
				t.Id = xtype.Func
				f := b.FuncDataTypeHead(toks, i)
				*i++
				f.RetType, ok = b.FuncRetDataType(toks, i)
				if !ok {
					*i--
				}
				t.Tag = &f
				dtv.WriteString(f.DataTypeString())
				ok = true
				goto ret
			case tokens.LBRACKET:
				*i++
				if *i > len(toks) {
					if err {
						b.pusherr(tok, "invalid_syntax")
					}
					return
				}
				tok = toks[*i]
				if tok.Id == tokens.Brace && tok.Kind == tokens.RBRACKET {
					arrays = false
					dtv.WriteString(x.Prefix_Slice)
					continue
				}
				*i-- // Start from bracket
				if arrays {
					t = b.MapOrArrayDataType(toks, i, err)
				} else {
					t = b.MapDataType(toks, i, err)
				}
				if t.Id == xtype.Void {
					if err {
						b.pusherr(tok, "invalid_syntax")
					}
					return
				}
				t.Kind = dtv.String() + t.Kind
				ok = true
				return
			}
			/*if err {
				ast.pusherrtok(tok, "invalid_syntax")
			}*/
			return
		default:
			if err {
				b.pusherr(tok, "invalid_syntax")
			}
			return
		}
	}
	if err {
		b.pusherr(toks[first], "invalid_type")
	}
ret:
	t.Kind = dtv.String()
	return
}

func (b *Builder) arrayDataType(toks Toks, i *int, err bool) (t models.DataType) {
	defer func() { t.Original = t }()
	if *i+1 >= len(toks) {
		return
	}
	*i++
	exprI := *i
	t, ok := b.DataType(toks, i, true, err)
	if !ok {
		return
	}
	var exprs [][]any
	if t.Tag != nil {
		exprs = t.Tag.([][]any)
	}
	_, exprToks := RangeLast(toks[:exprI])
	exprToks = exprToks[1 : len(exprToks)-1]
	tok := exprToks[0]
	if len(exprToks) == 1 && tok.Id == tokens.Operator && tok.Kind == tokens.TRIPLE_DOT {
		exprs = append([][]any{{uint64(0), models.Expr{}}}, exprs...)
	} else {
		exprs = append([][]any{{uint64(0), b.Expr(exprToks)}}, exprs...)
	}
	t.Tag = exprs
	t.Kind = x.Prefix_Array + t.Kind
	return t
}

func (b *Builder) MapOrArrayDataType(toks Toks, i *int, err bool) models.DataType {
	t := b.MapDataType(toks, i, err)
	if t.Id == xtype.Void { // Array
		return b.arrayDataType(toks, i, err)
	}
	return t
}

// MapDataType builds map data-type.
func (b *Builder) MapDataType(toks Toks, i *int, err bool) (t models.DataType) {
	typeToks, colon := SplitColon(toks, i)
	if typeToks == nil || colon == -1 {
		return
	}
	return b.mapDataType(toks, typeToks, colon, err)
}

func (b *Builder) mapDataType(toks, typeToks Toks, colon int, err bool) (t models.DataType) {
	defer func() { t.Original = t }()
	t.Id = xtype.Map
	t.Tok = toks[0]
	colonTok := toks[colon]
	if colon == 0 || colon+1 >= len(typeToks) {
		if err {
			b.pusherr(colonTok, "missing_expr")
		}
		return t
	}
	keyTypeToks := typeToks[:colon]
	valueTypeToks := typeToks[colon+1:]
	types := make([]models.DataType, 2)
	j := 0
	types[0], _ = b.DataType(keyTypeToks, &j, true, err)
	j = 0
	types[1], _ = b.DataType(valueTypeToks, &j, true, err)
	t.Tag = types
	t.Kind = t.MapKind()
	return t
}

// FuncDataTypeHead builds head part of function data-type.
func (b *Builder) FuncDataTypeHead(toks Toks, i *int) models.Func {
	var f models.Func
	brace := 1
	firstIndex := *i
	for *i++; *i < len(toks); *i++ {
		tok := toks[*i]
		switch tok.Id {
		case tokens.Brace:
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				brace++
			default:
				brace--
			}
		}
		if brace == 0 {
			b.Params(&f, toks[firstIndex+1:*i])
			return f
		}
	}
	b.pusherr(toks[firstIndex], "invalid_type")
	return f
}

func (b *Builder) pushTypeToTypes(ids *Toks, types *[]models.DataType, toks Toks, errTok Tok) {
	if len(toks) == 0 {
		b.pusherr(errTok, "missing_expr")
		return
	}
	tok := toks[0]
	if tok.Id == tokens.Id && len(toks) > 1 {
		*ids = append(*ids, tok)
		toks = toks[1:]
	} else {
		*ids = append(*ids, Tok{Kind: xapi.Ignore})
	}
	index := new(int)
	currentDt, ok := b.DataType(toks, index, false, true)
	if !ok {
		return
	} else if *index < len(toks)-1 {
		b.pusherr(toks[*index], "invalid_syntax")
	}
	*types = append(*types, currentDt)
}

func (b *Builder) funcMultiTypeRet(toks Toks, i *int) (t models.RetType, ok bool) {
	defer func() { t.Type.Original = t.Type }()
	start := *i
	tok := toks[*i]
	t.Type.Kind += tok.Kind
	*i++
	if *i >= len(toks) {
		*i--
		t.Type, ok = b.DataType(toks, i, false, false)
		return
	}
	tok = toks[*i]
	// Is slice?
	if tok.Id == tokens.Brace && tok.Kind == tokens.RBRACKET {
		*i--
		t.Type, ok = b.DataType(toks, i, false, false)
		return
	}
	var types []models.DataType
	braceCount := 1
	last := *i
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		t.Type.Kind += tok.Kind
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount == 0 {
			if tok.Id == tokens.Colon { // Is map
				*i = start
				t.Type, ok = b.DataType(toks, i, false, false)
				return
			}
			b.pushTypeToTypes(&t.Identifiers, &types, toks[last:*i], toks[last-1])
			break
		} else if braceCount > 1 {
			continue
		}
		switch tok.Id {
		case tokens.Comma:
		case tokens.Colon:
			*i = start
			t.Type, ok = b.DataType(toks, i, false, false)
			return
		default:
			continue
		}
		b.pushTypeToTypes(&t.Identifiers, &types, toks[last:*i], toks[*i-1])
		last = *i + 1
	}
	if len(types) > 1 {
		t.Type.MultiTyped = true
		t.Type.Tag = types
	} else {
		t.Type = types[0]
	}
	ok = true
	return
}

// FuncRetDataType builds ret data-type of function.
func (b *Builder) FuncRetDataType(toks Toks, i *int) (t models.RetType, ok bool) {
	defer func() { t.Type.Original = t.Type }()
	t.Type.Id = xtype.Void
	t.Type.Kind = xtype.TypeMap[t.Type.Id]
	if *i >= len(toks) {
		return
	}
	tok := toks[*i]
	// Multityped?
	if tok.Id == tokens.Brace && tok.Kind == tokens.LBRACKET {
		return b.funcMultiTypeRet(toks, i)
	}
	t.Type, ok = b.DataType(toks, i, false, false)
	return
}

func (b *Builder) pushStatementToBlock(bs *blockStatement) {
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
	s := b.Statement(bs)
	if s.Data == nil {
		return
	}
	s.WithTerminator = bs.withTerminator
	bs.block.Tree = append(bs.block.Tree, s)
}

// Block builds AST model of statements of code block.
func (b *Builder) Block(toks Toks) (block *models.Block) {
	block = new(models.Block)
	bs := new(blockStatement)
	bs.block = block
	bs.srcToks = &toks
	for {
		bs.pos, bs.withTerminator = NextStatementPos(toks, 0)
		statementToks := toks[:bs.pos]
		bs.blockToks = &toks
		bs.toks = statementToks
		b.pushStatementToBlock(bs)
	next:
		if len(bs.nextToks) > 0 {
			bs.toks = bs.nextToks
			bs.nextToks = nil
			b.pushStatementToBlock(bs)
			goto next
		}
		if bs.pos >= len(toks) {
			break
		}
		toks = toks[bs.pos:]
	}
	return
}

// Statement builds AST model of statement.
func (b *Builder) Statement(bs *blockStatement) (s models.Statement) {
	s, ok := b.AssignStatement(bs.toks, false)
	if ok {
		return s
	}
	tok := bs.toks[0]
	switch tok.Id {
	case tokens.Id:
		s, ok := b.IdStatement(bs.toks)
		if ok {
			return s
		}
	case tokens.Const, tokens.Volatile:
		return b.VarStatement(bs.toks)
	case tokens.Ret:
		return b.RetStatement(bs.toks)
	case tokens.Iter:
		return b.IterExpr(bs.toks)
	case tokens.Break:
		return b.BreakStatement(bs.toks)
	case tokens.Continue:
		return b.ContinueStatement(bs.toks)
	case tokens.If:
		return b.IfExpr(bs)
	case tokens.Else:
		return b.ElseBlock(bs)
	case tokens.Comment:
		return b.CommentStatement(bs.toks[0])
	case tokens.Defer:
		return b.DeferStatement(bs.toks)
	case tokens.Co:
		return b.ConcurrentCallStatement(bs.toks)
	case tokens.Goto:
		return b.GotoStatement(bs.toks)
	case tokens.Type:
		t := b.Type(bs.toks)
		s.Tok = t.Tok
		s.Data = t
		return
	case tokens.Match:
		return b.MatchCase(bs.toks)
	case tokens.Brace:
		if tok.Kind == tokens.LBRACE {
			return b.blockStatement(bs.toks)
		}
	}
	if IsFuncCall(bs.toks) != nil {
		return b.ExprStatement(bs)
	}
	tok = Tok{
		File:   tok.File,
		Id:     tokens.Ret,
		Kind:   tokens.RET,
		Row:    tok.Row,
		Column: tok.Column,
	}
	bs.toks = append([]Tok{tok}, bs.toks...)
	return b.RetStatement(bs.toks)
}

func (b *Builder) blockStatement(toks Toks) models.Statement {
	i := new(int)
	tok := toks[0]
	toks = Range(i, tokens.LBRACE, tokens.RBRACE, toks)
	if *i < len(toks) {
		b.pusherr(toks[*i], "invalid_syntax")
	}
	block := b.Block(toks)
	return models.Statement{Tok: tok, Data: block}
}

func (b *Builder) assignInfo(toks Toks) (info AssignInfo) {
	info.Ok = true
	braceCount := 0
	for i, tok := range toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		} else if tok.Id != tokens.Operator {
			continue
		} else if !IsAssignOperator(tok.Kind) {
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
			info.Ok = IsSuffixOperator(info.Setter.Kind)
			break
		}
		info.Right = toks[i+1:]
		if IsSuffixOperator(info.Setter.Kind) {
			if info.Right != nil {
				b.pusherr(info.Right[0], "invalid_syntax")
				info.Right = nil
			}
		}
		break
	}
	return
}

func (b *Builder) pushAssignLeft(lefts *[]models.AssignLeft, last, current int, info AssignInfo) {
	var left models.AssignLeft
	left.Expr.Toks = info.Left[last:current]
	if last-current == 0 {
		b.pusherr(info.Left[current-1], "missing_expr")
		return
	}
	// Variable is new?
	if left.Expr.Toks[0].Id == tokens.Id &&
		current-last > 1 &&
		left.Expr.Toks[1].Id == tokens.Colon {
		if info.IsExpr {
			b.pusherr(left.Expr.Toks[0], "notallow_declares")
		}
		left.Var.New = true
		left.Var.IdTok = left.Expr.Toks[0]
		left.Var.Id = left.Var.IdTok.Kind
		left.Var.SetterTok = info.Setter
		// Has specific data-type?
		if current-last > 2 {
			left.Var.Type, _ = b.DataType(left.Expr.Toks[2:], new(int), true, false)
		}
	} else {
		if left.Expr.Toks[0].Id == tokens.Id {
			left.Var.IdTok = left.Expr.Toks[0]
			left.Var.Id = left.Var.IdTok.Kind
		}
		left.Expr = b.Expr(left.Expr.Toks)
	}
	*lefts = append(*lefts, left)
}

func (b *Builder) assignLefts(info AssignInfo) []models.AssignLeft {
	var lefts []models.AssignLeft
	braceCount := 0
	lastIndex := 0
	for i, tok := range info.Left {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		} else if tok.Id != tokens.Comma {
			continue
		}
		b.pushAssignLeft(&lefts, lastIndex, i, info)
		lastIndex = i + 1
	}
	if lastIndex < len(info.Left) {
		b.pushAssignLeft(&lefts, lastIndex, len(info.Left), info)
	}
	return lefts
}

func (b *Builder) pushAssignExpr(exps *[]models.Expr, last, current int, info AssignInfo) {
	toks := info.Right[last:current]
	if toks == nil {
		b.pusherr(info.Right[current-1], "missing_expr")
		return
	}
	*exps = append(*exps, b.Expr(toks))
}

func (b *Builder) assignExprs(info AssignInfo) []models.Expr {
	var exprs []models.Expr
	braceCount := 0
	lastIndex := 0
	for i, tok := range info.Right {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		} else if tok.Id != tokens.Comma {
			continue
		}
		b.pushAssignExpr(&exprs, lastIndex, i, info)
		lastIndex = i + 1
	}
	if lastIndex < len(info.Right) {
		b.pushAssignExpr(&exprs, lastIndex, len(info.Right), info)
	}
	return exprs
}

// AssignStatement builds AST model of assignment statement.
func (b *Builder) AssignStatement(toks Toks, isExpr bool) (s models.Statement, _ bool) {
	assign, ok := b.AssignExpr(toks, isExpr)
	if !ok {
		return
	}
	s.Tok = toks[0]
	s.Data = assign
	return s, true
}

// AssignExpr builds AST model of assignment expression.
func (b *Builder) AssignExpr(toks Toks, isExpr bool) (assign models.Assign, ok bool) {
	if !CheckAssignToks(toks) {
		return
	}
	info := b.assignInfo(toks)
	if !info.Ok {
		return
	}
	ok = true
	info.IsExpr = isExpr
	assign.IsExpr = isExpr
	assign.Setter = info.Setter
	assign.Left = b.assignLefts(info)
	if isExpr && len(assign.Left) > 1 {
		b.pusherr(assign.Setter, "notallow_multiple_assign")
	}
	if info.Right != nil {
		assign.Right = b.assignExprs(info)
	}
	return
}

// BuildReturnStatement builds AST model of return statement.
func (b *Builder) IdStatement(toks Toks) (s models.Statement, ok bool) {
	if len(toks) == 1 {
		return
	}
	tok := toks[1]
	switch tok.Id {
	case tokens.Colon:
		if len(toks) == 2 { // Label?
			return b.LabelStatement(toks[0]), true
		}
		return b.VarStatement(toks), true
	}
	return
}

// LabelStatement builds AST model of label.
func (b *Builder) LabelStatement(tok Tok) models.Statement {
	var l models.Label
	l.Tok = tok
	l.Label = tok.Kind
	return models.Statement{Tok: tok, Data: l}
}

// ExprStatement builds AST model of expression.
func (b *Builder) ExprStatement(bs *blockStatement) models.Statement {
	expr := models.ExprStatement{
		Expr: b.Expr(bs.toks),
	}
	return models.Statement{
		Tok:  bs.toks[0],
		Data: expr,
	}
}

// Args builds AST model of arguments.
func (b *Builder) Args(toks Toks) *models.Args {
	args := new(models.Args)
	last := 0
	braceCount := 0
	for i, tok := range toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 || tok.Id != tokens.Comma {
			continue
		}
		b.pushArg(args, toks[last:i], tok)
		last = i + 1
	}
	if last < len(toks) {
		if last == 0 {
			if len(toks) > 0 {
				b.pushArg(args, toks[last:], toks[last])
			}
		} else {
			b.pushArg(args, toks[last:], toks[last-1])
		}
	}
	return args
}

func (b *Builder) pushArg(args *models.Args, toks Toks, err Tok) {
	if len(toks) == 0 {
		b.pusherr(err, "invalid_syntax")
		return
	}
	var arg models.Arg
	arg.Tok = toks[0]
	if arg.Tok.Id == tokens.Id {
		if len(toks) > 1 {
			tok := toks[1]
			if tok.Id == tokens.Operator && tok.Kind == tokens.EQUAL {
				args.Targeted = true
				arg.TargetId = arg.Tok.Kind
				toks = toks[2:]
			}
		}
	}
	arg.Expr = b.Expr(toks)
	args.Src = append(args.Src, arg)
}

func (b *Builder) varBegin(v *models.Var, i *int, toks Toks) {
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		if tok.Id == tokens.Id {
			break
		}
		switch tok.Id {
		case tokens.Const:
			if v.Const {
				b.pusherr(tok, "already_constant")
				break
			}
			v.Const = true
		case tokens.Volatile:
			if v.Volatile {
				b.pusherr(tok, "already_volatile")
				break
			}
			v.Volatile = true
		default:
			b.pusherr(tok, "invalid_syntax")
		}
	}
}

func (b *Builder) varTypeNExpr(v *models.Var, toks Toks, i int) {
	tok := toks[i]
	t, ok := b.DataType(toks, &i, true, false)
	if ok {
		v.Type = t
		i++
		if i >= len(toks) {
			return
		}
		tok = toks[i]
	}
	if tok.Id == tokens.Operator {
		if tok.Kind != tokens.EQUAL {
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
func (b *Builder) Var(toks Toks, begin bool) (v models.Var) {
	v.Pub = b.pub
	b.pub = false
	i := 0
	v.DefTok = toks[i]
	if begin {
		b.varBegin(&v, &i, toks)
		if i >= len(toks) {
			return
		}
	}
	v.IdTok = toks[i]
	if v.IdTok.Id != tokens.Id {
		b.pusherr(v.IdTok, "invalid_syntax")
	}
	v.Id = v.IdTok.Kind
	v.Type.Id = xtype.Void
	v.Type.Kind = xtype.TypeMap[v.Type.Id]
	// Skip type definer operator(':')
	i++
	if i >= len(toks) {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	// If exist deftok, this token is now should be colon.
	if v.DefTok.File != nil {
		if toks[i].Id != tokens.Colon {
			b.pusherr(toks[i], "invalid_syntax")
			return
		}
		i++
	}
	if i < len(toks) {
		b.varTypeNExpr(&v, toks, i)
	}
	return
}

// VarStatement builds AST model of variable declaration statement.
func (b *Builder) VarStatement(toks Toks) models.Statement {
	vast := b.Var(toks, true)
	return models.Statement{
		Tok:  vast.IdTok,
		Data: vast,
	}
}

// CommentStatement builds AST model of comment statement.
func (b *Builder) CommentStatement(tok Tok) (s models.Statement) {
	s.Tok = tok
	tok.Kind = strings.TrimSpace(tok.Kind[2:])
	if strings.HasPrefix(tok.Kind, "cxx:") {
		s.Data = models.CxxEmbed{
			Tok:     tok,
			Content: tok.Kind[4:],
		}
	} else {
		s.Data = models.Comment{
			Content: tok.Kind,
		}
	}
	return
}

// DeferStatement builds AST model of deferred call statement.
func (b *Builder) DeferStatement(toks Toks) (s models.Statement) {
	var d models.Defer
	d.Tok = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		b.pusherr(d.Tok, "missing_expr")
		return
	}
	if IsFuncCall(toks) == nil {
		b.pusherr(d.Tok, "expr_not_func_call")
	}
	d.Expr = b.Expr(toks)
	s.Tok = d.Tok
	s.Data = d
	return
}

func (b *Builder) ConcurrentCallStatement(toks Toks) (s models.Statement) {
	var cc models.ConcurrentCall
	cc.Tok = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		b.pusherr(cc.Tok, "missing_expr")
		return
	}
	if IsFuncCall(toks) == nil {
		b.pusherr(cc.Tok, "expr_not_func_call")
	}
	cc.Expr = b.Expr(toks)
	s.Tok = cc.Tok
	s.Data = cc
	return
}

func (b *Builder) GotoStatement(toks Toks) (s models.Statement) {
	s.Tok = toks[0]
	if len(toks) == 1 {
		b.pusherr(s.Tok, "missing_goto_label")
		return
	} else if len(toks) > 2 {
		b.pusherr(toks[2], "invalid_syntax")
	}
	idTok := toks[1]
	if idTok.Id != tokens.Id {
		b.pusherr(idTok, "invalid_syntax")
		return
	}
	var gt models.Goto
	gt.Tok = s.Tok
	gt.Label = idTok.Kind
	s.Data = gt
	return
}

// RetStatement builds AST model of return statement.
func (b *Builder) RetStatement(toks Toks) models.Statement {
	var ret models.Ret
	ret.Tok = toks[0]
	if len(toks) > 1 {
		ret.Expr = b.Expr(toks[1:])
	}
	return models.Statement{
		Tok:  ret.Tok,
		Data: ret,
	}
}

func (b *Builder) getWhileIterProfile(toks Toks) models.IterWhile {
	return models.IterWhile{
		Expr: b.Expr(toks),
	}
}

func (b *Builder) getForeachVarsToks(toks Toks) []Toks {
	vars, errs := Parts(toks, tokens.Comma, true)
	b.Errors = append(b.Errors, errs...)
	return vars
}

func (b *Builder) getVarProfile(toks Toks) (vast models.Var) {
	if len(toks) == 0 {
		return
	}
	vast.IdTok = toks[0]
	if vast.IdTok.Id != tokens.Id {
		b.pusherr(vast.IdTok, "invalid_syntax")
		return
	}
	vast.Id = vast.IdTok.Kind
	if len(toks) == 1 {
		return
	}
	if colon := toks[1]; colon.Id != tokens.Colon {
		b.pusherr(colon, "invalid_syntax")
		return
	}
	vast.New = true
	i := new(int)
	*i = 2
	if *i >= len(toks) {
		return
	}
	vast.Type, _ = b.DataType(toks, i, false, true)
	if *i < len(toks)-1 {
		b.pusherr(toks[*i], "invalid_syntax")
	}
	return
}

func (b *Builder) getForeachIterVars(varsToks []Toks) []models.Var {
	var vars []models.Var
	for _, toks := range varsToks {
		vars = append(vars, b.getVarProfile(toks))
	}
	return vars
}

func (b *Builder) getForeachIterProfile(varToks, exprToks Toks, inTok Tok) models.IterForeach {
	var foreach models.IterForeach
	foreach.InTok = inTok
	foreach.Expr = b.Expr(exprToks)
	if len(varToks) == 0 {
		foreach.KeyA.Id = xapi.Ignore
		foreach.KeyB.Id = xapi.Ignore
	} else {
		varsToks := b.getForeachVarsToks(varToks)
		if len(varsToks) == 0 {
			return foreach
		}
		if len(varsToks) > 2 {
			b.pusherr(inTok, "much_foreach_vars")
		}
		vars := b.getForeachIterVars(varsToks)
		foreach.KeyA = vars[0]
		if len(vars) > 1 {
			foreach.KeyB = vars[1]
		} else {
			foreach.KeyB.Id = xapi.Ignore
		}
	}
	return foreach
}

func (b *Builder) getForIterProfile(toks Toks, errtok Tok) models.IterProfile {
	parts, errs := Parts(toks, tokens.Comma, false)
	switch {
	case len(errs) > 0:
		b.Errors = append(b.Errors, errs...)
		return nil
	case len(parts) != 3:
		b.pusherr(errtok, "invalid_syntax")
		return nil
	}
	var fp models.IterFor
	once := parts[0]
	if len(once) > 0 {
		fp.Once = b.forStatement(once)
	}
	condition := parts[1]
	if len(condition) > 0 {
		fp.Condition = b.Expr(condition)
	}
	next := parts[2]
	if len(next) > 0 {
		fp.Next = b.forStatement(next)
	}
	return fp
}

func (b *Builder) forStatement(toks Toks) models.Statement {
	s := b.Statement(&blockStatement{toks: toks})
	switch s.Data.(type) {
	case models.ExprStatement, models.Assign:
	default:
		b.pusherr(s.Tok, "invalid_syntax")
	}
	return s
}

func (b *Builder) getIterProfile(toks Toks, errtok Tok) models.IterProfile {
	braceCount := 0
	comma := false
	for i, tok := range toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
				continue
			default:
				braceCount--
			}
		}
		if braceCount != 0 {
			continue
		}
		switch tok.Id {
		case tokens.In:
			varToks := toks[:i]
			exprToks := toks[i+1:]
			return b.getForeachIterProfile(varToks, exprToks, tok)
		case tokens.Comma:
			comma = true
		}
	}
	if comma {
		return b.getForIterProfile(toks, errtok)
	}
	return b.getWhileIterProfile(toks)
}

func (b *Builder) IterExpr(toks Toks) (s models.Statement) {
	var iter models.Iter
	iter.Tok = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		b.pusherr(iter.Tok, "body_not_exist")
		return
	}
	exprToks := BlockExpr(toks)
	if len(exprToks) > 0 {
		iter.Profile = b.getIterProfile(exprToks, iter.Tok)
	}
	i := new(int)
	*i = len(exprToks)
	blockToks := b.getrange(i, tokens.LBRACE, tokens.RBRACE, &toks)
	if blockToks == nil {
		b.pusherr(iter.Tok, "body_not_exist")
		return
	}
	if *i < len(toks) {
		b.pusherr(toks[*i], "invalid_syntax")
	}
	iter.Block = b.Block(blockToks)
	return models.Statement{
		Tok:  iter.Tok,
		Data: iter,
	}
}

func (b *Builder) caseexprs(toks *Toks, caseIsDefault bool) []models.Expr {
	var exprs []models.Expr
	pushExpr := func(toks Toks, tok Tok) {
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
	braceCount := 0
	j := 0
	var i int
	var tok Tok
	for i, tok = range *toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LPARENTHESES, tokens.LBRACE, tokens.LBRACKET:
				braceCount++
			default:
				braceCount--
			}
			continue
		} else if braceCount != 0 {
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
	b.pusherr((*toks)[0], "invalid_syntax")
	*toks = nil
	return nil
}

func (b *Builder) caseblock(toks *Toks) *models.Block {
	braceCount := 0
	for i, tok := range *toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LPARENTHESES, tokens.LBRACE, tokens.LBRACKET:
				braceCount++
			default:
				braceCount--
			}
			continue
		} else if braceCount != 0 {
			continue
		}
		switch tok.Id {
		case tokens.Case, tokens.Default:
			blockToks := (*toks)[:i]
			*toks = (*toks)[i:]
			return b.Block(blockToks)
		}
	}
	block := b.Block(*toks)
	*toks = nil
	return block
}

func (b *Builder) getcase(toks *Toks) models.Case {
	var c models.Case
	tok := (*toks)[0]
	*toks = (*toks)[1:]
	c.Exprs = b.caseexprs(toks, tok.Id == tokens.Default)
	c.Block = b.caseblock(toks)
	return c
}

func (b *Builder) cases(toks Toks) ([]models.Case, *models.Case) {
	var cases []models.Case
	var def *models.Case
	for len(toks) > 0 {
		tok := toks[0]
		switch tok.Id {
		case tokens.Case:
			cases = append(cases, b.getcase(&toks))
		case tokens.Default:
			c := b.getcase(&toks)
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
func (b *Builder) MatchCase(toks Toks) (s models.Statement) {
	var match models.Match
	match.Tok = toks[0]
	s.Tok = match.Tok
	toks = toks[1:]
	exprToks := BlockExpr(toks)
	if len(exprToks) > 0 {
		match.Expr = b.Expr(exprToks)
	}
	i := new(int)
	*i = len(exprToks)
	blockToks := b.getrange(i, tokens.LBRACE, tokens.RBRACE, &toks)
	if blockToks == nil {
		b.pusherr(match.Tok, "body_not_exist")
		return
	}
	match.Cases, match.Default = b.cases(blockToks)
	s.Data = match
	return
}

// IfExpr builds AST model of if expression.
func (b *Builder) IfExpr(bs *blockStatement) (s models.Statement) {
	var ifast models.If
	ifast.Tok = bs.toks[0]
	bs.toks = bs.toks[1:]
	exprToks := BlockExpr(bs.toks)
	i := new(int)
	if len(exprToks) == 0 {
		if len(bs.toks) == 0 || bs.pos >= len(*bs.srcToks) {
			b.pusherr(ifast.Tok, "missing_expr")
			return
		}
		exprToks = bs.toks
		*bs.srcToks = (*bs.srcToks)[bs.pos:]
		bs.pos, bs.withTerminator = NextStatementPos(*bs.srcToks, 0)
		bs.toks = (*bs.srcToks)[:bs.pos]
	} else {
		*i = len(exprToks)
	}
	blockToks := b.getrange(i, tokens.LBRACE, tokens.RBRACE, &bs.toks)
	if blockToks == nil {
		b.pusherr(ifast.Tok, "body_not_exist")
		return
	}
	if *i < len(bs.toks) {
		if bs.toks[*i].Id == tokens.Else {
			bs.nextToks = bs.toks[*i:]
		} else {
			b.pusherr(bs.toks[*i], "invalid_syntax")
		}
	}
	ifast.Expr = b.Expr(exprToks)
	ifast.Block = b.Block(blockToks)
	return models.Statement{
		Tok:  ifast.Tok,
		Data: ifast,
	}
}

// ElseIfEpxr builds AST model of else if expression.
func (b *Builder) ElseIfExpr(bs *blockStatement) (s models.Statement) {
	var elif models.ElseIf
	elif.Tok = bs.toks[1]
	bs.toks = bs.toks[2:]
	exprToks := BlockExpr(bs.toks)
	i := new(int)
	if len(exprToks) == 0 {
		if len(bs.toks) == 0 || bs.pos >= len(*bs.srcToks) {
			b.pusherr(elif.Tok, "missing_expr")
			return
		}
		exprToks = bs.toks
		*bs.srcToks = (*bs.srcToks)[bs.pos:]
		bs.pos, bs.withTerminator = NextStatementPos(*bs.srcToks, 0)
		bs.toks = (*bs.srcToks)[:bs.pos]
	} else {
		*i = len(exprToks)
	}
	blockToks := b.getrange(i, tokens.LBRACE, tokens.RBRACE, &bs.toks)
	if blockToks == nil {
		b.pusherr(elif.Tok, "body_not_exist")
		return
	}
	if *i < len(bs.toks) {
		if bs.toks[*i].Id == tokens.Else {
			bs.nextToks = bs.toks[*i:]
		} else {
			b.pusherr(bs.toks[*i], "invalid_syntax")
		}
	}
	elif.Expr = b.Expr(exprToks)
	elif.Block = b.Block(blockToks)
	return models.Statement{
		Tok:  elif.Tok,
		Data: elif,
	}
}

// ElseBlock builds AST model of else block.
func (b *Builder) ElseBlock(bs *blockStatement) (s models.Statement) {
	if len(bs.toks) > 1 && bs.toks[1].Id == tokens.If {
		return b.ElseIfExpr(bs)
	}
	var elseast models.Else
	elseast.Tok = bs.toks[0]
	bs.toks = bs.toks[1:]
	i := new(int)
	blockToks := b.getrange(i, tokens.LBRACE, tokens.RBRACE, &bs.toks)
	if blockToks == nil {
		if *i < len(bs.toks) {
			b.pusherr(elseast.Tok, "else_have_expr")
		} else {
			b.pusherr(elseast.Tok, "body_not_exist")
		}
		return
	}
	if *i < len(bs.toks) {
		b.pusherr(bs.toks[*i], "invalid_syntax")
	}
	elseast.Block = b.Block(blockToks)
	return models.Statement{
		Tok:  elseast.Tok,
		Data: elseast,
	}
}

// BreakStatement builds AST model of break statement.
func (b *Builder) BreakStatement(toks Toks) models.Statement {
	var breakAST models.Break
	breakAST.Tok = toks[0]
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
	}
	return models.Statement{
		Tok:  breakAST.Tok,
		Data: breakAST,
	}
}

// ContinueStatement builds AST model of continue statement.
func (b *Builder) ContinueStatement(toks Toks) models.Statement {
	var continueAST models.Continue
	continueAST.Tok = toks[0]
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
	}
	return models.Statement{
		Tok:  continueAST.Tok,
		Data: continueAST,
	}
}

// Expr builds AST model of expression.
func (b *Builder) Expr(toks Toks) (e models.Expr) {
	e.Processes = b.exprProcesses(toks)
	e.Toks = toks
	return
}

type exprProcessInfo struct {
	processes        []Toks
	part             Toks
	operator         bool
	value            bool
	singleOperatored bool
	pushedError      bool
	braceCount       int
	toks             Toks
	i                int
}

func (b *Builder) exprOperatorPart(info *exprProcessInfo, tok Tok) {
	if IsExpressionOperator(tok.Kind) ||
		IsAssignOperator(tok.Kind) {
		info.part = append(info.part, tok)
		return
	}
	if !info.operator {
		if IsUnaryOperator(tok.Kind) && !info.singleOperatored {
			info.part = append(info.part, tok)
			info.singleOperatored = true
			return
		}
		if info.braceCount == 0 && IsSolidOperator(tok.Kind) {
			b.pusherr(tok, "operator_overflow")
		}
	}
	info.singleOperatored = false
	info.operator = false
	info.value = true
	if info.braceCount > 0 {
		info.part = append(info.part, tok)
		return
	}
	info.processes = append(info.processes, info.part)
	info.processes = append(info.processes, Toks{tok})
	info.part = Toks{}
}

func (b *Builder) exprValuePart(info *exprProcessInfo, tok Tok) {
	if info.i > 0 && info.braceCount == 0 {
		lt := info.toks[info.i-1]
		if (lt.Id == tokens.Id || lt.Id == tokens.Value) &&
			(tok.Id == tokens.Id || tok.Id == tokens.Value) {
			b.pusherr(tok, "invalid_syntax")
			info.pushedError = true
		}
	}
	b.checkExprTok(tok)
	info.part = append(info.part, tok)
	info.operator = RequireOperatorToProcess(tok, info.i, len(info.toks))
	info.value = false
}

func (b *Builder) exprBracePart(info *exprProcessInfo, tok Tok) bool {
	switch tok.Kind {
	case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
		if tok.Kind == tokens.LBRACKET {
			oldIndex := info.i
			_, ok := b.DataType(info.toks, &info.i, false, false)
			if ok {
				info.part = append(info.part, info.toks[oldIndex:info.i+1]...)
				return true
			}
			info.i = oldIndex
		}
		info.singleOperatored = false
		info.braceCount++
	default:
		info.braceCount--
	}
	return false
}

func (b *Builder) exprProcesses(toks Toks) []Toks {
	var info exprProcessInfo
	info.toks = toks
	for ; info.i < len(info.toks); info.i++ {
		tok := info.toks[info.i]
		switch tok.Id {
		case tokens.Operator:
			b.exprOperatorPart(&info, tok)
			continue
		case tokens.Brace:
			skipStep := b.exprBracePart(&info, tok)
			if skipStep {
				continue
			}
		case tokens.Comma:
			info.singleOperatored = false
		}
		b.exprValuePart(&info, tok)
	}
	if len(info.part) > 0 {
		info.processes = append(info.processes, info.part)
	}
	if info.value {
		b.pusherr(info.processes[len(info.processes)-1][0], "operator_overflow")
		info.pushedError = true
	}
	if info.pushedError {
		return nil
	}
	return info.processes
}

func (b *Builder) checkExprTok(tok Tok) {
	if lex.NumRegexp.MatchString(tok.Kind) {
		var result bool
		if strings.Contains(tok.Kind, tokens.DOT) ||
			(!strings.HasPrefix(tok.Kind, "0x") && strings.ContainsAny(tok.Kind, "eE")) {
			result = xbits.CheckBitFloat(tok.Kind, 64)
		} else {
			result = xbits.CheckBitInt(tok.Kind, xbits.MaxInt)
			if !result {
				result = xbits.CheckBitUInt(tok.Kind, xbits.MaxInt)
			}
		}
		if !result {
			b.pusherr(tok, "invalid_numeric_range")
		}
	}
}

func (b *Builder) getrange(i *int, open, close string, toks *Toks) Toks {
	rang := Range(i, open, close, *toks)
	if rang != nil {
		return rang
	}
	if b.Ended() {
		return nil
	}
	*i = 0
	*toks = b.nextBuilderStatement()
	rang = Range(i, open, close, *toks)
	return rang
}

func (b *Builder) skipStatement(i *int, toks *Toks) Toks {
	start := *i
	*i, _ = NextStatementPos(*toks, start)
	stoks := (*toks)[start:*i]
	if stoks[len(stoks)-1].Id == tokens.SemiColon {
		if len(stoks) == 1 {
			return b.skipStatement(i, toks)
		}
		stoks = stoks[:len(stoks)-1]
	}
	return stoks
}

func (b *Builder) nextBuilderStatement() Toks {
	return b.skipStatement(&b.Pos, &b.Toks)
}
