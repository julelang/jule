package analysis

import (
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

func assign_shift(v *value, right uint64) {
	switch {
	case right <= 6:
		v.data.DataType.Id = types.I8
	case right <= 7:
		v.data.DataType.Id = types.U8
	case right <= 14:
		v.data.DataType.Id = types.I16
	case right <= 15:
		v.data.DataType.Id = types.U16
	case right <= 30:
		v.data.DataType.Id = types.I32
	case right <= 31:
		v.data.DataType.Id = types.U32
	case right <= 62:
		v.data.DataType.Id = types.I64
	case right <= 63:
		v.data.DataType.Id = types.U64
	case right <= 127:
		v.data.DataType.Id = types.F32
	default:
		v.data.DataType.Id = types.F64
	}
}

func normalize_bitsize(v *value) {
	id := types.VOID
	switch t := v.expr.(type) {
	case float64:
		id = types.FloatFromBits(types.BitsizeFloat(t))
	case int64:
		id = types.IntFromBits(types.BitsizeInt(t))
	case uint64:
		id = types.UIntFromBits(types.BitsizeUInt(t))
	default:
		return
	}

	// Save rune types.
	if !lex.IsRune(v.data.Value) {
		v.data.DataType.Id = id
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	}
}

func to_num_float(expr any) float64 {
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

func to_num_unsigned(expr any) uint64 {
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

func to_num_signed(expr any) int64 {
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
	p  *Analyzer
	l  value
	r  value
	op lex.Token
}

func (s *solver) value_is_rune() bool {
	if !lex.IsRune(s.l.data.Value) || !lex.IsRune(s.r.data.Value) {
		return false
	}
	return true
}

func (s *solver) eq(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case bool:
		v.expr = left == s.r.expr.(bool)
	case string:
		v.expr = left == s.r.expr.(string)
	case float64:
		v.expr = left == to_num_float(s.r.expr)
	case int64:
		v.expr = left == to_num_signed(s.r.expr)
	case uint64:
		v.expr = left == to_num_unsigned(s.r.expr)
	}
}

func (s *solver) not_eq(v *value) {
	if !s.is_const_expr() {
		return
	}
	s.eq(v)
	v.expr = !v.expr.(bool)
}

func (s *solver) lt(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case float64:
		v.expr = left < to_num_float(s.r.expr)
	case int64:
		v.expr = left < to_num_signed(s.r.expr)
	case uint64:
		v.expr = left < to_num_unsigned(s.r.expr)
	}
}

func (s *solver) gt(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case float64:
		v.expr = left > to_num_float(s.r.expr)
	case int64:
		v.expr = left > to_num_signed(s.r.expr)
	case uint64:
		v.expr = left > to_num_unsigned(s.r.expr)
	}
}

func (s *solver) lt_eq(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case float64:
		v.expr = left <= to_num_float(s.r.expr)
	case int64:
		v.expr = left <= to_num_signed(s.r.expr)
	case uint64:
		v.expr = left <= to_num_unsigned(s.r.expr)
	}
}

func (s *solver) gt_eq(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case float64:
		v.expr = left >= to_num_float(s.r.expr)
	case int64:
		v.expr = left >= to_num_signed(s.r.expr)
	case uint64:
		v.expr = left >= to_num_unsigned(s.r.expr)
	}
}

func (s *solver) add(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case string:
		v.expr = left + s.r.expr.(string)
	case float64:
		v.expr = left + to_num_float(s.r.expr)
	case int64:
		v.expr = int64(left + to_num_signed(s.r.expr))
	case uint64:
		v.expr = uint64(left + to_num_unsigned(s.r.expr))
	}
}

func (s *solver) sub(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case float64:
		v.expr = left - to_num_float(s.r.expr)
	case int64:
		v.expr = int64(left - to_num_signed(s.r.expr))
	case uint64:
		v.expr = uint64(left - to_num_unsigned(s.r.expr))
	}
}

func (s *solver) mul(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case float64:
		v.expr = left * to_num_float(s.r.expr)
	case int64:
		v.expr = int64(left * to_num_signed(s.r.expr))
	case uint64:
		v.expr = uint64(left * to_num_unsigned(s.r.expr))
	}
}

func (s *solver) div(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case float64:
		right := to_num_float(s.r.expr)
		if right != 0 {
			v.expr = left / right
		} else {
			s.p.pusherrtok(s.op, "divide_by_zero")
			v.expr = float64(0)
		}
	case int64:
		right := to_num_float(s.r.expr)
		if right != 0 {
			v.expr = float64(left) / right
		} else {
			s.p.pusherrtok(s.op, "divide_by_zero")
			v.expr = int64(0)
		}
	case uint64:
		right := to_num_float(s.r.expr)
		if right != 0 {
			v.expr = float64(left) / right
		} else {
			s.p.pusherrtok(s.op, "divide_by_zero")
			v.expr = float64(0)
		}
	}
}

func (s *solver) mod(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case int64:
		right := to_num_signed(s.r.expr)
		if right != 0 {
			v.expr = left % right
		} else {
			s.p.pusherrtok(s.op, "divide_by_zero")
			v.expr = int64(0)
		}
	case uint64:
		right := to_num_unsigned(s.r.expr)
		if right != 0 {
			v.expr = left % right
		} else {
			s.p.pusherrtok(s.op, "divide_by_zero")
			v.expr = uint64(0)
		}
	}
}

func (s *solver) bitwise_and(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case int64:
		v.expr = left & to_num_signed(s.r.expr)
	case uint64:
		v.expr = left & to_num_unsigned(s.r.expr)
	}
}

func (s *solver) bitwise_or(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case int64:
		v.expr = left | to_num_signed(s.r.expr)
	case uint64:
		v.expr = left | to_num_unsigned(s.r.expr)
	}
}

func (s *solver) bitwise_xor(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case int64:
		v.expr = left ^ to_num_signed(s.r.expr)
	case uint64:
		v.expr = left ^ to_num_unsigned(s.r.expr)
	}
}

func (s *solver) right_shift_unsigned(v *value) {
	left := to_num_unsigned(s.l.expr)
	right := to_num_unsigned(s.r.expr)
	v.expr = left >> right
	assign_shift(v, right)
}

func (s *solver) right_shift(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case int64:
		if left < 0 {
			right := to_num_unsigned(s.r.expr)
			v.expr = left >> right
			assign_shift(v, right)
		} else {
			s.right_shift_unsigned(v)
		}
	case uint64:
		s.right_shift_unsigned(v)
	}
}

func (s *solver) left_shift_unsigned(v *value) {
	left := to_num_unsigned(s.l.expr)
	right := to_num_unsigned(s.r.expr)
	v.expr = left << right
	assign_shift(v, right)
}

func (s *solver) left_shift(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case int64:
		if left < 0 {
			right := to_num_unsigned(s.r.expr)
			v.expr = left << right
			assign_shift(v, right)
		} else {
			s.left_shift_unsigned(v)
		}
	case uint64:
		s.left_shift_unsigned(v)
	}
}

func (s *solver) and(v *value) {
	if !s.is_const_expr() {
		return
	}
	switch left := s.l.expr.(type) {
	case bool:
		v.expr = left && s.r.expr.(bool)
	}
}

func (s *solver) or(v *value) {
	if !s.is_const_expr() {
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
			s.r.data.DataType.Kind, s.l.data.DataType.Kind)
		return
	}
	if !types.IsPtr(s.l.data.DataType) {
		s.l, s.r = s.r, s.l
	}
	switch s.op.Kind {
	case lex.KND_PLUS, lex.KND_MINUS:
		v.data.DataType = s.l.data.DataType
	case lex.KND_EQS, lex.KND_NOT_EQ, lex.KND_LT, lex.KND_GT,
		lex.KND_GREAT_EQ, lex.KND_LESS_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, "pointer")
	}
	return
}

func (s *solver) enum() (v value) {
	if types.IsEnum(s.l.data.DataType) {
		s.l.data.DataType = s.l.data.DataType.Tag.(*Enum).DataType
	}
	if types.IsEnum(s.r.data.DataType) {
		s.r.data.DataType = s.r.data.DataType.Tag.(*Enum).DataType
	}
	return s.solve()
}

func (s *solver) str() (v value) {
	v.data.Token = s.op
	// Not both string?
	if s.l.data.DataType.Id != s.r.data.DataType.Id {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.l.data.DataType.Kind, s.r.data.DataType.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_PLUS:
		v.data.DataType.Id = types.STR
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.add(&v)
	case lex.KND_EQS:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.eq(&v)
	case lex.KND_NOT_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.not_eq(&v)
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
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
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
			s.r.data.DataType.Kind, s.l.data.DataType.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_EQS:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.eq(&v)
	case lex.KND_NOT_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.not_eq(&v)
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, lex.KND_BOOL)
	}
	return
}

func (s *solver) float_mod() (v value, ok bool) {
	if !types.IsInteger(s.l.data.DataType.Id) {
		if !types.IsInteger(s.r.data.DataType.Id) {
			return
		}
		s.l, s.r = s.r, s.l
	}
	switch {
	case types.IsSignedInteger(s.l.data.DataType.Id):
		switch {
		case int_assignable(types.I64, s.r):
			return s.signed(), true
		case int_assignable(types.U64, s.r):
			return s.unsigned(), true
		}
	case types.IsUnsignedInteger(s.l.data.DataType.Id):
		if int_assignable(types.I64, s.r) ||
			int_assignable(types.U64, s.r) {
			return s.unsigned(), true
		}
	}
	return
}

func (s *solver) float() (v value) {
	v.data.Token = s.op
	if !types.IsNumeric(s.l.data.DataType.Id) ||
		!types.IsNumeric(s.r.data.DataType.Id) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.DataType.Kind, s.l.data.DataType.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_EQS:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.eq(&v)
	case lex.KND_NOT_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.not_eq(&v)
	case lex.KND_LT:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.lt(&v)
	case lex.KND_GT:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.gt(&v)
	case lex.KND_GREAT_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.gt_eq(&v)
	case lex.KND_LESS_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.lt_eq(&v)
	case lex.KND_PLUS:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.add(&v)
	case lex.KND_MINUS:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.sub(&v)
	case lex.KND_STAR:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.mul(&v)
	case lex.KND_SOLIDUS:
		// Ignore float if expression has integer
		if types.IsInteger(s.l.data.DataType.Id) && types.IsInteger(s.r.data.DataType.Id) {
		} else if types.IsInteger(s.l.data.DataType.Id) {
			s.r.data.DataType = s.l.data.DataType
		} else if types.IsInteger(s.r.data.DataType.Id) {
			s.l.data.DataType = s.r.data.DataType
		}
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.div(&v)
	case lex.KND_PERCENT:
		var ok bool
		v, ok = s.float_mod()
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
	if !types.IsNumeric(s.l.data.DataType.Id) ||
		!types.IsNumeric(s.r.data.DataType.Id) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.DataType.Kind, s.l.data.DataType.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_EQS:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.eq(&v)
	case lex.KND_NOT_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.not_eq(&v)
	case lex.KND_LT:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.lt(&v)
	case lex.KND_GT:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.gt(&v)
	case lex.KND_GREAT_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.gt_eq(&v)
	case lex.KND_LESS_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.lt_eq(&v)
	case lex.KND_PLUS:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.add(&v)
	case lex.KND_MINUS:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.sub(&v)
	case lex.KND_STAR:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.mul(&v)
	case lex.KND_SOLIDUS:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.div(&v)
	case lex.KND_PERCENT:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.mod(&v)
	case lex.KND_AMPER:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.bitwise_and(&v)
	case lex.KND_VLINE:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.bitwise_or(&v)
	case lex.KND_CARET:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.bitwise_xor(&v)
	case lex.KND_RSHIFT:
		v.data.DataType.Id = types.U64
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		if !is_ok_for_shifting(s.r) {
			s.p.pusherrtok(s.op, "bitshift_must_unsigned")
		}
		s.right_shift(&v)
	case lex.KND_LSHIFT:
		v.data.DataType.Id = types.U64
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		if !is_ok_for_shifting(s.r) {
			s.p.pusherrtok(s.op, "bitshift_must_unsigned")
		}
		s.left_shift(&v)
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_int", s.op.Kind)
	}
	return
}

func (s *solver) unsigned() (v value) {
	v.data.Token = s.op
	if !types.IsNumeric(s.l.data.DataType.Id) ||
		!types.IsNumeric(s.r.data.DataType.Id) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.DataType.Kind, s.l.data.DataType.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_EQS:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.eq(&v)
	case lex.KND_NOT_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.not_eq(&v)
	case lex.KND_LT:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.lt(&v)
	case lex.KND_GT:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.gt(&v)
	case lex.KND_GREAT_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.gt_eq(&v)
	case lex.KND_LESS_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		s.lt_eq(&v)
	case lex.KND_PLUS:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.add(&v)
	case lex.KND_MINUS:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.sub(&v)
	case lex.KND_STAR:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.mul(&v)
	case lex.KND_SOLIDUS:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.div(&v)
	case lex.KND_PERCENT:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.mod(&v)
	case lex.KND_AMPER:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.bitwise_and(&v)
	case lex.KND_VLINE:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.bitwise_or(&v)
	case lex.KND_CARET:
		v.data.DataType = s.l.data.DataType
		if types.TypeGreaterThan(s.r.data.DataType.Id, v.data.DataType.Id) {
			v.data.DataType = s.r.data.DataType
		}
		s.bitwise_xor(&v)
	case lex.KND_RSHIFT:
		v.data.DataType.Id = types.U64
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		if !is_ok_for_shifting(s.r) {
			s.p.pusherrtok(s.op, "bitshift_must_unsigned")
		}
		s.right_shift(&v)
	case lex.KND_LSHIFT:
		v.data.DataType.Id = types.U64
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		if !is_ok_for_shifting(s.r) {
			s.p.pusherrtok(s.op, "bitshift_must_unsigned")
		}
		s.left_shift(&v)
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_uint", s.op.Kind)
	}
	return
}

func (s *solver) logical() (v value) {
	if s.l.data.DataType.Id != types.BOOL ||
		s.r.data.DataType.Id != types.BOOL {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "logical_not_bool")
		return
	}
	v.data.Token = s.op
	v.data.DataType.Id = types.BOOL
	v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	if !s.is_const_expr() {
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
			s.r.data.DataType.Kind, s.l.data.DataType.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_EQS, lex.KND_NOT_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, s.l.data.DataType.Kind)
	}
	return
}

func (s *solver) slice() (v value) {
	v.data.Token = s.op
	if !s.types_are_compatible(true) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.DataType.Kind, s.l.data.DataType.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_EQS, lex.KND_NOT_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype",
			s.op.Kind, s.l.data.DataType.Kind)
	}
	return
}

func (s *solver) nil() (v value) {
	v.data.Token = s.op
	if !s.types_are_compatible(false) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.DataType.Kind, s.l.data.DataType.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_NOT_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		if s.is_const_expr() {
			v.expr = s.l.expr != nil && s.r.expr != nil
		}
	case lex.KND_EQS:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		if s.is_const_expr() {
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
	if s.l.data.DataType.Kind != s.r.data.DataType.Kind {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.DataType.Kind, s.l.data.DataType.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_NOT_EQ, lex.KND_EQS:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, lex.KND_STRUCT)
	}
	return
}

func (s *solver) traitv() (v value) {
	v.data.Token = s.op
	if !s.types_are_compatible(true) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.DataType.Kind, s.l.data.DataType.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_NOT_EQ, lex.KND_EQS:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, lex.KND_TRAIT)
	}
	return
}

func (s *solver) fnv() (v value) {
	v.data.Token = s.op
	if (!types.IsPure(s.l.data.DataType) || s.l.data.DataType.Id != types.NIL) &&
		(!types.IsPure(s.r.data.DataType) || s.r.data.DataType.Id != types.NIL) {
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "incompatible_types",
			s.r.data.DataType.Kind, s.l.data.DataType.Kind)
		return
	}
	switch s.op.Kind {
	case lex.KND_NOT_EQ:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	case lex.KND_EQS:
		v.data.DataType.Id = types.BOOL
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	default:
		s.p.eval.has_error = true
		s.p.pusherrtok(s.op, "operator_not_for_juletype", s.op.Kind, lex.KND_NIL)
	}
	return
}

func (s *solver) types_are_compatible(ignore_any bool) bool {
	checker := types.Checker{
		L:           s.l.data.DataType,
		R:           s.r.data.DataType,
		IgnoreAny:   ignore_any,
		ErrTok:      s.op,
		AllowAssign: true,
	}
	ok := checker.Check()
	s.p.pusherrs(checker.Errors...)
	return ok
}

func (s *solver) is_const_expr() bool { return s.l.constant && s.r.constant }

func (s *solver) finalize(v *value) {
	if types.IsVoid(v.data.DataType) {
		v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	} else {
		v.constant = s.is_const_expr()
		if v.constant {
			if s.value_is_rune() {
				// Save rune literal.
				v.data.Value = "'" + string(rune(v.expr.(int64))) + "'"
			}
			normalize_bitsize(v)
			v.model = get_const_expr_model(*v)
		}
	}

	if s.l.cast_type != nil && s.r.cast_type == nil {
		v.cast_type = s.l.cast_type
	} else if s.r.cast_type != nil && s.l.cast_type == nil {
		v.cast_type = s.r.cast_type
	}
}

func (s *solver) solve() (v value) {
	switch {
	case s.op.Kind == lex.KND_DBL_AMPER || s.op.Kind == lex.KND_DBL_VLINE:
		v = s.logical()
	case types.IsFn(s.l.data.DataType) || types.IsFn(s.r.data.DataType):
		v = s.fnv()
	case types.IsArray(s.l.data.DataType) || types.IsArray(s.r.data.DataType):
		v = s.array()
	case types.IsSlice(s.l.data.DataType) || types.IsSlice(s.r.data.DataType):
		v = s.slice()
	case types.IsPtr(s.l.data.DataType) || types.IsPtr(s.r.data.DataType):
		v = s.ptr()
	case types.IsEnum(s.l.data.DataType) || types.IsEnum(s.r.data.DataType):
		v = s.enum()
	case types.IsStruct(s.l.data.DataType) || types.IsStruct(s.r.data.DataType):
		v = s.structure()
	case types.IsTrait(s.l.data.DataType) || types.IsTrait(s.r.data.DataType):
		v = s.traitv()
	case s.l.data.DataType.Id == types.NIL || s.r.data.DataType.Id == types.NIL:
		v = s.nil()
	case s.l.data.DataType.Id == types.ANY || s.r.data.DataType.Id == types.ANY:
		v = s.any()
	case s.l.data.DataType.Id == types.BOOL || s.r.data.DataType.Id == types.BOOL:
		v = s.bool()
	case s.l.data.DataType.Id == types.STR || s.r.data.DataType.Id == types.STR:
		v = s.str()
	case types.IsFloat(s.l.data.DataType.Id) || types.IsFloat(s.r.data.DataType.Id):
		v = s.float()
	case types.IsUnsignedInteger(s.l.data.DataType.Id) ||
		types.IsUnsignedInteger(s.r.data.DataType.Id):
		v = s.unsigned()
	case types.IsSignedNumeric(s.l.data.DataType.Id) ||
		types.IsSignedNumeric(s.r.data.DataType.Id):
		v = s.signed()
	}
	s.finalize(&v)
	return
}
