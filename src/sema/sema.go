// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"unsafe"

	"github.com/julelang/jule/ast"
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
func (s *_Sema) check_type_with_refers(t *TypeSymbol, referencer *_Referencer) (ok bool) {
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
func (s *_Sema) check_type(t *TypeSymbol) (ok bool) {
	return s.check_type_with_refers(t, nil)
}

// Evaluates expression with type prefixed Eval and returns result.
func (s *_Sema) evalp(expr *ast.Expr, p *TypeSymbol) *Data {
	e := _Eval{
		s:      s,
		lookup: s,
	}

	if p != nil {
		e.prefix = p.Kind
	}

	return e.eval(expr)
}

// Evaluates expression with Eval and returns result.
func (s *_Sema) eval(expr *ast.Expr) *Data { return s.evalp(expr, nil) }

func (s *_Sema) check_type_alias_decl_kind(ta *TypeAlias) (ok bool) {
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

func (s *_Sema) check_type_alias_decl(ta *TypeAlias) {
	if lex.Is_ignore_ident(ta.Ident) {
		s.push_err(ta.Token, "ignore_ident")
	} else if s.is_duplicated_ident(_uintptr(ta), ta.Ident, ta.Cpp_linked) {
		s.push_err(ta.Token, "duplicated_ident", ta.Ident)
	}
	s.check_type_alias_decl_kind(ta)
}

// Checks current package file's type alias declarations.
func (s *_Sema) check_type_alias_decls() (ok bool) {
	for _, ta := range s.file.Type_aliases {
		s.check_type_alias_decl(ta)
		
		// Break checking if type alias has error.
		if len(s.errors) > 0 {
			return false
		}
	}
	return true
}

func (s *_Sema) check_enum_decl(e *Enum) {
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
		e.Kind = &TypeSymbol{
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

	case types.Is_int(t.To_str()):
		// TODO: Implement here.

	default:
		s.push_err(e.Token, "invalid_type_source")
	}
}

// Checks current package file's enum declarations.
func (s *_Sema) check_enum_decls() (ok bool) {
	for _, e := range s.file.Enums {
		s.check_enum_decl(e)
		
		// Break checking if type alias has error.
		if len(s.errors) > 0 {
			return false
		}
	}
	return true
}

func (s *_Sema) check_decl_generics(generics []*ast.Generic) (ok bool) {
	ok = true
	for i, g := range generics {
		if lex.Is_ignore_ident(g.Ident) {
			s.push_err(g.Token, "ignore_ident")
			ok = false
			continue
		}

		// Check duplications.
	duplication_lookup:
		for j, ct := range generics {
			switch {
			case j >= i:
				// Skip current and following generics.
				break duplication_lookup

			case g.Ident == ct.Ident:
				s.push_err(g.Token, "duplicated_ident", g.Ident)
				ok = false
				break duplication_lookup
			}
		}
	}
	return
}

func (s *_Sema) check_fn_decl_params_dup(f *Fn) (ok bool) {
	ok = true
check:
	for i, p := range f.Params {
		// Lookup in generics.
		for _, g := range f.Generics {
			if p.Ident == g.Ident {
				ok = false
				s.push_err(p.Token, "duplicated_ident", p.Ident)
				continue check
			}
		}

	params_lookup:
		for j, jp := range f.Params {
			switch {
			case j >= i:
				// Skip current and following parameters.
				break params_lookup

			case lex.Is_anon_ident(p.Ident) || lex.Is_anon_ident(jp.Ident):
				// Skip anonymous parameters.
				break params_lookup

			case p.Ident == jp.Ident:
				ok = false
				s.push_err(p.Token, "duplicated_ident", p.Ident)
				continue check
			}
		}
	}
	return
}

func (s *_Sema) check_fn_decl_result_dup(f *Fn) (ok bool) {
	ok = true
	
	if f.Is_void() {
		return
	}

	// Check duplications.
	for i, v := range f.Result.Idents {
		if lex.Is_ignore_ident(v.Kind) {
			continue // Skip anonymous return variables.
		}

		// Lookup in generics.
		for _, g := range f.Generics {
			if v.Kind == g.Ident {
				goto exist
			}
		}

		// Lookup in parameters.
		for _, p := range f.Params {
			if v.Kind == p.Ident {
				goto exist
			}
		}

		// Lookup in return identifiers.
	itself_lookup:
		for j, jv := range f.Result.Idents {
			switch {
			case j >= i:
				// Skip current and following identifiers.
				break itself_lookup

			case jv.Kind == v.Kind:
				goto exist
			}
		}
		continue
	exist:
		s.push_err(v, "duplicated_ident", v.Kind)
		ok = false
	}

	return
}

// Checks generics, parameters and return type.
// Not checks scope, and other things.
func (s *_Sema) check_fn_decl_prototype(f *Fn) (ok bool) {
	// TODO:
	//  - Check return type.
	//  |- Check parameter types.
	//  |- Build non-generic types if function has generic types.
	switch {
	case !s.check_decl_generics(f.Generics):
		return false

	case !s.check_fn_decl_params_dup(f):
		return false

	case !s.check_fn_decl_result_dup(f):
		return false

	default:
		return true
	}
}

func (s *_Sema) check_trait_decl_method(f *Fn) {
	if lex.Is_ignore_ident(f.Ident) {
		s.push_err(f.Token, "ignore_ident")
	}

	s.check_fn_decl_prototype(f)
}

func (s *_Sema) check_trait_decl_methods(t *Trait) {
	for i, f := range t.Methods {
		s.check_trait_decl_method(f)
		
		// Break checking if type alias has error.
		if len(s.errors) > 0 {
			return
		}

		// Check duplications.
	duplicate_lookup:
		for j, jf := range t.Methods {
			// NOTE:
			//  Ignore identifier checking is unnecessary here.
			//  Because ignore identifiers logs error.
			//  Errors breaks checking, so here is unreachable code for
			//  ignore identified methods.
			switch {
			case j >= i:
				// Skip current and following methods.
				break duplicate_lookup
			
			case f.Ident == jf.Ident:
				s.push_err(f.Token, "duplicated_ident", f.Ident)
				break duplicate_lookup
			}
		}
	}
}

func (s *_Sema) check_trait_decl(t *Trait) {
	if lex.Is_ignore_ident(t.Ident) {
		s.push_err(t.Token, "ignore_ident")
	} else if s.is_duplicated_ident(_uintptr(t), t.Ident, false) {
		s.push_err(t.Token, "duplicated_ident", t.Ident)
	}

	s.check_trait_decl_methods(t)
}

// Checks current package file's trait declarations.
func (s *_Sema) check_trait_decls() (ok bool) {
	for _, t := range s.file.Traits {
		s.check_trait_decl(t)

		// Break checking if type alias has error.
		if len(s.errors) > 0 {
			return false
		}
	}
	return true
}

func (s *_Sema) check_trait_impl_methods(base *Trait, ipl *Impl) (ok bool) {
	ok = true
	for _, f := range ipl.Methods {
		if base.Find_method(f.Ident) == nil {
			s.push_err(f.Token, "trait_have_not_ident", base.Ident, f.Ident)
			ok = false
		}
	}
	return
}

func (s *_Sema) impl_to_struct(dest *Struct, ipl *Impl) (ok bool) {
	ok = true
	for _, f := range ipl.Methods {
		if dest.Find_method(f.Ident) != nil {
			s.push_err(f.Token, "struct_already_have_ident", dest.Ident, f.Ident)
			ok = false
			continue
		}

		dest.Methods = append(dest.Methods, f)
	}
	return
}

// Implement trait to destination.
func (s *_Sema) impl_trait(decl *Impl) {
	base := s.find_trait(decl.Base.Kind)
	if base == nil {
		s.push_err(decl.Base, "impl_base_not_exist", decl.Base.Kind)
		return
	}

	// Cpp-link state always false because cpp-linked
	// definitions haven't support implementations.
	const CPP_LINKED = false

	dest := s.find_struct(decl.Dest.Kind, CPP_LINKED)
	if dest == nil {
		s.push_err(decl.Dest, "impl_dest_not_exist", decl.Dest.Kind)
		return
	}

	dest.Implements = append(dest.Implements, base)

	switch  {
	case !s.check_trait_impl_methods(base, decl):
		return

	case !s.impl_to_struct(dest, decl):
		return
	}

	// TODO: Check structure implements trait correctly.
}

// Implement implementation.
func (s *_Sema) impl_impl(decl *Impl) {
	switch {
	case decl.Is_trait_impl():
		s.impl_trait(decl)

	case decl.Is_struct_impl():
		// TODO: Implement here.
	}
}

// Implement implementations.
func (s *_Sema) impl_impls() (ok bool) {
	for _, decl := range s.file.Impls {
		s.impl_impl(decl)

		// Break checking if type alias has error.
		if len(s.errors) > 0 {
			return false
		}
	}
	return true
}

func (s *_Sema) check_global_decl(decl *Var) {
	if lex.Is_ignore_ident(decl.Ident) {
		s.push_err(decl.Token, "ignore_ident")
	} else if s.is_duplicated_ident(_uintptr(decl), decl.Ident, false) {
		s.push_err(decl.Token, "duplicated_ident", decl.Ident)
	}

	if decl.Value == nil {
		s.push_err(decl.Token, "variable_not_initialized")
	}

	if decl.Is_auto_typed() {
		if decl.Value == nil {
			s.push_err(decl.Token, "missing_autotype_value")
		}
	} else {
		_ = s.check_type(decl.Kind)
	}
}

// Checks current package file's global variable declarations.
func (s *_Sema) check_global_decls() (ok bool) {
	for _, decl := range s.file.Vars {
		s.check_global_decl(decl)

		// Break checking if type alias has error.
		if len(s.errors) > 0 {
			return false
		}
	}
	return true
}

func (s *_Sema) check_struct_decl(strct *Struct) {
	if lex.Is_ignore_ident(strct.Ident) {
		s.push_err(strct.Token, "ignore_ident")
	} else if s.is_duplicated_ident(_uintptr(strct), strct.Ident, false) {
		s.push_err(strct.Token, "duplicated_ident", strct.Ident)
	}

	ok := s.check_decl_generics(strct.Generics)
	if !ok {
		return
	}

	// TODO: Check fields and methods if not have any generic type.
}

// Checks current package file's structure declarations.
func (s *_Sema) check_struct_decls() (ok bool) {
	for _, strct := range s.file.Structs {
		s.check_struct_decl(strct)
		
		// Break checking if type alias has error.
		if len(s.errors) > 0 {
			return false
		}
	}
	return true
}

func (s *_Sema) check_fn_decl(f *Fn) {
	if lex.Is_ignore_ident(f.Ident) {
		s.push_err(f.Token, "ignore_ident")
	} else if s.is_duplicated_ident(_uintptr(f), f.Ident, false) {
		s.push_err(f.Token, "duplicated_ident", f.Ident)
	}

	ok := s.check_fn_decl_prototype(f)
	if !ok {
		return
	}

	// TODO: Check scope if function not has any generic type.
}

// Checks current package file's function declarations.
func (s *_Sema) check_fn_decls() (ok bool) {
	for _, f := range s.file.Funcs {
		s.check_fn_decl(f)
		
		// Break checking if type alias has error.
		if len(s.errors) > 0 {
			return false
		}
	}
	return true
}

// Checks all declarations of current package file.
// Reports whether checking is success.
func (s *_Sema) check_file_decls() (ok bool) {
	switch {
	case !s.check_type_alias_decls():
		return false

	case !s.check_enum_decls():
		return false

	case !s.check_trait_decls():
		return false

	case !s.impl_impls():
		return false

	case !s.check_global_decls():
		return false

	case !s.check_fn_decls():
		return false

	case !s.check_struct_decls():
		return false

	default:
		return true
	}
}

// Checks declarations of all package files.
// Breaks checking if checked file failed.
func (s *_Sema) check_package_decls() {
	for _, f := range s.files {
		s.set_current_file(f)
		ok := s.check_file_decls()
		if !ok {
			return
		}
	}
}

func (s *_Sema) check_data_for_auto_type(d *Data, err_token lex.Token) {
	switch {
	case d.Is_nil():
		s.push_err(err_token, "nil_for_autotype")

	case d.Is_void():
		s.push_err(err_token, "void_for_autotype")
	}
}

func (s *_Sema) check_type_global(decl *Var) {
	data := s.evalp(decl.Value.Expr, decl.Kind)
	if data == nil {
		return // Skip checks if error ocurrs.
	}

	if decl.Is_auto_typed() {
		// Build new TypeSymbol because
		// auto-type symbols are nil.
		decl.Kind = &TypeSymbol{Kind: data.Kind}
		s.check_data_for_auto_type(data, decl.Value.Expr.Token)
		// TODO: Check assignment validity.
	} else {
		// TODO: Check type compatibility.
	}
}

// Checks types of current package file's global variables.
func (s *_Sema) check_global_types() (ok bool) {
	for _, decl := range s.file.Vars {
		s.check_type_global(decl)

		// Break checking if type alias has error.
		if len(s.errors) > 0 {
			return false
		}
	}
	return true
}

// Checks all types of current package file.
// Reports whether checking is success.
func (s *_Sema) check_file_types() (ok bool) {
	// TODO: Implement other declarations.
	switch {
	case !s.check_global_types():
		return false

	default:
		return true
	}
}

// Checks all types of all package files.
// Breaks checking if checked file failed.
func (s *_Sema) check_package_types() {
	for _, f := range s.files {
		s.set_current_file(f)
		ok := s.check_file_types()
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

	s.check_package_decls()
	// Break checking if imports has error.
	if len(s.errors) > 0 {
		return
	}

	s.check_package_types()
}
