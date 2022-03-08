package parser

import (
	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/pkg/x"
)

var builtinFuncs = []*function{
	{
		Ast: ast.FuncAST{
			Id: "out",
			RetType: ast.DataTypeAST{
				Code: x.Void,
			},
			Params: []ast.ParameterAST{{
				Id: "v",
				Type: ast.DataTypeAST{
					Value: "any",
					Code:  x.Any,
				},
			}},
		},
	},
	{
		Ast: ast.FuncAST{
			Id: "outln",
			RetType: ast.DataTypeAST{
				Code: x.Void,
			},
			Params: []ast.ParameterAST{{
				Id: "v",
				Type: ast.DataTypeAST{
					Value: "any",
					Code:  x.Any,
				},
			}},
		},
	},
}
