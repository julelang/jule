package parser

import (
	"github.com/the-xlang/xxc/ast"
	"github.com/the-xlang/xxc/pkg/x"
)

var builtinFuncs = []*function{
	{
		Ast: ast.Func{
			Id:      "out",
			RetType: ast.DataType{Id: x.Void},
			Params: []ast.Param{{
				Id:    "v",
				Const: true,
				Type: ast.DataType{
					Val: "any",
					Id:  x.Any,
				},
				Default: ast.Expr{Model: exprNode{`""`}},
			}},
		},
	},
	{
		Ast: ast.Func{
			Id:      "outln",
			RetType: ast.DataType{Id: x.Void},
			Params: []ast.Param{{
				Id:    "v",
				Const: true,
				Type: ast.DataType{
					Val: "any",
					Id:  x.Any,
				},
				Default: ast.Expr{Model: exprNode{`""`}},
			}},
		},
	},
}

var strDefs = &defmap{
	Globals: []*ast.Var{
		{
			Id:    "len",
			Const: true,
			Type:  ast.DataType{Id: x.Size, Val: "size"},
			Tag:   "_buffer.size()",
		},
	},
}

var arrDefs = &defmap{
	Globals: []*ast.Var{
		{
			Id:    "len",
			Const: true,
			Type:  ast.DataType{Id: x.Size, Val: "size"},
			Tag:   "_buffer.size()",
		},
	},
}

var mapDefs = &defmap{
	Globals: []*ast.Var{
		{
			Id:    "len",
			Const: true,
			Type:  ast.DataType{Id: x.Size, Val: "size"},
			Tag:   "size()",
		},
	},
	Funcs: []*function{
		{Ast: ast.Func{Id: "clear"}},
		{Ast: ast.Func{Id: "keys"}},
		{Ast: ast.Func{Id: "values"}},
		{Ast: ast.Func{
			Id:      "has",
			Params:  []ast.Param{{Id: "key", Const: true}},
			RetType: ast.DataType{Id: x.Bool, Val: "bool"},
		}},
		{Ast: ast.Func{
			Id:     "del",
			Params: []ast.Param{{Id: "key", Const: true}},
		}},
	},
}

// Use this at before use mapDefs if necessary.
// Because some definitions is responsive for map data-types.
func readyMapDefs(mapt ast.DataType) {
	types := mapt.Tag.([]ast.DataType)
	keyt := types[0]
	valt := types[1]

	keysFunc, _, _ := mapDefs.funcById("keys", nil)
	keysFunc.Ast.RetType = keyt
	keysFunc.Ast.RetType.Val = "[]" + keysFunc.Ast.RetType.Val

	valuesFunc, _, _ := mapDefs.funcById("values", nil)
	valuesFunc.Ast.RetType = valt
	valuesFunc.Ast.RetType.Val = "[]" + valuesFunc.Ast.RetType.Val

	hasFunc, _, _ := mapDefs.funcById("has", nil)
	hasFunc.Ast.Params[0].Type = keyt

	delFunc, _, _ := mapDefs.funcById("del", nil)
	delFunc.Ast.Params[0].Type = keyt
}
