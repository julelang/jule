package cxx

import "github.com/julelang/jule/sema"

func get_derive_fn_decl_clone(s *sema.Struct) string {
	obj := gen_struct_kind(s)
	obj += " clone(void) const "
	return obj
}

func get_derive_fn_def_clone(s *sema.Struct) string {
	obj := gen_struct_kind(s)
	obj += " " + obj + "::clone(void) const "
	return obj
}
