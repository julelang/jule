package ast

import "github.com/the-xlang/xxc/lex/tokens"

// IsFuncCall returns function expressions without call expression
// if tokens are function call, nil if not.
func IsFuncCall(toks Toks) Toks {
	switch toks[0].Id {
	case tokens.Brace, tokens.Id, tokens.DataType:
	default:
		tok := toks[len(toks)-1]
		if tok.Id != tokens.Brace && tok.Kind != tokens.RPARENTHESES {
			return nil
		}
	}
	tok := toks[len(toks)-1]
	if tok.Id != tokens.Brace || tok.Kind != tokens.RPARENTHESES {
		return nil
	}
	braceCount := 0
	// Loops i >= 1 because expression must be has function expression at begin.
	// For this reason, ignore first token.
	for i := len(toks) - 1; i >= 1; i-- {
		tok := toks[i]
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.RPARENTHESES:
				braceCount++
			case tokens.LPARENTHESES:
				braceCount--
			}
			if braceCount == 0 {
				return toks[:i]
			}
		}
	}
	return nil
}

// RequireOperatorToProcess reports operator required for process or not.
func RequireOperatorToProcess(tok Tok, index, len int) bool {
	switch tok.Id {
	case tokens.Comma:
		return false
	case tokens.Brace:
		if tok.Kind == tokens.LPARENTHESES ||
			tok.Kind == tokens.LBRACE {
			return false
		}
	}
	return index < len-1
}

// BlockExpr returns expression tokens comes before block if exist, nil if not.
func BlockExpr(toks Toks) (expr Toks) {
	braceCount := 0
	for i, tok := range toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE:
				if braceCount > 0 {
					braceCount++
					break
				}
				return toks[:i]
			case tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
	}
	return nil
}
