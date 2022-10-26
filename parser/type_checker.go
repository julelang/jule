package parser

import (
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/pkg/juletype"
)

type type_checker struct {
	errtok       lex.Token
	p            *Parser
	l         Type
	r        Type
	error_logged bool
	ignore_any   bool
	allow_assign bool
}

func (tc *type_checker) check_ref() bool {
	if tc.l.Kind == tc.r.Kind {
		return true
	} else if !tc.allow_assign {
		return false
	}
	tc.l = un_ptr_or_ref_type(tc.l)
	return tc.check()
}

func (tc *type_checker) check_ptr() bool {
	if tc.r.Id == juletype.NIL {
		return true
	} else if type_is_unsafe_ptr(tc.l) {
		return true
	}
	return tc.l.Kind == tc.r.Kind
}

func (tc *type_checker) check_trait() bool {
	if tc.r.Id == juletype.NIL {
		return true
	}
	t := tc.l.Tag.(*trait)
	lm := tc.l.Modifiers()
	ref := false
	switch {
	case type_is_ref(tc.r):
		ref = true
		tc.r = un_ptr_or_ref_type(tc.r)
		if !type_is_struct(tc.r) {
			break
		}
		fallthrough
	case type_is_struct(tc.r):
		if lm != "" {
			return false
		}
		rm := tc.r.Modifiers()
		if rm != "" {
			return false
		}
		s := tc.r.Tag.(*structure)
		if !s.hasTrait(t) {
			return false
		}
		if t.has_reference_receiver() && !ref {
			tc.error_logged = true
			tc.p.pusherrtok(tc.errtok, "trait_has_reference_parametered_function")
			return false
		}
		return true
	case type_is_trait(tc.r):
		return t == tc.r.Tag.(*trait) && lm == tc.r.Modifiers()
	}
	return false
}

func (tc *type_checker) check_struct() bool {
	s1, s2 := tc.l.Tag.(*structure), tc.r.Tag.(*structure)
	switch {
	case s1.Ast.Id != s2.Ast.Id,
		s1.Ast.Token.File != s2.Ast.Token.File:
		return false
	}
	if len(s1.Ast.Generics) == 0 {
		return true
	}
	n1, n2 := len(s1.generics), len(s2.generics)
	if n1 != n2 {
		return false
	}
	for i, g1 := range s1.generics {
		g2 := s2.generics[i]
		if !types_equals(g1, g2) {
			return false
		}
	}
	return true
}

func (tc *type_checker) check_slice() bool {
	if tc.r.Id == juletype.NIL {
		return true
	}
	return tc.l.Kind == tc.r.Kind
}

func (tc *type_checker) check_array() bool {
	if !type_is_array(tc.r) {
		return false
	}
	return tc.l.Size.N == tc.r.Size.N
}

func (tc *type_checker) check_map() bool {
	if tc.r.Id == juletype.NIL {
		return true
	}
	return tc.l.Kind == tc.r.Kind
}

func (tc *type_checker) check() bool {
	switch {
	case type_is_trait(tc.l), type_is_trait(tc.r):
		if type_is_trait(tc.r) {
			tc.l, tc.r = tc.r, tc.l
		}
		return tc.check_trait()
	case type_is_ref(tc.l), type_is_ref(tc.r):
		if type_is_ref(tc.r) {
			tc.l, tc.r = tc.r, tc.l
		}
		return tc.check_ref()
	case type_is_ptr(tc.l), type_is_ptr(tc.r):
		if !type_is_ptr(tc.l) && type_is_ptr(tc.r) {
			tc.l, tc.r = tc.r, tc.l
		}
		return tc.check_ptr()
	case type_is_slc(tc.l), type_is_slc(tc.r):
		if type_is_slc(tc.r) {
			tc.l, tc.r = tc.r, tc.l
		}
		return tc.check_slice()
	case type_is_array(tc.l), type_is_array(tc.r):
		if type_is_array(tc.r) {
			tc.l, tc.r = tc.r, tc.l
		}
		return tc.check_array()
	case type_is_map(tc.l), type_is_map(tc.r):
		if type_is_map(tc.r) {
			tc.l, tc.r = tc.r, tc.l
		}
		return tc.check_map()
	case type_is_nil_compatible(tc.l):
		return tc.r.Id == juletype.NIL
	case type_is_nil_compatible(tc.r):
		return tc.l.Id == juletype.NIL
	case type_is_enum(tc.l), type_is_enum(tc.r):
		return tc.l.Id == tc.r.Id && tc.l.Kind == tc.r.Kind
	case type_is_struct(tc.l), type_is_struct(tc.r):
		if tc.r.Id == juletype.STRUCT {
			tc.l, tc.r = tc.r, tc.l
		}
		return tc.check_struct()
	}
	return juletype.TypesAreCompatible(tc.l.Id, tc.r.Id, tc.ignore_any)
}
