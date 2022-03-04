package lex

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/the-xlang/x/pkg/io"
	"github.com/the-xlang/x/pkg/x"
)

// Lex is lexer of Fract.
type Lex struct {
	wg sync.WaitGroup

	File     *io.File
	Position int
	Column   int
	Line     int
	Errors   []string

	braces []Token
}

// New Lex instance.
func NewLex(f *io.File) *Lex {
	l := new(Lex)
	l.File = f
	l.Line = 1
	l.Column = 1
	l.Position = 0
	return l
}

func (l *Lex) pusherr(err string) {
	l.Errors = append(l.Errors,
		fmt.Sprintf("%s %d:%d %s", l.File.Path, l.Line, l.Column, x.Errors[err]))
}

func (l *Lex) pusherrtok(tok Token, err string) {
	l.Errors = append(l.Errors,
		fmt.Sprintf("%s %d:%d %s", l.File.Path, tok.Row, tok.Column, x.Errors[err]))
}

// Tokenize all source content.
func (l *Lex) Tokenize() []Token {
	var tokens []Token
	l.Errors = nil
	for l.Position < len(l.File.Content) {
		token := l.Tok()
		if token.Id != NA {
			tokens = append(tokens, token)
		}
	}
	l.wg.Add(1)
	go l.checkRangesAsync()
	l.wg.Wait()
	return tokens
}

func (l *Lex) checkRangesAsync() {
	defer func() { l.wg.Done() }()
	for _, token := range l.braces {
		switch token.Kind {
		case "(":
			l.pusherrtok(token, "wait_close_parentheses")
		case "{":
			l.pusherrtok(token, "wait_close_brace")
		case "[":
			l.pusherrtok(token, "wait_close_bracket")
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
		l.Position++
	}
	return sb.String()
}

// resume to lex from position.
func (l *Lex) resume() string {
	var ln string
	runes := l.File.Content[l.Position:]
	// Skip spaces.
	for i, r := range runes {
		if unicode.IsSpace(r) {
			l.Column++
			l.Position++
			if r == '\n' {
				l.Newln()
			}
			continue
		}
		ln = string(runes[i:])
		break
	}
	return ln
}

func (l *Lex) lncomment() {
	l.Position += 2
	for ; l.Position < len(l.File.Content); l.Position++ {
		if l.File.Content[l.Position] == '\n' {
			l.Position++
			l.Newln()
			return
		}
	}
}

func (l *Lex) rangecomment() {
	l.Position += 2
	for ; l.Position < len(l.File.Content); l.Position++ {
		run := l.File.Content[l.Position]
		if run == '\n' {
			l.Newln()
			continue
		}
		l.Column += len(string(run))
		if strings.HasPrefix(string(l.File.Content[l.Position:]), "*/") {
			l.Column += 2
			l.Position += 2
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
	l.Position += len(value)
	return value
}

var escSeqRegexp = regexp.MustCompile(`^\\([\\'"abfnrtv]|U.{8}|u.{4}|x..|[0-7]{1,3})`)

func (l *Lex) escseq(content string) string {
	seq := escSeqRegexp.FindString(content)
	if seq != "" {
		l.Position += len([]rune(seq))
		return seq
	}
	l.Position++
	l.pusherr("invalid_escape_sequence")
	return seq
}

func (l *Lex) getrune(content string) string {
	if content[0] == '\\' {
		return l.escseq(content)
	}
	run, _ := utf8.DecodeRuneInString(content)
	l.Position++
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
			l.Position++
			l.Newln()
			return ""
		}
		run := l.getrune(content[index:])
		sb.WriteString(run)
		length := len(run)
		l.Column += length
		if run == "'" {
			l.Position++
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
	sb.WriteByte('"')
	l.Column++
	content = content[1:]
	for i := 0; i < len(content); i++ {
		ch := content[i]
		if ch == '\n' {
			l.pusherr("missing_string_end")
			l.Position++
			l.Newln()
			return ""
		}
		run := l.getrune(content[i:])
		sb.WriteString(run)
		length := len(run)
		l.Column += length
		if run == `"` {
			l.Position++
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
	l.Line++
	l.Column = 1
}

func (l *Lex) punct(content, kind string, id uint8, token *Token) bool {
	if !strings.HasPrefix(content, kind) {
		return false
	}
	token.Kind = kind
	token.Id = id
	l.Position += len([]rune(kind))
	return true
}

func (l *Lex) kw(content, kind string, id uint8, token *Token) bool {
	if !iskw(content, kind) {
		return false
	}
	token.Kind = kind
	token.Id = id
	l.Position += len([]rune(kind))
	return true
}

// Tok generates next token from resume at position.
func (l *Lex) Tok() Token {
	token := Token{
		File: l.File,
		Id:   NA,
	}
	content := l.resume()
	if content == "" {
		return token
	}
	// Set token values.
	token.Column = l.Column
	token.Row = l.Line

	//* Tokenize

	switch {
	case content[0] == '\'':
		token.Kind = l.rune(content)
		token.Id = Value
		return token
	case content[0] == '"':
		token.Kind = l.str(content)
		token.Id = Value
		return token
	case strings.HasPrefix(content, "//"):
		l.lncomment()
		return token
	case strings.HasPrefix(content, "/*"):
		l.rangecomment()
		return token
	case l.punct(content, "(", Brace, &token):
		l.braces = append(l.braces, token)
	case l.punct(content, ")", Brace, &token):
		length := len(l.braces)
		if length == 0 {
			l.pusherrtok(token, "extra_closed_parentheses")
			break
		} else if l.braces[length-1].Kind != "(" {
			l.wg.Add(1)
			go l.pushWrongOrderCloseErrAsync(token)
		}
		l.rmrange(length-1, token.Kind)
	case l.punct(content, "{", Brace, &token):
		l.braces = append(l.braces, token)
	case l.punct(content, "}", Brace, &token):
		length := len(l.braces)
		if length == 0 {
			l.pusherrtok(token, "extra_closed_braces")
			break
		} else if l.braces[length-1].Kind != "{" {
			l.wg.Add(1)
			go l.pushWrongOrderCloseErrAsync(token)
		}
		l.rmrange(length-1, token.Kind)
	case l.punct(content, "[", Brace, &token):
		l.braces = append(l.braces, token)
	case l.punct(content, "]", Brace, &token):
		length := len(l.braces)
		if length == 0 {
			l.pusherrtok(token, "extra_closed_brackets")
			break
		} else if l.braces[length-1].Kind != "[" {
			l.wg.Add(1)
			go l.pushWrongOrderCloseErrAsync(token)
		}
		l.rmrange(length-1, token.Kind)
	case
		l.punct(content, ":", Colon, &token),
		l.punct(content, ";", SemiColon, &token),
		l.punct(content, ",", Comma, &token),
		l.punct(content, "@", At, &token),
		l.punct(content, "...", Operator, &token),
		l.punct(content, "+=", Operator, &token),
		l.punct(content, "-=", Operator, &token),
		l.punct(content, "*=", Operator, &token),
		l.punct(content, "/=", Operator, &token),
		l.punct(content, "%=", Operator, &token),
		l.punct(content, "<<=", Operator, &token),
		l.punct(content, ">>=", Operator, &token),
		l.punct(content, "^=", Operator, &token),
		l.punct(content, "&=", Operator, &token),
		l.punct(content, "|=", Operator, &token),
		l.punct(content, "==", Operator, &token),
		l.punct(content, "!=", Operator, &token),
		l.punct(content, ">=", Operator, &token),
		l.punct(content, "<=", Operator, &token),
		l.punct(content, "&&", Operator, &token),
		l.punct(content, "||", Operator, &token),
		l.punct(content, "<<", Operator, &token),
		l.punct(content, ">>", Operator, &token),
		l.punct(content, "+", Operator, &token),
		l.punct(content, "-", Operator, &token),
		l.punct(content, "*", Operator, &token),
		l.punct(content, "/", Operator, &token),
		l.punct(content, "%", Operator, &token),
		l.punct(content, "~", Operator, &token),
		l.punct(content, "&", Operator, &token),
		l.punct(content, "|", Operator, &token),
		l.punct(content, "^", Operator, &token),
		l.punct(content, "!", Operator, &token),
		l.punct(content, "<", Operator, &token),
		l.punct(content, ">", Operator, &token),
		l.punct(content, "=", Operator, &token),
		l.kw(content, "i8", DataType, &token),
		l.kw(content, "i16", DataType, &token),
		l.kw(content, "i32", DataType, &token),
		l.kw(content, "i64", DataType, &token),
		l.kw(content, "u8", DataType, &token),
		l.kw(content, "u16", DataType, &token),
		l.kw(content, "u32", DataType, &token),
		l.kw(content, "u64", DataType, &token),
		l.kw(content, "f32", DataType, &token),
		l.kw(content, "f64", DataType, &token),
		l.kw(content, "size", DataType, &token),
		l.kw(content, "ssize", DataType, &token),
		l.kw(content, "bool", DataType, &token),
		l.kw(content, "rune", DataType, &token),
		l.kw(content, "str", DataType, &token),
		l.kw(content, "true", Value, &token),
		l.kw(content, "false", Value, &token),
		l.kw(content, "nil", Value, &token),
		l.kw(content, "const", Const, &token),
		l.kw(content, "ret", Ret, &token),
		l.kw(content, "type", Type, &token),
		l.kw(content, "new", New, &token),
		l.kw(content, "free", Free, &token),
		l.kw(content, "iter", Iter, &token),
		l.kw(content, "break", Break, &token),
		l.kw(content, "continue", Continue, &token),
		l.kw(content, "in", In, &token),
		l.kw(content, "if", If, &token),
		l.kw(content, "else", Else, &token),
		l.kw(content, "volatile", Volatile, &token):
	default:
		lex := l.id(content)
		if lex != "" {
			token.Kind = "_" + lex
			token.Id = Id
			break
		}
		lex = l.num(content)
		if lex != "" {
			token.Kind = lex
			token.Id = Value
			break
		}
		l.pusherr("invalid_token")
		l.Column++
		l.Position++
		return token
	}
	l.Column += len(token.Kind)
	return token
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
		token := l.braces[index]
		if token.Kind != close {
			continue
		}
		l.braces = append(l.braces[:index], l.braces[index+1:]...)
		break
	}
}

func (l *Lex) pushWrongOrderCloseErrAsync(token Token) {
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
	l.pusherrtok(token, message)
}
