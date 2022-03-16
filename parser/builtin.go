package parser

import (
	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/pkg/x"
)

var builtinFuncs = []*function{
	{
		Ast: ast.Func{
			Id: "out",
			RetType: ast.DataType{
				Code: x.Void,
			},
			Params: []ast.Parameter{{
				Id:    "v",
				Const: true,
				Type: ast.DataType{
					Value: "any",
					Code:  x.Any,
				},
			}},
		},
	},
	{
		Ast: ast.Func{
			Id: "outln",
			RetType: ast.DataType{
				Code: x.Void,
			},
			Params: []ast.Parameter{{
				Id:    "v",
				Const: true,
				Type: ast.DataType{
					Value: "any",
					Code:  x.Any,
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
			Type:  ast.DataType{Code: x.Size, Value: "size"},
			Tag:   "length()",
		},
	},
}
