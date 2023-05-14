package cxx

import (
	"unsafe"

	"github.com/julelang/jule/sema"
)

type _RangeSetter interface {
	setup_vars(key_a *sema.Var, key_b *sema.Var) string
	next_steps(key_a *sema.Var, key_b *sema.Var, begin string) string
}

type _IndexRangeSetter struct {}

func (*_IndexRangeSetter) setup_vars(key_a *sema.Var, key_b *sema.Var) string {
	indent := indent()

	obj := ""
	if key_a != nil {
		obj += gen_var(key_a)
		obj += var_out_ident(key_a)
		obj += " = 0;\n"
		obj += indent
	}

	if key_b != nil {
		obj += gen_var(key_b)
		obj += var_out_ident(key_b)
		obj += " = *__julec_range_begin;\n"
		obj += indent
	}

	return obj
}

func (*_IndexRangeSetter) next_steps(key_a *sema.Var, key_b *sema.Var, begin string) string {
	indent := indent()
	
	obj := "++__julec_range_begin;\n"
	obj += indent
	
	obj += "if (__julec_range_begin != __julec_range_end) { "
	if key_a != nil {
		obj += "++" + var_out_ident(key_a) + "; "
	}
	if key_b != nil {
		obj += var_out_ident(key_b) + " = *__julec_range_begin; "
	}

	obj += "goto " + begin + "; }\n"
	return obj
}

type _MapRangeSetter struct {}

func (*_MapRangeSetter) setup_vars(key_a *sema.Var, key_b *sema.Var) string {
	indent := indent()
	obj := ""

	if key_a != nil {
		obj += gen_var(key_a)
		obj += var_out_ident(key_a)
		obj += " = __julec_range_begin->first;\n"
		obj += indent
	}

	if key_b != nil {
		obj += gen_var(key_b)
		obj += var_out_ident(key_b)
		obj += " = __julec_range_begin->second;\n"
		obj += indent
	}

	return obj
}

func (*_MapRangeSetter) next_steps(key_a *sema.Var, key_b *sema.Var, begin string) string {
	indent := indent()

	obj := "++__julec_range_begin;\n"
	obj += indent
	
	obj += "if (__julec_range_begin != __julec_range_end) { "
	if key_a != nil {
		obj += var_out_ident(key_a)
		obj += " = __julec_range_begin->first; "
	}
	if key_b != nil {
		obj += var_out_ident(key_b)
		obj += " = __julec_range_begin->second; "
	}

	obj += "goto " + begin + "; }\n"

	return obj
}

// In Jule: (uintptr)(PTR)
func _uintptr[T any](t *T) uintptr { return uintptr(unsafe.Pointer(t)) }

func gen_if(i *sema.If) string {
	obj := "if ("
	obj += gen_expr_model(i.Expr)
	obj += ") "
	obj += gen_scope(i.Scope)
	return obj
}

func gen_conditional(c *sema.Conditional) string {
	obj := gen_if(c.If)

	for _, elif := range c.Elifs {
		obj += " else "
		obj += gen_if(elif)
	}

	if c.Default != nil {
		obj += " else "
		obj += gen_scope(c.Default.Scope)
	}

	return obj
}

func gen_inf_iter(it *sema.InfIter) string {
	begin := iter_begin_label_ident(_uintptr(it))
	end := iter_end_label_ident(_uintptr(it))
	next := iter_next_label_ident(_uintptr(it))
	indent := indent()

	obj := begin + ":;\n"
	obj += indent
	obj += gen_scope(it.Scope)
	obj += "\n"
	obj += indent
	obj += next + ":;\n"
	obj += indent
	obj += "goto " + begin + ";\n"
	obj += indent
	obj += end + ":;"

	return obj
}

func gen_while_iter(it *sema.WhileIter) string {
	begin := iter_begin_label_ident(_uintptr(it))
	end := iter_end_label_ident(_uintptr(it))
	next := iter_next_label_ident(_uintptr(it))
	indent := indent()

	obj := begin + ":;\n"
	obj += indent
	obj += "if (!("
	obj += gen_expr_model(it.Expr)
	obj += ")) { goto "
	obj += end
	obj += "; }\n"
	obj += indent
	obj += gen_scope(it.Scope)
	obj += "\n"
	obj += indent
	obj += next + ":;\n"
	obj += indent
	obj += "goto " + begin + ";\n"
	obj += indent
	obj += end + ":;"

	return obj
}

func get_range_setter(it *sema.RangeIter) _RangeSetter {
	switch {
	case it.Expr.Kind.Slc() != nil:
		return &_IndexRangeSetter{}

	case it.Expr.Kind.Arr() != nil:
		return &_IndexRangeSetter{}

	case it.Expr.Kind.Map() != nil:
		return &_MapRangeSetter{}

	default: // Str
		return &_IndexRangeSetter{}
	}
}

func gen_range_iter(it *sema.RangeIter) string {
	add_indent()

	begin := iter_begin_label_ident(_uintptr(it))
	end := iter_end_label_ident(_uintptr(it))
	next := iter_next_label_ident(_uintptr(it))
	_indent := indent()
	setter := get_range_setter(it)

	obj := "{\n"
	obj += _indent
	obj += "auto __julec_range_expr = "
	obj += gen_expr_model(it.Expr.Model) + ";\n"
	obj += _indent
	obj += "if (__julec_range_expr.begin() != __julec_range_expr.end()) {\n"

	add_indent()
	_indent = indent()

	obj += _indent
	obj += "auto __julec_range_begin = __julec_range_expr.begin();\n"
	obj += _indent
	obj += "const auto __julec_range_end = __julec_range_expr.end();\n"
	obj += _indent
	obj += setter.setup_vars(it.Key_a, it.Key_b)
	obj += begin + ":;\n"
	obj += _indent
	obj += gen_scope(it.Scope)
	obj += "\n"
	obj += _indent
	obj += next + ":;\n"
	obj += _indent
	obj += setter.next_steps(it.Key_a, it.Key_b, begin)
	obj += _indent
	obj += end + ":;\n"

	done_indent()
	_indent = indent()

	obj += _indent
	obj += "}\n"

	done_indent()
	_indent = indent()

	obj += _indent
	obj += "}"

	return obj
}

// Generates C++ code of statement.
func gen_st(st sema.St) string {
	switch st.(type) {
	case *sema.Scope:
		return gen_scope(st.(*sema.Scope))

	case *sema.Var:
		return gen_var(st.(*sema.Var))

	case *sema.TypeAlias:
		return "// " + gen_type_alias(st.(*sema.TypeAlias))

	case *sema.Data:
		return gen_expr_model(st.(*sema.Data).Model) + ";"

	case *sema.Conditional:
		return gen_conditional(st.(*sema.Conditional))

	case *sema.InfIter:
		return gen_inf_iter(st.(*sema.InfIter))

	case *sema.WhileIter:
		return gen_while_iter(st.(*sema.WhileIter))

	case *sema.RangeIter:
		return gen_range_iter(st.(*sema.RangeIter))

	default:
		return "<unimplemented stmt>"
	}
}

// Generates C++ code of scope.
func gen_scope(s *sema.Scope) string {
	obj := "{\n"
	add_indent()

	for _, st := range s.Stmts {
		obj += indent()
		obj += gen_st(st)
		obj += "\n"
	}

	done_indent()
	obj += indent()
	obj += "}"
	
	if s.Deferred {
		obj = "__JULEC_DEFER(" + obj + ");"
	}

	return obj
}

// Generates C++ code of function's scope.
func gen_fn_scope(f *sema.FnIns) string {
	// TODO: Add return variables to root scope.
	return gen_scope(f.Scope)
}
