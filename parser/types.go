package parser

import (
	"strings"

	"github.com/jule-lang/jule/lex/tokens"
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

func typeIsVoid(t DataType) bool {
	return t.Id == juletype.Void && !t.MultiTyped
}

func typeIsVariadicable(t DataType) bool {
	return typeIsSlice(t)
}

func typeIsAllowForConst(t DataType) bool {
	if !typeIsPure(t) {
		return false
	}
	switch t.Id {
	case juletype.Str, juletype.Bool:
		return true
	default:
		return juletype.IsNumeric(t.Id)
	}
}

func typeIsStruct(dt DataType) bool {
	return dt.Id == juletype.Struct
}

func typeIsTrait(dt DataType) bool {
	return dt.Id == juletype.Trait
}

func typeIsEnum(dt DataType) bool {
	return dt.Id == juletype.Enum
}

func unptrType(t DataType) DataType {
	t.Kind = t.Kind[1:]
	return t
}

func typeHasThisGeneric(generic *GenericType, t DataType) bool {
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
		types := t.Tag.([]DataType)
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

func typeHasGenerics(generics []*GenericType, t DataType) bool {
	for _, generic := range generics {
		if typeHasThisGeneric(generic, t) {
			return true
		}
	}
	return false
}

func typeIsThisGeneric(generic *GenericType, t DataType) bool {
	id, _ := t.KindId()
	return id == generic.Id
}

func typeIsGeneric(generics []*GenericType, t DataType) bool {
	if t.Id != juletype.Id {
		return false
	}
	for _, generic := range generics {
		if typeIsThisGeneric(generic, t) {
			return true
		}
	}
	return false
}

func typeIsExplicitPtr(t DataType) bool {
	if t.Kind == "" {
		return false
	}
	return t.Kind[0] == '*'
}

func typeIsPtr(t DataType) bool {
	return typeIsExplicitPtr(t)
}

func typeIsSlice(t DataType) bool {
	return t.Id == juletype.Slice && strings.HasPrefix(t.Kind, jule.Prefix_Slice)
}

func typeIsArray(t DataType) bool {
	return t.Id == juletype.Array && strings.HasPrefix(t.Kind, jule.Prefix_Array)
}

func typeIsMap(t DataType) bool {
	if t.Kind == "" || t.Id != juletype.Map {
		return false
	}
	return t.Kind[0] == '[' && t.Kind[len(t.Kind)-1] == ']'
}

func typeIsFunc(t DataType) bool {
	return t.Id == juletype.Func && strings.HasPrefix(t.Kind, tokens.FN)
}

// Includes single ptr types.
func typeIsPure(t DataType) bool {
	return !typeIsPtr(t) &&
		!typeIsSlice(t) &&
		!typeIsArray(t) &&
		!typeIsMap(t) &&
		!typeIsFunc(t)
}

func subIdAccessorOfType(t DataType) string {
	if typeIsPtr(t) {
		return "->"
	}
	return tokens.DOT
}

func typeIsNilCompatible(t DataType) bool {
	return t.Id == juletype.Nil ||
		typeIsFunc(t) ||
		typeIsPtr(t) ||
		typeIsSlice(t) ||
		typeIsTrait(t) ||
		typeIsMap(t)
}

func checkSliceCompatiblity(arrT, t DataType) bool {
	if t.Id == juletype.Nil {
		return true
	}
	return arrT.Kind == t.Kind
}

func checkArrayCompatiblity(arrT, t DataType) bool {
	if !typeIsArray(t) {
		return false
	}
	return arrT.Size.N == t.Size.N
}

func checkMapCompability(mapT, t DataType) bool {
	if t.Id == juletype.Nil {
		return true
	}
	return mapT.Kind == t.Kind
}

func typeIsLvalue(t DataType) bool {
	return typeIsPtr(t) || typeIsSlice(t) || typeIsMap(t)
}

func checkPtrCompability(t1, t2 DataType) bool {
	if t2.Id == juletype.Nil {
		return true
	}
	return t1.Kind == t2.Kind
}

func typesEquals(t1, t2 DataType) bool {
	return t1.Id == t2.Id && t1.Kind == t2.Kind
}

func checkTraitCompability(t1, t2 DataType) bool {
	if t2.Id == juletype.Nil {
		return true
	}
	t := t1.Tag.(*trait)
	t1ptr := t1.Pointers()
	switch {
	case typeIsTrait(t2):
		return t == t2.Tag.(*trait) && t1ptr == t2.Pointers()
	case typeIsPtr(t2):
		t2 = unptrType(t2)
		if !typeIsStruct(t2) {
			break
		}
		fallthrough
	case typeIsStruct(t2):
		if t1ptr != "" {
			return false
		}
		t2ptr := t2.Pointers()
		if t2ptr != "" {
			return false
		}
		s := t2.Tag.(*structure)
		return s.hasTrait(t)
	}
	return false
}

func checkStructCompability(t1, t2 DataType) bool {
	s1, s2 := t1.Tag.(*structure), t2.Tag.(*structure)
	switch {
	case s1.Ast.Id != s2.Ast.Id,
		s1.Ast.Tok.File != s2.Ast.Tok.File:
		return false
	}
	if len(s1.Ast.Generics) == 0 {
		return true
	}
	n1, n2 := len(s1.generics), len(s2.generics)
	if n1 != n2 {
		return false
	}
	for i, g1 := range s1.generics {
		g2 := s2.generics[i]
		if !typesEquals(g1, g2) {
			return false
		}
	}
	return true
}

func typesAreCompatible(t1, t2 DataType, ignoreany bool) bool {
	switch {
	case typeIsTrait(t1), typeIsTrait(t2):
		if typeIsTrait(t2) {
			t1, t2 = t2, t1
		}
		return checkTraitCompability(t1, t2)
	case typeIsPtr(t1), typeIsPtr(t2):
		if typeIsPtr(t2) {
			t1, t2 = t2, t1
		}
		return checkPtrCompability(t1, t2)
	case typeIsSlice(t1), typeIsSlice(t2):
		if typeIsSlice(t2) {
			t1, t2 = t2, t1
		}
		return checkSliceCompatiblity(t1, t2)
	case typeIsArray(t1), typeIsArray(t2):
		if typeIsArray(t2) {
			t1, t2 = t2, t1
		}
		return checkArrayCompatiblity(t1, t2)
	case typeIsMap(t1), typeIsMap(t2):
		if typeIsMap(t2) {
			t1, t2 = t2, t1
		}
		return checkMapCompability(t1, t2)
	case typeIsNilCompatible(t1):
		return t2.Id == juletype.Nil
	case typeIsNilCompatible(t2):
		return t1.Id == juletype.Nil
	case typeIsEnum(t1), typeIsEnum(t2):
		return t1.Id == t2.Id && t1.Kind == t2.Kind
	case typeIsStruct(t1), typeIsStruct(t2):
		if t2.Id == juletype.Struct {
			t1, t2 = t2, t1
		}
		return checkStructCompability(t1, t2)
	}
	return juletype.TypesAreCompatible(t1.Id, t2.Id, ignoreany)
}
