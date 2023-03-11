package sema

import (
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// Semantic analyzer for tables.
// Accepts tables as files of package.
type _Sema struct {
	errors   []build.Log
	tables   []*SymbolTable
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
	for _, file := range s.tables {
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

func (s *_Sema) check(tables []*SymbolTable) {
	s.tables = tables
	
	s.check_imports()
	// Break checking if imports has error.
	if len(s.errors) > 0 {
		return
	}


}
