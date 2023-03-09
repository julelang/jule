package parser

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Reports whether token is statement finish point.
func is_st(current lex.Token, prev lex.Token) (ok bool, terminated bool) {
	ok = current.Id == lex.ID_SEMICOLON || prev.Row < current.Row
	terminated = current.Id == lex.ID_SEMICOLON
	return
}

// Reports position of the next statement if exist, len(toks) if not.
func next_st_pos(tokens []lex.Token, start int) (int, bool) {
	brace_n := 0
	i := start
	for ; i < len(tokens); i++ {
		var ok, terminated bool
		tok := tokens[i]
		if tok.Id == lex.ID_RANGE {
			switch tok.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				if brace_n == 0 && i > start {
					ok, terminated = is_st(tok, tokens[i-1])
					if ok {
						goto ret
					}
				}
				brace_n++
				continue
			default:
				brace_n--
				if brace_n == 0 && i+1 < len(tokens) {
					ok, terminated = is_st(tokens[i+1], tok)
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
			ok, terminated = is_st(tok, tokens[i-1])
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

// Returns current statement tokens.
// Starts selection at *i.
func skip_st(i *int, tokens []lex.Token) []lex.Token {
	start := *i
	*i, _ = next_st_pos(tokens, start)
	st_tokens := tokens[start:*i]
	if st_tokens[len(st_tokens)-1].Id == lex.ID_SEMICOLON {
		if len(st_tokens) == 1 {
			return skip_st(i, tokens)
		}
		// -1 for eliminate statement terminator.
		st_tokens = st_tokens[:len(st_tokens)-1]
	}
	return st_tokens
}

// Splits all statements.
func split_stms(tokens []lex.Token) [][]lex.Token {
	var stms [][]lex.Token = nil
	pos := 0
	for pos < len(tokens) {
		stms = append(stms, skip_st(&pos, tokens))
	}
	return stms
}

type scope_parser struct {
	p    *parser
	s    *ast.Scope
}

func (sp *scope_parser) push_err(token lex.Token, key string) {
	sp.p.push_err(token, key)
}

func (sp *scope_parser) build_var_st(tokens []lex.Token) ast.NodeData {
	v := sp.p.build_var(tokens)
	v.Scope = sp.s
	return v
}

func (sp *scope_parser) build_st(tokens []lex.Token) ast.NodeData {
	token := tokens[0]
	switch token.Id {
	case lex.ID_CONST, lex.ID_LET, lex.ID_MUT:
		return sp.build_var_st(tokens)
	}
	sp.push_err(token, "invalid_syntax")
	return nil
}

func (sp *scope_parser) build(tokens []lex.Token) *ast.Scope {
	stms := split_stms(tokens)
	sp.s = &ast.Scope{}
	for _, st := range stms {
		data := sp.build_st(st)
		if data != nil {
			sp.s.Tree = append(sp.s.Tree, data)
		}
	}
	return sp.s
}
