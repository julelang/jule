package ast

import "github.com/the-xlang/xxc/lex/tokens"

// IsSigleOperator is returns true
// if operator is unary or smilar to unary,
// returns false if not.
func IsSingleOperator(kind string) bool {
	return kind == tokens.MINUS ||
		kind == tokens.PLUS ||
		kind == tokens.TILDE ||
		kind == tokens.EXCLAMATION ||
		kind == tokens.STAR ||
		kind == tokens.AMPER
}

// IsSolidOperator returns true operator kind is not repeatable, false if not.
func IsSolidOperator(kind string) bool {
	return kind == tokens.PLUS ||
		kind == tokens.MINUS ||
		kind == tokens.STAR ||
		kind == tokens.SLASH ||
		kind == tokens.PERCENT ||
		kind == tokens.AMPER ||
		kind == tokens.VLINE ||
		kind == tokens.CARET ||
		kind == tokens.LESS ||
		kind == tokens.GREAT ||
		kind == tokens.TILDE ||
		kind == tokens.EXCLAMATION
}

// IsExprOperator reports operator kind is allow as expression operator or not.
func IsExprOperator(kind string) bool { return kind == tokens.TRIPLE_DOT }
