package parser

import (
	"github.com/the-xlang/xxc/pkg/x"
)

var builtinFuncs = []*function{
	{
		Ast: Func{
			Id:      "out",
			RetType: DataType{Id: x.Void},
			Params: []Param{{
				Id:      "v",
				Const:   true,
				Type:    DataType{Val: "any", Id: x.Any},
				Default: Expr{Model: exprNode{`""`}},
			}},
		},
	},
	{
		Ast: Func{
			Id:      "outln",
			RetType: DataType{Id: x.Void},
			Params: []Param{{
				Id:      "v",
				Const:   true,
				Type:    DataType{Val: "any", Id: x.Any},
				Default: Expr{Model: exprNode{`""`}},
			}},
		},
	},
}

var strDefs = &defmap{
	Globals: []*Var{
		{
			Id:    "len",
			Const: true,
			Type:  DataType{Id: x.Size, Val: "size"},
			Tag:   "_buffer.size()",
		},
	},
}

var arrDefs = &defmap{
	Globals: []*Var{
		{
			Id:    "len",
			Const: true,
			Type:  DataType{Id: x.Size, Val: "size"},
			Tag:   "_buffer.size()",
		},
	},
}

var mapDefs = &defmap{
	Globals: []*Var{
		{
			Id:    "len",
			Const: true,
			Type:  DataType{Id: x.Size, Val: "size"},
			Tag:   "size()",
		},
	},
	Funcs: []*function{
		{Ast: Func{Id: "clear"}},
		{Ast: Func{Id: "keys"}},
		{Ast: Func{Id: "values"}},
		{Ast: Func{
			Id:      "has",
			Params:  []Param{{Id: "key", Const: true}},
			RetType: DataType{Id: x.Bool, Val: "bool"},
		}},
		{Ast: Func{
			Id:     "del",
			Params: []Param{{Id: "key", Const: true}},
		}},
	},
}

// Use this at before use mapDefs if necessary.
// Because some definitions is responsive for map data-types.
func readyMapDefs(mapt DataType) {
	types := mapt.Tag.([]DataType)
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
