package cxx

import (
	"runtime"
	"strconv"

	"github.com/julelang/jule/build"
	"github.com/julelang/jule/constant"
	"github.com/julelang/jule/constant/lit"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/sema"
)

// Ignore expression for std::tie function.
const CPP_IGNORE = "std::ignore"

// Represents default expression for type.
const CPP_DEFAULT_EXPR = "{}"

func get_accessor(t *sema.TypeKind) string {
	if t.Ref() != nil || t.Ptr() != nil {
		return "->"
	}
	return lex.KND_DOT
}

func get_str_model(c *constant.Const) string {
	content := c.Read_str()
	s := ""
	if lex.Is_raw_str(content) {
		s = lit.To_raw_str([]byte(content))
	} else {
		s = lit.To_str([]byte(content))
	}
	return as_jt("str") + `("` + s + `")`
}

func get_bool_model(c *constant.Const) string {
	if c.Read_bool() {
		return "true"
	}
	return "false"
}

func get_nil_model() string { return "nil" }

func get_f32_model(c *constant.Const) string {
	return strconv.FormatFloat(c.Read_f64(), 'e', -1, 32) + "f"
}

func get_f64_model(c *constant.Const) string {
	return strconv.FormatFloat(c.Read_f64(), 'e', -1, 64)
}

func get_float_model(d *sema.Data) string {
	switch {
	case d.Kind.Prim().Is_f32():
		return get_f32_model(d.Constant)

	default: // 64-bit
		return get_f64_model(d.Constant)
	}
}

func get_i64_model(c *constant.Const) string {
	fmt := strconv.FormatInt(c.Read_i64(), 10)
	if build.Is_64bit(runtime.GOARCH) {
		return fmt + "LL"
	}
	return fmt + "L"
}

func get_u64_model(c *constant.Const) string {
	fmt := strconv.FormatUint(c.Read_u64(), 10)
	if build.Is_64bit(runtime.GOARCH) {
		return fmt + "LLU"
	}
	return fmt + "LU"
}

func gen_const_expr(d *sema.Data) string {
	switch {
	case d.Constant.Is_str():
		return get_str_model(d.Constant)

	case d.Constant.Is_bool():
		return get_bool_model(d.Constant)

	case d.Constant.Is_f64():
		return get_float_model(d)

	case d.Constant.Is_i64():
		return get_i64_model(d.Constant)

	case d.Constant.Is_u64():
		return get_u64_model(d.Constant)

	case d.Constant.Is_nil():
		return get_nil_model()

	default:
		return "" // Here is should be unreachable code.
	}
}

func gen_const_expr_model(m *constant.Const) string {
	switch {
	case m.Is_str():
		return get_str_model(m)

	case m.Is_bool():
		return get_bool_model(m)

	case m.Is_f64():
		return get_f64_model(m)

	case m.Is_i64():
		return get_i64_model(m)

	case m.Is_u64():
		return get_u64_model(m)

	case m.Is_nil():
		return get_nil_model()

	default:
		return "" // Here is should be unreachable code.
	}
}

func gen_binop_expr_model(m *sema.BinopExprModel) string {
	switch m.Op {
	case lex.KND_SOLIDUS:
		obj := "__julec_div("
		obj += gen_expr_model(m.L)
		obj += ","
		obj += gen_expr_model(m.R)
		obj += ")"
		return obj

	default:
		obj := "("
		obj += gen_expr_model(m.L)
		obj += m.Op
		obj += gen_expr_model(m.R)
		obj += ")"
		return obj
	}
}

func gen_var_expr_model(m *sema.Var) string {
	return var_out_ident(m)
}

func gen_struct_expr_model(m *sema.Struct) string {
	return struct_out_ident(m)
}

func gen_fn_expr_model(m *sema.Fn) string {
	return fn_out_ident(m)
}

func gen_unary_expr_model(m *sema.UnaryExprModel) string {
	switch m.Op {
	case lex.KND_CARET:
		return "~" + gen_expr_model(m.Expr)

	default:
		return m.Op + gen_expr_model(m.Expr)
	}
}

func gen_get_ref_ptr_expr_model(m *sema.GetRefPtrExprModel) string {
	return "(" + gen_expr_model(m.Expr) + ").__alloc"
}

func gen_struct_lit_expr_model(m *sema.StructLitExprModel) string {
	obj := struct_out_ident(m.Strct.Decl)
	obj += "("
	if len(m.Args) > 0 {
		for _, f := range m.Strct.Fields {
			for _, arg := range m.Args {
				if arg.Field == f {
					obj += gen_expr_model(arg.Expr) + ","
					break;
				}
			}
		}
		obj = obj[:len(obj)-1] // Remove last comma.
	}
	obj += ")"
	return obj
}

func gen_alloc_struct_lit_expr_model(m *sema.AllocStructLitExprModel) string {
	obj := "__julec_new_structure<"
	obj += struct_out_ident(m.Lit.Strct.Decl)
	obj += ">(new( std::nothrow ) ";
	obj += gen_struct_lit_expr_model(m.Lit)
	obj += ")"
	return obj
}

func gen_casting_expr_model(m *sema.CastingExprModel) string {
	obj := ""
	switch {
	case m.ExprKind.Ptr() != nil || m.Kind.Ptr() != nil:
		obj += "(("
		obj += gen_type_kind(m.Kind)
		obj += ")("
		obj += gen_expr_model(m.Expr)
		obj += "))"
		return obj

	case m.ExprKind.Trt() != nil || (m.ExprKind.Prim() != nil && m.ExprKind.Prim().Is_any()):
		obj += gen_expr_model(m.Expr)
		obj += get_accessor(m.ExprKind)
		obj += "operator "
		obj += gen_type_kind(m.Kind)
		obj += "()"
		return obj
	}

	obj += "static_cast<"
	obj += gen_type_kind(m.Kind)
	obj += ">("
	obj += gen_expr_model(m.Expr)
	obj += ")"
	return obj
}

func gen_expr_model(m sema.ExprModel) string {
	switch m.(type) {
	case *constant.Const:
		return gen_const_expr_model(m.(*constant.Const))

	case *sema.Var:
		return gen_var_expr_model(m.(*sema.Var))

	case *sema.Struct:
		return gen_struct_expr_model(m.(*sema.Struct))

	case *sema.BinopExprModel:
		return gen_binop_expr_model(m.(*sema.BinopExprModel))

	case *sema.UnaryExprModel:
		return gen_unary_expr_model(m.(*sema.UnaryExprModel))

	case *sema.GetRefPtrExprModel:
		return gen_get_ref_ptr_expr_model(m.(*sema.GetRefPtrExprModel))

	case *sema.StructLitExprModel:
		return gen_struct_lit_expr_model(m.(*sema.StructLitExprModel))

	case *sema.AllocStructLitExprModel:
		return gen_alloc_struct_lit_expr_model(m.(*sema.AllocStructLitExprModel))

	case *sema.CastingExprModel:
		return gen_casting_expr_model(m.(*sema.CastingExprModel))

	default:
		return "<unimplemented_expression_model>"
	}
}

func gen_expr(v *sema.Value) string {
	if v.Data.Is_const() {
		return gen_const_expr(v.Data)
	}
	return gen_expr_model(v.Data.Model)
}

func get_init_expr(t *sema.TypeKind) string {
	enm := t.Enm()
	if enm == nil {
		return CPP_DEFAULT_EXPR
	}
	return "{" + gen_expr(enm.Items[0].Value) + "}"
}
