// Copyright 2021 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package lex

import (
	"strings"
	"unicode/utf8"

	"github.com/julelang/jule/build"
)

// [keyword]id
var _KEYWORDS = map[string]uint8{
	KND_I8:       ID_PRIM,
	KND_I16:      ID_PRIM,
	KND_I32:      ID_PRIM,
	KND_I64:      ID_PRIM,
	KND_U8:       ID_PRIM,
	KND_U16:      ID_PRIM,
	KND_U32:      ID_PRIM,
	KND_U64:      ID_PRIM,
	KND_F32:      ID_PRIM,
	KND_F64:      ID_PRIM,
	KND_UINT:     ID_PRIM,
	KND_INT:      ID_PRIM,
	KND_UINTPTR:  ID_PRIM,
	KND_BOOL:     ID_PRIM,
	KND_STR:      ID_PRIM,
	KND_ANY:      ID_PRIM,
	KND_TRUE:     ID_LIT,
	KND_FALSE:    ID_LIT,
	KND_NIL:      ID_LIT,
	KND_CONST:    ID_CONST,
	KND_RET:      ID_RET,
	KND_TYPE:     ID_TYPE,
	KND_ITER:     ID_FOR,
	KND_BREAK:    ID_BREAK,
	KND_CONTINUE: ID_CONTINUE,
	KND_IN:       ID_IN,
	KND_IF:       ID_IF,
	KND_ELSE:     ID_ELSE,
	KND_USE:      ID_USE,
	KND_PUB:      ID_PUB,
	KND_GOTO:     ID_GOTO,
	KND_ENUM:     ID_ENUM,
	KND_STRUCT:   ID_STRUCT,
	KND_CO:       ID_CO,
	KND_MATCH:    ID_MATCH,
	KND_SELF:     ID_SELF,
	KND_TRAIT:    ID_TRAIT,
	KND_IMPL:     ID_IMPL,
	KND_CPP:      ID_CPP,
	KND_FALL:     ID_FALL,
	KND_FN:       ID_FN,
	KND_LET:      ID_LET,
	KND_UNSAFE:   ID_UNSAFE,
	KND_MUT:      ID_MUT,
	KND_DEFER:    ID_DEFER,
}

type _OpPair struct {
	op string
	id uint8
}

var _BASIC_OPS = [...]_OpPair{
	{KND_DBLCOLON, ID_DBLCOLON},
	{KND_COLON, ID_COLON},
	{KND_SEMICOLON, ID_SEMICOLON},
	{KND_COMMA, ID_COMMA},
	{KND_TRIPLE_DOT, ID_OP},
	{KND_DOT, ID_DOT},
	{KND_PLUS_EQ, ID_OP},
	{KND_MINUS_EQ, ID_OP},
	{KND_STAR_EQ, ID_OP},
	{KND_SOLIDUS_EQ, ID_OP},
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

type _Lex struct {
	first_token_of_line bool
	ranges              []Token
	data                []rune
	file                *File
	pos                 int
	column              int
	row                 int
	errors              []build.Log
}

func make_err(row int, col int, f *File, key string, args ...any) build.Log {
	return build.Log{
		Type:   build.ERR,
		Row:    row,
		Column: col,
		Path:   f.Path(),
		Text:   build.Errorf(key, args...),
	}
}

func (l *_Lex) push_err(key string, args ...any) {
	l.errors = append(l.errors, make_err(l.row, l.column, l.file, key, args...))
}

func (l *_Lex) push_err_tok(tok Token, key string) {
	l.errors = append(l.errors, make_err(tok.Row, tok.Column, l.file, key))
}

// lexs all source content.
func (l *_Lex) lex() []Token {
	var toks []Token
	l.errors = nil
	l.new_line()
	for l.pos < len(l.data) {
		t := l.token()
		l.first_token_of_line = false
		if t.Id != ID_NA {
			toks = append(toks, t)
		}
	}
	l.check_ranges()
	l.data = nil
	return toks
}

func (l *_Lex) check_ranges() {
	for _, t := range l.ranges {
		switch t.Kind {
		case KND_LPAREN:
			l.push_err_tok(t, "wait_close_parentheses")

		case KND_LBRACE:
			l.push_err_tok(t, "wait_close_brace")

		case KND_LBRACKET:
			l.push_err_tok(t, "wait_close_bracket")
		}
	}
}

// is_kw returns true if part is keyword, false if not.
func is_kw(ln, kw string) bool {
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
	return Is_space(r) || Is_punct(r) || !Is_letter(r)
}

// id returns identifer if next token is identifer,
// returns empty string if not.
func (l *_Lex) id(ln string) string {
	if !Is_ident_rune(ln) {
		return ""
	}

	ident := ""
	for _, r := range ln {
		if r != '_' &&
			!Is_decimal(byte(r)) &&
			!Is_letter(r) {
			break
		}
		ident += string(r)
		l.pos++
	}

	return ident
}

// resume to lex from position.
func (l *_Lex) resume() string {
	var ln string
	runes := l.data[l.pos:]
	// Skip spaces.
	for i, r := range runes {
		if Is_space(r) {
			l.pos++
			switch r {
			case '\n':
				l.new_line()

			case '\t':
				l.column += 4

			default:
				l.column++
			}
			continue
		}
		ln = string(runes[i:])
		break
	}
	return ln
}

func (l *_Lex) lex_line_comment(t *Token) {
	start := l.pos
	l.pos += 2
	for ; l.pos < len(l.data); l.pos++ {
		if l.data[l.pos] == '\n' {
			if l.first_token_of_line {
				t.Id = ID_COMMENT
				t.Kind = string(l.data[start:l.pos])
			}
			return
		}
	}
	if l.first_token_of_line {
		t.Id = ID_COMMENT
		t.Kind = string(l.data[start:])
	}
}

func (l *_Lex) lex_range_comment() {
	l.pos += 2
	for ; l.pos < len(l.data); l.pos++ {
		r := l.data[l.pos]
		if r == '\n' {
			l.new_line()
			continue
		}
		l.column += len(string(r))
		if strings.HasPrefix(string(l.data[l.pos:]), KND_RNG_RCOMMENT) {
			l.column += 2
			l.pos += 2
			return
		}
	}
	l.push_err("missing_block_comment")
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
		if !Is_decimal(b) {
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
		case Is_decimal(b):
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

func float_num(txt string, i int) (literal string) {
	i++ // Skip dot
	for ; i < len(txt); i++ {
		b := txt[i]
		if i > 1 && (b == 'e' || b == 'E') {
			return float_fmt_e(txt, i)
		} else if !Is_decimal(b) {
			break
		}
	}

	if i == 1 { // Just dot
		return
	}
	return txt[:i]
}

func common_num(txt string) (literal string) {
	i := 0
loop:
	for ; i < len(txt); i++ {
		b := txt[i]
		switch {
		case b == '.':
			return float_num(txt, i)

		case is_float_fmt_e(b, i):
			return float_fmt_e(txt, i)

		case !Is_decimal(b):
			break loop
		}
	}

	if i == 0 {
		return
	}
	return txt[:i]
}

func binary_num(txt string) (literal string) {
	if !strings.HasPrefix(txt, "0b") {
		return ""
	}
	if len(txt) < 2 {
		return
	}
	const binaryStart = 2
	i := binaryStart
	for ; i < len(txt); i++ {
		if !Is_binary(txt[i]) {
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
		case Is_decimal(b):
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

func octal_num(txt string) (literal string) {
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
		} else if !Is_octal(b) {
			break
		}
	}
	if i == octalStart {
		return
	}
	return txt[:i]
}

func hex_num(txt string) (literal string) {
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

		case !Is_hex(b):
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
func (l *_Lex) num(txt string) (literal string) {
	literal = hex_num(txt)
	if literal != "" {
		goto end
	}
	literal = octal_num(txt)
	if literal != "" {
		goto end
	}
	literal = binary_num(txt)
	if literal != "" {
		goto end
	}
	literal = common_num(txt)
end:
	l.pos += len(literal)
	return
}

func hex_escape(txt string, n int) (seq string) {
	if len(txt) < n {
		return
	}
	const hexStart = 2
	for i := hexStart; i < n; i++ {
		if !Is_hex(txt[i]) {
			return
		}
	}
	seq = txt[:n]
	return
}

// Pattern (RegEx): ^\\U.{8}
func big_unicode_point_escape(txt string) string { return hex_escape(txt, 10) }

// Pattern (RegEx): ^\\u.{4}
func little_unicode_point_escape(txt string) string { return hex_escape(txt, 6) }

// Pattern (RegEx): ^\\x..
func hex_byte_escape(txt string) string { return hex_escape(txt, 4) }

// Patter (RegEx): ^\\[0-7]{3}
func byte_escape(txt string) (seq string) {
	if len(txt) < 4 {
		return
	} else if !Is_octal(txt[1]) || !Is_octal(txt[2]) || !Is_octal(txt[3]) {
		return
	}
	return txt[:4]
}

func (l *_Lex) escape_seq(txt string) string {
	seq := ""
	if len(txt) < 2 {
		goto end
	}
	switch txt[1] {
	case '\\', '\'', '"', 'a', 'b', 'f', 'n', 'r', 't', 'v':
		l.pos += 2
		return txt[:2]
	case 'U':
		seq = big_unicode_point_escape(txt)
	case 'u':
		seq = little_unicode_point_escape(txt)
	case 'x':
		seq = hex_byte_escape(txt)
	default:
		seq = byte_escape(txt)
	}
end:
	if seq == "" {
		l.pos++
		l.push_err("invalid_escape_sequence")
		return ""
	}
	l.pos += len(seq)
	return seq
}

func (l *_Lex) get_rune(txt string, raw bool) string {
	if !raw && txt[0] == '\\' {
		return l.escape_seq(txt)
	}

	r, _ := utf8.DecodeRuneInString(txt)
	l.pos++
	return string(r)
}

func (l *_Lex) lex_rune(txt string) string {
	run := "'"
	l.column++
	n := 0
	for i := 1; i < len(txt); i++ {
		if txt[i] == '\n' {
			l.push_err("missing_rune_end")
			l.pos++
			l.new_line()
			return ""
		}

		r := l.get_rune(txt[i:], false)
		run += r
		length := len(r)
		l.column += length
		if r == "'" {
			l.pos++
			break
		}
		if length > 1 {
			i += length - 1
		}
		n++
	}

	if n == 0 {
		l.push_err("rune_empty")
	} else if n > 1 {
		l.push_err("rune_overflow")
	}

	return run
}

func (l *_Lex) lex_str(txt string) string {
	s := ""
	mark := txt[0]
	raw := mark == '`'
	s += string(mark)
	l.column++

	for i := 1; i < len(txt); i++ {
		ch := txt[i]
		if ch == '\n' {
			l.new_line()
			if !raw {
				l.push_err("missing_string_end")
				l.pos++
				return ""
			}
		}
		r := l.get_rune(txt[i:], raw)
		s += r
		n := len(r)
		l.column += n
		if ch == mark {
			l.pos++
			break
		}
		if n > 1 {
			i += n - 1
		}
	}

	return s
}

func (l *_Lex) new_line() {
	l.first_token_of_line = true
	l.row++
	l.column = 1
}

func (l *_Lex) is_op(txt, kind string, id uint8, t *Token) bool {
	if !strings.HasPrefix(txt, kind) {
		return false
	}
	t.Kind = kind
	t.Id = id
	l.pos += len([]rune(kind))
	return true
}

func (l *_Lex) is_kw(txt, kind string, id uint8, t *Token) bool {
	if !is_kw(txt, kind) {
		return false
	}
	t.Kind = kind
	t.Id = id
	l.pos += len([]rune(kind))
	return true
}

func (l *_Lex) lex_kws(txt string, tok *Token) bool {
	for k, v := range _KEYWORDS {
		if l.is_kw(txt, k, v, tok) {
			return true
		}
	}
	return false
}

func (l *_Lex) lex_basic_ops(txt string, tok *Token) bool {
	for _, pair := range _BASIC_OPS {
		if l.is_op(txt, pair.op, pair.id, tok) {
			return true
		}
	}
	return false
}

func (l *_Lex) lex_id(txt string, t *Token) bool {
	lex := l.id(txt)
	if lex == "" {
		return false
	}

	t.Kind = lex
	t.Id = ID_IDENT
	return true
}

func (l *_Lex) lex_num(txt string, t *Token) bool {
	lex := l.num(txt)
	if lex == "" {
		return false
	}

	t.Kind = lex
	t.Id = ID_LIT
	return true
}

// lex.token generates next token from resume at position.
func (l *_Lex) token() Token {
	t := Token{File: l.file, Id: ID_NA}

	txt := l.resume()
	if txt == "" {
		return t
	}

	// Set token values.
	t.Column = l.column
	t.Row = l.row

	//* lex.Tokenenize
	switch {
	case l.lex_num(txt, &t):
	case txt[0] == '\'':
		t.Kind = l.lex_rune(txt)
		t.Id = ID_LIT
		return t

	case txt[0] == '"' || txt[0] == '`':
		t.Kind = l.lex_str(txt)
		t.Id = ID_LIT
		return t

	case strings.HasPrefix(txt, KND_LN_COMMENT):
		l.lex_line_comment(&t)
		return t

	case strings.HasPrefix(txt, KND_RNG_LCOMMENT):
		l.lex_range_comment()
		return t

	case l.is_op(txt, KND_LPAREN, ID_RANGE, &t):
		l.ranges = append(l.ranges, t)

	case l.is_op(txt, KND_RPARENT, ID_RANGE, &t):
		l.push_range_close(t, KND_LPAREN)

	case l.is_op(txt, KND_LBRACE, ID_RANGE, &t):
		l.ranges = append(l.ranges, t)

	case l.is_op(txt, KND_RBRACE, ID_RANGE, &t):
		l.push_range_close(t, KND_LBRACE)

	case l.is_op(txt, KND_LBRACKET, ID_RANGE, &t):
		l.ranges = append(l.ranges, t)

	case l.is_op(txt, KND_RBRACKET, ID_RANGE, &t):
		l.push_range_close(t, KND_LBRACKET)

	case l.lex_basic_ops(txt, &t) || l.lex_kws(txt, &t) || l.lex_id(txt, &t):
		// Skip.

	default:
		r, sz := utf8.DecodeRuneInString(txt)
		l.push_err("invalid_token", r)
		l.column += sz
		l.pos++
		return t
	}

	l.column += len(t.Kind)
	return t
}

func get_close_kind_of_brace(left string) string {
	switch left {
	case KND_RPARENT:
		return KND_LPAREN

	case KND_RBRACE:
		return KND_LBRACE

	case KND_RBRACKET:
		return KND_LBRACKET

	default:
		return ""
	}
}

func (l *_Lex) remove_range(i int, kind string) {
	close := get_close_kind_of_brace(kind)
	for ; i >= 0; i-- {
		tok := l.ranges[i]
		if tok.Kind != close {
			continue
		}
		l.ranges = append(l.ranges[:i], l.ranges[i+1:]...)
		break
	}
}

func (l *_Lex) push_range_close(t Token, left string) {
	n := len(l.ranges)
	if n == 0 {
		switch t.Kind {
		case KND_RBRACKET:
			l.push_err_tok(t, "extra_closed_brackets")

		case KND_RBRACE:
			l.push_err_tok(t, "extra_closed_braces")

		case KND_RPARENT:
			l.push_err_tok(t, "extra_closed_parentheses")
		}
		return
	} else if l.ranges[n-1].Kind != left {
		l.push_wrong_order_close_err(t)
	}
	l.remove_range(n-1, t.Kind)
}

func (l *_Lex) push_wrong_order_close_err(t Token) {
	var msg string
	switch l.ranges[len(l.ranges)-1].Kind {
	case KND_LPAREN:
		msg = "expected_parentheses_close"

	case KND_LBRACE:
		msg = "expected_brace_close"

	case KND_LBRACKET:
		msg = "expected_bracket_close"
	}

	l.push_err_tok(t, msg)
}

// Lex source code into fileset.
// Returns nil if f is nil.
// Returns nil slice for errors if no any error.
func Lex(f *File, text string) []build.Log {
	if f == nil {
		return nil
	}

	lex := _Lex{
		file: f,
		pos:  0,
		row:  -1, // For true row
		data: ([]rune)(text),
	}
	
	lex.new_line()
	tokens := lex.lex()
	
	if len(lex.errors) > 0 {
		return lex.errors
	}

	f.tokens = tokens
	return nil
}
