package cxx

import "github.com/julelang/jule/sema"

// Generates C++ code of scope.
func gen_scope(s *sema.Scope) string {
	return "{}"
}

// Generates C++ code of function's scope.
func gen_fn_scope(f *sema.FnIns) string {
	// TODO: Add return variables to root scope.
	return "{}"
}
