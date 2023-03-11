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
