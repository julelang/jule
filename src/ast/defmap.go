package ast

import "github.com/julelang/jule/lex"

func IsAccessable(finder *lex.File, target *lex.File, defIsPub bool) bool {
	return defIsPub || finder == nil || target == nil || finder.Dir() == target.Dir()
}

// Defmap is definition map.
type Defmap struct {
	Namespaces []*Namespace
	Enums      []*Enum
	Structs    []*Struct
	Traits     []*Trait
	Types      []*TypeAlias
	Fns        []*Fn
	Globals    []*Var
	Side       *Defmap
}

func (dm *Defmap) FindNsById(id string) int {
	for i, t := range dm.Namespaces {
		if t != nil && t.Id == id {
			return i
		}
	}
	return -1
}

func (dm *Defmap) NsById(id string) *Namespace {
	i := dm.FindNsById(id)
	if i == -1 {
		return nil
	}
	return dm.Namespaces[i]
}

func (dm *Defmap) FindStructById(id string, f *lex.File) (int, *Defmap, bool) {
	for i, s := range dm.Structs {
		if s != nil && s.Id == id {
			if IsAccessable(f, s.Token.File, s.Public) {
				return i, dm, false
			}
		}
	}
	if dm.Side != nil {
		i, m, _ := dm.Side.FindStructById(id, f)
		return i, m, true
	}
	return -1, nil, false
}

func (dm *Defmap) StructById(id string, f *lex.File) (*Struct, *Defmap, bool) {
	i, m, canshadow := dm.FindStructById(id, f)
	if i == -1 {
		return nil, nil, false
	}
	return m.Structs[i], m, canshadow
}

func (dm *Defmap) FindTraitById(id string, f *lex.File) (int, *Defmap, bool) {
	for i, t := range dm.Traits {
		if t != nil && t.Id == id {
			if IsAccessable(f, t.Token.File, t.Public) {
				return i, dm, false
			}
		}
	}
	if dm.Side != nil {
		i, m, _ := dm.Side.FindTraitById(id, f)
		return i, m, true
	}
	return -1, nil, false
}

func (dm *Defmap) TraitById(id string, f *lex.File) (*Trait, *Defmap, bool) {
	i, m, canshadow := dm.FindTraitById(id, f)
	if i == -1 {
		return nil, nil, false
	}
	return m.Traits[i], m, canshadow
}

func (dm *Defmap) FindEnumById(id string, f *lex.File) (int, *Defmap, bool) {
	for i, e := range dm.Enums {
		if e != nil && e.Id == id {
			if IsAccessable(f, e.Token.File, e.Public) {
				return i, dm, false
			}
		}
	}
	if dm.Side != nil {
		i, m, _ := dm.Side.FindEnumById(id, f)
		return i, m, true
	}
	return -1, nil, false
}

func (dm *Defmap) EnumById(id string, f *lex.File) (*Enum, *Defmap, bool) {
	i, m, canshadow := dm.FindEnumById(id, f)
	if i == -1 {
		return nil, nil, false
	}
	return m.Enums[i], m, canshadow
}

func (dm *Defmap) FindTypeById(id string, f *lex.File) (int, *Defmap, bool) {
	for i, t := range dm.Types {
		if t != nil && t.Id == id {
			if IsAccessable(f, t.Token.File, t.Public) {
				return i, dm, false
			}
		}
	}
	if dm.Side != nil {
		i, m, _ := dm.Side.FindTypeById(id, f)
		return i, m, true
	}
	return -1, nil, false
}

func (dm *Defmap) TypeById(id string, f *lex.File) (*TypeAlias, *Defmap, bool) {
	i, m, canshadow := dm.FindTypeById(id, f)
	if i == -1 {
		return nil, nil, false
	}
	return m.Types[i], m, canshadow
}

func (dm *Defmap) FindFnById(id string, f *lex.File) (int, *Defmap, bool) {
	for i, fn := range dm.Fns {
		if fn != nil && fn.Id == id {
			if IsAccessable(f, fn.Token.File, fn.Public) {
				return i, dm, false
			}
		}
	}
	if dm.Side != nil {
		i, m, _ := dm.Side.FindFnById(id, f)
		return i, m, true
	}
	return -1, nil, false
}

// FnById returns function by specified id.
//
// Special case:
//
//	FnById(id) -> nil: if function is not exist.
func (dm *Defmap) FnById(id string, f *lex.File) (*Fn, *Defmap, bool) {
	i, m, canshadow := dm.FindFnById(id, f)
	if i == -1 {
		return nil, nil, false
	}
	return m.Fns[i], m, canshadow
}

func (dm *Defmap) FindGlobalById(id string, f *lex.File) (int, *Defmap, bool) {
	for i, g := range dm.Globals {
		if g != nil && g.DataType.Id != void_t && g.Id == id {
			if IsAccessable(f, g.Token.File, g.Public) {
				return i, dm, false
			}
		}
	}
	if dm.Side != nil {
		i, m, _ := dm.Side.FindGlobalById(id, f)
		return i, m, true
	}
	return -1, nil, false
}

func (dm *Defmap) GlobalById(id string, f *lex.File) (*Var, *Defmap, bool) {
	i, m, canshadow := dm.FindGlobalById(id, f)
	if i == -1 {
		return nil, nil, false
	}
	return m.Globals[i], m, canshadow
}

// FindById returns index of definition with type if exist.
//
// Special case is;
//
//	FindById(id) -> -1, ' ' if id is not exist
//
// Types;
// 'g' -> global
// 'f' -> function
// 'e' -> enum
// 's' -> struct
// 't' -> type alias
// 'i' -> trait
func (dm *Defmap) FindById(id string, f *lex.File) (int, *Defmap, byte) {
	var finders = map[byte]func(string, *lex.File) (int, *Defmap, bool){
		'g': dm.FindGlobalById,
		'f': dm.FindFnById,
		'e': dm.FindEnumById,
		's': dm.FindStructById,
		't': dm.FindTypeById,
		'i': dm.FindTraitById,
	}
	for code, finder := range finders {
		i, m, _ := finder(id, f)
		if i != -1 {
			return i, m, code
		}
	}
	return -1, nil, ' '
}

// PushDefines pushes defines to destination Defmap.
func (dm *Defmap) PushDefines(dest *Defmap) {
	dest.Types = append(dest.Types, dm.Types...)
	dest.Traits = append(dest.Traits, dm.Traits...)
	dest.Structs = append(dest.Structs, dm.Structs...)
	dest.Enums = append(dest.Enums, dm.Enums...)
	dest.Globals = append(dest.Globals, dm.Globals...)
	dest.Fns = append(dest.Fns, dm.Fns...)
}
