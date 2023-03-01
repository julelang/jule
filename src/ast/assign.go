package ast

import "github.com/julelang/jule/lex"

// AssignInfo is the assignment information.
type AssignInfo struct {
	Left   []lex.Token
	Right  []lex.Token
	Setter lex.Token
	Ok     bool
}

// POSTFIX_OPS list of postfix operators.
var POSTFIX_OPS = [...]string{
	lex.KND_DBL_PLUS,
	lex.KND_DBL_MINUS,
}

// ASSING_OPS list of assign operators.
var ASSING_OPS = [...]string{
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

// IsAssign reports given token id is allow for
// assignment left-expression or not.
func IsAssign(id uint8) bool {
	switch id {
	case lex.ID_IDENT,
		lex.ID_CPP,
		lex.ID_LET,
		lex.ID_DOT,
		lex.ID_SELF,
		lex.ID_BRACE,
		lex.ID_OP:
		return true
	}
	return false
}

// IsPostfixOp reports operator kind is postfix operator or not.
func IsPostfixOp(kind string) bool {
	for _, operator := range POSTFIX_OPS {
		if kind == operator {
			return true
		}
	}
	return false
}

// IsAssignOp reports operator kind is
// assignment operator or not.
func IsAssignOp(kind string) bool {
	if IsPostfixOp(kind) {
		return true
	}
	for _, operator := range ASSING_OPS {
		if kind == operator {
			return true
		}
	}
	return false
}

// CheckAssignTokens checks assignment tokens and reports is ok or not.
func CheckAssignTokens(toks []lex.Token) bool {
	if len(toks) == 0 || !IsAssign(toks[0].Id) {
		return false
	}
	brace_n := 0
	for _, t := range toks {
		if t.Id == lex.ID_BRACE {
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
		} else if t.Id == lex.ID_OP && IsAssignOp(t.Kind) {
			return true
		}
	}
	return false
}
