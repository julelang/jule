package ast

import "github.com/jule-lang/jule/lex"

// UNARY_OPS list of unary operators.
var UNARY_OPS = [...]string{
	lex.KND_MINUS,
	lex.KND_PLUS,
	lex.KND_CARET,
	lex.KND_EXCL,
	lex.KND_STAR,
	lex.KND_AMPER,
}

// STRONG_OPS list of strong operators.
// These operators are strong, can't used as part of expression.
var STRONG_OPS = [...]string{
	lex.KND_PLUS,
	lex.KND_MINUS,
	lex.KND_STAR,
	lex.KND_SOLIDUS,
	lex.KND_PERCENT,
	lex.KND_AMPER,
	lex.KND_VLINE,
	lex.KND_CARET,
	lex.KND_LT,
	lex.KND_GT,
	lex.KND_EXCL,
	lex.KND_DBL_AMPER,
	lex.KND_DBL_VLINE,
}

// WEAK_OPS list of weak operators.
// These operators are weak, can used as part of expression.
var WEAK_OPS = [...]string{
	lex.KND_TRIPLE_DOT,
	lex.KND_COLON,
}

// IsUnaryOperator is returns true
// if operator is unary or smilar to unary,
// returns false if not.
func IsUnaryOperator(kind string) bool { return existOperator(kind, UNARY_OPS[:]) }

// IsSolidOperator returns true operator kind is not repeatable, false if not.
func IsSolidOperator(kind string) bool { return existOperator(kind, STRONG_OPS[:]) }

// IsExprOperator reports operator kind
// is allow as expression operator or not.
func IsExprOperator(kind string) bool { return existOperator(kind, WEAK_OPS[:]) }

func existOperator(kind string, operators []string) bool {
	for _, operator := range operators {
		if kind == operator {
			return true
		}
	}
	return false
}
