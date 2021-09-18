package lex

import (
	"fmt"
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
		if token.Type != NA {
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
	if ln == "" {
		return true
	} else if unicode.IsSpace(rune(ln[0])) {
		return true
	} else if unicode.IsPunct(rune(ln[0])) {
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

// lexNumeric returns numeric if next token is numeric,
// returns empty string if not.
func (l *Lex) lexNumeric(ln string) string {
	for index, run := range ln {
		if '0' <= run && '9' >= run {
			l.Position++
			continue
		}
		return ln[:index]
	}
	return ""
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
			if r == '\n' {
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
	token := Token{
		File: l.File,
		Type: NA,
	}
	ln := l.resume()
	if ln == "" {
		return token
	}
	// Set token values.
	token.Column = l.Column
	token.Line = l.Line

	//* Tokenize

	switch {
	case ln[0] == ';':
		token.Value = ";"
		token.Type = SemiColon
		l.Position++
	case ln[0] == ',':
		token.Value = ","
		token.Type = Comma
		l.Position++
	case ln[0] == '(':
		token.Value = "("
		token.Type = Brace
		l.Position++
	case ln[0] == ')':
		token.Value = ")"
		token.Type = Brace
		l.Position++
	case ln[0] == '{':
		token.Value = "{"
		token.Type = Brace
		l.Position++
	case ln[0] == '}':
		token.Value = "}"
		token.Type = Brace
		l.Position++
	case ln[0] == '[':
		token.Value = "["
		token.Type = Brace
		l.Position++
	case ln[0] == ']':
		token.Value = "]"
		token.Type = Brace
		l.Position++
	case strings.HasPrefix(ln, "//"):
		index := strings.IndexByte(ln, '\n')
		if index == -1 {
			l.Position += len(ln)
		} else {
			l.Position += index
		}
		return token
	case strings.HasPrefix(ln, "<<"):
		token.Value = "<<"
		token.Type = Operator
		l.Position += 2
	case strings.HasPrefix(ln, ">>"):
		token.Value = ">>"
		token.Type = Operator
		l.Position += 2
	case strings.HasPrefix(ln, "=="):
		token.Value = "=="
		token.Type = Operator
		l.Position += 2
	case strings.HasPrefix(ln, "!="):
		token.Value = "!="
		token.Type = Operator
		l.Position += 2
	case strings.HasPrefix(ln, ">="):
		token.Value = ">="
		token.Type = Operator
		l.Position += 2
	case strings.HasPrefix(ln, "<="):
		token.Value = "<="
		token.Type = Operator
		l.Position += 2
	case strings.HasPrefix(ln, "&&"):
		token.Value = "&&"
		token.Type = Operator
		l.Position += 2
	case strings.HasPrefix(ln, "||"):
		token.Value = "||"
		token.Type = Operator
		l.Position += 2
	case ln[0] == '+':
		token.Value = "+"
		token.Type = Operator
		l.Position++
	case ln[0] == '-':
		token.Value = "-"
		token.Type = Operator
		l.Position++
	case ln[0] == '*':
		token.Value = "*"
		token.Type = Operator
		l.Position++
	case ln[0] == '/':
		token.Value = "/"
		token.Type = Operator
		l.Position++
	case ln[0] == '%':
		token.Value = "%"
		token.Type = Operator
		l.Position++
	case ln[0] == '~':
		token.Value = "~"
		token.Type = Operator
		l.Position++
	case ln[0] == '&':
		token.Value = "&"
		token.Type = Operator
		l.Position++
	case ln[0] == '|':
		token.Value = "|"
		token.Type = Operator
		l.Position++
	case ln[0] == '^':
		token.Value = "^"
		token.Type = Operator
		l.Position++
	case ln[0] == '!':
		token.Value = "!"
		token.Type = Operator
		l.Position++
	case ln[0] == '<':
		token.Value = "<"
		token.Type = Operator
		l.Position++
	case ln[0] == '>':
		token.Value = ">"
		token.Type = Operator
		l.Position++
	case ln[0] == '=':
		token.Value = "="
		token.Type = Operator
		l.Position++
	case isKeyword(ln, "fun"):
		token.Value = "fun"
		token.Type = Fun
		l.Position += 3
	case isKeyword(ln, "var"):
		token.Value = "var"
		token.Type = Var
		l.Position += 3
	case isKeyword(ln, "any"):
		token.Value = "any"
		token.Type = Type
		l.Position += 3
	case isKeyword(ln, "bool"):
		token.Value = "bool"
		token.Type = Type
		l.Position += 4
	case isKeyword(ln, "int8"):
		token.Value = "int8"
		token.Type = Type
		l.Position += 4
	case isKeyword(ln, "int16"):
		token.Value = "int16"
		token.Type = Type
		l.Position += 5
	case isKeyword(ln, "int32"):
		token.Value = "int32"
		token.Type = Type
		l.Position += 5
	case isKeyword(ln, "int64"):
		token.Value = "int64"
		token.Type = Type
		l.Position += 5
	case isKeyword(ln, "uint8"):
		token.Value = "uint8"
		token.Type = Type
		l.Position += 5
	case isKeyword(ln, "uint16"):
		token.Value = "uint16"
		token.Type = Type
		l.Position += 6
	case isKeyword(ln, "uint32"):
		token.Value = "uint32"
		token.Type = Type
		l.Position += 6
	case isKeyword(ln, "uint64"):
		token.Value = "uint64"
		token.Type = Type
		l.Position += 6
	case isKeyword(ln, "float32"):
		token.Value = "float32"
		token.Type = Type
		l.Position += 7
	case isKeyword(ln, "float64"):
		token.Value = "float64"
		token.Type = Type
		l.Position += 7
	case isKeyword(ln, "return"):
		token.Value = "return"
		token.Type = Return
		l.Position += 6
	case isKeyword(ln, "bool"):
		token.Value = "bool"
		token.Type = Type
		l.Position += 4
	case isKeyword(ln, "true"):
		token.Value = "true"
		token.Type = Value
		l.Position += 4
	case isKeyword(ln, "false"):
		token.Value = "false"
		token.Type = Value
		l.Position += 5
	default:
		lex := l.lexName(ln)
		if lex != "" {
			token.Value = lex
			token.Type = Name
			break
		}
		lex = l.lexNumeric(ln)
		if lex != "" {
			token.Value = lex
			token.Type = Value
			break
		}
		l.pushError("invalid_token")
		l.Column++
		l.Position++
		return token
	}
	l.Column += len(token.Value)
	return token
}
