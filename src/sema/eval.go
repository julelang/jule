package sema

import "github.com/julelang/jule/ast"

// Value.
type Value struct {
	Expr *ast.Expr
}

// Evaluator.
type _Eval struct {
	// Used for error logging.
	s *_Sema
}

// TODO: Implement here.
// Reports whether evaluation in unsafe scope.
func (e *_Eval) is_unsafe() bool { return false }

func (e *_Eval) eval(v *Value) (ok bool) {
	return false
}
