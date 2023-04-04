package cxx

import "github.com/julelang/jule/sema"

// Generates C++ code of Prim TypeKind.
func gen_prim_kind(p *sema.Prim) string {
	return p.To_str()
}

// Generates C++ code of Tupe TypeKind.
func gen_tuple_kind(t *sema.Tuple) string {
	obj := "std::tuple<"
	for _, t := range t.Types {
		obj += gen_type_kind(t) + ","
	}
	obj = obj[:len(obj)-1] // Remove comma
	return obj + ">"
}

// Generates C++ code of Ref TypeKind.
func gen_ref_kind(r *sema.Ref) string {
	elem := gen_type_kind(r.Elem)
	ref := as_jt("ref")
	return ref + "<" + elem + ">"
}

// Generates C++ code of Ptr TypeKind.
func gen_ptr_kind(p *sema.Ptr) string {
	const CPP_POINTER_MARK = "*"

	elem := gen_type_kind(p.Elem)
	return elem + CPP_POINTER_MARK
}

// Generates C++ code of TypeKind.
func gen_type_kind(k *sema.TypeKind) string {
	switch {
	case k.Prim() != nil:
		return gen_prim_kind(k.Prim())

	case k.Tup() != nil:
		return gen_tuple_kind(k.Tup())

	case k.Ref() != nil:
		return gen_ref_kind(k.Ref())
	
	case k.Ptr() != nil:
		return gen_ptr_kind(k.Ptr())

	default:
		return "[<undefined_type_kind>]"
	}
}
