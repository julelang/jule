package ast

import "github.com/jule-lang/jule/lex"

// IsFuncCall returns function expressions without call expression
// if tokens are function call, nil if not.
func IsFuncCall(toks []lex.Token) []lex.Token {
	switch toks[0].Id {
	case lex.ID_BRACE, lex.ID_IDENT, lex.ID_DT:
	default:
		tok := toks[len(toks)-1]
		if tok.Id != lex.ID_BRACE && tok.Kind != lex.KND_RPARENT {
			return nil
		}
	}
	tok := toks[len(toks)-1]
	if tok.Id != lex.ID_BRACE || tok.Kind != lex.KND_RPARENT {
		return nil
	}
	brace_n := 0
	// Loops i >= 1 because expression must be has function expression at begin.
	// For this reason, ignore first token.
	for i := len(toks) - 1; i >= 1; i-- {
		tok := toks[i]
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_RPARENT:
				brace_n++
			case lex.KND_LPAREN:
				brace_n--
			}
			if brace_n == 0 {
				return toks[:i]
			}
		}
	}
	return nil
}

// BlockExpr returns expression tokens comes before block if exist, nil if not.
func BlockExpr(toks []lex.Token) (expr []lex.Token) {
	brace_n := 0
	for i, tok := range toks {
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_LBRACE:
				if brace_n > 0 {
					brace_n++
					break
				}
				return toks[:i]
			case lex.KND_LBRACKET, lex.KND_LPAREN:
				brace_n++
			default:
				brace_n--
			}
		}
	}
	return nil
}
