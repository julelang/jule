package ast

import (
	"os"
	"strings"
	"sync"

	"github.com/the-xlang/xxc/lex"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xbits"
	"github.com/the-xlang/xxc/pkg/xlog"
)

// Builder is builds AST tree.
type Builder struct {
	wg  sync.WaitGroup
	pub bool

	Tree []Obj
	Errs []xlog.CompilerLog
	Toks []lex.Tok
	Pos  int
}

// NewBuilder instance.
func NewBuilder(toks []lex.Tok) *Builder {
	b := new(Builder)
	b.Toks = toks
	b.Pos = 0
	return b
}

// pusherr appends error by specified token.
func (b *Builder) pusherr(tok lex.Tok, key string, args ...interface{}) {
	b.Errs = append(b.Errs, xlog.CompilerLog{
		Type:   xlog.Err,
		Row:    tok.Row,
		Column: tok.Column,
		Path:   tok.File.Path,
		Msg:    x.GetErr(key, args...),
	})
}

// Ended reports position is at end of tokens or not.
func (ast *Builder) Ended() bool { return ast.Pos >= len(ast.Toks) }

func (b *Builder) buildNode(toks []lex.Tok) {
	tok := toks[0]
	switch tok.Id {
	case lex.Use:
		b.Use(toks)
	case lex.At:
		b.Attribute(toks)
	case lex.Id:
		b.Id(toks)
	case lex.Const, lex.Volatile:
		b.GlobalVar(toks)
	case lex.Type:
		b.Type(toks)
	case lex.Comment:
		b.Comment(toks[0])
	case lex.Preprocessor:
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
		toks := b.skipStatement()
		b.pub = toks[0].Id == lex.Pub
		if b.pub {
			if len(toks) == 1 {
				if b.Ended() {
					b.pusherr(toks[0], "invalid_syntax")
					continue
				}
				toks = b.skipStatement()
			} else {
				toks = toks[1:]
			}
		}
		b.buildNode(toks)
	}
	b.wg.Wait()
}

// Type builds AST model of type defination statement.
func (b *Builder) Type(toks []lex.Tok) {
	i := 1 // Initialize value is 1 for skip keyword.
	if i >= len(toks) {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	tok := toks[i]
	if tok.Id != lex.Id {
		b.pusherr(tok, "invalid_syntax")
	}
	i++
	if i >= len(toks) {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	destType, _ := b.DataType(toks[i:], new(int), true)
	tok = toks[1]
	t := Type{
		Pub:  b.pub,
		Tok:  tok,
		Id:   tok.Kind,
		Type: destType,
	}
	b.pub = false
	b.Tree = append(b.Tree, Obj{tok, t})
}

// Comment builds AST model of comment.
func (b *Builder) Comment(tok lex.Tok) {
	tok.Kind = strings.TrimSpace(tok.Kind[2:])
	if strings.HasPrefix(tok.Kind, "cxx:") {
		b.Tree = append(b.Tree, Obj{tok, CxxEmbed{tok.Kind[4:]}})
	} else {
		b.Tree = append(b.Tree, Obj{tok, Comment{tok.Kind}})
	}
}

// Preprocessor builds AST model of preprocessor directives.
func (b *Builder) Preprocessor(toks []lex.Tok) {
	if len(toks) == 1 {
		b.pusherr(toks[0], "invalid_syntax")
		return
	}
	var pp Preprocessor
	toks = toks[1:] // Remove directive mark
	tok := toks[0]
	if tok.Id != lex.Id {
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
func (b *Builder) Pragma(pp *Preprocessor, toks []lex.Tok) bool {
	if len(toks) == 1 {
		b.pusherr(toks[0], "missing_pragma_directive")
		return false
	}
	toks = toks[1:] // Remove pragma identifier
	tok := toks[0]
	if tok.Id != lex.Id {
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

func (b *Builder) pragmaEnofi(d *Directive, toks []lex.Tok) bool {
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
		return false
	}
	d.Command = EnofiDirective{}
	return true
}

// Id builds AST model of global id statement.
func (b *Builder) Id(toks []lex.Tok) {
	if len(toks) == 1 {
		b.pusherr(toks[0], "invalid_syntax")
		return
	}
	tok := toks[1]
	switch tok.Id {
	case lex.Colon:
		b.GlobalVar(toks)
		return
	case lex.Brace:
		switch tok.Kind {
		case "(":
			f := b.Func(toks, false)
			s := Statement{f.Tok, f, false}
			b.Tree = append(b.Tree, Obj{f.Tok, s})
			return
		}
	}
	b.pusherr(tok, "invalid_syntax")
}

// Use builds AST model of use declaration.
func (b *Builder) Use(toks []lex.Tok) {
	var use Use
	use.Tok = toks[0]
	if len(toks) < 2 {
		b.pusherr(use.Tok, "missing_use_path")
		return
	}
	use.Path = b.usePath(toks[1:])
	b.Tree = append(b.Tree, Obj{use.Tok, use})
}

func (b *Builder) usePath(toks []lex.Tok) string {
	var path strings.Builder
	path.WriteString(x.StdlibPath)
	path.WriteRune(os.PathSeparator)
	for i, tok := range toks {
		if i%2 != 0 {
			if tok.Id != lex.Dot {
				b.pusherr(tok, "invalid_syntax")
			}
			path.WriteRune(os.PathSeparator)
			continue
		}
		if tok.Id != lex.Id {
			b.pusherr(tok, "invalid_syntax")
		}
		path.WriteString(tok.Kind)
	}
	return path.String()
}

// Attribute builds AST model of attribute.
func (b *Builder) Attribute(toks []lex.Tok) {
	var a Attribute
	i := 0
	a.Tok = toks[i]
	i++
	if b.Ended() {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	a.Tag = toks[i]
	if a.Tag.Id != lex.Id ||
		a.Tok.Column+1 != a.Tag.Column {
		b.pusherr(a.Tag, "invalid_syntax")
		return
	}
	b.Tree = append(b.Tree, Obj{a.Tok, a})
}

// Func builds AST model of function.
func (b *Builder) Func(toks []lex.Tok, anonymous bool) (f Func) {
	f.Tok = toks[0]
	i := 0
	f.Pub = b.pub
	b.pub = false
	if anonymous {
		f.Id = "anonymous"
	} else {
		if f.Tok.Id != lex.Id {
			b.pusherr(f.Tok, "invalid_syntax")
		}
		f.Id = f.Tok.Kind
		i++
	}
	f.RetType.Id = x.Void
	paramToks := b.getrange(&i, "(", ")", &toks)
	if len(paramToks) > 0 {
		b.Params(&f, paramToks)
	}
	if i >= len(toks) {
		if b.Ended() {
			b.pusherr(f.Tok, "body_not_exist")
			return
		}
		i = 0
		toks = b.skipStatement()
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
			toks = b.skipStatement()
		}
		tok = toks[i]
	}
	if tok.Id != lex.Brace || tok.Kind != "{" {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	blockToks := b.getrange(&i, "{", "}", &toks)
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
func (b *Builder) GlobalVar(toks []lex.Tok) {
	if toks == nil {
		return
	}
	s := b.VarStatement(toks)
	b.Tree = append(b.Tree, Obj{s.Tok, s})
}

// Params builds AST model of function parameters.
func (b *Builder) Params(fn *Func, toks []lex.Tok) {
	last := 0
	braceCount := 0
	for i, tok := range toks {
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 || tok.Id != lex.Comma {
			continue
		}
		b.pushParam(fn, toks[last:i], tok)
		last = i + 1
	}
	if last < len(toks) {
		if last == 0 {
			b.pushParam(fn, toks[last:], toks[last])
		} else {
			b.pushParam(fn, toks[last:], toks[last-1])
		}
	}
	b.wg.Add(1)
	go b.checkParamsAsync(fn)
}

func (b *Builder) checkParamsAsync(f *Func) {
	defer func() { b.wg.Done() }()
	for _, p := range f.Params {
		if p.Type.Tok.Id == lex.NA {
			b.pusherr(p.Tok, "missing_type")
		}
	}
}

func (b *Builder) pushParam(f *Func, toks []lex.Tok, errtok lex.Tok) {
	if len(toks) == 0 {
		b.pusherr(errtok, "invalid_syntax")
		return
	}
	past := Parameter{Tok: toks[0]}
	for i, tok := range toks {
		switch tok.Id {
		case lex.Const:
			if past.Const {
				b.pusherr(tok, "already_constant")
				continue
			}
			past.Const = true
		case lex.Volatile:
			if past.Volatile {
				b.pusherr(tok, "already_volatile")
				continue
			}
			past.Volatile = true
		case lex.Operator:
			if tok.Kind != "..." {
				b.pusherr(tok, "invalid_syntax")
				continue
			}
			if past.Variadic {
				b.pusherr(tok, "already_variadic")
				continue
			}
			past.Variadic = true
		case lex.Id:
			toks = toks[i:]
			if !xapi.IsIgnoreId(tok.Kind) {
				for _, param := range f.Params {
					if param.Id == tok.Kind {
						b.pusherr(tok, "parameter_exist", tok.Kind)
						break
					}
				}
				past.Id = tok.Kind
			}
			if len(toks) > 1 {
				i := 1
				past.Type, _ = b.DataType(toks, &i, true)
				i++
				if i < len(toks) {
					b.pusherr(toks[i], "invalid_syntax")
				}
				i = len(f.Params) - 1
				for ; i >= 0; i-- {
					param := &f.Params[i]
					if param.Type.Tok.Id != lex.NA {
						break
					}
					param.Type = past.Type
				}
			}
			goto end
		default:
			if t, ok := b.DataType(toks, &i, true); ok {
				if i+1 == len(toks) {
					past.Type = t
					goto end
				}
			}
			b.pusherr(tok, "invalid_syntax")
			goto end
		}
	}
end:
	f.Params = append(f.Params, past)
}

// DataType builds AST model of data type.
func (b *Builder) DataType(toks []lex.Tok, i *int, err bool) (t DataType, ok bool) {
	first := *i
	var dtv strings.Builder
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		switch tok.Id {
		case lex.DataType:
			t.Tok = tok
			t.Id = x.TypeFromId(t.Tok.Kind)
			dtv.WriteString(t.Tok.Kind)
			ok = true
			goto ret
		case lex.Id:
			t.Tok = tok
			t.Id = x.Id
			dtv.WriteString(t.Tok.Kind)
			ok = true
			goto ret
		case lex.Operator:
			if tok.Kind == "*" {
				dtv.WriteString(tok.Kind)
				break
			}
			if err {
				b.pusherr(tok, "invalid_syntax")
			}
			return
		case lex.Brace:
			switch tok.Kind {
			case "(":
				t.Tok = tok
				t.Id = x.Func
				val, f := b.FuncDataTypeHead(toks, i)
				f.RetType, _ = b.FuncRetDataType(toks, i)
				dtv.WriteString(val)
				t.Tag = f
				ok = true
				goto ret
			case "[":
				*i++
				if *i > len(toks) {
					if err {
						b.pusherr(tok, "invalid_syntax")
					}
					return
				}
				tok = toks[*i]
				if tok.Id == lex.Brace && tok.Kind == "]" {
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
func (b *Builder) MapDataType(toks []lex.Tok, i *int, err bool) (t DataType, _ string) {
	t.Id = x.Map
	t.Tok = toks[0]
	braceCount := 0
	colon := -1
	start := *i
	var mapToks []lex.Tok
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case "(", "[", "{":
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
		if colon == -1 && tok.Id == lex.Colon {
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
func (b *Builder) FuncDataTypeHead(toks []lex.Tok, i *int) (string, Func) {
	var f Func
	var typeVal strings.Builder
	typeVal.WriteByte('(')
	brace := 1
	firstIndex := *i
	for *i++; *i < len(toks); *i++ {
		tok := toks[*i]
		typeVal.WriteString(tok.Kind)
		switch tok.Id {
		case lex.Brace:
			switch tok.Kind {
			case "{", "[", "(":
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

func (b *Builder) pushTypeToTypes(types *[]DataType, toks []lex.Tok, errTok lex.Tok) {
	if len(toks) == 0 {
		b.pusherr(errTok, "missing_expr")
		return
	}
	currentDt, _ := b.DataType(toks, new(int), false)
	*types = append(*types, currentDt)
}

// FuncRetDataType builds ret data-type of funtion.
func (b *Builder) FuncRetDataType(toks []lex.Tok, i *int) (t DataType, ok bool) {
	if *i >= len(toks) {
		return
	}
	tok := toks[*i]
	start := *i
	if tok.Id == lex.Brace && tok.Kind == "[" { // Multityped?
		t.Val += tok.Kind
		*i++
		if *i >= len(toks) {
			*i--
			goto end
		}
		if tok.Id == lex.Brace && tok.Kind == "]" {
			*i--
			goto end
		}
		var types []DataType
		braceCount := 1
		last := *i
		for ; *i < len(toks); *i++ {
			tok := toks[*i]
			t.Val += tok.Kind
			if tok.Id == lex.Brace {
				switch tok.Kind {
				case "(", "[", "{":
					braceCount++
				default:
					braceCount--
				}
			}
			if braceCount == 0 {
				if tok.Id == lex.Colon {
					*i = start
					goto end
				}
				b.pushTypeToTypes(&types, toks[last:*i], toks[last-1])
				break
			} else if braceCount > 1 {
				continue
			}
			switch tok.Id {
			case lex.Comma:
			case lex.Colon:
				*i = start
				goto end
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
end:
	return b.DataType(toks, i, false)
}

// IsSigleOperator is returns true
// if operator is unary or smilar to unary,
// returns false if not.
func IsSingleOperator(kind string) bool {
	return kind == "-" ||
		kind == "+" ||
		kind == "~" ||
		kind == "!" ||
		kind == "*" ||
		kind == "&"
}

func (b *Builder) pushStatementToBlock(bs *blockStatement) {
	if len(bs.toks) == 0 {
		return
	}
	lastTok := bs.toks[len(bs.toks)-1]
	if lastTok.Id == lex.SemiColon {
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
func IsStatement(current, prev lex.Tok) (ok bool, withTerminator bool) {
	ok = current.Id == lex.SemiColon || prev.Row < current.Row
	withTerminator = current.Id == lex.SemiColon
	return
}

func nextStatementPos(toks []lex.Tok, start int) (int, bool) {
	braceCount := 0
	i := start
	for ; i < len(toks); i++ {
		var isStatement, withTerminator bool
		tok := toks[i]
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case "{", "[", "(":
				braceCount++
				continue
			default:
				braceCount--
				if braceCount == 0 {
					if i+1 < len(toks) {
						isStatement, withTerminator = IsStatement(toks[i+1], tok)
						if isStatement {
							i++
							goto ret
						}
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
	block          *BlockAST
	blockToks      *[]lex.Tok
	toks           []lex.Tok
	nextToks       []lex.Tok
	withTerminator bool
}

// Block builds AST model of statements of code block.
func (b *Builder) Block(toks []lex.Tok) (block BlockAST) {
	for {
		if b.Pos == -1 {
			return
		}
		i, withTerminator := nextStatementPos(toks, 0)
		statementToks := toks[:i]
		bs := new(blockStatement)
		bs.block = &block
		bs.blockToks = &toks
		bs.toks = statementToks
		bs.withTerminator = withTerminator
		b.pushStatementToBlock(bs)
	next:
		if len(bs.nextToks) > 0 {
			bs.toks = bs.nextToks
			bs.nextToks = nil
			b.pushStatementToBlock(bs)
			goto next
		}
		if i >= len(toks) {
			break
		}
		toks = toks[i:]
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
	case lex.Id:
		return b.IdStatement(bs.toks)
	case lex.Const, lex.Volatile:
		return b.VarStatement(bs.toks)
	case lex.Ret:
		return b.RetStatement(bs.toks)
	case lex.Free:
		return b.FreeStatement(bs.toks)
	case lex.Iter:
		return b.IterExpr(bs.toks)
	case lex.Break:
		return b.BreakStatement(bs.toks)
	case lex.Continue:
		return b.ContinueStatement(bs.toks)
	case lex.If:
		return b.IfExpr(bs)
	case lex.Else:
		return b.ElseBlock(bs)
	case lex.Operator:
		if tok.Kind == "<" {
			return b.RetStatement(bs.toks)
		}
	case lex.Comment:
		return b.CommentStatement(bs.toks[0])
	}
	return b.ExprStatement(bs.toks)
}

type assignInfo struct {
	selectorToks []lex.Tok
	exprToks     []lex.Tok
	setter       lex.Tok
	ok           bool
	isExpr       bool
}

func (b *Builder) assignInfo(toks []lex.Tok) (info assignInfo) {
	info.ok = true
	braceCount := 0
	for i, tok := range toks {
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case "(", "[", "{":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		}
		if tok.Id == lex.Operator &&
			tok.Kind[len(tok.Kind)-1] == '=' {
			info.selectorToks = toks[:i]
			if info.selectorToks == nil {
				b.pusherr(tok, "invalid_syntax")
				info.ok = false
			}
			info.setter = tok
			if i+1 >= len(toks) {
				b.pusherr(tok, "missing_expr")
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
	if selector.Expr.Toks[0].Id == lex.Id &&
		current-last > 1 &&
		selector.Expr.Toks[1].Id == lex.Colon {
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
		if selector.Expr.Toks[0].Id == lex.Id {
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
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case "(", "[", "{":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		} else if tok.Id != lex.Comma {
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
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case "(", "[", "{":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		} else if tok.Id != lex.Comma {
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
	return id == lex.Id ||
		id == lex.Brace ||
		id == lex.Operator
}

func isAssignOperator(kind string) bool {
	return kind == "=" ||
		kind == "+=" ||
		kind == "-=" ||
		kind == "/=" ||
		kind == "*=" ||
		kind == "%=" ||
		kind == ">>=" ||
		kind == "<<=" ||
		kind == "|=" ||
		kind == "&=" ||
		kind == "^="
}

func checkAssignToks(toks []lex.Tok) bool {
	if !isAssignTok(toks[0].Id) {
		return false
	}
	braceCount := 0
	for _, tok := range toks {
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		}
		if tok.Id == lex.Operator &&
			isAssignOperator(tok.Kind) {
			return true
		}
	}
	return false
}

// AssignStatement builds AST model of assignment statement.
func (b *Builder) AssignStatement(toks []lex.Tok, isExpr bool) (s Statement, _ bool) {
	assign, ok := b.AssignExpr(toks, isExpr)
	if !ok {
		return
	}
	s.Tok = toks[0]
	s.Val = assign
	return s, true
}

// AssignExpr builds AST model of assignment expression.
func (b *Builder) AssignExpr(toks []lex.Tok, isExpr bool) (assign Assign, ok bool) {
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
func (b *Builder) IdStatement(toks []lex.Tok) (s Statement) {
	if len(toks) == 1 {
		b.pusherr(toks[0], "invalid_syntax")
		return
	}
	switch toks[1].Id {
	case lex.Colon:
		return b.VarStatement(toks)
	case lex.Brace:
		switch toks[1].Kind {
		case "(":
			return b.FuncCallStatement(toks)
		}
	}
	b.pusherr(toks[0], "invalid_syntax")
	return
}

// FuncCallStatement builds AST model of function call statement.
func (b *Builder) FuncCallStatement(toks []lex.Tok) Statement {
	return b.ExprStatement(toks)
}

// ExprStatement builds AST model of expression.
func (b *Builder) ExprStatement(toks []lex.Tok) Statement {
	block := ExprStatement{b.Expr(toks)}
	return Statement{toks[0], block, false}
}

// Args builds AST model of arguments.
func (b *Builder) Args(toks []lex.Tok) []Arg {
	var args []Arg
	last := 0
	braceCount := 0
	for i, tok := range toks {
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 || tok.Id != lex.Comma {
			continue
		}
		b.pushArg(&args, toks[last:i], tok)
		last = i + 1
	}
	if last < len(toks) {
		if last == 0 {
			b.pushArg(&args, toks[last:], toks[last])
		} else {
			b.pushArg(&args, toks[last:], toks[last-1])
		}
	}
	return args
}

func (b *Builder) pushArg(args *[]Arg, toks []lex.Tok, err lex.Tok) {
	if len(toks) == 0 {
		b.pusherr(err, "invalid_syntax")
		return
	}
	var arg Arg
	arg.Tok = toks[0]
	arg.Expr = b.Expr(toks)
	*args = append(*args, arg)
}

// VarStatement builds AST model of variable declaration statement.
func (b *Builder) VarStatement(toks []lex.Tok) (s Statement) {
	var vast Var
	vast.Pub = b.pub
	b.pub = false
	i := 0
	vast.DefTok = toks[i]
	for ; i < len(toks); i++ {
		tok := toks[i]
		if tok.Id == lex.Id {
			break
		}
		switch tok.Id {
		case lex.Const:
			if vast.Const {
				b.pusherr(tok, "invalid_constant")
				break
			}
			vast.Const = true
		case lex.Volatile:
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
	if vast.IdTok.Id != lex.Id {
		b.pusherr(vast.IdTok, "invalid_syntax")
	}
	vast.Id = vast.IdTok.Kind
	vast.Type = DataType{Id: x.Void}
	// Skip type definer operator(':')
	i++
	if vast.DefTok.File != nil {
		if toks[i].Id != lex.Colon {
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
				goto ret
			}
			tok = toks[i]
		}
		if tok.Id == lex.Operator {
			if tok.Kind != "=" {
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
		}
	}
ret:
	return Statement{vast.IdTok, vast, false}
}

func (b *Builder) CommentStatement(tok lex.Tok) (s Statement) {
	s.Tok = tok
	tok.Kind = strings.TrimSpace(tok.Kind[2:])
	if strings.HasPrefix(tok.Kind, "cxx:") {
		s.Val = CxxEmbed{tok.Kind[4:]}
	} else {
		s.Val = Comment{tok.Kind}
	}
	return
}

// RetStatement builds AST model of return statement.
func (b *Builder) RetStatement(toks []lex.Tok) Statement {
	var returnModel Ret
	returnModel.Tok = toks[0]
	if len(toks) > 1 {
		returnModel.Expr = b.Expr(toks[1:])
	}
	return Statement{returnModel.Tok, returnModel, false}
}

func (b *Builder) FreeStatement(toks []lex.Tok) Statement {
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

func blockExprToks(toks []lex.Tok) (expr []lex.Tok) {
	braceCount := 0
	for i, tok := range toks {
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case "{":
				if braceCount > 0 {
					braceCount++
					break
				}
				return toks[:i]
			case "(", "[":
				braceCount++
			default:
				braceCount--
			}
		}
	}
	return nil
}

func (b *Builder) getWhileIterProfile(toks []lex.Tok) WhileProfile {
	return WhileProfile{b.Expr(toks)}
}

func (b *Builder) pushVarsToksPart(vars *[][]lex.Tok, toks []lex.Tok, errTok lex.Tok) {
	if len(toks) == 0 {
		b.pusherr(errTok, "missing_expr")
	}
	*vars = append(*vars, toks)
}

func (b *Builder) getForeachVarsToks(toks []lex.Tok) [][]lex.Tok {
	var vars [][]lex.Tok
	braceCount := 0
	last := 0
	for i, tok := range toks {
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case "(", "[", "{":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		}
		if tok.Id == lex.Comma {
			part := toks[last:i]
			b.pushVarsToksPart(&vars, part, tok)
			last = i + 1
		}
	}
	if last < len(toks) {
		part := toks[last:]
		b.pushVarsToksPart(&vars, part, toks[last])
	}
	return vars
}

func (b *Builder) getForeachIterVars(varsToks [][]lex.Tok) []Var {
	var vars []Var
	for _, toks := range varsToks {
		var vast Var
		vast.IdTok = toks[0]
		if vast.IdTok.Id != lex.Id {
			b.pusherr(vast.IdTok, "invalid_syntax")
			vars = append(vars, vast)
			continue
		}
		vast.Id = vast.IdTok.Kind
		if len(toks) == 1 {
			vars = append(vars, vast)
			continue
		}
		if colon := toks[1]; colon.Id != lex.Colon {
			b.pusherr(colon, "invalid_syntax")
			vars = append(vars, vast)
			continue
		}
		vast.New = true
		i := new(int)
		*i = 2
		if *i >= len(toks) {
			vars = append(vars, vast)
			continue
		}
		vast.Type, _ = b.DataType(toks, i, true)
		if *i < len(toks)-1 {
			b.pusherr(toks[*i], "invalid_syntax")
		}
		vars = append(vars, vast)
	}
	return vars
}

func (b *Builder) getForeachIterProfile(varToks, exprToks []lex.Tok, inTok lex.Tok) ForeachProfile {
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

func (b *Builder) getIterProfile(toks []lex.Tok) IterProfile {
	braceCount := 0
	for i, tok := range toks {
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case "(", "[", "{":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount != 0 {
			continue
		}
		if tok.Id == lex.In {
			varToks := toks[:i]
			exprToks := toks[i+1:]
			return b.getForeachIterProfile(varToks, exprToks, tok)
		}
	}
	return b.getWhileIterProfile(toks)
}

func (b *Builder) IterExpr(toks []lex.Tok) (s Statement) {
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
	blockToks := b.getrange(i, "{", "}", &toks)
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

func (b *Builder) IfExpr(bs *blockStatement) (s Statement) {
	var ifast If
	ifast.Tok = bs.toks[0]
	bs.toks = bs.toks[1:]
	exprToks := blockExprToks(bs.toks)
	if len(exprToks) == 0 {
		b.pusherr(ifast.Tok, "missing_expr")
	}
	i := new(int)
	*i = len(exprToks)
	blockToks := b.getrange(i, "{", "}", &bs.toks)
	if blockToks == nil {
		b.pusherr(ifast.Tok, "body_not_exist")
		return
	}
	if *i < len(bs.toks) {
		if bs.toks[*i].Id == lex.Else {
			bs.nextToks = bs.toks[*i:]
		} else {
			b.pusherr(bs.toks[*i], "invalid_syntax")
		}
	}
	ifast.Expr = b.Expr(exprToks)
	ifast.Block = b.Block(blockToks)
	return Statement{ifast.Tok, ifast, false}
}

func (b *Builder) ElseIfExpr(bs *blockStatement) (s Statement) {
	var elif ElseIf
	elif.Tok = bs.toks[1]
	bs.toks = bs.toks[2:]
	exprToks := blockExprToks(bs.toks)
	if len(exprToks) == 0 {
		b.pusherr(elif.Tok, "missing_expr")
	}
	i := new(int)
	*i = len(exprToks)
	blockToks := b.getrange(i, "{", "}", &bs.toks)
	if blockToks == nil {
		b.pusherr(elif.Tok, "body_not_exist")
		return
	}
	if *i < len(bs.toks) {
		if bs.toks[*i].Id == lex.Else {
			bs.nextToks = bs.toks[*i:]
		} else {
			b.pusherr(bs.toks[*i], "invalid_syntax")
		}
	}
	elif.Expr = b.Expr(exprToks)
	elif.Block = b.Block(blockToks)
	return Statement{elif.Tok, elif, false}
}

func (b *Builder) ElseBlock(bs *blockStatement) (s Statement) {
	if len(bs.toks) > 1 && bs.toks[1].Id == lex.If {
		return b.ElseIfExpr(bs)
	}
	var elseast Else
	elseast.Tok = bs.toks[0]
	bs.toks = bs.toks[1:]
	i := new(int)
	blockToks := b.getrange(i, "{", "}", &bs.toks)
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

func (b *Builder) BreakStatement(toks []lex.Tok) Statement {
	var breakAST Break
	breakAST.Tok = toks[0]
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
	}
	return Statement{breakAST.Tok, breakAST, false}
}

func (b *Builder) ContinueStatement(toks []lex.Tok) Statement {
	var continueAST Continue
	continueAST.Tok = toks[0]
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
	}
	return Statement{continueAST.Tok, continueAST, false}
}

// Expr builds AST model of expression.
func (b *Builder) Expr(toks []lex.Tok) (e Expr) {
	e.Processes = b.getExprProcesses(toks)
	e.Toks = toks
	return
}

func isOverflowOperator(kind string) bool {
	return kind == "+" ||
		kind == "-" ||
		kind == "*" ||
		kind == "/" ||
		kind == "%" ||
		kind == "&" ||
		kind == "|" ||
		kind == "^" ||
		kind == "<" ||
		kind == ">" ||
		kind == "~" ||
		kind == "!"
}

func isExprOperator(kind string) bool { return kind == "..." }

func (b *Builder) getExprProcesses(toks []lex.Tok) [][]lex.Tok {
	var processes [][]lex.Tok
	var part []lex.Tok
	operator := false
	value := false
	braceCount := 0
	pushedError := false
	singleOperatored := false
	newKeyword := false
	for i := 0; i < len(toks); i++ {
		tok := toks[i]
		switch tok.Id {
		case lex.Operator:
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
			processes = append(processes, []lex.Tok{tok})
			part = []lex.Tok{}
			continue
		case lex.Brace:
			switch tok.Kind {
			case "(", "[", "{":
				if tok.Kind == "[" {
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
		case lex.New:
			newKeyword = true
		case lex.Id:
			if braceCount == 0 {
				newKeyword = false
			}
		}
		if i > 0 && braceCount == 0 {
			lt := toks[i-1]
			if (lt.Id == lex.Id || lt.Id == lex.Value) &&
				(tok.Id == lex.Id || tok.Id == lex.Value) {
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

func requireOperatorForProcess(tok lex.Tok, index, len int) bool {
	switch tok.Id {
	case lex.Comma:
		return false
	case lex.Brace:
		if tok.Kind == "(" ||
			tok.Kind == "{" {
			return false
		}
	}
	return index < len-1
}

func (b *Builder) checkExprTok(tok lex.Tok) {
	if tok.Kind[0] >= '0' && tok.Kind[0] <= '9' {
		var result bool
		if strings.Contains(tok.Kind, ".") ||
			strings.ContainsAny(tok.Kind, "eE") {
			result = xbits.CheckBitFloat(tok.Kind, 64)
		} else {
			result = xbits.CheckBitInt(tok.Kind, 64)
			if !result {
				result = xbits.CheckBitUInt(tok.Kind, 64)
			}
		}
		if !result {
			b.pusherr(tok, "invalid_numeric_range")
		}
	}
}

func (b *Builder) getrange(i *int, open, close string, toks *[]lex.Tok) []lex.Tok {
	rang := getrange(i, open, close, *toks)
	if rang != nil {
		return rang
	}
	if b.Ended() {
		return nil
	}
	*i = 0
	*toks = b.skipStatement()
	rang = getrange(i, open, close, *toks)
	return rang
}

func getrange(i *int, open, close string, toks []lex.Tok) []lex.Tok {
	if *i >= len(toks) {
		return nil
	}
	tok := toks[*i]
	if tok.Id == lex.Brace && tok.Kind == open {
		*i++
		braceCount := 1
		start := *i
		for ; braceCount > 0 && *i < len(toks); *i++ {
			tok := toks[*i]
			if tok.Id != lex.Brace {
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

func (b *Builder) skipStatement() []lex.Tok {
	start := b.Pos
	b.Pos, _ = nextStatementPos(b.Toks, start)
	toks := b.Toks[start:b.Pos]
	if toks[len(toks)-1].Id == lex.SemiColon {
		if len(toks) == 1 {
			return b.skipStatement()
		}
		toks = toks[:len(toks)-1]
	}
	return toks
}
