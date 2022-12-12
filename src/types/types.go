package types

import (
	"strings"

	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/pkg/jule"
	"github.com/julelang/jule/pkg/juletype"
)

type Type = models.Type
type GenericType = models.GenericType
type Fn = models.Fn

func VariadicToSlice(t Type) Type {
	t.Original = nil
	t.ComponentType = new(Type)
	*t.ComponentType = t
	t.Id = juletype.SLICE
	t.Kind = jule.PREFIX_SLICE + t.ComponentType.Kind
	return t
}

func FindGeneric(id string, generics []*GenericType) *GenericType {
	for _, g := range generics {
		if g.Id == id {
			return g
		}
	}
	return nil
}

func IsVoid(t Type) bool {
	return t.Id == juletype.VOID && !t.MultiTyped
}

func IsAllowForConst(t Type) bool {
	if !IsPure(t) {
		return false
	}
	switch t.Id {
	case juletype.STR, juletype.BOOL:
		return true
	default:
		return juletype.IsNumeric(t.Id)
	}
}

func IsVariadicable(t Type) bool { return IsSlice(t) }
func IsStruct(t Type) bool { return t.Id == juletype.STRUCT }
func IsTrait(t Type) bool { return t.Id == juletype.TRAIT }
func IsEnum(t Type) bool { return t.Id == juletype.ENUM }

func DerefPtrOrRef(t Type) Type {
	t.Kind = t.Kind[1:]
	return t
}

func HasThisGeneric(generic *GenericType, t Type) bool {
	switch {
	case IsFn(t):
		f := t.Tag.(*Fn)
		for _, p := range f.Params {
			if HasThisGeneric(generic, p.Type) {
				return true
			}
		}
		return HasThisGeneric(generic, f.RetType.Type)
	case t.MultiTyped, IsMap(t):
		types := t.Tag.([]Type)
		for _, t := range types {
			if HasThisGeneric(generic, t) {
				return true
			}
		}
		return false
	case IsSlice(t), IsArray(t):
		return HasThisGeneric(generic, *t.ComponentType)
	}
	return IsThisGeneric(generic, t)
}

func HasGenerics(generics []*GenericType, t Type) bool {
	for _, g := range generics {
		if HasThisGeneric(g, t) {
			return true
		}
	}
	return false
}

func IsThisGeneric(generic *GenericType, t Type) bool {
	id, _ := t.KindId()
	return id == generic.Id
}

func IsGeneric(generics []*GenericType, t Type) bool {
	if t.Id != juletype.ID {
		return false
	}
	for _, generic := range generics {
		if IsThisGeneric(generic, t) {
			return true
		}
	}
	return false
}

func IsExplicitPtr(t Type) bool {
	if t.Kind == "" {
		return false
	}
	return t.Kind[0] == '*' && !IsUnsafePtr(t)
}

func IsUnsafePtr(t Type) bool {
	if t.Id != juletype.UNSAFE {
		return false
	}
	return len(t.Kind)-len(lex.KND_UNSAFE) == 1
}

func IsPtr(t Type) bool {
	return IsExplicitPtr(t) || IsUnsafePtr(t)
}

func IsRef(t Type) bool {
	return t.Kind != "" && t.Kind[0] == '&'
}

func IsSlice(t Type) bool {
	return t.Id == juletype.SLICE && strings.HasPrefix(t.Kind, jule.PREFIX_SLICE)
}

func IsArray(t Type) bool {
	return t.Id == juletype.ARRAY && strings.HasPrefix(t.Kind, jule.PREFIX_ARRAY)
}

func IsMap(t Type) bool {
	if t.Kind == "" || t.Id != juletype.MAP {
		return false
	}
	return t.Kind[0] == '[' && t.Kind[len(t.Kind)-1] == ']'
}

func IsFn(t Type) bool {
	return t.Id == juletype.FN &&
		(strings.HasPrefix(t.Kind, lex.KND_FN) ||
			strings.HasPrefix(t.Kind, lex.KND_UNSAFE+" "+lex.KND_FN))
}

// Includes single ptr types.
func IsPure(t Type) bool {
	return !IsPtr(t) &&
		!IsRef(t) &&
		!IsSlice(t) &&
		!IsArray(t) &&
		!IsMap(t) &&
		!IsFn(t)
}

func ValidForRef(t Type) bool {
	return !(IsTrait(t) ||
		IsEnum(t) ||
		IsPtr(t) ||
		IsRef(t) ||
		IsSlice(t) ||
		IsArray(t))
}

func IsMut(t Type) bool {
	return IsSlice(t) || IsPtr(t) || IsRef(t)
}

func GetAccessor(t Type) string {
	if IsRef(t) || IsPtr(t) {
		return "->"
	}
	return lex.KND_DOT
}

func IsNilCompatible(t Type) bool {
	return t.Id == juletype.NIL ||
		IsFn(t) ||
		IsPtr(t) ||
		IsSlice(t) ||
		IsTrait(t) ||
		IsMap(t)
}

func IsLvalue(t Type) bool {
	return IsRef(t) || IsPtr(t) || IsSlice(t) || IsMap(t)
}

func Equals(t1, t2 Type) bool {
	return t1.Id == t2.Id && t1.Kind == t2.Kind
}
