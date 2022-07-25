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
	Traits     []*trait
	Types      []*Type
	Funcs      []*function
	Globals    []*Var
}

func (dm *Defmap) findNsById(id string) int {
	for i, t := range dm.Namespaces {
		if t != nil && t.Id == id {
			return i
		}
	}
	return -1
}

func (dm *Defmap) nsById(id string) *namespace {
	i := dm.findNsById(id)
	if i == -1 {
		return nil
	}
	return dm.Namespaces[i]
}

func (dm *Defmap) findStructById(id string, f *File) (int, bool) {
	for i, s := range dm.Structs {
		if s != nil && s.Ast.Id == id {
			if s.Ast.Pub || f == nil || f.Dir == s.Ast.Tok.File.Dir {
				return i, false
			}
		}
	}
	return -1, false
}

func (dm *Defmap) structById(id string, f *File) (*xstruct, bool) {
	i, canshadow := dm.findStructById(id, f)
	if i == -1 {
		return nil, false
	}
	return dm.Structs[i], canshadow
}

func (dm *Defmap) findTraitById(id string, f *File) (int, bool) {
	for i, t := range dm.Traits {
		if t != nil && t.Ast.Id == id {
			if t.Ast.Pub || f == nil || f.Dir == t.Ast.Tok.File.Dir {
				return i, false
			}
		}
	}
	return -1, false
}

func (dm *Defmap) traitById(id string, f *File) (*trait, bool) {
	i, canshadow := dm.findTraitById(id, f)
	if i == -1 {
		return nil, false
	}
	return dm.Traits[i], canshadow
}

func (dm *Defmap) findEnumById(id string, f *File) (int, bool) {
	for i, e := range dm.Enums {
		if e != nil && e.Id == id {
			if e.Pub || f == nil || f.Dir == e.Tok.File.Dir {
				return i, false
			}
		}
	}
	return -1, false
}

func (dm *Defmap) enumById(id string, f *File) (*Enum, bool) {
	i, canshadow := dm.findEnumById(id, f)
	if i == -1 {
		return nil, false
	}
	return dm.Enums[i], canshadow
}

func (dm *Defmap) findTypeById(id string, f *File) (int, bool) {
	for i, t := range dm.Types {
		if t != nil && t.Id == id {
			if t.Pub || f == nil || f.Dir == t.Tok.File.Dir {
				return i, false
			}
		}
	}
	return -1, false
}

func (dm *Defmap) typeById(id string, f *File) (*Type, bool) {
	i, canshadow := dm.findTypeById(id, f)
	if i == -1 {
		return nil, false
	}
	return dm.Types[i], canshadow
}

func (dm *Defmap) findFuncById(id string, f *File) (int, bool) {
	for i, fn := range dm.Funcs {
		if fn != nil && fn.Ast.Id == id {
			if fn.Ast.Pub || f == nil || f.Dir == fn.Ast.Tok.File.Dir {
				return i, false
			}
		}
	}
	return -1, false
}

// funcById returns function by specified id.
//
// Special case:
//  funcById(id) -> nil: if function is not exist.
func (dm *Defmap) funcById(id string, f *File) (*function, bool) {
	i, canshadow := dm.findFuncById(id, f)
	if i == -1 {
		return nil, false
	}
	return dm.Funcs[i], canshadow
}

func (dm *Defmap) findGlobalById(id string, f *File) (int, bool) {
	for i, v := range dm.Globals {
		if v != nil && v.Type.Id != xtype.Void && v.Id == id {
			if v.Pub || f == nil || f.Dir == v.IdTok.File.Dir {
				return i, false
			}
		}
	}
	return -1, false
}

func (dm *Defmap) globalById(id string, f *File) (*Var, bool) {
	i, canshadow := dm.findGlobalById(id, f)
	if i == -1 {
		return nil, false
	}
	return dm.Globals[i], canshadow
}

// findById returns index of definition with type if exist.
//
// Special case is;
//  findById(id) -> -1, ' ' if id is not exist
//
// Types;
// 'g' -> global
// 'f' -> function
// 'e' -> enum
// 's' -> struct
// 't' -> type alias
// 'i' -> trait
func (dm *Defmap) findById(id string, f *File) (int, byte) {
	var finders = map[byte]func(string, *xio.File) (int, bool){
		'g': dm.findGlobalById,
		'f': dm.findFuncById,
		'e': dm.findEnumById,
		's': dm.findStructById,
		't': dm.findTypeById,
		'i': dm.findTraitById,
	}
	for code, finder := range finders {
		i, _ := finder(id, f)
		if i != -1 {
			return i, code
		}
	}
	return -1, ' '
}
