// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"unsafe"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/constant"
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

// TODO: Implement built-in definitions.

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

func (s *_Sema) check_generic_quantity(required int, given int, error_token lex.Token) (ok bool) {
	switch {
	case required == 0 && given > 0:
		s.push_err(error_token, "not_has_generics")
		return false

	case required > 0 && given == 0:
		s.push_err(error_token, "has_generics")
		return false

	case required < given:
		s.push_err(error_token, "generics_overflow")
		return false

	case required > given:
		s.push_err(error_token, "missing_generics")
		return false

	default:
		return true
	}
}

// Returns package by identifier.
// Returns nil if not exist any package in this identifier.
//
// Lookups:
//  - Current file's imported packages.
func (s *_Sema) Find_package(ident string) *Package {
	return s.file.Find_package(ident)
}

// Returns package by selector.
// Returns nil if selector returns false for all packages.
// Returns nil if selector is nil.
//
// Lookups:
//  - Current file's imported packages.
func (s *_Sema) Select_package(selector func(*Package) bool) *Package {
	return s.file.Select_package(selector)
}

// Returns variable by identifier and cpp linked state.
// Returns nil if not exist any variable in this identifier.
//
// Lookups:
//  - Package file's symbol table.
//  - Current file's public denifes of imported packages.
func (s *_Sema) Find_var(ident string, cpp_linked bool) *Var {
	// Lookup package files.
	v := find_var_in_package(s.files, ident, cpp_linked)
	if v != nil {
		return v
	}

	// Lookup current file's public denifes of imported packages.
	for _, pkg := range s.file.Packages {
		v := pkg.Find_var(ident, cpp_linked)
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
func (s *_Sema) Find_type_alias(ident string, cpp_linked bool) *TypeAlias {
	// Lookup package files.
	ta := find_type_alias_in_package(s.files, ident, cpp_linked)
	if ta != nil {
		return ta
	}

	// Lookup current file's public denifes of imported packages.
	for _, pkg := range s.file.Packages {
		ta := pkg.Find_type_alias(ident, cpp_linked)
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
func (s *_Sema) Find_struct(ident string, cpp_linked bool) *Struct {
	// Lookup package files.
	strct := find_struct_in_package(s.files, ident, cpp_linked)
	if strct != nil {
		return strct
	}

	// Lookup current file's public denifes of imported packages.
	for _, pkg := range s.file.Packages {
		strct := pkg.Find_struct(ident, cpp_linked)
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
func (s *_Sema) Find_fn(ident string, cpp_linked bool) *Fn {
	// Lookup package files.
	f := find_fn_in_package(s.files, ident, cpp_linked)
	if f != nil {
		return f
	}

	// Lookup current file's public denifes of imported packages.
	for _, pkg := range s.file.Packages {
		f := pkg.Find_fn(ident, cpp_linked)
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
func (s *_Sema) Find_trait(ident string) *Trait {
	// Lookup package files.
	t := find_trait_in_package(s.files, ident)
	if t != nil {
		return t
	}

	// Lookup current file's public denifes of imported packages.
	for _, pkg := range s.file.Packages {
		t := pkg.Find_trait(ident)
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
func (s *_Sema) Find_enum(ident string) *Enum {
	// Lookup package files.
	e := find_enum_in_package(s.files, ident)
	if e != nil {
		return e
	}

	// Lookup current file's public denifes of imported packages.
	for _, pkg := range s.file.Packages {
		e := pkg.Find_enum(ident)
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
func (s *_Sema) check_type_with_refers(t *TypeSymbol, l Lookup, referencer *_Referencer) (ok bool) {
	if t.checked() {
		return true
	}
	tc := _TypeChecker{
		s:          s,
		lookup:     l,
		referencer: referencer,
	}
	tc.check(t)
	return t.checked()
}

// Checks type and builds result as kind.
// Skips already checked types.
func (s *_Sema) check_type(t *TypeSymbol, l Lookup) (ok bool) {
	return s.check_type_with_refers(t, l, nil)
}

// Builds type with type aliases for generics.
// Returns nil if error occur or failed.
func (s *_Sema) build_type_with_generics(t *ast.Type, generics []*TypeAlias) *TypeKind {
	tc := _TypeChecker{
		s:            s,
		lookup:       s,
		use_generics: generics,
	}
	return tc.check_decl(t)
}

// Same as s.build_type_with_generics but not uses any generics.
func (s *_Sema) build_type(t *ast.Type) *TypeKind {
	return s.build_type_with_generics(t, nil)
}

// Evaluates expression with type prefixed Eval and returns result.
// Checks variable dependencies if exist.
func (s *_Sema) evalpd(expr *ast.Expr, l Lookup, p *TypeSymbol, owner *Var) *Data {
	e := _Eval{
		s:      s,
		lookup: l,
		owner:  owner,
	}

	switch l.(type) {
	case *_ScopeChecker:
		e.unsafety = l.(*_ScopeChecker).is_unsafe()
	}

	if p != nil {
		e.prefix = p.Kind
	}

	return e.eval(expr)
}

// Evaluates expression with type prefixed Eval and returns result.
func (s *_Sema) evalp(expr *ast.Expr, l Lookup, p *TypeSymbol) *Data {
	return s.evalpd(expr, l, p, nil)
}

// Evaluates expression with Eval and returns result.
func (s *_Sema) eval(expr *ast.Expr, l Lookup) *Data { return s.evalp(expr, l, nil) }

func (s *_Sema) check_assign_type(dest *TypeKind, d *Data, error_token lex.Token, deref bool) {
	atc := _AssignTypeChecker{
		s:           s,
		error_token: error_token,
		dest:        dest,
		d:           d,
		deref:       deref,
	}
	atc.check()
}

func (s *_Sema) check_type_compatibility(dest *TypeKind, src *TypeKind, error_token lex.Token, deref bool) {
	dest_kind := dest.To_str()
	if src == nil {
		s.push_err(error_token, "incompatible_types", dest_kind, "<untyped>")
		return
	}
	src_kind := src.To_str()

	// Tuple to single type, always fails.
	if src.Tup() != nil {
		s.push_err(error_token, "incompatible_types", dest_kind, src_kind)
		return
	}

	if dest.Prim() != nil && dest.Prim().Is_any() {
		return
	}

	tcc := _TypeCompatibilityChecker{
		s:           s,
		error_token: error_token,
		dest:        dest,
		src:         src,
		deref:       deref,
	}
	ok := tcc.check()

	switch {
	case ok:
		// Ok.

	case dest_kind == src_kind:
		// Ok.

	default:
		s.push_err(error_token, "incompatible_types", dest_kind, src_kind)
	}
}

// Builds non-generic types but skips generic types.
// Builds generic identifiers as primitive type.
//
// Useful:
//  - For non-generic type parsed string type kinds.
//  - For checking non-generic types.
func (s *_Sema) build_non_generic_type_kind(ast *ast.Type,
	generics []*ast.Generic, ignore_with_trait_pattern bool) *TypeKind {
	tc := _TypeChecker{
		s:                         s,
		lookup:                    s,
		ignore_generics:           generics,
		ignore_with_trait_pattern: ignore_with_trait_pattern,
	}
	return tc.check_decl(ast)
}

func (s *_Sema) build_fn_non_generic_type_kinds(f *FnIns, with_trait_pattern bool) {
	var generics []*ast.Generic
	if f.Decl.Is_method() {
		generics = append(f.Decl.Generics, f.Decl.Owner.Generics...)
	} else {
		generics = f.Decl.Generics
	}

	for _, p := range f.Params {
		if !p.Decl.Is_self() {
			p.Kind = s.build_non_generic_type_kind(p.Decl.Kind.Decl, generics, with_trait_pattern)
		}
	}
	if !f.Decl.Is_void() {
		f.Result = s.build_non_generic_type_kind(f.Decl.Result.Kind.Decl, generics, with_trait_pattern)
	}
}

func (s *_Sema) get_trait_check_fn_kind(f *Fn) string {
	ins := f.instance_force()
	s.build_fn_non_generic_type_kinds(ins, true)
	return to_trait_kind_str(ins)
}

func (s *_Sema) reload_fn_ins_types(f *FnIns) (ok bool) {
	generics := make([]*TypeAlias, len(f.Generics))
	for i, g := range f.Generics {
		generics[i] = &TypeAlias{
			Ident: f.Decl.Generics[i].Ident,
			Kind:  &TypeSymbol{
				Kind: g,
			},
		}
	}

	ok = true
	for _, p := range f.Params {
		if !p.Decl.Is_self() {
			p.Kind = s.build_type_with_generics(p.Decl.Kind.Decl, generics)
			ok = p.Kind != nil && ok
		}
	}

	if !f.Decl.Is_void() {
		f.Result = s.build_type_with_generics(f.Decl.Result.Kind.Decl, generics)
		ok = f.Result != nil && ok
	}

	return ok
}

func (s *_Sema) check_validity_for_init_expr(left_mut bool, d *Data, error_token lex.Token) {
	if d.Lvalue && left_mut && !d.Mutable && is_mut(d.Kind) {
		s.push_err(error_token, "assignment_non_mut_to_mut")
		return
	}
	atc := _AssignTypeChecker{
		s:           s,
		d:           d,
		error_token: error_token,
	}
	_ = atc.check_validity()
}

func (s *_Sema) check_type_alias_decl_kind(ta *TypeAlias, l Lookup) (ok bool) {
	ok = s.check_type_with_refers(ta.Kind, l, &_Referencer{
		ident:  ta.Ident,
		owner:  _uintptr(ta),
		refers: &ta.Refers,
	})
	if ok && ta.Kind.Kind.Arr() != nil && ta.Kind.Kind.Arr().Auto {
		s.push_err(ta.Kind.Decl.Token, "array_auto_sized")
		ok = false
	}
	return
}

func (s *_Sema) check_type_alias_decl(ta *TypeAlias, l Lookup) {
	if lex.Is_ignore_ident(ta.Ident) {
		s.push_err(ta.Token, "ignore_ident")
	}
	s.check_type_alias_decl_kind(ta, l)
}

// Checks type alias declaration with duplicated identifiers.
func (s *_Sema) check_type_alias_decl_dup(ta *TypeAlias) {
	if s.is_duplicated_ident(_uintptr(ta), ta.Ident, ta.Cpp_linked) {
		s.push_err(ta.Token, "duplicated_ident", ta.Ident)
	}
	s.check_type_alias_decl_kind(ta, s)
}

// Checks current package file's type alias declarations.
func (s *_Sema) check_type_alias_decls() (ok bool) {
	for _, ta := range s.file.Type_aliases {
		s.check_type_alias_decl_dup(ta)
		
		// Break checking if type alias has error.
		if len(s.errors) > 0 {
			return false
		}
	}
	return true
}

func (s *_Sema) check_enum_items_dup(e *Enum) {
	for _, item := range e.Items {
		if lex.Is_ignore_ident(item.Ident) {
			s.push_err(item.Token, "ignore_ident")
		} else {
			for _, citem := range e.Items {
				if item == citem {
					break
				} else if item.Ident == citem.Ident {
					s.push_err(item.Token, "duplicated_ident", item.Ident)
					break
				}
			}
		}
	}
}

func (s *_Sema) check_enum_items_str(e *Enum) {
	for _, item := range e.Items {
		if item.Auto_expr() {
			item.Value = &Value{
				Data: &Data{
					Constant: constant.New_str(item.Ident),
				},
			}
			item.Value.Data.Model = item.Value.Data.Constant
		} else {
			d := s.eval(item.Value.Expr, s)
			if d == nil {
				continue
			}

			if !d.Is_const() {
				s.push_err(item.Value.Expr.Token, "expr_not_const")
			}

			s.check_assign_type(e.Kind.Kind, d, item.Token, false)
			item.Value.Data = d
		}
	}
}

func (s *_Sema) check_enum_items_int(e *Enum) {
	max := uint64(types.Max_of(e.Kind.Kind.Prim().To_str()))
	for i, item := range e.Items {
		if max == 0 {
			s.push_err(item.Token, "overflow_limits")
		} else {
			max--
		}

		if item.Auto_expr() {
			item.Value = &Value{
				Data: &Data{
					Constant: constant.New_u64(max - (max - uint64(i))),
				},
			}
			item.Value.Data.Model = item.Value.Data.Constant
		} else {
			d := s.eval(item.Value.Expr, s)
			if d == nil {
				continue
			}

			if !d.Is_const() {
				s.push_err(item.Value.Expr.Token, "expr_not_const")
			}

			s.check_assign_type(e.Kind.Kind, d, item.Token, false)
			item.Value.Data = d
		}
	}
}

func (s *_Sema) check_enum_decl(e *Enum) {
	if lex.Is_ignore_ident(e.Ident) {
		s.push_err(e.Token, "ignore_ident")
	} else if s.is_duplicated_ident(_uintptr(e), e.Ident, false) {
		s.push_err(e.Token, "duplicated_ident", e.Ident)
	}

	s.check_enum_items_dup(e)

	if e.Kind != nil {
		if !s.check_type_with_refers(e.Kind, s, &_Referencer{
			ident:  e.Ident,
			owner: _uintptr(e),
			refers: &e.Refers,
		}) {
			return
		}
	} else {
		// Set to default type.
		e.Kind = &TypeSymbol{
			Decl: nil,
			Kind: &TypeKind{
				kind: &Prim{kind: types.TypeKind_I32},
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
		s.check_enum_items_str(e)

	case types.Is_int(t.To_str()):
		s.check_enum_items_int(e)

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

func (s *_Sema) check_fn_decl_types(f *Fn) (ok bool) {
	ok = true

	generics := f.Generics
	if f.Owner != nil {
		generics = append(generics, f.Owner.Generics...)
	}

	for _, p := range f.Params {
		if !p.Is_self() {
			kind := s.build_non_generic_type_kind(p.Kind.Decl, generics, false)
			ok = kind != nil && ok
			p.Kind.Kind = kind
		}
	}

	if !f.Is_void() {
		kind := s.build_non_generic_type_kind(f.Result.Kind.Decl, generics, false)
		ok = kind != nil && ok
		f.Result.Kind.Kind = kind
	}

	return ok
}

// Checks generics, parameters and return type.
// Not checks scope, and other things.
func (s *_Sema) check_fn_decl_prototype(f *Fn) (ok bool) {
	switch {
	case !s.check_decl_generics(f.Generics):
		return false

	case !s.check_fn_decl_params_dup(f):
		return false

	case !s.check_fn_decl_result_dup(f):
		return false

	case !s.check_fn_decl_types(f):
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

		if len(dest.Generics) > 0 && len(f.Generics) > 0 {
			for _, fg := range f.Generics {
				for _, dg := range dest.Generics {
					if fg.Ident == dg.Ident {
						s.push_err(fg.Token, "method_has_generic_with_same_ident")
						ok = false
					}
				}
			}
		}

		f.Owner = dest
		dest.Methods = append(dest.Methods, f)
	}
	return
}

// Implement trait to destination.
func (s *_Sema) impl_trait(decl *Impl) {
	base := s.Find_trait(decl.Base.Kind)
	if base == nil {
		s.push_err(decl.Base, "impl_base_not_exist", decl.Base.Kind)
		return
	}

	// Cpp-link state always false because cpp-linked
	// definitions haven't support implementations.
	const CPP_LINKED = false

	dest := s.Find_struct(decl.Dest.Kind, CPP_LINKED)
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
}

func (s *_Sema) impl_struct(decl *Impl) {
	// Cpp-link state always false because cpp-linked
	// definitions haven't support implementations.
	const CPP_LINKED = false

	dest := s.Find_struct(decl.Dest.Kind, CPP_LINKED)
	if dest == nil {
		s.push_err(decl.Base, "impl_dest_not_exist", decl.Base.Kind)
		return
	}

	switch  {
	case !s.impl_to_struct(dest, decl):
		return
	}
}

// Implement implementation.
func (s *_Sema) impl_impl(decl *Impl) {
	// TODO: Illegal implementation from out of package.
	switch {
	case decl.Is_trait_impl():
		s.impl_trait(decl)

	case decl.Is_struct_impl():
		s.impl_struct(decl)
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

// Checks variable declaration.
// No checks duplicated identifiers.
func (s *_Sema) check_var_decl(decl *Var, l Lookup) {
	if lex.Is_ignore_ident(decl.Ident) {
		s.push_err(decl.Token, "ignore_ident")
	}

	if decl.Value == nil || decl.Value.Expr == nil {
		s.push_err(decl.Token, "variable_not_initialized")
	}

	if decl.Is_auto_typed() {
		if decl.Value == nil {
			s.push_err(decl.Token, "missing_autotype_value")
		}
	} else {
		_ = s.check_type(decl.Kind, l)
	}
}

// Checks variable declaration.
// Checks duplicated identifiers by Sema.
func (s *_Sema) check_var_decl_dup(decl *Var) {
	if s.is_duplicated_ident(_uintptr(decl), decl.Ident, false) {
		s.push_err(decl.Token, "duplicated_ident", decl.Ident)
	}
	s.check_var_decl(decl, s)
}

// Checks current package file's global variable declarations.
func (s *_Sema) check_global_decls() (ok bool) {
	for _, decl := range s.file.Vars {
		s.check_var_decl_dup(decl)

		// Break checking if type alias has error.
		if len(s.errors) > 0 {
			return false
		}
	}
	return true
}

func (s *_Sema) check_struct_trait_impl(strct *Struct, trt *Trait) (ok bool) {
	for _, tf := range trt.Methods {
		exist := false
		sf := strct.Find_method(tf.Ident)
		if sf != nil {
			tf_k := s.get_trait_check_fn_kind(tf)
			sf_k := s.get_trait_check_fn_kind(sf)
			exist = tf_k == sf_k
		}
		if !exist {
			ins := tf.instance_force()
			s.build_fn_non_generic_type_kinds(ins, false)
			s.push_err(strct.Token, "not_impl_trait_def", trt.Ident, ins.To_str())
			ok = false
		}
	}
	return
}

func (s *_Sema) check_struct_impls(strct *Struct) (ok bool) {
	ok = true
	for _, trt := range strct.Implements {
		ok = s.check_struct_trait_impl(strct, trt) && ok
	}
	return ok
}

func (s *_Sema) check_struct_fields(st *Struct) (ok bool) {
	ok = true
	tc := _TypeChecker{
		s:               s,
		lookup:          s,
		ignore_generics: st.Generics,
		referencer:      &_Referencer{
			ident: st.Ident,
			strct: st,
		},
	}
	for _, f := range st.Fields {
		f.Kind.Kind = tc.check_decl(f.Kind.Decl)
		ok = f.Kind.Kind != nil && ok

		for _, cf := range st.Fields {
			if f == cf {
				break
			} else if f.Ident == cf.Ident {
				s.push_err(f.Token, "duplicated_ident", f.Ident)
				ok = false
			}
		}
	}
	return ok
}

func (s *_Sema) check_struct_decl(strct *Struct) {
	if lex.Is_ignore_ident(strct.Ident) {
		s.push_err(strct.Token, "ignore_ident")
	} else if s.is_duplicated_ident(_uintptr(strct), strct.Ident, false) {
		s.push_err(strct.Token, "duplicated_ident", strct.Ident)
	}

	strct.sema = s
	switch {
	case !s.check_decl_generics(strct.Generics):
		return
		
	case !s.check_struct_fields(strct):
		return

	case !s.check_struct_impls(strct):
		return
	}
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

func (s *_Sema) check_var(v *Var) {
	if !v.Constant {
		v.Value.Data.Constant = nil
	}

	if v.Is_auto_typed() {
		// Build new TypeSymbol because
		// auto-type symbols are nil.
		v.Kind = &TypeSymbol{Kind: v.Value.Data.Kind}

		s.check_data_for_auto_type(v.Value.Data, v.Value.Expr.Token)
		s.check_validity_for_init_expr(v.Mutable, v.Value.Data, v.Value.Expr.Token)
	} else {
		arr := v.Kind.Kind.Arr()
		if arr != nil {
			if arr.Auto {
				data_arr := v.Value.Data.Kind.Arr()
				if data_arr != nil {
					arr.N = data_arr.N
				}
			}
		}

		s.check_assign_type(v.Kind.Kind, v.Value.Data, v.Value.Expr.Token, false)
	}
}

func (s *_Sema) check_type_var(decl *Var, l Lookup) {
	decl.Value.Data = s.evalpd(decl.Value.Expr, l, decl.Kind, decl)
	if decl.Value.Data == nil {
		return // Skip checks if error ocurrs.
	}

	s.check_var(decl)
}

// Checks types of current package file's global variables.
func (s *_Sema) check_global_types() {
	for _, decl := range s.file.Vars {
		s.check_type_var(decl, s)
	}

	// Re-check depended.
	for _, decl := range s.file.Vars {
		if decl.Value.Data == nil && len(decl.Depends) > 0 {
			s.check_type_var(decl, s)
		}
	}
}

func (s *_Sema) check_fn_ins(f *FnIns) {
	sc := new_scope_checker(s)
	sc.check(f.Decl.Scope, f.Scope)
}

func (s *_Sema) check_type_fn(f *Fn) {
	if len(f.Instances) == 0 {
		if len(f.Generics) > 0 {
			return
		}

		ins := f.instance()
		f.Instances = append(f.Instances, ins)
		s.reload_fn_ins_types(ins)
	}

	for _, ins := range f.Instances {
		s.check_fn_ins(ins)
	}
}

// Checks types of current package file's functions.
func (s *_Sema) check_fn_types() (ok bool) {
	for _, decl := range s.file.Funcs {
		s.check_type_fn(decl)
	}
	return true
}

// Checks all types of current package file.
// Reports whether checking is success.
func (s *_Sema) check_file_types() {
	// TODO: Implement other declarations.
	s.check_global_types()
	s.check_fn_types()
}

// Checks all types of all package files.
// Breaks checking if checked file failed.
func (s *_Sema) check_package_types() {
	for _, f := range s.files {
		s.set_current_file(f)
		s.check_file_types()
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
