package parser

import (
	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/pkg/x"
)

func typeIsVoidReturn(t ast.DataTypeAST) bool {
	return t.Code == x.Void && !t.MultiTyped
}

func typeOfArrayElements(t ast.DataTypeAST) ast.DataTypeAST {
	// Remove array syntax "[]"
	t.Value = t.Value[2:]
	return t
}

func typeIsPtr(t ast.DataTypeAST) bool {
	if t.Value == "" {
		return false
	}
	return t.Value[0] == '*'
}

func typeIsArray(t ast.DataTypeAST) bool {
	if t.Value == "" {
		return false
	}
	return t.Value[0] == '['
}

func typeIsSingle(dt ast.DataTypeAST) bool {
	return !typeIsPtr(dt) &&
		!typeIsArray(dt) &&
		dt.Code != x.Function
}

func typeIsNilCompatible(t ast.DataTypeAST) bool {
	return t.Code == x.Function || typeIsPtr(t)
}

func checkArrayCompatiblity(arrT, t ast.DataTypeAST) bool {
	if t.Code == x.Nil {
		return true
	}
	return arrT.Value == t.Value
}

func typeIsLvalue(t ast.DataTypeAST) bool {
	return typeIsPtr(t) || typeIsArray(t)
}

func typesAreCompatible(t1, t2 ast.DataTypeAST, ignoreany bool) bool {
	switch {
	case typeIsArray(t1) || typeIsArray(t2):
		if typeIsArray(t2) {
			t1, t2 = t2, t1
		}
		return checkArrayCompatiblity(t1, t2)
	case typeIsNilCompatible(t1) || typeIsNilCompatible(t2):
		return t1.Code == x.Nil || t2.Code == x.Nil
	}
	return x.TypesAreCompatible(t1.Code, t2.Code, ignoreany)
}

func typeIsVariadicable(t ast.DataTypeAST) bool {
	return typeIsArray(t)
}

func typeIsMut(t ast.DataTypeAST) bool {
	return typeIsPtr(t)
}
