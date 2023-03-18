package sema

import (
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

// This file reserved for type compatibility checking.

type _TypeCompatibilityChecker struct {
	s           *_Sema    // Used for error logging.
	dest        *TypeKind
	src         *TypeKind
	error_token lex.Token
}

func (tcc *_TypeCompatibilityChecker) push_err(key string, args ...any) {
	tcc.s.push_err(tcc.error_token, key, args...)
}

func (tcc *_TypeCompatibilityChecker) check() (ok bool) {
	// TODO: Check other cases.
	return types.Types_are_compatible(tcc.dest.To_str(), tcc.src.To_str())
}
