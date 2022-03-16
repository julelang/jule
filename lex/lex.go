package lex

import (
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/the-xlang/x/pkg/x"
	"github.com/the-xlang/x/pkg/xio"
	"github.com/the-xlang/x/pkg/xlog"
)

// Lex is lexer of Fract.
type Lex struct {
	wg               sync.WaitGroup
	firstTokenOfLine bool

	File   *xio.File
	Pos    int
	Column int
	Row    int
	// Logs are only errors
	Logs []xlog.CompilerLog

	braces []Token
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

func (l *Lex) pusherr(err string) {
	l.Logs = append(l.Logs, xlog.CompilerLog{
		Type:    xlog.Error,
		Row:     l.Row,
		Column:  l.Column,
		Path:    l.File.Path,
		Message: x.Errors[err],
	})
}

func (l *Lex) pusherrtok(tok Token, err string) {
	l.Logs = append(l.Logs, xlog.CompilerLog{
		Type:    xlog.Error,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    l.File.Path,
		Message: x.Errors[err],
	})
}

// Lex all source content.
func (l *Lex) Lex() []Token {
	var toks []Token
	l.Logs = nil
	l.Newln()
	for l.Pos < len(l.File.Content) {
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
	for _, run := range ln {
		if run != '_' &&
			('0' > run || '9' < run) &&
			!unicode.IsLetter(run) {
			break
		}
		sb.WriteRune(run)
		l.Pos++
	}
	return sb.String()
}

// resume to lex from position.
func (l *Lex) resume() string {
	var ln string
	runes := l.File.Content[l.Pos:]
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

func (l *Lex) lncomment(tok *Token) {
	start := l.Pos
	l.Pos += 2
	for ; l.Pos < len(l.File.Content); l.Pos++ {
		if l.File.Content[l.Pos] == '\n' {
			if l.firstTokenOfLine {
				tok.Id = Comment
				tok.Kind = string(l.File.Content[start:l.Pos])
			}
			return
		}
	}
	if l.firstTokenOfLine {
		tok.Id = Comment
		tok.Kind = string(l.File.Content[start:])
	}
}

func (l *Lex) rangecomment() {
	l.Pos += 2
	for ; l.Pos < len(l.File.Content); l.Pos++ {
		run := l.File.Content[l.Pos]
		if run == '\n' {
			l.Newln()
			continue
		}
		l.Column += len(string(run))
		if strings.HasPrefix(string(l.File.Content[l.Pos:]), "*/") {
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
func (l *Lex) num(content string) string {
	value := numRegexp.FindString(content)
	l.Pos += len(value)
	return value
}

var escSeqRegexp = regexp.MustCompile(`^\\([\\'"abfnrtv]|U.{8}|u.{4}|x..|[0-7]{1,3})`)

func (l *Lex) escseq(content string) string {
	seq := escSeqRegexp.FindString(content)
	if seq != "" {
		l.Pos += len([]rune(seq))
		return seq
	}
	l.Pos++
	l.pusherr("invalid_escape_sequence")
	return seq
}

func (l *Lex) getrune(content string, raw bool) string {
	if !raw && content[0] == '\\' {
		return l.escseq(content)
	}
	run, _ := utf8.DecodeRuneInString(content)
	l.Pos++
	return string(run)
}

func (l *Lex) rune(content string) string {
	var sb strings.Builder
	sb.WriteByte('\'')
	l.Column++
	content = content[1:]
	count := 0
	for index := 0; index < len(content); index++ {
		if content[index] == '\n' {
			l.pusherr("missing_rune_end")
			l.Pos++
			l.Newln()
			return ""
		}
		run := l.getrune(content[index:], false)
		sb.WriteString(run)
		length := len(run)
		l.Column += length
		if run == "'" {
			l.Pos++
			break
		}
		if length > 1 {
			index += length - 1
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

func (l *Lex) str(content string) string {
	var sb strings.Builder
	mark := content[0]
	raw := mark == '`'
	sb.WriteByte(mark)
	l.Column++
	content = content[1:]
	for i := 0; i < len(content); i++ {
		ch := content[i]
		if ch == '\n' {
			defer l.Newln()
			if !raw {
				l.pusherr("missing_string_end")
				l.Pos++
				return ""
			}
		}
		run := l.getrune(content[i:], raw)
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
	l.firstTokenOfLine = true
	l.Row++
	l.Column = 1
}

func (l *Lex) punct(content, kind string, id uint8, tok *Token) bool {
	if !strings.HasPrefix(content, kind) {
		return false
	}
	tok.Kind = kind
	tok.Id = id
	l.Pos += len([]rune(kind))
	return true
}

func (l *Lex) kw(content, kind string, id uint8, tok *Token) bool {
	if !iskw(content, kind) {
		return false
	}
	tok.Kind = kind
	tok.Id = id
	l.Pos += len([]rune(kind))
	return true
}

// Tok generates next token from resume at position.
func (l *Lex) Tok() Token {
	defer func() { l.firstTokenOfLine = false }()
	tok := Token{
		File: l.File,
		Id:   NA,
	}
	content := l.resume()
	if content == "" {
		return tok
	}
	// Set token values.
	tok.Column = l.Column
	tok.Row = l.Row

	//* Tokenize

	switch {
	case content[0] == '\'':
		tok.Kind = l.rune(content)
		tok.Id = Value
		return tok
	case content[0] == '"', content[0] == '`':
		tok.Kind = l.str(content)
		tok.Id = Value
		return tok
	case strings.HasPrefix(content, "//"):
		l.lncomment(&tok)
		goto ret
	case strings.HasPrefix(content, "/*"):
		l.rangecomment()
		return tok
	case l.punct(content, "(", Brace, &tok):
		l.braces = append(l.braces, tok)
	case l.punct(content, ")", Brace, &tok):
		length := len(l.braces)
		if length == 0 {
			l.pusherrtok(tok, "extra_closed_parentheses")
			break
		} else if l.braces[length-1].Kind != "(" {
			l.wg.Add(1)
			go l.pushWrongOrderCloseErrAsync(tok)
		}
		l.rmrange(length-1, tok.Kind)
	case l.punct(content, "{", Brace, &tok):
		l.braces = append(l.braces, tok)
	case l.punct(content, "}", Brace, &tok):
		length := len(l.braces)
		if length == 0 {
			l.pusherrtok(tok, "extra_closed_braces")
			break
		} else if l.braces[length-1].Kind != "{" {
			l.wg.Add(1)
			go l.pushWrongOrderCloseErrAsync(tok)
		}
		l.rmrange(length-1, tok.Kind)
	case l.punct(content, "[", Brace, &tok):
		l.braces = append(l.braces, tok)
	case l.punct(content, "]", Brace, &tok):
		length := len(l.braces)
		if length == 0 {
			l.pusherrtok(tok, "extra_closed_brackets")
			break
		} else if l.braces[length-1].Kind != "[" {
			l.wg.Add(1)
			go l.pushWrongOrderCloseErrAsync(tok)
		}
		l.rmrange(length-1, tok.Kind)
	case
		l.firstTokenOfLine && l.punct(content, "#", Preprocessor, &tok),
		l.punct(content, ":", Colon, &tok),
		l.punct(content, ";", SemiColon, &tok),
		l.punct(content, ",", Comma, &tok),
		l.punct(content, "@", At, &tok),
		l.punct(content, "...", Operator, &tok),
		l.punct(content, ".", Dot, &tok),
		l.punct(content, "+=", Operator, &tok),
		l.punct(content, "-=", Operator, &tok),
		l.punct(content, "*=", Operator, &tok),
		l.punct(content, "/=", Operator, &tok),
		l.punct(content, "%=", Operator, &tok),
		l.punct(content, "<<=", Operator, &tok),
		l.punct(content, ">>=", Operator, &tok),
		l.punct(content, "^=", Operator, &tok),
		l.punct(content, "&=", Operator, &tok),
		l.punct(content, "|=", Operator, &tok),
		l.punct(content, "==", Operator, &tok),
		l.punct(content, "!=", Operator, &tok),
		l.punct(content, ">=", Operator, &tok),
		l.punct(content, "<=", Operator, &tok),
		l.punct(content, "&&", Operator, &tok),
		l.punct(content, "||", Operator, &tok),
		l.punct(content, "<<", Operator, &tok),
		l.punct(content, ">>", Operator, &tok),
		l.punct(content, "+", Operator, &tok),
		l.punct(content, "-", Operator, &tok),
		l.punct(content, "*", Operator, &tok),
		l.punct(content, "/", Operator, &tok),
		l.punct(content, "%", Operator, &tok),
		l.punct(content, "~", Operator, &tok),
		l.punct(content, "&", Operator, &tok),
		l.punct(content, "|", Operator, &tok),
		l.punct(content, "^", Operator, &tok),
		l.punct(content, "!", Operator, &tok),
		l.punct(content, "<", Operator, &tok),
		l.punct(content, ">", Operator, &tok),
		l.punct(content, "=", Operator, &tok),
		l.kw(content, "i8", DataType, &tok),
		l.kw(content, "i16", DataType, &tok),
		l.kw(content, "i32", DataType, &tok),
		l.kw(content, "i64", DataType, &tok),
		l.kw(content, "u8", DataType, &tok),
		l.kw(content, "u16", DataType, &tok),
		l.kw(content, "u32", DataType, &tok),
		l.kw(content, "u64", DataType, &tok),
		l.kw(content, "f32", DataType, &tok),
		l.kw(content, "f64", DataType, &tok),
		l.kw(content, "byte", DataType, &tok),
		l.kw(content, "sbyte", DataType, &tok),
		l.kw(content, "size", DataType, &tok),
		l.kw(content, "ssize", DataType, &tok),
		l.kw(content, "bool", DataType, &tok),
		l.kw(content, "rune", DataType, &tok),
		l.kw(content, "str", DataType, &tok),
		l.kw(content, "true", Value, &tok),
		l.kw(content, "false", Value, &tok),
		l.kw(content, "nil", Value, &tok),
		l.kw(content, "const", Const, &tok),
		l.kw(content, "ret", Ret, &tok),
		l.kw(content, "type", Type, &tok),
		l.kw(content, "new", New, &tok),
		l.kw(content, "free", Free, &tok),
		l.kw(content, "iter", Iter, &tok),
		l.kw(content, "break", Break, &tok),
		l.kw(content, "continue", Continue, &tok),
		l.kw(content, "in", In, &tok),
		l.kw(content, "if", If, &tok),
		l.kw(content, "else", Else, &tok),
		l.kw(content, "volatile", Volatile, &tok),
		l.kw(content, "use", Use, &tok),
		l.kw(content, "pub", Pub, &tok):
	default:
		lex := l.id(content)
		if lex != "" {
			tok.Kind = lex
			tok.Id = Id
			break
		}
		lex = l.num(content)
		if lex != "" {
			tok.Kind = lex
			tok.Id = Value
			break
		}
		l.pusherr("invalid_token")
		l.Column++
		l.Pos++
		return tok
	}
	l.Column += len(tok.Kind)
ret:
	return tok
}

func (l *Lex) rmrange(index int, kind string) {
	var close string
	switch kind {
	case ")":
		close = "("
	case "}":
		close = "{"
	case "]":
		close = "["
	}
	for ; index >= 0; index-- {
		tok := l.braces[index]
		if tok.Kind != close {
			continue
		}
		l.braces = append(l.braces[:index], l.braces[index+1:]...)
		break
	}
}

func (l *Lex) pushWrongOrderCloseErrAsync(tok Token) {
	defer func() { l.wg.Done() }()
	var message string
	switch l.braces[len(l.braces)-1].Kind {
	case "(":
		message = "expected_parentheses_close"
	case "{":
		message = "expected_brace_close"
	case "[":
		message = "expected_bracket_close"
	}
	l.pusherrtok(tok, message)
}
