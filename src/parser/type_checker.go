package parser

import (
	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
	"github.com/julelang/jule/pkg/juletype"
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
	tc.l = types.DerefPtrOrRef(tc.l)
	return tc.check()
}

func (tc *type_checker) check_ptr() bool {
	if tc.r.Id == juletype.NIL {
		return true
	} else if types.IsUnsafePtr(tc.l) {
		return true
	}
	return tc.l.Kind == tc.r.Kind
}

func (tc *type_checker) check_trait() bool {
	if tc.r.Id == juletype.NIL {
		return true
	}
	t := tc.l.Tag.(*models.Trait)
	lm := tc.l.Modifiers()
	ref := false
	switch {
	case types.IsRef(tc.r):
		ref = true
		tc.r = types.DerefPtrOrRef(tc.r)
		if !types.IsStruct(tc.r) {
			break
		}
		fallthrough
	case types.IsStruct(tc.r):
		if lm != "" {
			return false
		}
		rm := tc.r.Modifiers()
		if rm != "" {
			return false
		}
		s := tc.r.Tag.(*models.Struct)
		if !s.HasTrait(t) {
			return false
		}
		if trait_has_reference_receiver(t) && !ref {
			tc.error_logged = true
			tc.p.pusherrtok(tc.errtok, "trait_has_reference_parametered_function")
			return false
		}
		return true
	case types.IsTrait(tc.r):
		return t == tc.r.Tag.(*models.Trait) && lm == tc.r.Modifiers()
	}
	return false
}

func (tc *type_checker) check_struct() bool {
	if tc.r.Tag == nil {
		return false
	}
	s1, s2 := tc.l.Tag.(*models.Struct), tc.r.Tag.(*models.Struct)
	switch {
	case s1.Id != s2.Id,
		s1.Token.File != s2.Token.File:
		return false
	}
	if len(s1.Generics) == 0 {
		return true
	}
	n1, n2 := len(s1.GetGenerics()), len(s2.GetGenerics())
	if n1 != n2 {
		return false
	}
	for i, g1 := range s1.GetGenerics() {
		g2 := s2.GetGenerics()[i]
		if !types.Equals(g1, g2) {
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
	if !types.IsArray(tc.r) {
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
	case types.IsTrait(tc.l), types.IsTrait(tc.r):
		if types.IsTrait(tc.r) {
			tc.l, tc.r = tc.r, tc.l
		}
		return tc.check_trait()
	case types.IsRef(tc.l), types.IsRef(tc.r):
		if types.IsRef(tc.r) {
			tc.l, tc.r = tc.r, tc.l
		}
		return tc.check_ref()
	case types.IsPtr(tc.l), types.IsPtr(tc.r):
		if !types.IsPtr(tc.l) && types.IsPtr(tc.r) {
			tc.l, tc.r = tc.r, tc.l
		}
		return tc.check_ptr()
	case types.IsSlice(tc.l), types.IsSlice(tc.r):
		if types.IsSlice(tc.r) {
			tc.l, tc.r = tc.r, tc.l
		}
		return tc.check_slice()
	case types.IsArray(tc.l), types.IsArray(tc.r):
		if types.IsArray(tc.r) {
			tc.l, tc.r = tc.r, tc.l
		}
		return tc.check_array()
	case types.IsMap(tc.l), types.IsMap(tc.r):
		if types.IsMap(tc.r) {
			tc.l, tc.r = tc.r, tc.l
		}
		return tc.check_map()
	case types.IsNilCompatible(tc.l):
		return tc.r.Id == juletype.NIL
	case types.IsNilCompatible(tc.r):
		return tc.l.Id == juletype.NIL
	case types.IsEnum(tc.l), types.IsEnum(tc.r):
		return tc.l.Id == tc.r.Id && tc.l.Kind == tc.r.Kind
	case types.IsStruct(tc.l), types.IsStruct(tc.r):
		if tc.r.Id == juletype.STRUCT {
			tc.l, tc.r = tc.r, tc.l
		}
		return tc.check_struct()
	}
	return juletype.TypesAreCompatible(tc.l.Id, tc.r.Id, tc.ignore_any)
}
