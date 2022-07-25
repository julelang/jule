package parser

import (
	"strings"

	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xtype"
)

func findGeneric(id string, generics []*GenericType) *GenericType {
	for _, generic := range generics {
		if generic.Id == id {
			return generic
		}
	}
	return nil
}

func arrayIsAutoSized(t DataType) bool {
	return arrayExprIsAutoSized(t.Tag.([][]any)[0][1].(models.Expr))
}

func typeIsVoid(t DataType) bool {
	return t.Id == xtype.Void && !t.MultiTyped
}

func typeIsVariadicable(t DataType) bool {
	return typeIsSlice(t)
}

func typeIsMut(t DataType) bool {
	return typeIsPtr(t)
}

func typeIsAllowForConst(t DataType) bool {
	if !typeIsPure(t) {
		return false
	}
	return t.Id == xtype.Str || xtype.IsNumeric(t.Id)
}

func typeIsStruct(dt DataType) bool {
	return dt.Id == xtype.Struct
}

func typeIsTrait(dt DataType) bool {
	return dt.Id == xtype.Trait
}

func typeIsEnum(dt DataType) bool {
	return dt.Id == xtype.Enum
}

func unptrType(t DataType) DataType {
	t.Kind = t.Kind[1:]
	return t
}

func typeIsGeneric(generics []*GenericType, t DataType) bool {
	if t.Id != xtype.Id {
		return false
	}
	id, _ := t.KindId()
	for _, generic := range generics {
		if id == generic.Id {
			return true
		}
	}
	return false
}

func typeOfSliceComponents(t DataType) DataType {
	// Remove slice syntax
	t.Kind = t.Kind[len(x.Prefix_Slice):]
	return t
}

func typeOfArrayComponents(t DataType) DataType {
	return t.ArrayComponent()
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
	return strings.HasPrefix(t.Kind, x.Prefix_Slice)
}

func typeIsArray(t DataType) bool {
	return strings.HasPrefix(t.Kind, x.Prefix_Array)
}

func typeIsMap(t DataType) bool {
	if t.Kind == "" || t.Id != xtype.Map {
		return false
	}
	return t.Kind[0] == '[' && t.Kind[len(t.Kind)-1] == ']'
}

func typeIsFunc(t DataType) bool {
	if t.Id != xtype.Func || t.Kind == "" {
		return false
	}
	return t.Kind[0] == '('
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
	return t.Id == xtype.Nil ||
		typeIsFunc(t) ||
		typeIsPtr(t) ||
		typeIsSlice(t) ||
		typeIsTrait(t) ||
		typeIsMap(t)
}

func checkSliceCompatiblity(arrT, t DataType) bool {
	if t.Id == xtype.Nil {
		return true
	}
	return arrT.Kind == t.Kind
}

func checkArrayCompatiblity(arrT, t DataType) bool {
	if !typeIsArray(t) {
		return false
	}
	i := arrT.Tag.([][]any)[0][0].(uint64)
	j := t.Tag.([][]any)[0][0].(uint64)
	return i == j
}

func checkMapCompability(mapT, t DataType) bool {
	if t.Id == xtype.Nil {
		return true
	}
	return mapT.Kind == t.Kind
}

func typeIsLvalue(t DataType) bool {
	return typeIsPtr(t) || typeIsSlice(t) || typeIsMap(t)
}

func checkPtrCompability(t1, t2 DataType) bool {
	if t2.Id == xtype.Nil {
		return true
	}
	return t1.Kind == t2.Kind
}

func typesEquals(t1, t2 DataType) bool {
	return t1.Id == t2.Id && t1.Kind == t2.Kind
}

func checkTraitCompability(t1, t2 DataType) bool {
	t := t1.Tag.(*trait)
	switch {
	case typeIsTrait(t2):
		return t == t2.Tag.(*trait)
	case typeIsStruct(t2):
		s := t2.Tag.(*xstruct)
		return s.hasTrait(t)
	}
	return true
}

func checkStructCompability(t1, t2 DataType) bool {
	s1, s2 := t1.Tag.(*xstruct), t2.Tag.(*xstruct)
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
	case typeIsTrait(t1), typeIsTrait(t2):
		if typeIsTrait(t2) {
			t1, t2 = t2, t1
		}
		return checkTraitCompability(t1, t2)
	case typeIsNilCompatible(t1), typeIsNilCompatible(t2):
		return t1.Id == xtype.Nil || t2.Id == xtype.Nil
	case typeIsEnum(t1), typeIsEnum(t2):
		return t1.Id == t2.Id && t1.Kind == t2.Kind
	case typeIsStruct(t1), typeIsStruct(t2):
		if t2.Id == xtype.Struct {
			t1, t2 = t2, t1
		}
		return checkStructCompability(t1, t2)
	}
	return xtype.TypesAreCompatible(t1.Id, t2.Id, ignoreany)
}
