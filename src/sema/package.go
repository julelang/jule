// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// Importer.
// Used by semantic analyzer for import use declarations.
type Importer interface {
	// Returns *ImportInfo by path.
	// This function accepted as returns already imported and checked package.
	// If returns not-nil value, will be used instead of Import_package
	// if possible and package content is not checked by Sema.
	Get_import(path string) *ImportInfo
	// Path is the directory path of package to import.
	// Should return abstract syntax tree of package files.
	// Logs accepts as error.
	Import_package(path string) ([]*ast.Ast, []build.Log)
	// Invoked after the package is imported.
	Imported(*ImportInfo)
}

// Returns variable by identifier and cpp linked state.
// Returns nil if not exist any variable in this identifier.
func find_var_in_package(files []*SymbolTable,
	ident string, cpp_linked bool) *Var {
	for _, f := range files {
		v := f.Find_var(ident, cpp_linked)
		if v != nil {
			return v
		}
	}
	return nil
}

// Returns type alias by identifier and cpp linked state.
// Returns nil if not exist any type alias in this identifier.
func find_type_alias_in_package(files []*SymbolTable,
	ident string, cpp_linked bool) *TypeAlias {
	for _, f := range files {
		ta := f.Find_type_alias(ident, cpp_linked)
		if ta != nil {
			return ta
		}
	}
	return nil
}

// Returns struct by identifier and cpp linked state.
// Returns nil if not exist any struct in this identifier.
func find_struct_in_package(files []*SymbolTable,
	ident string, cpp_linked bool) *Struct {
	for _, f := range files {
		s := f.Find_struct(ident, cpp_linked)
		if s != nil {
			return s
		}
	}
	return nil
}

// Returns function by identifier and cpp linked state.
// Returns nil if not exist any function in this identifier.
func find_fn_in_package(files []*SymbolTable, ident string, cpp_linked bool) *Fn {
	for _, f := range files {
		f := f.Find_fn(ident, cpp_linked)
		if f != nil {
			return f
		}
	}
	return nil
}

// Returns trait by identifier.
// Returns nil if not exist any trait in this identifier.
func find_trait_in_package(files []*SymbolTable, ident string) *Trait {
	for _, f := range files {
		t := f.Find_trait(ident)
		if t != nil {
			return t
		}
	}
	return nil
}

// Returns enum by identifier.
// Returns nil if not exist any enum in this identifier.
func find_enum_in_package(files []*SymbolTable, ident string) *Enum {
	for _, f := range files {
		e := f.find_enum(ident)
		if e != nil {
			return e
		}
	}
	return nil
}

// ImportInfo must implement Lookup.

// ImportInfo.
// Represents imported package by use declaration.
type ImportInfo struct {
	// Use declaration token.
	Token lex.Token

	// Absolute path.
	Path string

	// Use declaration path string.
	Link_path string

	// Package identifier (aka package name).
	// Empty if package is cpp header.
	Ident string

	// True if imported with Importer.Get_import function.
	Duplicate bool

	// Is cpp header.
	Cpp bool

	// Is standard library package.
	Std bool

	// Is imported all defines implicitly.
	Import_all bool

	// Identifiers of selected definition.
	Selected []lex.Token

	// Nil if package is cpp header.
	Package *Package
}

// Returns nil always.
func (*ImportInfo) Find_package(string) *ImportInfo { return nil }

// Returns always nil.
func (*ImportInfo) Select_package(func(*ImportInfo) bool) *ImportInfo { return nil }

// Returns variable by identifier and cpp linked state.
// Returns nil if not exist any variable in this identifier.
//
// Lookups by import way such as identifier selection.
// Just lookups non-cpp-linked defines.
func (i *ImportInfo) Find_var(ident string, cpp_linked bool) *Var {
	if !i.is_lookupable(ident) {
		return nil
	}

	return find_var_in_package(i.Package.Files, ident, false)
}

// Returns type alias by identifier.
// Returns nil if not exist any type alias in this identifier.
//
// Lookups by import way such as identifier selection.
// Just lookups non-cpp-linked defines.
func (i *ImportInfo) Find_type_alias(ident string, cpp_linked bool) *TypeAlias {
	if !i.is_lookupable(ident) {
		return nil
	}

	return find_type_alias_in_package(i.Package.Files, ident, false)
}

// Returns struct by identifier and cpp linked state.
// Returns nil if not exist any struct in this identifier.
//
// Lookups by import way such as identifier selection.
// Just lookups non-cpp-linked defines.
func (i *ImportInfo) Find_struct(ident string, cpp_linked bool) *Struct {
	if !i.is_lookupable(ident) {
		return nil
	}

	return find_struct_in_package(i.Package.Files, ident, false)
}

// Returns function by identifier and cpp linked state.
// Returns nil if not exist any function in this identifier.
//
// Lookups by import way such as identifier selection.
// Just lookups non-cpp-linked defines.
func (i *ImportInfo) Find_fn(ident string, cpp_linked bool) *Fn {
	if !i.is_lookupable(ident) {
		return nil
	}

	return find_fn_in_package(i.Package.Files, ident, false)
}

// Returns trait by identifier.
// Returns nil if not exist any trait in this identifier.
//
// Lookups by import way such as identifier selection.
func (i *ImportInfo) Find_trait(ident string) *Trait {
	if !i.is_lookupable(ident) {
		return nil
	}

	return find_trait_in_package(i.Package.Files, ident)
}

// Returns enum by identifier.
// Returns nil if not exist any enum in this identifier.
//
// Lookups by import way such as identifier selection.
func (i *ImportInfo) Find_enum(ident string) *Enum {
	if !i.is_lookupable(ident) {
		return nil
	}

	return find_enum_in_package(i.Package.Files, ident)
}

func (i *ImportInfo) is_lookupable(ident string) bool {
	if !i.Import_all {
		if len(i.Selected) > 0 {
			if !i.exist_ident(ident) {
				return false
			}
		}
	}
	return true
}

// Reports whether identifier is selected.
func (i *ImportInfo) exist_ident(ident string) bool {
	for _, sident := range i.Selected {
		if sident.Kind == ident {
			return true
		}
	}

	return false
}

// Package must implement Lookup.

// Package.
type Package struct {
	// Symbol table for each package's file.
	Files []*SymbolTable
}

// Returns nil always.
func (*Package) Find_package(string) *Package { return nil }

// Returns always nil.
func (*Package) Select_package(func(*Package) bool) *Package { return nil }

// Returns variable by identifier and cpp linked state.
// Returns nil if not exist any variable in this identifier.
func (p *Package) Find_var(ident string, cpp_linked bool) *Var {
	return find_var_in_package(p.Files, ident, cpp_linked)
}

// Returns type alias by identifier and cpp linked state.
// Returns nil if not exist any type alias in this identifier.
func (p *Package) Find_type_alias(ident string, cpp_linked bool) *TypeAlias {
	return find_type_alias_in_package(p.Files, ident, cpp_linked)
}

// Returns struct by identifier and cpp linked state.
// Returns nil if not exist any struct in this identifier.
func (p *Package) Find_struct(ident string, cpp_linked bool) *Struct {
	return find_struct_in_package(p.Files, ident, cpp_linked)
}

// Returns function by identifier and cpp linked state.
// Returns nil if not exist any function in this identifier.
func (p *Package) Find_fn(ident string, cpp_linked bool) *Fn {
	return find_fn_in_package(p.Files, ident, cpp_linked)
}

// Returns trait by identifier.
// Returns nil if not exist any trait in this identifier.
func (p *Package) Find_trait(ident string) *Trait {
	return find_trait_in_package(p.Files, ident)
}

// Returns enum by identifier.
// Returns nil if not exist any enum in this identifier.
func (p *Package) Find_enum(ident string) *Enum {
	return find_enum_in_package(p.Files, ident)
}
