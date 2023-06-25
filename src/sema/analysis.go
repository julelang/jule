// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
)

// Builds symbol table of AST.
func build_symbols(ast *ast.Ast, importer Importer, owner *_SymbolBuilder) (*SymbolTable, []build.Log) {
	sb := &_SymbolBuilder{
		ast:      ast,
		importer: importer,
		owner:    owner,
	}
	sb.build()

	if len(sb.errors) == 0 {
		return sb.table, nil
	}
	return nil, sb.errors
}

func analyze_package(files []*ast.Ast, importer Importer) (*Package, []build.Log) {
	// Build symbol tables of files.
	tables := make([]*SymbolTable, len(files))
	for i, f := range files {
		table, errors := build_symbols(f, importer, nil)
		if len(errors) > 0 {
			return nil, errors
		}
		tables[i] = table
	}

	sema := _Sema{}
	sema.check(tables)
	if len(sema.errors) > 0 {
		return nil, sema.errors
	}

	pkg := &Package{
		Files: sema.files,
	}

	return pkg, nil
}

// Builds symbol table of package's ASTs.
// Returns nil if files is nil.
// Returns nil if pwd is empty.
// Returns nil if pstd is empty.
// Accepts current working directory is pwd.
//
// Parameters:
//
//	files:    abstract syntax trees of files
//	importer: importer that used for use declarations
//
// Dependent Parameters:
//
//	working-directory: uses working directory path provided by build
//	std-path: uses standard library path provided by build
//
// Risks:
//   - You can pass nil to importer, but panics if importer is nil and
//     semantic analyzer used nil importer.
func Analyze_package(files []*ast.Ast, importer Importer) (*Package, []build.Log) {
	if len(files) == 0 {
		return nil, nil
	}

	return analyze_package(files, importer)
}

// Builds symbol table of AST.
// Returns nil if f is nil.
// Returns nil if pwd is empty.
// Returns nil if pstd is empty.
// Accepts current working directory is pwd.
//
// Parameters:
//
//	f:        file's abstract syntax tree
//	importer: importer that used for use declarations
//
// Dependent Parameters:
//
//	working-directory: uses working directory path provided by build
//	std-path: uses standard library path provided by build
//
// Risks:
//   - You can pass nil to importer, but panics if importer is nil and
//     semantic analyzer used nil importer.
func Analyze_file(f *ast.Ast, importer Importer) (*SymbolTable, []build.Log) {
	files := []*ast.Ast{f}
	pkg, errors := Analyze_package(files, importer)
	if len(errors) > 0 {
		return nil, errors
	}

	// Select first table, because package has only one file.
	// We give just one file.
	table := pkg.Files[0]
	return table, nil
}
