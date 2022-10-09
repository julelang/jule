package parser

import (
	"strings"

	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/pkg/jule"
	"github.com/jule-lang/jule/pkg/juletype"
)

func findGeneric(id string, generics []*GenericType) *GenericType {
	for _, generic := range generics {
		if generic.Id == id {
			return generic
		}
	}
	return nil
}

func typeIsVoid(t Type) bool {
	return t.Id == juletype.VOID && !t.MultiTyped
}

func typeIsVariadicable(t Type) bool {
	return typeIsSlice(t)
}

func typeIsAllowForConst(t Type) bool {
	if !typeIsPure(t) {
		return false
	}
	switch t.Id {
	case juletype.STR, juletype.BOOL:
		return true
	default:
		return juletype.IsNumeric(t.Id)
	}
}

func typeIsStruct(dt Type) bool {
	return dt.Id == juletype.STRUCT
}

func typeIsTrait(dt Type) bool {
	return dt.Id == juletype.TRAIT
}

func typeIsEnum(dt Type) bool {
	return dt.Id == juletype.ENUM
}

func un_ptr_or_ref_type(t Type) Type {
	t.Kind = t.Kind[1:]
	return t
}

func typeHasThisGeneric(generic *GenericType, t Type) bool {
	switch {
	case typeIsFunc(t):
		f := t.Tag.(*Func)
		for _, p := range f.Params {
			if typeHasThisGeneric(generic, p.Type) {
				return true
			}
		}
		return typeHasThisGeneric(generic, f.RetType.Type)
	case t.MultiTyped, typeIsMap(t):
		types := t.Tag.([]Type)
		for _, t := range types {
			if typeHasThisGeneric(generic, t) {
				return true
			}
		}
		return false
	case typeIsSlice(t), typeIsArray(t):
		return typeHasThisGeneric(generic, *t.ComponentType)
	}
	return typeIsThisGeneric(generic, t)
}

func typeHasGenerics(generics []*GenericType, t Type) bool {
	for _, generic := range generics {
		if typeHasThisGeneric(generic, t) {
			return true
		}
	}
	return false
}

func typeIsThisGeneric(generic *GenericType, t Type) bool {
	id, _ := t.KindId()
	return id == generic.Id
}

func typeIsGeneric(generics []*GenericType, t Type) bool {
	if t.Id != juletype.ID {
		return false
	}
	for _, generic := range generics {
		if typeIsThisGeneric(generic, t) {
			return true
		}
	}
	return false
}

func typeIsExplicitPtr(t Type) bool {
	if t.Kind == "" {
		return false
	}
	return t.Kind[0] == '*' && !typeIsUnsafePtr(t)
}

func typeIsUnsafePtr(t Type) bool {
	if t.Id != juletype.UNSAFE {
		return false
	}
	return len(t.Kind)-len(lex.KND_UNSAFE) == 1
}

func typeIsPtr(t Type) bool {
	return typeIsExplicitPtr(t) || typeIsUnsafePtr(t)
}

func typeIsRef(t Type) bool {
	return t.Kind != "" && t.Kind[0] == '&'
}

func typeIsSlice(t Type) bool {
	return t.Id == juletype.SLICE && strings.HasPrefix(t.Kind, jule.PREFIX_SLICE)
}

func typeIsArray(t Type) bool {
	return t.Id == juletype.ARRAY && strings.HasPrefix(t.Kind, jule.PREFIX_ARRAY)
}

func typeIsMap(t Type) bool {
	if t.Kind == "" || t.Id != juletype.MAP {
		return false
	}
	return t.Kind[0] == '[' && t.Kind[len(t.Kind)-1] == ']'
}

func typeIsFunc(t Type) bool {
	return t.Id == juletype.FN &&
		(strings.HasPrefix(t.Kind, lex.KND_FN) ||
			strings.HasPrefix(t.Kind, lex.KND_UNSAFE+" "+lex.KND_FN))
}

// Includes single ptr types.
func typeIsPure(t Type) bool {
	return !typeIsPtr(t) &&
		!typeIsRef(t) &&
		!typeIsSlice(t) &&
		!typeIsArray(t) &&
		!typeIsMap(t) &&
		!typeIsFunc(t)
}

func is_valid_type_for_reference(t Type) bool {
	return !(typeIsTrait(t) ||
		typeIsEnum(t) ||
		typeIsPtr(t) ||
		typeIsRef(t) ||
		typeIsSlice(t) ||
		typeIsArray(t))
}

func type_is_mutable(t Type) bool {
	return typeIsSlice(t) || typeIsPtr(t) || typeIsRef(t)
}

func subIdAccessorOfType(t Type) string {
	if typeIsRef(t) || typeIsPtr(t) {
		return "->"
	}
	return lex.KND_DOT
}

func typeIsNilCompatible(t Type) bool {
	return t.Id == juletype.NIL ||
		typeIsFunc(t) ||
		typeIsPtr(t) ||
		typeIsSlice(t) ||
		typeIsTrait(t) ||
		typeIsMap(t)
}

func typeIsLvalue(t Type) bool {
	return typeIsRef(t) || typeIsPtr(t) || typeIsSlice(t) || typeIsMap(t)
}

func typesEquals(t1, t2 Type) bool {
	return t1.Id == t2.Id && t1.Kind == t2.Kind
}
