// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"unsafe"

	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// In Jule: (uintptr)(PTR)
func _uintptr[T any](t *T) uintptr { return uintptr(unsafe.Pointer(t)) }

// Sema must implement Lookup.

// Semantic analyzer for tables.
// Accepts tables as files of package.
type _Sema struct {
	errors []build.Log
	files  []*SymbolTable // Package files.
	file   *SymbolTable   // Current package file.
}

func (s *_Sema) set_current_file(f *SymbolTable) { s.file = f }

func (s *_Sema) push_err(token lex.Token, key string, args ...any) {
	s.errors = append(s.errors, build.Log{
		Type:   build.ERR,
		Row:    token.Row,
		Column: token.Column,
		Path:   token.File.Path(),
		Text:   build.Errorf(key, args...),
	})
}

// Reports whether define is accessible in the current package.
func (s *_Sema) is_accessible_define(public bool, token lex.Token) bool {
	return public || s.file.File.Dir() == token.File.Dir()
}

// Reports this identifier duplicated in package's global scope.
// The "self" parameter represents address of exception identifier.
// If founded identifier address equals to self, will be skipped.
func (s *_Sema) is_duplicate_identifier(self uintptr, ident string, cpp_linked bool) bool {
	is_duplicated := func(f *SymbolTable) bool {
		for _, v := range f.Vars {
			if _uintptr(v) != self && v.Ident == ident && v.Cpp_linked == cpp_linked {
				return true
			}
		}

		for _, ta := range f.Type_aliases {
			if _uintptr(ta) != self && ta.Ident == ident && ta.Cpp_linked == cpp_linked {
				return true
			}
		}

		for _, s := range f.Structs {
			if _uintptr(s) != self && s.Ident == ident && s.Cpp_linked == cpp_linked {
				return true
			}
		}

		for _, f := range f.Funcs {
			if _uintptr(f) != self && f.Ident == ident && f.Cpp_linked == cpp_linked {
				return true
			}
		}

		if cpp_linked {
			return false
		}

		for _, t := range f.Traits {
			if _uintptr(t) != self && t.Ident == ident {
				return true
			}
		}

		for _, e := range f.Enums {
			if _uintptr(e) != self && e.Ident == ident {
				return true
			}
		}

		return false
	}

	for _, f := range s.files {
		if is_duplicated(f) {
			return true
		}
	}

	return false
}

// Returns package by identifier.
// Returns nil if not exist any package in this identifier.
//
// Lookups:
//  - Current file's imported packages.
func (s *_Sema) find_package(ident string) *Package {
	return s.file.find_package(ident)
}

// Returns variable by identifier and cpp linked state.
// Returns nil if not exist any variable in this identifier.
//
// Lookups:
//  - Package file's symbol table.
//  - Current file's public denifes of imported packages.
func (s *_Sema) find_var(ident string, cpp_linked bool) *Var {
	// Lookup package files.
	v := find_var_in_package(s.files, ident, cpp_linked)
	if v != nil {
		return v
	}

	// Lookup current file's public denifes of imported packages.
	for _, pkg := range s.file.Packages {
		v := pkg.find_var(ident, cpp_linked)
		if v != nil && s.is_accessible_define(v.Public, v.Token) {
			return v
		}
	}

	return nil
}

// Returns type alias by identifier and cpp linked state.
// Returns nil if not exist any type alias in this identifier.
//
// Lookups:
//  - Package file's symbol table.
//  - Current file's public denifes of imported packages.
func (s *_Sema) find_type_alias(ident string, cpp_linked bool) *TypeAlias {
	// Lookup package files.
	ta := find_type_alias_in_package(s.files, ident, cpp_linked)
	if ta != nil {
		return ta
	}

	// Lookup current file's public denifes of imported packages.
	for _, pkg := range s.file.Packages {
		ta := pkg.find_type_alias(ident, cpp_linked)
		if ta != nil && s.is_accessible_define(ta.Public, ta.Token) {
			return ta
		}
	}

	return nil
}

// Returns struct by identifier and cpp linked state.
// Returns nil if not exist any struct in this identifier.
//
// Lookups:
//  - Package file's symbol table.
//  - Current file's public denifes of imported packages.
func (s *_Sema) find__struct(ident string, cpp_linked bool) *Struct {
	// Lookup package files.
	strct := find_struct_in_package(s.files, ident, cpp_linked)
	if strct != nil {
		return strct
	}

	// Lookup current file's public denifes of imported packages.
	for _, pkg := range s.file.Packages {
		strct := pkg.find_struct(ident, cpp_linked)
		if strct != nil && s.is_accessible_define(strct.Public, strct.Token) {
			return strct
		}
	}

	return nil
}

// Returns function by identifier and cpp linked state.
// Returns nil if not exist any function in this identifier.
//
// Lookups:
//  - Package file's symbol table.
//  - Current file's public denifes of imported packages.
func (s *_Sema) find_fn(ident string, cpp_linked bool) *Fn {
	// Lookup package files.
	f := find_fn_in_package(s.files, ident, cpp_linked)
	if f != nil {
		return f
	}

	// Lookup current file's public denifes of imported packages.
	for _, pkg := range s.file.Packages {
		f := pkg.find_fn(ident, cpp_linked)
		if f != nil && s.is_accessible_define(f.Public, f.Token) {
			return f
		}
	}

	return nil
}

// Returns trait by identifier.
// Returns nil if not exist any trait in this identifier.
//
// Lookups:
//  - Package file's symbol table.
//  - Current file's public denifes of imported packages.
func (s *_Sema) find_trait(ident string) *Trait {
	// Lookup package files.
	t := find_trait_in_package(s.files, ident)
	if t != nil {
		return t
	}

	// Lookup current file's public denifes of imported packages.
	for _, pkg := range s.file.Packages {
		t := pkg.find_trait(ident)
		if t != nil && s.is_accessible_define(t.Public, t.Token) {
			return t
		}
	}

	return nil
}

// Returns enum by identifier.
// Returns nil if not exist any enum in this identifier.
//
// Lookups:
//  - Package file's symbol table.
//  - Current file's public denifes of imported packages.
func (s *_Sema) find_enums(ident string) *Enum {
	// Lookup package files.
	e := find_enum_in_package(s.files, ident)
	if e != nil {
		return e
	}

	// Lookup current file's public denifes of imported packages.
	for _, pkg := range s.file.Packages {
		e := pkg.find_enum(ident)
		if e != nil && s.is_accessible_define(e.Public, e.Token) {
			return e
		}
	}

	return nil
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
		s.push_err(ta.Token, "ignore_ident")
	} else if s.is_duplicate_identifier(_uintptr(ta), ta.Ident, ta.Cpp_linked) {
		s.push_err(ta.Token, "duplicated_ident", ta.Ident)
	}

	// TODO: Detect cycles.
	// TODO: Check type validity.
}

// Checks current package file's type aliases.
func (s *_Sema) check_type_aliases() (ok bool) {
	for _, ta := range s.file.Type_aliases {
		s.check_type_alias(ta)
		
		// Break checking if type alias has error.
		if len(s.errors) > 0 {
			return false
		}
	}
	return true
}

// Checks current package file.
// Reports whether checking is success.
func (s *_Sema) check_file() (ok bool) {
	ok = s.check_type_aliases()
	return ok
}

// Checks all package files.
// Breaks checking if checked file failed.
func (s *_Sema) check_package_files() {
	for _, f := range s.files {
		s.set_current_file(f)
		ok := s.check_file()
		if !ok {
			return
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

	s.check_package_files()
}
