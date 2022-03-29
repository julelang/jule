package parser

import (
	"strings"

	"github.com/the-xlang/xxc/ast"
	"github.com/the-xlang/xxc/pkg/x"
)

func typeIsVoidRet(t ast.DataType) bool { return t.Id == x.Void && !t.MultiTyped }

func typeOfArrayElements(t ast.DataType) ast.DataType {
	// Remove array syntax "[]"
	t.Val = t.Val[2:]
	return t
}

func typeIsPtr(t ast.DataType) bool {
	if t.Val == "" {
		return false
	}
	return t.Val[0] == '*'
}

func typeIsArray(t ast.DataType) bool {
	if t.Val == "" {
		return false
	}
	return strings.HasPrefix(t.Val, "[]")
}

func typeIsMap(t ast.DataType) bool {
	if t.Val == "" {
		return false
	}
	return t.Id == x.Map && t.Val[0] == '[' && !strings.HasPrefix(t.Val, "[]")
}

func typeIsFunc(t ast.DataType) bool {
	if t.Id != x.Func || t.Val == "" {
		return false
	}
	return t.Val[0] == '('
}

func typeIsSingle(t ast.DataType) bool {
	return !typeIsPtr(t) &&
		!typeIsArray(t) &&
		!typeIsMap(t) &&
		!typeIsFunc(t)
}

func subIdAccessorOfType(t ast.DataType) string {
	if typeIsPtr(t) {
		return "->"
	}
	return "."
}

func typeIsNilCompatible(t ast.DataType) bool {
	return t.Id == x.Func || typeIsPtr(t) || typeIsArray(t) || typeIsMap(t)
}

func checkArrayCompatiblity(arrT, t ast.DataType) bool {
	if t.Id == x.Nil {
		return true
	}
	return arrT.Val == t.Val
}

func checkMapCompability(mapT, t ast.DataType) bool {
	if t.Id == x.Nil {
		return true
	}
	/*t1types := t1.Tag.([]ast.DataType)
	t2types := t2.Tag.([]ast.DataType)
	if !typesAreCompatible(t1types[0], t2types[0], ignoreany) {
		return false
	}
	return typesAreCompatible(t1types[1], t2types[1], ignoreany)*/
	return mapT.Val == t.Val
}

func typeIsLvalue(t ast.DataType) bool {
	return typeIsPtr(t) || typeIsArray(t) || typeIsMap(t)
}

func typesAreCompatible(t1, t2 ast.DataType, ignoreany bool) bool {
	switch {
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
		return t1.Id == x.Nil || t2.Id == x.Nil
	}
	return x.TypesAreCompatible(t1.Id, t2.Id, ignoreany)
}

func typeIsVariadicable(t ast.DataType) bool { return typeIsArray(t) }
func typeIsMut(t ast.DataType) bool          { return typeIsPtr(t) }
