package ast

import (
	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/lex/tokens"
)

type blockStatement struct {
	pos            int
	block          *models.Block
	srcToks        *[]lex.Token
	toks           []lex.Token
	nextToks       []lex.Token
	withTerminator bool
}

// IsStatement reports token is
// statement finish point or not.
func IsStatement(current, prev lex.Token) (ok bool, withTerminator bool) {
	ok = current.Id == tokens.SemiColon || prev.Row < current.Row
	withTerminator = current.Id == tokens.SemiColon
	return
}

// NextStatementPos reports position of the next statement
// if exist, len(toks) if not.
func NextStatementPos(toks []lex.Token, start int) (int, bool) {
	brace_n := 0
	i := start
	for ; i < len(toks); i++ {
		var isStatement, withTerminator bool
		tok := toks[i]
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				if brace_n == 0 && i > start {
					isStatement, withTerminator = IsStatement(tok, toks[i-1])
					if isStatement {
						goto ret
					}
				}
				brace_n++
				continue
			default:
				brace_n--
				if brace_n == 0 && i+1 < len(toks) {
					isStatement, withTerminator = IsStatement(toks[i+1], tok)
					if isStatement {
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
			isStatement, withTerminator = IsStatement(tok, toks[i-1])
		} else {
			isStatement, withTerminator = IsStatement(tok, tok)
		}
		if !isStatement {
			continue
		}
	ret:
		if withTerminator {
			i++
		}
		return i, withTerminator
	}
	return i, false
}
