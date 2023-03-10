package parser

import (
	"github.com/julelang/jule/lex"
)

// Returns function expressions without call expression
// if tokens are function call, nil if not.
func is_fn_call(tokens []lex.Token) []lex.Token {
	switch tokens[0].Id {
	case lex.ID_RANGE, lex.ID_IDENT, lex.ID_DT:
		// Ignore.
	default:
		tok := tokens[len(tokens)-1]
		if tok.Id != lex.ID_RANGE && tok.Kind != lex.KND_RPARENT {
			return nil
		}
	}
	tok := tokens[len(tokens)-1]
	if tok.Id != lex.ID_RANGE || tok.Kind != lex.KND_RPARENT {
		return nil
	}
	brace_n := 0
	// Loops i >= 1 because expression must be has function expression at begin.
	// For this reason, ignore first token.
	for i := len(tokens) - 1; i >= 1; i-- {
		tok := tokens[i]
		if tok.Id == lex.ID_RANGE {
			switch tok.Kind {
			case lex.KND_RPARENT:
				brace_n++
			case lex.KND_LPAREN:
				brace_n--
			}
			if brace_n == 0 {
				return tokens[:i]
			}
		}
	}
	return nil
}

type call_data struct {
	expr_tokens     []lex.Token
	args_tokens     []lex.Token
	generics_tokens []lex.Token
}

func get_call_data(tokens []lex.Token) *call_data {
	data := &call_data{}
	data.expr_tokens, data.args_tokens = lex.RangeLast(tokens)
	if len(data.expr_tokens) == 0 {
		return data
	}

	// Below is call expression
	token := data.expr_tokens[len(data.expr_tokens)-1]
	if token.Id == lex.ID_RANGE && token.Kind == lex.KND_RBRACKET {
		data.expr_tokens, data.generics_tokens = lex.RangeLast(data.expr_tokens)
	}
	return data
}

// Returns expression tokens comes before block if exist, nil if not.
func get_block_expr(tokens []lex.Token) []lex.Token {
	brace_n := 0
	for i, tok := range tokens {
		if tok.Id == lex.ID_RANGE {
			switch tok.Kind {
			case lex.KND_LBRACE:
				if brace_n > 0 {
					brace_n++
					break
				}
				return tokens[:i]
			case lex.KND_LBRACKET, lex.KND_LPAREN:
				brace_n++
			default:
				brace_n--
			}
		}
	}
	return nil
}

// Returns colon index and range tokens.
// Returns nil slice and -1 if not found.
// Starts search at *i.
func split_colon(tokens []lex.Token, i *int) (range_tokens []lex.Token, colon int) {
	colon = -1
	range_n := 0
	start := *i
	for ; *i < len(tokens); *i++ {
		token := tokens[*i]
		if token.Id == lex.ID_RANGE {
			switch token.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				range_n++
				continue
			default:
				range_n--
			}
		}
		if range_n == 0 {
			if start+1 > *i {
				return
			}
			range_tokens = tokens[start+1 : *i]
			break
		} else if range_n != 1 {
			continue
		}
		if colon == -1 && token.Id == lex.ID_COLON {
			colon = *i - start - 1
		}
	}
	return
}
