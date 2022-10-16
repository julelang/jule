package parser

import (
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/pkg/juletype"
)

type type_checker struct {
	errtok       lex.Token
	p            *Parser
	left         Type
	right        Type
	error_logged bool
	ignore_any   bool
	allow_assign bool
}

func (tc *type_checker) check_ref() bool {
	if tc.left.Kind == tc.right.Kind {
		return true
	} else if !tc.allow_assign {
		return false
	}
	tc.left = un_ptr_or_ref_type(tc.left)
	return tc.check()
}

func (tc *type_checker) check_ptr() bool {
	if tc.right.Id == juletype.NIL {
		return true
	} else if type_is_unsafe_ptr(tc.left) {
		return true
	}
	return tc.left.Kind == tc.right.Kind
}

func (tc *type_checker) check_trait() bool {
	if tc.right.Id == juletype.NIL {
		return true
	}
	t := tc.left.Tag.(*trait)
	lm := tc.left.Modifiers()
	ref := false
	switch {
	case type_is_ref(tc.right):
		ref = true
		tc.right = un_ptr_or_ref_type(tc.right)
		if !type_is_struct(tc.right) {
			break
		}
		fallthrough
	case type_is_struct(tc.right):
		if lm != "" {
			return false
		}
		rm := tc.right.Modifiers()
		if rm != "" {
			return false
		}
		s := tc.right.Tag.(*structure)
		if !s.hasTrait(t) {
			return false
		}
		if t.has_reference_receiver() && !ref {
			tc.error_logged = true
			tc.p.pusherrtok(tc.errtok, "trait_has_reference_parametered_function")
			return false
		}
		return true
	case type_is_trait(tc.right):
		return t == tc.right.Tag.(*trait) && lm == tc.right.Modifiers()
	}
	return false
}

func (tc *type_checker) check_struct() bool {
	s1, s2 := tc.left.Tag.(*structure), tc.right.Tag.(*structure)
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
	if tc.right.Id == juletype.NIL {
		return true
	}
	return tc.left.Kind == tc.right.Kind
}

func (tc *type_checker) check_array() bool {
	if !type_is_array(tc.right) {
		return false
	}
	return tc.left.Size.N == tc.right.Size.N
}

func (tc *type_checker) check_map() bool {
	if tc.right.Id == juletype.NIL {
		return true
	}
	return tc.left.Kind == tc.right.Kind
}

func (tc *type_checker) check() bool {
	switch {
	case type_is_trait(tc.left), type_is_trait(tc.right):
		if type_is_trait(tc.right) {
			tc.left, tc.right = tc.right, tc.left
		}
		return tc.check_trait()
	case type_is_ref(tc.left), type_is_ref(tc.right):
		if type_is_ref(tc.right) {
			tc.left, tc.right = tc.right, tc.left
		}
		return tc.check_ref()
	case type_is_ptr(tc.left), type_is_ptr(tc.right):
		if !type_is_ptr(tc.left) && type_is_ptr(tc.right) {
			tc.left, tc.right = tc.right, tc.left
		}
		return tc.check_ptr()
	case type_is_slc(tc.left), type_is_slc(tc.right):
		if type_is_slc(tc.right) {
			tc.left, tc.right = tc.right, tc.left
		}
		return tc.check_slice()
	case type_is_array(tc.left), type_is_array(tc.right):
		if type_is_array(tc.right) {
			tc.left, tc.right = tc.right, tc.left
		}
		return tc.check_array()
	case type_is_map(tc.left), type_is_map(tc.right):
		if type_is_map(tc.right) {
			tc.left, tc.right = tc.right, tc.left
		}
		return tc.check_map()
	case type_is_nil_compatible(tc.left):
		return tc.right.Id == juletype.NIL
	case type_is_nil_compatible(tc.right):
		return tc.left.Id == juletype.NIL
	case type_is_enum(tc.left), type_is_enum(tc.right):
		return tc.left.Id == tc.right.Id && tc.left.Kind == tc.right.Kind
	case type_is_struct(tc.left), type_is_struct(tc.right):
		if tc.right.Id == juletype.STRUCT {
			tc.left, tc.right = tc.right, tc.left
		}
		return tc.check_struct()
	}
	return juletype.TypesAreCompatible(tc.left.Id, tc.right.Id, tc.ignore_any)
}
