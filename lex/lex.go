package lex

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/the-xlang/x/pkg/x"
)

func (l *Lex) pushError(err string) {
	l.Errors = append(l.Errors,
		fmt.Sprintf("%s %d:%d %s", l.File.Path, l.Line, l.Column, x.Errors[err]))
}

// Tokenize all source content.
func (l *Lex) Tokenize() []Token {
	var tokens []Token
	l.Errors = nil
	for l.Position < len(l.File.Content) {
		token := l.Token()
		if token.Id != NA {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

// isKeyword returns true if part is keyword, false if not.
func isKeyword(ln, kw string) bool {
	if !strings.HasPrefix(ln, kw) {
		return false
	}
	ln = ln[len(kw):]
	switch {
	case
		ln == "",
		unicode.IsSpace(rune(ln[0])),
		unicode.IsPunct(rune(ln[0])):
		return true
	}
	return false
}

// lexName returns name if next token is name,
// returns empty string if not.
func (l *Lex) lexName(ln string) string {
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
				l.NewLine()
			}
			continue
		}
		ln = string(runes[i:])
		break
	}
	return ln
}

func (l *Lex) lexLineComment() {
	l.Position += 2
	for ; l.Position < len(l.File.Content); l.Position++ {
		if l.File.Content[l.Position] == '\n' {
			l.Position++
			l.NewLine()
			return
		}
	}
}

func (l *Lex) lexBlockComment() {
	l.Position += 2
	for ; l.Position < len(l.File.Content); l.Position++ {
		run := l.File.Content[l.Position]
		if run == '\n' {
			l.NewLine()
			continue
		}
		l.Column += len(string(run))
		if strings.HasPrefix(string(l.File.Content[l.Position:]), "*/") {
			l.Column += 2
			l.Position += 2
			return
		}
	}
	l.pushError("missing_block_comment")
}

var numericRegexp = *regexp.MustCompile(`^((0x[[:xdigit:]]+)|(\d+((\.\d+)?((e|E)(\-|\+|)\d+)?|(\.\d+))))`)

// lexNumeric returns numeric if next token is numeric,
// returns empty string if not.
func (l *Lex) lexNumeric(content string) string {
	value := numericRegexp.FindString(content)
	l.Position += len(value)
	return value
}

var escapeSequenceRegexp = regexp.MustCompile(`^\\([\\'"abfnrtv]|U.{8}|u.{4}|x..|[0-7]{1,3})`)

func (l *Lex) getEscapeSequence(content string) string {
	seq := escapeSequenceRegexp.FindString(content)
	if seq != "" {
		l.Position += len(seq)
		return seq
	}
	l.Position++
	l.pushError("invalid_escape_sequence")
	return seq
}

func (l *Lex) getRune(content string) string {
	if content[0] == '\\' {
		return l.getEscapeSequence(content)
	}
	run, _ := utf8.DecodeRuneInString(content)
	l.Position++
	return string(run)
}

func (l *Lex) lexRune(content string) string {
	var sb strings.Builder
	sb.WriteByte('\'')
	l.Column++
	content = content[1:]
	count := 0
	for index := 0; index < len(content); index++ {
		if content[index] == '\n' {
			l.pushError("missing_rune_end")
			l.Position++
			l.NewLine()
			return ""
		}
		run := l.getRune(content[index:])
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
		l.pushError("rune_empty")
	} else if count > 1 {
		l.pushError("rune_overflow")
	}
	return sb.String()
}

func (l *Lex) lexString(content string) string {
	var sb strings.Builder
	sb.WriteByte('"')
	l.Column++
	content = content[1:]
	for index, run := range content {
		if run == '\n' {
			l.pushError("missing_string_end")
			l.Position++
			l.NewLine()
			return ""
		}
		run := l.getRune(content[index:])
		sb.WriteString(run)
		length := len(run)
		l.Column += length
		if run == `"` {
			l.Position++
			break
		}
		if length > 1 {
			index += length - 1
		}
	}
	return sb.String()
}

// NewLine sets ready lexer to a new line lexing.
func (l *Lex) NewLine() {
	l.Line++
	l.Column = 1
}

func (l *Lex) lexPunct(content, kind string, id uint8, token *Token) bool {
	if !strings.HasPrefix(content, kind) {
		return false
	}
	token.Kind = kind
	token.Id = id
	l.Position += len([]rune(kind))
	return true
}

func (l *Lex) lexKeyword(content, kind string, id uint8, token *Token) bool {
	if !isKeyword(content, kind) {
		return false
	}
	token.Kind = kind
	token.Id = id
	l.Position += len([]rune(kind))
	return true
}

// Token generates next token from resume at position.
func (l *Lex) Token() Token {
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
		token.Kind = l.lexRune(content)
		token.Id = Value
		return token
	case content[0] == '"':
		token.Kind = l.lexString(content)
		token.Id = Value
		return token
	case strings.HasPrefix(content, "//"):
		l.lexLineComment()
		return token
	case strings.HasPrefix(content, "/*"):
		l.lexBlockComment()
		return token
	case
		l.lexPunct(content, ":", Colon, &token),
		l.lexPunct(content, ";", SemiColon, &token),
		l.lexPunct(content, ",", Comma, &token),
		l.lexPunct(content, "(", Brace, &token),
		l.lexPunct(content, ")", Brace, &token),
		l.lexPunct(content, "{", Brace, &token),
		l.lexPunct(content, "}", Brace, &token),
		l.lexPunct(content, "[", Brace, &token),
		l.lexPunct(content, "]", Brace, &token),
		l.lexPunct(content, "<<", Operator, &token),
		l.lexPunct(content, ">>", Operator, &token),
		l.lexPunct(content, "==", Operator, &token),
		l.lexPunct(content, "!=", Operator, &token),
		l.lexPunct(content, ">=", Operator, &token),
		l.lexPunct(content, "<=", Operator, &token),
		l.lexPunct(content, "&&", Operator, &token),
		l.lexPunct(content, "||", Operator, &token),
		l.lexPunct(content, "+", Operator, &token),
		l.lexPunct(content, "-", Operator, &token),
		l.lexPunct(content, "*", Operator, &token),
		l.lexPunct(content, "/", Operator, &token),
		l.lexPunct(content, "%", Operator, &token),
		l.lexPunct(content, "~", Operator, &token),
		l.lexPunct(content, "&", Operator, &token),
		l.lexPunct(content, "|", Operator, &token),
		l.lexPunct(content, "^", Operator, &token),
		l.lexPunct(content, "!", Operator, &token),
		l.lexPunct(content, "<", Operator, &token),
		l.lexPunct(content, ">", Operator, &token),
		l.lexPunct(content, "=", Operator, &token),
		l.lexKeyword(content, "int8", DataType, &token),
		l.lexKeyword(content, "int16", DataType, &token),
		l.lexKeyword(content, "int32", DataType, &token),
		l.lexKeyword(content, "int64", DataType, &token),
		l.lexKeyword(content, "uint8", DataType, &token),
		l.lexKeyword(content, "uint16", DataType, &token),
		l.lexKeyword(content, "uint32", DataType, &token),
		l.lexKeyword(content, "uint64", DataType, &token),
		l.lexKeyword(content, "float32", DataType, &token),
		l.lexKeyword(content, "float64", DataType, &token),
		l.lexKeyword(content, "bool", DataType, &token),
		l.lexKeyword(content, "rune", DataType, &token),
		l.lexKeyword(content, "str", DataType, &token),
		l.lexKeyword(content, "true", Value, &token),
		l.lexKeyword(content, "false", Value, &token),
		l.lexKeyword(content, "null", Value, &token),
		l.lexKeyword(content, "const", Const, &token),
		l.lexKeyword(content, "ret", Return, &token),
		l.lexKeyword(content, "type", Type, &token):
	default:
		lex := l.lexName(content)
		if lex != "" {
			token.Kind = "_" + lex
			token.Id = Name
			break
		}
		lex = l.lexNumeric(content)
		if lex != "" {
			token.Kind = lex
			token.Id = Value
			break
		}
		l.pushError("invalid_token")
		l.Column++
		l.Position++
		return token
	}
	l.Column += len(token.Kind)
	return token
}
