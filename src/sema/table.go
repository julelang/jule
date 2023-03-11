// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

// Symbol table.
// Builds by semantic analyzer.
type SymbolTable struct {
	Packages     []*Package   // Imported packages.
	Vars         []*Var       // Variables.
	Type_aliases []*TypeAlias // Type aliases.
	Structs      []*Struct    // Structures.
	Funcs        []*Fn        // Functions.
	Traits       []*Trait     // Traits.
	Enums        []*Enum      // Enums.
}

// Returns package by identifier.
// Returns nil if not exist any package in this identifier.
func (st *SymbolTable) Find_package(ident string) *Package {
	for _, pkg := range st.Packages {
		if pkg.Ident == ident {
			return pkg
		}
	}
	return nil
}

// Returns package by path.
// Returns nil if not exist any package in this path.
func (st *SymbolTable) Find_package_by_path(path string) *Package {
	for _, pkg := range st.Packages {
		if pkg.Path == path {
			return pkg
		}
	}
	return nil
}

// Returns variable by identifier and cpp linked state.
// Returns nil if not exist any variable in this identifier.
func (st *SymbolTable) Find_var(ident string, cpp_linked bool) *Var {
	for _, v := range st.Vars {
		if v.Ident == ident && v.Cpp_linked == cpp_linked {
			return v
		}
	}
	return nil
}

// Returns type alias by identifier and cpp linked state.
// Returns nil if not exist any type alias in this identifier.
func (st *SymbolTable) Find_type_alias(ident string, cpp_linked bool) *TypeAlias {
	for _, ta := range st.Type_aliases {
		if ta.Ident == ident && ta.Cpp_linked == cpp_linked {
			return ta
		}
	}
	return nil
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
func (st *SymbolTable) Find_enums(ident string) *Enum {
	for _, e := range st.Enums {
		if e.Ident == ident {
			return e
		}
	}
	return nil
}
