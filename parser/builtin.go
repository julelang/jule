package parser

import (
	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/pkg/x"
)

var builtinFunctions = []*function{
	{
		Name: "_out",
		ReturnType: ast.TypeAST{
			Code: x.Void,
		},
		Params: []ast.ParameterAST{{
			Name: "v",
			Type: ast.TypeAST{
				Value: "any",
				Code:  x.Any,
			},
		}},
	},
	{
		Name: "_outln",
		ReturnType: ast.TypeAST{
			Code: x.Void,
		},
		Params: []ast.ParameterAST{{
			Name: "v",
			Type: ast.TypeAST{
				Value: "any",
				Code:  x.Any,
			},
		}},
	},
}
