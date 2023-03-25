package cxx

import "github.com/julelang/jule/sema"

// Generates C++ code of Tupe TypeKind.
func gen_tuple_kind(t *sema.Tuple) string {
	obj := "std::tuple<"
	for _, t := range t.Types {
		obj += gen_type_kind(t) + ","
	}
	obj = obj[:len(obj)-1] // Remove comma
	return obj + ">"
}

// Generates C++ code of TypeKind.
func gen_type_kind(k *sema.TypeKind) string {
	switch {
	case k.Tup() != nil:
		return gen_tuple_kind(k.Tup())

	default:
		return "[<undefined_type_kind>]"
	}
}
