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
				Id: "v",
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
				Id: "v",
				Type: ast.DataType{
					Value: "any",
					Code:  x.Any,
				},
			}},
		},
	},
}
