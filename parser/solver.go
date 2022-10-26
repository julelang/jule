package parser

import (
	"github.com/jule-lang/jule/lex"
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
	v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
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
	p  *Parser
	l  value
	r  value
	op lex.Token
}

func (s *solver) eq(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case bool:
		v.expr = left == s.r.expr.(bool)
	case string:
		v.expr = left == s.r.expr.(string)
	case float64:
		v.expr = left == tonumf(s.r.expr)
	case int64:
		v.expr = left == tonums(s.r.expr)
	case uint64:
		v.expr = left == tonumu(s.r.expr)
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
	switch left := s.l.expr.(type) {
	case float64:
		v.expr = left < tonumf(s.r.expr)
	case int64:
		v.expr = left < tonums(s.r.expr)
	case uint64:
		v.expr = left < tonumu(s.r.expr)
	}
}

func (s *solver) gt(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case float64:
		v.expr = left > tonumf(s.r.expr)
	case int64:
		v.expr = left > tonums(s.r.expr)
	case uint64:
		v.expr = left > tonumu(s.r.expr)
	}
}

func (s *solver) lteq(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case float64:
		v.expr = left <= tonumf(s.r.expr)
	case int64:
		v.expr = left <= tonums(s.r.expr)
	case uint64:
		v.expr = left <= tonumu(s.r.expr)
	}
}

func (s *solver) gteq(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case float64:
		v.expr = left >= tonumf(s.r.expr)
	case int64:
		v.expr = left >= tonums(s.r.expr)
	case uint64:
		v.expr = left >= tonumu(s.r.expr)
	}
}

func (s *solver) add(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case string:
		v.expr = left + s.r.expr.(string)
	case float64:
		v.expr = left + tonumf(s.r.expr)
	case int64:
		v.expr = int64(left + tonums(s.r.expr))
	case uint64:
		v.expr = uint64(left + tonumu(s.r.expr))
	}
}

func (s *solver) sub(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case float64:
		v.expr = left - tonumf(s.r.expr)
	case int64:
		v.expr = int64(left - tonums(s.r.expr))
	case uint64:
		v.expr = uint64(left - tonumu(s.r.expr))
	}
}

func (s *solver) mul(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case float64:
		v.expr = left * tonumf(s.r.expr)
	case int64:
		v.expr = int64(left * tonums(s.r.expr))
	case uint64:
		v.expr = uint64(left * tonumu(s.r.expr))
	}
}

func (s *solver) div(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case float64:
		right := tonumf(s.r.expr)
		if right != 0 {
			v.expr = left / right
		} else {
			s.p.pusherrtok(s.op, "divide_by_zero")
			v.expr = float64(0)
		}
	case int64:
		right := tonumf(s.r.expr)
		if right != 0 {
			v.expr = float64(left) / right
		} else {
			s.p.pusherrtok(s.op, "divide_by_zero")
			v.expr = int64(0)
		}
	case uint64:
		right := tonumf(s.r.expr)
		if right != 0 {
			v.expr = float64(left) / right
		} else {
			s.p.pusherrtok(s.op, "divide_by_zero")
			v.expr = float64(0)
		}
	}
}

func (s *solver) mod(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case int64:
		right := tonums(s.r.expr)
		if right != 0 {
			v.expr = left % right
		} else {
			s.p.pusherrtok(s.op, "divide_by_zero")
			v.expr = int64(0)
		}
	case uint64:
		right := tonumu(s.r.expr)
		if right != 0 {
			v.expr = left % right
		} else {
			s.p.pusherrtok(s.op, "divide_by_zero")
			v.expr = uint64(0)
		}
	}
}

func (s *solver) bitwiseAnd(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case int64:
		v.expr = left & tonums(s.r.expr)
	case uint64:
		v.expr = left & tonumu(s.r.expr)
	}
}

func (s *solver) bitwiseOr(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case int64:
		v.expr = left | tonums(s.r.expr)
	case uint64:
		v.expr = left | tonumu(s.r.expr)
	}
}

func (s *solver) bitwiseXor(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case int64:
		v.expr = left ^ tonums(s.r.expr)
	case uint64:
		v.expr = left ^ tonumu(s.r.expr)
	}
}

func (s *solver) urshift(v *value) {
	left := tonumu(s.l.expr)
	right := tonumu(s.r.expr)
	v.expr = left >> right
	setshift(v, right)
}

func (s *solver) rshift(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case int64:
		if left < 0 {
			right := tonumu(s.r.expr)
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
	left := tonumu(s.l.expr)
	right := tonumu(s.r.expr)
	v.expr = left << right
	setshift(v, right)
}

func (s *solver) lshift(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case int64:
		if left < 0 {
			right := tonumu(s.r.expr)
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
	switch left := s.l.expr.(type) {
	case bool:
		v.expr = left && s.r.expr.(bool)
	}
}

func (s *solver) or(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.l.expr.(type) {
	case bool:
		v.expr = left || s.r.expr.(bool)
	}
}

func (s *solver) ptr() (v value) {
	v.data.Token = s.op
	if !s.types_are_compatible(true) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.Type.Kind, s.l.data.Type.Kind)
		return
	}
	if !type_is_ptr(s.l.data.Type) {
		s.l, s.r = s.r, s.l
	}
	switch s.op.Kind {
	case lex.KND_PLUS, lex.KND_MINUS:
		v.data.Type = s.l.data.Type
	case lex.KND_EQS, lex.KND_NOT_EQ, lex.KND_LT, lex.KND_GT,
		lex.KND_GREAT_EQ, lex.KND_LESS_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, "pointer")
	}
	return
}

func (s *solver) enum() (v value) {
	if type_is_enum(s.l.data.Type) {
		s.l.data.Type = s.l.data.Type.Tag.(*Enum).Type
	}
	if type_is_enum(s.r.data.Type) {
		s.r.data.Type = s.r.data.Type.Tag.(*Enum).Type
	}
	return s.solve()
}

func (s *solver) str() (v value) {
	v.data.Token = s.op
	// Not both string?
	if s.l.data.Type.Id != s.r.data.Type.Id {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.l.data.Type.Kind, s.r.data.Type.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_PLUS:
		v.data.Type.Id = juletype.STR
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.add(&v)
	case lex.KND_EQS:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.eq(&v)
	case lex.KND_NOT_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.noteq(&v)
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, lex.KND_STR)
	}
	return
}

func (s *solver) any() (v value) {
	v.data.Token = s.op
	switch s.op.Kind {
	case lex.KND_EQS, lex.KND_NOT_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, lex.KND_ANY)
	}
	return
}

func (s *solver) bool() (v value) {
	v.data.Token = s.op
	if !s.types_are_compatible(true) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.Type.Kind, s.l.data.Type.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_EQS:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.eq(&v)
	case lex.KND_NOT_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.noteq(&v)
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, lex.KND_BOOL)
	}
	return
}

func (s *solver) floatMod() (v value, ok bool) {
	if !juletype.IsInteger(s.l.data.Type.Id) {
		if !juletype.IsInteger(s.r.data.Type.Id) {
			return
		}
		s.l, s.r = s.r, s.l
	}
	switch {
	case juletype.IsSignedInteger(s.l.data.Type.Id):
		switch {
		case int_assignable(juletype.I64, s.r):
			return s.signed(), true
		case int_assignable(juletype.U64, s.r):
			return s.unsigned(), true
		}
	case juletype.IsUnsignedInteger(s.l.data.Type.Id):
		if int_assignable(juletype.I64, s.r) ||
			int_assignable(juletype.U64, s.r) {
			return s.unsigned(), true
		}
	}
	return
}

func (s *solver) float() (v value) {
	v.data.Token = s.op
	if !juletype.IsNumeric(s.l.data.Type.Id) ||
		!juletype.IsNumeric(s.r.data.Type.Id) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.Type.Kind, s.l.data.Type.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_EQS:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.eq(&v)
	case lex.KND_NOT_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.noteq(&v)
	case lex.KND_LT:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.lt(&v)
	case lex.KND_GT:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.gt(&v)
	case lex.KND_GREAT_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.gteq(&v)
	case lex.KND_LESS_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.lteq(&v)
	case lex.KND_PLUS:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.add(&v)
	case lex.KND_MINUS:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.sub(&v)
	case lex.KND_STAR:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.mul(&v)
	case lex.KND_SOLIDUS:
		// Ignore float if expression has integer
		if juletype.IsInteger(s.l.data.Type.Id) && juletype.IsInteger(s.r.data.Type.Id) {
		} else if juletype.IsInteger(s.l.data.Type.Id) {
			s.r.data.Type = s.l.data.Type
		} else if juletype.IsInteger(s.r.data.Type.Id) {
			s.l.data.Type = s.r.data.Type
		}
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.div(&v)
	case lex.KND_PERCENT:
		var ok bool
		v, ok = s.floatMod()
		if ok {
			break
		}
		fallthrough
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_float", s.op.Kind)
	}
	return
}

func (s *solver) signed() (v value) {
	v.data.Token = s.op
	if !juletype.IsNumeric(s.l.data.Type.Id) ||
		!juletype.IsNumeric(s.r.data.Type.Id) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.Type.Kind, s.l.data.Type.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_EQS:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.eq(&v)
	case lex.KND_NOT_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.noteq(&v)
	case lex.KND_LT:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.lt(&v)
	case lex.KND_GT:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.gt(&v)
	case lex.KND_GREAT_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.gteq(&v)
	case lex.KND_LESS_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.lteq(&v)
	case lex.KND_PLUS:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.add(&v)
	case lex.KND_MINUS:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.sub(&v)
	case lex.KND_STAR:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.mul(&v)
	case lex.KND_SOLIDUS:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.div(&v)
	case lex.KND_PERCENT:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.mod(&v)
	case lex.KND_AMPER:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.bitwiseAnd(&v)
	case lex.KND_VLINE:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.bitwiseOr(&v)
	case lex.KND_CARET:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.bitwiseXor(&v)
	case lex.KND_RSHIFT:
		v.data.Type.Id = juletype.U64
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		if !okForShifting(s.r) {
			s.p.pusherrtok(s.op, "bitshift_must_unsigned")
		}
		s.rshift(&v)
	case lex.KND_LSHIFT:
		v.data.Type.Id = juletype.U64
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		if !okForShifting(s.r) {
			s.p.pusherrtok(s.op, "bitshift_must_unsigned")
		}
		s.lshift(&v)
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_int", s.op.Kind)
	}
	return
}

func (s *solver) unsigned() (v value) {
	v.data.Token = s.op
	if !juletype.IsNumeric(s.l.data.Type.Id) ||
		!juletype.IsNumeric(s.r.data.Type.Id) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.Type.Kind, s.l.data.Type.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_EQS:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.eq(&v)
	case lex.KND_NOT_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.noteq(&v)
	case lex.KND_LT:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.lt(&v)
	case lex.KND_GT:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.gt(&v)
	case lex.KND_GREAT_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.gteq(&v)
	case lex.KND_LESS_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		s.lteq(&v)
	case lex.KND_PLUS:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.add(&v)
	case lex.KND_MINUS:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.sub(&v)
	case lex.KND_STAR:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.mul(&v)
	case lex.KND_SOLIDUS:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.div(&v)
	case lex.KND_PERCENT:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.mod(&v)
	case lex.KND_AMPER:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.bitwiseAnd(&v)
	case lex.KND_VLINE:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.bitwiseOr(&v)
	case lex.KND_CARET:
		v.data.Type = s.l.data.Type
		if juletype.TypeGreaterThan(s.r.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.r.data.Type
		}
		s.bitwiseXor(&v)
	case lex.KND_RSHIFT:
		v.data.Type.Id = juletype.U64
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		if !okForShifting(s.r) {
			s.p.pusherrtok(s.op, "bitshift_must_unsigned")
		}
		s.rshift(&v)
	case lex.KND_LSHIFT:
		v.data.Type.Id = juletype.U64
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		if !okForShifting(s.r) {
			s.p.pusherrtok(s.op, "bitshift_must_unsigned")
		}
		s.lshift(&v)
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_uint", s.op.Kind)
	}
	return
}

func (s *solver) logical() (v value) {
	if s.l.data.Type.Id != juletype.BOOL ||
		s.r.data.Type.Id != juletype.BOOL {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "logical_not_bool")
		return
	}
	v.data.Token = s.op
	v.data.Type.Id = juletype.BOOL
	v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	if !s.isConstExpr() {
		return
	}
	switch s.op.Kind {
	case lex.KND_DBL_AMPER:
		s.and(&v)
	case lex.KND_DBL_VLINE:
		s.or(&v)
	}
	return
}

func (s *solver) array() (v value) {
	v.data.Token = s.op
	if !s.types_are_compatible(true) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.Type.Kind, s.l.data.Type.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_EQS, lex.KND_NOT_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, s.l.data.Type.Kind)
	}
	return
}

func (s *solver) slice() (v value) {
	v.data.Token = s.op
	if !s.types_are_compatible(true) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.Type.Kind, s.l.data.Type.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_EQS, lex.KND_NOT_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype",
			s.op.Kind, s.l.data.Type.Kind)
	}
	return
}

func (s *solver) nil() (v value) {
	v.data.Token = s.op
	if !s.types_are_compatible(false) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.Type.Kind, s.l.data.Type.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_NOT_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		if s.isConstExpr() {
			v.expr = s.l.expr != nil && s.r.expr != nil
		}
	case lex.KND_EQS:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
		if s.isConstExpr() {
			v.expr = s.l.expr == nil && s.r.expr == nil
		}
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, lex.KND_NIL)
	}
	return
}

func (s *solver) structure() (v value) {
	v.data.Token = s.op
	if s.l.data.Type.Kind != s.r.data.Type.Kind {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.Type.Kind, s.l.data.Type.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_NOT_EQ, lex.KND_EQS:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, lex.KND_STRUCT)
	}
	return
}

func (s *solver) juletrait() (v value) {
	v.data.Token = s.op
	if !s.types_are_compatible(true) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.Type.Kind, s.l.data.Type.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_NOT_EQ, lex.KND_EQS:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, lex.KND_TRAIT)
	}
	return
}

func (s *solver) function() (v value) {
	v.data.Token = s.op
	if (!type_is_pure(s.l.data.Type) || s.l.data.Type.Id != juletype.NIL) &&
		(!type_is_pure(s.r.data.Type) || s.r.data.Type.Id != juletype.NIL) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.Type.Kind, s.l.data.Type.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_NOT_EQ:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	case lex.KND_EQS:
		v.data.Type.Id = juletype.BOOL
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, lex.KND_NIL)
	}
	return
}

func (s *solver) types_are_compatible(ignore_any bool) bool {
	checker := type_checker{
		p:            s.p,
		l:            s.l.data.Type,
		r:            s.r.data.Type,
		ignore_any:   ignore_any,
		errtok:       s.op,
		allow_assign: true,
	}
	ok := checker.check()
	return ok
}

func (s *solver) isConstExpr() bool { return s.l.constExpr && s.r.constExpr }

func (s *solver) finalize(v *value) {
	if type_is_void(v.data.Type) {
		v.data.Type.Kind = juletype.TYPE_MAP[v.data.Type.Id]
	} else {
		v.constExpr = s.isConstExpr()
		if v.constExpr {
			bitize(v)
			v.model = getModel(*v)
		}
	}
}

func (s *solver) solve() (v value) {
	switch {
	case s.op.Kind == lex.KND_DBL_AMPER || s.op.Kind == lex.KND_DBL_VLINE:
		v = s.logical()
	case type_is_fn(s.l.data.Type) || type_is_fn(s.r.data.Type):
		v = s.function()
	case type_is_array(s.l.data.Type) || type_is_array(s.r.data.Type):
		v = s.array()
	case type_is_slc(s.l.data.Type) || type_is_slc(s.r.data.Type):
		v = s.slice()
	case type_is_ptr(s.l.data.Type) || type_is_ptr(s.r.data.Type):
		v = s.ptr()
	case type_is_enum(s.l.data.Type) || type_is_enum(s.r.data.Type):
		v = s.enum()
	case type_is_struct(s.l.data.Type) || type_is_struct(s.r.data.Type):
		v = s.structure()
	case type_is_trait(s.l.data.Type) || type_is_trait(s.r.data.Type):
		v = s.juletrait()
	case s.l.data.Type.Id == juletype.NIL || s.r.data.Type.Id == juletype.NIL:
		v = s.nil()
	case s.l.data.Type.Id == juletype.ANY || s.r.data.Type.Id == juletype.ANY:
		v = s.any()
	case s.l.data.Type.Id == juletype.BOOL || s.r.data.Type.Id == juletype.BOOL:
		v = s.bool()
	case s.l.data.Type.Id == juletype.STR || s.r.data.Type.Id == juletype.STR:
		v = s.str()
	case juletype.IsFloat(s.l.data.Type.Id) || juletype.IsFloat(s.r.data.Type.Id):
		v = s.float()
	case juletype.IsUnsignedInteger(s.l.data.Type.Id) ||
		juletype.IsUnsignedInteger(s.r.data.Type.Id):
		v = s.unsigned()
	case juletype.IsSignedNumeric(s.l.data.Type.Id) ||
		juletype.IsSignedNumeric(s.r.data.Type.Id):
		v = s.signed()
	}
	s.finalize(&v)
	return
}
