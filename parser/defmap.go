package parser

import (
	"github.com/the-xlang/xxc/pkg/x"
)

type defmap struct {
	Namespaces []*namespace
	Types      []*Type
	Funcs      []*function
	Globals    []*Var
	parent     *defmap
	justPub    bool
}

func (dm *defmap) findNsById(id string, parent bool) (int, *defmap) {
	for i, t := range dm.Namespaces {
		if t != nil && t.Id == id {
			return i, dm
		}
	}
	if parent && dm.parent != nil {
		return dm.parent.findNsById(id, parent)
	}
	return -1, nil
}

func (dm *defmap) nsById(id string, parent bool) *namespace {
	i, m := dm.findNsById(id, parent)
	if i == -1 {
		return nil
	}
	return m.Namespaces[i]
}

func (dm *defmap) findTypeById(id string, f *File) (int, *defmap, bool) {
	for i, t := range dm.Types {
		if t != nil && t.Id == id {
			if !dm.justPub || f == t.Tok.File || t.Pub {
				return i, dm, false
			}
		}
	}
	if dm.parent != nil {
		i, m, _ := dm.parent.findTypeById(id, f)
		return i, m, true
	}
	return -1, nil, false
}

func (dm *defmap) typeById(id string, f *File) (*Type, *defmap, bool) {
	i, m, canshadow := dm.findTypeById(id, f)
	if i == -1 {
		return nil, nil, false
	}
	return m.Types[i], m, canshadow
}

func (dm *defmap) findFuncById(id string, f *File) (int, *defmap, bool) {
	for i, fn := range dm.Funcs {
		if fn != nil && fn.Ast.Id == id {
			if !dm.justPub || f == fn.Ast.Tok.File || fn.Ast.Pub {
				return i, dm, false
			}
		}
	}
	if dm.parent != nil {
		i, m, _ := dm.parent.findFuncById(id, f)
		return i, m, true
	}
	return -1, nil, false
}

// funcById returns function by specified id.
//
// Special case:
//  funcById(id) -> nil: if function is not exist.
func (dm *defmap) funcById(id string, f *File) (*function, *defmap, bool) {
	i, m, canshadow := dm.findFuncById(id, f)
	if i == -1 {
		return nil, nil, false
	}
	return m.Funcs[i], m, canshadow
}

func (dm *defmap) findGlobalById(id string, f *File) (int, *defmap, bool) {
	for i, v := range dm.Globals {
		if v != nil && v.Type.Id != x.Void && v.Id == id {
			if !dm.justPub || f == v.IdTok.File || v.Pub {
				return i, dm, false
			}
		}
	}
	if dm.parent != nil {
		i, m, _ := dm.parent.findGlobalById(id, f)
		return i, m, true
	}
	return -1, nil, false
}

func (dm *defmap) globalById(id string, f *File) (*Var, *defmap, bool) {
	i, m, canshadow := dm.findGlobalById(id, f)
	if i == -1 {
		return nil, nil, false
	}
	return m.Globals[i], m, canshadow
}

// defById returns index of definition with type if exist.
//
// Special case is;
//  defById(id) -> -1, ' ' if id is not exist
//
// Types;
// 'g' -> global
// 'f' -> function
func (dm *defmap) defById(id string, f *File) (int, *defmap, byte) {
	var i int
	var m *defmap
	i, m, _ = dm.findGlobalById(id, f)
	if i != -1 {
		return i, m, 'g'
	}
	i, m, _ = dm.findFuncById(id, f)
	if i != -1 {
		return i, m, 'f'
	}
	return -1, m, ' '
}
