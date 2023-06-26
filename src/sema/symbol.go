// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// Directive pass.
type Pass struct {
	Token lex.Token
	Text  string
}

func build_doc(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	doc := ""
	for _, c := range cg.Comments {
		doc += c.Text
		doc += " " // Write space for each newline.
	}
	return doc
}

func build_type(t *ast.TypeDecl) *TypeSymbol {
	if t == nil {
		return nil
	}
	return &TypeSymbol{
		Decl: t,
		Kind: nil,
	}
}

func build_expr(expr *ast.Expr) *Value {
	if expr == nil {
		return nil
	}

	return &Value{
		Data: nil,
		Expr: expr,
	}
}

func build_type_alias(decl *ast.TypeAliasDecl) *TypeAlias {
	return &TypeAlias{
		Scope:      decl.Scope,
		Public:     decl.Public,
		Cpp_linked: decl.Cpp_linked,
		Token:      decl.Token,
		Ident:      decl.Ident,
		Kind:       build_type(decl.Kind),
		Doc:        build_doc(decl.Doc_comments),
	}
}

func build_field(decl *ast.FieldDecl) *Field {
	return &Field{
		Token:   decl.Token,
		Public:  decl.Public,
		Mutable: decl.Mutable,
		Ident:   decl.Ident,
		Kind:    build_type(decl.Kind),
	}
}

func build_fields(decls []*ast.FieldDecl) []*Field {
	fields := make([]*Field, len(decls))
	for i, decl := range decls {
		fields[i] = build_field(decl)
	}
	return fields
}

func build_struct(decl *ast.StructDecl) *Struct {
	return &Struct{
		Token:      decl.Token,
		Ident:      decl.Ident,
		Fields:     build_fields(decl.Fields),
		Public:     decl.Public,
		Cpp_linked: decl.Cpp_linked,
		Directives: decl.Directives,
		Doc:        build_doc(decl.Doc_comments),
		Generics:   decl.Generics,
	}
}

func build_param(decl *ast.ParamDecl) *Param {
	return &Param{
		Token:    decl.Token,
		Mutable:  decl.Mutable,
		Variadic: decl.Variadic,
		Kind:     build_type(decl.Kind),
		Ident:    decl.Ident,
	}
}

func build_params(decls []*ast.ParamDecl) []*Param {
	params := make([]*Param, len(decls))
	for i, decl := range decls {
		params[i] = build_param(decl)
	}
	return params
}

func build_ret_type(decl *ast.RetTypeDecl) *RetType {
	if decl.Idents == nil && decl.Kind == nil {
		return nil // Void type.
	}
	return &RetType{
		Kind:   build_type(decl.Kind),
		Idents: decl.Idents,
	}
}

func build_fn(decl *ast.FnDecl) *Fn {
	return &Fn{
		Token:      decl.Token,
		Global:     decl.Global,
		Unsafety:   decl.Unsafety,
		Public:     decl.Public,
		Cpp_linked: decl.Cpp_linked,
		Ident:      decl.Ident,
		Directives: decl.Directives,
		Doc:        build_doc(decl.Doc_comments),
		Scope:      decl.Scope,
		Generics:   decl.Generics,
		Result:     build_ret_type(decl.Result),
		Params:     build_params(decl.Params),
	}
}

func build_methods(decls []*ast.FnDecl) []*Fn {
	methods := make([]*Fn, len(decls))
	for i, decl := range decls {
		methods[i] = build_fn(decl)
	}
	return methods
}

func build_trait(decl *ast.TraitDecl) *Trait {
	return &Trait{
		Token:   decl.Token,
		Ident:   decl.Ident,
		Public:  decl.Public,
		Doc:     build_doc(decl.Doc_comments),
		Methods: build_methods(decl.Methods),
	}
}

func build_enum_item(decl *ast.EnumItemDecl) *EnumItem {
	return &EnumItem{
		Token: decl.Token,
		Ident: decl.Ident,
		Value: build_expr(decl.Expr),
	}
}

func build_enum_items(decls []*ast.EnumItemDecl) []*EnumItem {
	items := make([]*EnumItem, len(decls))
	for i, decl := range decls {
		items[i] = build_enum_item(decl)
	}
	return items
}

func build_enum(decl *ast.EnumDecl) *Enum {
	return &Enum{
		Token:  decl.Token,
		Public: decl.Public,
		Ident:  decl.Ident,
		Kind:   build_type(decl.Kind),
		Items:  build_enum_items(decl.Items),
		Doc:    build_doc(decl.Doc_comments),
	}
}

func build_var(decl *ast.VarDecl) *Var {
	return &Var{
		Scope:      decl.Scope,
		Token:      decl.Token,
		Ident:      decl.Ident,
		Cpp_linked: decl.Cpp_linked,
		Constant:   decl.Constant,
		Mutable:    decl.Mutable,
		Public:     decl.Public,
		Doc:        build_doc(decl.Doc_comments),
		Kind:       build_type(decl.Kind),
		Value:      build_expr(decl.Expr),
	}
}

func build_impl(decl *ast.Impl) *Impl {
	return &Impl{
		Base:    decl.Base,
		Dest:    decl.Dest,
		Methods: build_methods(decl.Methods),
	}
}

// Symbol table builder.
// Just builds symbols, not analyze metadatas
// like struct's implemented traits.
type _SymbolBuilder struct {
	owner    *_SymbolBuilder
	importer Importer
	errors   []build.Log
	ast      *ast.Ast
	table    *SymbolTable
}

func (s *_SymbolBuilder) get_root() *_SymbolBuilder {
	root := s
	for root.owner != nil {
		root = root.owner
	}
	return root
}

func (s *_SymbolBuilder) push_err(token lex.Token, key string, args ...any) {
	s.errors = append(s.errors, build.Log{
		Type:   build.ERR,
		Row:    token.Row,
		Column: token.Column,
		Path:   token.File.Path(),
		Text:   build.Errorf(key, args...),
	})
}

func (s *_SymbolBuilder) check_cpp_use_decl_path(decl *ast.UseDecl) (ok bool) {
	if build.Is_std_header_path(decl.Link_path) {
		return true
	}

	ext := filepath.Ext(decl.Link_path)
	if !build.Is_valid_header_ext(ext) && !build.Is_valid_cpp_ext(ext) {
		s.push_err(decl.Token, "invalid_cpp_ext", ext)
		return false
	}

	info, err := os.Stat(decl.Link_path)
	// Exist?
	if err != nil || info.IsDir() {
		s.push_err(decl.Token, "use_not_found", decl.Link_path)
		return false
	}

	return true
}

func (s *_SymbolBuilder) build_cpp_header_import(decl *ast.UseDecl) *ImportInfo {
	path := decl.Link_path

	if !build.Is_std_header_path(decl.Link_path) {
		ok := s.check_cpp_use_decl_path(decl)
		if !ok {
			return nil
		}

		// Set to absolute path for correct include path.
		var err error
		path, err = filepath.Abs(decl.Link_path)
		if err != nil {
			s.push_err(decl.Token, "use_not_found", decl.Link_path)
		}
	}

	return &ImportInfo{
		Token:      decl.Token,
		Path:       path,
		Link_path:  decl.Link_path,
		Ident:      "", // Cpp headers haven't identifiers.
		Cpp_linked: true,
		Std:        false,
		Package:    nil, // Cpp headers haven't symbol table.
	}
}

func (s *_SymbolBuilder) build_std_import(decl *ast.UseDecl) *ImportInfo {
	path := decl.Link_path[len("std::"):] // Skip "std::" prefix.
	path = strings.Replace(path, lex.KND_DBLCOLON, string(filepath.Separator), -1)
	path = filepath.Join(build.PATH_STDLIB, path)
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

	return &ImportInfo{
		Import_all: decl.Full,
		Token:      decl.Token,
		Path:       path,
		Link_path:  decl.Link_path,
		Ident:      ident,
		Cpp_linked: false,
		Std:        true,
		Package: &Package{
			Files: nil, // Appends by import algorithm.
		},
	}
}

func (s *_SymbolBuilder) build_ident_import(decl *ast.UseDecl) *ImportInfo {
	path := decl.Link_path
	path = strings.Replace(path, lex.KND_DBLCOLON, string(filepath.Separator), -1)
	path = filepath.Join(s.get_root().ast.File.Dir(), path)

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

	return &ImportInfo{
		Import_all: decl.Full,
		Token:      decl.Token,
		Path:       path,
		Link_path:  decl.Link_path,
		Ident:      ident,
		Cpp_linked: false,
		Std:        false,
		Package: &Package{
			Files: nil, // Appends by import algorithm.
		},
	}
}

func (s *_SymbolBuilder) build_import(decl *ast.UseDecl) *ImportInfo {
	switch {
	case decl.Cpp_linked:
		return s.build_cpp_header_import(decl)

	case decl.Std:
		return s.build_std_import(decl)

	default:
		return s.build_ident_import(decl)
	}
}

func (s *_SymbolBuilder) check_duplicate_use_decl(pkg *ImportInfo) (ok bool) {
	// Find package by path to detect cpp header imports.
	lpkg := s.table.Select_package(func(spkg *ImportInfo) bool {
		return pkg.Path == spkg.Path
	})
	if lpkg == nil {
		return true
	}
	s.push_err(pkg.Token, "duplicate_use_decl", pkg.Link_path)
	return false
}

func (s *_SymbolBuilder) impl_import_selections(imp *ImportInfo, decl *ast.UseDecl) {
	for _, ident := range decl.Selected {
		if imp.exist_ident(ident.Kind) {
			s.push_err(ident, "duplicated_import_selection", ident.Kind)
			continue
		}

		imp.Selected = append(imp.Selected, ident)
	}
}

func (s *_SymbolBuilder) get_as_link_path(path string) string {
	if strings.HasPrefix(path, build.PATH_STDLIB) {
		path = path[len(build.PATH_STDLIB):]
		return "std" + strings.ReplaceAll(path, string(filepath.Separator), lex.KND_DBLCOLON)
	}

	root, _ := filepath.Abs(s.get_root().ast.File.Dir())
	path = path[len(root):]
	if path[0] == filepath.Separator {
		path = path[1:]
	}
	return strings.ReplaceAll(path, string(filepath.Separator), lex.KND_DBLCOLON)
}

func (s *_SymbolBuilder) push_cross_cycle_error(target *_SymbolBuilder, imp *ImportInfo, error_token lex.Token) {
	const PADDING = 4

	message := ""

	push := func(sb *_SymbolBuilder, path string) {
		refers_to := build.Errorf("refers_to", s.get_as_link_path(sb.table.File.Dir()), s.get_as_link_path(path))
		message = strings.Repeat(" ", PADDING) + refers_to + "\n" + message
	}

	push(s, imp.Path)

	owner := s.owner
	old := s

	for owner.owner != nil {
		push(old.owner, old.table.File.Dir())

		if owner.owner == target {
			push(target, owner.table.File.Dir())
			break
		}

		old = owner
		owner = owner.owner
	}

	s.push_err(error_token, "pkg_illegal_cross_cycle", message)
}

func (s *_SymbolBuilder) check_import_cycles(imp *ImportInfo, decl *ast.UseDecl) bool {
	if imp.Path == s.table.File.Dir() {
		s.push_err(decl.Token, "pkg_illegal_cycle_refers_itself", s.get_as_link_path(imp.Path))
		return false
	}

	if s.owner == nil {
		return true
	}

	if s.owner.table.File.Dir() == imp.Path {
		s.push_cross_cycle_error(s.owner, imp, decl.Token)
		return false
	}

	owner := s.owner
iter:
	if owner.table.File.Dir() == imp.Path {
		s.push_cross_cycle_error(owner, imp, decl.Token)
		return false
	}

	if owner.owner != nil {
		owner = owner.owner
		goto iter
	}

	return true
}

func (s *_SymbolBuilder) import_package(imp *ImportInfo, decl *ast.UseDecl) (ok bool) {
	if imp.Cpp_linked {
		return true
	}

	port := s.importer.Get_import(imp.Path)
	if port != nil {
		imp.Package = port.Package
		imp.Duplicate = true
	} else {
		if !s.check_import_cycles(imp, decl) {
			return false
		}

		asts, errors := s.importer.Import_package(imp.Path)
		if len(errors) > 0 {
			s.errors = append(s.errors, errors...)
			return false
		}

		for _, ast := range asts {
			table, errors := build_symbols(ast, s.importer, s)

			// Break import if file has error(s).
			if len(errors) > 0 {
				s.errors = append(s.errors, errors...)
				s.push_err(imp.Token, "used_package_has_errors", imp.Link_path)
				return false
			}

			imp.Package.Files = append(imp.Package.Files, table)
		}
	}

	s.impl_import_selections(imp, decl)

	return true
}

func (s *_SymbolBuilder) import_use_decl(decl *ast.UseDecl) *ImportInfo {
	imp := s.build_import(decl)
	// Break analysis if error occurs.
	if imp == nil {
		return nil
	}

	ok := s.check_duplicate_use_decl(imp)
	if !ok {
		return nil
	}

	ok = s.import_package(imp, decl)
	s.table.Imports = append(s.table.Imports, imp)
	if ok {
		s.importer.Imported(imp)
		return imp
	}
	return nil
}

func (s *_SymbolBuilder) import_use_decls() {
	for _, decl := range s.ast.Use_decls {
		s.import_use_decl(decl)

		// Break analysis if error occurs.
		if len(s.errors) > 0 {
			break
		}
	}
}

func (s *_SymbolBuilder) append_decl(decl ast.Node) {
	switch decl.Data.(type) {
	case *ast.TypeAliasDecl:
		ta := build_type_alias(decl.Data.(*ast.TypeAliasDecl))
		s.table.Type_aliases = append(s.table.Type_aliases, ta)

	case *ast.StructDecl:
		srct := build_struct(decl.Data.(*ast.StructDecl))
		s.table.Structs = append(s.table.Structs, srct)

	case *ast.FnDecl:
		f := build_fn(decl.Data.(*ast.FnDecl))
		s.table.Funcs = append(s.table.Funcs, f)

	case *ast.TraitDecl:
		t := build_trait(decl.Data.(*ast.TraitDecl))
		s.table.Traits = append(s.table.Traits, t)

	case *ast.VarDecl:
		v := build_var(decl.Data.(*ast.VarDecl))
		s.table.Vars = append(s.table.Vars, v)

	case *ast.EnumDecl:
		e := build_enum(decl.Data.(*ast.EnumDecl))
		s.table.Enums = append(s.table.Enums, e)

	default:
		s.push_err(decl.Token, "invalid_syntax")
	}
}

func (s *_SymbolBuilder) append_decls() {
	for _, decl := range s.ast.Decls {
		s.append_decl(decl)
	}
}

func (s *_SymbolBuilder) append_impls() {
	s.table.Impls = make([]*Impl, len(s.ast.Impls))
	for i, decl := range s.ast.Impls {
		s.table.Impls[i] = build_impl(decl)
	}
}

func (s *_SymbolBuilder) push_directive_pass(d *ast.Directive) {
	pass := Pass{
		Token: d.Token,
	}
	for _, arg := range d.Args {
		if arg != "" {
			pass.Text += arg + " "
		}
	}
	pass.Text = strings.TrimSpace(pass.Text)
	s.table.Passes = append(s.table.Passes, pass)
}

func (s *_SymbolBuilder) append_top_directives() {
	for _, d := range s.ast.Top_directives {
		switch d.Tag {
		case build.DIRECTIVE_PASS:
			s.push_directive_pass(d)
		}
	}
}

func (s *_SymbolBuilder) build() {
	s.table = &SymbolTable{
		File: s.ast.File,
	}

	s.append_top_directives()

	s.import_use_decls()
	// Break analysis if use declarations has error.
	if len(s.errors) > 0 {
		return
	}

	s.append_decls()
	// Break analysis if declarations has error.
	if len(s.errors) > 0 {
		return
	}

	s.append_impls()
}
