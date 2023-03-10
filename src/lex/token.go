package lex

import (
	"strings"
	"unicode/utf8"

	"github.com/julelang/jule/build"
)

// Token identities.
const ID_NA = 0
const ID_DT = 1
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

const IGNORE_ID = "_"
const ANONYMOUS_ID = "<anonymous>"

const COMMENT_PRAGMA_SEP = ":"
const DIRECTIVE_COMMENT_PREFIX = "jule" + COMMENT_PRAGMA_SEP

const MARK_ARRAY = "..."
const PREFIX_SLICE = "[]"
const PREFIX_ARRAY = "[" + MARK_ARRAY + "]"

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

var SPACES = [...]rune{
	' ',
	'\t',
	'\v',
	'\r',
	'\n',
}

// UNARY_OPS list of unary operators.
var UNARY_OPS = [...]string{
	KND_MINUS,
	KND_PLUS,
	KND_CARET,
	KND_EXCL,
	KND_STAR,
	KND_AMPER,
}

// STRONG_OPS list of strong operators.
// These operators are strong, can't used as part of expression.
var STRONG_OPS = [...]string{
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

// WEAK_OPS list of weak operators.
// These operators are weak, can used as part of expression.
var WEAK_OPS = [...]string{
	KND_TRIPLE_DOT,
	KND_COLON,
}

func exist_op(kind string, operators []string) bool {
	for _, operator := range operators {
		if kind == operator {
			return true
		}
	}
	return false
}

// IsUnaryOp is returns true if operator is unary or smilar to unary,
// returns false if not.
func IsUnaryOp(kind string) bool { return exist_op(kind, UNARY_OPS[:]) }

// IsStrongOp returns true operator kind is not repeatable, false if not.
func IsStrongOp(kind string) bool { return exist_op(kind, STRONG_OPS[:]) }

// IsExprOp reports operator kind is allow as expression operator or not.
func IsExprOp(kind string) bool { return exist_op(kind, WEAK_OPS[:]) }

// Token is lexer token.
type Token struct {
	File   *File
	Row    int
	Column int
	Kind   string
	Id     uint8
}

// Prec returns operator precedence of token.
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

func IsStr(k string) bool    { return k != "" && (k[0] == '"' || IsRawStr(k)) }
func IsRawStr(k string) bool { return k != "" && k[0] == '`' }
func IsRune(k string) bool   { return k != "" && k[0] == '\'' }
func IsNil(k string) bool    { return k == KND_NIL }
func IsBool(k string) bool   { return k == KND_TRUE || k == KND_FALSE }

func contains_any(s string, bytes string) bool {
	for _, b := range bytes {
		i := strings.Index(s, string(b))
		if i >= 0 {
			return true
		}
	}
	return false
}

func IsFloat(k string) bool {
	if strings.HasPrefix(k, "0x") {
		return contains_any(k, ".pP")
	}
	return contains_any(k, ".eE")
}

func IsNum(k string) bool {
	if k == "" {
		return false
	}
	return k[0] == '-' || (k[0] >= '0' && k[0] <= '9')
}

func IsLiteral(k string) bool {
	return IsNum(k) || IsStr(k) || IsRune(k) || IsNil(k) || IsBool(k)
}

// IsIgnoreId reports identifier is ignore or not.
func IsIgnoreId(id string) bool { return id == IGNORE_ID }

// IsAnonymousId reports whether identifier is anonymous.
func IsAnonymousId(id string) bool { return id == ANONYMOUS_ID }

func rune_exist(r rune, runes []rune) bool {
	for _, cr := range runes {
		if r == cr {
			return true
		}
	}
	return false
}

// IsPunct reports rune is punctuation or not.
func IsPunct(r rune) bool { return rune_exist(r, PUNCTS[:]) }

// IsSpace reports byte is whitespace or not.
func IsSpace(r rune) bool { return rune_exist(r, SPACES[:]) }

// IsLetter reports rune is letter or not.
func IsLetter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

// IsIdentifierRune returns true if first rune of string is allowed to
// first char for identifiers, false if not.
func IsIdentifierRune(s string) bool {
	if s == "" {
		return false
	}
	if s[0] != '_' {
		r, _ := utf8.DecodeRuneInString(s)
		if !IsLetter(r) {
			return false
		}
	}
	return true
}

// IsDecimal reports byte is decimal sequence or not.
func IsDecimal(b byte) bool { return '0' <= b && b <= '9' }

// IsBinary reports byte is binary sequence or not.
func IsBinary(b byte) bool { return b == '0' || b == '1' }

// IsOctal reports byte is octal sequence or not.
func IsOctal(b byte) bool { return '0' <= b && b <= '7' }

// IsHex reports byte is hexadecimal sequence or not.
func IsHex(b byte) bool {
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
	n := 1
	start := *i
	for ; n != 0 && *i < len(tokens); *i++ {
		tok := tokens[*i]
		if tok.Id != ID_RANGE {
			continue
		}
		switch tok.Kind {
		case open:
			n++
		case close:
			n--
		}
	}
	return tokens[start : *i-1]
}

// RangeLast returns last range from tokens.
// Returns tokens without range tokens and range tokens.
//
// Special cases are;
//  RangeLast(toks) = toks, nil if len(toks) == 0
//  RangeLast(toks) = toks, nil if toks is not has range at last
func RangeLast(tokens []Token) (cutted []Token, cut []Token) {
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
// Parts(toks) = nil if len(toks) == 0
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
