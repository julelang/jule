package ast

import "github.com/the-xlang/xxc/lex/tokens"

// MapDataTypeInfo returns data type tokens of map type.
// Starts at i.
func MapDataTypeInfo(toks Toks, i *int) (typeToks Toks, colon int) {
	typeToks = nil
	colon = -1
	braceCount := 0
	start := *i
	for ; *i < len(toks); *i++ {
		tok := toks[*i]
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount == 0 {
			if start+1 > *i {
				return
			}
			typeToks = toks[start+1 : *i]
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
