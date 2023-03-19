package sema

import (
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
