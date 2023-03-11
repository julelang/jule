package sema

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

type _Sema struct {
	pwd    string
	pstd   string
	errors []build.Log
	ast    *ast.Ast
	table  *SymbolTable
}

func (s *_Sema) push_err(token lex.Token, key string, args ...any) {
	s.errors = append(s.errors, build.Log{
		Type:   build.ERR,
		Row:    token.Row,
		Column: token.Column,
		Path:   token.File.Path(),
		Text:   build.Errorf(key, args...),
	})
}

func (s *_Sema) check_cpp_use_decl_path(decl *ast.UseDecl) (ok bool) {
	ext := filepath.Ext(decl.Link_path)
	if !build.Is_valid_header_ext(ext) {
		s.push_err(decl.Token, "invalid_header_ext", ext)
		return false
	}

	save_pwd := func() bool {
		err := os.Chdir(s.pwd)
		if err != nil {
			s.push_err(decl.Token, "pwd_cannot_set", decl.Link_path)
			return false
		}
		return true
	}

	err := os.Chdir(decl.Token.File.Dir())
	if err != nil {
		s.push_err(decl.Token, "use_not_found", decl.Link_path)
		_ = save_pwd()
		return false
	}

	info, err := os.Stat(decl.Link_path)
	// Exist?
	if err != nil || info.IsDir() {
		s.push_err(decl.Token, "use_not_found", decl.Link_path)
		_ = save_pwd()
		return false
	}

	return save_pwd()
}

func (s *_Sema) build_cpp_header_package(decl *ast.UseDecl) *Package {
	if !build.Is_std_header_path(decl.Link_path) {
		ok := s.check_cpp_use_decl_path(decl)
		if !ok {
			return nil
		}
	}

	// Set to absolute path for correct include path
	path, err := filepath.Abs(decl.Link_path)
	if err != nil {
		s.push_err(decl.Token, "use_not_found", decl.Link_path)
	}

	return &Package{
		Path:  path,
		Ident: "",    // Cpp headers haven't identifiers.
		Cpp:   true,
		Std:   false,
		Table: nil,   // Cpp headers haven't symbol table.
	}
}

func (s *_Sema) build_std_package(decl *ast.UseDecl) *Package {
	path := decl.Link_path[len("std::"):] // Skip "std::" prefix.
	path = strings.Replace(path, lex.KND_DBLCOLON, string(filepath.Separator), -1)
	path = filepath.Join(s.pstd, path)
	path, err := filepath.Abs(path)
	if err != nil {
		s.push_err(decl.Token, "use_not_found", decl.Link_path)
		return nil
	}

	info, err := os.Stat(path)
	// Exist?
	if err != nil || !info.IsDir() {
		s.push_err(decl.Token, "use_not_found", decl.Link_path)
		return nil
	}

	// Select last identifier of namespace chain.
	ident := decl.Link_path[strings.LastIndex(decl.Link_path, lex.KND_DBLCOLON)+1:]

	return &Package{
		Path:  path,
		Ident: ident,
		Cpp:   false,
		Std:   true,
		Table: nil,
	}
}

func (s *_Sema) build_package(decl *ast.UseDecl) *Package {
	switch {
	case decl.Cpp:
		return s.build_cpp_header_package(decl)

	case decl.Std:
		return s.build_std_package(decl)

	default:
		return nil
	}
}

func (s *_Sema) check_duplicate_use_decl(pkg *Package, link_path string, error_token lex.Token) (ok bool) {
	// Find package by path to detect cpp header imports.
	lpkg := s.table.Find_package_by_path(pkg.Path)
	if lpkg == nil {
		return true
	}
	s.push_err(error_token, "duplicate_use_decl", link_path)
	return false
}

func (s *_Sema) analyze_use_decl(decl *ast.UseDecl) *Package {
	pkg := s.build_package(decl)
	// Break analysis if error occurs.
	if pkg == nil {
		return nil
	}

	ok := s.check_duplicate_use_decl(pkg, decl.Link_path, decl.Token)
	if !ok {
		return nil
	}

	s.table.Packages = append(s.table.Packages, pkg)

	// TODO: Implement here.
	return nil
}

func (s *_Sema) analyze_use_decls() {
	for _, decl := range s.ast.UseDecls {
		s.analyze_use_decl(decl)

		// Break analysis if error occurs.
		if len(s.errors) > 0 {
			break
		}
	}
}

func (s *_Sema) analyze() {
	s.table = &SymbolTable{}
	s.analyze_use_decls()

	// Break analysis if use declarations has error.
	if len(s.errors) > 0 {
		return
	}
}
