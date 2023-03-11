// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// Semantic analyzer for tables.
// Accepts tables as files of package.
type _Sema struct {
	errors   []build.Log
	files   []*SymbolTable
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

func (s *_Sema) check_import(pkg *Package) {
	if pkg.Cpp || len(pkg.Files) == 0{
		return
	}
	sema := _Sema{}
	sema.check(pkg.Files)
	if len(sema.errors) > 0 {
		s.errors = append(s.errors, sema.errors...)
	}
}

func (s *_Sema) check_imports() {
	for _, file := range s.files {
		for _, pkg := range file.Packages {
			s.check_import(pkg)

			// Break checking if package has error.
			if len(s.errors) > 0 {
				s.push_err(pkg.Token, "used_package_has_errors", pkg.Link_path)
				return
			}
		}
	}
}

func (s *_Sema) check_type_alias(ta *TypeAlias) {
	if lex.Is_ignore_ident(ta.Ident) {
		s.push_err(ta.Token, "ignore_id")
		return
	}
}

func (s *_Sema) check_package_type_aliases(files []*SymbolTable) {
	for _, file := range files {
		for _, ta := range file.Type_aliases {
			s.check_type_alias(ta)
			
			// Break checking if type alias has error.
			if len(s.errors) > 0 {
				return
			}
		}
	}
}

func (s *_Sema) check(files []*SymbolTable) {
	s.files = files
	
	s.check_imports()
	// Break checking if imports has error.
	if len(s.errors) > 0 {
		return
	}

	s.check_package_type_aliases(s.files)
}
