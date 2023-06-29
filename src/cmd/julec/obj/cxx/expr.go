package cxx

import (
	"math"
	"runtime"
	"strconv"

	"github.com/julelang/jule/build"
	"github.com/julelang/jule/constant"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/sema"
	"github.com/julelang/jule/types"
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

func decompose_common_esq(b byte) string {
	switch b {
	case '\\':
		return "\\\\"

	case '\'':
		return "'"

	case '"':
		return `\"`

	case '\a':
		return `\a`

	case '\b':
		return `\b`

	case '\f':
		return `\f`

	case '\n':
		return `\n`

	case '\r':
		return `\r`

	case '\t':
		return `\t`

	case '\v':
		return `\v`

	default:
		return ""
	}
}

func sbtoa(b byte) string {
	if b == 0 {
		return "\\x00"
	}

	if b <= 127 { // ASCII
		seq := decompose_common_esq(b)
		if seq != "" {
			return seq
		}

		if 32 <= b && b <= 126 {
			return string(b)
		}
	}

	seq := strconv.FormatUint(uint64(b), 8)
	return "\\" + seq
}

func get_str_model(c *constant.Const) string {
	content := c.Read_str()
	bytes := []byte(content)
	len := strconv.FormatInt(int64(len(bytes)), 10)

	lit := ""
	for _, b := range bytes {
		lit += sbtoa(b)
	}

	return as_jt("str") + `("` + lit + `", ` + len + ")"
}

func get_bool_model(c *constant.Const) string {
	if c.Read_bool() {
		return "true"
	}
	return "false"
}

func get_nil_model() string { return "nullptr" }

func gen_float_special_cases(x float64) string {
	switch {
	case math.IsNaN(x):
		return "NAN"

	case math.IsInf(x, 1):
		return "INFINITY"

	case math.IsInf(x, -1):
		return "-INFINITY"

	default:
		return ""
	}
}

func get_f32_model(c *constant.Const) string {
	x := c.As_f64()

	// Special cases.
	f := gen_float_special_cases(x)
	if f != "" {
		return f
	}

	switch {
	case x == types.MAX_F32:
		return "jule::MAX_F32"

	case x == types.MIN_F32:
		return "jule::MIN_F32"
	}

	return strconv.FormatFloat(x, 'e', -1, 32) + "f"
}

func get_f64_model(c *constant.Const) string {
	x := c.As_f64()

	// Special cases.
	f := gen_float_special_cases(x)
	if f != "" {
		return f
	}

	switch {
	case x == types.MAX_F64:
		return "jule::MAX_F64"

	case x == types.MIN_F64:
		return "jule::MIN_F64"
	}

	return strconv.FormatFloat(x, 'e', -1, 64)
}

func get_float_model(d *sema.Data) string {
	switch {
	case d.Kind.Prim().Is_f32():
		return get_f32_model(d.Constant)

	default: // 64-bit
		return get_f64_model(d.Constant)
	}
}

func i64toa(x int64) string {
	switch {
	case x == types.MAX_I64:
		return "jule::MAX_I64"

	case x == types.MIN_I64:
		return "jule::MIN_I64"
	}

	fmt := strconv.FormatInt(x, 10)
	if build.Is_64bit(runtime.GOARCH) {
		return fmt + "LL"
	}
	return fmt + "L"
}

func get_i64_model(c *constant.Const) string {
	return i64toa(c.Read_i64())
}

func get_u64_model(c *constant.Const) string {
	x := c.Read_u64()

	switch {
	case x == types.MAX_U64:
		return "jule::MAX_U64"
	}

	fmt := strconv.FormatUint(x, 10)
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
		obj := "jule::div("
		obj += gen_expr(m.Left)
		obj += ","
		obj += gen_expr(m.Right)
		obj += ")"
		return obj

	default:
		obj := "("
		obj += gen_expr_model(m.Left)
		obj += " "
		obj += m.Op
		obj += " "
		obj += gen_expr_model(m.Right)
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
		return "(~" + gen_expr(m.Expr) + ")"

	default:
		return "(" + m.Op + gen_expr(m.Expr) + ")"
	}
}

func gen_get_ref_ptr_expr_model(m *sema.GetRefPtrExprModel) string {
	return "(" + gen_expr(m.Expr) + ").alloc"
}

func gen_cpp_struct_lit_expr_model(m *sema.StructLitExprModel) string {
	obj := "(" + struct_ins_out_ident(m.Strct)
	obj += "){"
	if len(m.Args) > 0 {
	iter:
		for _, f := range m.Strct.Fields {
			obj += field_out_ident(f.Decl) + ": "
			for _, arg := range m.Args {
				if arg.Field == f {
					obj += gen_expr(arg.Expr) + ","
					continue iter
				}
			}
			obj += get_init_expr(f.Kind) + ","
		}
		obj = obj[:len(obj)-1] // Remove last comma.
	}
	obj += "}"
	return obj
}

func gen_struct_lit_expr_model(m *sema.StructLitExprModel) string {
	if m.Strct.Decl.Cpp_linked {
		return gen_cpp_struct_lit_expr_model(m)
	}

	obj := struct_ins_out_ident(m.Strct)
	obj += "("
	if len(m.Args) > 0 {
	iter:
		for _, f := range m.Strct.Fields {
			for _, arg := range m.Args {
				if arg.Field == f {
					obj += gen_expr(arg.Expr) + ","
					continue iter
				}
			}
			obj += get_init_expr(f.Kind) + ","
		}
		obj = obj[:len(obj)-1] // Remove last comma.
	}
	obj += ")"
	return obj
}

func gen_alloc_struct_lit_expr_model(m *sema.AllocStructLitExprModel) string {
	obj := "jule::new_struct<"
	obj += struct_out_ident(m.Lit.Strct.Decl)
	obj += ">(new( std::nothrow ) "
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
		obj += gen_expr(m.Expr)
		obj += "))"

	case m.ExprKind.Trt() != nil || (m.ExprKind.Prim() != nil && m.ExprKind.Prim().Is_any()):
		obj += gen_expr_model(m.Expr)
		obj += get_accessor(m.ExprKind)
		obj += "operator "
		obj += gen_type_kind(m.Kind)
		obj += "()"

	default:
		obj += "static_cast<"
		obj += gen_type_kind(m.Kind)
		obj += ">("
		obj += gen_expr(m.Expr)
		obj += ")"
	}
	return obj
}

func gen_arg_expr_models(models []sema.ExprModel) string {
	if len(models) == 0 {
		return ""
	}

	obj := ""
	for _, m := range models {
		obj += gen_expr(m) + ","
	}
	obj = obj[:len(obj)-1] // Remove last comma.
	return obj
}

func gen_fn_call_expr_model(m *sema.FnCallExprModel) string {
	obj := gen_expr_model(m.Expr)
	if !m.Func.Is_builtin() && m.Func.Decl.Cpp_linked && len(m.Func.Generics) > 0 {
		if !has_directive(m.Func.Decl.Directives, build.DIRECTIVE_CDEF) {
			obj += "<"
			for _, g := range m.Func.Generics {
				obj += gen_type_kind(g) + ","
			}
			obj = obj[:len(obj)-1] // Remove last comma.
			obj += ">"
		}
	}
	obj += "("
	obj += gen_arg_expr_models(m.Args)
	obj += ")"

	if m.IsCo {
		obj = "__JULE_CO(" + obj + ")"
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
	obj += gen_expr(m.Index)
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
			pair_obj += gen_expr(pair.Key)
			pair_obj += ","
			pair_obj += gen_expr(pair.Val)
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
	obj += ".slice("
	obj += gen_expr(m.Left)
	if m.Right != nil {
		obj += ","
		obj += gen_expr(m.Right)
	}
	obj += ")"
	return obj
}

func gen_trait_sub_ident_expr_model(m *sema.TraitSubIdentExprModel) string {
	obj := gen_expr_model(m.Expr)
	obj += ".get()._method_"
	obj += m.Ident
	return obj
}

func gen_struct_sub_ident_expr_model(m *sema.StructSubIdentExprModel) string {
	obj := gen_expr_model(m.Expr)
	obj += get_accessor(m.ExprKind)
	if m.Field != nil {
		obj += field_out_ident(m.Field.Decl)
	} else {
		obj += fn_ins_out_ident(m.Method)
	}
	return obj
}

func gen_common_ident_expr_model(m *sema.CommonIdentExprModel) string {
	return m.Ident
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
		obj += gen_expr(d.Model) + ","
	}
	obj = obj[:len(obj)-1] // Remove last comma.
	obj += ")"
	return obj
}

func gen_builtin_new_call_expr_model(m *sema.BuiltinNewCallExprModel) string {
	obj := "jule::new_ref<"
	obj += gen_type_kind(m.Kind)
	obj += ">("
	if m.Init != nil {
		obj += gen_expr(m.Init)
	}
	obj += ")"
	return obj
}

func gen_builtin_out_call_expr_model(m *sema.BuiltinOutCallExprModel) string {
	obj := "jule::out("
	obj += gen_expr(m.Expr)
	obj += ")"
	return obj
}

func gen_builtin_outln_call_expr_model(m *sema.BuiltinOutlnCallExprModel) string {
	obj := "jule::outln("
	obj += gen_expr(m.Expr)
	obj += ")"
	return obj
}

func gen_builtin_real_call_expr_model(m *sema.BuiltinRealCallExprModel) string {
	obj := "jule::real("
	obj += gen_expr(m.Expr)
	obj += ")"
	return obj
}

func gen_builtin_drop_call_expr_model(m *sema.BuiltinDropCallExprModel) string {
	obj := "jule::drop("
	obj += gen_expr(m.Expr)
	obj += ")"
	return obj
}

func gen_builtin_panic_call_expr_model(m *sema.BuiltinPanicCallExprModel) string {
	obj := "jule::panic("
	obj += gen_expr(m.Expr)
	obj += ")"
	return obj
}

func gen_builtin_make_call_expr_model(m *sema.BuiltinMakeCallExprModel) string {
	obj := gen_type_kind(m.Kind)
	obj += "::alloc("
	if m.Size != nil {
		obj += gen_expr(m.Size)
	} else {
		obj += "0"
	}
	obj += ")"
	return obj
}

func gen_builtin_clone_call_expr_model(m *sema.BuiltinCloneCallExprModel) string {
	obj := "jule::clone("
	obj += gen_expr_model(m.Expr)
	obj += ")"
	return obj
}

func gen_sizeof_expr_model(m *sema.SizeofExprModel) string {
	obj := "sizeof("
	obj += gen_expr(m.Expr)
	obj += ")"
	return obj
}

func gen_alignof_expr_model(m *sema.AlignofExprModel) string {
	obj := "alignof("
	obj += gen_expr(m.Expr)
	obj += ")"
	return obj
}

func gen_str_constructor_expr_model(m *sema.StrConstructorCallExprModel) string {
	return "jule::to_str(" + gen_expr(m.Expr) + ")"
}

func gen_rune_expr_model(m *sema.RuneExprModel) string {
	if m.Code <= 127 { // ASCII
		b := sbtoa(byte(m.Code))
		if b == "'" {
			b = "\\'"
		}
		return "'" + b + "'"
	}
	return i64toa(int64(m.Code))
}

func gen_builtin_error_trait_sub_ident_expr_model(m *sema.BuiltinErrorTraitSubIdentExprModel) string {
	obj := gen_expr_model(m.Expr)
	obj += ".get()."
	obj += m.Ident
	return obj
}

func gen_explicit_deref_expr_model(m *sema.ExplicitDerefExprModel) string {
	obj := gen_expr_model(m.Expr)
	obj += ".get()"
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

	case *sema.StructSubIdentExprModel:
		return gen_struct_sub_ident_expr_model(m.(*sema.StructSubIdentExprModel))

	case *sema.CommonIdentExprModel:
		return gen_common_ident_expr_model(m.(*sema.CommonIdentExprModel))

	case *sema.CommonSubIdentExprModel:
		return gen_common_sub_ident_expr_model(m.(*sema.CommonSubIdentExprModel))

	case *sema.TupleExprModel:
		return gen_tuple_expr_model(m.(*sema.TupleExprModel))

	case *sema.BuiltinOutCallExprModel:
		return gen_builtin_out_call_expr_model(m.(*sema.BuiltinOutCallExprModel))

	case *sema.BuiltinOutlnCallExprModel:
		return gen_builtin_outln_call_expr_model(m.(*sema.BuiltinOutlnCallExprModel))

	case *sema.BuiltinNewCallExprModel:
		return gen_builtin_new_call_expr_model(m.(*sema.BuiltinNewCallExprModel))

	case *sema.BuiltinRealCallExprModel:
		return gen_builtin_real_call_expr_model(m.(*sema.BuiltinRealCallExprModel))

	case *sema.BuiltinDropCallExprModel:
		return gen_builtin_drop_call_expr_model(m.(*sema.BuiltinDropCallExprModel))

	case *sema.BuiltinPanicCallExprModel:
		return gen_builtin_panic_call_expr_model(m.(*sema.BuiltinPanicCallExprModel))

	case *sema.BuiltinMakeCallExprModel:
		return gen_builtin_make_call_expr_model(m.(*sema.BuiltinMakeCallExprModel))

	case *sema.BuiltinCloneCallExprModel:
		return gen_builtin_clone_call_expr_model(m.(*sema.BuiltinCloneCallExprModel))

	case *sema.SizeofExprModel:
		return gen_sizeof_expr_model(m.(*sema.SizeofExprModel))

	case *sema.AlignofExprModel:
		return gen_alignof_expr_model(m.(*sema.AlignofExprModel))

	case *sema.StrConstructorCallExprModel:
		return gen_str_constructor_expr_model(m.(*sema.StrConstructorCallExprModel))

	case *sema.RuneExprModel:
		return gen_rune_expr_model(m.(*sema.RuneExprModel))

	case *sema.BuiltinErrorTraitSubIdentExprModel:
		return gen_builtin_error_trait_sub_ident_expr_model(m.(*sema.BuiltinErrorTraitSubIdentExprModel))

	case *sema.ExplicitDerefExprModel:
		return gen_explicit_deref_expr_model(m.(*sema.ExplicitDerefExprModel))

	default:
		return "<unimplemented_expression_model>"
	}
}

func gen_expr(e sema.ExprModel) string {
	obj := gen_expr_model(e)

	if obj != "" && obj[0] == '(' {
		switch e.(type) {
		case *sema.BinopExprModel:
			obj = obj[1 : len(obj)-1] // Remove unnecessary parentheses.
		}
	}

	return obj
}

func gen_val(v *sema.Value) string {
	if v.Data.Is_const() {
		return gen_const_expr(v.Data)
	}
	return gen_expr(v.Data.Model)
}

func get_init_expr(t *sema.TypeKind) string {
	if t.Ptr() != nil {
		return "nullptr"
	}

	enm := t.Enm()
	if enm == nil {
		return gen_type_kind(t) + "()"
	}
	return gen_val(enm.Items[0].Value)
}
