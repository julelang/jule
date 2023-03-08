package parser

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/julelang/jule"
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// FileInfo stores information about file parsing.
type FileInfo struct {
	File   *lex.File
	Tree   []ast.Node
	Tokens []lex.Token
	Errors []build.Log
}

// PackageInfo stores information about package parsing.
type PackageInfo struct {
	Files []*FileInfo
}

func parse_file(f *lex.File) *FileInfo {
	finf := &FileInfo{
		File: f,
	}

	lex := lex.New(finf.File)
	finf.Tokens = lex.Lex()
	if len(lex.Logs) > 0 {
		finf.Errors = lex.Logs
		return finf
	}

	p := new_parser(finf.Tokens)
	p.build()
	if len(p.errors) > 0 {
		finf.Errors = p.errors
	} else {
		finf.Tree = p.tree
	}

	return finf
}

func ParseFile(path string) (*FileInfo, string) {
	if !build.IsJule(path) {
		return nil, build.Errorf("file_not_jule", path)
	}
	file := lex.NewFile(path)
	if !file.IsOk() {
		return nil, "path is not exist or inaccessible: " + path
	}
	return parse_file(file), ""
}

func ParsePackage(path string) (*PackageInfo, string) {
	dirents, err := os.ReadDir(path)
	if err != nil {
		return nil, err.Error()
	}

	pinfo := &PackageInfo{}
	for _, dirent := range dirents {
		name := dirent.Name()
		// Skip directories.
		if dirent.IsDir() ||
			!strings.HasSuffix(name, jule.EXT) ||
			!build.IsPassFileAnnotation(name) {
			continue
		}
		path := filepath.Join(path, name)
		finfo := parse_file(lex.NewFile(path))
		pinfo.Files = append(pinfo.Files, finfo)
	}
	return pinfo, ""
}
