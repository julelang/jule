package parser

import (
	"github.com/the-xlang/x/ast"
)

type defmap struct {
	Funcs   []*function
	Globals []ast.Var
	Types   []ast.Type
}

func (dm *defmap) typeIndexById(id string) int {
	for i, t := range dm.Types {
		if t.Id == id {
			return i
		}
	}
	return -1
}

func (dm *defmap) typeById(id string) *ast.Type {
	i := dm.typeIndexById(id)
	if i == -1 {
		return nil
	}
	return &dm.Types[i]
}

func (dm *defmap) funcIndexById(id string) int {
	for i, f := range dm.Funcs {
		if f.Ast.Id == id {
			return i
		}
	}
	return -1
}

// funcById returns function by specified id.
//
// Special case:
//  funcById(id) -> nil: if function is not exist.
func (dm *defmap) funcById(id string) *function {
	i := dm.funcIndexById(id)
	if i == -1 {
		return nil
	}
	return dm.Funcs[i]
}

func (dm *defmap) globalIndexById(id string) int {
	for i, v := range dm.Globals {
		if v.Id == id {
			return i
		}
	}
	return -1
}

func (dm *defmap) globalById(id string) *ast.Var {
	i := dm.globalIndexById(id)
	if i == -1 {
		return nil
	}
	return &dm.Globals[i]
}

// defById returns index of definition with type if exist.
//
// Special case is;
//  defById(id) -> -1, ' ' if id is not exist
//
// Types;
// 'g' -> global
func (dm *defmap) defById(id string) (int, byte) {
	i := dm.globalIndexById(id)
	if i != -1 {
		return i, 'g'
	}
	return -1, ' '
}
