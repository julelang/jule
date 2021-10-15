package parser

import (
	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
)

func typeIsPointer(t ast.DataTypeAST) bool {
	if t.Value == "" {
		return false
	}
	return t.Value[0] == '*'
}

func typeIsSingle(dt ast.DataTypeAST) bool {
	return !typeIsPointer(dt) && dt.Code != x.Function
}

func defaultValueOfType(t ast.DataTypeAST) string {
	if typeIsPointer(t) {
		return "null"
	}
	return x.DefaultValueOfType(t.Code)
}

func typesAreCompatible(t1, t2 ast.DataTypeAST, ignoreany bool) bool {
	if (typeIsPointer(t1) || typeIsPointer(t2)) &&
		(t1.Code == x.Null || t2.Code == x.Null) {
		return true
	}
	return x.TypesAreCompatible(t1.Code, t2.Code, ignoreany)
}

func (p *Parser) readyType(dt ast.DataTypeAST) (ast.DataTypeAST, bool) {
	switch dt.Code {
	case x.Name:
		t := p.typeByName(dt.Token.Value)
		if t == nil {
			return dt, false
		}
		t.Type.Value = dt.Value[:len(dt.Value)-len(dt.Token.Value)] + t.Type.Value
		return p.readyType(t.Type)
	case x.Function:
		funAST := dt.Tag.(ast.FunctionAST)
		for index, param := range funAST.Params {
			funAST.Params[index].Type, _ = p.readyType(param.Type)
		}
		funAST.ReturnType, _ = p.readyType(funAST.ReturnType)
		dt.Value = dt.Tag.(ast.FunctionAST).DataTypeString()
	}
	return dt, true
}

func (p *Parser) checkType(real, check ast.DataTypeAST, ignoreAny bool, errToken lex.Token) {
	real, ok := p.readyType(real)
	if !ok {
		return
	}
	check, ok = p.readyType(check)
	if !ok {
		return
	}
	if typeIsSingle(real) && typeIsSingle(check) {
		if !x.TypesAreCompatible(check.Code, real.Code, false) {
			p.PushErrorToken(errToken, "incompatible_datatype")
		}
	} else {
		if typeIsPointer(real) && check.Code == x.Null {
			return
		}
		if real.Value != check.Value {
			p.PushErrorToken(errToken, "incompatible_datatype")
		}
	}
}
