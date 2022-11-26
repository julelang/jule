package ast

import "github.com/julelang/jule/lex"

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

// IsUnaryOp is returns true if operator is unary or smilar to unary,
// returns false if not.
func IsUnaryOp(kind string) bool { return existOp(kind, UNARY_OPS[:]) }

// IsStrongOp returns true operator kind is not repeatable, false if not.
func IsStrongOp(kind string) bool { return existOp(kind, STRONG_OPS[:]) }

// IsExprOp reports operator kind is allow as expression operator or not.
func IsExprOp(kind string) bool { return existOp(kind, WEAK_OPS[:]) }

func existOp(kind string, operators []string) bool {
	for _, operator := range operators {
		if kind == operator {
			return true
		}
	}
	return false
}
