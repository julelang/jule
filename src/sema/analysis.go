package sema

import (
	"strings"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
)

// Builds symbol table of AST.
func build_symbols(pwd string, pstd string, ast *ast.Ast,
	importer Importer) (*SymbolTable, []build.Log) {
	sb := &_SymbolBuilder{
		ast:      ast,
		importer: importer,
		pwd:      pwd,
		pstd:     pstd,
	}
	sb.build()

	if len(sb.errors) == 0 {
		return sb.table, nil
	}
	return nil, sb.errors
}

func analyze_package(pwd string, pstd string, files []*ast.Ast,
	importer Importer) ([]*SymbolTable, []build.Log){
	// Build symbol tables of files.
	tables := make([]*SymbolTable, len(files))
	for i, f := range files {
		table, errors := build_symbols(pwd, pstd, f, importer)
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

	return sema.tables, nil
}

// Builds symbol table of package's ASTs.
// Returns nil if files is nil.
// Returns nil if pwd is empty.
// Returns nil if pstd is empty.
// Accepts current working directory is pwd.
//
// Parameters:
//   pwd:      working directory path
//   pstd:     standard library directory path
//   files:    abstract syntax trees of files
//   importer: importer that used for use declarations
//
// Risks:
//  - You can pass nil to importer, but panics if importer is nil and
//    semantic analyzer used nil importer.
func Analyze_package(pwd string, pstd string, files []*ast.Ast,
	importer Importer) ([]*SymbolTable, []build.Log) {
	if len(files) == 0 {
		return nil, nil
	}

	pwd = strings.TrimSpace(pwd)
	if pwd == "" {
		return nil, nil
	}

	pstd = strings.TrimSpace(pstd)
	if pstd == "" {
		return nil, nil
	}

	return analyze_package(pwd, pstd, files, importer)
}

// Builds symbol table of AST.
// Returns nil if f is nil.
// Returns nil if pwd is empty.
// Returns nil if pstd is empty.
// Accepts current working directory is pwd.
//
// Parameters:
//   pwd:      working directory path
//   pstd:     standard library directory path
//   f:        file's abstract syntax tree
//   importer: importer that used for use declarations
//
// Risks:
//  - You can pass nil to importer, but panics if importer is nil and
//    semantic analyzer used nil importer.
func Analyze_file(pwd string, pstd string, f *ast.Ast,
	importer Importer) (*SymbolTable, []build.Log) {
	files := []*ast.Ast{f}
	tables, errors := Analyze_package(pwd, pstd, files, importer)
	if len(errors) > 0 {
		return nil, errors
	}

	// Select first table, because package has only one file.
	// We give just one file.
	table := tables[0]
	return table, nil
}
