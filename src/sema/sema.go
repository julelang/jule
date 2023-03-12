// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"unsafe"

	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

// In Jule: (uintptr)(PTR)
func _uintptr[T any](t *T) uintptr { return uintptr(unsafe.Pointer(t)) }

func compiler_err(token lex.Token, key string, args ...any) build.Log {
	return build.Log{
		Type:   build.ERR,
		Row:    token.Row,
		Column: token.Column,
		Path:   token.File.Path(),
		Text:   build.Errorf(key, args...),
	}
}

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
	s.errors = append(s.errors, compiler_err(token, key, args...))
}

// Reports whether define is accessible in the current package.
func (s *_Sema) is_accessible_define(public bool, token lex.Token) bool {
	return public || s.file.File.Dir() == token.File.Dir()
}

// Reports this identifier duplicated in package's global scope.
// The "self" parameter represents address of exception identifier.
// If founded identifier address equals to self, will be skipped.
func (s *_Sema) is_duplicated_ident(self uintptr, ident string, cpp_linked bool) bool {
	for _, f := range s.files {
		if f.is_duplicated_ident(self, ident, cpp_linked) {
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

// Returns package by selector.
// Returns nil if selector returns false for all packages.
// Returns nil if selector is nil.
//
// Lookups:
//  - Current file's imported packages.
func (s *_Sema) select_package(selector func(*Package) bool) *Package {
	return s.file.select_package(selector)
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
func (s *_Sema) find_struct(ident string, cpp_linked bool) *Struct {
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
func (s *_Sema) find_enum(ident string) *Enum {
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

// Checks type, builds result as kind and collect referred type aliases.
// Skips already checked types.
func (s *_Sema) check_type_with_refers(t *Type, referencer *_Referencer) (ok bool) {
	if t.checked() {
		return true
	}
	tc := _TypeChecker{
		s:          s,
		lookup:     s,
		referencer: referencer,
	}
	tc.check(t)
	return t.checked()
}

// Checks type and builds result as kind.
// Skips already checked types.
func (s *_Sema) check_type(t *Type) (ok bool) {
	return s.check_type_with_refers(t, nil)
}

func (s *_Sema) check_type_alias_kind(ta *TypeAlias) (ok bool) {
	ok = s.check_type_with_refers(ta.Kind, &_Referencer{
		ident:  ta.Ident,
		refers: &ta.Refers,
	})
	if ok && ta.Kind.Kind.Arr() != nil && ta.Kind.Kind.Arr().Auto {
		s.push_err(ta.Kind.Decl.Token, "array_auto_sized")
		ok = false
	}
	return
}

func (s *_Sema) check_type_alias(ta *TypeAlias) {
	if lex.Is_ignore_ident(ta.Ident) {
		s.push_err(ta.Token, "ignore_ident")
	} else if s.is_duplicated_ident(_uintptr(ta), ta.Ident, ta.Cpp_linked) {
		s.push_err(ta.Token, "duplicated_ident", ta.Ident)
	}
	s.check_type_alias_kind(ta)
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

func (s *_Sema) check_enum(e *Enum) {
	if lex.Is_ignore_ident(e.Ident) {
		s.push_err(e.Token, "ignore_ident")
	} else if s.is_duplicated_ident(_uintptr(e), e.Ident, false) {
		s.push_err(e.Token, "duplicated_ident", e.Ident)
	}

	if e.Kind != nil {
		if !s.check_type_with_refers(e.Kind, &_Referencer{
			ident:  e.Ident,
			refers: &e.Refers,
		}) {
			return
		}
	} else {
		// Set to default type.
		e.Kind = &Type{
			Decl: nil,
			Kind: &TypeKind{
				kind: &Prim{kind: lex.KND_I32},
			},
		}
	}

	t := e.Kind.Kind.Prim()
	if t == nil {
		s.push_err(e.Token, "invalid_type_source")
		return
	}

	// Check items.
	switch {
	case t.Is_str():
		// TODO: Implement here.

	case types.Is_int(t.Kind()):
		// TODO: Implement here.

	default:
		s.push_err(e.Token, "invalid_type_source")
	}
}

// Checks current package file's enums.
func (s *_Sema) check_enums() (ok bool) {
	for _, e := range s.file.Enums {
		s.check_enum(e)
		
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
	if !ok {
		return false
	}

	ok = s.check_enums()
	if !ok {
		return false
	}

	// TODO: Implement other declarations.

	return true
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
