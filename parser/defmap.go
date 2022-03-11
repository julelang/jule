package parser

import (
	"github.com/the-xlang/x/ast"
)

type defmap struct {
	Funcs   []*function
	Globals []ast.Var
	Types   []ast.Type
}

func (dm *defmap) typeById(id string) *ast.Type {
	for _, t := range dm.Types {
		if t.Id == id {
			return &t
		}
	}
	return nil
}

// FuncById returns function by specified name.
//
// Special case:
//  FuncById(name) -> nil: if function is not exist.
func (dm *defmap) FuncById(id string) *function {
	for _, f := range dm.Funcs {
		if f.Ast.Id == id {
			return f
		}
	}
	return nil
}

func (dm *defmap) globalById(id string) *ast.Var {
	for _, v := range dm.Globals {
		if v.Id == id {
			return &v
		}
	}
	return nil
}
