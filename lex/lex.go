package lex

import (
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xio"
	"github.com/the-xlang/xxc/pkg/xlog"
)

// Lex is lexer of Fract.
type Lex struct {
	wg             sync.WaitGroup
	firstTokOfLine bool

	File   *xio.File
	Pos    int
	Column int
	Row    int
	// Logs are only errors
	Logs []xlog.CompilerLog

	braces []Tok
}

// New Lex instance.
func NewLex(f *xio.File) *Lex {
	l := new(Lex)
	l.File = f
	l.Pos = 0
	l.Row = -1 // For true row
	l.Newln()
	return l
}

func (l *Lex) pusherr(key string, args ...interface{}) {
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
		Msg:    x.Errs[err],
	})
}

// Lex all source content.
func (l *Lex) Lex() []Tok {
	var toks []Tok
	l.Logs = nil
	l.Newln()
	for l.Pos < len(l.File.Text) {
		tok := l.Tok()
		if tok.Id != NA {
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
		case "(":
			l.pusherrtok(tok, "wait_close_parentheses")
		case "{":
			l.pusherrtok(tok, "wait_close_brace")
		case "[":
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
	return ln == "" ||
		unicode.IsSpace(rune(ln[0])) ||
		unicode.IsPunct(rune(ln[0]))
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
	runes := l.File.Text[l.Pos:]
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
	for ; l.Pos < len(l.File.Text); l.Pos++ {
		if l.File.Text[l.Pos] == '\n' {
			if l.firstTokOfLine {
				tok.Id = Comment
				tok.Kind = string(l.File.Text[start:l.Pos])
			}
			return
		}
	}
	if l.firstTokOfLine {
		tok.Id = Comment
		tok.Kind = string(l.File.Text[start:])
	}
}

func (l *Lex) rangecomment() {
	l.Pos += 2
	for ; l.Pos < len(l.File.Text); l.Pos++ {
		run := l.File.Text[l.Pos]
		if run == '\n' {
			l.Newln()
			continue
		}
		l.Column += len(string(run))
		if strings.HasPrefix(string(l.File.Text[l.Pos:]), "*/") {
			l.Column += 2
			l.Pos += 2
			return
		}
	}
	l.pusherr("missing_block_comment")
}

var numRegexp = *regexp.MustCompile(`^((0x[[:xdigit:]]+)|(\d+((\.\d+)?((e|E)(\-|\+|)\d+)?|(\.\d+))))`)

// num returns numeric if next token is numeric,
// returns empty string if not.
func (l *Lex) num(txt string) string {
	val := numRegexp.FindString(txt)
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

func (l *Lex) punct(txt, kind string, id uint8, tok *Tok) bool {
	if !strings.HasPrefix(txt, kind) {
		return false
	}
	tok.Kind = kind
	tok.Id = id
	l.Pos += len([]rune(kind))
	return true
}

func (l *Lex) kw(txt, kind string, id uint8, tok *Tok) bool {
	if !iskw(txt, kind) {
		return false
	}
	tok.Kind = kind
	tok.Id = id
	l.Pos += len([]rune(kind))
	return true
}

// Tok generates next token from resume at position.
func (l *Lex) Tok() Tok {
	defer func() { l.firstTokOfLine = false }()

	tok := Tok{File: l.File, Id: NA}

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
		tok.Id = Value
		return tok
	case txt[0] == '"', txt[0] == '`':
		tok.Kind = l.str(txt)
		tok.Id = Value
		return tok
	case strings.HasPrefix(txt, "//"):
		l.lncomment(&tok)
		goto ret
	case strings.HasPrefix(txt, "/*"):
		l.rangecomment()
		return tok
	case l.punct(txt, "(", Brace, &tok):
		l.braces = append(l.braces, tok)
	case l.punct(txt, ")", Brace, &tok):
		len := len(l.braces)
		if len == 0 {
			l.pusherrtok(tok, "extra_closed_parentheses")
			break
		} else if l.braces[len-1].Kind != "(" {
			l.wg.Add(1)
			go l.pushWrongOrderCloseErrAsync(tok)
		}
		l.rmrange(len-1, tok.Kind)
	case l.punct(txt, "{", Brace, &tok):
		l.braces = append(l.braces, tok)
	case l.punct(txt, "}", Brace, &tok):
		len := len(l.braces)
		if len == 0 {
			l.pusherrtok(tok, "extra_closed_braces")
			break
		} else if l.braces[len-1].Kind != "{" {
			l.wg.Add(1)
			go l.pushWrongOrderCloseErrAsync(tok)
		}
		l.rmrange(len-1, tok.Kind)
	case l.punct(txt, "[", Brace, &tok):
		l.braces = append(l.braces, tok)
	case l.punct(txt, "]", Brace, &tok):
		len := len(l.braces)
		if len == 0 {
			l.pusherrtok(tok, "extra_closed_brackets")
			break
		} else if l.braces[len-1].Kind != "[" {
			l.wg.Add(1)
			go l.pushWrongOrderCloseErrAsync(tok)
		}
		l.rmrange(len-1, tok.Kind)
	case
		l.firstTokOfLine && l.punct(txt, "#", Preprocessor, &tok),
		l.punct(txt, ":", Colon, &tok),
		l.punct(txt, ";", SemiColon, &tok),
		l.punct(txt, ",", Comma, &tok),
		l.punct(txt, "@", At, &tok),
		l.punct(txt, "...", Operator, &tok),
		l.punct(txt, ".", Dot, &tok),
		l.punct(txt, "+=", Operator, &tok),
		l.punct(txt, "-=", Operator, &tok),
		l.punct(txt, "*=", Operator, &tok),
		l.punct(txt, "/=", Operator, &tok),
		l.punct(txt, "%=", Operator, &tok),
		l.punct(txt, "<<=", Operator, &tok),
		l.punct(txt, ">>=", Operator, &tok),
		l.punct(txt, "^=", Operator, &tok),
		l.punct(txt, "&=", Operator, &tok),
		l.punct(txt, "|=", Operator, &tok),
		l.punct(txt, "==", Operator, &tok),
		l.punct(txt, "!=", Operator, &tok),
		l.punct(txt, ">=", Operator, &tok),
		l.punct(txt, "<=", Operator, &tok),
		l.punct(txt, "&&", Operator, &tok),
		l.punct(txt, "||", Operator, &tok),
		l.punct(txt, "<<", Operator, &tok),
		l.punct(txt, ">>", Operator, &tok),
		l.punct(txt, "+", Operator, &tok),
		l.punct(txt, "-", Operator, &tok),
		l.punct(txt, "*", Operator, &tok),
		l.punct(txt, "/", Operator, &tok),
		l.punct(txt, "%", Operator, &tok),
		l.punct(txt, "~", Operator, &tok),
		l.punct(txt, "&", Operator, &tok),
		l.punct(txt, "|", Operator, &tok),
		l.punct(txt, "^", Operator, &tok),
		l.punct(txt, "!", Operator, &tok),
		l.punct(txt, "<", Operator, &tok),
		l.punct(txt, ">", Operator, &tok),
		l.punct(txt, "=", Operator, &tok),
		l.kw(txt, "i8", DataType, &tok),
		l.kw(txt, "i16", DataType, &tok),
		l.kw(txt, "i32", DataType, &tok),
		l.kw(txt, "i64", DataType, &tok),
		l.kw(txt, "u8", DataType, &tok),
		l.kw(txt, "u16", DataType, &tok),
		l.kw(txt, "u32", DataType, &tok),
		l.kw(txt, "u64", DataType, &tok),
		l.kw(txt, "f32", DataType, &tok),
		l.kw(txt, "f64", DataType, &tok),
		l.kw(txt, "byte", DataType, &tok),
		l.kw(txt, "sbyte", DataType, &tok),
		l.kw(txt, "size", DataType, &tok),
		l.kw(txt, "ssize", DataType, &tok),
		l.kw(txt, "bool", DataType, &tok),
		l.kw(txt, "rune", DataType, &tok),
		l.kw(txt, "str", DataType, &tok),
		l.kw(txt, "true", Value, &tok),
		l.kw(txt, "false", Value, &tok),
		l.kw(txt, "nil", Value, &tok),
		l.kw(txt, "const", Const, &tok),
		l.kw(txt, "ret", Ret, &tok),
		l.kw(txt, "type", Type, &tok),
		l.kw(txt, "new", New, &tok),
		l.kw(txt, "free", Free, &tok),
		l.kw(txt, "iter", Iter, &tok),
		l.kw(txt, "break", Break, &tok),
		l.kw(txt, "continue", Continue, &tok),
		l.kw(txt, "in", In, &tok),
		l.kw(txt, "if", If, &tok),
		l.kw(txt, "else", Else, &tok),
		l.kw(txt, "volatile", Volatile, &tok),
		l.kw(txt, "use", Use, &tok),
		l.kw(txt, "pub", Pub, &tok):
	default:
		lex := l.id(txt)
		if lex != "" {
			tok.Kind = lex
			tok.Id = Id
			break
		}
		lex = l.num(txt)
		if lex != "" {
			tok.Kind = lex
			tok.Id = Value
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
	case ")":
		close = "("
	case "}":
		close = "{"
	case "]":
		close = "["
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

func (l *Lex) pushWrongOrderCloseErrAsync(tok Tok) {
	defer func() { l.wg.Done() }()
	var msg string
	switch l.braces[len(l.braces)-1].Kind {
	case "(":
		msg = "expected_parentheses_close"
	case "{":
		msg = "expected_brace_close"
	case "[":
		msg = "expected_bracket_close"
	}
	l.pusherrtok(tok, msg)
}
