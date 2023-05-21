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
	obj := struct_ins_out_ident(m.Strct)
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

func gen_arg_expr_models(models []sema.ExprModel) string {
	if len(models) == 0 {
		return ""
	}

	obj := ""
	for _, m := range models {
		obj += gen_expr_model(m) + ","
	}
	obj = obj[:len(obj)-1] // Remove last comma.
	return obj
}

func gen_fn_call_expr_model(m *sema.FnCallExprModel) string {
	obj := gen_expr_model(m.Expr)
	obj += "("
	obj += gen_arg_expr_models(m.Args)
	obj += ")"

	if m.IsCo {
		obj = "__JULEC_CO(" + obj + ")"
	}

	return obj
}

func gen_slice_expr_model(m *sema.SliceExprModel) string {
	obj := as_slice_kind(m.Elem_kind)
	obj += "({"
	obj += gen_arg_expr_models(m.Elems)
	obj += "})"
	return obj
}

func gen_indexing_expr_model(m *sema.IndexigExprModel) string {
	obj := gen_expr_model(m.Expr)
	obj += "["
	obj += gen_expr_model(m.Index)
	obj += "]"
	return obj
}

func gen_anon_fn_expr_model(m *sema.AnonFnExprModel) string {
	obj := gen_fn_kind(m.Func)
	if m.Global {
		obj += "([]"
	} else {
		obj += "([&]"
	}
	obj += gen_params_ins(m.Func.Params)
	obj += " mutable -> "
	obj += gen_fn_ins_result(m.Func)
	obj += " "
	obj += gen_fn_scope(m.Func)
	obj += ")"
	return obj
}

func gen_map_expr_model(m *sema.MapExprModel) string {
	obj := as_jt("map")
	obj += "<"
	obj += gen_type_kind(m.Key_kind)
	obj += ","
	obj += gen_type_kind(m.Val_kind)
	obj += ">({"
	if len(m.Entries) > 0 {
		for _, pair := range m.Entries {
			pair_obj := "{"
			pair_obj += gen_expr_model(pair.Key)
			pair_obj += ","
			pair_obj += gen_expr_model(pair.Val)
			pair_obj += "}"
			obj += pair_obj
			obj += ","
		}
		obj = obj[:len(obj)-1] // Remove last comma.
	}
	obj += "})"
	return obj
}

func gen_slicing_expr_model(m *sema.SlicingExprModel) string {
	obj := gen_expr_model(m.Expr)
	obj += ".___slice("
	obj += gen_expr_model(m.L)
	if m.R != nil {
		obj += ","
		obj += gen_expr_model(m.R)
	}
	obj += ")"
	return obj
}

func gen_trait_sub_ident_expr_model(m *sema.TraitSubIdentExprModel) string {
	obj := gen_expr_model(m.Expr)
	obj += "._get()."
	obj += m.Ident
	return obj
}

func gen_struct_sub_ident_expr_model(m *sema.StrctSubIdentExprModel) string {
	obj := gen_expr_model(m.Expr)
	obj += get_accessor(m.ExprKind)
	if m.Field != nil {
		obj += field_out_ident(m.Field.Decl)
	} else {
		obj += fn_ins_out_ident(m.Method)
	}
	return obj
}

func gen_common_sub_ident_expr_model(m *sema.CommonSubIdentExprModel) string {
	obj := gen_expr_model(m.Expr)
	obj += "."
	obj += m.Ident
	return obj
}

func gen_array_expr_model(m *sema.ArrayExprModel) string {
	obj := gen_array_kind(m.Kind)
	obj += "({"
	obj += gen_arg_expr_models(m.Elems)
	obj += "})"
	return obj
}

func gen_fn_ins_expr_model(m *sema.FnIns) string {
	return fn_ins_out_ident(m)
}

func gen_tuple_expr_model(m *sema.TupleExprModel) string {
	obj := "std::make_tuple("
	for _, d := range m.Datas {
		obj += gen_expr_model(d.Model) + ","
	}
	obj = obj[:len(obj)-1] // Remove last comma.
	obj += ")"
	return obj
}

func gen_builtin_new_call_expr_model(m *sema.BuiltinNewCallExprModel) string {
	obj := "_new<"
	obj += gen_type_kind(m.Kind)
	obj += ">("
	if m.Init != nil {
		obj += gen_expr_model(m.Init)
	}
	obj += ")"
	return obj
}

func gen_builtin_out_call_expr_model(m *sema.BuiltinOutCallExprModel) string {
	obj := "_out("
	obj += gen_expr_model(m.Expr)
	obj += ")"
	return obj
}

func gen_builtin_real_call_expr_model(m *sema.BuiltinRealCallExprModel) string {
	obj := "_real("
	obj += gen_expr_model(m.Expr)
	obj += ")"
	return obj
}

func gen_builtin_drop_call_expr_model(m *sema.BuiltinDropCallExprModel) string {
	obj := "_drop("
	obj += gen_expr_model(m.Expr)
	obj += ")"
	return obj
}

func gen_expr_model(m sema.ExprModel) string {
	switch m.(type) {
	case *sema.TypeKind:
		return gen_type_kind(m.(*sema.TypeKind))

	case *constant.Const:
		return gen_const_expr_model(m.(*constant.Const))

	case *sema.Var:
		return gen_var_expr_model(m.(*sema.Var))

	case *sema.Struct:
		return gen_struct_expr_model(m.(*sema.Struct))

	case *sema.FnIns:
		return gen_fn_ins_expr_model(m.(*sema.FnIns))

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

	case *sema.FnCallExprModel:
		return gen_fn_call_expr_model(m.(*sema.FnCallExprModel))

	case *sema.SliceExprModel:
		return gen_slice_expr_model(m.(*sema.SliceExprModel))

	case *sema.ArrayExprModel:
		return gen_array_expr_model(m.(*sema.ArrayExprModel))

	case *sema.IndexigExprModel:
		return gen_indexing_expr_model(m.(*sema.IndexigExprModel))

	case *sema.AnonFnExprModel:
		return gen_anon_fn_expr_model(m.(*sema.AnonFnExprModel))

	case *sema.MapExprModel:
		return gen_map_expr_model(m.(*sema.MapExprModel))

	case *sema.SlicingExprModel:
		return gen_slicing_expr_model(m.(*sema.SlicingExprModel))

	case *sema.TraitSubIdentExprModel:
		return gen_trait_sub_ident_expr_model(m.(*sema.TraitSubIdentExprModel))

	case *sema.StrctSubIdentExprModel:
		return gen_struct_sub_ident_expr_model(m.(*sema.StrctSubIdentExprModel))

	case *sema.CommonSubIdentExprModel:
		return gen_common_sub_ident_expr_model(m.(*sema.CommonSubIdentExprModel))

	case *sema.TupleExprModel:
		return gen_tuple_expr_model(m.(*sema.TupleExprModel))

	case *sema.BuiltinNewCallExprModel:
		return gen_builtin_new_call_expr_model(m.(*sema.BuiltinNewCallExprModel))

	case *sema.BuiltinOutCallExprModel:
		return gen_builtin_out_call_expr_model(m.(*sema.BuiltinOutCallExprModel))

	case *sema.BuiltinRealCallExprModel:
		return gen_builtin_real_call_expr_model(m.(*sema.BuiltinRealCallExprModel))

	case *sema.BuiltinDropCallExprModel:
		return gen_builtin_drop_call_expr_model(m.(*sema.BuiltinDropCallExprModel))

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
