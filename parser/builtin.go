package parser

import (
	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/pkg/x"
)

var builtinFunctions = []*function{
	{
		ast: ast.FunctionAST{
			Name: "_out",
			ReturnType: ast.DataTypeAST{
				Code: x.Void,
			},
			Params: []ast.ParameterAST{{
				Name: "v",
				Type: ast.DataTypeAST{
					Value: "any",
					Code:  x.Any,
				},
			}},
		},
	},
	{
		ast: ast.FunctionAST{
			Name: "_outln",
			ReturnType: ast.DataTypeAST{
				Code: x.Void,
			},
			Params: []ast.ParameterAST{{
				Name: "v",
				Type: ast.DataTypeAST{
					Value: "any",
					Code:  x.Any,
				},
			}},
		},
	},
}
