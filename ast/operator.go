package ast

import "github.com/the-xlang/xxc/lex/tokens"

// UnaryOperators.
var UnaryOperators = [...]string{
	0: tokens.MINUS,
	1: tokens.PLUS,
	2: tokens.TILDE,
	3: tokens.EXCLAMATION,
	4: tokens.STAR,
	5: tokens.AMPER,
}

// SolidOperators.
var SolidOperators = [...]string{
	0:  tokens.PLUS,
	1:  tokens.MINUS,
	2:  tokens.STAR,
	3:  tokens.SLASH,
	4:  tokens.PERCENT,
	5:  tokens.AMPER,
	6:  tokens.VLINE,
	7:  tokens.CARET,
	8:  tokens.LESS,
	9:  tokens.GREAT,
	10: tokens.TILDE,
	11: tokens.EXCLAMATION,
}

// ExpressionOperators.
var ExpressionOperators = [...]string{
	0: tokens.TRIPLE_DOT,
	1: tokens.COLON,
}

// IsUnaryOperator is returns true
// if operator is unary or smilar to unary,
// returns false if not.
func IsUnaryOperator(kind string) bool {
	return existOperator(kind, UnaryOperators[:])
}

// IsSolidOperator returns true operator kind is not repeatable, false if not.
func IsSolidOperator(kind string) bool {
	return existOperator(kind, SolidOperators[:])
}

// IsExpressionOperator reports operator kind
// is allow as expression operator or not.
func IsExpressionOperator(kind string) bool {
	return existOperator(kind, ExpressionOperators[:])
}

func existOperator(kind string, operators []string) bool {
	for _, operator := range operators {
		if kind == operator {
			return true
		}
	}
	return false
}
