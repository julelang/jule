package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

// This file reserved for type compatibility checking.

func trait_has_reference_receiver(t *Trait) bool {
	for _, f := range t.Methods {
		p := f.Params[0]
		if p.Is_ref() && p.Is_self() {
			return true
		}
	}
	return false
}

type _TypeCompatibilityChecker struct {
	s           *_Sema    // Used for error logging.
	dest        *TypeKind
	src         *TypeKind
	error_token lex.Token

	// References uses elem's type itself if true.
	deref       bool
}

func (tcc *_TypeCompatibilityChecker) push_err(key string, args ...any) {
	tcc.s.push_err(tcc.error_token, key, args...)
}

func (tcc *_TypeCompatibilityChecker) check_trait() (ok bool) {
	if tcc.src.Is_nil() {
		return true
	}

	trt := tcc.dest.Trt()
	ref := false
	switch {
	case tcc.src.Ref() != nil:
		ref = true
		tcc.src = tcc.src.Ref().Elem
		if tcc.src.Strct() == nil {
			return false
		}
		fallthrough

	case tcc.src.Strct() != nil:
		s := tcc.src.Strct()
		if !s.Decl.Is_implements(trt) {
			return false
		}

		if trait_has_reference_receiver(trt) && !ref {
			tcc.push_err("trait_has_reference_parametered_function")
			return false
		}

		return true

	case tcc.src.Trt() != nil:
		return trt == tcc.src.Trt()
	
	default:
		return false
	}
}

func (tcc *_TypeCompatibilityChecker) check_ref() (ok bool) {
	if tcc.dest.To_str() == tcc.src.To_str() {
		return true
	} else if !tcc.deref {
		return false
	}
	tcc.src = tcc.src.Ref().Elem
	return tcc.check()
}

func (tcc *_TypeCompatibilityChecker) check_ptr() (ok bool) {
	if tcc.src.Is_nil() {
		return true
	} else if tcc.src.Ptr() != nil && tcc.src.Ptr().Is_unsafe() {
		return true
	}
	return tcc.dest.To_str() == tcc.src.To_str()
}

func (tcc *_TypeCompatibilityChecker) check_slc() (ok bool) {
	if tcc.src.Is_nil() {
		return true
	}
	return tcc.dest.To_str() == tcc.src.To_str()
}

func (tcc *_TypeCompatibilityChecker) check_arr() (ok bool) {
	src := tcc.src.Arr()
	if src == nil {
		return false
	}
	dest := tcc.dest.Arr()
	return dest.N == src.N
}

func (tcc *_TypeCompatibilityChecker) check_map() (ok bool) {
	if tcc.src.Is_nil() {
		return true
	}
	return tcc.dest.To_str() == tcc.src.To_str()
}

func (tcc *_TypeCompatibilityChecker) check_struct() (ok bool) {
	src := tcc.src.Strct()
	if src == nil {
		return false
	}
	dest := tcc.dest.Strct()
	switch {
	case dest.Decl != src.Decl:
		return false

	case len(dest.Generics) == 0:
		return true
	}

	for i, dg := range dest.Generics {
		sg := src.Generics[i]
		if dg.To_str() != sg.To_str() {
			return false
		}
	}
	return true
}

func (tcc *_TypeCompatibilityChecker) check_enum() (ok bool) {
	r := tcc.src.Enm()
	if r == nil {
		return false
	}
	return tcc.dest.Enm() == r
}

func (tcc *_TypeCompatibilityChecker) check() (ok bool) {
	switch {
	case tcc.dest.Trt() != nil:
		return tcc.check_trait()

	case tcc.dest.Ref() != nil:
		return tcc.check_ref()

	case tcc.dest.Ptr() != nil:
		return tcc.check_ptr()

	case tcc.dest.Slc() != nil:
		return tcc.check_slc()

	case tcc.dest.Arr() != nil:
		return tcc.check_arr()

	case tcc.dest.Map() != nil:
		return tcc.check_map()

	case tcc.dest.Enm() != nil:
		return tcc.check_enum()

	case tcc.dest.Strct() != nil:
		return tcc.check_struct()
	
	case is_nil_compatible(tcc.dest):
		return tcc.src.Is_nil()

	default:
		return types.Types_are_compatible(tcc.dest.To_str(), tcc.src.To_str())
	}
}

// Checks value and type compatibility for assignment.
type _AssignTypeChecker struct {
	s           *_Sema    // Used for error logging and type checking.
	dest        *TypeKind
	d           *Data
	error_token lex.Token
	deref       bool     // Same as TypeCompatibilityChecker.deref field.
}

func (tcc *_AssignTypeChecker) push_err(key string, args ...any) {
	tcc.s.push_err(tcc.error_token, key, args...)
}

func (atc *_AssignTypeChecker) check_validity() bool {
	valid := true

	switch {
	case atc.d.Kind.Func() != nil:
		f := atc.d.Kind.Func()
		if f.Decl.Is_method() {
			atc.push_err("method_as_anonymous_fn")
			valid = false
		} else if len(f.Decl.Generics) > 0 {
			atc.push_err("genericed_fn_as_anonymous_fn")
			valid = false
		}

	case atc.d.Kind.Tup() != nil:
		atc.push_err("tuple_assign_to_single")
		valid = false
	}

	return valid
}

func (atc *_AssignTypeChecker) check() {
	// TODO: Check constants.
	switch {
	case atc.d == nil:
		// Skip Data is nil.
		return

	case !atc.check_validity():
		// Data is invalid and error(s) logged about it.
		return
	
	default:
		atc.s.check_type_compatibility(atc.dest, atc.d.Kind, atc.error_token, atc.deref)
	}
}

type _DynamicTypeAnnotation struct {
	e           *_Eval
	f           *FnIns
	p           *ParamIns
	a           *Data
	error_token lex.Token
	
	generics  []*ast.Generic
	k         **TypeKind
}

func (dta *_DynamicTypeAnnotation) push_generic(k *TypeKind, i int) {
	if k.Enm() != nil {
		dta.e.push_err(dta.error_token, "enum_not_supports_as_generic")
	}
	dta.f.Generics[i] = k
}

func (dta *_DynamicTypeAnnotation) annotate_prim(k *TypeKind) (ok bool) {
	kind := (*dta.k).To_str()
	for i, g := range dta.generics {
		if kind != g.Ident {
			continue
		}

		t := &dta.f.Generics[i]
		if t == nil {
			dta.push_generic(k, i)
		}
		*dta.k = k
		return true
	}

	return false
}

func (dta *_DynamicTypeAnnotation) annotate_slc(k *TypeKind) (ok bool) {
	pslc := (*dta.k).Slc()
	if pslc == nil {
		return false
	}

	slc := k.Slc()
	dta.k = &pslc.Elem
	return dta.annotate_kind(slc.Elem)
}

func (dta *_DynamicTypeAnnotation) annotate_map(k *TypeKind) (ok bool) {
	pmap := (*dta.k).Map()
	if pmap == nil {
		return false
	}

	m := k.Map()
	check := func(k **TypeKind, ck *TypeKind) (ok bool) {
		old := dta.k
		dta.k = k
		ok = dta.annotate_kind(ck)
		dta.k = old
		return ok
	}
	return check(&pmap.Key, m.Key) && check(&pmap.Val, m.Val)
}

func (dta *_DynamicTypeAnnotation) annotate_kind(k *TypeKind) (ok bool) {
	// TODO: Implement other types.
	switch {
	case k.Prim() != nil:
		return dta.annotate_prim(k)

	case k.Slc() != nil:
		return dta.annotate_slc(k)

	case k.Map() != nil:
		return dta.annotate_map(k)

	default:
		return false
	}
}

func (dta *_DynamicTypeAnnotation) annotate() (ok bool) {
	dta.generics = dta.f.Decl.Generics
	dta.k = &dta.p.Kind

	return dta.annotate_kind(dta.a.Kind)
}

type _FnCallArgChecker struct {
	e                  *_Eval
	args               []*ast.Expr
	error_token        lex.Token
	f                  *FnIns
	dynamic_annotation bool
}

func (fcac *_FnCallArgChecker) push_err_token(token lex.Token, key string, args ...any) {
	fcac.e.s.push_err(token, key, args...)
}

func (fcac *_FnCallArgChecker) push_err(key string, args ...any) {
	fcac.push_err_token(fcac.error_token, key, args...)
}

func (fcac *_FnCallArgChecker) get_params() []*ParamIns {
	if fcac.f.Decl.Is_method() {
		return fcac.f.Params[1:] // Remove receiver parameter.
	}
	return fcac.f.Params
}

func (fcac *_FnCallArgChecker) tuple_as_params(params []*ParamIns) bool {
	return len(params) > 1 && len(fcac.args) == 1
}

func (fcac *_FnCallArgChecker) check_counts(params []*ParamIns) (ok bool) {
	n := len(params)
	if n > 0 && params[n-1].Decl.Variadic {
		n--
	}

	diff := n - len(fcac.args)
	switch {
	case diff <= 0:
		return true

	case diff > len(params):
		fcac.push_err("argument_overflow")
		return false
	}

	idents := ""
	for ; diff > 0; diff-- {
		idents += ", " + params[n-diff].Decl.Ident
	}
	idents = idents[2:] // Remove first separator.
	fcac.push_err("missing_expr_for", idents)

	return false
}

func (fcac *_FnCallArgChecker) check_arg(p *ParamIns, arg *Data, error_token lex.Token) (ok bool) {
	if fcac.dynamic_annotation {
		dta := _DynamicTypeAnnotation{
			e:           fcac.e,
			f:           fcac.f,
			p:           p,
			a:           arg,
			error_token: error_token,
		}
		ok = dta.annotate()
		if !ok {
			fcac.push_err_token(error_token, "dynamic_type_annotation_failed")
			return false
		}
	}

	fcac.e.s.check_validity_for_init_expr(p.Decl.Mutable, arg, error_token)
	fcac.e.s.check_assign_type(p.Kind, arg, error_token, false)
	return true
}

func (fcac *_FnCallArgChecker) try_tuple_as_params(params []*ParamIns) (ok bool) {
	d := fcac.e.eval_expr_kind(fcac.args[0].Kind)
	if d == nil {
		return false
	}

	tup := d.Kind.Tup()
	if tup == nil {
		return false
	}

	if len(tup.Types) != len(params) {
		return false
	}

	for i, arg := range tup.Types {
		param := params[i]
		d := Data{Kind: arg}
		ok = fcac.check_arg(param, &d, fcac.args[0].Token) && ok
	}

	return ok
}

func (fcac *_FnCallArgChecker) push(p *ParamIns, arg *ast.Expr) (ok bool) {
	d := fcac.e.eval_expr_kind(arg.Kind)
	if d == nil {
		return false
	}
	return fcac.check_arg(p, d, arg.Token)
}

func (fcac *_FnCallArgChecker) push_variadic(p *ParamIns, i int) (ok bool) {
	variadiced := false
	more := i+1 < len(fcac.args)
	for ; i < len(fcac.args); i++ {
		arg := fcac.args[i]
		d := fcac.e.eval_expr_kind(arg.Kind)
		if d == nil {
			ok = false
			continue
		}

		if d.Variadiced {
			variadiced = true
			d.Kind = d.Kind.Slc().Elem
		}

		ok = fcac.check_arg(p, d, arg.Token) && ok
	}

	if variadiced && more {
		fcac.push_err("more_args_with_variadiced")
	}

	return ok
}

func (fcac *_FnCallArgChecker) check_args(params []*ParamIns) (ok bool) {
	i := 0
iter:
	for i < len(params) {
		p := params[i]
		switch {
		case p.Decl.Variadic:
			// Variadiced parameters always last.
			ok = fcac.push_variadic(p, i) && ok
			break iter

		default:
			ok = fcac.push(p, fcac.args[i]) && ok
			i++
		}
	}

	return ok
}

func (fcac *_FnCallArgChecker) check_dynamic_type_annotation() (ok bool) {
	for _, g := range fcac.f.Generics {
		if g == nil {
			fcac.push_err("dynamic_type_annotation_failed")
			return false
		}
	}
	return true
}

func (fcac *_FnCallArgChecker) check() (ok bool) {
	params := fcac.get_params()

	if fcac.tuple_as_params(params) {
		ok = fcac.try_tuple_as_params(params)
		if ok {
			return true
		}
	}

	ok = fcac.check_counts(params)
	if !ok {
		return false
	}

	ok = fcac.check_args(params)
	if ok && fcac.dynamic_annotation {
		ok = fcac.check_dynamic_type_annotation()
	}

	return ok
}
