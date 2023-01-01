package ast

import "github.com/julelang/jule/lex"

// IsSt reports token is statement finish point or not.
func IsSt(current, prev lex.Token) (ok bool, terminated bool) {
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
		var is_st, terminated bool
		tok := toks[i]
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				if brace_n == 0 && i > start {
					is_st, terminated = IsSt(tok, toks[i-1])
					if is_st {
						goto ret
					}
				}
				brace_n++
				continue
			default:
				brace_n--
				if brace_n == 0 && i+1 < len(toks) {
					is_st, terminated = IsSt(toks[i+1], tok)
					if is_st {
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
			is_st, terminated = IsSt(tok, toks[i-1])
		} else {
			is_st, terminated = IsSt(tok, tok)
		}
		if !is_st {
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
