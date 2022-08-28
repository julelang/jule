package ast

import (
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/lex/tokens"
)

// AssignInfo is the assignment information.
type AssignInfo struct {
	Left   []lex.Token
	Right  []lex.Token
	Setter lex.Token
	Ok     bool
}

// PostfixOperators.
var PostfixOperators = [...]string{
	0: tokens.DOUBLE_PLUS,
	1: tokens.DOUBLE_MINUS,
}

// AssignOperators.
var AssignOperators = [...]string{
	0:  tokens.EQUAL,
	1:  tokens.PLUS_EQUAL,
	2:  tokens.MINUS_EQUAL,
	3:  tokens.SLASH_EQUAL,
	4:  tokens.STAR_EQUAL,
	5:  tokens.PERCENT_EQUAL,
	6:  tokens.RSHIFT_EQUAL,
	7:  tokens.LSHIFT_EQUAL,
	8:  tokens.VLINE_EQUAL,
	9:  tokens.AMPER_EQUAL,
	10: tokens.CARET_EQUAL,
}

// IsAssign reports given token id is allow for
// assignment left-expression or not.
func IsAssign(id uint8) bool {
	switch id {
	case tokens.Id,
		tokens.Let,
		tokens.Dot,
		tokens.Self,
		tokens.Brace,
		tokens.Operator:
		return true
	}
	return false
}

// IsPostfixOperator reports operator kind is postfix operator or not.
func IsPostfixOperator(kind string) bool {
	for _, operator := range PostfixOperators {
		if kind == operator {
			return true
		}
	}
	return false
}

// IsAssignOperator reports operator kind is
// assignment operator or not.
func IsAssignOperator(kind string) bool {
	if IsPostfixOperator(kind) {
		return true
	}
	for _, operator := range AssignOperators {
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
		if t.Id == tokens.Brace {
			switch t.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				brace_n++
			default:
				brace_n--
			}
		}
		if brace_n < 0 {
			return false
		} else if brace_n > 0 {
			continue
		} else if t.Id == tokens.Operator && IsAssignOperator(t.Kind) {
			return true
		}
	}
	return false
}
