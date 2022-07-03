package ast

import "github.com/the-xlang/xxc/lex/tokens"

// AssignInfo is the assignment information.
type AssignInfo struct {
	Left   Toks
	Right  Toks
	Setter Tok
	Ok     bool
	IsExpr bool
}

// IsAssign reports given token id is allow for
// assignment left-expression or not.
func IsAssign(id uint8) bool {
	return id == tokens.Id ||
		id == tokens.Brace ||
		id == tokens.Operator
}

// IsPostfixOperator reports operator kind is postfix operator or not.
func IsPostfixOperator(kind string) bool {
	return kind == tokens.DOUBLE_PLUS ||
		kind == tokens.DOUBLE_MINUS
}

// IsAssignOperator reports operator kind is
// assignment operator or not.
func IsAssignOperator(kind string) bool {
	return IsPostfixOperator(kind) ||
		kind == tokens.EQUAL ||
		kind == tokens.PLUS_EQUAL ||
		kind == tokens.MINUS_EQUAL ||
		kind == tokens.SLASH_EQUAL ||
		kind == tokens.STAR_EQUAL ||
		kind == tokens.PERCENT_EQUAL ||
		kind == tokens.RSHIFT_EQUAL ||
		kind == tokens.LSHIFT_EQUAL ||
		kind == tokens.VLINE_EQUAL ||
		kind == tokens.AMPER_EQUAL ||
		kind == tokens.CARET_EQUAL
}

// CheckAssignToks checks assignment tokens and reports is ok or not.
func CheckAssignToks(toks Toks) bool {
	if len(toks) == 0 || !IsAssign(toks[0].Id) {
		return false
	}
	braceCount := 0
	for _, tok := range toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount < 0 {
			return false
		} else if braceCount > 0 {
			continue
		}
		if tok.Id == tokens.Operator &&
			IsAssignOperator(tok.Kind) {
			return true
		}
	}
	return false
}
