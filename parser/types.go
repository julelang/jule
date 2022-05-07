package parser

import (
	"strings"

	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/xtype"
)

func typeIsVoid(t DataType) bool          { return t.Id == xtype.Void && !t.MultiTyped }
func typeIsVariadicable(t DataType) bool  { return typeIsArray(t) }
func typeIsMut(t DataType) bool           { return typeIsPtr(t) }
func typeIsAllowForConst(t DataType) bool { return typeIsSingle(t) }

func typeOfArrayItems(t DataType) DataType {
	// Remove array syntax "[]"
	t.Val = t.Val[2:]
	return t
}

func typeIsSinglePtr(t DataType) bool {
	return t.Id == xtype.Voidptr || t.Id == xtype.UIntptr || t.Id == xtype.Intptr
}

func typeIsPtr(t DataType) bool {
	if t.Val == "" {
		return false
	}
	return t.Val[0] == '*' || typeIsSinglePtr(t)
}

func typeIsArray(t DataType) bool {
	if t.Val == "" {
		return false
	}
	return strings.HasPrefix(t.Val, "[]")
}

func typeIsMap(t DataType) bool {
	if t.Val == "" {
		return false
	}
	return t.Id == xtype.Map && t.Val[0] == '[' && !strings.HasPrefix(t.Val, "[]")
}

func typeIsFunc(t DataType) bool {
	if t.Id != xtype.Func || t.Val == "" {
		return false
	}
	return t.Val[0] == '('
}

func typeIsSingle(t DataType) bool {
	return !typeIsPtr(t) &&
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
	return typeIsFunc(t) || typeIsPtr(t) || typeIsArray(t) || typeIsMap(t)
}

func checkArrayCompatiblity(arrT, t DataType) bool {
	if t.Id == xtype.Nil {
		return true
	}
	return arrT.Val == t.Val
}

func checkMapCompability(mapT, t DataType) bool {
	if t.Id == xtype.Nil {
		return true
	}
	/*t1types := t1.Tag.([]DataType)
	t2types := t2.Tag.([]DataType)
	if !typesAreCompatible(t1types[0], t2types[0], ignoreany) {
		return false
	}
	return typesAreCompatible(t1types[1], t2types[1], ignoreany)*/
	return mapT.Val == t.Val
}

func typeIsLvalue(t DataType) bool {
	return typeIsPtr(t) || typeIsArray(t) || typeIsMap(t)
}

func checkPtrCompability(t1, t2 DataType) bool {
	if typeIsPtr(t2) {
		return true
	}
	if typeIsSingle(t2) && xtype.IsIntegerType(t2.Id) {
		return true
	}
	return false
}

func typesAreCompatible(t1, t2 DataType, ignoreany bool) bool {
	switch {
	case typeIsPtr(t1) || typeIsPtr(t2):
		if typeIsPtr(t2) {
			t1, t2 = t2, t1
		}
		return checkPtrCompability(t1, t2)
	case typeIsArray(t1) || typeIsArray(t2):
		if typeIsArray(t2) {
			t1, t2 = t2, t1
		}
		return checkArrayCompatiblity(t1, t2)
	case typeIsMap(t1) || typeIsMap(t2):
		if typeIsMap(t2) {
			t1, t2 = t2, t1
		}
		return checkMapCompability(t1, t2)
	case typeIsNilCompatible(t1) || typeIsNilCompatible(t2):
		return t1.Id == xtype.Nil || t2.Id == xtype.Nil
	case t1.Id == xtype.Enum || t2.Id == xtype.Enum:
		return t1.Id == t2.Id && t1.Val == t2.Val
	case t1.Id == xtype.Struct || t2.Id == xtype.Struct:
		return t1.Tag == t2.Tag
	}
	return xtype.TypesAreCompatible(t1.Id, t2.Id, ignoreany)
}
