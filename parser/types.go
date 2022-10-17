package parser

import (
	"strings"

	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/pkg/jule"
	"github.com/jule-lang/jule/pkg/juletype"
)

func variadic_to_slice_t(t Type) Type {
	t.Original = nil
	t.ComponentType = new(Type)
	*t.ComponentType = t
	t.Id = juletype.SLICE
	t.Kind = jule.PREFIX_SLICE + t.ComponentType.Kind
	return t
}

func find_generic(id string, generics []*GenericType) *GenericType {
	for _, g := range generics {
		if g.Id == id {
			return g
		}
	}
	return nil
}

func type_is_void(t Type) bool {
	return t.Id == juletype.VOID && !t.MultiTyped
}

func type_is_variadicable(t Type) bool {
	return type_is_slc(t)
}

func type_is_allow_for_const(t Type) bool {
	if !type_is_pure(t) {
		return false
	}
	switch t.Id {
	case juletype.STR, juletype.BOOL:
		return true
	default:
		return juletype.IsNumeric(t.Id)
	}
}

func type_is_struct(dt Type) bool { return dt.Id == juletype.STRUCT }

func type_is_trait(dt Type) bool { return dt.Id == juletype.TRAIT }

func type_is_enum(dt Type) bool { return dt.Id == juletype.ENUM }

func un_ptr_or_ref_type(t Type) Type {
	t.Kind = t.Kind[1:]
	return t
}

func type_has_this_generic(generic *GenericType, t Type) bool {
	switch {
	case type_is_fn(t):
		f := t.Tag.(*Func)
		for _, p := range f.Params {
			if type_has_this_generic(generic, p.Type) {
				return true
			}
		}
		return type_has_this_generic(generic, f.RetType.Type)
	case t.MultiTyped, type_is_map(t):
		types := t.Tag.([]Type)
		for _, t := range types {
			if type_has_this_generic(generic, t) {
				return true
			}
		}
		return false
	case type_is_slc(t), type_is_array(t):
		return type_has_this_generic(generic, *t.ComponentType)
	}
	return type_is_this_generic(generic, t)
}

func type_has_generics(generics []*GenericType, t Type) bool {
	for _, generic := range generics {
		if type_has_this_generic(generic, t) {
			return true
		}
	}
	return false
}

func type_is_this_generic(generic *GenericType, t Type) bool {
	id, _ := t.KindId()
	return id == generic.Id
}

func type_is_generic(generics []*GenericType, t Type) bool {
	if t.Id != juletype.ID {
		return false
	}
	for _, generic := range generics {
		if type_is_this_generic(generic, t) {
			return true
		}
	}
	return false
}

func type_is_explicit_ptr(t Type) bool {
	if t.Kind == "" {
		return false
	}
	return t.Kind[0] == '*' && !type_is_unsafe_ptr(t)
}

func type_is_unsafe_ptr(t Type) bool {
	if t.Id != juletype.UNSAFE {
		return false
	}
	return len(t.Kind)-len(lex.KND_UNSAFE) == 1
}

func type_is_ptr(t Type) bool {
	return type_is_explicit_ptr(t) || type_is_unsafe_ptr(t)
}

func type_is_ref(t Type) bool {
	return t.Kind != "" && t.Kind[0] == '&'
}

func type_is_slc(t Type) bool {
	return t.Id == juletype.SLICE && strings.HasPrefix(t.Kind, jule.PREFIX_SLICE)
}

func type_is_array(t Type) bool {
	return t.Id == juletype.ARRAY && strings.HasPrefix(t.Kind, jule.PREFIX_ARRAY)
}

func type_is_map(t Type) bool {
	if t.Kind == "" || t.Id != juletype.MAP {
		return false
	}
	return t.Kind[0] == '[' && t.Kind[len(t.Kind)-1] == ']'
}

func type_is_fn(t Type) bool {
	return t.Id == juletype.FN &&
		(strings.HasPrefix(t.Kind, lex.KND_FN) ||
			strings.HasPrefix(t.Kind, lex.KND_UNSAFE+" "+lex.KND_FN))
}

// Includes single ptr types.
func type_is_pure(t Type) bool {
	return !type_is_ptr(t) &&
		!type_is_ref(t) &&
		!type_is_slc(t) &&
		!type_is_array(t) &&
		!type_is_map(t) &&
		!type_is_fn(t)
}

func is_valid_type_for_reference(t Type) bool {
	return !(type_is_trait(t) ||
		type_is_enum(t) ||
		type_is_ptr(t) ||
		type_is_ref(t) ||
		type_is_slc(t) ||
		type_is_array(t))
}

func type_is_mutable(t Type) bool {
	return type_is_slc(t) || type_is_ptr(t) || type_is_ref(t)
}

func accessor_of_type(t Type) string {
	if type_is_ref(t) || type_is_ptr(t) {
		return "->"
	}
	return lex.KND_DOT
}

func type_is_nil_compatible(t Type) bool {
	return t.Id == juletype.NIL ||
		type_is_fn(t) ||
		type_is_ptr(t) ||
		type_is_slc(t) ||
		type_is_trait(t) ||
		type_is_map(t)
}

func type_is_lvalue(t Type) bool {
	return type_is_ref(t) || type_is_ptr(t) || type_is_slc(t) || type_is_map(t)
}

func types_equals(t1, t2 Type) bool {
	return t1.Id == t2.Id && t1.Kind == t2.Kind
}
