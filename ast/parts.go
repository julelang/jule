package ast

import (
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/xlog"
)

// Range returns between of open and close braces.
//
// Special case is:
//  Range(i, open, close, toks) = nil if *i > len(toks)
//  Range(i, open, close, toks) = bil if (toks[i*]) Id != tokens.Brace && Kind != open
func Range(i *int, open, close string, toks Toks) Toks {
	if *i >= len(toks) {
		return nil
	}
	tok := toks[*i]
	if tok.Id == tokens.Brace && tok.Kind == open {
		*i++
		braceCount := 1
		start := *i
		for ; braceCount != 0 && *i < len(toks); *i++ {
			tok := toks[*i]
			if tok.Id != tokens.Brace {
				continue
			}
			switch tok.Kind {
			case open:
				braceCount++
			case close:
				braceCount--
			}
		}
		return toks[start : *i-1]
	}
	return nil
}

// RangeLast returns last range from tokens.
//
// Special cases are;
//  RangeLast(toks) = toks, nil if len(toks) == 0
//  RangeLast(toks) = toks, nil if toks is not has range at last
func RangeLast(toks Toks) (cutted, cut Toks) {
	if len(toks) == 0 {
		return toks, nil
	} else if toks[len(toks)-1].Id != tokens.Brace {
		return toks, nil
	}
	braceCount := 0
	for i := len(toks) - 1; i >= 0; i-- {
		tok := toks[i]
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.RBRACE, tokens.RBRACKET, tokens.RPARENTHESES:
				braceCount++
				continue
			default:
				braceCount--
			}
		}
		if braceCount == 0 {
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
func Parts(toks Toks, id uint8, exprMust bool) ([]Toks, []xlog.CompilerLog) {
	if len(toks) == 0 {
		return nil, nil
	}
	parts := make([]Toks, 0)
	errs := make([]xlog.CompilerLog, 0)
	braceCount := 0
	last := 0
	for i, tok := range toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
				continue
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
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
		parts = append(parts, Toks{})
	}
	return parts, errs
}

// SplitColon returns colon index and range tokens.
// Starts at i.
func SplitColon(toks Toks, i *int) (rangeToks Toks, colon int) {
	rangeToks = nil
	colon = -1
	braceCount := 0
	start := *i
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
				continue
			default:
				braceCount--
			}
		}
		if braceCount == 0 {
			if start+1 > *i {
				return
			}
			rangeToks = toks[start+1 : *i]
			break
		} else if braceCount != 1 {
			continue
		}
		if colon == -1 && tok.Id == tokens.Colon {
			colon = *i - start - 1
		}
	}
	return
}
