// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import "github.com/julelang/jule/lex"

// SymbolTable must implement Lookup.

// Symbol table.
// Builds by semantic analyzer.
type SymbolTable struct {
	File         *lex.File    // Owner fileset of this symbol table.
	Packages     []*Package   // Imported packages.
	Vars         []*Var       // Variables.
	Type_aliases []*TypeAlias // Type aliases.
	Structs      []*Struct    // Structures.
	Funcs        []*Fn        // Functions.
	Traits       []*Trait     // Traits.
	Enums        []*Enum      // Enums.
	Impls        []*Impl      // Implementations.
}

// Returns package by identifier.
// Returns nil if not exist any package in this identifier.
func (st *SymbolTable) find_package(ident string) *Package {
	for _, pkg := range st.Packages {
		if pkg.Ident == ident {
			return pkg
		}
	}
	return nil
}

// Returns package by selector.
// Returns nil if selector returns false for all packages.
// Returns nil if selector is nil.
func (st *SymbolTable) select_package(selector func(*Package) bool) *Package {
	if selector == nil {
		return nil
	}
	for _, pkg := range st.Packages {
		if selector(pkg) {
			return pkg
		}
	}
	return nil
}

// Returns variable by identifier and cpp linked state.
// Returns nil if not exist any variable in this identifier.
func (st *SymbolTable) find_var(ident string, cpp_linked bool) *Var {
	for _, v := range st.Vars {
		if v.Ident == ident && v.Cpp_linked == cpp_linked {
			return v
		}
	}
	return nil
}

// Returns type alias by identifier and cpp linked state.
// Returns nil if not exist any type alias in this identifier.
func (st *SymbolTable) find_type_alias(ident string, cpp_linked bool) *TypeAlias {
	for _, ta := range st.Type_aliases {
		if ta.Ident == ident && ta.Cpp_linked == cpp_linked {
			return ta
		}
	}
	return nil
}

// Returns struct by identifier and cpp linked state.
// Returns nil if not exist any struct in this identifier.
func (st *SymbolTable) find_struct(ident string, cpp_linked bool) *Struct {
	for _, s := range st.Structs {
		if s.Ident == ident && s.Cpp_linked == cpp_linked {
			return s
		}
	}
	return nil
}

// Returns function by identifier and cpp linked state.
// Returns nil if not exist any function in this identifier.
func (st *SymbolTable) find_fn(ident string, cpp_linked bool) *Fn {
	for _, f := range st.Funcs {
		if f.Ident == ident && f.Cpp_linked == cpp_linked {
			return f
		}
	}
	return nil
}

// Returns trait by identifier.
// Returns nil if not exist any trait in this identifier.
func (st *SymbolTable) find_trait(ident string) *Trait {
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

// Reports this identifier duplicated in symbol table.
// The "self" parameter represents address of exception identifier.
// If founded identifier address equals to self, will be skipped.
func (st *SymbolTable) is_duplicated_ident(self uintptr, ident string, cpp_linked bool) bool {
	for _, v := range st.Vars {
		if _uintptr(v) != self && v.Ident == ident && v.Cpp_linked == cpp_linked {
			return true
		}
	}

	for _, ta := range st.Type_aliases {
		if _uintptr(ta) != self && ta.Ident == ident && ta.Cpp_linked == cpp_linked {
			return true
		}
	}

	for _, s := range st.Structs {
		if _uintptr(s) != self && s.Ident == ident && s.Cpp_linked == cpp_linked {
			return true
		}
	}

	for _, f := range st.Funcs {
		if _uintptr(f) != self && f.Ident == ident && f.Cpp_linked == cpp_linked {
			return true
		}
	}

	if cpp_linked {
		return false
	}

	for _, t := range st.Traits {
		if _uintptr(t) != self && t.Ident == ident {
			return true
		}
	}

	for _, e := range st.Enums {
		if _uintptr(e) != self && e.Ident == ident {
			return true
		}
	}

	return false
}
