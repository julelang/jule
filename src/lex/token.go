// Copyright 2021 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package lex

import (
	"strings"
	"unicode/utf8"

	"github.com/julelang/jule/build"
)

// Token identities.
const ID_NA = 0
const ID_PRIM = 1
const ID_IDENT = 2
const ID_RANGE = 3
const ID_RET = 4
const ID_SEMICOLON = 5
const ID_LIT = 6
const ID_OP = 7
const ID_COMMA = 8
const ID_CONST = 9
const ID_TYPE = 10
const ID_COLON = 11
const ID_FOR = 12
const ID_BREAK = 13
const ID_CONTINUE = 14
const ID_IN = 15
const ID_IF = 16
const ID_ELSE = 17
const ID_COMMENT = 18
const ID_USE = 19
const ID_DOT = 20
const ID_PUB = 21
const ID_GOTO = 22
const ID_DBLCOLON = 23
const ID_ENUM = 24
const ID_STRUCT = 25
const ID_CO = 26
const ID_MATCH = 27
const ID_SELF = 28
const ID_TRAIT = 29
const ID_IMPL = 30
const ID_CPP = 31
const ID_FALL = 32
const ID_FN = 33
const ID_LET = 34
const ID_UNSAFE = 35
const ID_MUT = 36
const ID_DEFER = 37

// Token kinds.
const KND_DBLCOLON = "::"
const KND_COLON = ":"
const KND_SEMICOLON = ";"
const KND_COMMA = ","
const KND_TRIPLE_DOT = "..."
const KND_DOT = "."
const KND_PLUS_EQ = "+="
const KND_MINUS_EQ = "-="
const KND_STAR_EQ = "*="
const KND_SOLIDUS_EQ = "/="
const KND_PERCENT_EQ = "%="
const KND_LSHIFT_EQ = "<<="
const KND_RSHIFT_EQ = ">>="
const KND_CARET_EQ = "^="
const KND_AMPER_EQ = "&="
const KND_VLINE_EQ = "|="
const KND_EQS = "=="
const KND_NOT_EQ = "!="
const KND_GREAT_EQ = ">="
const KND_LESS_EQ = "<="
const KND_DBL_AMPER = "&&"
const KND_DBL_VLINE = "||"
const KND_LSHIFT = "<<"
const KND_RSHIFT = ">>"
const KND_DBL_PLUS = "++"
const KND_DBL_MINUS = "--"
const KND_PLUS = "+"
const KND_MINUS = "-"
const KND_STAR = "*"
const KND_SOLIDUS = "/"
const KND_PERCENT = "%"
const KND_AMPER = "&"
const KND_VLINE = "|"
const KND_CARET = "^"
const KND_EXCL = "!"
const KND_LT = "<"
const KND_GT = ">"
const KND_EQ = "="
const KND_LN_COMMENT = "//"
const KND_RNG_LCOMMENT = "/*"
const KND_RNG_RCOMMENT = "*/"
const KND_LPAREN = "("
const KND_RPARENT = ")"
const KND_LBRACKET = "["
const KND_RBRACKET = "]"
const KND_LBRACE = "{"
const KND_RBRACE = "}"
const KND_I8 = "i8"
const KND_I16 = "i16"
const KND_I32 = "i32"
const KND_I64 = "i64"
const KND_U8 = "u8"
const KND_U16 = "u16"
const KND_U32 = "u32"
const KND_U64 = "u64"
const KND_F32 = "f32"
const KND_F64 = "f64"
const KND_UINT = "uint"
const KND_INT = "int"
const KND_UINTPTR = "uintptr"
const KND_BOOL = "bool"
const KND_STR = "str"
const KND_ANY = "any"
const KND_TRUE = "true"
const KND_FALSE = "false"
const KND_NIL = "nil"
const KND_CONST = "const"
const KND_RET = "ret"
const KND_TYPE = "type"
const KND_ITER = "for"
const KND_BREAK = "break"
const KND_CONTINUE = "continue"
const KND_IN = "in"
const KND_IF = "if"
const KND_ELSE = "else"
const KND_USE = "use"
const KND_PUB = "pub"
const KND_GOTO = "goto"
const KND_ENUM = "enum"
const KND_STRUCT = "struct"
const KND_CO = "co"
const KND_MATCH = "match"
const KND_SELF = "self"
const KND_TRAIT = "trait"
const KND_IMPL = "impl"
const KND_CPP = "cpp"
const KND_FALL = "fall"
const KND_FN = "fn"
const KND_LET = "let"
const KND_UNSAFE = "unsafe"
const KND_MUT = "mut"
const KND_DEFER = "defer"

// Ignore identifier.
const IGNORE_IDENT = "_"
// Anonymous identifier.
const ANON_IDENT = "<anonymous>"

// Directive seperator of directive comments.
const DIRECTIVE_SEP = ":"
// Prefix of directive comments.
const DIRECTIVE_PREFIX = "jule" + DIRECTIVE_SEP

// Punctuations.
var PUNCTS = [...]rune{
	'!',
	'#',
	'$',
	',',
	'.',
	'\'',
	'"',
	':',
	';',
	'<',
	'>',
	'=',
	'?',
	'-',
	'+',
	'*',
	'(',
	')',
	'[',
	']',
	'{',
	'}',
	'%',
	'&',
	'/',
	'\\',
	'@',
	'^',
	'_',
	'`',
	'|',
	'~',
	'¦',
}

// Space characters.
var SPACES = [...]rune{
	' ',
	'\t',
	'\v',
	'\r',
	'\n',
}

// Kind list of unary operators.
var UNARY_OPS = [...]string{
	KND_MINUS,
	KND_PLUS,
	KND_CARET,
	KND_EXCL,
	KND_STAR,
	KND_AMPER,
}

// Kind list of binary operators.
var BIN_OPS = [...]string{
	KND_PLUS,
	KND_MINUS,
	KND_STAR,
	KND_SOLIDUS,
	KND_PERCENT,
	KND_AMPER,
	KND_VLINE,
	KND_CARET,
	KND_LT,
	KND_GT,
	KND_EXCL,
	KND_DBL_AMPER,
	KND_DBL_VLINE,
}

// Kind list of weak operators.
// These operators are weak, can used as part of expression.
var WEAK_OPS = [...]string{
	KND_TRIPLE_DOT,
	KND_COLON,
}

// List of postfix operators.
var POSTFIX_OPS = [...]string{
	KND_DBL_PLUS,
	KND_DBL_MINUS,
}

// List of assign operators.
var ASSING_OPS = [...]string{
	KND_EQ,
	KND_PLUS_EQ,
	KND_MINUS_EQ,
	KND_SOLIDUS_EQ,
	KND_STAR_EQ,
	KND_PERCENT_EQ,
	KND_RSHIFT_EQ,
	KND_LSHIFT_EQ,
	KND_VLINE_EQ,
	KND_AMPER_EQ,
	KND_CARET_EQ,
}

// Token is lexer token.
type Token struct {
	File   *File
	Row    int
	Column int
	Kind   string
	Id     uint8
}

// Returns operator precedence of token.
// Returns -1 if token is not operator or invalid operator for operator precedence.
func (t *Token) Prec() int {
	if t.Id != ID_OP {
		return -1
	}

	switch t.Kind {
	case KND_STAR, KND_PERCENT, KND_SOLIDUS,
		KND_RSHIFT, KND_LSHIFT, KND_AMPER:
		return 5

	case KND_PLUS, KND_MINUS, KND_VLINE, KND_CARET:
		return 4

		case KND_EQS, KND_NOT_EQ, KND_LT,
		KND_LESS_EQ, KND_GT, KND_GREAT_EQ:
		return 3

	case KND_DBL_AMPER:
		return 2

	case KND_DBL_VLINE:
		return 1

	default:
		return -1
	}
}

func exist_op(kind string, operators []string) bool {
	for _, operator := range operators {
		if kind == operator {
			return true
		}
	}
	return false
}

// Reports whether kind is unary operator.
func Is_unary_op(kind string) bool { return exist_op(kind, UNARY_OPS[:]) }
// Reports whether kind is binary operator.
func Is_bin_op(kind string) bool { return exist_op(kind, BIN_OPS[:]) }
// Reports whether kind is weak operator.
func Is_weak_op(kind string) bool { return exist_op(kind, WEAK_OPS[:]) }
// Reports whether kind is string literal.
func Is_str(k string) bool { return k != "" && (k[0] == '"' || Is_raw_str(k)) }
// Reports whether kind is raw string literal.
func Is_raw_str(k string) bool { return k != "" && k[0] == '`' }
// Reports whether kind is rune literal.
// Literal value can be byte or rune.
func Is_rune(k string) bool { return k != "" && k[0] == '\'' }
// Reports whether kind is nil literal.
func Is_nil(k string) bool { return k == KND_NIL }
// Reports whether kind is boolean literal.
func Is_bool(k string) bool { return k == KND_TRUE || k == KND_FALSE }

func contains_any(s string, bytes string) bool {
	for _, b := range bytes {
		i := strings.Index(s, string(b))
		if i >= 0 {
			return true
		}
	}
	return false
}

// Reports whether kind is float.
func Is_float(k string) bool {
	if strings.HasPrefix(k, "0x") {
		return contains_any(k, ".pP")
	}
	return contains_any(k, ".eE")
}

// Reports whether kind is numeric.
func Is_num(k string) bool {
	if k == "" {
		return false
	}

	b := k[0]
	return b == '.' || ('0' <= b && b <= '9')
}

// Reports whether kind is literal.
func Is_lit(k string) bool {
	return Is_num(k) || Is_str(k) || Is_rune(k) || Is_nil(k) || Is_bool(k)
}

// Reports whether identifier is ignore.
func Is_ignore_ident(ident string) bool { return ident == IGNORE_IDENT }
// Reports whether identifier is anonymous.
func Is_anon_ident(ident string) bool { return ident == ANON_IDENT }

func rune_exist(r rune, runes []rune) bool {
	for _, cr := range runes {
		if r == cr {
			return true
		}
	}

	return false
}

// Reports whether rune is punctuation.
func Is_punct(r rune) bool { return rune_exist(r, PUNCTS[:]) }
// Reports wheter byte is whitespace.
func Is_space(r rune) bool { return rune_exist(r, SPACES[:]) }

// Reports whether rune is letter.
func Is_letter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

// Reports whether firs rune of string is allowed
// to first rune for identifier.
func Is_ident_rune(s string) bool {
	if s == "" {
		return false
	}
	if s[0] != '_' {
		r, _ := utf8.DecodeRuneInString(s)
		if !Is_letter(r) {
			return false
		}
	}
	return true
}

// Reports whether byte is decimal sequence.
func Is_decimal(b byte) bool { return '0' <= b && b <= '9' }
// Reports whether byte is binary sequence.
func Is_binary(b byte) bool { return b == '0' || b == '1' }
// Reports whether byte is octal sequence.
func Is_octal(b byte) bool { return '0' <= b && b <= '7' }

// Reports whether byte is hexadecimal sequence.
func Is_hex(b byte) bool {
	switch {
	case '0' <= b && b <= '9':
		return true

	case 'a' <= b && b <= 'f':
		return true

	case 'A' <= b && b <= 'F':
		return true

	default:
		return false
	}
}

// Returns between of open and close ranges.
// Starts selection at *i.
// Moves one *i for each selected token.
// *i points to close range token after selection.
//
// Special case is:
//  Range(i, open, close, tokens) = nil if i == nil
//  Range(i, open, close, tokens) = nil if *i > len(tokens)
//  Range(i, open, close, tokens) = nil if tokens[i*].Id != ID_RANGE
//  Range(i, open, close, tokens) = nil if tokens[i*].Kind != open
func Range(i *int, open string, close string, tokens []Token) []Token {
	if i == nil || *i >= len(tokens) {
		return nil
	}

	tok := tokens[*i]
	if tok.Id != ID_RANGE || tok.Kind != open {
		return nil
	}

	*i++
	range_n := 1
	start := *i
	for ; range_n != 0 && *i < len(tokens); *i++ {
		token := tokens[*i]
		if token.Id == ID_RANGE {
			switch token.Kind {
			case open:
				range_n++
			case close:
				range_n--
			}
		}
	}

	return tokens[start : *i-1]
}

// Range_last returns last range from tokens.
// Returns tokens without range tokens and range tokens.
// Range tokens includes left and right range tokens.
//
// Special cases are;
//  Range_last(toks) = toks, nil if len(toks) == 0
//  Range_last(toks) = toks, nil if toks is not has range at last
func Range_last(tokens []Token) (cutted []Token, cut []Token) {
	if len(tokens) == 0 {
		return tokens, nil
	} else if tokens[len(tokens)-1].Id != ID_RANGE {
		return tokens, nil
	}
	brace_n := 0
	for i := len(tokens) - 1; i >= 0; i-- {
		token := tokens[i]
		if token.Id == ID_RANGE {
			switch token.Kind {
			case KND_RBRACE, KND_RBRACKET, KND_RPARENT:
				brace_n++
				continue
			default:
				brace_n--
			}
		}
		if brace_n == 0 {
			return tokens[:i], tokens[i:]
		}
	}
	return tokens, nil
}

// Returns parts separated by given token identifier.
// It's skips parentheses ranges.
// Logs missing_expr if expr_must == true and not exist any expression for part.
//
// Special case is;
//  Parts(toks) = nil if len(toks) == 0
func Parts(tokens []Token, id uint8, expr_must bool) ([][]Token, []build.Log) {
	if len(tokens) == 0 {
		return nil, nil
	}

	var parts [][]Token = nil
	var errors []build.Log = nil

	range_n := 0
	last := 0
	for i, token := range tokens {
		if token.Id == ID_RANGE {
			switch token.Kind {
			case KND_LBRACE, KND_LBRACKET, KND_LPAREN:
				range_n++
				continue
			default:
				range_n--
			}
		}
		if range_n > 0 {
			continue
		}
		if token.Id == id {
			if expr_must && i-last <= 0 {
				errors = append(errors, make_err(token.Row, token.Column, token.File, "missing_expr"))
			}
			parts = append(parts, tokens[last:i])
			last = i + 1
		}
	}

	if last < len(tokens) {
		parts = append(parts, tokens[last:])
	} else if !expr_must {
		parts = append(parts, []Token{})
	}

	return parts, errors
}

// Reports given token id is allow for
// assignment left-expression or not.
func Is_assign(id uint8) bool {
	switch id {
	case ID_IDENT,
		ID_CPP,
		ID_LET,
		ID_DOT,
		ID_SELF,
		ID_RANGE,
		ID_OP:
		return true

	default:
		return false
	}
}

// Reports whether operator kind is postfix operator.
func Is_postfix_op(kind string) bool {
	for _, op := range POSTFIX_OPS {
		if kind == op {
			return true
		}
	}

	return false
}

// Reports whether operator kind is assignment operator.
func Is_assign_op(kind string) bool {
	if Is_postfix_op(kind) {
		return true
	}

	for _, op := range ASSING_OPS {
		if kind == op {
			return true
		}
	}

	return false
}
