package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Value data.
type Data struct {
	Kind *TypeKind
}

// Reports whether Data is nil literal.
func (d *Data) Is_nil() bool { return d.Kind == nil }

// Value.
type Value struct {
	Expr *ast.Expr
	Data *Data
}

// Evaluator.
type _Eval struct {
	s      *_Sema  // Used for error logging.
	lookup _Lookup
}

// TODO: Implement here.
// Reports whether evaluation in unsafe scope.
func (e *_Eval) is_unsafe() bool { return false }

func (e *_Eval) lit_str(lit *ast.LitExpr) *Data {
	s := lit.Value[1:len(lit.Value)-1] // Remove string quotes.

	return &Data{
		Kind: &TypeKind{
			kind: build_prim_type(s),
		},
	}
}

func (e *_Eval) lit_nil() *Data {
	// Return new Data with nil kind.
	// Nil kind represents "nil" literal.
	return &Data{
		Kind: nil,
	}
}

func (e *_Eval) eval_lit(lit *ast.LitExpr) *Data {
	switch {
	case lit.Is_nil():
		return e.lit_nil()

	case lex.Is_str(lit.Value):
		return e.lit_str(lit)

	default:
		return nil
	}
}

func (e *_Eval) eval_expr_kind(kind ast.ExprData) *Data {
	// TODO: Implement other types.
	switch kind.(type) {
	case *ast.LitExpr:
		return e.eval_lit(kind.(*ast.LitExpr))

	default:
		return nil
	}
}

// Returns value data of evaluated expression.
// Returns nil if error occurs.
func (e *_Eval) eval(expr *ast.Expr) *Data {
	return e.eval_expr_kind(expr.Kind)
}
