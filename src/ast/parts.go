package ast

import (
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

func compilerErr(t lex.Token, key string, args ...any) build.Log {
	return build.Log{
		Type:    build.ERR,
		Row:     t.Row,
		Column:  t.Column,
		Path:    t.File.Path(),
		Text: build.Errorf(key, args...),
	}
}

// Range returns between of open and close braces.
//
// Special case is:
//  Range(i, open, close, toks) = nil if *i > len(toks)
//  Range(i, open, close, toks) = bil if (toks[i*]) Id != tokens.Brace && Kind != open
func Range(i *int, open, close string, toks []lex.Token) []lex.Token {
	if *i >= len(toks) {
		return nil
	}
	tok := toks[*i]
	if tok.Id != lex.ID_BRACE || tok.Kind != open {
		return nil
	}
	*i++
	brace_n := 1
	start := *i
	for ; brace_n != 0 && *i < len(toks); *i++ {
		tok := toks[*i]
		if tok.Id != lex.ID_BRACE {
			continue
		}
		switch tok.Kind {
		case open:
			brace_n++
		case close:
			brace_n--
		}
	}
	return toks[start : *i-1]
}

// RangeLast returns last range from tokens.
//
// Special cases are;
//  RangeLast(toks) = toks, nil if len(toks) == 0
//  RangeLast(toks) = toks, nil if toks is not has range at last
func RangeLast(toks []lex.Token) (cutted, cut []lex.Token) {
	if len(toks) == 0 {
		return toks, nil
	} else if toks[len(toks)-1].Id != lex.ID_BRACE {
		return toks, nil
	}
	brace_n := 0
	for i := len(toks) - 1; i >= 0; i-- {
		tok := toks[i]
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_RBRACE, lex.KND_RBRACKET, lex.KND_RPARENT:
				brace_n++
				continue
			default:
				brace_n--
			}
		}
		if brace_n == 0 {
			return toks[:i], toks[i:]
		}
	}
	return toks, nil
}

// Parts returns parts separated by given token identifier.
// It's skips parentheses ranges.
//
// Special case is;
//  Parts(toks) = nil if len(toks) == 0
func Parts(toks []lex.Token, id uint8, exprMust bool) ([][]lex.Token, []build.Log) {
	if len(toks) == 0 {
		return nil, nil
	}
	var parts [][]lex.Token
	var errs []build.Log
	brace_n := 0
	last := 0
	for i, tok := range toks {
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				brace_n++
				continue
			default:
				brace_n--
			}
		}
		if brace_n > 0 {
			continue
		}
		if tok.Id == id {
			if exprMust && i-last <= 0 {
				errs = append(errs, compilerErr(tok, "missing_expr"))
			}
			parts = append(parts, toks[last:i])
			last = i + 1
		}
	}
	if last < len(toks) {
		parts = append(parts, toks[last:])
	} else if !exprMust {
		parts = append(parts, []lex.Token{})
	}
	return parts, errs
}

// SplitColon returns colon index and range tokens.
// Starts at i.
func SplitColon(toks []lex.Token, i *int) (rangeToks []lex.Token, colon int) {
	rangeToks = nil
	colon = -1
	brace_n := 0
	start := *i
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				brace_n++
				continue
			default:
				brace_n--
			}
		}
		if brace_n == 0 {
			if start+1 > *i {
				return
			}
			rangeToks = toks[start+1 : *i]
			break
		} else if brace_n != 1 {
			continue
		}
		if colon == -1 && tok.Id == lex.ID_COLON {
			colon = *i - start - 1
		}
	}
	return
}
