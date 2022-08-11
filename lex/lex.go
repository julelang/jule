package lex

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/jule"
	"github.com/jule-lang/jule/pkg/juleio"
	"github.com/jule-lang/jule/pkg/julelog"
)

type File = juleio.File

// Lex is lexer of Fract.
type Lex struct {
	firstTokOfLine bool

	File   *File
	Pos    int
	Column int
	Row    int
	// Logs are only errors
	Logs []julelog.CompilerLog

	braces []Tok
}

// New Lex instance.
func NewLex(f *File) *Lex {
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

func (l *Lex) pusherrtok(tok Tok, err string) {
	l.Logs = append(l.Logs, julelog.CompilerLog{
		Type:    julelog.Error,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    l.File.Path(),
		Message: jule.GetError(err),
	})
}

// Lex all source content.
func (l *Lex) Lex() []Tok {
	var toks []Tok
	l.Logs = nil
	l.Newln()
	for l.Pos < len(l.File.Data) {
		tok := l.Tok()
		if tok.Id != tokens.NA {
			toks = append(toks, tok)
		}
	}
	l.checkRanges()
	return toks
}

func (l *Lex) checkRanges() {
	for _, tok := range l.braces {
		switch tok.Kind {
		case tokens.LPARENTHESES:
			l.pusherrtok(tok, "wait_close_parentheses")
		case tokens.LBRACE:
			l.pusherrtok(tok, "wait_close_brace")
		case tokens.LBRACKET:
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
	if ln == "" {
		return true
	}
	r, _ := utf8.DecodeRuneInString(ln)
	if r == '_' {
		return false
	}
	return unicode.IsSpace(r) ||
		unicode.IsPunct(r) ||
		!unicode.IsLetter(r)
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
	runes := l.File.Data[l.Pos:]
	// Skip spaces.
	for i, r := range runes {
		if unicode.IsSpace(r) {
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

func (l *Lex) lncomment(tok *Tok) {
	start := l.Pos
	l.Pos += 2
	for ; l.Pos < len(l.File.Data); l.Pos++ {
		if l.File.Data[l.Pos] == '\n' {
			if l.firstTokOfLine {
				tok.Id = tokens.Comment
				tok.Kind = string(l.File.Data[start:l.Pos])
			}
			return
		}
	}
	if l.firstTokOfLine {
		tok.Id = tokens.Comment
		tok.Kind = string(l.File.Data[start:])
	}
}

func (l *Lex) rangecomment() {
	l.Pos += 2
	for ; l.Pos < len(l.File.Data); l.Pos++ {
		run := l.File.Data[l.Pos]
		if run == '\n' {
			l.Newln()
			continue
		}
		l.Column += len(string(run))
		if strings.HasPrefix(string(l.File.Data[l.Pos:]), tokens.RANGE_COMMENT_CLOSE) {
			l.Column += 2
			l.Pos += 2
			return
		}
	}
	l.pusherr("missing_block_comment")
}

func scientific(txt string, i int) (literal string) {
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
	for ; i < len(txt); i++ {
		b := txt[i]
		if !IsDecimal(b) {
			break
		}
	}
	return txt[:i]
}

func floatNum(txt string, i int) (literal string) {
	i++ // Skip dot
	for ; i < len(txt); i++ {
		b := txt[i]
		if b == 'e' || b == 'E' {
			return scientific(txt, i)
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
	for ; i < len(txt); i++ {
		b := txt[i]
		if b == '.' {
			return floatNum(txt, i)
		} else if isScientific(b, i) {
			return scientific(txt, i)
		} else if !IsDecimal(b) {
			break
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

func isScientific(b byte, i int) bool { return i > 0 && (b == 'e' || b == 'E') }

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
		if isScientific(b, i) {
			return scientific(txt, i)
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
	if !strings.HasPrefix(txt, "0x") {
		return ""
	}
	if len(txt) < 3 {
		return
	}
	const hexStart = 2
	i := hexStart
	for ; i < len(txt); i++ {
		if !IsHex(txt[i]) {
			break
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

// Pattern (RegEx): ^\\jule..
func hexByteEsq(txt string) string { return hexEsq(txt, 4) }

// Patter (RegEx): ^\\[0-7]{1,3}
func byteEsq(txt string) (seq string) {
	if len(txt) < 2 {
		return
	}
	if !IsOctal(txt[1]) {
		return
	}
	switch {
	case len(txt) == 2:
		seq = txt[:2]
	case !IsOctal(txt[2]):
		seq = txt[:2]
	case len(txt) == 3:
		seq = txt[:3]
	case !IsOctal(txt[3]):
		seq = txt[:3]
	default:
		seq = txt[:4]
	}
	return
}

func (l *Lex) escseq(txt string) string {
	seq := ""
	if len(txt) < 2 {
		goto end
	}
	switch txt[1] {
	case '\'', '"', 'a', 'b', 'f', 'n', 'r', 't', 'v':
		l.Pos += 2
		return txt[:3]
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

func (l *Lex) isop(txt, kind string, id uint8, tok *Tok) bool {
	if !strings.HasPrefix(txt, kind) {
		return false
	}
	tok.Kind = kind
	tok.Id = id
	l.Pos += len([]rune(kind))
	return true
}

func (l *Lex) iskw(txt, kind string, id uint8, tok *Tok) bool {
	if !iskw(txt, kind) {
		return false
	}
	tok.Kind = kind
	tok.Id = id
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
	4:  {tokens.AT, tokens.At},
	5:  {tokens.TRIPLE_DOT, tokens.Operator},
	6:  {tokens.DOT, tokens.Dot},
	7:  {tokens.PLUS_EQUAL, tokens.Operator},
	8:  {tokens.MINUS_EQUAL, tokens.Operator},
	9:  {tokens.STAR_EQUAL, tokens.Operator},
	10: {tokens.SLASH_EQUAL, tokens.Operator},
	11: {tokens.PERCENT_EQUAL, tokens.Operator},
	12: {tokens.LSHIFT_EQUAL, tokens.Operator},
	13: {tokens.RSHIFT_EQUAL, tokens.Operator},
	14: {tokens.CARET_EQUAL, tokens.Operator},
	15: {tokens.AMPER_EQUAL, tokens.Operator},
	16: {tokens.VLINE_EQUAL, tokens.Operator},
	17: {tokens.EQUALS, tokens.Operator},
	18: {tokens.NOT_EQUALS, tokens.Operator},
	19: {tokens.GREAT_EQUAL, tokens.Operator},
	20: {tokens.LESS_EQUAL, tokens.Operator},
	21: {tokens.AND, tokens.Operator},
	22: {tokens.OR, tokens.Operator},
	23: {tokens.LSHIFT, tokens.Operator},
	24: {tokens.RSHIFT, tokens.Operator},
	25: {tokens.DOUBLE_PLUS, tokens.Operator},
	26: {tokens.DOUBLE_MINUS, tokens.Operator},
	27: {tokens.PLUS, tokens.Operator},
	28: {tokens.MINUS, tokens.Operator},
	29: {tokens.STAR, tokens.Operator},
	30: {tokens.SOLIDUS, tokens.Operator},
	31: {tokens.PERCENT, tokens.Operator},
	32: {tokens.AMPER, tokens.Operator},
	33: {tokens.VLINE, tokens.Operator},
	34: {tokens.CARET, tokens.Operator},
	35: {tokens.EXCLAMATION, tokens.Operator},
	36: {tokens.LESS, tokens.Operator},
	37: {tokens.GREAT, tokens.Operator},
	38: {tokens.EQUAL, tokens.Operator},
}

func (l *Lex) lexKeywords(txt string, tok *Tok) bool {
	for key, value := range keywords {
		if l.iskw(txt, key, value, tok) {
			return true
		}
	}
	return false
}

func (l *Lex) lexBasicOps(txt string, tok *Tok) bool {
	for _, pair := range basicOps {
		if l.isop(txt, pair.op, pair.id, tok) {
			return true
		}
	}
	return false
}

func (l *Lex) lexIdentifier(txt string, tok *Tok) bool {
	lex := l.id(txt)
	if lex == "" {
		return false
	}
	tok.Kind = lex
	tok.Id = tokens.Id
	return true
}

func (l *Lex) lexNumeric(txt string, tok *Tok) bool {
	lex := l.num(txt)
	if lex == "" {
		return false
	}
	tok.Kind = lex
	tok.Id = tokens.Value
	return true
}

// Tok generates next token from resume at position.
func (l *Lex) Tok() Tok {
	defer func() { l.firstTokOfLine = false }()

	tok := Tok{File: l.File, Id: tokens.NA}

	txt := l.resume()
	if txt == "" {
		return tok
	}

	// Set token values.
	tok.Column = l.Column
	tok.Row = l.Row

	//* Tokenize
	switch {
	case l.lexNumeric(txt, &tok):
	case txt[0] == '\'':
		tok.Kind = l.rune(txt)
		tok.Id = tokens.Value
		return tok
	case txt[0] == '"', txt[0] == '`':
		tok.Kind = l.str(txt)
		tok.Id = tokens.Value
		return tok
	case strings.HasPrefix(txt, tokens.LINE_COMMENT):
		l.lncomment(&tok)
		return tok
	case strings.HasPrefix(txt, tokens.RANGE_COMMENT_OPEN):
		l.rangecomment()
		return tok
	case l.isop(txt, tokens.LPARENTHESES, tokens.Brace, &tok):
		l.braces = append(l.braces, tok)
	case l.isop(txt, tokens.RPARENTHESES, tokens.Brace, &tok):
		l.pushRangeClose(tok, tokens.LPARENTHESES)
	case l.isop(txt, tokens.LBRACE, tokens.Brace, &tok):
		l.braces = append(l.braces, tok)
	case l.isop(txt, tokens.RBRACE, tokens.Brace, &tok):
		l.pushRangeClose(tok, tokens.LBRACE)
	case l.isop(txt, tokens.LBRACKET, tokens.Brace, &tok):
		l.braces = append(l.braces, tok)
	case l.isop(txt, tokens.RBRACKET, tokens.Brace, &tok):
		l.pushRangeClose(tok, tokens.LBRACKET)
	case
		l.firstTokOfLine && l.isop(txt, tokens.SHARP, tokens.Preprocessor, &tok),
		l.lexBasicOps(txt, &tok),
		l.lexKeywords(txt, &tok),
		l.lexIdentifier(txt, &tok):
	default:
		r, sz := utf8.DecodeRuneInString(txt)
		l.pusherr("invalid_token", r)
		l.Column += sz
		l.Pos++
		return tok
	}
	l.Column += len(tok.Kind)
	return tok
}

func getCloseKindOfBrace(open string) string {
	var close string
	switch open {
	case tokens.RPARENTHESES:
		close = tokens.LPARENTHESES
	case tokens.RBRACE:
		close = tokens.LBRACE
	case tokens.RBRACKET:
		close = tokens.LBRACKET
	}
	return close
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

func (l *Lex) pushRangeClose(tok Tok, open string) {
	len := len(l.braces)
	if len == 0 {
		switch tok.Kind {
		case tokens.RBRACKET:
			l.pusherrtok(tok, "extra_closed_brackets")
		case tokens.RBRACE:
			l.pusherrtok(tok, "extra_closed_braces")
		case tokens.RPARENTHESES:
			l.pusherrtok(tok, "extra_closed_parentheses")
		}
		return
	} else if l.braces[len-1].Kind != open {
		l.pushWrongOrderCloseErr(tok)
	}
	l.rmrange(len-1, tok.Kind)
}

func (l *Lex) pushWrongOrderCloseErr(tok Tok) {
	var msg string
	switch l.braces[len(l.braces)-1].Kind {
	case tokens.LPARENTHESES:
		msg = "expected_parentheses_close"
	case tokens.LBRACE:
		msg = "expected_brace_close"
	case tokens.LBRACKET:
		msg = "expected_bracket_close"
	}
	l.pusherrtok(tok, msg)
}
