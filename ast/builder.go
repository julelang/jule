package ast

import (
	"os"
	"strings"
	"sync"

	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
	"github.com/the-xlang/x/pkg/xapi"
	"github.com/the-xlang/x/pkg/xbits"
	"github.com/the-xlang/x/pkg/xlog"
)

// Builder is builds AST tree.
type Builder struct {
	wg  sync.WaitGroup
	pub bool

	Tree   []Obj
	Errors []xlog.CompilerLog
	Tokens []lex.Token
	Pos    int
}

// NewBuilder instance.
func NewBuilder(toks []lex.Token) *Builder {
	ast := new(Builder)
	ast.Tokens = toks
	ast.Pos = 0
	return ast
}

// pusherr appends error by specified token.
func (b *Builder) pusherr(tok lex.Token, err string) {
	message := x.Errors[err]
	b.Errors = append(b.Errors, xlog.CompilerLog{
		Type:    xlog.Error,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    tok.File.Path,
		Message: message,
	})
}

// Ended reports position is at end of tokens or not.
func (ast *Builder) Ended() bool {
	return ast.Pos >= len(ast.Tokens)
}

func (b *Builder) buildNode(toks []lex.Token) {
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
				b.pusherr(toks[0], "invalid_syntax")
				continue
			}
			toks = toks[1:]
		}
		b.buildNode(toks)
	}
	b.wg.Wait()
}

// Type builds AST model of type defination statement.
func (b *Builder) Type(toks []lex.Token) {
	pos := 1 // Initialize value is 1 for skip keyword.
	if pos >= len(toks) {
		b.pusherr(toks[pos-1], "invalid_syntax")
		return
	}
	tok := toks[pos]
	if tok.Id != lex.Id {
		b.pusherr(tok, "invalid_syntax")
	}
	pos++
	if pos >= len(toks) {
		b.pusherr(toks[pos-1], "invalid_syntax")
		return
	}
	destType, _ := b.DataType(toks[pos:], new(int), true)
	tok = toks[1]
	typeAST := Type{
		Pub:   b.pub,
		Token: tok,
		Id:    tok.Kind,
		Type:  destType,
	}
	b.pub = false
	b.Tree = append(b.Tree, Obj{tok, typeAST})
}

// Comment builds AST model of comment.
func (b *Builder) Comment(tok lex.Token) {
	tok.Kind = strings.TrimSpace(tok.Kind[2:])
	if strings.HasPrefix(tok.Kind, "cxx:") {
		b.Tree = append(b.Tree, Obj{tok, CxxEmbed{tok.Kind[4:]}})
	} else {
		b.Tree = append(b.Tree, Obj{tok, Comment{tok.Kind}})
	}
}

// Preprocessor builds AST model of preprocessor directives.
func (b *Builder) Preprocessor(toks []lex.Token) {
	if len(toks) == 1 {
		b.pusherr(toks[0], "invalid_syntax")
		return
	}
	var pp Preprocessor
	toks = toks[1:] // Remove directive mark
	tok := toks[0]
	if tok.Id != lex.Id {
		b.pusherr(pp.Token, "invalid_syntax")
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
		b.Tree = append(b.Tree, Obj{pp.Token, pp})
	}
}

// Pragma builds AST model of preprocessor pragma directive.
// Returns true if success, returns false if not.
func (b *Builder) Pragma(pp *Preprocessor, toks []lex.Token) bool {
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

func (b *Builder) pragmaEnofi(d *Directive, toks []lex.Token) bool {
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
		return false
	}
	d.Command = EnofiDirective{}
	return true
}

// Id builds AST model of global id statement.
func (b *Builder) Id(toks []lex.Token) {
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
			funAST := b.Func(toks, false)
			statement := Statement{funAST.Token, funAST, false}
			b.Tree = append(b.Tree, Obj{funAST.Token, statement})
			return
		}
	}
	b.pusherr(tok, "invalid_syntax")
}

// Use builds AST model of use declaration.
func (b *Builder) Use(toks []lex.Token) {
	var use Use
	use.Token = toks[0]
	if len(toks) < 2 {
		b.pusherr(use.Token, "missing_use_path")
		return
	}
	use.Path = b.usePath(toks[1:])
	b.Tree = append(b.Tree, Obj{use.Token, use})
}

func (b *Builder) usePath(toks []lex.Token) string {
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
func (b *Builder) Attribute(toks []lex.Token) {
	var attribute Attribute
	i := 0
	attribute.Token = toks[i]
	i++
	if b.Ended() {
		b.pusherr(toks[i-1], "invalid_syntax")
		return
	}
	attribute.Tag = toks[i]
	if attribute.Tag.Id != lex.Id ||
		attribute.Token.Column+1 != attribute.Tag.Column {
		b.pusherr(attribute.Tag, "invalid_syntax")
		return
	}
	b.Tree = append(b.Tree, Obj{attribute.Token, attribute})
}

// Func builds AST model of function.
func (b *Builder) Func(toks []lex.Token, anonymous bool) (f Func) {
	f.Token = toks[0]
	i := 0
	f.Pub = b.pub
	b.pub = false
	if anonymous {
		f.Id = "anonymous"
	} else {
		if f.Token.Id != lex.Id {
			b.pusherr(f.Token, "invalid_syntax")
		}
		f.Id = f.Token.Kind
		i++
	}
	f.RetType.Code = x.Void
	paramToks := getRange(&i, "(", ")", toks)
	if len(paramToks) > 0 {
		b.Params(&f, paramToks)
	}
	if i >= len(toks) {
		b.pusherr(f.Token, "body_not_exist")
		return
	}
	tok := toks[i]
	t, ok := b.FuncRetDataType(toks, &i)
	if ok {
		f.RetType = t
		i++
		if i >= len(toks) {
			b.pusherr(f.Token, "body_not_exist")
			return
		}
		tok = toks[i]
	}
	if tok.Id != lex.Brace || tok.Kind != "{" {
		b.pusherr(tok, "invalid_syntax")
		return
	}
	blockToks := getRange(&i, "{", "}", toks)
	if blockToks == nil {
		b.pusherr(f.Token, "body_not_exist")
		return
	}
	if i < len(toks) {
		b.pusherr(toks[i], "invalid_syntax")
	}
	f.Block = b.Block(blockToks)
	return
}

// GlobalVar builds AST model of global variable.
func (b *Builder) GlobalVar(toks []lex.Token) {
	if toks == nil {
		return
	}
	statement := b.VarStatement(toks)
	b.Tree = append(b.Tree, Obj{statement.Token, statement})
}

// Params builds AST model of function parameters.
func (b *Builder) Params(fn *Func, toks []lex.Token) {
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
		if p.Type.Token.Id == lex.NA {
			b.pusherr(p.Token, "missing_type")
		}
	}
}

func (b *Builder) pushParam(f *Func, toks []lex.Token, errtok lex.Token) {
	if len(toks) == 0 {
		b.pusherr(errtok, "invalid_syntax")
		return
	}
	past := Parameter{Token: toks[0]}
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
						b.pusherr(tok, "parameter_exist")
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
					if param.Type.Token.Id != lex.NA {
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
func (b *Builder) DataType(toks []lex.Token, i *int, err bool) (dt DataType, ok bool) {
	first := *i
	var dtv strings.Builder
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		switch tok.Id {
		case lex.DataType:
			dt.Token = tok
			dt.Code = x.TypeFromId(dt.Token.Kind)
			dtv.WriteString(dt.Token.Kind)
			ok = true
			goto ret
		case lex.Id:
			dt.Token = tok
			dt.Code = x.Id
			dtv.WriteString(dt.Token.Kind)
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
				dt.Token = tok
				dt.Code = x.Func
				value, f := b.FuncDataTypeHead(toks, i)
				f.RetType, _ = b.FuncRetDataType(toks, i)
				dtv.WriteString(value)
				dt.Tag = f
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
				if tok.Id != lex.Brace || tok.Kind != "]" {
					if err {
						b.pusherr(tok, "invalid_syntax")
					}
					return
				}
				dtv.WriteString("[]")
				continue
			}
			/*if err {
				ast.PushErrorToken(token, "invalid_syntax")
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
	dt.Value = dtv.String()
	return
}

func (b *Builder) FuncDataTypeHead(toks []lex.Token, i *int) (string, Func) {
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

func (b *Builder) pushTypeToTypes(types *[]DataType, toks []lex.Token, errTok lex.Token) {
	if len(toks) == 0 {
		b.pusherr(errTok, "missing_value")
		return
	}
	currentDt, _ := b.DataType(toks, new(int), false)
	*types = append(*types, currentDt)
}

func (b *Builder) FuncRetDataType(toks []lex.Token, i *int) (dt DataType, ok bool) {
	if *i >= len(toks) {
		return
	}
	tok := toks[*i]
	if tok.Id == lex.Brace && tok.Kind == "[" { // Multityped?
		dt.Value += tok.Kind
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
			dt.Value += tok.Kind
			if tok.Id == lex.Brace {
				switch tok.Kind {
				case "(", "[", "{":
					braceCount++
				default:
					braceCount--
				}
			}
			if braceCount == 0 {
				b.pushTypeToTypes(&types, toks[last:*i], toks[last-1])
				break
			} else if braceCount > 1 {
				continue
			}
			if tok.Id != lex.Comma {
				continue
			}
			b.pushTypeToTypes(&types, toks[last:*i], toks[*i-1])
			last = *i + 1
		}
		if len(types) > 1 {
			dt.MultiTyped = true
			dt.Tag = types
		} else {
			dt = types[0]
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
	if s.Value == nil {
		return
	}
	s.WithTerminator = bs.withTerminator
	bs.block.Tree = append(bs.block.Tree, s)
}

// IsStatement reports token is
// statement finish point or not.
func IsStatement(current, prev lex.Token) (ok bool, withTerminator bool) {
	ok = current.Id == lex.SemiColon || prev.Row < current.Row
	withTerminator = current.Id == lex.SemiColon
	return
}

func nextStatementPos(toks []lex.Token, start int) (int, bool) {
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
	blockToks      *[]lex.Token
	toks           []lex.Token
	nextToks       []lex.Token
	withTerminator bool
}

// Block builds AST model of statements of code block.
func (b *Builder) Block(toks []lex.Token) (block BlockAST) {
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
	selectorToks []lex.Token
	exprToks     []lex.Token
	setter       lex.Token
	ok           bool
	isExpr       bool
}

func (b *Builder) assignInfo(toks []lex.Token) (info assignInfo) {
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
				b.pusherr(tok, "missing_value")
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
	selector.Expr.Tokens = info.selectorToks[last:current]
	if last-current == 0 {
		b.pusherr(info.selectorToks[current-1], "missing_value")
		return
	}
	// Variable is new?
	if selector.Expr.Tokens[0].Id == lex.Id &&
		current-last > 1 &&
		selector.Expr.Tokens[1].Id == lex.Colon {
		if info.isExpr {
			b.pusherr(selector.Expr.Tokens[0], "notallow_declares")
		}
		selector.Var.New = true
		selector.Var.IdToken = selector.Expr.Tokens[0]
		selector.Var.Id = selector.Var.IdToken.Kind
		selector.Var.SetterToken = info.setter
		// Has specific data-type?
		if current-last > 2 {
			selector.Var.Type, _ = b.DataType(selector.Expr.Tokens[2:], new(int), false)
		}
	} else {
		if selector.Expr.Tokens[0].Id == lex.Id {
			selector.Var.IdToken = selector.Expr.Tokens[0]
			selector.Var.Id = selector.Var.IdToken.Kind
		}
		selector.Expr = b.Expr(selector.Expr.Tokens)
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
		b.pusherr(info.exprToks[current-1], "missing_value")
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

func checkAssignToks(toks []lex.Token) bool {
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
func (b *Builder) AssignStatement(toks []lex.Token, isExpr bool) (s Statement, _ bool) {
	assign, ok := b.AssignExpr(toks, isExpr)
	if !ok {
		return
	}
	s.Token = toks[0]
	s.Value = assign
	return s, true
}

// AssignExpr builds AST model of assignment expression.
func (b *Builder) AssignExpr(toks []lex.Token, isExpr bool) (assign Assign, ok bool) {
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
func (b *Builder) IdStatement(toks []lex.Token) (s Statement) {
	if len(toks) == 1 {
		b.pusherr(toks[0], "invalid_syntax")
		return
	}
	switch toks[1].Id {
	case lex.Colon:
		return b.VarStatement(toks)
	}
	b.pusherr(toks[0], "invalid_syntax")
	return
}

// ExprStatement builds AST model of expression.
func (b *Builder) ExprStatement(toks []lex.Token) Statement {
	block := ExprStatement{b.Expr(toks)}
	return Statement{toks[0], block, false}
}

// Args builds AST model of arguments.
func (b *Builder) Args(toks []lex.Token) []Arg {
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

func (b *Builder) pushArg(args *[]Arg, toks []lex.Token, err lex.Token) {
	if len(toks) == 0 {
		b.pusherr(err, "invalid_syntax")
		return
	}
	var arg Arg
	arg.Token = toks[0]
	arg.Expr = b.Expr(toks)
	*args = append(*args, arg)
}

// VarStatement builds AST model of variable declaration statement.
func (b *Builder) VarStatement(toks []lex.Token) (s Statement) {
	var vast Var
	vast.Pub = b.pub
	b.pub = false
	i := 0
	vast.DefToken = toks[i]
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
	vast.IdToken = toks[i]
	if vast.IdToken.Id != lex.Id {
		b.pusherr(vast.IdToken, "invalid_syntax")
	}
	vast.Id = vast.IdToken.Kind
	vast.Type = DataType{Code: x.Void}
	// Skip type definer operator(':')
	i++
	if vast.DefToken.File != nil {
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
				b.pusherr(tok, "missing_value")
				return
			}
			vast.Value = b.Expr(valueToks)
			vast.SetterToken = tok
		}
	}
ret:
	return Statement{vast.IdToken, vast, false}
}

func (b *Builder) CommentStatement(tok lex.Token) (s Statement) {
	s.Token = tok
	tok.Kind = strings.TrimSpace(tok.Kind[2:])
	if strings.HasPrefix(tok.Kind, "cxx:") {
		s.Value = CxxEmbed{tok.Kind[4:]}
	} else {
		s.Value = Comment{tok.Kind}
	}
	return
}

// RetStatement builds AST model of return statement.
func (b *Builder) RetStatement(toks []lex.Token) Statement {
	var returnModel Ret
	returnModel.Token = toks[0]
	if len(toks) > 1 {
		returnModel.Expr = b.Expr(toks[1:])
	}
	return Statement{returnModel.Token, returnModel, false}
}

func (b *Builder) FreeStatement(toks []lex.Token) Statement {
	var free Free
	free.Token = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		b.pusherr(free.Token, "missing_expression")
	} else {
		free.Expr = b.Expr(toks)
	}
	return Statement{free.Token, free, false}
}

func blockExprToks(toks []lex.Token) (expr []lex.Token) {
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

func (b *Builder) getWhileIterProfile(toks []lex.Token) WhileProfile {
	return WhileProfile{b.Expr(toks)}
}

func (b *Builder) pushVarsToksPart(vars *[][]lex.Token, toks []lex.Token, errTok lex.Token) {
	if len(toks) == 0 {
		b.pusherr(errTok, "missing_value")
	}
	*vars = append(*vars, toks)
}

func (b *Builder) getForeachVarsToks(toks []lex.Token) [][]lex.Token {
	var vars [][]lex.Token
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

func (b *Builder) getForeachIterVars(varsToks [][]lex.Token) []Var {
	var vars []Var
	for _, toks := range varsToks {
		var vast Var
		vast.IdToken = toks[0]
		if vast.IdToken.Id != lex.Id {
			b.pusherr(vast.IdToken, "invalid_syntax")
			vars = append(vars, vast)
			continue
		}
		vast.Id = vast.IdToken.Kind
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

func (b *Builder) getForeachIterProfile(varToks, exprToks []lex.Token, inTok lex.Token) ForeachProfile {
	var profile ForeachProfile
	profile.InToken = inTok
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

func (b *Builder) getIterProfile(toks []lex.Token) IterProfile {
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

func (b *Builder) IterExpr(toks []lex.Token) (s Statement) {
	var iter Iter
	iter.Token = toks[0]
	toks = toks[1:]
	if len(toks) == 0 {
		b.pusherr(iter.Token, "body_not_exist")
		return
	}
	exprToks := blockExprToks(toks)
	if len(exprToks) > 0 {
		iter.Profile = b.getIterProfile(exprToks)
	}
	index := new(int)
	*index = len(exprToks)
	blockToks := getRange(index, "{", "}", toks)
	if blockToks == nil {
		b.pusherr(iter.Token, "body_not_exist")
		return
	}
	if *index < len(toks) {
		b.pusherr(toks[*index], "invalid_syntax")
	}
	iter.Block = b.Block(blockToks)
	return Statement{iter.Token, iter, false}
}

func (b *Builder) IfExpr(bs *blockStatement) (s Statement) {
	var ifast If
	ifast.Token = bs.toks[0]
	bs.toks = bs.toks[1:]
	exprToks := blockExprToks(bs.toks)
	if len(exprToks) == 0 {
		b.pusherr(ifast.Token, "missing_expression")
	}
	i := new(int)
	*i = len(exprToks)
	blockToks := getRange(i, "{", "}", bs.toks)
	if blockToks == nil {
		b.pusherr(ifast.Token, "body_not_exist")
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
	return Statement{ifast.Token, ifast, false}
}

func (b *Builder) ElseIfExpr(bs *blockStatement) (s Statement) {
	var elif ElseIf
	elif.Token = bs.toks[1]
	bs.toks = bs.toks[2:]
	exprToks := blockExprToks(bs.toks)
	if len(exprToks) == 0 {
		b.pusherr(elif.Token, "missing_expression")
	}
	i := new(int)
	*i = len(exprToks)
	blockToks := getRange(i, "{", "}", bs.toks)
	if blockToks == nil {
		b.pusherr(elif.Token, "body_not_exist")
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
	return Statement{elif.Token, elif, false}
}

func (b *Builder) ElseBlock(bs *blockStatement) (s Statement) {
	if len(bs.toks) > 1 && bs.toks[1].Id == lex.If {
		return b.ElseIfExpr(bs)
	}
	var elseast Else
	elseast.Token = bs.toks[0]
	bs.toks = bs.toks[1:]
	i := new(int)
	blockToks := getRange(i, "{", "}", bs.toks)
	if blockToks == nil {
		if *i < len(bs.toks) {
			b.pusherr(elseast.Token, "else_have_expr")
		} else {
			b.pusherr(elseast.Token, "body_not_exist")
		}
		return
	}
	if *i < len(bs.toks) {
		b.pusherr(bs.toks[*i], "invalid_syntax")
	}
	elseast.Block = b.Block(blockToks)
	return Statement{elseast.Token, elseast, false}
}

func (b *Builder) BreakStatement(toks []lex.Token) Statement {
	var breakAST Break
	breakAST.Token = toks[0]
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
	}
	return Statement{breakAST.Token, breakAST, false}
}

func (b *Builder) ContinueStatement(toks []lex.Token) Statement {
	var continueAST Continue
	continueAST.Token = toks[0]
	if len(toks) > 1 {
		b.pusherr(toks[1], "invalid_syntax")
	}
	return Statement{continueAST.Token, continueAST, false}
}

// Expr builds AST model of expression.
func (b *Builder) Expr(toks []lex.Token) (e Expr) {
	e.Processes = b.getExprProcesses(toks)
	e.Tokens = toks
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

func (b *Builder) getExprProcesses(toks []lex.Token) [][]lex.Token {
	var processes [][]lex.Token
	var part []lex.Token
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
			processes = append(processes, []lex.Token{tok})
			part = []lex.Token{}
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

func requireOperatorForProcess(tok lex.Token, index, len int) bool {
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

func (b *Builder) checkExprTok(tok lex.Token) {
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

func getRange(i *int, open, close string, toks []lex.Token) []lex.Token {
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

func (b *Builder) skipStatement() []lex.Token {
	start := b.Pos
	b.Pos, _ = nextStatementPos(b.Tokens, start)
	toks := b.Tokens[start:b.Pos]
	if toks[len(toks)-1].Id == lex.SemiColon {
		if len(toks) == 1 {
			return b.skipStatement()
		}
		toks = toks[:len(toks)-1]
	}
	return toks
}
