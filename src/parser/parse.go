package parser

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// FileInfo stores information about file parsing.
type FileInfo struct {
	File   *lex.File
	Tree   []ast.Node
	Errors []build.Log
}

// PackageInfo stores information about package parsing.
type PackageInfo struct {
	Files []*FileInfo
}

// Parses fileset's tokens and builds AST.
// Returns nil if f is nil.
func ParseFile(f *lex.File) *FileInfo {
	if f == nil {
		return nil
	}

	finf := &FileInfo{
		File: f,
	}

	tree, errors := parse_fileset(f)
	if errors != nil {
		finf.Errors = errors
	} else {
		finf.Tree = tree
	}

	return finf
}

// Parses fileset's tokens and builds AST.
// Returns nil if filesets is nil.
// Skip fileset if nil.
func ParsePackage(filesets []*lex.File) *PackageInfo {
	if filesets == nil {
		return nil
	}

	pinf := &PackageInfo{}
	for _, f := range filesets {
		if f == nil {
			continue
		}

		finfo := ParseFile(f)
		pinf.Files = append(pinf.Files, finfo)
	}

	return pinf
}

func parse_fileset(f *lex.File) ([]ast.Node, []build.Log) {
	p := parser{
		file: f,
	}
	p.parse()
	return p.tree, p.errors
}
