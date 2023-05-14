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

// Chain conditional node.
type If struct {
	Expr  ExprModel
	Scope *Scope
}

// Default scope of conditional chain.
type Else struct {
	Scope *Scope
}

// Conditional chain.
type Conditional struct {
	If      *If
	Elifs   []*If
	Default *Else
}

// Infinity iteration.
type InfIter struct {
	Scope *Scope
}

// While iteration.
type WhileIter struct {
	Expr  ExprModel
	Scope *Scope
}

// Scope checker.
type _ScopeChecker struct {
	s       *_Sema
	parent  *_ScopeChecker
	table   *SymbolTable
	scope   *Scope
	tree    *ast.ScopeTree
	is_iter bool
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
	sc.s.check_type_var(v, sc)
	
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

func (sc *_ScopeChecker) check_child_sc(tree *ast.ScopeTree, ssc *_ScopeChecker) *Scope {
	s := &Scope{
		Parent: sc.scope,
	}

	ssc.parent = sc
	ssc.check(tree, s)

	return s
}

func (sc *_ScopeChecker) check_child(tree *ast.ScopeTree) *Scope {
	ssc := new_scope_checker(sc.s)
	return sc.check_child_sc(tree, ssc)
}

func (sc *_ScopeChecker) check_anon_scope(tree *ast.ScopeTree) {
	s := sc.check_child(tree)
	sc.scope.Stmts = append(sc.scope.Stmts, s)
}

func (sc *_ScopeChecker) check_expr(expr *ast.Expr) {
	d := sc.s.eval(expr, sc)
	if d == nil {
		return
	}

	sc.scope.Stmts = append(sc.scope.Stmts, d)
}

func (sc *_ScopeChecker) check_if(i *ast.If) *If {
	s := sc.check_child(i.Scope)

	d := sc.s.eval(i.Expr, sc)
	if d == nil {
		return nil
	}

	prim := d.Kind.Prim()
	if prim == nil {
		sc.s.push_err(i.Expr.Token, "if_require_bool_expr")
		return nil
	}

	if !prim.Is_bool() {
		sc.s.push_err(i.Expr.Token, "if_require_bool_expr")
		return nil
	}

	return &If{
		Expr:  d.Model,
		Scope: s,
	}
}

func (sc *_ScopeChecker) check_else(e *ast.Else) *Else {
	s := sc.check_child(e.Scope)
	return &Else{
		Scope: s,
	}
}

func (sc *_ScopeChecker) check_conditional(conditional *ast.Conditional) {
	c := &Conditional{}

	c.If = sc.check_if(conditional.If)

	c.Elifs = make([]*If, len(conditional.Elifs))
	for i, elif := range conditional.Elifs {
		c.Elifs[i] = sc.check_if(elif)
	}

	c.Default = sc.check_else(conditional.Default)

	sc.scope.Stmts = append(sc.scope.Stmts, c)
}

func (sc *_ScopeChecker) check_iter_scope(tree *ast.ScopeTree) *Scope {
	ssc := new_scope_checker(sc.s)
	ssc.is_iter = true
	return sc.check_child_sc(tree, ssc)
}

func (sc *_ScopeChecker) check_inf_iter(it *ast.Iter) {
	kind := &InfIter{}

	kind.Scope = sc.check_iter_scope(it.Scope)

	sc.scope.Stmts = append(sc.scope.Stmts, kind)
}

func (sc *_ScopeChecker) check_while_iter(it *ast.Iter) {
	kind := &WhileIter{}

	kind.Scope = sc.check_iter_scope(it.Scope)

	wh := it.Kind.(*ast.WhileKind)
	d := sc.s.eval(wh.Expr, sc)
	if d == nil {
		return
	}

	prim := d.Kind.Prim()
	if prim == nil {
		sc.s.push_err(it.Token, "iter_while_require_bool_expr")
		return
	}

	if !prim.Is_bool() {
		sc.s.push_err(it.Token, "iter_while_require_bool_expr")
		return
	}

	kind.Expr = d.Model

	sc.scope.Stmts = append(sc.scope.Stmts, kind)
}

func (sc *_ScopeChecker) check_iter(it *ast.Iter) {
	if it.Is_inf() {
		sc.check_inf_iter(it)
		return
	}

	switch it.Kind.(type) {
	case *ast.WhileKind:
		sc.check_while_iter(it)

	default:
		println("error <unimplemented iteration kind>")
	}
}

func (sc *_ScopeChecker) check_node(node ast.NodeData) {
	switch node.(type) {
	case *ast.Comment:
		// Ignore.
		break

	case *ast.ScopeTree:
		sc.check_anon_scope(node.(*ast.ScopeTree))

	case *ast.VarDecl:
		sc.check_var_decl(node.(*ast.VarDecl))

	case *ast.TypeAliasDecl:
		sc.check_type_alias_decl(node.(*ast.TypeAliasDecl))

	case *ast.Expr:
		sc.check_expr(node.(*ast.Expr))

	case *ast.Conditional:
		sc.check_conditional(node.(*ast.Conditional))

	case *ast.Iter:
		sc.check_iter(node.(*ast.Iter))

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
