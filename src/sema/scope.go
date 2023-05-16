package sema

import (
	"github.com/julelang/jule/ast"
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
	L     []ExprModel // Nil models represents ingored expressions.
	R     ExprModel
}

type _ScopeLabel struct {
	label *Label
	pos   int
	scope *_ScopeChecker
}

type _ScopeGoto struct {
	gt    *ast.GotoSt
	scope *_ScopeChecker
	pos   int
}

// Scope checker.
type _ScopeChecker struct {
	s           *_Sema
	parent      *_ScopeChecker
	child_index int // Index of child scope.
	table       *SymbolTable
	scope       *Scope
	tree        *ast.ScopeTree
	it          uintptr
	labels      *[]*_ScopeLabel // All labels of all scopes.
	gotos       *[]*_ScopeGoto  // All gotos of all scopes.
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

	if conditional.Default != nil {
		c.Default = sc.check_else(conditional.Default)
	}

	sc.scope.Stmts = append(sc.scope.Stmts, c)
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

func (sc *_ScopeChecker) check_while_iter(it *ast.Iter) {
	kind := &WhileIter{}

	sc.scope.Stmts = append(sc.scope.Stmts, kind)

	kind.Scope = sc.check_iter_scope(_uintptr(kind), it.Scope)

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

func (sc *_ScopeChecker) check_valid_iter_label(it uintptr) bool {
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

func (sc *_ScopeChecker) check_cont(c *ast.ContSt) {
	if sc.it == 0 {
		sc.s.push_err(c.Token, "continue_at_out_of_valid_scope")
	}

	cont := &ContSt{It: sc.it}

	if c.Label.File != nil { // Label given.
		label := find_label_parent(c.Label.Kind, sc.parent)
		if label == nil {
			sc.s.push_err(c.Label, "label_not_exist", c.Label.Kind)
			return
		} else if label.pos+1 >= len(label.scope.scope.Stmts) {
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
		if !sc.check_valid_iter_label(cont.It) {
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

func (sc *_ScopeChecker) check_assign(left *Data, error_token lex.Token) (ok bool) {
	switch {
	case left.Kind.Fnc() != nil:
		sc.s.push_err(error_token, "assign_type_not_support_value")
		return false
	}

	if !left.Lvalue {
		sc.s.push_err(error_token, "assign_require_lvalue")
		return false
	}

	if left.Is_const() {
		sc.s.push_err(error_token, "assign_const")
		return false
	} else if !left.Mutable {
		sc.s.push_err(error_token, "assignment_to_non_mut")
		return false
	}

	return true
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

	_ = sc.check_assign(d, a.Setter)

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
	return sc.table.def_by_ident(ident, false) == nil
}

func (sc *_ScopeChecker) check_single_assign(a *ast.AssignSt) {
	r := sc.s.eval(a.R, sc)
	if r == nil {
		return
	}

	if lex.Is_ignore_ident(a.L[0].Ident) {
		sc.scope.Stmts = append(sc.scope.Stmts, r)
		return
	}

	l := sc.s.eval(a.L[0].Expr, sc)
	if l == nil {
		return
	}

	if !sc.check_assign(l, a.Setter) {
		return
	}

	if r.Kind.Tup() != nil {
		sc.s.push_err(a.Setter, "missing_multi_assign_idents")
		return
	}

	if a.Setter.Kind != lex.KND_EQ && !r.Is_const() {
		a.Setter.Kind = a.Setter.Kind[:len(a.Setter.Kind)-1]

		solver := _BinopSolver{
			e:  &_Eval{
				s:        sc.s,
				lookup:   sc,
				unsafety: sc.is_unsafe(),
			},
			op: a.Setter,
		}

		r = solver.solve_explicit(l, r)
		a.Setter.Kind += lex.KND_EQ
	}

	checker := _AssignTypeChecker{
		s:           sc.s,
		dest:        l.Kind,
		d:           r,
		error_token: a.Setter,
	}
	checker.check()

	sc.scope.Stmts = append(sc.scope.Stmts, &Assign{
		L:  l.Model,
		R:  r.Model,
		Op: a.Setter.Kind,
	})
}

func (sc *_ScopeChecker) check_multi_assign(a *ast.AssignSt) {
	rd := sc.s.eval(a.R, sc)
	if rd == nil {
		return
	}

	var r []*Data
	if rd.Kind.Tup() != nil {
		r = rd.Model.(*TupleExprModel).Datas
	} else {
		r = append(r, rd)
	}

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

		if lex.Is_ignore_ident(lexpr.Ident) {
			st.L = append(st.L, nil)
			continue
		}

		r := r[i]

		if sc.is_new_assign_ident(lexpr.Ident) {
			// Add new variable declaration statement.
			v := &Var{
				Ident:   lexpr.Ident,
				Token:   lexpr.Token,
				Mutable: lexpr.Mutable,
				Scope:   sc.tree,
				Value:   &Value{
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

		if !sc.check_assign(l, a.Setter) {
			continue
		}

		sc.s.check_validity_for_init_expr(l.Mutable, r, a.Setter)

		checker := _AssignTypeChecker{
			s:           sc.s,
			dest:        l.Kind,
			d:           r,
			error_token: a.Setter,
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

	if len(a.L) == 1 && !sc.is_new_assign_ident(a.L[0].Ident) {
		sc.check_single_assign(a)
		return
	}

	sc.check_multi_assign(a)
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

	default:
		println("error <unimplemented scope node>")
	}
}

func (sc *_ScopeChecker) check_tree() {
	for _, node := range sc.tree.Stmts {
		sc.check_node(node)
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
		sc.check_goto(gt, label)
	}
}

// Checks scope tree.
func (sc *_ScopeChecker) check(tree *ast.ScopeTree, s *Scope) {
	s.Deferred = tree.Deferred
	s.Unsafety = tree.Unsafety

	sc.tree = tree
	sc.scope = s

	sc.check_tree()

	if sc.parent == nil { // If parent scope.
		sc.check_gotos()
	}
}

func (sc *_ScopeChecker) new_child_checker() *_ScopeChecker {
	base := new_scope_checker_base(sc.s)
	base.labels = sc.labels
	base.gotos =  sc.gotos
	base.child_index = sc.child_index + 1
	return base
}

func new_scope_checker_base(s *_Sema) *_ScopeChecker {
	return &_ScopeChecker{
		s:      s,
		table:  &SymbolTable{},
	}
}

func new_scope_checker(s *_Sema) *_ScopeChecker {
	base := new_scope_checker_base(s)
	base.labels = new([]*_ScopeLabel)
	base.gotos =  new([]*_ScopeGoto)
	return base
}

// Returns label by identifier.
// Returns nil if not exist any label in this identifier.
// Lookups given scope and parent scopes.
func find_label_parent(ident string, scope *_ScopeChecker) *_ScopeLabel {
	label := scope.find_label_scope(ident)
	for label == nil {
		if scope.parent == nil {
			return nil
		}

		scope = scope.parent
		label = scope.find_label_scope(ident)
	}

	return label
}
