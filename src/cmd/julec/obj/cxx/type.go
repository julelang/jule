package cxx

import (
	"strconv"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/sema"
)

// Reports wherher tag is exist in directives.
func has_directive(directives []*ast.Directive, tag string) bool {
	for _, dr := range directives {
		if dr.Tag == tag {
			return true
		}
	}
	return false
}

// Generates C++ code of Prim TypeKind.
func gen_prim_kind(p *sema.Prim) string {
	return as_jt(p.To_str())
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

// Returns C++ code of reference type with element type.
func as_ref_kind(elem string) string {
	ref := as_jt("ref")
	return ref + "<" + elem + ">"
}

// Generates C++ code of Ref TypeKind.
func gen_ref_kind(r *sema.Ref) string {
	elem := gen_type_kind(r.Elem)
	return as_ref_kind(elem)
}

// Generates C++ code of Ptr TypeKind.
func gen_ptr_kind(p *sema.Ptr) string {
	const CPP_POINTER_MARK = "*"
	if p.Is_unsafe() {
		return "void" + CPP_POINTER_MARK
	}

	elem := gen_type_kind(p.Elem)
	return elem + CPP_POINTER_MARK
}

// Generates C++ code of Enum TypeKind.
func gen_enum_kind(e *sema.Enum) string {
	return gen_type_kind(e.Kind.Kind)
}

func as_slice_kind(elem *sema.TypeKind) string {
	elem_s := gen_type_kind(elem)
	slc := as_jt("slice")
	return slc + "<" + elem_s + ">"
}

// Generates C++ code of Slc TypeKind.
func gen_slice_kind(s *sema.Slc) string {
	return as_slice_kind(s.Elem)
}

// Generates C++ code of Map TypeKind.
func gen_map_kind(m *sema.Map) string {
	key := gen_type_kind(m.Key)
	val := gen_type_kind(m.Val)
	_map := as_jt("map")
	return _map + "<" + key + "," + val + ">"
}

func gen_trait_kind_from_ident(ident string) string {
	trt := as_jt("trait")
	return trt + "<" + ident + ">"
}

// Generates C++ code of Trait TypeKind.
func gen_trait_kind(t *sema.Trait) string {
	ident := trait_out_ident(t)
	return gen_trait_kind_from_ident(ident)
}

// Generates C++ code of Struct TypeKind.
func gen_struct_kind(s *sema.Struct) string {
	rep := ""
	if s.Cpp_linked && !has_directive(s.Directives, build.DIRECTIVE_TYPEDEF) {
		rep += "struct "
	}

	rep += struct_out_ident(s)
	return rep
}

// Generates C++ code of Struct instance TypeKind.
func gen_struct_kind_ins(s *sema.StructIns) string {
	return struct_ins_out_ident(s)
}

// Generates C++ code of Arr TypeKind.
func gen_array_kind(a *sema.Arr) string {
	arr := as_jt("array")
	elem := gen_type_kind(a.Elem)
	size := strconv.Itoa(a.N)
	return arr + "<" + elem + "," + size + ">"
}

func gen_fn_anon_decl(f *sema.FnIns) string {
	decl := gen_fn_ins_result(f)

	decl += "("
	if len(f.Params) > 0 {
		for _, param := range f.Params {
			if param.Decl.Is_self() {
				continue
			}

			decl += gen_param_ins_prototype(param)
			decl += ","
		}
		decl = decl[:len(decl)-1] // Remove last comma.
	} else {
		decl += "void"
	}
	decl += ")"

	return decl
}

// Generates C++ code of Fn TypeKind.
func gen_fn_kind(f *sema.FnIns) string {
	fnc := as_jt("fn")
	decl := gen_fn_anon_decl(f)
	return fnc + "<" + decl + ">"
}

// Generates C++ code of TypeKind.
func gen_type_kind(k *sema.TypeKind) string {
	switch {
	case k.Cpp_linked:
		return k.Cpp_ident

	case k.Prim() != nil:
		return gen_prim_kind(k.Prim())

	case k.Tup() != nil:
		return gen_tuple_kind(k.Tup())

	case k.Ref() != nil:
		return gen_ref_kind(k.Ref())

	case k.Ptr() != nil:
		return gen_ptr_kind(k.Ptr())

	case k.Enm() != nil:
		return gen_enum_kind(k.Enm())

	case k.Slc() != nil:
		return gen_slice_kind(k.Slc())

	case k.Map() != nil:
		return gen_map_kind(k.Map())

	case k.Trt() != nil:
		return gen_trait_kind(k.Trt())

	case k.Strct() != nil:
		return gen_struct_kind_ins(k.Strct())

	case k.Arr() != nil:
		return gen_array_kind(k.Arr())

	case k.Fnc() != nil:
		return gen_fn_kind(k.Fnc())

	default:
		return "[<unimplemented_type_kind>]"
	}
}
