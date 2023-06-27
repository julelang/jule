package sema

import (
	"unsafe"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/constant"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

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
	Elifs   []*If // First not is root condition.
	Default *Else
}

// Infinity iteration.
type InfIter struct {
	Scope *Scope
}

// While iteration.
type WhileIter struct {
	Expr  ExprModel // Can be nil if iteration is while-next kind.
	Next  St        // Nil if iteration is not while-next kind.
	Scope *Scope
}

// Reports whether iteration is while-next kind.
func (wi *WhileIter) Is_while_next() bool { return wi.Next != nil }

// Range iteration.
type RangeIter struct {
	Expr  *Data
	Scope *Scope
	Key_a *Var
	Key_b *Var
}

// Continue statement.
type ContSt struct {
	It uintptr
}

// Break statement.
type BreakSt struct {
	It   uintptr
	Mtch uintptr
}

// Label.
type Label struct {
	Ident string
}

// Goto statement.
type GotoSt struct {
	Ident string
}

// Postfix assignment.
type Postfix struct {
	Expr ExprModel
	Op   string
}

// Assigment.
type Assign struct {
	L  ExprModel
	R  ExprModel
	Op string
}

// Mult-declarative assignment.
type MultiAssign struct {
	L []ExprModel // Nil models represents ingored expressions.
	R ExprModel
}

// Match-Case.
type Match struct {
	Expr       ExprModel
	Type_match bool
	Cases      []*Case
	Default    *Case
}

// Match-Case case.
type Case struct {
	Owner *Match
	Scope *Scope
	Exprs []ExprModel
	Next  *Case
}

// Reports whether case is default.
func (c *Case) Is_default() bool { return c.Exprs == nil }

// Fall statement.
type FallSt struct {
	Dest_case uintptr
}

// Return statement.
type RetSt struct {
	Vars []*Var // Used "_" identifier to pass ignored vars for ordering.
	Expr ExprModel
}

// Built-in recover function call statement.
type Recover struct {
	Handler      *FnIns
	Handler_expr ExprModel
	Scope        *Scope
}

type _ScopeLabel struct {
	token lex.Token
	label *Label
	pos   int
	scope *_ScopeChecker
	used  bool
}

type _ScopeGoto struct {
	gt    *ast.GotoSt
	scope *_ScopeChecker
	pos   int
}

// Scope checker.
type _ScopeChecker struct {
	s           *_Sema
	owner       *FnIns
	parent      *_ScopeChecker
	child_index int // Index of child scope.
	table       *SymbolTable
	scope       *Scope
	tree        *ast.ScopeTree
	it          uintptr
	cse         uintptr
	labels      *[]*_ScopeLabel // All labels of all scopes.
	gotos       *[]*_ScopeGoto  // All gotos of all scopes.
	i           int
}

// Reports whether scope is unsafe.
func (sc *_ScopeChecker) is_unsafe() bool {
	scope := sc

iter:
	if scope.scope.Unsafety {
		return true
	}

	if scope.parent != nil {
		scope = scope.parent
		goto iter
	}

	return false
}

// Reports scope is root.
// Accepts anonymous functions as root.
func (sc *_ScopeChecker) is_root() bool { return sc.parent == nil || sc.owner != nil }

// Returns root scope.
// Accepts anonymous functions as root.
func (sc *_ScopeChecker) get_root() *_ScopeChecker {
	root := sc
	for root.parent != nil && root.owner == nil {
		root = root.parent
	}
	return root
}

// Returns imported package by identifier.
// Returns nil if not exist any package in this identifier.
//
// Lookups:
//   - Sema.
func (sc *_ScopeChecker) Find_package(ident string) *ImportInfo {
	return sc.s.Find_package(ident)
}

// Returns imported package by selector.
// Returns nil if selector returns false for all packages.
// Returns nil if selector is nil.
//
// Lookups:
//   - Sema.
func (sc *_ScopeChecker) Select_package(selector func(*ImportInfo) bool) *ImportInfo {
	return sc.s.Select_package(selector)
}

// Returns variable by identifier and cpp linked state.
// Returns nil if not exist any variable in this identifier.
//
// Lookups:
//   - Current scope.
//   - Parent scopes.
//   - Sema.
func (sc *_ScopeChecker) Find_var(ident string, cpp_linked bool) *Var {
	// Search reverse for correct shadowing.
	const REVERSE = true
	v := sc.table.__find_var(ident, cpp_linked, REVERSE)
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
//   - Current scope.
//   - Parent scopes.
//   - Sema.
func (sc *_ScopeChecker) Find_type_alias(ident string, cpp_linked bool) *TypeAlias {
	// Search reverse for correct shadowing.
	const REVERSE = true
	ta := sc.table.__find_type_alias(ident, cpp_linked, REVERSE)
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
//   - Sema.
func (sc *_ScopeChecker) Find_struct(ident string, cpp_linked bool) *Struct {
	return sc.s.Find_struct(ident, cpp_linked)
}

// Returns function by identifier and cpp linked state.
// Returns nil if not exist any function in this identifier.
//
// Lookups:
//   - Sema.
func (sc *_ScopeChecker) Find_fn(ident string, cpp_linked bool) *Fn {
	return sc.s.Find_fn(ident, cpp_linked)
}

// Returns trait by identifier.
// Returns nil if not exist any trait in this identifier.
//
// Lookups:
//   - Sema.
func (sc *_ScopeChecker) Find_trait(ident string) *Trait {
	return sc.s.Find_trait(ident)
}

// Returns enum by identifier.
// Returns nil if not exist any enum in this identifier.
//
// Lookups:
//   - Sema.
func (sc *_ScopeChecker) Find_enum(ident string) *Enum {
	return sc.s.Find_enum(ident)
}

// Returns label by identifier.
// Returns nil if not exist any label in this identifier.
// Just lookups current scope.
func (sc *_ScopeChecker) find_label(ident string) *Label {
	for _, st := range sc.scope.Stmts {
		switch st.(type) {
		case *Label:
			label := st.(*Label)
			if label.Ident == ident {
				return label
			}
		}
	}
	return nil
}

// Returns label by identifier.
// Returns nil if not exist any label in this identifier.
// Just lookups current scope.
func (sc *_ScopeChecker) find_label_scope(ident string) *_ScopeLabel {
	label := sc.find_label_all(ident)
	if label != nil && label.scope == sc {
		return label
	}

	return nil
}

// Returns label by identifier.
// Returns nil if not exist any label in this identifier.
// Lookups all labels.
func (sc *_ScopeChecker) find_label_all(ident string) *_ScopeLabel {
	for _, lbl := range *sc.labels {
		if lbl.label.Ident == ident {
			return lbl
		}
	}
	return nil
}

// Reports this identifier duplicated in scope.
// The "self" parameter represents address of exception identifier.
// If founded identifier address equals to self, will be skipped.
func (sc *_ScopeChecker) is_duplicated_ident(itself uintptr, ident string) bool {
	v := sc.Find_var(ident, false)
	if v != nil && _uintptr(v) != itself && v.Scope == sc.tree {
		return true
	}

	ta := sc.Find_type_alias(ident, false)
	if ta != nil && _uintptr(ta) != itself && ta.Scope == sc.tree {
		return true
	}

	return false
}

func (sc *_ScopeChecker) check_var_decl(decl *ast.VarDecl) {
	v := build_var(decl)

	defer func() {
		sc.table.Vars = append(sc.table.Vars, v)
		sc.scope.Stmts = append(sc.scope.Stmts, v)
	}()

	if sc.is_duplicated_ident(_uintptr(v), v.Ident) {
		sc.s.push_err(v.Token, "duplicated_ident", v.Ident)
	}

	sc.s.check_var_decl(v, sc)
	if !v.Is_auto_typed() && (v.Kind == nil || v.Kind.Kind == nil) {
		return
	}

	sc.s.check_type_var(v, sc)
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

func (sc *_ScopeChecker) get_child() *Scope {
	return &Scope{
		Parent: sc.scope,
	}
}

func (sc *_ScopeChecker) check_child_ssc(tree *ast.ScopeTree, s *Scope, ssc *_ScopeChecker) {
	ssc.parent = sc
	ssc.check(tree, s)
}

func (sc *_ScopeChecker) check_child_sc(tree *ast.ScopeTree, ssc *_ScopeChecker) *Scope {
	s := sc.get_child()
	sc.check_child_ssc(tree, s, ssc)
	return s
}

func (sc *_ScopeChecker) check_child(tree *ast.ScopeTree) *Scope {
	ssc := sc.new_child_checker()
	return sc.check_child_sc(tree, ssc)
}

func (sc *_ScopeChecker) check_anon_scope(tree *ast.ScopeTree) {
	s := sc.check_child(tree)
	sc.scope.Stmts = append(sc.scope.Stmts, s)
}

func (sc *_ScopeChecker) try_call_recover(d *Data) bool {
	switch d.Model.(type) {
	case *Recover:
		// Ok.

	default:
		return false
	}

	rec := d.Model.(*Recover)
	rec.Handler = d.Kind.Fnc() // Argument function.
	rec.Scope = &Scope{}

	sc.scope.Stmts = append(sc.scope.Stmts, rec)

	sc.tree.Stmts = sc.tree.Stmts[sc.i+1:]
	sc.scope = rec.Scope
	sc.check_tree()
	return true
}

func (sc *_ScopeChecker) check_expr(expr *ast.Expr) {
	d := sc.s.eval(expr, sc)
	if d == nil {
		return
	}

	if expr.Is_fn_call() {
		ok := sc.try_call_recover(d)
		if ok {
			return
		}
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
	sc.scope.Stmts = append(sc.scope.Stmts, c)

	c.Elifs = make([]*If, len(conditional.Tail)+1)

	c.Elifs[0] = sc.check_if(conditional.Head)
	for i, elif := range conditional.Tail {
		c.Elifs[i+1] = sc.check_if(elif)
	}

	if conditional.Default != nil {
		c.Default = sc.check_else(conditional.Default)
	}
}

func (sc *_ScopeChecker) check_iter_scope_sc(it uintptr, tree *ast.ScopeTree, ssc *_ScopeChecker) *Scope {
	ssc.it = it
	return sc.check_child_sc(tree, ssc)
}

func (sc *_ScopeChecker) check_iter_scope(it uintptr, tree *ast.ScopeTree) *Scope {
	ssc := sc.new_child_checker()
	return sc.check_iter_scope_sc(it, tree, ssc)
}

func (sc *_ScopeChecker) check_inf_iter(it *ast.Iter) {
	kind := &InfIter{}

	sc.scope.Stmts = append(sc.scope.Stmts, kind)

	kind.Scope = sc.check_iter_scope(_uintptr(kind), it.Scope)
}

func (sc *_ScopeChecker) is_valid_ast_st_for_next_st(n ast.NodeData) bool {
	switch n.(type) {
	case *ast.AssignSt:
		return !n.(*ast.AssignSt).Declarative

	case *ast.FnCallExpr, *ast.Expr:
		return true

	default:
		return false
	}
}

func (sc *_ScopeChecker) is_valid_st_for_next_st(st St) bool {
	switch st.(type) {
	case *FnCallExprModel,
		*Postfix,
		*Assign,
		*MultiAssign:
		return true

	case *Data:
		switch st.(*Data).Model.(type) {
		case *FnCallExprModel:
			return true

		default:
			return false
		}

	default:
		return false
	}
}

func (sc *_ScopeChecker) check_while_iter(it *ast.Iter) {
	wh := it.Kind.(*ast.WhileKind)
	if wh.Expr == nil && wh.Next == nil {
		sc.check_inf_iter(it)
		return
	}

	kind := &WhileIter{}

	sc.scope.Stmts = append(sc.scope.Stmts, kind)

	kind.Scope = sc.check_iter_scope(_uintptr(kind), it.Scope)

	if wh.Expr != nil {
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
	}

	if wh.Is_while_next() {
		if !sc.is_valid_ast_st_for_next_st(wh.Next) {
			sc.s.push_err(wh.Next_token, "invalid_stmt_for_next")
			return
		}

		n := len(sc.scope.Stmts)
		sc.check_node(wh.Next)
		if n < len(sc.scope.Stmts) {
			st := sc.scope.Stmts[n]
			sc.scope.Stmts = sc.scope.Stmts[:n] // Remove statement.
			if !sc.is_valid_st_for_next_st(st) {
				sc.s.push_err(wh.Next_token, "invalid_stmt_for_next")
			}

			kind.Next = st
		}
	}
}

func (sc *_ScopeChecker) check_range_iter(it *ast.Iter) {
	rang := it.Kind.(*ast.RangeKind)

	d := sc.s.eval(rang.Expr, sc)
	if d == nil {
		return
	}

	kind := &RangeIter{
		Expr: d,
	}

	rc := _RangeChecker{
		sc:   sc,
		kind: kind,
		rang: rang,
		d:    d,
	}
	ok := rc.check()
	if !ok {
		return
	}

	sc.scope.Stmts = append(sc.scope.Stmts, kind)

	ssc := sc.new_child_checker()

	if kind.Key_a != nil {
		ssc.table.Vars = append(ssc.table.Vars, kind.Key_a)
	}

	if kind.Key_b != nil {
		ssc.table.Vars = append(ssc.table.Vars, kind.Key_b)
	}

	kind.Scope = sc.check_iter_scope_sc(_uintptr(kind), it.Scope, ssc)
}

func (sc *_ScopeChecker) check_iter(it *ast.Iter) {
	if it.Is_inf() {
		sc.check_inf_iter(it)
		return
	}

	switch it.Kind.(type) {
	case *ast.WhileKind:
		sc.check_while_iter(it)

	case *ast.RangeKind:
		sc.check_range_iter(it)

	default:
		println("error <unimplemented iteration kind>")
	}
}

func (sc *_ScopeChecker) check_valid_cont_label(it uintptr) bool {
	scope := sc

iter:
	if scope.it == it {
		return true
	}

	if scope.parent != nil {
		scope = scope.parent
		goto iter
	}

	return false
}

func (sc *_ScopeChecker) check_valid_break_label(ptr uintptr) bool {
	scope := sc

iter:
	if scope.it == ptr {
		return true
	}

	if scope.cse != 0 {
		mtch := _uintptr((*Case)(unsafe.Pointer(scope.cse)).Owner)
		if mtch == ptr {
			return true
		}
	}

	if scope.parent != nil {
		scope = scope.parent
		goto iter
	}

	return false
}

func (sc *_ScopeChecker) check_cont_valid_scope(c *ast.ContSt) *ContSt {
	if c.Label.Id != lex.ID_NA {
		return &ContSt{}
	}

	scope := sc
iter:
	switch {
	case scope.it == 0 && scope.parent != nil && scope.owner == nil:
		scope = scope.parent
		goto iter

	case scope.it != 0:
		return &ContSt{It: scope.it}
	}

	sc.s.push_err(c.Token, "continue_at_out_of_valid_scope")
	return nil
}

func (sc *_ScopeChecker) check_cont(c *ast.ContSt) {
	cont := sc.check_cont_valid_scope(c)
	if cont == nil {
		return
	}

	if c.Label.Id != lex.ID_NA { // Label given.
		label := find_label_parent(c.Label.Kind, sc.parent)
		if label == nil {
			sc.s.push_err(c.Label, "label_not_exist", c.Label.Kind)
			return
		}

		label.used = true

		if label.pos+1 >= len(label.scope.scope.Stmts) {
			sc.s.push_err(c.Label, "invalid_label")
			return
		}

		i := label.pos + 1
		if i >= len(label.scope.scope.Stmts) {
			sc.s.push_err(c.Label, "invalid_label")
		} else {
			st := label.scope.scope.Stmts[i]
			switch st.(type) {
			case *InfIter:
				cont.It = _uintptr(st.(*InfIter))

			case *RangeIter:
				cont.It = _uintptr(st.(*RangeIter))

			case *WhileIter:
				cont.It = _uintptr(st.(*WhileIter))

			default:
				sc.s.push_err(c.Label, "invalid_label")
			}
		}
	}

	if cont.It != 0 {
		if !sc.check_valid_cont_label(cont.It) {
			sc.s.push_err(c.Label, "invalid_label")
		}
	}

	sc.scope.Stmts = append(sc.scope.Stmts, cont)
}

func (sc *_ScopeChecker) check_label(l *ast.LabelSt) {
	if sc.find_label(l.Ident) != nil {
		sc.s.push_err(l.Token, "label_exist", l.Ident)
		return
	}

	label := &Label{
		Ident: l.Ident,
	}

	sc.scope.Stmts = append(sc.scope.Stmts, label)
	*sc.labels = append(*sc.labels, &_ScopeLabel{
		token: l.Token,
		label: label,
		pos:   len(sc.scope.Stmts) - 1,
		scope: sc,
	})
}

func (sc *_ScopeChecker) push_goto(gt *ast.GotoSt) {
	sc.scope.Stmts = append(sc.scope.Stmts, &GotoSt{
		Ident: gt.Label.Kind,
	})

	*sc.gotos = append(*sc.gotos, &_ScopeGoto{
		gt:    gt,
		pos:   len(sc.scope.Stmts) - 1,
		scope: sc,
	})
}

func (sc *_ScopeChecker) check_postfix(a *ast.AssignSt) {
	if len(a.L) > 1 {
		sc.s.push_err(a.Setter, "invalid_syntax")
		return
	}

	d := sc.s.eval(a.L[0].Expr, sc)
	if d == nil {
		return
	}

	_ = check_assign(sc.s, d, nil, a.Setter)

	if d.Kind.Ptr() != nil {
		ptr := d.Kind.Ptr()
		if !ptr.Is_unsafe() && !sc.is_unsafe() {
			sc.s.push_err(a.L[0].Expr.Token, "unsafe_behavior_at_out_of_unsafe_scope")
			return
		}
	}

	check_t := d.Kind
	if d.Kind.Ref() != nil {
		check_t = d.Kind.Ref().Elem
	}

	if check_t.Prim() == nil || !types.Is_num(check_t.Prim().kind) {
		sc.s.push_err(a.Setter, "operator_not_for_juletype", a.Setter.Kind, d.Kind.To_str())
		return
	}

	sc.scope.Stmts = append(sc.scope.Stmts, &Postfix{
		Expr: d.Model,
		Op:   a.Setter.Kind,
	})
}

func (sc *_ScopeChecker) is_new_assign_ident(ident string) bool {
	if lex.Is_ignore_ident(ident) || ident == "" {
		return false
	}

	return sc.table.def_by_ident(ident, false) == nil
}

func (sc *_ScopeChecker) check_single_assign(a *ast.AssignSt) {
	r := sc.s.eval(a.R, sc)
	if r == nil {
		return
	}

	if lex.Is_ignore_ident(a.L[0].Ident) {
		if r.Kind.Is_void() {
			sc.s.push_err(a.R.Token, "invalid_expr")
		}

		sc.scope.Stmts = append(sc.scope.Stmts, r)
		return
	}

	l := sc.s.eval(a.L[0].Expr, sc)
	if l == nil {
		return
	}

	if !check_assign(sc.s, l, r, a.Setter) {
		return
	}

	if r.Kind.Tup() != nil {
		sc.s.push_err(a.Setter, "missing_multi_assign_idents")
		return
	}

	sc.scope.Stmts = append(sc.scope.Stmts, &Assign{
		L:  l.Model,
		R:  r.Model,
		Op: a.Setter.Kind,
	})

	if a.Setter.Kind != lex.KND_EQ && !r.Is_const() {
		a.Setter.Kind = a.Setter.Kind[:len(a.Setter.Kind)-1]

		solver := _BinopSolver{
			e: &_Eval{
				s:        sc.s,
				lookup:   sc,
				unsafety: sc.is_unsafe(),
			},
			op: a.Setter,
		}

		r = solver.solve_explicit(l, r)
		if r == nil {
			return
		}
		a.Setter.Kind += lex.KND_EQ
	}

	checker := _AssignTypeChecker{
		s:           sc.s,
		dest:        l.Kind,
		d:           r,
		error_token: a.Setter,
		deref:       true,
	}
	checker.check()
}

func (sc *_ScopeChecker) check_multi_assign(a *ast.AssignSt) {
	rd := sc.s.eval(a.R, sc)
	if rd == nil {
		return
	}

	r := get_datas_from_tuple_data(rd)

	switch {
	case len(a.L) > len(r):
		sc.s.push_err(a.Setter, "overflow_multi_assign_idents")
		return

	case len(a.L) < len(r):
		sc.s.push_err(a.Setter, "missing_multi_assign_idents")
		return
	}

	st := &MultiAssign{
		R: rd.Model,
	}

	if rd.Kind.Tup() == nil {
		st.R = &TupleExprModel{Datas: r}
	}

	for i := range a.L {
		lexpr := a.L[i]
		r := r[i]

		if lex.Is_ignore_ident(lexpr.Ident) {
			if r.Kind.Is_void() {
				sc.s.push_err(a.R.Token, "invalid_expr")
			}

			st.L = append(st.L, nil)
			continue
		}

		if a.Declarative && sc.is_new_assign_ident(lexpr.Ident) {
			// Add new variable declaration statement.
			v := &Var{
				Ident:   lexpr.Ident,
				Token:   lexpr.Token,
				Mutable: lexpr.Mutable,
				Scope:   sc.tree,
				Value: &Value{
					Expr: a.R,
					Data: r,
				},
			}

			sc.s.check_var(v)

			st.L = append(st.L, v)
			sc.table.Vars = append(sc.table.Vars, v)
			sc.scope.Stmts = append(sc.scope.Stmts, v)

			continue
		}

		if lexpr.Mutable {
			sc.s.push_err(lexpr.Token, "duplicated_ident", lexpr.Ident)
		}

		l := sc.s.eval(lexpr.Expr, sc)
		if l == nil {
			continue
		}

		if !check_assign(sc.s, l, r, a.Setter) {
			continue
		}

		sc.s.check_validity_for_init_expr(l.Mutable, l.Kind, r, a.Setter)

		checker := _AssignTypeChecker{
			s:           sc.s,
			dest:        l.Kind,
			d:           r,
			error_token: a.Setter,
			deref:       true,
		}
		checker.check()

		st.L = append(st.L, l.Model)
	}

	sc.scope.Stmts = append(sc.scope.Stmts, st)
}

func (sc *_ScopeChecker) check_assign_st(a *ast.AssignSt) {
	if lex.Is_postfix_op(a.Setter.Kind) {
		sc.check_postfix(a)
		return
	}

	if len(a.L) == 1 && !a.Declarative {
		sc.check_single_assign(a)
		return
	}

	sc.check_multi_assign(a)
}

func (sc *_ScopeChecker) check_case_scope(c *Case, tree *ast.ScopeTree) *Scope {
	ssc := sc.new_child_checker()
	ssc.cse = _uintptr(c)
	return sc.check_child_sc(tree, ssc)
}

func (sc *_ScopeChecker) check_case(m *Match, i int, c *ast.Case, expr *Data) *Case {
	_case := m.Cases[i]
	_case.Exprs = make([]ExprModel, len(c.Exprs))

	for i, e := range c.Exprs {
		if m.Type_match {
			eval := _Eval{
				s:      sc.s,
				lookup: sc,
			}

			d := eval.eval(e)
			if d != nil {
				_case.Exprs[i] = d.Kind
				if count_match_type(m, d.Kind) > 1 {
					sc.s.push_err(e.Token, "duplicate_match_type", d.Kind.To_str())
				}
			}

			trt := expr.Kind.Trt()
			if trt != nil {
				_ = sc.s.check_type_compatibility(expr.Kind, d.Kind, e.Token, false)
			}

			continue
		}

		d := sc.s.eval(e, sc)
		if d == nil {
			continue
		}

		_case.Exprs[i] = d.Model

		checker := _AssignTypeChecker{
			s:           sc.s,
			dest:        expr.Kind,
			d:           d,
			error_token: e.Token,
			deref:       true,
		}
		checker.check()
	}

	_case.Scope = sc.check_case_scope(_case, c.Scope)
	return _case
}

func (sc *_ScopeChecker) check_cases(m *ast.MatchCase, rm *Match, expr *Data) {
	rm.Cases = make([]*Case, len(m.Cases))
	for i := range m.Cases {
		_case := &Case{
			Owner: rm,
		}

		if i > 0 {
			rm.Cases[i-1].Next = _case
		}

		rm.Cases[i] = _case
	}

	if rm.Default != nil && len(m.Cases) > 0 {
		rm.Cases[len(rm.Cases)-1].Next = rm.Default
	}

	for i, c := range m.Cases {
		sc.check_case(rm, i, c, expr)
	}
}

func (sc *_ScopeChecker) check_default(m *Match, d *ast.Else) *Case {
	def := &Case{
		Owner: m,
	}
	def.Scope = sc.check_case_scope(def, d.Scope)
	return def
}

func (sc *_ScopeChecker) check_type_match(m *ast.MatchCase) {
	d := sc.s.eval(m.Expr, sc)
	if d == nil {
		return
	}

	if !((d.Kind.Prim() != nil && d.Kind.Prim().Is_any()) || d.Kind.Trt() != nil) {
		sc.s.push_err(m.Expr.Token, "type_case_has_not_valid_expr")
		return
	}

	tm := &Match{
		Type_match: true,
		Expr:       d.Model,
	}

	sc.scope.Stmts = append(sc.scope.Stmts, tm)

	if m.Default != nil {
		tm.Default = sc.check_default(tm, m.Default)
	}
	sc.check_cases(m, tm, d)
}

func (sc *_ScopeChecker) check_common_match(m *ast.MatchCase) {
	var d *Data
	if m.Expr == nil {
		d = &Data{
			Constant: constant.New_bool(true),
			Kind:     &TypeKind{kind: build_prim_type(types.TypeKind_BOOL)},
		}
		d.Model = d.Constant
	} else {
		d = sc.s.eval(m.Expr, sc)
		if d == nil {
			return
		}
	}

	mc := &Match{
		Expr: d.Model,
	}

	sc.scope.Stmts = append(sc.scope.Stmts, mc)

	if m.Default != nil {
		mc.Default = sc.check_default(mc, m.Default)
	}
	sc.check_cases(m, mc, d)
}

func (sc *_ScopeChecker) check_match(m *ast.MatchCase) {
	if m.Type_match {
		sc.check_type_match(m)
		return
	}
	sc.check_common_match(m)
}

func (sc *_ScopeChecker) check_fall(f *ast.FallSt) {
	if sc.cse == 0 || len(sc.scope.Stmts)+1 < len(sc.scope.Stmts) {
		sc.s.push_err(f.Token, "fallthrough_wrong_use")
		return
	}

	_case := (*Case)(unsafe.Pointer(sc.cse))
	if _case.Next == nil {
		sc.s.push_err(f.Token, "fallthrough_into_final_case")
		return
	}

	sc.scope.Stmts = append(sc.scope.Stmts, &FallSt{
		Dest_case: _uintptr(_case.Next),
	})
}

func (sc *_ScopeChecker) check_break_with_label(b *ast.BreakSt) *BreakSt {
	brk := sc.check_plain_break(b)
	if brk == nil {
		return nil
	}

	// Set pointer to zero.
	// Pointer will set by label.
	brk.It = 0
	brk.Mtch = 0

	label := find_label_parent(b.Label.Kind, sc.parent)
	if label == nil {
		sc.s.push_err(b.Label, "label_not_exist", b.Label.Kind)
		return nil
	}

	label.used = true

	if label.pos+1 >= len(label.scope.scope.Stmts) {
		sc.s.push_err(b.Label, "invalid_label")
		return nil
	}

	i := label.pos + 1
	if i >= len(label.scope.scope.Stmts) {
		sc.s.push_err(b.Label, "invalid_label")
	} else {
		st := label.scope.scope.Stmts[i]
		switch st.(type) {
		case *InfIter:
			brk.It = _uintptr(st.(*InfIter))

		case *RangeIter:
			brk.It = _uintptr(st.(*RangeIter))

		case *WhileIter:
			brk.It = _uintptr(st.(*WhileIter))

		case *Match:
			brk.Mtch = _uintptr(st.(*Match))

		default:
			sc.s.push_err(b.Label, "invalid_label")
		}
	}

	if brk.It != 0 {
		if !sc.check_valid_break_label(brk.It) {
			sc.s.push_err(b.Label, "invalid_label")
		}
	}

	if brk.Mtch != 0 {
		if !sc.check_valid_break_label(brk.Mtch) {
			sc.s.push_err(b.Label, "invalid_label")
		}
	}

	return brk
}

func (sc *_ScopeChecker) check_plain_break(b *ast.BreakSt) *BreakSt {
	scope := sc
iter:
	switch {
	case scope.it == 0 && scope.cse == 0 && scope.parent != nil && scope.owner == nil:
		scope = scope.parent
		goto iter

	case scope.it != 0:
		return &BreakSt{It: scope.it}

	case scope.cse != 0:
		return &BreakSt{Mtch: _uintptr((*Case)(unsafe.Pointer(scope.cse)).Owner)}
	}

	sc.s.push_err(b.Token, "break_at_out_of_valid_scope")
	return nil
}

func (sc *_ScopeChecker) check_break(b *ast.BreakSt) {
	if b.Label.Id != lex.ID_NA { // Label given.
		brk := sc.check_break_with_label(b)
		sc.scope.Stmts = append(sc.scope.Stmts, brk)
		return
	}

	brk := sc.check_plain_break(b)
	sc.scope.Stmts = append(sc.scope.Stmts, brk)
}

func (sc *_ScopeChecker) check_ret(r *ast.RetSt) {
	rt := &RetSt{}
	sc.scope.Stmts = append(sc.scope.Stmts, rt)

	var d *Data = nil

	if r.Expr != nil {
		d = sc.s.eval(r.Expr, sc)
		if d == nil {
			return
		}
	}

	rtc := &_RetTypeChecker{
		sc:          sc,
		f:           sc.get_root().owner,
		error_token: r.Token,
	}
	ok := rtc.check(d)
	if !ok {
		return
	}

	if d == nil && len(rtc.vars) == 0 {
		return
	}

	rt.Vars = rtc.vars

	if d != nil {
		rt.Expr = d.Model
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

	case *ast.ContSt:
		sc.check_cont(node.(*ast.ContSt))

	case *ast.LabelSt:
		sc.check_label(node.(*ast.LabelSt))

	case *ast.GotoSt:
		sc.push_goto(node.(*ast.GotoSt))

	case *ast.AssignSt:
		sc.check_assign_st(node.(*ast.AssignSt))

	case *ast.MatchCase:
		sc.check_match(node.(*ast.MatchCase))

	case *ast.FallSt:
		sc.check_fall(node.(*ast.FallSt))

	case *ast.BreakSt:
		sc.check_break(node.(*ast.BreakSt))

	case *ast.RetSt:
		sc.check_ret(node.(*ast.RetSt))

	default:
		println("error <unimplemented scope node>")
	}
}

func (sc *_ScopeChecker) check_tree() {
	sc.i = 0
	for ; sc.i < len(sc.tree.Stmts); sc.i++ {
		sc.check_node(sc.tree.Stmts[sc.i])
	}
}

func st_is_def(st St) bool {
	switch st.(type) {
	case *Var:
		return true

	default:
		return false
	}
}

func (sc *_ScopeChecker) check_same_scope_goto(gt *_ScopeGoto, label *_ScopeLabel) {
	if label.pos < gt.pos { // Label at above.
		return
	}

	i := label.pos
	for ; i > gt.pos; i-- {
		s := label.scope.scope.Stmts[i]
		if st_is_def(s) {
			sc.s.push_err(gt.gt.Token, "goto_jumps_declarations", gt.gt.Label.Kind)
			break
		}
	}
}

func (sc *_ScopeChecker) check_label_parents(gt *_ScopeGoto, label *_ScopeLabel) bool {
	scope := label.scope

parent_scopes:
	if scope.parent != nil && scope.parent != gt.scope {
		scope = scope.parent
		for i := 0; i < len(scope.scope.Stmts); i++ {
			switch {
			case i >= label.pos:
				return true

			case st_is_def(scope.scope.Stmts[i]):
				sc.s.push_err(gt.gt.Token, "goto_jumps_declarations", gt.gt.Label.Kind)
				return false
			}
		}

		goto parent_scopes
	}

	return true
}

func (sc *_ScopeChecker) check_goto_scope(gt *_ScopeGoto, label *_ScopeLabel) {
	for i := gt.pos; i < len(gt.scope.scope.Stmts); i++ {
		switch {
		case i >= label.pos:
			return

		case st_is_def(gt.scope.scope.Stmts[i]):
			sc.s.push_err(gt.gt.Token, "goto_jumps_declarations", gt.gt.Label.Kind)
			return
		}
	}
}

func (sc *_ScopeChecker) check_diff_scope_goto(gt *_ScopeGoto, label *_ScopeLabel) {
	switch {
	case label.scope.child_index > 0 && gt.scope.child_index == 0:
		if !sc.check_label_parents(gt, label) {
			return
		}

	case label.scope.child_index < gt.scope.child_index: // Label at parent blocks.
		return
	}

	scope := label.scope
	for i := label.pos - 1; i >= 0; i-- {
		s := scope.scope.Stmts[i]
		switch s.(type) {
		case *Scope:
			if i <= gt.pos {
				return
			}
		}

		if st_is_def(s) {
			sc.s.push_err(gt.gt.Token, "goto_jumps_declarations", gt.gt.Label.Kind)
			break
		}
	}

	// Parent Scopes
	if scope.parent != nil && scope.parent != gt.scope {
		_ = sc.check_label_parents(gt, label)
	} else { // goto Scope
		sc.check_goto_scope(gt, label)
	}
}

func (sc *_ScopeChecker) check_goto(gt *_ScopeGoto, label *_ScopeLabel) {
	switch {
	case gt.scope == label.scope:
		sc.check_same_scope_goto(gt, label)

	case label.scope.child_index > 0:
		sc.check_diff_scope_goto(gt, label)
	}
}

func (sc *_ScopeChecker) check_gotos() {
	for _, gt := range *sc.gotos {
		label := sc.find_label_all(gt.gt.Label.Kind)
		if label == nil {
			sc.s.push_err(gt.gt.Token, "label_not_exist", gt.gt.Label.Kind)
			continue
		}

		label.used = true
		sc.check_goto(gt, label)
	}
}

func (sc *_ScopeChecker) check_labels() {
	for _, l := range *sc.labels {
		if !l.used {
			sc.s.push_err(l.token, "declared_but_not_used", l.label.Ident)
		}
	}
}

func (sc *_ScopeChecker) check_vars() {
	for _, v := range sc.table.Vars {
		if !v.Used && !lex.Is_ignore_ident(v.Ident) && !lex.Is_anon_ident(v.Ident) && v.Ident != lex.KND_SELF {
			sc.s.push_err(v.Token, "declared_but_not_used", v.Ident)
		}
	}
}

func (sc *_ScopeChecker) check_aliases() {
	for _, a := range sc.table.Type_aliases {
		if !a.Used && !lex.Is_ignore_ident(a.Ident) && !lex.Is_anon_ident(a.Ident) {
			sc.s.push_err(a.Token, "declared_but_not_used", a.Ident)
		}
	}
}

// Checks scope tree.
func (sc *_ScopeChecker) check(tree *ast.ScopeTree, s *Scope) {
	s.Deferred = tree.Deferred
	s.Unsafety = tree.Unsafety

	sc.tree = tree
	sc.scope = s

	sc.check_tree()

	sc.check_vars()
	sc.check_aliases()

	if sc.is_root() {
		sc.check_gotos()
		sc.check_labels()
	}
}

func (sc *_ScopeChecker) new_child_checker() *_ScopeChecker {
	base := new_scope_checker_base(sc.s, nil)
	base.parent = sc
	base.labels = sc.labels
	base.gotos = sc.gotos
	base.child_index = sc.child_index + 1
	return base
}

func new_scope_checker_base(s *_Sema, owner *FnIns) *_ScopeChecker {
	return &_ScopeChecker{
		s:     s,
		owner: owner,
		table: &SymbolTable{},
	}
}

func new_scope_checker(s *_Sema, owner *FnIns) *_ScopeChecker {
	base := new_scope_checker_base(s, owner)
	base.labels = new([]*_ScopeLabel)
	base.gotos = new([]*_ScopeGoto)
	return base
}

// Returns label by identifier.
// Returns nil if not exist any label in this identifier.
// Lookups given scope and parent scopes.
func find_label_parent(ident string, scope *_ScopeChecker) *_ScopeLabel {
	label := scope.find_label_scope(ident)
	for label == nil {
		if scope.parent == nil || scope.owner != nil {
			return nil
		}

		scope = scope.parent
		label = scope.find_label_scope(ident)
	}

	return label
}

func count_match_type(m *Match, t *TypeKind) int {
	n := 0
	kind := t.To_str()
loop:
	for _, c := range m.Cases {
		if c == nil {
			continue
		}

		for _, expr := range c.Exprs {
			// Break loop because this expression is not parsed yet.
			// So, parsed cases finished.
			if expr == nil {
				break loop
			}

			if kind == expr.(*TypeKind).To_str() {
				n++
			}
		}
	}
	return n
}

func get_datas_from_tuple_data(d *Data) []*Data {
	if d.Kind.Tup() != nil {
		switch d.Model.(type) {
		case *TupleExprModel:
			return d.Model.(*TupleExprModel).Datas

		default:
			t := d.Kind.Tup()
			r := make([]*Data, len(t.Types))
			for i, kind := range t.Types {
				r[i] = &Data{
					Mutable: true, // Function return.
					Kind:    kind,
				}
			}
			return r
		}
	} else {
		return []*Data{d}
	}
}

func check_mut(s *_Sema, left *Data, right *Data, error_token lex.Token) (ok bool) {
	switch {
	case !left.Mutable:
		s.push_err(error_token, "assignment_to_non_mut")
		return false

	case right != nil && !right.Mutable && is_mut(right.Kind):
		s.push_err(error_token, "assignment_non_mut_to_mut")
		return false

	default:
		return true
	}
}

func check_assign(s *_Sema, left *Data, right *Data, error_token lex.Token) (ok bool) {
	f := left.Kind.Fnc()
	if f != nil && f.Decl != nil && f.Decl.Global {
		s.push_err(error_token, "assign_type_not_support_value")
		return false
	}

	switch {
	case !left.Lvalue:
		s.push_err(error_token, "assign_require_lvalue")
		return false

	case left.Is_const():
		s.push_err(error_token, "assign_const")
		return false

	case !check_mut(s, left, right, error_token):
		return false

	default:
		return true
	}
}
