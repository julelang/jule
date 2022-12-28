package lex

import (
	"os"
	"strings"
	"unicode/utf8"

	"github.com/julelang/jule/build"
)

// Lex is lexer of Jule.
type Lex struct {
	firstTokenOfLine bool
	braces           []Token
	data             []rune

	File   *File
	Pos    int
	Column int
	Row    int
	// Logs are only errors
	Logs []build.Log
}

// New Lex instance.
func New(f *File) *Lex {
	l := new(Lex)
	l.File = f
	l.Pos = 0
	l.Row = -1 // For true row
	l.Newln()
	return l
}

func (l *Lex) pusherr(key string, args ...any) {
	l.Logs = append(l.Logs, build.Log{
		Type:    build.ERR,
		Row:     l.Row,
		Column:  l.Column,
		Path:    l.File.Path(),
		Message: build.Errorf(key, args...),
	})
}

func (l *Lex) pusherrtok(tok Token, err string) {
	l.Logs = append(l.Logs, build.Log{
		Type:    build.ERR,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    l.File.Path(),
		Message: build.Errorf(err),
	})
}

func (l* Lex) buff_data() {
	bytes, err := os.ReadFile(l.File.Path())
	if err != nil {
		panic("buffering failed: " + err.Error())
	}
	l.data = []rune(string(bytes))
}

// Lex all source content.
func (l *Lex) Lex() []Token {
	l.buff_data()
	var toks []Token
	l.Logs = nil
	l.Newln()
	for l.Pos < len(l.data) {
		t := l.Token()
		l.firstTokenOfLine = false
		if t.Id != ID_NA {
			toks = append(toks, t)
		}
	}
	l.checkRanges()
	l.data = nil
	return toks
}

func (l *Lex) checkRanges() {
	for _, t := range l.braces {
		switch t.Kind {
		case KND_LPAREN:
			l.pusherrtok(t, "wait_close_parentheses")
		case KND_LBRACE:
			l.pusherrtok(t, "wait_close_brace")
		case KND_LBRACKET:
			l.pusherrtok(t, "wait_close_bracket")
		}
	}
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
	return IsSpace(r) || IsPunct(r) || !IsLetter(r)
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
	runes := l.data[l.Pos:]
	// Skip spaces.
	for i, r := range runes {
		if IsSpace(r) {
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
	for ; l.Pos < len(l.data); l.Pos++ {
		if l.data[l.Pos] == '\n' {
			if l.firstTokenOfLine {
				t.Id = ID_COMMENT
				t.Kind = string(l.data[start:l.Pos])
			}
			return
		}
	}
	if l.firstTokenOfLine {
		t.Id = ID_COMMENT
		t.Kind = string(l.data[start:])
	}
}

func (l *Lex) rangecomment() {
	l.Pos += 2
	for ; l.Pos < len(l.data); l.Pos++ {
		r := l.data[l.Pos]
		if r == '\n' {
			l.Newln()
			continue
		}
		l.Column += len(string(r))
		if strings.HasPrefix(string(l.data[l.Pos:]), KND_RNG_RCOMMENT) {
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

func float_fmt_p(txt string, i int) string { return float_fmt_e(txt, i) }

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
			l.Newln()
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
var KEYWORDS = map[string]uint8{
	KND_I8:          ID_DT,
	KND_I16:         ID_DT,
	KND_I32:         ID_DT,
	KND_I64:         ID_DT,
	KND_U8:          ID_DT,
	KND_U16:         ID_DT,
	KND_U32:         ID_DT,
	KND_U64:         ID_DT,
	KND_F32:         ID_DT,
	KND_F64:         ID_DT,
	KND_UINT:        ID_DT,
	KND_INT:         ID_DT,
	KND_UINTPTR:     ID_DT,
	KND_BOOL:        ID_DT,
	KND_STR:         ID_DT,
	KND_ANY:         ID_DT,
	KND_TRUE:        ID_LITERAL,
	KND_FALSE:       ID_LITERAL,
	KND_NIL:         ID_LITERAL,
	KND_CONST:       ID_CONST,
	KND_RET:         ID_RET,
	KND_TYPE:        ID_TYPE,
	KND_ITER:         ID_ITER,
	KND_BREAK:       ID_BREAK,
	KND_CONTINUE:    ID_CONTINUE,
	KND_IN:          ID_IN,
	KND_IF:          ID_IF,
	KND_ELSE:        ID_ELSE,
	KND_USE:         ID_USE,
	KND_PUB:         ID_PUB,
	KND_GOTO:        ID_GOTO,
	KND_ENUM:        ID_ENUM,
	KND_STRUCT:      ID_STRUCT,
	KND_CO:          ID_CO,
	KND_MATCH:       ID_MATCH,
	KND_CASE:        ID_CASE,
	KND_DEFAULT:     ID_DEFAULT,
	KND_SELF:        ID_SELF,
	KND_TRAIT:       ID_TRAIT,
	KND_IMPL:        ID_IMPL,
	KND_CPP:         ID_CPP,
	KND_FALLTHROUGH: ID_FALLTHROUGH,
	KND_FN:          ID_FN,
	KND_LET:         ID_LET,
	KND_UNSAFE:      ID_UNSAFE,
	KND_MUT:         ID_MUT,
	KND_DEFER:       ID_DEFER,
}

type oppair struct {
	op string
	id uint8
}

var BASIC_OPS = [...]oppair{
	{KND_DBLCOLON, ID_DBLCOLON},
	{KND_COLON, ID_COLON},
	{KND_SEMICOLON, ID_SEMICOLON},
	{KND_COMMA, ID_COMMA},
	{KND_TRIPLE_DOT, ID_OP},
	{KND_DOT, ID_DOT},
	{KND_PLUS_EQ, ID_OP},
	{KND_MINUS_EQ, ID_OP},
	{KND_STAR_EQ, ID_OP},
	{KND_SLASH_EQ, ID_OP},
	{KND_PERCENT_EQ, ID_OP},
	{KND_LSHIFT_EQ, ID_OP},
	{KND_RSHIFT_EQ, ID_OP},
	{KND_CARET_EQ, ID_OP},
	{KND_AMPER_EQ, ID_OP},
	{KND_VLINE_EQ, ID_OP},
	{KND_EQS, ID_OP},
	{KND_NOT_EQ, ID_OP},
	{KND_GREAT_EQ, ID_OP},
	{KND_LESS_EQ, ID_OP},
	{KND_DBL_AMPER, ID_OP},
	{KND_DBL_VLINE, ID_OP},
	{KND_LSHIFT, ID_OP},
	{KND_RSHIFT, ID_OP},
	{KND_DBL_PLUS, ID_OP},
	{KND_DBL_MINUS, ID_OP},
	{KND_PLUS, ID_OP},
	{KND_MINUS, ID_OP},
	{KND_STAR, ID_OP},
	{KND_SOLIDUS, ID_OP},
	{KND_PERCENT, ID_OP},
	{KND_AMPER, ID_OP},
	{KND_VLINE, ID_OP},
	{KND_CARET, ID_OP},
	{KND_EXCL, ID_OP},
	{KND_LT, ID_OP},
	{KND_GT, ID_OP},
	{KND_EQ, ID_OP},
}

func (l *Lex) lexKeywords(txt string, tok *Token) bool {
	for k, v := range KEYWORDS {
		if l.iskw(txt, k, v, tok) {
			return true
		}
	}
	return false
}

func (l *Lex) lexBasicOps(txt string, tok *Token) bool {
	for _, pair := range BASIC_OPS {
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
	t.Id = ID_IDENT
	return true
}

func (l *Lex) lexNumeric(txt string, t *Token) bool {
	lex := l.num(txt)
	if lex == "" {
		return false
	}
	t.Kind = lex
	t.Id = ID_LITERAL
	return true
}

// lex.Token generates next token from resume at position.
func (l *Lex) Token() Token {
	t := Token{File: l.File, Id: ID_NA}

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
		t.Id = ID_LITERAL
		return t
	case txt[0] == '"' || txt[0] == '`':
		t.Kind = l.str(txt)
		t.Id = ID_LITERAL
		return t
	case strings.HasPrefix(txt, KND_LN_COMMENT):
		l.lncomment(&t)
		return t
	case strings.HasPrefix(txt, KND_RNG_LCOMMENT):
		l.rangecomment()
		return t
	case l.isop(txt, KND_LPAREN, ID_BRACE, &t):
		l.braces = append(l.braces, t)
	case l.isop(txt, KND_RPARENT, ID_BRACE, &t):
		l.pushRangeClose(t, KND_LPAREN)
	case l.isop(txt, KND_LBRACE, ID_BRACE, &t):
		l.braces = append(l.braces, t)
	case l.isop(txt, KND_RBRACE, ID_BRACE, &t):
		l.pushRangeClose(t, KND_LBRACE)
	case l.isop(txt, KND_LBRACKET, ID_BRACE, &t):
		l.braces = append(l.braces, t)
	case l.isop(txt, KND_RBRACKET, ID_BRACE, &t):
		l.pushRangeClose(t, KND_LBRACKET)
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
	case KND_RPARENT:
		right = KND_LPAREN
	case KND_RBRACE:
		right = KND_LBRACE
	case KND_RBRACKET:
		right = KND_LBRACKET
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
		case KND_RBRACKET:
			l.pusherrtok(t, "extra_closed_brackets")
		case KND_RBRACE:
			l.pusherrtok(t, "extra_closed_braces")
		case KND_RPARENT:
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
	case KND_LPAREN:
		msg = "expected_parentheses_close"
	case KND_LBRACE:
		msg = "expected_brace_close"
	case KND_LBRACKET:
		msg = "expected_bracket_close"
	}
	l.pusherrtok(t, msg)
}
