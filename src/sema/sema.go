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
	ast    *ast.Ast
	pwd    string
	pstd   string
	errors []build.Log
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

func (s *_Sema) check_cpp_use_decl_path(decl *ast.UseDecl) *ImportInfo {
	if build.Is_std_header_path(decl.LinkString) {
		return nil
	}

	ext := filepath.Ext(decl.LinkString)
	if !build.Is_valid_header_ext(ext) {
		s.push_err(decl.Token, "invalid_header_ext", ext)
		return nil
	}

	err := os.Chdir(decl.Token.File.Dir())
	if err != nil {
		s.push_err(decl.Token, "use_not_found", decl.LinkString)
		return nil
	}

	// Save pwd.
	defer func() {
		err := os.Chdir(s.pwd)
		if err != nil {
			s.push_err(decl.Token, "pwd_cannot_set", decl.LinkString)
		}
	}()

	info, err := os.Stat(decl.LinkString)
	// Exist?
	if err != nil || info.IsDir() {
		s.push_err(decl.Token, "use_not_found", decl.LinkString)
		return nil
	}

	// Set to absolute path for correct include path
	abs, err := filepath.Abs(decl.LinkString)
	if err != nil {
		s.push_err(decl.Token, "use_not_found", decl.LinkString)
	}

	return &ImportInfo{
		Path: abs,
		Cpp:  true,
		Std:  false,
	}
}

func (s *_Sema) check_std_use_decl_path(decl *ast.UseDecl) *ImportInfo {
	path := decl.LinkString[len("std::"):] // Skip "std::" prefix.
	path = strings.Replace(path, lex.KND_DBLCOLON, string(filepath.Separator), -1)
	path = filepath.Join(s.pstd, path)
	path, err := filepath.Abs(path)
	if err != nil {
		s.push_err(decl.Token, "use_not_found", decl.LinkString)
		return nil
	}

	info, err := os.Stat(path)
	// Exist?
	if err != nil || !info.IsDir() {
		s.push_err(decl.Token, "use_not_found", decl.LinkString)
		return nil
	}
	return &ImportInfo{
		Path: path,
		Cpp:  false,
		Std:  true,
	}
}

func (s *_Sema) check_use_decl_path(decl *ast.UseDecl) *ImportInfo {
	switch {
	case decl.Cpp:
		return s.check_cpp_use_decl_path(decl)

	case decl.Std:
		return s.check_std_use_decl_path(decl)

	default:
		return nil
	}
}

func (s *_Sema) analyze_use_decl(decl *ast.UseDecl) *ImportInfo {
	info := s.check_use_decl_path(decl)
	// Break analysis if error occurs.
	if info == nil {
		return nil
	}

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
	s.analyze_use_decls()

	// Break analysis if use declarations has error.
	if len(s.errors) > 0 {
		return
	}
}
