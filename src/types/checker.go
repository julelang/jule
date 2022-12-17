package types

import (
	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/pkg/jule"
	"github.com/julelang/jule/pkg/juletype"
)

// Checker is type checker.
type Checker struct {
	ErrTok      lex.Token
	L           Type
	R           Type
	ErrorLogged bool
	IgnoreAny   bool
	AllowAssign bool
	Errors      []build.CompilerLog
}

// pusherrtok appends new error by token.
func (c *Checker) pusherrtok(tok lex.Token, key string, args ...any) {
	c.Errors = append(c.Errors, build.CompilerLog{
		Type:    build.ERR,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    tok.File.Path(),
		Message: jule.GetError(key, args...),
	})
}

func (c *Checker) check_ref() bool {
	if c.L.Kind == c.R.Kind {
		return true
	} else if !c.AllowAssign {
		return false
	}
	c.L = DerefPtrOrRef(c.L)
	return c.Check()
}

func (c *Checker) check_ptr() bool {
	if c.R.Id == juletype.NIL {
		return true
	} else if IsUnsafePtr(c.L) {
		return true
	}
	return c.L.Kind == c.R.Kind
}

func trait_has_reference_receiver(t *models.Trait) bool {
	for _, f := range t.Defines.Funcs {
		if IsRef(f.Receiver.Type) {
			return true
		}
	}
	return false
}

func (c *Checker) check_trait() bool {
	if c.R.Id == juletype.NIL {
		return true
	}
	t := c.L.Tag.(*models.Trait)
	lm := c.L.Modifiers()
	ref := false
	switch {
	case IsRef(c.R):
		ref = true
		c.R = DerefPtrOrRef(c.R)
		if !IsStruct(c.R) {
			break
		}
		fallthrough
	case IsStruct(c.R):
		if lm != "" {
			return false
		}
		rm := c.R.Modifiers()
		if rm != "" {
			return false
		}
		s := c.R.Tag.(*models.Struct)
		if !s.HasTrait(t) {
			return false
		}
		if trait_has_reference_receiver(t) && !ref {
			c.ErrorLogged = true
			c.pusherrtok(c.ErrTok, "trait_has_reference_parametered_function")
			return false
		}
		return true
	case IsTrait(c.R):
		return t == c.R.Tag.(*models.Trait) && lm == c.R.Modifiers()
	}
	return false
}

func (c *Checker) check_struct() bool {
	if c.R.Tag == nil {
		return false
	}
	s1, s2 := c.L.Tag.(*models.Struct), c.R.Tag.(*models.Struct)
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
		if !Equals(g1, g2) {
			return false
		}
	}
	return true
}

func (c *Checker) check_slice() bool {
	if c.R.Id == juletype.NIL {
		return true
	}
	return c.L.Kind == c.R.Kind
}

func (c *Checker) check_array() bool {
	if !IsArray(c.R) {
		return false
	}
	return c.L.Size.N == c.R.Size.N
}

func (c *Checker) check_map() bool {
	if c.R.Id == juletype.NIL {
		return true
	}
	return c.L.Kind == c.R.Kind
}

// Check checks type compatilility and reports.
func (c *Checker) Check() bool {
	switch {
	case IsTrait(c.L), IsTrait(c.R):
		if IsTrait(c.R) {
			c.L, c.R = c.R, c.L
		}
		return c.check_trait()
	case IsRef(c.L), IsRef(c.R):
		if IsRef(c.R) {
			c.L, c.R = c.R, c.L
		}
		return c.check_ref()
	case IsPtr(c.L), IsPtr(c.R):
		if !IsPtr(c.L) && IsPtr(c.R) {
			c.L, c.R = c.R, c.L
		}
		return c.check_ptr()
	case IsSlice(c.L), IsSlice(c.R):
		if IsSlice(c.R) {
			c.L, c.R = c.R, c.L
		}
		return c.check_slice()
	case IsArray(c.L), IsArray(c.R):
		if IsArray(c.R) {
			c.L, c.R = c.R, c.L
		}
		return c.check_array()
	case IsMap(c.L), IsMap(c.R):
		if IsMap(c.R) {
			c.L, c.R = c.R, c.L
		}
		return c.check_map()
	case IsNilCompatible(c.L):
		return c.R.Id == juletype.NIL
	case IsNilCompatible(c.R):
		return c.L.Id == juletype.NIL
	case IsEnum(c.L), IsEnum(c.R):
		return c.L.Id == c.R.Id && c.L.Kind == c.R.Kind
	case IsStruct(c.L), IsStruct(c.R):
		if c.R.Id == juletype.STRUCT {
			c.L, c.R = c.R, c.L
		}
		return c.check_struct()
	}
	return juletype.TypesAreCompatible(c.L.Id, c.R.Id, c.IgnoreAny)
}
