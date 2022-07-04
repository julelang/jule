package parser

import (
	"github.com/the-xlang/xxc/pkg/xio"
	"github.com/the-xlang/xxc/pkg/xtype"
)

// Defmap is definition map.
type Defmap struct {
	Namespaces []*namespace
	Enums      []*Enum
	Structs    []*xstruct
	Types      []*Type
	Funcs      []*function
	Globals    []*Var
	parent     *Defmap
}

func (dm *Defmap) findNsById(id string, parent bool) (int, *Defmap) {
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

func (dm *Defmap) nsById(id string, parent bool) *namespace {
	i, m := dm.findNsById(id, parent)
	if i == -1 {
		return nil
	}
	return m.Namespaces[i]
}

func (dm *Defmap) findStructById(id string, f *File) (int, *Defmap, bool) {
	for i, s := range dm.Structs {
		if s != nil && s.Ast.Id == id {
			if s.Ast.Pub || f == nil || f.Dir == s.Ast.Tok.File.Dir {
				return i, dm, false
			}
		}
	}
	if dm.parent != nil {
		i, m, _ := dm.parent.findStructById(id, f)
		return i, m, true
	}
	return -1, nil, false
}

func (dm *Defmap) structById(id string, f *File) (*xstruct, *Defmap, bool) {
	i, m, canshadow := dm.findStructById(id, f)
	if i == -1 {
		return nil, nil, false
	}
	return m.Structs[i], m, canshadow
}

func (dm *Defmap) findEnumById(id string, f *File) (int, *Defmap, bool) {
	for i, e := range dm.Enums {
		if e != nil && e.Id == id {
			if e.Pub || f == nil || f.Dir == e.Tok.File.Dir {
				return i, dm, false
			}
		}
	}
	if dm.parent != nil {
		i, m, _ := dm.parent.findEnumById(id, f)
		return i, m, true
	}
	return -1, nil, false
}

func (dm *Defmap) enumById(id string, f *File) (*Enum, *Defmap, bool) {
	i, m, canshadow := dm.findEnumById(id, f)
	if i == -1 {
		return nil, nil, false
	}
	return m.Enums[i], m, canshadow
}

func (dm *Defmap) findTypeById(id string, f *File) (int, *Defmap, bool) {
	for i, t := range dm.Types {
		if t != nil && t.Id == id {
			if t.Pub || f == nil || f.Dir == t.Tok.File.Dir {
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

func (dm *Defmap) typeById(id string, f *File) (*Type, *Defmap, bool) {
	i, m, canshadow := dm.findTypeById(id, f)
	if i == -1 {
		return nil, nil, false
	}
	return m.Types[i], m, canshadow
}

func (dm *Defmap) findFuncById(id string, f *File) (int, *Defmap, bool) {
	for i, fn := range dm.Funcs {
		if fn != nil && fn.Ast.Id == id {
			if fn.Ast.Pub || f == nil || f.Dir == fn.Ast.Tok.File.Dir {
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
func (dm *Defmap) funcById(id string, f *File) (*function, *Defmap, bool) {
	i, m, canshadow := dm.findFuncById(id, f)
	if i == -1 {
		return nil, nil, false
	}
	return m.Funcs[i], m, canshadow
}

func (dm *Defmap) findGlobalById(id string, f *File) (int, *Defmap, bool) {
	for i, v := range dm.Globals {
		if v != nil && v.Type.Id != xtype.Void && v.Id == id {
			if v.Pub || f == nil || f.Dir == v.IdTok.File.Dir {
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

func (dm *Defmap) globalById(id string, f *File) (*Var, *Defmap, bool) {
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
// 'e' -> enum
// 's' -> struct
func (dm *Defmap) defById(id string, f *File) (int, *Defmap, byte) {
	var finders = map[byte]func(string, *xio.File) (int, *Defmap, bool){
		'g': dm.findGlobalById,
		'f': dm.findFuncById,
		'e': dm.findEnumById,
		's': dm.findStructById,
	}
	for code, finder := range finders {
		i, m, _ := finder(id, f)
		if i != -1 {
			return i, m, code
		}
	}
	return -1, nil, ' '
}
