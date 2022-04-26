package parser

import "github.com/the-xlang/xxc/pkg/xtype"

// Defmap is definition map.
type Defmap struct {
	Namespaces []*namespace
	Types      []*Type
	Enums      []*Enum
	Funcs      []*function
	Globals    []*Var
	parent     *Defmap
	justPub    bool
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

func (dm *Defmap) findTypeById(id string, f *File) (int, *Defmap, bool) {
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

func (dm *Defmap) typeById(id string, f *File) (*Type, *Defmap, bool) {
	i, m, canshadow := dm.findTypeById(id, f)
	if i == -1 {
		return nil, nil, false
	}
	return m.Types[i], m, canshadow
}

func (dm *Defmap) findEnumById(id string, f *File) (int, *Defmap, bool) {
	for i, t := range dm.Enums {
		if t != nil && t.Id == id {
			if !dm.justPub || f == t.Tok.File || t.Pub {
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

func (dm *Defmap) findFuncById(id string, f *File) (int, *Defmap, bool) {
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
func (dm *Defmap) defById(id string, f *File) (int, *Defmap, byte) {
	var i int
	var m *Defmap
	i, m, _ = dm.findGlobalById(id, f)
	if i != -1 {
		return i, m, 'g'
	}
	i, m, _ = dm.findFuncById(id, f)
	if i != -1 {
		return i, m, 'f'
	}
	i, m, _ = dm.findEnumById(id, f)
	if i != -1 {
		return i, m, 'e'
	}
	return -1, m, ' '
}
