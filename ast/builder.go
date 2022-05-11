package ast

import (
	"os"
	"strings"
	"sync"

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

	Tree []Obj
	Errs []xlog.CompilerLog
	Toks Toks
	Pos  int
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
		Type:   xlog.Err,
		Row:    tok.Row,
		Column: tok.Column,
		Path:   tok.File.Path,
		Msg:    x.GetErr(key, args...),
	}
}

// pusherr appends error by specified token.
func (b *Builder) pusherr(tok Tok, key string, args ...any) {
	b.Errs = append(b.Errs, compilerErr(tok, key, args...))
}

// Parts returns parts separated by given token identifier.
// It's skips parentheses ranges.
//
// Special case is;
//  Parts(toks) = nil if len(toks) == 0
func Parts(toks Toks, id uint8) ([]Toks, []xlog.CompilerLog) {
	if len(toks) == 0 {
		return nil, nil
	}
	parts := make([]Toks, 0)
	errs := make([]xlog.CompilerLog, 0)
	braceCount := 0
	last := 0
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
		if braceCount > 0 {
			continue
		}
		if tok.Id == id {
			if i-last <= 0 {
				errs = append(errs, compilerErr(tok, "missing_expr"))
			}
			parts = append(parts, toks[last:i])
			last = i + 1
		}
	}
	if last < len(toks) {
		parts = append(parts, toks[last:])
	}
	return parts, errs
}

// Ended reports position is at end of tokens or not.
func (ast *Builder) Ended() bool { return ast.Pos >= len(ast.Toks) }

func (b *Builder) buildNode(toks Toks) {
	tok := toks[0]
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
		t := b.Type(toks)
		t.Pub = b.pub
		b.pub = false
		b.Tree = append(b.Tree, Obj{t.Tok, t})
	case tokens.Enum:
		b.Enum(toks)
	case tokens.Struct:
		b.Struct(toks)
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
	b.wg.Wait()
}

// Type builds AST model of type definition statement.
func (b *Builder) Type(toks Toks) (t Type) {
	i := 1 // Initialize value is 1 for skip keyword.
	if i >= len(toks) {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	tok := toks[i]
	if tok.Id != tokens.Id {
		b.pusherr(tok, "invalid_syntax")
	}
	i++
	if i >= len(toks) {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	destType, _ := b.DataType(toks[i:], new(int), true)
	tok = toks[1]
	return Type{
		Tok:  tok,
		Id:   tok.Kind,
		Type: destType,
	}
}

func (b *Builder) buildEnumItemExpr(i *int, toks Toks) Expr {
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
	return Expr{}
}

func (b *Builder) buildEnumItems(toks Toks) []*EnumItem {
	items := make([]*EnumItem, 0)
	for i := 0; i < len(toks); i++ {
		tok := toks[i]
		item := new(EnumItem)
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
	var enum Enum
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
		enum.Type, _ = b.DataType(toks, &i, true)
		i++
		if i >= len(toks) {
			b.pusherr(enum.Tok, "body_not_exist")
			return
		}
	} else {
		enum.Type = DataType{Id: xtype.U32, Val: tokens.U32}
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
	b.Tree = append(b.Tree, Obj{enum.Tok, enum})
}

// Comment builds AST model of comment.
func (b *Builder) Comment(tok Tok) {
	tok.Kind = strings.TrimSpace(tok.Kind[2:])
	if strings.HasPrefix(tok.Kind, "cxx:") {
		b.Tree = append(b.Tree, Obj{tok, CxxEmbed{tok.Kind[4:]}})
		return
	}
	b.Tree = append(b.Tree, Obj{tok, Comment{tok.Kind}})
}

// Preprocessor builds AST model of preprocessor directives.
func (b *Builder) Preprocessor(toks Toks) {
	if len(toks) == 1 {
		b.pusherr(toks[0], "invalid_syntax")
		return
	}
	var pp Preprocessor
	toks = toks[1:] // Remove directive mark
	tok := toks[0]
	if tok.Id != tokens.Id {
		b.pusherr(pp.Tok, "invalid_syntax")
		return
	}
	ok := false
	switch tok.Kind {
	case "pragma":
		ok = b.Pragma(&pp, toks)
	default:
		b.pusherr(tok, "invalid_preprocessor")
		return
	}
	if ok {
		b.Tree = append(b.Tree, Obj{pp.Tok, pp})
	}
}

// Pragma builds AST model of preprocessor pragma directive.
// Returns true if success, returns false if not.
func (b *Builder) Pragma(pp *Preprocessor, toks Toks) bool {
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
	var d Directive
	ok := false
	switch tok.Kind {
	case "enofi":
		ok = b.pragmaEnofi(&d, toks)
	default:
		b.pusherr(tok, "invalid_pragma_directive")
	}
	pp.Command = d
	return ok
}

func (b *Builder) pragmaEnofi(d *Directive, toks Toks) bool {
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
		return false
	}
	d.Command = EnofiDirective{}
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
		case tokens.LPARENTHESES: // Function.
			f := b.Func(toks, false)
			s := Statement{f.Tok, f, false}
			b.Tree = append(b.Tree, Obj{f.Tok, s})
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
	var ns Namespace
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
	b.Tree = append(b.Tree, Obj{ns.Tok, ns})
}

func (b *Builder) structFields(toks Toks) []*Var {
	fields := make([]*Var, 0)
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
		b.Var(varToks)
		vast := b.Var(varToks)
		vast.Pub = pub
		fields = append(fields, &vast)
	}
	return fields
}

// Struct builds AST model of structure.
func (b *Builder) Struct(toks Toks) {
	var s Struct
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
	b.Tree = append(b.Tree, Obj{s.Tok, s})
}

// Use builds AST model of use declaration.
func (b *Builder) Use(toks Toks) {
	var use Use
	use.Tok = toks[0]
	if len(toks) < 2 {
		b.pusherr(use.Tok, "missing_use_path")
		return
	}
	use.Path = b.usePath(toks[1:])
	b.Tree = append(b.Tree, Obj{use.Tok, use})
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
	var a Attribute
	i := 0
	a.Tok = toks[i]
	i++
	if b.Ended() {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	a.Tag = toks[i]
	if a.Tag.Id != tokens.Id ||
		a.Tok.Column+1 != a.Tag.Column {
		b.pusherr(a.Tag, "invalid_syntax")
		return
	}
	b.Tree = append(b.Tree, Obj{a.Tok, a})
}

// Func builds AST model of function.
func (b *Builder) Func(toks Toks, anon bool) (f Func) {
	f.Tok = toks[0]
	i := 0
	f.Pub = b.pub
	b.pub = false
	if anon {
		f.Id = x.Anonymous
	} else {
		if f.Tok.Id != tokens.Id {
			b.pusherr(f.Tok, "invalid_syntax")
		}
		f.Id = f.Tok.Kind
		i++
	}
	f.RetType.Id = xtype.Void
	f.RetType.Val = xtype.VoidTypeStr
	paramToks := b.getrange(&i, tokens.LPARENTHESES, tokens.RPARENTHESES, &toks)
	if len(paramToks) > 0 {
		b.Params(&f, paramToks)
	}
	if i >= len(toks) {
		if b.Ended() {
			b.pusherr(f.Tok, "body_not_exist")
			return
		}
		i = 0
		toks = b.nextBuilderStatement()
	}
	tok := toks[i]
	t, ok := b.FuncRetDataType(toks, &i)
	if ok {
		f.RetType = t
		i++
		if i >= len(toks) {
			if b.Ended() {
				b.pusherr(f.Tok, "body_not_exist")
				return
			}
			i = 0
			toks = b.nextBuilderStatement()
		}
		tok = toks[i]
	}
	if tok.Id != tokens.Brace || tok.Kind != tokens.LBRACE {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	blockToks := b.getrange(&i, tokens.LBRACE, tokens.RBRACE, &toks)
	if blockToks == nil {
		b.pusherr(f.Tok, "body_not_exist")
		return
	}
	if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	f.Block = b.Block(blockToks)
	return
}

// GlobalVar builds AST model of global variable.
func (b *Builder) GlobalVar(toks Toks) {
	if toks == nil {
		return
	}
	s := b.VarStatement(toks)
	b.Tree = append(b.Tree, Obj{s.Tok, s})
}

// Params builds AST model of function parameters.
func (b *Builder) Params(fn *Func, toks Toks) {
	parts, errs := Parts(toks, tokens.Comma)
	b.Errs = append(b.Errs, errs...)
	for _, part := range parts {
		if len(parts) > 0 {
			b.pushParam(fn, part)
		}

	}
	b.wg.Add(1)
	go b.checkParamsAsync(fn)
}

func (b *Builder) checkParamsAsync(f *Func) {
	defer func() { b.wg.Done() }()
	for i, p := range f.Params {
		if p.Type.Tok.Id == tokens.NA {
			if p.Tok.Id == tokens.NA {
				b.pusherr(p.Tok, "missing_type")
			} else {
				p.Type.Tok = p.Tok
				p.Type.Id = xtype.Id
				p.Type.Val = p.Type.Tok.Kind
				f.Params[i] = p
				p.Tok = lex.Tok{}
			}
		}
	}
}

func (b *Builder) pushParam(f *Func, toks Toks) {
	var p Param
	for i, tok := range toks {
		switch tok.Id {
		case tokens.Const:
			if p.Const {
				b.pusherr(tok, "already_constant")
				continue
			}
			p.Const = true
		case tokens.Volatile:
			if p.Volatile {
				b.pusherr(tok, "already_volatile")
				continue
			}
			p.Volatile = true
		case tokens.Operator:
			if tok.Kind != tokens.TRIPLE_DOT {
				b.pusherr(tok, "invalid_syntax")
				continue
			}
			if p.Variadic {
				b.pusherr(tok, "already_variadic")
				continue
			}
			p.Variadic = true
		case tokens.Id:
			toks = toks[i:]
			p.Tok = tok
			if !xapi.IsIgnoreId(tok.Kind) {
				for _, param := range f.Params {
					if param.Id == tok.Kind {
						b.pusherr(tok, "parameter_exist", tok.Kind)
						break
					}
				}
				p.Id = tok.Kind
			} else {
				p.Id = x.Anonymous
			}
			if len(toks) == 1 {
				goto end
			}
			toks = toks[1:]
			if tok := toks[0]; tok.Id == tokens.Brace && tok.Kind == tokens.LBRACE {
				braceCount := 0
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
					if braceCount > 0 {
						continue
					}
					exprToks := toks[1:i]
					toks = toks[i+1:]
					if len(exprToks) > 0 {
						p.Default = b.Expr(exprToks)
					}
					break
				}
			}
			if len(toks) > 0 {
				i := 0
				p.Type, _ = b.DataType(toks, &i, true)
				i++
				if i < len(toks) {
					b.pusherr(toks[i], "invalid_syntax")
				}
				i = len(f.Params) - 1
				for ; i >= 0; i-- {
					param := &f.Params[i]
					if param.Type.Tok.Id != tokens.NA {
						break
					}
					param.Type = p.Type
				}
			}
			goto end
		default:
			if t, ok := b.DataType(toks, &i, true); ok {
				if i+1 == len(toks) {
					p.Type = t
					goto end
				}
			}
			b.pusherr(tok, "invalid_syntax")
			goto end
		}
	}
end:
	f.Params = append(f.Params, p)
}

// DataType builds AST model of data type.
func (b *Builder) DataType(toks Toks, i *int, err bool) (t DataType, ok bool) {
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
			t.Tok = tok
			t.Id = xtype.Id
			t.OriginalId = t.Tok.Kind
			dtv.WriteString(t.Tok.Kind)
			ok = true
			goto ret
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
				val, f := b.FuncDataTypeHead(toks, i)
				f.RetType, _ = b.FuncRetDataType(toks, i)
				dtv.WriteString(val)
				t.Tag = f
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
					dtv.WriteString("[]")
					continue
				}
				*i-- // Start from bracket
				dt, val := b.MapDataType(toks, i, err)
				if val == "" {
					if err {
						b.pusherr(tok, "invalid_syntax")
					}
					return
				}
				t = dt
				dtv.WriteString(val)
				ok = true
				goto ret
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
	t.Val = dtv.String()
	return
}

// MapDataType builds map data-type.
func (b *Builder) MapDataType(toks Toks, i *int, err bool) (t DataType, _ string) {
	t.Id = xtype.Map
	t.Tok = toks[0]
	braceCount := 0
	colon := -1
	start := *i
	var mapToks Toks
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount == 0 {
			if start+1 > *i {
				return
			}
			mapToks = toks[start+1 : *i]
			break
		} else if braceCount != 1 {
			continue
		}
		if colon == -1 && tok.Id == tokens.Colon {
			colon = *i - start - 1
		}
	}
	if mapToks == nil || colon == -1 {
		return
	}
	colonTok := toks[colon]
	if colon == 0 || colon+1 >= len(mapToks) {
		b.pusherr(colonTok, "missing_expr")
		return t, " " // Space for ignore "invalid_syntax" error
	}
	keyTypeToks := mapToks[:colon]
	valTypeToks := mapToks[colon+1:]
	types := make([]DataType, 2)
	j := 0
	types[0], _ = b.DataType(keyTypeToks, &j, err)
	if j < len(keyTypeToks) && err {
		b.pusherr(keyTypeToks[j], "invalid_syntax")
	}
	j = 0
	types[1], _ = b.DataType(valTypeToks, &j, err)
	if j < len(valTypeToks) && err {
		b.pusherr(valTypeToks[j], "invalid_syntax")
	}
	t.Tag = types
	var val strings.Builder
	val.WriteByte('[')
	val.WriteString(types[0].Val)
	val.WriteByte(':')
	val.WriteString(types[1].Val)
	val.WriteByte(']')
	return t, val.String()
}

// FuncDataTypeHead builds head part of function data-type.
func (b *Builder) FuncDataTypeHead(toks Toks, i *int) (string, Func) {
	var f Func
	var typeVal strings.Builder
	typeVal.WriteByte('(')
	brace := 1
	firstIndex := *i
	for *i++; *i < len(toks); *i++ {
		tok := toks[*i]
		typeVal.WriteString(tok.Kind)
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
			*i++
			return typeVal.String(), f
		}
	}
	b.pusherr(toks[firstIndex], "invalid_type")
	return "", f
}

func (b *Builder) pushTypeToTypes(types *[]DataType, toks Toks, errTok Tok) {
	if len(toks) == 0 {
		b.pusherr(errTok, "missing_expr")
		return
	}
	currentDt, _ := b.DataType(toks, new(int), false)
	*types = append(*types, currentDt)
}

func (b *Builder) funcMultiTypeRet(toks Toks, i *int) (t DataType, ok bool) {
	start := *i
	tok := toks[*i]
	t.Val += tok.Kind
	*i++
	if *i >= len(toks) {
		*i--
		return b.DataType(toks, i, false)
	}
	tok = toks[*i]
	if tok.Id == tokens.Brace && tok.Kind == tokens.RBRACKET {
		*i--
		return b.DataType(toks, i, false)
	}
	var types []DataType
	braceCount := 1
	last := *i
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		t.Val += tok.Kind
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount == 0 {
			if tok.Id == tokens.Colon {
				*i = start
				return b.DataType(toks, i, false)
			}
			b.pushTypeToTypes(&types, toks[last:*i], toks[last-1])
			break
		} else if braceCount > 1 {
			continue
		}
		switch tok.Id {
		case tokens.Comma:
		case tokens.Colon:
			*i = start
			return b.DataType(toks, i, false)
		default:
			continue
		}
		b.pushTypeToTypes(&types, toks[last:*i], toks[*i-1])
		last = *i + 1
	}
	if len(types) > 1 {
		t.MultiTyped = true
		t.Tag = types
	} else {
		t = types[0]
	}
	ok = true
	return
}

// FuncRetDataType builds ret data-type of function.
func (b *Builder) FuncRetDataType(toks Toks, i *int) (t DataType, ok bool) {
	if *i >= len(toks) {
		return
	}
	tok := toks[*i]
	if tok.Id == tokens.Brace && tok.Kind == tokens.LBRACKET { // Multityped?
		return b.funcMultiTypeRet(toks, i)
	}
	return b.DataType(toks, i, false)
}

// IsSigleOperator is returns true
// if operator is unary or smilar to unary,
// returns false if not.
func IsSingleOperator(kind string) bool {
	return kind == tokens.MINUS ||
		kind == tokens.PLUS ||
		kind == tokens.TILDE ||
		kind == tokens.EXCLAMATION ||
		kind == tokens.STAR ||
		kind == tokens.AMPER
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
	if s.Val == nil {
		return
	}
	s.WithTerminator = bs.withTerminator
	bs.block.Tree = append(bs.block.Tree, s)
}

// IsStatement reports token is
// statement finish point or not.
func IsStatement(current, prev Tok) (ok bool, withTerminator bool) {
	ok = current.Id == tokens.SemiColon || prev.Row < current.Row
	withTerminator = current.Id == tokens.SemiColon
	return
}

func nextStatementPos(toks Toks, start int) (int, bool) {
	braceCount := 0
	i := start
	for ; i < len(toks); i++ {
		var isStatement, withTerminator bool
		tok := toks[i]
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				if braceCount == 0 && i > start {
					isStatement, withTerminator = IsStatement(tok, toks[i-1])
					if isStatement {
						goto ret
					}
				}
				braceCount++
				continue
			default:
				braceCount--
				if braceCount == 0 && i+1 < len(toks) {
					isStatement, withTerminator = IsStatement(toks[i+1], tok)
					if isStatement {
						i++
						goto ret
					}
				}
				continue
			}
		}
		if braceCount != 0 {
			continue
		}
		if i > start {
			isStatement, withTerminator = IsStatement(tok, toks[i-1])
		} else {
			isStatement, withTerminator = IsStatement(tok, tok)
		}
		if !isStatement {
			continue
		}
	ret:
		if withTerminator {
			i++
		}
		return i, withTerminator
	}
	return i, false
}

type blockStatement struct {
	pos            int
	block          *Block
	srcToks        *Toks
	blockToks      *Toks
	toks           Toks
	nextToks       Toks
	withTerminator bool
}

// Block builds AST model of statements of code block.
func (b *Builder) Block(toks Toks) (block Block) {
	bs := new(blockStatement)
	bs.block = &block
	bs.srcToks = &toks
	for {
		bs.pos, bs.withTerminator = nextStatementPos(toks, 0)
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
func (b *Builder) Statement(bs *blockStatement) (s Statement) {
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
	case tokens.Free:
		return b.FreeStatement(bs.toks)
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
	case tokens.Try:
		return b.TryBlock(bs)
	case tokens.Catch:
		return b.CatchBlock(bs)
	case tokens.Type:
		t := b.Type(bs.toks)
		s.Tok = t.Tok
		s.Val = t
		return
	case tokens.Operator:
	case tokens.Brace:
		if tok.Kind == tokens.LBRACE {
			return b.blockStatement(bs.toks)
		}
	}
	if IsFuncCall(bs.toks) != nil {
		return b.ExprStatement(bs.toks)
	}
	bs.toks = append([]Tok{{Id: tokens.Ret, Kind: tokens.RET}}, bs.toks...)
	return b.RetStatement(bs.toks)
}

func (b *Builder) blockStatement(toks Toks) Statement {
	i := new(int)
	tok := toks[0]
	toks = getrange(i, tokens.LBRACE, tokens.RBRACE, toks)
	if *i < len(toks) {
		b.pusherr(toks[*i], "invalid_syntax")
	}
	block := b.Block(toks)
	return Statement{Tok: tok, Val: block}
}

type assignInfo struct {
	selectorToks Toks
	exprToks     Toks
	setter       Tok
	ok           bool
	isExpr       bool
}

// IsFuncCall returns function expressions without call expression
// if tokens are function call, nil if not.
func IsFuncCall(toks Toks) Toks {
	if t := toks[len(toks)-1]; t.Id != tokens.Brace && t.Kind != tokens.RPARENTHESES {
		return nil
	}
	braceCount := 0
	for i := len(toks) - 1; i >= 0; i-- {
		tok := toks[i]
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.RPARENTHESES:
				braceCount++
			case tokens.LPARENTHESES:
				braceCount--
			}
			if braceCount == 0 {
				return toks[:i]
			}
		}
	}
	return nil
}

func (b *Builder) assignInfo(toks Toks) (info assignInfo) {
	info.ok = true
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
		}
		if tok.Id == tokens.Operator &&
			tok.Kind[len(tok.Kind)-1] == '=' {
			info.selectorToks = toks[:i]
			if info.selectorToks == nil {
				b.pusherr(tok, "invalid_syntax")
				info.ok = false
			}
			info.setter = tok
			if i+1 >= len(toks) {
				// b.pusherr(tok, "missing_expr")
				info.ok = false
			} else {
				info.exprToks = toks[i+1:]
			}
			return
		}
	}
	return
}

func (b *Builder) pushAssignSelector(selectors *[]AssignSelector, last, current int, info assignInfo) {
	var selector AssignSelector
	selector.Expr.Toks = info.selectorToks[last:current]
	if last-current == 0 {
		b.pusherr(info.selectorToks[current-1], "missing_expr")
		return
	}
	// Variable is new?
	if selector.Expr.Toks[0].Id == tokens.Id &&
		current-last > 1 &&
		selector.Expr.Toks[1].Id == tokens.Colon {
		if info.isExpr {
			b.pusherr(selector.Expr.Toks[0], "notallow_declares")
		}
		selector.Var.New = true
		selector.Var.IdTok = selector.Expr.Toks[0]
		selector.Var.Id = selector.Var.IdTok.Kind
		selector.Var.SetterTok = info.setter
		// Has specific data-type?
		if current-last > 2 {
			selector.Var.Type, _ = b.DataType(selector.Expr.Toks[2:], new(int), false)
		}
	} else {
		if selector.Expr.Toks[0].Id == tokens.Id {
			selector.Var.IdTok = selector.Expr.Toks[0]
			selector.Var.Id = selector.Var.IdTok.Kind
		}
		selector.Expr = b.Expr(selector.Expr.Toks)
	}
	*selectors = append(*selectors, selector)
}

func (b *Builder) assignSelectors(info assignInfo) []AssignSelector {
	var selectors []AssignSelector
	braceCount := 0
	lastIndex := 0
	for i, tok := range info.selectorToks {
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
		b.pushAssignSelector(&selectors, lastIndex, i, info)
		lastIndex = i + 1
	}
	if lastIndex < len(info.selectorToks) {
		b.pushAssignSelector(&selectors, lastIndex, len(info.selectorToks), info)
	}
	return selectors
}

func (b *Builder) pushAssignExpr(exps *[]Expr, last, current int, info assignInfo) {
	toks := info.exprToks[last:current]
	if toks == nil {
		b.pusherr(info.exprToks[current-1], "missing_expr")
		return
	}
	*exps = append(*exps, b.Expr(toks))
}

func (b *Builder) assignExprs(info assignInfo) []Expr {
	var exprs []Expr
	braceCount := 0
	lastIndex := 0
	for i, tok := range info.exprToks {
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
	if lastIndex < len(info.exprToks) {
		b.pushAssignExpr(&exprs, lastIndex, len(info.exprToks), info)
	}
	return exprs
}

func isAssignTok(id uint8) bool {
	return id == tokens.Id ||
		id == tokens.Brace ||
		id == tokens.Operator
}

func isAssignOperator(kind string) bool {
	return kind == tokens.EQUAL ||
		kind == tokens.PLUS_EQUAL ||
		kind == tokens.MINUS_EQUAL ||
		kind == tokens.SLASH_EQUAL ||
		kind == tokens.STAR_EQUAL ||
		kind == tokens.PERCENT_EQUAL ||
		kind == tokens.RSHIFT_EQUAL ||
		kind == tokens.LSHIFT_EQUAL ||
		kind == tokens.VLINE_EQUAL ||
		kind == tokens.AMPER_EQUAL ||
		kind == tokens.CARET_EQUAL
}

func checkAssignToks(toks Toks) bool {
	if len(toks) == 0 || !isAssignTok(toks[0].Id) {
		return false
	}
	braceCount := 0
	for _, tok := range toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount < 0 {
			return false
		} else if braceCount > 0 {
			continue
		}
		if tok.Id == tokens.Operator &&
			isAssignOperator(tok.Kind) {
			return true
		}
	}
	return false
}

// AssignStatement builds AST model of assignment statement.
func (b *Builder) AssignStatement(toks Toks, isExpr bool) (s Statement, _ bool) {
	assign, ok := b.AssignExpr(toks, isExpr)
	if !ok {
		return
	}
	s.Tok = toks[0]
	s.Val = assign
	return s, true
}

// AssignExpr builds AST model of assignment expression.
func (b *Builder) AssignExpr(toks Toks, isExpr bool) (assign Assign, ok bool) {
	if !checkAssignToks(toks) {
		return
	}
	info := b.assignInfo(toks)
	if !info.ok {
		return
	}
	ok = true
	info.isExpr = isExpr
	assign.IsExpr = isExpr
	assign.Setter = info.setter
	assign.SelectExprs = b.assignSelectors(info)
	if isExpr && len(assign.SelectExprs) > 1 {
		b.pusherr(assign.Setter, "notallow_multiple_assign")
	}
	assign.ValueExprs = b.assignExprs(info)
	return
}

// BuildReturnStatement builds AST model of return statement.
func (b *Builder) IdStatement(toks Toks) (s Statement, _ bool) {
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
func (b *Builder) LabelStatement(tok Tok) Statement {
	var l Label
	l.Tok = tok
	l.Label = tok.Kind
	return Statement{Tok: tok, Val: l}
}

// ExprStatement builds AST model of expression.
func (b *Builder) ExprStatement(toks Toks) Statement {
	block := ExprStatement{b.Expr(toks)}
	return Statement{toks[0], block, false}
}

// Args builds AST model of arguments.
func (b *Builder) Args(toks Toks) *Args {
	args := new(Args)
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
			b.pushArg(args, toks[last:], toks[last])
		} else {
			b.pushArg(args, toks[last:], toks[last-1])
		}
	}
	return args
}

func (b *Builder) pushArg(args *Args, toks Toks, err Tok) {
	if len(toks) == 0 {
		b.pusherr(err, "invalid_syntax")
		return
	}
	var arg Arg
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

// Var builds AST model of variable statement.
func (b *Builder) Var(toks Toks) (vast Var) {
	vast.Pub = b.pub
	b.pub = false
	i := 0
	vast.DefTok = toks[i]
	for ; i < len(toks); i++ {
		tok := toks[i]
		if tok.Id == tokens.Id {
			break
		}
		switch tok.Id {
		case tokens.Const:
			if vast.Const {
				b.pusherr(tok, "invalid_constant")
				break
			}
			vast.Const = true
		case tokens.Volatile:
			if vast.Volatile {
				b.pusherr(tok, "invalid_volatile")
				break
			}
			vast.Volatile = true
		default:
			b.pusherr(tok, "invalid_syntax")
		}
	}
	if i >= len(toks) {
		return
	}
	vast.IdTok = toks[i]
	if vast.IdTok.Id != tokens.Id {
		b.pusherr(vast.IdTok, "invalid_syntax")
	}
	vast.Id = vast.IdTok.Kind
	vast.Type.Id = xtype.Void
	vast.Type.Val = xtype.VoidTypeStr
	// Skip type definer operator(':')
	i++
	if i >= len(toks) {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	if vast.DefTok.File != nil {
		if toks[i].Id != tokens.Colon {
			b.pusherr(toks[i], "invalid_syntax")
			return
		}
		i++
	} else {
		i++
	}
	if i < len(toks) {
		tok := toks[i]
		t, ok := b.DataType(toks, &i, false)
		if ok {
			vast.Type = t
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
			vast.Val = b.Expr(valueToks)
			vast.SetterTok = tok
		} else {
			b.pusherr(tok, "invalid_syntax")
		}
	}
	return
}

// VarStatement builds AST model of variable declaration statement.
func (b *Builder) VarStatement(toks Toks) Statement {
	vast := b.Var(toks)
	return Statement{vast.IdTok, vast, false}
}

// CommentStatement builds AST model of comment statement.
func (b *Builder) CommentStatement(tok Tok) (s Statement) {
	s.Tok = tok
	tok.Kind = strings.TrimSpace(tok.Kind[2:])
	if strings.HasPrefix(tok.Kind, "cxx:") {
		s.Val = CxxEmbed{tok.Kind[4:]}
	} else {
		s.Val = Comment{tok.Kind}
	}
	return
}

// DeferStatement builds AST model of deferred call statement.
func (b *Builder) DeferStatement(toks Toks) (s Statement) {
	var d Defer
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
	s.Val = d
	return
}

func (b *Builder) ConcurrentCallStatement(toks Toks) (s Statement) {
	var cc ConcurrentCall
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
	s.Val = cc
	return
}

func (b *Builder) GotoStatement(toks Toks) (s Statement) {
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
	var gt Goto
	gt.Tok = s.Tok
	gt.Label = idTok.Kind
	s.Val = gt
	return
}

// RetStatement builds AST model of return statement.
func (b *Builder) RetStatement(toks Toks) Statement {
	var returnModel Ret
	returnModel.Tok = toks[0]
	if len(toks) > 1 {
		returnModel.Expr = b.Expr(toks[1:])
	}
	return Statement{returnModel.Tok, returnModel, false}
}

// FreeStatement builds AST model of free statement.
func (b *Builder) FreeStatement(toks Toks) Statement {
	var free Free
	free.Tok = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		b.pusherr(free.Tok, "missing_expr")
	} else {
		free.Expr = b.Expr(toks)
	}
	return Statement{free.Tok, free, false}
}

func blockExprToks(toks Toks) (expr Toks) {
	braceCount := 0
	for i, tok := range toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE:
				if braceCount > 0 {
					braceCount++
					break
				}
				return toks[:i]
			case tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
	}
	return nil
}

func (b *Builder) getWhileIterProfile(toks Toks) WhileProfile {
	return WhileProfile{b.Expr(toks)}
}

func (b *Builder) getForeachVarsToks(toks Toks) []Toks {
	vars, errs := Parts(toks, tokens.Comma)
	b.Errs = append(b.Errs, errs...)
	return vars
}

func (b *Builder) getVarProfile(toks Toks) (vast Var) {
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
	vast.Type, _ = b.DataType(toks, i, true)
	if *i < len(toks)-1 {
		b.pusherr(toks[*i], "invalid_syntax")
	}
	return
}

func (b *Builder) getForeachIterVars(varsToks []Toks) []Var {
	var vars []Var
	for _, toks := range varsToks {
		vars = append(vars, b.getVarProfile(toks))
	}
	return vars
}

func (b *Builder) getForeachIterProfile(varToks, exprToks Toks, inTok Tok) ForeachProfile {
	var profile ForeachProfile
	profile.InTok = inTok
	profile.Expr = b.Expr(exprToks)
	if len(varToks) == 0 {
		profile.KeyA.Id = xapi.Ignore
		profile.KeyB.Id = xapi.Ignore
	} else {
		varsToks := b.getForeachVarsToks(varToks)
		if len(varsToks) == 0 {
			return profile
		}
		if len(varsToks) > 2 {
			b.pusherr(inTok, "much_foreach_vars")
		}
		vars := b.getForeachIterVars(varsToks)
		profile.KeyA = vars[0]
		if len(vars) > 1 {
			profile.KeyB = vars[1]
		} else {
			profile.KeyB.Id = xapi.Ignore
		}
	}
	return profile
}

func (b *Builder) getIterProfile(toks Toks) IterProfile {
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
		if braceCount != 0 {
			continue
		}
		if tok.Id == tokens.In {
			varToks := toks[:i]
			exprToks := toks[i+1:]
			return b.getForeachIterProfile(varToks, exprToks, tok)
		}
	}
	return b.getWhileIterProfile(toks)
}

func (b *Builder) IterExpr(toks Toks) (s Statement) {
	var iter Iter
	iter.Tok = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		b.pusherr(iter.Tok, "body_not_exist")
		return
	}
	exprToks := blockExprToks(toks)
	if len(exprToks) > 0 {
		iter.Profile = b.getIterProfile(exprToks)
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
	return Statement{iter.Tok, iter, false}
}

// TryBlock build try block.
func (b *Builder) TryBlock(bs *blockStatement) (s Statement) {
	var try Try
	try.Tok = bs.toks[0]
	bs.toks = bs.toks[1:]
	i := new(int)
	blockToks := b.getrange(i, tokens.LBRACE, tokens.RBRACE, &bs.toks)
	if blockToks == nil {
		b.pusherr(try.Tok, "body_not_exist")
		return
	}
	if *i < len(bs.toks) {
		if bs.toks[*i].Id == tokens.Catch {
			bs.nextToks = bs.toks[*i:]
		} else {
			b.pusherr(bs.toks[*i], "invalid_syntax")
		}
	}
	try.Block = b.Block(blockToks)
	return Statement{try.Tok, try, false}
}

// CatchBlock build catch block.
func (b *Builder) CatchBlock(bs *blockStatement) (s Statement) {
	var catch Catch
	catch.Tok = bs.toks[0]
	bs.toks = bs.toks[1:]
	varToks := blockExprToks(bs.toks)
	i := new(int)
	*i = len(varToks)
	blockToks := b.getrange(i, tokens.LBRACE, tokens.RBRACE, &bs.toks)
	if blockToks == nil {
		b.pusherr(catch.Tok, "body_not_exist")
		return
	}
	if *i < len(bs.toks) {
		if bs.toks[*i].Id == tokens.Catch {
			bs.nextToks = bs.toks[*i:]
		} else {
			b.pusherr(bs.toks[*i], "invalid_syntax")
		}
	}
	if len(varToks) > 0 {
		catch.Var = b.getVarProfile(varToks)
	}
	catch.Block = b.Block(blockToks)
	return Statement{catch.Tok, catch, false}
}

// IfExpr builds AST model of if expression.
func (b *Builder) IfExpr(bs *blockStatement) (s Statement) {
	var ifast If
	ifast.Tok = bs.toks[0]
	bs.toks = bs.toks[1:]
	exprToks := blockExprToks(bs.toks)
	i := new(int)
	if len(exprToks) == 0 {
		if len(bs.toks) == 0 || bs.pos >= len(*bs.srcToks) {
			b.pusherr(ifast.Tok, "missing_expr")
			return
		}
		exprToks = bs.toks
		*bs.srcToks = (*bs.srcToks)[bs.pos:]
		bs.pos, bs.withTerminator = nextStatementPos(*bs.srcToks, 0)
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
	return Statement{ifast.Tok, ifast, false}
}

// ElseIfEpxr builds AST model of else if expression.
func (b *Builder) ElseIfExpr(bs *blockStatement) (s Statement) {
	var elif ElseIf
	elif.Tok = bs.toks[1]
	bs.toks = bs.toks[2:]
	exprToks := blockExprToks(bs.toks)
	i := new(int)
	if len(exprToks) == 0 {
		if len(bs.toks) == 0 || bs.pos >= len(*bs.srcToks) {
			b.pusherr(elif.Tok, "missing_expr")
			return
		}
		exprToks = bs.toks
		*bs.srcToks = (*bs.srcToks)[bs.pos:]
		bs.pos, bs.withTerminator = nextStatementPos(*bs.srcToks, 0)
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
	return Statement{elif.Tok, elif, false}
}

// ElseBlock builds AST model of else block.
func (b *Builder) ElseBlock(bs *blockStatement) (s Statement) {
	if len(bs.toks) > 1 && bs.toks[1].Id == tokens.If {
		return b.ElseIfExpr(bs)
	}
	var elseast Else
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
	return Statement{elseast.Tok, elseast, false}
}

// BreakStatement builds AST model of break statement.
func (b *Builder) BreakStatement(toks Toks) Statement {
	var breakAST Break
	breakAST.Tok = toks[0]
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
	}
	return Statement{breakAST.Tok, breakAST, false}
}

// ContinueStatement builds AST model of continue statement.
func (b *Builder) ContinueStatement(toks Toks) Statement {
	var continueAST Continue
	continueAST.Tok = toks[0]
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
	}
	return Statement{continueAST.Tok, continueAST, false}
}

// Expr builds AST model of expression.
func (b *Builder) Expr(toks Toks) (e Expr) {
	e.Processes = b.getExprProcesses(toks)
	e.Toks = toks
	return
}

func isOverflowOperator(kind string) bool {
	return kind == tokens.PLUS ||
		kind == tokens.MINUS ||
		kind == tokens.STAR ||
		kind == tokens.SLASH ||
		kind == tokens.PERCENT ||
		kind == tokens.AMPER ||
		kind == tokens.VLINE ||
		kind == tokens.CARET ||
		kind == tokens.LESS ||
		kind == tokens.GREAT ||
		kind == tokens.TILDE ||
		kind == tokens.EXCLAMATION
}

func isExprOperator(kind string) bool { return kind == tokens.TRIPLE_DOT }

func (b *Builder) getExprProcesses(toks Toks) []Toks {
	var processes []Toks
	var part Toks
	operator := false
	value := false
	braceCount := 0
	pushedError := false
	singleOperatored := false
	newKeyword := false
	for i := 0; i < len(toks); i++ {
		tok := toks[i]
		switch tok.Id {
		case tokens.Operator:
			if newKeyword ||
				isExprOperator(tok.Kind) ||
				isAssignOperator(tok.Kind) {
				part = append(part, tok)
				continue
			}
			if !operator {
				if IsSingleOperator(tok.Kind) && !singleOperatored {
					part = append(part, tok)
					singleOperatored = true
					continue
				}
				if braceCount == 0 && isOverflowOperator(tok.Kind) {
					b.pusherr(tok, "operator_overflow")
				}
			}
			singleOperatored = false
			operator = false
			value = true
			if braceCount > 0 {
				part = append(part, tok)
				continue
			}
			processes = append(processes, part)
			processes = append(processes, Toks{tok})
			part = Toks{}
			continue
		case tokens.Brace:
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				if tok.Kind == tokens.LBRACKET {
					oldIndex := i
					_, ok := b.DataType(toks, &i, false)
					if ok {
						part = append(part, toks[oldIndex:i+1]...)
						continue
					}
					i = oldIndex
				}
				singleOperatored = false
				braceCount++
			default:
				braceCount--
			}
		case tokens.New:
			newKeyword = true
		case tokens.Id:
			if braceCount == 0 {
				newKeyword = false
			}
		}
		if i > 0 && braceCount == 0 {
			lt := toks[i-1]
			if (lt.Id == tokens.Id || lt.Id == tokens.Value) &&
				(tok.Id == tokens.Id || tok.Id == tokens.Value) {
				b.pusherr(tok, "invalid_syntax")
				pushedError = true
			}
		}
		b.checkExprTok(tok)
		part = append(part, tok)
		operator = requireOperatorForProcess(tok, i, len(toks))
		value = false
	}
	if len(part) > 0 {
		processes = append(processes, part)
	}
	if value {
		b.pusherr(processes[len(processes)-1][0], "operator_overflow")
		pushedError = true
	}
	if pushedError {
		return nil
	}
	return processes
}

func requireOperatorForProcess(tok Tok, index, len int) bool {
	switch tok.Id {
	case tokens.Comma:
		return false
	case tokens.Brace:
		if tok.Kind == tokens.LPARENTHESES ||
			tok.Kind == tokens.LBRACE {
			return false
		}
	}
	return index < len-1
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
	rang := getrange(i, open, close, *toks)
	if rang != nil {
		return rang
	}
	if b.Ended() {
		return nil
	}
	*i = 0
	*toks = b.nextBuilderStatement()
	rang = getrange(i, open, close, *toks)
	return rang
}

func getrange(i *int, open, close string, toks Toks) Toks {
	if *i >= len(toks) {
		return nil
	}
	tok := toks[*i]
	if tok.Id == tokens.Brace && tok.Kind == open {
		*i++
		braceCount := 1
		start := *i
		for ; braceCount > 0 && *i < len(toks); *i++ {
			tok := toks[*i]
			if tok.Id != tokens.Brace {
				continue
			}
			if tok.Kind == open {
				braceCount++
			} else if tok.Kind == close {
				braceCount--
			}
		}
		return toks[start : *i-1]
	}
	return nil
}

func (b *Builder) skipStatement(i *int, toks *Toks) Toks {
	start := *i
	*i, _ = nextStatementPos(*toks, start)
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
