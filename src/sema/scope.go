package sema

import "github.com/julelang/jule/ast"

// Statement type.
type St = any

// Scope.
type Scope struct {
	Parent   *Scope
	Unsafety bool
	Deferred bool
	Stmts    []St
}

// Scope checker.
type _ScopeChecker struct {
	s      *_Sema
	parent *_ScopeChecker
	table  *SymbolTable
	scope  *Scope
	tree   *ast.ScopeTree
}

// Returns package by identifier.
// Returns nil if not exist any package in this identifier.
//
// Lookups:
//  - Sema.
func (sc *_ScopeChecker) Find_package(ident string) *Package {
	return sc.s.Find_package(ident)
}

// Returns package by selector.
// Returns nil if selector returns false for all packages.
// Returns nil if selector is nil.
//
// Lookups:
//  - Sema.
func (sc *_ScopeChecker) Select_package(selector func(*Package) bool) *Package {
	return sc.s.Select_package(selector)
}

// Returns variable by identifier and cpp linked state.
// Returns nil if not exist any variable in this identifier.
//
// Lookups:
//  - Current scope.
//  - Parent scopes.
//  - Sema.
func (sc *_ScopeChecker) Find_var(ident string, cpp_linked bool) *Var {
	v := sc.table.Find_var(ident, cpp_linked)
	if v != nil {
		return v
	}

	parent := sc.parent
	for parent != nil {
		v := parent.table.Find_var(ident, cpp_linked)
		if v != nil {
			return v
		}
		parent = parent.parent
	}

	return sc.s.Find_var(ident, cpp_linked)
}

// Returns type alias by identifier and cpp linked state.
// Returns nil if not exist any type alias in this identifier.
//
// Lookups:
//  - Current scope.
//  - Parent scopes.
//  - Sema.
func (sc *_ScopeChecker) Find_type_alias(ident string, cpp_linked bool) *TypeAlias {
	ta := sc.table.Find_type_alias(ident, cpp_linked)
	if ta != nil {
		return ta
	}

	parent := sc.parent
	for parent != nil {
		ta := parent.table.Find_type_alias(ident, cpp_linked)
		if ta != nil {
			return ta
		}
		parent = parent.parent
	}

	return sc.s.Find_type_alias(ident, cpp_linked)
}

// Returns struct by identifier and cpp linked state.
// Returns nil if not exist any struct in this identifier.
//
// Lookups:
//  - Sema.
func (sc *_ScopeChecker) Find_struct(ident string, cpp_linked bool) *Struct {
	return sc.s.Find_struct(ident, cpp_linked)
}

// Returns function by identifier and cpp linked state.
// Returns nil if not exist any function in this identifier.
//
// Lookups:
//  - Sema.
func (sc *_ScopeChecker) Find_fn(ident string, cpp_linked bool) *Fn {
	return sc.s.Find_fn(ident, cpp_linked)
}

// Returns trait by identifier.
// Returns nil if not exist any trait in this identifier.
//
// Lookups:
//  - Sema.
func (sc *_ScopeChecker) Find_trait(ident string) *Trait {
	return sc.s.Find_trait(ident)
}

// Returns enum by identifier.
// Returns nil if not exist any enum in this identifier.
//
// Lookups:
//  - Sema.
func (sc *_ScopeChecker) Find_enum(ident string) *Enum {
	return sc.s.Find_enum(ident)
}

// Reports this identifier duplicated in scope.
// The "self" parameter represents address of exception identifier.
// If founded identifier address equals to self, will be skipped.
func (sc *_ScopeChecker) is_duplicated_ident(self uintptr, ident string) bool {
	v := sc.Find_var(ident, false)
	if v != nil && _uintptr(v) != self && v.Scope == sc.tree {
		return true
	}

	ta := sc.Find_type_alias(ident, false)
	if ta != nil && _uintptr(ta) != self && ta.Scope == sc.tree {
		return true
	}

	return false
}

func (sc *_ScopeChecker) check_var_decl(decl *ast.VarDecl) {
	v := build_var(decl)
	if sc.is_duplicated_ident(_uintptr(v), v.Ident) {
		sc.s.push_err(v.Token, "duplicated_ident", v.Ident)
	}
	sc.s.check_var_decl(v, sc)
	sc.s.check_type_var(v)
	
	sc.table.Vars = append(sc.table.Vars, v)
	sc.scope.Stmts = append(sc.scope.Stmts, v)
}

func (sc *_ScopeChecker) check_type_alias_decl(decl *ast.TypeAliasDecl) {
	ta := build_type_alias(decl)
	if sc.is_duplicated_ident(_uintptr(ta), ta.Ident) {
		sc.s.push_err(ta.Token, "duplicated_ident", ta.Ident)
	}
	sc.s.check_type_alias_decl(ta, sc)

	sc.table.Type_aliases = append(sc.table.Type_aliases, ta)
	sc.scope.Stmts = append(sc.scope.Stmts, ta)
}

func (sc *_ScopeChecker) check_sub_scope(tree *ast.ScopeTree) {
	s := &Scope{
		Parent: sc.scope,
	}

	ssc := new_scope_checker(sc.s)
	ssc.parent = sc
	ssc.check(tree, s)

	sc.scope.Stmts = append(sc.scope.Stmts, s)
}

func (sc *_ScopeChecker) check_node(node ast.NodeData) {
	switch node.(type) {
	case *ast.ScopeTree:
		sc.check_sub_scope(node.(*ast.ScopeTree))

	case *ast.VarDecl:
		sc.check_var_decl(node.(*ast.VarDecl))

	case *ast.TypeAliasDecl:
		sc.check_type_alias_decl(node.(*ast.TypeAliasDecl))

	case *ast.Comment:
		// Skip.
		break

	default:
		println("error <unimplemented scope node>")
	}
}

func (sc *_ScopeChecker) check_tree() {
	for _, node := range sc.tree.Stmts {
		sc.check_node(node)
	}
}

// Checks scope tree.
func (sc *_ScopeChecker) check(tree *ast.ScopeTree, s *Scope) {
	s.Deferred = tree.Deferred
	s.Unsafety = tree.Unsafety

	sc.tree = tree
	sc.scope = s

	sc.check_tree()
}

func new_scope_checker(s *_Sema) *_ScopeChecker {
	return &_ScopeChecker{
		s:     s,
		table: &SymbolTable{},
	}
}
