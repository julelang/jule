package parser

import (
	"github.com/the-xlang/xxc/ast"
)

type defmap struct {
	Namespaces []*namespace
	Types      []*ast.Type
	Funcs      []*function
	Globals    []*ast.Var
	justPub    bool
}

func (dm *defmap) findNsById(id string) int {
	for i, t := range dm.Namespaces {
		if t != nil && t.Id == id {
			return i
		}
	}
	return -1
}

func (dm *defmap) nsById(id string) *namespace {
	i := dm.findNsById(id)
	if i == -1 {
		return nil
	}
	return dm.Namespaces[i]
}

func (dm *defmap) findTypeById(id string) int {
	for i, t := range dm.Types {
		if t != nil && t.Id == id {
			if !dm.justPub || t.Pub {
				return i
			}
		}
	}
	return -1
}

func (dm *defmap) typeById(id string) *ast.Type {
	i := dm.findTypeById(id)
	if i == -1 {
		return nil
	}
	return dm.Types[i]
}

func (dm *defmap) findFuncById(id string) int {
	for i, f := range dm.Funcs {
		if f != nil && f.Ast.Id == id {
			if !dm.justPub || f.Ast.Pub {
				return i
			}
		}
	}
	return -1
}

// funcById returns function by specified id.
//
// Special case:
//  funcById(id) -> nil: if function is not exist.
func (dm *defmap) funcById(id string) *function {
	i := dm.findFuncById(id)
	if i == -1 {
		return nil
	}
	return dm.Funcs[i]
}

func (dm *defmap) findGlobalById(id string) int {
	for i, v := range dm.Globals {
		if v != nil && v.Id == id {
			if !dm.justPub || v.Pub {
				return i
			}
		}
	}
	return -1
}

func (dm *defmap) globalById(id string) *ast.Var {
	i := dm.findGlobalById(id)
	if i == -1 {
		return nil
	}
	return dm.Globals[i]
}

// defById returns index of definition with type if exist.
//
// Special case is;
//  defById(id) -> -1, ' ' if id is not exist
//
// Types;
// 'g' -> global
// 'f' -> function
func (dm *defmap) defById(id string) (int, byte) {
	var i int
	i = dm.findGlobalById(id)
	if i != -1 {
		return i, 'g'
	}
	i = dm.findFuncById(id)
	if i != -1 {
		return i, 'f'
	}
	return -1, ' '
}
