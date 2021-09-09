package lex

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/the-xlang/x/pkg/x"
)

// error thrown exception.
func (l *Lex) error(err string) {
	l.Errors = append(l.Errors,
		fmt.Sprintf("%s %d:%d %s", l.File.Path, l.Line, l.Column, x.Errors[err]))
}

// Tokenize all source content.
func (l *Lex) Tokenize() []Token {
	var tokens []Token
	l.Errors = nil
	for l.Position < len(l.File.Content) {
		token := l.Token()
		if token.Type != NA {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

// isKeyword returns true if part is keyword, false if not.
func isKeyword(ln, kw string) bool {
	return regexp.MustCompile("^" + kw + `(\s+|$|[[:punct:]])`).MatchString(ln)
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

// Resume to lex from position.
func (l *Lex) resume() string {
	var ln string
	runes := l.File.Content[l.Position:]
	// Skip spaces.
	for i, r := range runes {
		if unicode.IsSpace(r) {
			l.Column++
			l.Position++
			if r == '\n' || r == '\r' {
				l.Line++
				l.Column = 1
			}
			continue
		}
		ln = string(runes[i:])
		break
	}
	return ln
}

// Token generates next token from resume at position.
func (l *Lex) Token() Token {
	tk := Token{
		File: l.File,
		Type: NA,
	}
	ln := l.resume()
	if ln == "" {
		return tk
	}
	// Set token values.
	tk.Column = l.Column
	tk.Line = l.Line

	//* Tokenize

	switch {
	case ln[0] == '(':
		tk.Value = "("
		tk.Type = Brace
		l.Position++
	case ln[0] == ')':
		tk.Value = ")"
		tk.Type = Brace
		l.Position++
	case ln[0] == '{':
		tk.Value = "{"
		tk.Type = Brace
		l.Position++
	case ln[0] == '}':
		tk.Value = "}"
		tk.Type = Brace
		l.Position++
	case isKeyword(ln, "int8"):
		tk.Value = "int8"
		tk.Type = Type
		l.Position += 5
	case isKeyword(ln, "int16"):
		tk.Value = "int16"
		tk.Type = Type
		l.Position += 5
	case isKeyword(ln, "int32"):
		tk.Value = "int32"
		tk.Type = Type
		l.Position += 5
	case isKeyword(ln, "int64"):
		tk.Value = "int64"
		tk.Type = Type
		l.Position += 5
	default:
		if chk := l.lexName(ln); chk != "" {
			tk.Value = chk
			tk.Type = Name
			break
		}
		l.error("invalid_token")
		l.Column++
		l.Position++
		return tk
	}
	l.Column += len(tk.Value)
	return tk
}
