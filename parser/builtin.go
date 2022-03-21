package parser

import (
	"github.com/the-xlang/xxc/ast"
	"github.com/the-xlang/xxc/pkg/x"
)

var builtinFuncs = []*function{
	{
		Ast: ast.Func{
			Id: "out",
			RetType: ast.DataType{
				Id: x.Void,
			},
			Params: []ast.Parameter{{
				Id:    "v",
				Const: true,
				Type: ast.DataType{
					Val: "any",
					Id:  x.Any,
				},
			}},
		},
	},
	{
		Ast: ast.Func{
			Id: "outln",
			RetType: ast.DataType{
				Id: x.Void,
			},
			Params: []ast.Parameter{{
				Id:    "v",
				Const: true,
				Type: ast.DataType{
					Val: "any",
					Id:  x.Any,
				},
			}},
		},
	},
}

var strDefs = &defmap{
	Globals: []ast.Var{
		{
			Id:    "len",
			Const: true,
			Type:  ast.DataType{Id: x.Size, Val: "size"},
			Tag:   "length()",
		},
	},
}

var arrDefs = &defmap{
	Globals: []ast.Var{
		{
			Id:    "len",
			Const: true,
			Type:  ast.DataType{Id: x.Size, Val: "size"},
			Tag:   "_buffer.size()",
		},
	},
}
