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

func get_f64_model(d *sema.Data) string {
	switch {
	case d.Kind.Prim().Is_f32():
		return strconv.FormatFloat(d.Constant.Read_f64(), 'e', -1, 32) + "f"

	default: // 64-bit
		return strconv.FormatFloat(d.Constant.Read_f64(), 'e', -1, 64)
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
		return get_f64_model(d)

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

func gen_expr(d *sema.Value, lkp sema.Lookup) string {
	if d.Data.Is_const() {
		return gen_const_expr(d.Data)
	}

	return ""
}

func get_init_expr(t *sema.TypeKind) string {
	enm := t.Enm()
	if enm == nil {
		return CPP_DEFAULT_EXPR
	}
	return "{" + gen_expr(enm.Items[0].Value, enm.Sema) + "}"
}
