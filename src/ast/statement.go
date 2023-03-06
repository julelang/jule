package ast

import "github.com/julelang/jule/lex"

// Reports token is statement finish point or not.
func is_st(current, prev lex.Token) (ok bool, terminated bool) {
	ok = current.Id == lex.ID_SEMICOLON || prev.Row < current.Row
	terminated = current.Id == lex.ID_SEMICOLON
	return
}

// NextStPos reports position of the next statement
// if exist, len(toks) if not.
func NextStPos(toks []lex.Token, start int) (int, bool) {
	brace_n := 0
	i := start
	for ; i < len(toks); i++ {
		var ok, terminated bool
		tok := toks[i]
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				if brace_n == 0 && i > start {
					ok, terminated = is_st(tok, toks[i-1])
					if ok {
						goto ret
					}
				}
				brace_n++
				continue
			default:
				brace_n--
				if brace_n == 0 && i+1 < len(toks) {
					ok, terminated = is_st(toks[i+1], tok)
					if ok {
						i++
						goto ret
					}
				}
				continue
			}
		}
		if brace_n != 0 {
			continue
		} else if i > start {
			ok, terminated = is_st(tok, toks[i-1])
		} else {
			ok, terminated = is_st(tok, tok)
		}
		if !ok {
			continue
		}
	ret:
		if terminated {
			i++
		}
		return i, terminated
	}
	return i, false
}
