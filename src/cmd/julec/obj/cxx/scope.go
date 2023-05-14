package cxx

import (
	"unsafe"

	"github.com/julelang/jule/sema"
)

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
