// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package parser

import "github.com/julelang/jule/lex"

// Assignment information.
type _AssignInfo struct {
	l      []lex.Token
	r      []lex.Token
	setter lex.Token
	ok     bool
}

// Checks assignment tokens and whether reports is ok or not.
func check_assign_tokens(tokens []lex.Token) bool {
	if len(tokens) == 0 || !lex.Is_assign(tokens[0].Id) {
		return false
	}

	brace_n := 0
	for _, t := range tokens {
		if t.Id == lex.ID_RANGE {
			switch t.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				brace_n++

			default:
				brace_n--
			}
		}

		if brace_n < 0 {
			return false
		} else if brace_n > 0 {
			continue
		} else if t.Id == lex.ID_OP && lex.Is_assign_op(t.Kind) {
			return true
		}
	}

	return false
}
