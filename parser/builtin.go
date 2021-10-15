package parser

import (
	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/pkg/x"
)

var builtinFunctions = []*function{
	{
		name: "_out",
		returnType: ast.DataTypeAST{
			Code: x.Void,
		},
		params: []ast.ParameterAST{{
			Name: "v",
			Type: ast.DataTypeAST{
				Value: "any",
				Code:  x.Any,
			},
		}},
	},
	{
		name: "_outln",
		returnType: ast.DataTypeAST{
			Code: x.Void,
		},
		params: []ast.ParameterAST{{
			Name: "v",
			Type: ast.DataTypeAST{
				Value: "any",
				Code:  x.Any,
			},
		}},
	},
}
