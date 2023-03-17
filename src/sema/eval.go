package sema

import (
	"strconv"
	"strings"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/constant/lit"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

// Value data.
type Data struct {
	Kind    *TypeKind
	Mutable bool
	Lvalue  bool

	// This field is reminder.
	// Will write to every constant processing points.
	// Changed after add constant evaluation support.
	// So, reminder flag for constants.
	Constant bool
}

// Reports whether Data is nil literal.
func (d *Data) Is_nil() bool { return d.Kind == nil }
// Reports whether Data is void.
func (d *Data) Is_void() bool { return d.Kind != nil && d.Kind.kind == nil }

func build_void_data() *Data {
	return &Data{
		Kind: &TypeKind{
			kind: nil,
		},
	}
}

// Value.
type Value struct {
	Expr *ast.Expr
	Data *Data
}

func kind_by_bitsize(expr any) string {
	switch expr.(type) {
	case float64:
		x := expr.(float64)
		return types.Float_from_bits(types.Bitsize_of_float(x))

	case int64:
		x := expr.(int64)
		return types.Int_from_bits(types.Bitsize_of_int(x))

	case uint64:
		x := expr.(uint64)
		return types.Uint_from_bits(types.Bitsize_of_uint(x))

	default:
		return ""
	}
}

// Evaluator.
type _Eval struct {
	s      *_Sema  // Used for error logging.
	lookup _Lookup
}

// TODO: Implement here.
// Reports whether evaluation in unsafe scope.
func (e *_Eval) is_unsafe() bool { return false }

func (e *_Eval) lit_nil() *Data {
	// Return new Data with nil kind.
	// Nil kind represents "nil" literal.
	return &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: true,
		Kind:     nil,
	}
}

func (e *_Eval) lit_str(lit *ast.LitExpr) *Data {
	return &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: true,
		Kind:     &TypeKind{
			kind: build_prim_type(types.TypeKind_STR),
		},
	}
}

func (e *_Eval) lit_bool(lit *ast.LitExpr) *Data {
	return &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: true,
		Kind:     &TypeKind{
			kind: build_prim_type(types.TypeKind_BOOL),
		},
	}
}

func (e *_Eval) lit_rune(l *ast.LitExpr) *Data {
	const BYTE_KIND = types.TypeKind_U8
	const RUNE_KIND = types.TypeKind_I32
	
	data := &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: true,
	}

	_, is_byte := lit.Is_byte_lit(l.Value)
	if is_byte {
		data.Kind = &TypeKind{
			kind: build_prim_type(BYTE_KIND),
		}
	} else {
		data.Kind = &TypeKind{
			kind: build_prim_type(RUNE_KIND),
		}
	}

	return data
}

func (e *_Eval) lit_float(l *ast.LitExpr) *Data {
	const FLOAT_KIND = types.TypeKind_F64

	return &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: true,
		Kind:     &TypeKind{
			kind: build_prim_type(FLOAT_KIND),
		},
	}
}

func (e *_Eval) lit_int(l *ast.LitExpr) *Data {
	data := l.Value
	base := 0

	switch {
	case strings.HasPrefix(data, "0x"): // Hexadecimal
		data = data[2:]
		base = 0b00010000

		case strings.HasPrefix(data, "0b"): // Binary
		data = data[2:]
		base = 0b10

	case data[0] == '0' && len(data) > 1: // Octal
		data = data[1:]
		base = 0b1000

	default: // Decimal
		base = 0b1010
	}

	is_neg := data[0] == '-'
	var value any = nil
	const BIT_SIZE = 0b01000000
	if is_neg {
		value, _ = strconv.ParseInt(data, base, BIT_SIZE)
	} else {
		temp_value, err := strconv.ParseInt(data, base, BIT_SIZE)
		if err == nil {
			value = temp_value
		} else {
			value, _ = strconv.ParseUint(data, base, BIT_SIZE)
		}
	}

	return &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: true,
		Kind:     &TypeKind{
			kind: build_prim_type(kind_by_bitsize(value)),
		},
	}
}

func (e *_Eval) lit_num(l *ast.LitExpr) *Data {
	switch {
	case lex.Is_float(l.Value):
		return e.lit_float(l)

	default:
		return e.lit_int(l)
	}
}

func (e *_Eval) eval_lit(lit *ast.LitExpr) *Data {
	switch {
	case lit.Is_nil():
		return e.lit_nil()

	case lex.Is_str(lit.Value):
		return e.lit_str(lit)

	case lex.Is_bool(lit.Value):
		return e.lit_bool(lit)

	case lex.Is_rune(lit.Value):
		return e.lit_rune(lit)

	case lex.Is_num(lit.Value):
		return e.lit_num(lit)

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
