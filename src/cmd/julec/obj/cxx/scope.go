package cxx

import "github.com/julelang/jule/sema"

// Generates C++ code of statement.
func gen_st(st sema.St) string {
	switch st.(type) {
	case *sema.Scope:
		return gen_scope(st.(*sema.Scope))

	case *sema.Var:
		return gen_var(st.(*sema.Var))

	case *sema.TypeAlias:
		return "// " + gen_type_alias(st.(*sema.TypeAlias))

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
	return obj
}

// Generates C++ code of function's scope.
func gen_fn_scope(f *sema.FnIns) string {
	// TODO: Add return variables to root scope.
	return gen_scope(f.Scope)
}
