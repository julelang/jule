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
	// Path is the directory path of package to import.
	// Should return abstract syntax tree of package files.
	// Logs accepts as error.
	Import_package(path string) ([]*ast.Ast, []build.Log)

	// Invoked after the package is imported.
	Imported(pkg *Package)
}

// Returns variable by identifier and cpp linked state.
// Returns nil if not exist any variable in this identifier.
func find_var_in_package(files []*SymbolTable,
	ident string, cpp_linked bool) *Var {
	for _, f := range files {
		v := f.find_var(ident, cpp_linked)
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
		ta := f.find_type_alias(ident, cpp_linked)
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
		s := f.find_struct(ident, cpp_linked)
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
		f := f.find_fn(ident, cpp_linked)
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
		t := f.find_trait(ident)
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

// Package must implement Lookup.

// Package.
// Represents imported package by use declaration.
type Package struct {
	// Use declaration token.
	Token lex.Token

	// Absolute path.
	Path string

	// Use declaration path string.
	Link_path string

	// Package identifier (aka package name).
	// Empty if package is cpp header.
	Ident string

	// Is cpp header.
	Cpp bool

	// Is standard library package.
	Std bool

	// Symbol table for each package's file.
	// Nil if package is cpp header.
	Files []*SymbolTable
}

// Returns nil always.
func (p *Package) find_package() *Package { return nil }

// Returns variable by identifier and cpp linked state.
// Returns nil if not exist any variable in this identifier.
func (p *Package) find_var(ident string, cpp_linked bool) *Var {
	return find_var_in_package(p.Files, ident, cpp_linked)
}

// Returns type alias by identifier and cpp linked state.
// Returns nil if not exist any type alias in this identifier.
func (p *Package) find_type_alias(ident string, cpp_linked bool) *TypeAlias {
	return find_type_alias_in_package(p.Files, ident, cpp_linked)
}

// Returns struct by identifier and cpp linked state.
// Returns nil if not exist any struct in this identifier.
func (p *Package) find_struct(ident string, cpp_linked bool) *Struct {
	return find_struct_in_package(p.Files, ident, cpp_linked)
}

// Returns function by identifier and cpp linked state.
// Returns nil if not exist any function in this identifier.
func (p *Package) find_fn(ident string, cpp_linked bool) *Fn {
	return find_fn_in_package(p.Files, ident, cpp_linked)
}

// Returns trait by identifier.
// Returns nil if not exist any trait in this identifier.
func (p *Package) find_trait(ident string) *Trait {
	return find_trait_in_package(p.Files, ident)
}

// Returns enum by identifier.
// Returns nil if not exist any enum in this identifier.
func (p *Package) find_enum(ident string) *Enum {
	return find_enum_in_package(p.Files, ident)
}
