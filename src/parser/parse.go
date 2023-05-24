// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package parser

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// FileInfo stores information about file parsing.
type FileInfo struct {
	Ast    *ast.Ast
	Errors []build.Log
}

// PackageInfo stores information about package parsing.
type PackageInfo struct {
	Files []*FileInfo
}

// Parses fileset's tokens and builds AST.
// Returns nil if f is nil.
func Parse_file(f *lex.File) *FileInfo {
	if f == nil {
		return nil
	}

	finf := &FileInfo{}
	finf.Ast, finf.Errors = parse_fileset(f)
	if finf.Errors != nil {
		finf.Ast = nil
	}

	return finf
}

// Parses fileset's tokens and builds AST.
// Returns nil if filesets is nil.
// Skip fileset if nil.
func Parse_package(filesets []*lex.File) *PackageInfo {
	if filesets == nil {
		return nil
	}

	pinf := &PackageInfo{}
	for _, f := range filesets {
		if f == nil {
			continue
		}

		finfo := Parse_file(f)
		pinf.Files = append(pinf.Files, finfo)
	}

	return pinf
}

func parse_fileset(f *lex.File) (*ast.Ast, []build.Log) {
	p := _Parser{}
	p.parse(f)
	return p.ast, p.errors
}
