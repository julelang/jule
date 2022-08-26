package parser

import (
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/julebits"
	"github.com/jule-lang/jule/pkg/juletype"
)

func setshift(v *value, right uint64) {
	switch {
	case right <= 6:
		v.data.Type.Id = juletype.I8
	case right <= 7:
		v.data.Type.Id = juletype.U8
	case right <= 14:
		v.data.Type.Id = juletype.I16
	case right <= 15:
		v.data.Type.Id = juletype.U16
	case right <= 30:
		v.data.Type.Id = juletype.I32
	case right <= 31:
		v.data.Type.Id = juletype.U32
	case right <= 62:
		v.data.Type.Id = juletype.I64
	case right <= 63:
		v.data.Type.Id = juletype.U64
	case right <= 127:
		v.data.Type.Id = juletype.F32
	default:
		v.data.Type.Id = juletype.F64
	}
}

func bitize(v *value) {
	switch t := v.expr.(type) {
	case float64:
		v.data.Type.Id = juletype.FloatFromBits(julebits.BitsizeFloat(t))
	case int64:
		v.data.Type.Id = juletype.IntFromBits(julebits.BitsizeInt(t))
	case uint64:
		v.data.Type.Id = juletype.UIntFromBits(julebits.BitsizeUInt(t))
	default:
		return
	}
	v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
}

func tonumf(expr any) float64 {
	switch t := expr.(type) {
	case float64:
		return t
	case int64:
		return float64(t)
	case uint64:
		return float64(t)
	}
	return 0
}

func tonumu(expr any) uint64 {
	switch t := expr.(type) {
	case float64:
		return uint64(t)
	case int64:
		return uint64(t)
	case uint64:
		return t
	}
	return 0
}

func tonums(expr any) int64 {
	switch t := expr.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case uint64:
		return int64(t)
	}
	return 0
}

type solver struct {
	p         *Parser
	left      []lex.Token
	left_val  value
	right     []lex.Token
	right_val value
	operator  lex.Token
}

func (s *solver) eq(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case bool:
		v.expr = left == s.right_val.expr.(bool)
	case string:
		v.expr = left == s.right_val.expr.(string)
	case float64:
		v.expr = left == tonumf(s.right_val.expr)
	case int64:
		v.expr = left == tonums(s.right_val.expr)
	case uint64:
		v.expr = left == tonumu(s.right_val.expr)
	}
}

func (s *solver) noteq(v *value) {
	if !s.isConstExpr() {
		return
	}
	s.eq(v)
	v.expr = !v.expr.(bool)
}

func (s *solver) lt(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case float64:
		v.expr = left < tonumf(s.right_val.expr)
	case int64:
		v.expr = left < tonums(s.right_val.expr)
	case uint64:
		v.expr = left < tonumu(s.right_val.expr)
	}
}

func (s *solver) gt(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case float64:
		v.expr = left > tonumf(s.right_val.expr)
	case int64:
		v.expr = left > tonums(s.right_val.expr)
	case uint64:
		v.expr = left > tonumu(s.right_val.expr)
	}
}

func (s *solver) lteq(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case float64:
		v.expr = left <= tonumf(s.right_val.expr)
	case int64:
		v.expr = left <= tonums(s.right_val.expr)
	case uint64:
		v.expr = left <= tonumu(s.right_val.expr)
	}
}

func (s *solver) gteq(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case float64:
		v.expr = left >= tonumf(s.right_val.expr)
	case int64:
		v.expr = left >= tonums(s.right_val.expr)
	case uint64:
		v.expr = left >= tonumu(s.right_val.expr)
	}
}

func (s *solver) add(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case string:
		v.expr = left + s.right_val.expr.(string)
	case float64:
		v.expr = left + tonumf(s.right_val.expr)
	case int64:
		v.expr = int64(left + tonums(s.right_val.expr))
	case uint64:
		v.expr = uint64(left + tonumu(s.right_val.expr))
	}
}

func (s *solver) sub(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case float64:
		v.expr = left - tonumf(s.right_val.expr)
	case int64:
		v.expr = int64(left - tonums(s.right_val.expr))
	case uint64:
		v.expr = uint64(left - tonumu(s.right_val.expr))
	}
}

func (s *solver) mul(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case float64:
		v.expr = left * tonumf(s.right_val.expr)
	case int64:
		v.expr = int64(left * tonums(s.right_val.expr))
	case uint64:
		v.expr = uint64(left * tonumu(s.right_val.expr))
	}
}

func (s *solver) div(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case float64:
		right := tonumf(s.right_val.expr)
		if right != 0 {
			v.expr = left / right
		} else {
			s.p.pusherrtok(s.operator, "divide_by_zero")
			v.expr = float64(0)
		}
	case int64:
		right := tonumf(s.right_val.expr)
		if right != 0 {
			v.expr = float64(left) / right
		} else {
			s.p.pusherrtok(s.operator, "divide_by_zero")
			v.expr = int64(0)
		}
	case uint64:
		right := tonumf(s.right_val.expr)
		if right != 0 {
			v.expr = float64(left) / right
		} else {
			s.p.pusherrtok(s.operator, "divide_by_zero")
			v.expr = float64(0)
		}
	}
}

func (s *solver) mod(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case int64:
		right := tonums(s.right_val.expr)
		if right != 0 {
			v.expr = left % right
		} else {
			s.p.pusherrtok(s.operator, "divide_by_zero")
			v.expr = int64(0)
		}
	case uint64:
		right := tonumu(s.right_val.expr)
		if right != 0 {
			v.expr = left % right
		} else {
			s.p.pusherrtok(s.operator, "divide_by_zero")
			v.expr = uint64(0)
		}
	}
}

func (s *solver) bitwiseAnd(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case int64:
		v.expr = left & tonums(s.right_val.expr)
	case uint64:
		v.expr = left & tonumu(s.right_val.expr)
	}
}

func (s *solver) bitwiseOr(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case int64:
		v.expr = left | tonums(s.right_val.expr)
	case uint64:
		v.expr = left | tonumu(s.right_val.expr)
	}
}

func (s *solver) bitwiseXor(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case int64:
		v.expr = left ^ tonums(s.right_val.expr)
	case uint64:
		v.expr = left ^ tonumu(s.right_val.expr)
	}
}

func (s *solver) urshift(v *value) {
	left := tonumu(s.left_val.expr)
	right := tonumu(s.right_val.expr)
	v.expr = left >> right
	setshift(v, right)
}

func (s *solver) rshift(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case int64:
		if left < 0 {
			right := tonumu(s.right_val.expr)
			v.expr = left >> right
			setshift(v, right)
		} else {
			s.urshift(v)
		}
	case uint64:
		s.urshift(v)
	}
}

func (s *solver) ulshift(v *value) {
	left := tonumu(s.left_val.expr)
	right := tonumu(s.right_val.expr)
	v.expr = left << right
	setshift(v, right)
}

func (s *solver) lshift(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case int64:
		if left < 0 {
			right := tonumu(s.right_val.expr)
			v.expr = left << right
			setshift(v, right)
		} else {
			s.ulshift(v)
		}
	case uint64:
		s.ulshift(v)
	}
}

func (s *solver) and(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case bool:
		v.expr = left && s.right_val.expr.(bool)
	}
}

func (s *solver) or(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.left_val.expr.(type) {
	case bool:
		v.expr = left || s.right_val.expr.(bool)
	}
}

func (s *solver) ptr() (v value) {
	v.data.Token = s.operator
	if !typesAreCompatible(s.left_val.data.Type, s.right_val.data.Type, true) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "incompatible_types",
			s.right_val.data.Type.Kind, s.left_val.data.Type.Kind)
		return
	}
	if !typeIsPtr(s.left_val.data.Type) {
		s.left_val, s.right_val = s.right_val, s.left_val
	}
	switch s.operator.Kind {
	case tokens.PLUS, tokens.MINUS:
		v.data.Type = s.left_val.data.Type
	case tokens.EQUALS, tokens.NOT_EQUALS, tokens.LESS, tokens.GREAT,
		tokens.GREAT_EQUAL, tokens.LESS_EQUAL:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "operator_not_for_juletype", s.operator.Kind, "pointer")
	}
	return
}

func (s *solver) enum() (v value) {
	if typeIsEnum(s.left_val.data.Type) {
		s.left_val.data.Type = s.left_val.data.Type.Tag.(*Enum).Type
	}
	if typeIsEnum(s.right_val.data.Type) {
		s.right_val.data.Type = s.right_val.data.Type.Tag.(*Enum).Type
	}
	return s.solve()
}

func (s *solver) str() (v value) {
	v.data.Token = s.operator
	// Not both string?
	if s.left_val.data.Type.Id != s.right_val.data.Type.Id {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "incompatible_types",
			s.left_val.data.Type.Kind, s.right_val.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.PLUS:
		v.data.Type.Id = juletype.Str
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.add(&v)
	case tokens.EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.eq(&v)
	case tokens.NOT_EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.noteq(&v)
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "operator_not_for_juletype",
			s.operator.Kind, tokens.STR)
	}
	return
}

func (s *solver) any() (v value) {
	v.data.Token = s.operator
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "operator_not_for_juletype", s.operator.Kind, tokens.ANY)
	}
	return
}

func (s *solver) bool() (v value) {
	v.data.Token = s.operator
	if !typesAreCompatible(s.left_val.data.Type, s.right_val.data.Type, true) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "incompatible_types",
			s.right_val.data.Type.Kind, s.left_val.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.eq(&v)
	case tokens.NOT_EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.noteq(&v)
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "operator_not_for_juletype",
			s.operator.Kind, tokens.BOOL)
	}
	return
}

func (s *solver) floatMod() (v value, ok bool) {
	if !juletype.IsInteger(s.left_val.data.Type.Id) {
		if !juletype.IsInteger(s.right_val.data.Type.Id) {
			return
		}
		s.left_val, s.right_val = s.right_val, s.left_val
	}
	switch {
	case juletype.IsSignedInteger(s.left_val.data.Type.Id):
		switch {
		case integerAssignable(juletype.I64, s.right_val):
			return s.signed(), true
		case integerAssignable(juletype.U64, s.right_val):
			return s.unsigned(), true
		}
	case juletype.IsUnsignedInteger(s.left_val.data.Type.Id):
		if integerAssignable(juletype.I64, s.right_val) ||
			integerAssignable(juletype.U64, s.right_val) {
			return s.unsigned(), true
		}
	}
	return
}

func (s *solver) float() (v value) {
	v.data.Token = s.operator
	if !juletype.IsNumeric(s.left_val.data.Type.Id) ||
		!juletype.IsNumeric(s.right_val.data.Type.Id) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "incompatible_types",
			s.right_val.data.Type.Kind, s.left_val.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.eq(&v)
	case tokens.NOT_EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.noteq(&v)
	case tokens.LESS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.lt(&v)
	case tokens.GREAT:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.gt(&v)
	case tokens.GREAT_EQUAL:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.gteq(&v)
	case tokens.LESS_EQUAL:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.lteq(&v)
	case tokens.PLUS:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.add(&v)
	case tokens.MINUS:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.sub(&v)
	case tokens.STAR:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.mul(&v)
	case tokens.SOLIDUS:
		// Ignore float if expression has integer
		if juletype.IsInteger(s.left_val.data.Type.Id) && juletype.IsInteger(s.right_val.data.Type.Id) {
		} else if juletype.IsInteger(s.left_val.data.Type.Id) {
			s.right_val.data.Type = s.left_val.data.Type
		} else if juletype.IsInteger(s.right_val.data.Type.Id) {
			s.left_val.data.Type = s.right_val.data.Type
		}
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.div(&v)
	case tokens.PERCENT:
		var ok bool
		v, ok = s.floatMod()
		if ok {
			break
		}
		fallthrough
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "operator_not_for_float", s.operator.Kind)
	}
	return
}

func (s *solver) signed() (v value) {
	v.data.Token = s.operator
	if !juletype.IsNumeric(s.left_val.data.Type.Id) ||
		!juletype.IsNumeric(s.right_val.data.Type.Id) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "incompatible_types",
			s.right_val.data.Type.Kind, s.left_val.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.eq(&v)
	case tokens.NOT_EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.noteq(&v)
	case tokens.LESS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.lt(&v)
	case tokens.GREAT:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.gt(&v)
	case tokens.GREAT_EQUAL:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.gteq(&v)
	case tokens.LESS_EQUAL:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.lteq(&v)
	case tokens.PLUS:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.add(&v)
	case tokens.MINUS:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.sub(&v)
	case tokens.STAR:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.mul(&v)
	case tokens.SOLIDUS:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.div(&v)
	case tokens.PERCENT:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.mod(&v)
	case tokens.AMPER:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.bitwiseAnd(&v)
	case tokens.VLINE:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.bitwiseOr(&v)
	case tokens.CARET:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.bitwiseXor(&v)
	case tokens.RSHIFT:
		v.data.Type.Id = juletype.U64
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		if !okForShifting(s.right_val) {
			s.p.pusherrtok(s.operator, "bitshift_must_unsigned")
		}
		s.rshift(&v)
	case tokens.LSHIFT:
		v.data.Type.Id = juletype.U64
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		if !okForShifting(s.right_val) {
			s.p.pusherrtok(s.operator, "bitshift_must_unsigned")
		}
		s.lshift(&v)
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "operator_not_for_int", s.operator.Kind)
	}
	return
}

func (s *solver) unsigned() (v value) {
	v.data.Token = s.operator
	if !juletype.IsNumeric(s.left_val.data.Type.Id) ||
		!juletype.IsNumeric(s.right_val.data.Type.Id) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "incompatible_types",
			s.right_val.data.Type.Kind, s.left_val.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.eq(&v)
	case tokens.NOT_EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.noteq(&v)
	case tokens.LESS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.lt(&v)
	case tokens.GREAT:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.gt(&v)
	case tokens.GREAT_EQUAL:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.gteq(&v)
	case tokens.LESS_EQUAL:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		s.lteq(&v)
	case tokens.PLUS:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.add(&v)
	case tokens.MINUS:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.sub(&v)
	case tokens.STAR:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.mul(&v)
	case tokens.SOLIDUS:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.div(&v)
	case tokens.PERCENT:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.mod(&v)
	case tokens.AMPER:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.bitwiseAnd(&v)
	case tokens.VLINE:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.bitwiseOr(&v)
	case tokens.CARET:
		v.data.Type = s.left_val.data.Type
		if juletype.TypeGreaterThan(s.right_val.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.right_val.data.Type
		}
		s.bitwiseXor(&v)
	case tokens.RSHIFT:
		v.data.Type.Id = juletype.U64
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		if !okForShifting(s.right_val) {
			s.p.pusherrtok(s.operator, "bitshift_must_unsigned")
		}
		s.rshift(&v)
	case tokens.LSHIFT:
		v.data.Type.Id = juletype.U64
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		if !okForShifting(s.right_val) {
			s.p.pusherrtok(s.operator, "bitshift_must_unsigned")
		}
		s.lshift(&v)
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "operator_not_for_uint", s.operator.Kind)
	}
	return
}

func (s *solver) logical() (v value) {
	if s.left_val.data.Type.Id != juletype.Bool ||
		s.right_val.data.Type.Id != juletype.Bool {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "logical_not_bool")
		return
	}
	v.data.Token = s.operator
	v.data.Type.Id = juletype.Bool
	v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	if !s.isConstExpr() {
		return
	}
	switch s.operator.Kind {
	case tokens.DOUBLE_AMPER:
		s.and(&v)
	case tokens.DOUBLE_VLINE:
		s.or(&v)
	}
	return
}

func (s *solver) array() (v value) {
	v.data.Token = s.operator
	if !typesAreCompatible(s.left_val.data.Type, s.right_val.data.Type, true) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "incompatible_types",
			s.right_val.data.Type.Kind, s.left_val.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "operator_not_for_juletype", s.operator.Kind, s.left_val.data.Type.Kind)
	}
	return
}

func (s *solver) slice() (v value) {
	v.data.Token = s.operator
	if !typesAreCompatible(s.left_val.data.Type, s.right_val.data.Type, true) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "incompatible_types",
			s.right_val.data.Type.Kind, s.left_val.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "operator_not_for_juletype",
			s.operator.Kind, s.left_val.data.Type.Kind)
	}
	return
}

func (s *solver) nil() (v value) {
	v.data.Token = s.operator
	if !typesAreCompatible(s.left_val.data.Type, s.right_val.data.Type, false) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "incompatible_types",
			s.right_val.data.Type.Kind, s.left_val.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.NOT_EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		if s.isConstExpr() {
			v.expr = s.left_val.expr != nil && s.right_val.expr != nil
		}
	case tokens.EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		if s.isConstExpr() {
			v.expr = s.left_val.expr == nil && s.right_val.expr == nil
		}
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "operator_not_for_juletype",
			s.operator.Kind, tokens.NIL)
	}
	return
}

func (s *solver) structure() (v value) {
	v.data.Token = s.operator
	if s.left_val.data.Type.Kind != s.right_val.data.Type.Kind {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "incompatible_types",
			s.right_val.data.Type.Kind, s.left_val.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.NOT_EQUALS, tokens.EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "operator_not_for_juletype",
			s.operator.Kind, tokens.STRUCT)
	}
	return
}

func (s *solver) juletrait() (v value) {
	v.data.Token = s.operator
	if !typesAreCompatible(s.left_val.data.Type, s.right_val.data.Type, true) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "incompatible_types",
			s.right_val.data.Type.Kind, s.left_val.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.NOT_EQUALS, tokens.EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "operator_not_for_juletype",
			s.operator.Kind, tokens.TRAIT)
	}
	return
}

func (s *solver) function() (v value) {
	v.data.Token = s.operator
	if (!typeIsPure(s.left_val.data.Type) || s.left_val.data.Type.Id != juletype.Nil) &&
		(!typeIsPure(s.right_val.data.Type) || s.right_val.data.Type.Id != juletype.Nil) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "incompatible_types",
			s.right_val.data.Type.Kind, s.left_val.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.NOT_EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	case tokens.EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.operator, "operator_not_for_juletype",
			s.operator.Kind, tokens.NIL)
	}
	return
}

func (s *solver) isConstExpr() bool {
	return s.left_val.constExpr && s.right_val.constExpr
}

func (s *solver) solve() (v value) {
	defer func() {
		if typeIsVoid(v.data.Type) {
			v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		} else {
			v.constExpr = s.isConstExpr()
			if v.constExpr {
				bitize(&v)
				v.model = getModel(v)
			}
		}
	}()
	switch s.operator.Kind {
	case tokens.DOUBLE_AMPER, tokens.DOUBLE_VLINE:
		return s.logical()
	}
	switch {
	case typeIsFunc(s.left_val.data.Type), typeIsFunc(s.right_val.data.Type):
		return s.function()
	case typeIsArray(s.left_val.data.Type), typeIsArray(s.right_val.data.Type):
		return s.array()
	case typeIsSlice(s.left_val.data.Type), typeIsSlice(s.right_val.data.Type):
		return s.slice()
	case typeIsPtr(s.left_val.data.Type), typeIsPtr(s.right_val.data.Type):
		return s.ptr()
	case typeIsEnum(s.left_val.data.Type), typeIsEnum(s.right_val.data.Type):
		return s.enum()
	case typeIsStruct(s.left_val.data.Type), typeIsStruct(s.right_val.data.Type):
		return s.structure()
	case typeIsTrait(s.left_val.data.Type), typeIsTrait(s.right_val.data.Type):
		return s.juletrait()
	case s.left_val.data.Type.Id == juletype.Nil, s.right_val.data.Type.Id == juletype.Nil:
		return s.nil()
	case s.left_val.data.Type.Id == juletype.Any, s.right_val.data.Type.Id == juletype.Any:
		return s.any()
	case s.left_val.data.Type.Id == juletype.Bool, s.right_val.data.Type.Id == juletype.Bool:
		return s.bool()
	case s.left_val.data.Type.Id == juletype.Str, s.right_val.data.Type.Id == juletype.Str:
		return s.str()
	case juletype.IsFloat(s.left_val.data.Type.Id),
		juletype.IsFloat(s.right_val.data.Type.Id):
		return s.float()
	case juletype.IsUnsignedInteger(s.left_val.data.Type.Id),
		juletype.IsUnsignedInteger(s.right_val.data.Type.Id):
		return s.unsigned()
	case juletype.IsSignedNumeric(s.left_val.data.Type.Id),
		juletype.IsSignedNumeric(s.right_val.data.Type.Id):
		return s.signed()
	}
	return
}
