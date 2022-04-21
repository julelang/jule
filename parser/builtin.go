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

var strDefs = &Defmap{
	Globals: []*Var{
		{
			Id:    "len",
			Const: true,
			Type:  DataType{Id: x.Size, Val: "size"},
			Tag:   "_buffer.size()",
		},
	},
}

var arrDefs = &Defmap{
	Globals: []*Var{
		{
			Id:    "len",
			Const: true,
			Type:  DataType{Id: x.Size, Val: "size"},
			Tag:   "_buffer.size()",
		},
	},
	Funcs: []*function{
		{Ast: Func{Id: "clear"}},
		{Ast: Func{
			Id:     "find",
			Params: []Param{{Id: "value"}},
		}},
		{Ast: Func{
			Id:     "find_last",
			Params: []Param{{Id: "value"}},
		}},
		{Ast: Func{
			Id:     "erase",
			Params: []Param{{Id: "value"}},
		}},
		{Ast: Func{
			Id:     "erase_all",
			Params: []Param{{Id: "value"}},
		}},
		{Ast: Func{
			Id:     "append",
			Params: []Param{{Id: "items", Variadic: true}},
		}},
		{Ast: Func{
			Id: "insert",
			Params: []Param{
				{Id: "start", Type: DataType{Id: x.Size, Val: "size"}},
				{Id: "items", Variadic: true},
			},
			RetType: DataType{Id: x.Bool, Val: "bool"},
		}},
	},
}

func readyArrDefs(arrt DataType) {
	elemType := typeOfArrayElements(arrt)

	findFunc, _, _ := arrDefs.funcById("find", nil)
	findFunc.Ast.Params[0].Type = elemType
	findFunc.Ast.RetType = elemType
	findFunc.Ast.RetType.Val = "*" + findFunc.Ast.RetType.Val

	findLastFunc, _, _ := arrDefs.funcById("find_last", nil)
	findLastFunc.Ast.Params[0].Type = elemType
	findLastFunc.Ast.RetType = elemType
	findLastFunc.Ast.RetType.Val = "*" + findLastFunc.Ast.RetType.Val

	eraseFunc, _, _ := arrDefs.funcById("erase", nil)
	eraseFunc.Ast.Params[0].Type = elemType

	eraseAllFunc, _, _ := arrDefs.funcById("erase_all", nil)
	eraseAllFunc.Ast.Params[0].Type = elemType

	appendFunc, _, _ := arrDefs.funcById("append", nil)
	appendFunc.Ast.Params[0].Type = elemType

	insertFunc, _, _ := arrDefs.funcById("insert", nil)
	insertFunc.Ast.Params[1].Type = elemType
}

var mapDefs = &Defmap{
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
