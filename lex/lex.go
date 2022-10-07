package lex

import (
	"strings"
	"unicode/utf8"

	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/jule"
	"github.com/jule-lang/jule/pkg/juleio"
	"github.com/jule-lang/jule/pkg/julelog"
)

// Lex is lexer of Jule.
type Lex struct {
	firstTokenOfLine bool

	File   *juleio.File
	Pos    int
	Column int
	Row    int
	// Logs are only errors
	Logs []julelog.CompilerLog

	braces []Token
}

// New Lex instance.
func NewLex(f *juleio.File) *Lex {
	l := new(Lex)
	l.File = f
	l.Pos = 0
	l.Row = -1 // For true row
	l.Newln()
	return l
}

func (l *Lex) pusherr(key string, args ...any) {
	l.Logs = append(l.Logs, julelog.CompilerLog{
		Type:    julelog.Error,
		Row:     l.Row,
		Column:  l.Column,
		Path:    l.File.Path(),
		Message: jule.GetError(key, args...),
	})
}

func (l *Lex) pusherrtok(tok Token, err string) {
	l.Logs = append(l.Logs, julelog.CompilerLog{
		Type:    julelog.Error,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    l.File.Path(),
		Message: jule.GetError(err),
	})
}

// Lex all source content.
func (l *Lex) Lex() []Token {
	var toks []Token
	l.Logs = nil
	l.Newln()
	for l.Pos < len(l.File.Data) {
		tok := l.Token()
		if tok.Id != tokens.NA {
			toks = append(toks, tok)
		}
	}
	l.checkRanges()
	return toks
}

func (l *Lex) checkRanges() {
	for _, t := range l.braces {
		switch t.Kind {
		case tokens.LPARENTHESES:
			l.pusherrtok(t, "wait_close_parentheses")
		case tokens.LBRACE:
			l.pusherrtok(t, "wait_close_brace")
		case tokens.LBRACKET:
			l.pusherrtok(t, "wait_close_bracket")
		}
	}
}

// IsPunct reports rune is punctuation or not.
func IsPunct(r rune) bool {
	return r == '!' ||
		r == '#' ||
		r == '$' ||
		r == ',' ||
		r == '.' ||
		r == '\'' ||
		r == '"' ||
		r == ':' ||
		r == ';' ||
		r == '<' ||
		r == '>' ||
		r == '=' ||
		r == '?' ||
		r == '-' ||
		r == '+' ||
		r == '*' ||
		r == '(' ||
		r == ')' ||
		r == '[' ||
		r == ']' ||
		r == '{' ||
		r == '}' ||
		r == '%' ||
		r == '&' ||
		r == '/' ||
		r == '\\' ||
		r == '@' ||
		r == '^' ||
		r == '_' ||
		r == '`' ||
		r == '|' ||
		r == '~' ||
		r == '¦'
}

// IsLetter reports rune is letter or not.
func IsLetter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

// iskw returns true if part is keyword, false if not.
func iskw(ln, kw string) bool {
	if !strings.HasPrefix(ln, kw) {
		return false
	}
	ln = ln[len(kw):]
	if ln == "" {
		return true
	}
	r, _ := utf8.DecodeRuneInString(ln)
	if r == '_' {
		return false
	}
	return IsSpace(byte(r)) || IsPunct(r) || !IsLetter(r)
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

// id returns identifer if next token is identifer,
// returns empty string if not.
func (l *Lex) id(ln string) string {
	if !IsIdentifierRune(ln) {
		return ""
	}
	var sb strings.Builder
	for _, r := range ln {
		if r != '_' &&
			!IsDecimal(byte(r)) &&
			!IsLetter(r) {
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
		if IsSpace(byte(r)) {
			l.Pos++
			switch r {
			case '\n':
				l.Newln()
			case '\t':
				l.Column += 4
			default:
				l.Column++
			}
			continue
		}
		ln = string(runes[i:])
		break
	}
	return ln
}

func (l *Lex) lncomment(t *Token) {
	start := l.Pos
	l.Pos += 2
	for ; l.Pos < len(l.File.Data); l.Pos++ {
		if l.File.Data[l.Pos] == '\n' {
			if l.firstTokenOfLine {
				t.Id = tokens.Comment
				t.Kind = string(l.File.Data[start:l.Pos])
			}
			return
		}
	}
	if l.firstTokenOfLine {
		t.Id = tokens.Comment
		t.Kind = string(l.File.Data[start:])
	}
}

func (l *Lex) rangecomment() {
	l.Pos += 2
	for ; l.Pos < len(l.File.Data); l.Pos++ {
		r := l.File.Data[l.Pos]
		if r == '\n' {
			l.Newln()
			continue
		}
		l.Column += len(string(r))
		if strings.HasPrefix(string(l.File.Data[l.Pos:]), tokens.RANGE_COMMENT_CLOSE) {
			l.Column += 2
			l.Pos += 2
			return
		}
	}
	l.pusherr("missing_block_comment")
}

func float_fmt_e(txt string, i int) (literal string) {
	i++ // Skip E | e
	if i >= len(txt) {
		return
	}
	b := txt[i]
	if b == '+' || b == '-' {
		i++ // Skip operator
		if i >= len(txt) {
			return
		}
	}
	first := i
	for ; i < len(txt); i++ {
		b := txt[i]
		if !IsDecimal(b) {
			break
		}
	}
	if i == first {
		return ""
	}
	return txt[:i]
}

func float_fmt_p(txt string, i int) string {
	return float_fmt_e(txt, i)
}

func float_fmt_dotnp(txt string, i int) string {
	if txt[i] != '.' {
		return ""
	}
loop:
	for i++; i < len(txt); i++ {
		b := txt[i]
		switch {
		case IsDecimal(b):
			continue
		case is_float_fmt_p(b, i):
			return float_fmt_p(txt, i)
		default:
			break loop
		}
	}
	return ""
}

func float_fmt_dotfp(txt string, i int) string {
	// skip .f
	i += 2
	return float_fmt_e(txt, i)
}

func float_fmt_dotp(txt string, i int) string {
	// skip .
	i++
	return float_fmt_e(txt, i)
}

func floatNum(txt string, i int) (literal string) {
	i++ // Skip dot
	for ; i < len(txt); i++ {
		b := txt[i]
		if i > 1 && (b == 'e' || b == 'E') {
			return float_fmt_e(txt, i)
		} else if !IsDecimal(b) {
			break
		}
	}
	if i == 1 { // Just dot
		return
	}
	return txt[:i]
}

func commonNum(txt string) (literal string) {
	i := 0
loop:
	for ; i < len(txt); i++ {
		b := txt[i]
		switch {
		case b == '.':
			return floatNum(txt, i)
		case is_float_fmt_e(b, i):
			return float_fmt_e(txt, i)
		case !IsDecimal(b):
			break loop
		}
	}
	if i == 0 {
		return
	}
	return txt[:i]
}

func binaryNum(txt string) (literal string) {
	if !strings.HasPrefix(txt, "0b") {
		return ""
	}
	if len(txt) < 2 {
		return
	}
	const binaryStart = 2
	i := binaryStart
	for ; i < len(txt); i++ {
		if !IsBinary(txt[i]) {
			break
		}
	}
	if i == binaryStart {
		return
	}
	return txt[:i]
}

func is_float_fmt_e(b byte, i int) bool { return i > 0 && (b == 'e' || b == 'E') }
func is_float_fmt_p(b byte, i int) bool { return i > 0 && (b == 'p' || b == 'P') }

func is_float_fmt_dotnp(txt string, i int) bool {
	if txt[i] != '.' {
		return false
	}
loop:
	for i++; i < len(txt); i++ {
		b := txt[i]
		switch {
		case IsDecimal(b):
			continue
		case is_float_fmt_p(b, i):
			return true
		default:
			break loop
		}
	}
	return false
}

func is_float_fmt_dotp(txt string, i int) bool {
	txt = txt[i:]
	switch {
	case len(txt) < 3:
		fallthrough
	case txt[0] != '.':
		fallthrough
	case txt[1] != 'p' && txt[1] != 'P':
		return false
	default:
		return true
	}
}

func is_float_fmt_dotfp(txt string, i int) bool {
	txt = txt[i:]
	switch {
	case len(txt) < 4:
		fallthrough
	case txt[0] != '.':
		fallthrough
	case txt[1] != 'f' && txt[1] != 'F':
		fallthrough
	case txt[2] != 'p' && txt[1] != 'P':
		return false
	default:
		return true
	}
}

func octalNum(txt string) (literal string) {
	if txt[0] != '0' {
		return ""
	}
	if len(txt) < 2 {
		return
	}
	const octalStart = 1
	i := octalStart
	for ; i < len(txt); i++ {
		b := txt[i]
		if is_float_fmt_e(b, i) {
			return float_fmt_e(txt, i)
		} else if !IsOctal(b) {
			break
		}
	}
	if i == octalStart {
		return
	}
	return txt[:i]
}

func hexNum(txt string) (literal string) {
	if len(txt) < 3 {
		return
	} else if txt[0] != '0' || (txt[1] != 'x' && txt[1] != 'X') {
		return
	}
	const hexStart = 2
	i := hexStart
loop:
	for ; i < len(txt); i++ {
		b := txt[i]
		switch {
		case is_float_fmt_dotp(txt, i):
			return float_fmt_dotp(txt, i)
		case is_float_fmt_dotfp(txt, i):
			return float_fmt_dotfp(txt, i)
		case is_float_fmt_p(b, i):
			return float_fmt_p(txt, i)
		case is_float_fmt_dotnp(txt, i):
			return float_fmt_dotnp(txt, i)
		case !IsHex(b):
			break loop
		}
	}
	if i == hexStart {
		return
	}
	return txt[:i]
}

// num returns literal if next token is numeric,
// returns empty string if not.
func (l *Lex) num(txt string) (literal string) {
	literal = hexNum(txt)
	if literal != "" {
		goto end
	}
	literal = octalNum(txt)
	if literal != "" {
		goto end
	}
	literal = binaryNum(txt)
	if literal != "" {
		goto end
	}
	literal = commonNum(txt)
end:
	l.Pos += len(literal)
	return
}

// IsSpace reports byte is whitespace or not.
func IsSpace(b byte) bool {
	return b == ' ' ||
		b == '\t' ||
		b == '\v' ||
		b == '\r' ||
		b == '\n'
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
	}
	return false
}

func hexEsq(txt string, n int) (seq string) {
	if len(txt) < n {
		return
	}
	const hexStart = 2
	for i := hexStart; i < n; i++ {
		if !IsHex(txt[i]) {
			return
		}
	}
	seq = txt[:n]
	return
}

// Pattern (RegEx): ^\\U.{8}
func bigUnicodePointEsq(txt string) string { return hexEsq(txt, 10) }

// Pattern (RegEx): ^\\u.{4}
func littleUnicodePointEsq(txt string) string { return hexEsq(txt, 6) }

// Pattern (RegEx): ^\\x..
func hexByteEsq(txt string) string { return hexEsq(txt, 4) }

// Patter (RegEx): ^\\[0-7]{3}
func byteEsq(txt string) (seq string) {
	if len(txt) < 4 {
		return
	} else if !IsOctal(txt[1]) || !IsOctal(txt[2]) || !IsOctal(txt[3]) {
		return
	}
	return txt[:4]
}

func (l *Lex) escseq(txt string) string {
	seq := ""
	if len(txt) < 2 {
		goto end
	}
	switch txt[1] {
	case '\\', '\'', '"', 'a', 'b', 'f', 'n', 'r', 't', 'v':
		l.Pos += 2
		return txt[:2]
	case 'U':
		seq = bigUnicodePointEsq(txt)
	case 'u':
		seq = littleUnicodePointEsq(txt)
	case 'x':
		seq = hexByteEsq(txt)
	default:
		seq = byteEsq(txt)
	}
end:
	if seq == "" {
		l.Pos++
		l.pusherr("invalid_escape_sequence")
		return ""
	}
	l.Pos += len(seq)
	return seq
}

func (l *Lex) getrune(txt string, raw bool) string {
	if !raw && txt[0] == '\\' {
		return l.escseq(txt)
	}
	r, _ := utf8.DecodeRuneInString(txt)
	l.Pos++
	return string(r)
}

func (l *Lex) rune(txt string) string {
	var sb strings.Builder
	sb.WriteByte('\'')
	l.Column++
	txt = txt[1:]
	n := 0
	for i := 0; i < len(txt); i++ {
		if txt[i] == '\n' {
			l.pusherr("missing_rune_end")
			l.Pos++
			l.Newln()
			return ""
		}
		r := l.getrune(txt[i:], false)
		sb.WriteString(r)
		length := len(r)
		l.Column += length
		if r == "'" {
			l.Pos++
			break
		}
		if length > 1 {
			i += length - 1
		}
		n++
	}
	if n == 0 {
		l.pusherr("rune_empty")
	} else if n > 1 {
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
		r := l.getrune(txt[i:], raw)
		sb.WriteString(r)
		n := len(r)
		l.Column += n
		if ch == mark {
			l.Pos++
			break
		}
		if n > 1 {
			i += n - 1
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

func (l *Lex) isop(txt, kind string, id uint8, t *Token) bool {
	if !strings.HasPrefix(txt, kind) {
		return false
	}
	t.Kind = kind
	t.Id = id
	l.Pos += len([]rune(kind))
	return true
}

func (l *Lex) iskw(txt, kind string, id uint8, t *Token) bool {
	if !iskw(txt, kind) {
		return false
	}
	t.Kind = kind
	t.Id = id
	l.Pos += len([]rune(kind))
	return true
}

//               [keyword]id
var keywords = map[string]uint8{
	tokens.I8:          tokens.DataType,
	tokens.I16:         tokens.DataType,
	tokens.I32:         tokens.DataType,
	tokens.I64:         tokens.DataType,
	tokens.U8:          tokens.DataType,
	tokens.U16:         tokens.DataType,
	tokens.U32:         tokens.DataType,
	tokens.U64:         tokens.DataType,
	tokens.F32:         tokens.DataType,
	tokens.F64:         tokens.DataType,
	tokens.UINT:        tokens.DataType,
	tokens.INT:         tokens.DataType,
	tokens.UINTPTR:     tokens.DataType,
	tokens.BOOL:        tokens.DataType,
	tokens.STR:         tokens.DataType,
	tokens.ANY:         tokens.DataType,
	tokens.TRUE:        tokens.Value,
	tokens.FALSE:       tokens.Value,
	tokens.NIL:         tokens.Value,
	tokens.CONST:       tokens.Const,
	tokens.RET:         tokens.Ret,
	tokens.TYPE:        tokens.Type,
	tokens.FOR:         tokens.For,
	tokens.BREAK:       tokens.Break,
	tokens.CONTINUE:    tokens.Continue,
	tokens.IN:          tokens.In,
	tokens.IF:          tokens.If,
	tokens.ELSE:        tokens.Else,
	tokens.USE:         tokens.Use,
	tokens.PUB:         tokens.Pub,
	tokens.DEFER:       tokens.Defer,
	tokens.GOTO:        tokens.Goto,
	tokens.ENUM:        tokens.Enum,
	tokens.STRUCT:      tokens.Struct,
	tokens.CO:          tokens.Co,
	tokens.MATCH:       tokens.Match,
	tokens.CASE:        tokens.Case,
	tokens.DEFAULT:     tokens.Default,
	tokens.SELF:        tokens.Self,
	tokens.TRAIT:       tokens.Trait,
	tokens.IMPL:        tokens.Impl,
	tokens.CPP:         tokens.Cpp,
	tokens.FALLTHROUGH: tokens.Fallthrough,
	tokens.FN:          tokens.Fn,
	tokens.LET:         tokens.Let,
	tokens.UNSAFE:      tokens.Unsafe,
	tokens.MUT:         tokens.Mut,
}

type oppair struct {
	op string
	id uint8
}

var basicOps = [...]oppair{
	0:  {tokens.DOUBLE_COLON, tokens.DoubleColon},
	1:  {tokens.COLON, tokens.Colon},
	2:  {tokens.SEMICOLON, tokens.SemiColon},
	3:  {tokens.COMMA, tokens.Comma},
	4:  {tokens.TRIPLE_DOT, tokens.Operator},
	5:  {tokens.DOT, tokens.Dot},
	6:  {tokens.PLUS_EQUAL, tokens.Operator},
	7:  {tokens.MINUS_EQUAL, tokens.Operator},
	8:  {tokens.STAR_EQUAL, tokens.Operator},
	9:  {tokens.SLASH_EQUAL, tokens.Operator},
	10: {tokens.PERCENT_EQUAL, tokens.Operator},
	11: {tokens.LSHIFT_EQUAL, tokens.Operator},
	12: {tokens.RSHIFT_EQUAL, tokens.Operator},
	13: {tokens.CARET_EQUAL, tokens.Operator},
	14: {tokens.AMPER_EQUAL, tokens.Operator},
	15: {tokens.VLINE_EQUAL, tokens.Operator},
	16: {tokens.EQUALS, tokens.Operator},
	17: {tokens.NOT_EQUALS, tokens.Operator},
	18: {tokens.GREAT_EQUAL, tokens.Operator},
	19: {tokens.LESS_EQUAL, tokens.Operator},
	20: {tokens.DOUBLE_AMPER, tokens.Operator},
	21: {tokens.DOUBLE_VLINE, tokens.Operator},
	22: {tokens.LSHIFT, tokens.Operator},
	23: {tokens.RSHIFT, tokens.Operator},
	24: {tokens.DOUBLE_PLUS, tokens.Operator},
	25: {tokens.DOUBLE_MINUS, tokens.Operator},
	26: {tokens.PLUS, tokens.Operator},
	27: {tokens.MINUS, tokens.Operator},
	28: {tokens.STAR, tokens.Operator},
	29: {tokens.SOLIDUS, tokens.Operator},
	30: {tokens.PERCENT, tokens.Operator},
	31: {tokens.AMPER, tokens.Operator},
	32: {tokens.VLINE, tokens.Operator},
	33: {tokens.CARET, tokens.Operator},
	34: {tokens.EXCLAMATION, tokens.Operator},
	35: {tokens.LESS, tokens.Operator},
	36: {tokens.GREAT, tokens.Operator},
	37: {tokens.EQUAL, tokens.Operator},
}

func (l *Lex) lexKeywords(txt string, tok *Token) bool {
	for k, v := range keywords {
		if l.iskw(txt, k, v, tok) {
			return true
		}
	}
	return false
}

func (l *Lex) lexBasicOps(txt string, tok *Token) bool {
	for _, pair := range basicOps {
		if l.isop(txt, pair.op, pair.id, tok) {
			return true
		}
	}
	return false
}

func (l *Lex) lexIdentifier(txt string, t *Token) bool {
	lex := l.id(txt)
	if lex == "" {
		return false
	}
	t.Kind = lex
	t.Id = tokens.Id
	return true
}

func (l *Lex) lexNumeric(txt string, t *Token) bool {
	lex := l.num(txt)
	if lex == "" {
		return false
	}
	t.Kind = lex
	t.Id = tokens.Value
	return true
}

// lex.Token generates next token from resume at position.
func (l *Lex) Token() Token {
	defer func() { l.firstTokenOfLine = false }()

	t := Token{File: l.File, Id: tokens.NA}

	txt := l.resume()
	if txt == "" {
		return t
	}

	// Set token values.
	t.Column = l.Column
	t.Row = l.Row

	//* lex.Tokenenize
	switch {
	case l.lexNumeric(txt, &t):
	case txt[0] == '\'':
		t.Kind = l.rune(txt)
		t.Id = tokens.Value
		return t
	case txt[0] == '"' || txt[0] == '`':
		t.Kind = l.str(txt)
		t.Id = tokens.Value
		return t
	case strings.HasPrefix(txt, tokens.LINE_COMMENT):
		l.lncomment(&t)
		return t
	case strings.HasPrefix(txt, tokens.RANGE_COMMENT_OPEN):
		l.rangecomment()
		return t
	case l.isop(txt, tokens.LPARENTHESES, tokens.Brace, &t):
		l.braces = append(l.braces, t)
	case l.isop(txt, tokens.RPARENTHESES, tokens.Brace, &t):
		l.pushRangeClose(t, tokens.LPARENTHESES)
	case l.isop(txt, tokens.LBRACE, tokens.Brace, &t):
		l.braces = append(l.braces, t)
	case l.isop(txt, tokens.RBRACE, tokens.Brace, &t):
		l.pushRangeClose(t, tokens.LBRACE)
	case l.isop(txt, tokens.LBRACKET, tokens.Brace, &t):
		l.braces = append(l.braces, t)
	case l.isop(txt, tokens.RBRACKET, tokens.Brace, &t):
		l.pushRangeClose(t, tokens.LBRACKET)
	case
		l.lexBasicOps(txt, &t) ||
		l.lexKeywords(txt, &t) ||
		l.lexIdentifier(txt, &t):
	default:
		r, sz := utf8.DecodeRuneInString(txt)
		l.pusherr("invalid_token", r)
		l.Column += sz
		l.Pos++
		return t
	}
	l.Column += len(t.Kind)
	return t
}

func getCloseKindOfBrace(left string) string {
	var right string
	switch left {
	case tokens.RPARENTHESES:
		right = tokens.LPARENTHESES
	case tokens.RBRACE:
		right = tokens.LBRACE
	case tokens.RBRACKET:
		right = tokens.LBRACKET
	}
	return right
}

func (l *Lex) rmrange(i int, kind string) {
	close := getCloseKindOfBrace(kind)
	for ; i >= 0; i-- {
		tok := l.braces[i]
		if tok.Kind != close {
			continue
		}
		l.braces = append(l.braces[:i], l.braces[i+1:]...)
		break
	}
}

func (l *Lex) pushRangeClose(t Token, left string) {
	n := len(l.braces)
	if n == 0 {
		switch t.Kind {
		case tokens.RBRACKET:
			l.pusherrtok(t, "extra_closed_brackets")
		case tokens.RBRACE:
			l.pusherrtok(t, "extra_closed_braces")
		case tokens.RPARENTHESES:
			l.pusherrtok(t, "extra_closed_parentheses")
		}
		return
	} else if l.braces[n-1].Kind != left {
		l.pushWrongOrderCloseErr(t)
	}
	l.rmrange(n-1, t.Kind)
}

func (l *Lex) pushWrongOrderCloseErr(t Token) {
	var msg string
	switch l.braces[len(l.braces)-1].Kind {
	case tokens.LPARENTHESES:
		msg = "expected_parentheses_close"
	case tokens.LBRACE:
		msg = "expected_brace_close"
	case tokens.LBRACKET:
		msg = "expected_bracket_close"
	}
	l.pusherrtok(t, msg)
}
