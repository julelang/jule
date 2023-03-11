package parser

import "github.com/julelang/jule/lex"

// Assignment information.
type _AssignInfo struct {
	l      []lex.Token
	r      []lex.Token
	setter lex.Token
	ok     bool
}

// List of postfix operators.
var _POSTFIX_OPS = [...]string{
	lex.KND_DBL_PLUS,
	lex.KND_DBL_MINUS,
}

// List of assign operators.
var _ASSING_OPS = [...]string{
	lex.KND_EQ,
	lex.KND_PLUS_EQ,
	lex.KND_MINUS_EQ,
	lex.KND_SOLIDUS_EQ,
	lex.KND_STAR_EQ,
	lex.KND_PERCENT_EQ,
	lex.KND_RSHIFT_EQ,
	lex.KND_LSHIFT_EQ,
	lex.KND_VLINE_EQ,
	lex.KND_AMPER_EQ,
	lex.KND_CARET_EQ,
}

// Reports given token id is allow for
// assignment left-expression or not.
func is_assign(id uint8) bool {
	switch id {
	case lex.ID_IDENT,
		lex.ID_CPP,
		lex.ID_LET,
		lex.ID_DOT,
		lex.ID_SELF,
		lex.ID_RANGE,
		lex.ID_OP:
		return true
	}
	return false
}

// Rreports whether operator kind is postfix operator.
func is_postfix_op(kind string) bool {
	for _, operator := range _POSTFIX_OPS {
		if kind == operator {
			return true
		}
	}
	return false
}

// Reports whether operator kind is assignment operator.
func is_assign_op(kind string) bool {
	if is_postfix_op(kind) {
		return true
	}
	for _, operator := range _ASSING_OPS {
		if kind == operator {
			return true
		}
	}
	return false
}

// Checks assignment tokens and whether reports is ok or not.
func check_assign_tokens(tokens []lex.Token) bool {
	if len(tokens) == 0 || !is_assign(tokens[0].Id) {
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
		} else if t.Id == lex.ID_OP && is_assign_op(t.Kind) {
			return true
		}
	}
	return false
}
