package lex

import (
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xio"
	"github.com/the-xlang/xxc/pkg/xlog"
)

type File = xio.File

// Lex is lexer of Fract.
type Lex struct {
	wg             sync.WaitGroup
	firstTokOfLine bool

	File   *File
	Pos    int
	Column int
	Row    int
	// Logs are only errors
	Logs []xlog.CompilerLog

	braces []Tok
}

// New Lex instance.
func NewLex(f *File) *Lex {
	l := new(Lex)
	l.File = f
	l.Pos = 0
	l.Row = -1 // For true row
	l.Newln()
	return l
}

func (l *Lex) pusherr(key string, args ...any) {
	l.Logs = append(l.Logs, xlog.CompilerLog{
		Type:   xlog.Err,
		Row:    l.Row,
		Column: l.Column,
		Path:   l.File.Path,
		Msg:    x.GetErr(key, args...),
	})
}

func (l *Lex) pusherrtok(tok Tok, err string) {
	l.Logs = append(l.Logs, xlog.CompilerLog{
		Type:   xlog.Err,
		Row:    tok.Row,
		Column: tok.Column,
		Path:   l.File.Path,
		Msg:    x.GetErr(err),
	})
}

// Lex all source content.
func (l *Lex) Lex() []Tok {
	var toks []Tok
	l.Logs = nil
	l.Newln()
	for l.Pos < len(l.File.Data) {
		tok := l.Tok()
		if tok.Id != tokens.NA {
			toks = append(toks, tok)
		}
	}
	l.wg.Add(1)
	go l.checkRangesAsync()
	l.wg.Wait()
	return toks
}

func (l *Lex) checkRangesAsync() {
	defer func() { l.wg.Done() }()
	for _, tok := range l.braces {
		switch tok.Kind {
		case tokens.LPARENTHESES:
			l.pusherrtok(tok, "wait_close_parentheses")
		case tokens.LBRACE:
			l.pusherrtok(tok, "wait_close_brace")
		case tokens.LBRACKET:
			l.pusherrtok(tok, "wait_close_bracket")
		}
	}
}

// iskw returns true if part is keyword, false if not.
func iskw(ln, kw string) bool {
	if !strings.HasPrefix(ln, kw) {
		return false
	}
	ln = ln[len(kw):]
	r := rune(ln[0])
	return ln == "" ||
		unicode.IsSpace(r) ||
		(r != '_' && unicode.IsPunct(r))
}

// id returns identifer if next token is identifer,
// returns empty string if not.
func (l *Lex) id(ln string) string {
	if ln[0] != '_' {
		r, _ := utf8.DecodeRuneInString(ln)
		if !unicode.IsLetter(r) {
			return ""
		}
	}
	var sb strings.Builder
	for _, r := range ln {
		if r != '_' &&
			('0' > r || '9' < r) &&
			!unicode.IsLetter(r) {
			break
		}
		sb.WriteRune(r)
		l.Pos++
	}
	return sb.String()
}

// resume to lex from position.
func (l *Lex) resume() string {
	var ln string
	runes := l.File.Data[l.Pos:]
	// Skip spaces.
	for i, r := range runes {
		if unicode.IsSpace(r) {
			l.Pos++
			if r == '\n' {
				l.Newln()
			} else {
				l.Column++
			}
			continue
		}
		ln = string(runes[i:])
		break
	}
	return ln
}

func (l *Lex) lncomment(tok *Tok) {
	start := l.Pos
	l.Pos += 2
	for ; l.Pos < len(l.File.Data); l.Pos++ {
		if l.File.Data[l.Pos] == '\n' {
			if l.firstTokOfLine {
				tok.Id = tokens.Comment
				tok.Kind = string(l.File.Data[start:l.Pos])
			}
			return
		}
	}
	if l.firstTokOfLine {
		tok.Id = tokens.Comment
		tok.Kind = string(l.File.Data[start:])
	}
}

func (l *Lex) rangecomment() {
	l.Pos += 2
	for ; l.Pos < len(l.File.Data); l.Pos++ {
		run := l.File.Data[l.Pos]
		if run == '\n' {
			l.Newln()
			continue
		}
		l.Column += len(string(run))
		if strings.HasPrefix(string(l.File.Data[l.Pos:]), tokens.RANGE_COMMENT_CLOSE) {
			l.Column += 2
			l.Pos += 2
			return
		}
	}
	l.pusherr("missing_block_comment")
}

// NumRegexp is regular expression pattern for numericals.
var NumRegexp = *regexp.MustCompile(`^((0x[[:xdigit:]]+)|0b([0-1]{1,})|0([0-7]{1,})|(\d+((\.\d+)?((e|E)(\-|\+|)\d+)?|(\.\d+))))`)

// num returns numeric if next token is numeric,
// returns empty string if not.
func (l *Lex) num(txt string) string {
	val := NumRegexp.FindString(txt)
	l.Pos += len(val)
	return val
}

var escSeqRegexp = regexp.MustCompile(`^\\([\\'"abfnrtv]|U.{8}|u.{4}|x..|[0-7]{1,3})`)

func (l *Lex) escseq(txt string) string {
	seq := escSeqRegexp.FindString(txt)
	if seq != "" {
		l.Pos += len([]rune(seq))
		return seq
	}
	l.Pos++
	l.pusherr("invalid_escape_sequence")
	return seq
}

func (l *Lex) getrune(txt string, raw bool) string {
	if !raw && txt[0] == '\\' {
		return l.escseq(txt)
	}
	run, _ := utf8.DecodeRuneInString(txt)
	l.Pos++
	return string(run)
}

func (l *Lex) rune(txt string) string {
	var sb strings.Builder
	sb.WriteByte('\'')
	l.Column++
	txt = txt[1:]
	count := 0
	for i := 0; i < len(txt); i++ {
		if txt[i] == '\n' {
			l.pusherr("missing_rune_end")
			l.Pos++
			l.Newln()
			return ""
		}
		run := l.getrune(txt[i:], false)
		sb.WriteString(run)
		length := len(run)
		l.Column += length
		if run == "'" {
			l.Pos++
			break
		}
		if length > 1 {
			i += length - 1
		}
		count++
	}
	if count == 0 {
		l.pusherr("rune_empty")
	} else if count > 1 {
		l.pusherr("rune_overflow")
	}
	return sb.String()
}

func (l *Lex) str(txt string) string {
	var sb strings.Builder
	mark := txt[0]
	raw := mark == '`'
	sb.WriteByte(mark)
	l.Column++
	txt = txt[1:]
	for i := 0; i < len(txt); i++ {
		ch := txt[i]
		if ch == '\n' {
			defer l.Newln()
			if !raw {
				l.pusherr("missing_string_end")
				l.Pos++
				return ""
			}
		}
		run := l.getrune(txt[i:], raw)
		sb.WriteString(run)
		length := len(run)
		l.Column += length
		if ch == mark {
			l.Pos++
			break
		}
		if length > 1 {
			i += length - 1
		}
	}
	return sb.String()
}

// Newln sets ready lexer to a new line lexing.
func (l *Lex) Newln() {
	l.firstTokOfLine = true
	l.Row++
	l.Column = 1
}

func (l *Lex) isop(txt, kind string, id uint8, tok *Tok) bool {
	if !strings.HasPrefix(txt, kind) {
		return false
	}
	tok.Kind = kind
	tok.Id = id
	l.Pos += len([]rune(kind))
	return true
}

func (l *Lex) iskw(txt, kind string, id uint8, tok *Tok) bool {
	if !iskw(txt, kind) {
		return false
	}
	tok.Kind = kind
	tok.Id = id
	l.Pos += len([]rune(kind))
	return true
}

//               [keyword]id
var keywords = map[string]uint8{
	tokens.I8:       tokens.DataType,
	tokens.I16:      tokens.DataType,
	tokens.I32:      tokens.DataType,
	tokens.I64:      tokens.DataType,
	tokens.U8:       tokens.DataType,
	tokens.U16:      tokens.DataType,
	tokens.U32:      tokens.DataType,
	tokens.U64:      tokens.DataType,
	tokens.F32:      tokens.DataType,
	tokens.F64:      tokens.DataType,
	tokens.UINT:     tokens.DataType,
	tokens.INT:      tokens.DataType,
	tokens.UINTPTR:  tokens.DataType,
	tokens.INTPTR:   tokens.DataType,
	tokens.BOOL:     tokens.DataType,
	tokens.CHAR:     tokens.DataType,
	tokens.STR:      tokens.DataType,
	tokens.VOIDPTR:  tokens.DataType,
	tokens.TRUE:     tokens.Value,
	tokens.FALSE:    tokens.Value,
	tokens.NIL:      tokens.Value,
	tokens.CONST:    tokens.Const,
	tokens.RET:      tokens.Ret,
	tokens.TYPE:     tokens.Type,
	tokens.NEW:      tokens.New,
	tokens.FREE:     tokens.Free,
	tokens.ITER:     tokens.Iter,
	tokens.BREAK:    tokens.Break,
	tokens.CONTINUE: tokens.Continue,
	tokens.IN:       tokens.In,
	tokens.IF:       tokens.If,
	tokens.ELSE:     tokens.Else,
	tokens.VOLATILE: tokens.Volatile,
	tokens.USE:      tokens.Use,
	tokens.PUB:      tokens.Pub,
	tokens.DEFER:    tokens.Defer,
	tokens.GOTO:     tokens.Goto,
	tokens.ENUM:     tokens.Enum,
	tokens.STRUCT:   tokens.Struct,
	tokens.CO:       tokens.Co,
	tokens.TRY:      tokens.Try,
	tokens.CATCH:    tokens.Catch,
	tokens.SIZEOF:   tokens.Sizeof,
}

type oppair struct {
	op string
	id uint8
}

//                [  op  ]id
var basicOps = []oppair{
	{tokens.DOUBLE_COLON, tokens.DoubleColon},
	{tokens.COLON, tokens.Colon},
	{tokens.SEMICOLON, tokens.SemiColon},
	{tokens.COMMA, tokens.Comma},
	{tokens.AT, tokens.At},
	{tokens.TRIPLE_DOT, tokens.Operator},
	{tokens.DOT, tokens.Dot},
	{tokens.PLUS_EQUAL, tokens.Operator},
	{tokens.MINUS_EQUAL, tokens.Operator},
	{tokens.STAR_EQUAL, tokens.Operator},
	{tokens.SLASH_EQUAL, tokens.Operator},
	{tokens.PERCENT_EQUAL, tokens.Operator},
	{tokens.LSHIFT_EQUAL, tokens.Operator},
	{tokens.RSHIFT_EQUAL, tokens.Operator},
	{tokens.CARET_EQUAL, tokens.Operator},
	{tokens.AMPER_EQUAL, tokens.Operator},
	{tokens.VLINE_EQUAL, tokens.Operator},
	{tokens.EQUALS, tokens.Operator},
	{tokens.NOT_EQUALS, tokens.Operator},
	{tokens.GREAT_EQUAL, tokens.Operator},
	{tokens.LESS_EQUAL, tokens.Operator},
	{tokens.AND, tokens.Operator},
	{tokens.OR, tokens.Operator},
	{tokens.LSHIFT, tokens.Operator},
	{tokens.RSHIFT, tokens.Operator},
	{tokens.PLUS, tokens.Operator},
	{tokens.MINUS, tokens.Operator},
	{tokens.STAR, tokens.Operator},
	{tokens.SLASH, tokens.Operator},
	{tokens.PERCENT, tokens.Operator},
	{tokens.TILDE, tokens.Operator},
	{tokens.AMPER, tokens.Operator},
	{tokens.VLINE, tokens.Operator},
	{tokens.CARET, tokens.Operator},
	{tokens.EXCLAMATION, tokens.Operator},
	{tokens.LESS, tokens.Operator},
	{tokens.GREAT, tokens.Operator},
	{tokens.EQUAL, tokens.Operator},
}

func (l *Lex) lexKeywords(txt string, tok *Tok) bool {
	for key, value := range keywords {
		if l.iskw(txt, key, value, tok) {
			return true
		}
	}
	return false
}

func (l *Lex) lexBasicOps(txt string, tok *Tok) bool {
	for _, pair := range basicOps {
		if l.isop(txt, pair.op, pair.id, tok) {
			return true
		}
	}
	return false
}

// Tok generates next token from resume at position.
func (l *Lex) Tok() Tok {
	defer func() { l.firstTokOfLine = false }()

	tok := Tok{File: l.File, Id: tokens.NA}

	txt := l.resume()
	if txt == "" {
		return tok
	}

	// Set token values.
	tok.Column = l.Column
	tok.Row = l.Row

	//* Tokenize
	switch {
	case txt[0] == '\'':
		tok.Kind = l.rune(txt)
		tok.Id = tokens.Value
		return tok
	case txt[0] == '"', txt[0] == '`':
		tok.Kind = l.str(txt)
		tok.Id = tokens.Value
		return tok
	case strings.HasPrefix(txt, tokens.LINE_COMMENT):
		l.lncomment(&tok)
		goto ret
	case strings.HasPrefix(txt, tokens.RANGE_COMMENT_OPEN):
		l.rangecomment()
		return tok
	case l.isop(txt, tokens.LPARENTHESES, tokens.Brace, &tok):
		l.braces = append(l.braces, tok)
	case l.isop(txt, tokens.RPARENTHESES, tokens.Brace, &tok):
		l.pushRangeClose(tok, tokens.LPARENTHESES)
	case l.isop(txt, tokens.LBRACE, tokens.Brace, &tok):
		l.braces = append(l.braces, tok)
	case l.isop(txt, tokens.RBRACE, tokens.Brace, &tok):
		l.pushRangeClose(tok, tokens.LBRACE)
	case l.isop(txt, tokens.LBRACKET, tokens.Brace, &tok):
		l.braces = append(l.braces, tok)
	case l.isop(txt, tokens.RBRACKET, tokens.Brace, &tok):
		l.pushRangeClose(tok, tokens.LBRACKET)
	case
		l.firstTokOfLine && l.isop(txt, tokens.SHARP, tokens.Preprocessor, &tok),
		l.lexBasicOps(txt, &tok),
		l.lexKeywords(txt, &tok):
	default:
		lex := l.id(txt)
		if lex != "" {
			tok.Kind = lex
			tok.Id = tokens.Id
			break
		}
		lex = l.num(txt)
		if lex != "" {
			tok.Kind = lex
			tok.Id = tokens.Value
			break
		}
		r, sz := utf8.DecodeRuneInString(txt)
		l.pusherr("invalid_token", r)
		l.Column += sz
		l.Pos++
		return tok
	}
	l.Column += len(tok.Kind)
ret:
	return tok
}

func (l *Lex) rmrange(i int, kind string) {
	var close string
	switch kind {
	case tokens.RPARENTHESES:
		close = tokens.LPARENTHESES
	case tokens.RBRACE:
		close = tokens.LBRACE
	case tokens.RBRACKET:
		close = tokens.LBRACKET
	}
	for ; i >= 0; i-- {
		tok := l.braces[i]
		if tok.Kind != close {
			continue
		}
		l.braces = append(l.braces[:i], l.braces[i+1:]...)
		break
	}
}

func (l *Lex) pushRangeClose(tok Tok, open string) {
	len := len(l.braces)
	if len == 0 {
		switch tok.Kind {
		case tokens.RBRACKET:
			l.pusherrtok(tok, "extra_closed_brackets")
		case tokens.RBRACE:
			l.pusherrtok(tok, "extra_closed_braces")
		case tokens.RPARENTHESES:
			l.pusherrtok(tok, "extra_closed_parentheses")
		}
		return
	} else if l.braces[len-1].Kind != open {
		l.pushWrongOrderCloseErr(tok)
	}
	l.rmrange(len-1, tok.Kind)
}

func (l *Lex) pushWrongOrderCloseErr(tok Tok) {
	var msg string
	switch l.braces[len(l.braces)-1].Kind {
	case tokens.LPARENTHESES:
		msg = "expected_parentheses_close"
	case tokens.LBRACE:
		msg = "expected_brace_close"
	case tokens.LBRACKET:
		msg = "expected_bracket_close"
	}
	l.pusherrtok(tok, msg)
}
