// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import "github.com/julelang/jule/lex"

// SymbolTable must implement Lookup.

// Symbol table.
// Builds by semantic analyzer.
type SymbolTable struct {
	File         *lex.File     // Owner fileset of this symbol table.
	Passes       []Pass        // All passed flags with jule:pass directive.
	Imports      []*ImportInfo // Imported packages.
	Vars         []*Var        // Variables.
	Type_aliases []*TypeAlias  // Type aliases.
	Structs      []*Struct     // Structures.
	Funcs        []*Fn         // Functions.
	Traits       []*Trait      // Traits.
	Enums        []*Enum       // Enums.
	Impls        []*Impl       // Implementations.
}

// Returns imported package by identifier.
// Returns nil if not exist any package in this identifier.
func (st *SymbolTable) Find_package(ident string) *ImportInfo {
	for _, pkg := range st.Imports {
		if pkg.Ident == ident {
			return pkg
		}
	}
	return nil
}

// Returns imported package by selector.
// Returns nil if selector returns false for all packages.
// Returns nil if selector is nil.
func (st *SymbolTable) Select_package(selector func(*ImportInfo) bool) *ImportInfo {
	if selector == nil {
		return nil
	}
	for _, pkg := range st.Imports {
		if selector(pkg) {
			return pkg
		}
	}
	return nil
}

func (st *SymbolTable) __find_var(ident string, cpp_linked bool, reverse bool) *Var {
	if reverse {
		for i := len(st.Vars) - 1; i >= 0; i-- {
			v := st.Vars[i]
			if v.Ident == ident && v.Cpp_linked == cpp_linked {
				return v
			}
		}
	} else {
		for _, v := range st.Vars {
			if v.Ident == ident && v.Cpp_linked == cpp_linked {
				return v
			}
		}
	}
	return nil
}

// Returns variable by identifier and cpp linked state.
// Returns nil if not exist any variable in this identifier.
func (st *SymbolTable) Find_var(ident string, cpp_linked bool) *Var {
	return st.__find_var(ident, cpp_linked, false)
}

func (st *SymbolTable) __find_type_alias(ident string, cpp_linked bool, reverse bool) *TypeAlias {
	if reverse {
		for i := len(st.Type_aliases) - 1; i >= 0; i-- {
			ta := st.Type_aliases[i]
			if ta.Ident == ident && ta.Cpp_linked == cpp_linked {
				return ta
			}
		}
	} else {
		for _, ta := range st.Type_aliases {
			if ta.Ident == ident && ta.Cpp_linked == cpp_linked {
				return ta
			}
		}
	}
	return nil
}

// Returns type alias by identifier and cpp linked state.
// Returns nil if not exist any type alias in this identifier.
func (st *SymbolTable) Find_type_alias(ident string, cpp_linked bool) *TypeAlias {
	return st.__find_type_alias(ident, cpp_linked, false)
}

// Returns struct by identifier and cpp linked state.
// Returns nil if not exist any struct in this identifier.
func (st *SymbolTable) Find_struct(ident string, cpp_linked bool) *Struct {
	for _, s := range st.Structs {
		if s.Ident == ident && s.Cpp_linked == cpp_linked {
			return s
		}
	}
	return nil
}

// Returns function by identifier and cpp linked state.
// Returns nil if not exist any function in this identifier.
func (st *SymbolTable) Find_fn(ident string, cpp_linked bool) *Fn {
	for _, f := range st.Funcs {
		if f.Ident == ident && f.Cpp_linked == cpp_linked {
			return f
		}
	}
	return nil
}

// Returns trait by identifier.
// Returns nil if not exist any trait in this identifier.
func (st *SymbolTable) Find_trait(ident string) *Trait {
	for _, t := range st.Traits {
		if t.Ident == ident {
			return t
		}
	}
	return nil
}

// Returns enum by identifier.
// Returns nil if not exist any enum in this identifier.
func (st *SymbolTable) find_enum(ident string) *Enum {
	for _, e := range st.Enums {
		if e.Ident == ident {
			return e
		}
	}
	return nil
}

// Returns define by identifier.
// Returns nil if not exist any define in this identifier.
func (st *SymbolTable) def_by_ident(ident string, cpp_linked bool) any {
	for _, v := range st.Vars {
		if v.Ident == ident && v.Cpp_linked == cpp_linked {
			return v
		}
	}

	for _, ta := range st.Type_aliases {
		if ta.Ident == ident && ta.Cpp_linked == cpp_linked {
			return ta
		}
	}

	for _, s := range st.Structs {
		if s.Ident == ident && s.Cpp_linked == cpp_linked {
			return s
		}
	}

	for _, f := range st.Funcs {
		if f.Ident == ident && f.Cpp_linked == cpp_linked {
			return f
		}
	}

	if cpp_linked {
		return nil
	}

	for _, t := range st.Traits {
		if t.Ident == ident {
			return t
		}
	}

	for _, e := range st.Enums {
		if e.Ident == ident {
			return e
		}
	}

	return nil
}

// Reports this identifier duplicated in symbol table.
// The "self" parameter represents address of exception identifier.
// If founded identifier address equals to self, will be skipped.
func (st *SymbolTable) is_duplicated_ident(itself uintptr, ident string, cpp_linked bool) bool {
	for _, v := range st.Vars {
		if _uintptr(v) != itself && v.Ident == ident && v.Cpp_linked == cpp_linked {
			return true
		}
	}

	for _, ta := range st.Type_aliases {
		if _uintptr(ta) != itself && ta.Ident == ident && ta.Cpp_linked == cpp_linked {
			return true
		}
	}

	for _, s := range st.Structs {
		if _uintptr(s) != itself && s.Ident == ident && s.Cpp_linked == cpp_linked {
			return true
		}
	}

	for _, f := range st.Funcs {
		if _uintptr(f) != itself && f.Ident == ident && f.Cpp_linked == cpp_linked {
			return true
		}
	}

	if cpp_linked {
		return false
	}

	for _, t := range st.Traits {
		if _uintptr(t) != itself && t.Ident == ident {
			return true
		}
	}

	for _, e := range st.Enums {
		if _uintptr(e) != itself && e.Ident == ident {
			return true
		}
	}

	return false
}
