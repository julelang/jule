package parser

import (
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
	switch {
	case juletype.IsSignedInteger(v.data.Type.Id):
		v.expr = tonums(v.expr)
	case juletype.IsUnsignedInteger(v.data.Type.Id):
		v.expr = tonumu(v.expr)
	}
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
	p        *Parser
	left     Toks
	leftVal  value
	right    Toks
	rightVal value
	operator Tok
}

func (s *solver) eq(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.leftVal.expr.(type) {
	case bool:
		v.expr = left == s.rightVal.expr.(bool)
	case string:
		v.expr = left == s.rightVal.expr.(string)
	case float64:
		v.expr = left == tonumf(s.rightVal.expr)
	case int64:
		v.expr = left == tonums(s.rightVal.expr)
	case uint64:
		v.expr = left == tonumu(s.rightVal.expr)
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
	switch left := s.leftVal.expr.(type) {
	case float64:
		v.expr = left < tonumf(s.rightVal.expr)
	case int64:
		v.expr = left < tonums(s.rightVal.expr)
	case uint64:
		v.expr = left < tonumu(s.rightVal.expr)
	}
}

func (s *solver) gt(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.leftVal.expr.(type) {
	case float64:
		v.expr = left > tonumf(s.rightVal.expr)
	case int64:
		v.expr = left > tonums(s.rightVal.expr)
	case uint64:
		v.expr = left > tonumu(s.rightVal.expr)
	}
}

func (s *solver) lteq(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.leftVal.expr.(type) {
	case float64:
		v.expr = left <= tonumf(s.rightVal.expr)
	case int64:
		v.expr = left <= tonums(s.rightVal.expr)
	case uint64:
		v.expr = left <= tonumu(s.rightVal.expr)
	}
}

func (s *solver) gteq(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.leftVal.expr.(type) {
	case float64:
		v.expr = left >= tonumf(s.rightVal.expr)
	case int64:
		v.expr = left >= tonums(s.rightVal.expr)
	case uint64:
		v.expr = left >= tonumu(s.rightVal.expr)
	}
}

func (s *solver) add(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.leftVal.expr.(type) {
	case string:
		v.expr = left + s.rightVal.expr.(string)
	case float64:
		v.expr = left + tonumf(s.rightVal.expr)
	case int64:
		v.expr = int64(float64(left) + tonumf(s.rightVal.expr))
	case uint64:
		v.expr = uint64(float64(left) + tonumf(s.rightVal.expr))
	}
}

func (s *solver) sub(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.leftVal.expr.(type) {
	case float64:
		v.expr = left - tonumf(s.rightVal.expr)
	case int64:
		v.expr = int64(float64(left) - tonumf(s.rightVal.expr))
	case uint64:
		v.expr = uint64(float64(left) - tonumf(s.rightVal.expr))
	}
}

func (s *solver) mul(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.leftVal.expr.(type) {
	case float64:
		v.expr = left * tonumf(s.rightVal.expr)
	case int64:
		v.expr = int64(float64(left) * tonumf(s.rightVal.expr))
	case uint64:
		v.expr = uint64(float64(left) * tonumf(s.rightVal.expr))
	}
}

func (s *solver) div(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.leftVal.expr.(type) {
	case float64:
		right := tonumf(s.rightVal.expr)
		if right != 0 {
			v.expr = left / right
		} else {
			s.p.pusherrtok(s.operator, "divide_by_zero")
			v.expr = float64(0)
		}
	case int64:
		right := tonumf(s.rightVal.expr)
		if right != 0 {
			v.expr = float64(left) / right
		} else {
			s.p.pusherrtok(s.operator, "divide_by_zero")
			v.expr = int64(0)
		}
	case uint64:
		right := tonumf(s.rightVal.expr)
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
	switch left := s.leftVal.expr.(type) {
	case int64:
		right := tonums(s.rightVal.expr)
		if right != 0 {
			v.expr = left % right
		} else {
			s.p.pusherrtok(s.operator, "divide_by_zero")
			v.expr = int64(0)
		}
	case uint64:
		right := tonumu(s.rightVal.expr)
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
	switch left := s.leftVal.expr.(type) {
	case int64:
		v.expr = left & tonums(s.rightVal.expr)
	case uint64:
		v.expr = left & tonumu(s.rightVal.expr)
	}
}

func (s *solver) bitwiseOr(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.leftVal.expr.(type) {
	case int64:
		v.expr = left | tonums(s.rightVal.expr)
	case uint64:
		v.expr = left | tonumu(s.rightVal.expr)
	}
}

func (s *solver) bitwiseXor(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.leftVal.expr.(type) {
	case int64:
		v.expr = left ^ tonums(s.rightVal.expr)
	case uint64:
		v.expr = left ^ tonumu(s.rightVal.expr)
	}
}

func (s *solver) rshift(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.leftVal.expr.(type) {
	case int64:
		right := tonumu(s.rightVal.expr)
		v.expr = left >> right
		setshift(v, right)
	case uint64:
		right := tonumu(s.rightVal.expr)
		v.expr = left >> right
		setshift(v, right)
	}
}

func (s *solver) lshift(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.leftVal.expr.(type) {
	case int64:
		right := tonumu(s.rightVal.expr)
		v.expr = left << right
		setshift(v, right)
	case uint64:
		right := tonumu(s.rightVal.expr)
		v.expr = left << right
		setshift(v, right)
	}
}

func (s *solver) and(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.leftVal.expr.(type) {
	case bool:
		v.expr = left && s.rightVal.expr.(bool)
	}
}

func (s *solver) or(v *value) {
	if !s.isConstExpr() {
		return
	}
	switch left := s.leftVal.expr.(type) {
	case bool:
		v.expr = left || s.rightVal.expr.(bool)
	}
}

func (s *solver) ptr() (v value) {
	v.data.Tok = s.operator
	if !typesAreCompatible(s.leftVal.data.Type, s.rightVal.data.Type, true) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.data.Type.Kind, s.leftVal.data.Type.Kind)
		return
	}
	if !typeIsPtr(s.leftVal.data.Type) {
		s.leftVal, s.rightVal = s.rightVal, s.leftVal
	}
	switch s.operator.Kind {
	case tokens.PLUS, tokens.MINUS:
		v.data.Type = s.leftVal.data.Type
	case tokens.EQUALS, tokens.NOT_EQUALS, tokens.LESS, tokens.GREAT,
		tokens.GREAT_EQUAL, tokens.LESS_EQUAL:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_juletype", s.operator.Kind, "pointer")
	}
	return
}

func (s *solver) enum() (v value) {
	if typeIsEnum(s.leftVal.data.Type) {
		s.leftVal.data.Type = s.leftVal.data.Type.Tag.(*Enum).Type
	}
	if typeIsEnum(s.rightVal.data.Type) {
		s.rightVal.data.Type = s.rightVal.data.Type.Tag.(*Enum).Type
	}
	return s.solve()
}

func (s *solver) str() (v value) {
	v.data.Tok = s.operator
	// Not both string?
	if s.leftVal.data.Type.Id != s.rightVal.data.Type.Id {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.leftVal.data.Type.Kind, s.rightVal.data.Type.Kind)
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
		s.p.pusherrtok(s.operator, "operator_notfor_juletype",
			s.operator.Kind, tokens.STR)
	}
	return
}

func (s *solver) any() (v value) {
	v.data.Tok = s.operator
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_juletype", s.operator.Kind, tokens.ANY)
	}
	return
}

func (s *solver) bool() (v value) {
	v.data.Tok = s.operator
	if !typesAreCompatible(s.leftVal.data.Type, s.rightVal.data.Type, true) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.data.Type.Kind, s.leftVal.data.Type.Kind)
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
		s.p.pusherrtok(s.operator, "operator_notfor_juletype",
			s.operator.Kind, tokens.BOOL)
	}
	return
}

func (s *solver) float() (v value) {
	v.data.Tok = s.operator
	if !juletype.IsNumeric(s.leftVal.data.Type.Id) ||
		!juletype.IsNumeric(s.rightVal.data.Type.Id) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.data.Type.Kind, s.leftVal.data.Type.Kind)
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
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.add(&v)
	case tokens.MINUS:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.sub(&v)
	case tokens.STAR:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.mul(&v)
	case tokens.SOLIDUS:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.div(&v)
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_float", s.operator.Kind)
	}
	return
}

func (s *solver) signed() (v value) {
	v.data.Tok = s.operator
	if !juletype.IsNumeric(s.leftVal.data.Type.Id) ||
		!juletype.IsNumeric(s.rightVal.data.Type.Id) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.data.Type.Kind, s.leftVal.data.Type.Kind)
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
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.add(&v)
	case tokens.MINUS:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.sub(&v)
	case tokens.STAR:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.mul(&v)
	case tokens.SOLIDUS:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.div(&v)
	case tokens.PERCENT:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.mod(&v)
	case tokens.AMPER:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.bitwiseAnd(&v)
	case tokens.VLINE:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.bitwiseOr(&v)
	case tokens.CARET:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.bitwiseXor(&v)
	case tokens.RSHIFT:
		v.data.Type.Id = juletype.U64
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		if !okForShifting(s.rightVal) {
			s.p.pusherrtok(s.operator, "bitshift_must_unsigned")
		}
		s.rshift(&v)
	case tokens.LSHIFT:
		v.data.Type.Id = juletype.U64
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		if !okForShifting(s.rightVal) {
			s.p.pusherrtok(s.operator, "bitshift_must_unsigned")
		}
		s.lshift(&v)
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_int", s.operator.Kind)
	}
	return
}

func (s *solver) unsigned() (v value) {
	v.data.Tok = s.operator
	if !juletype.IsNumeric(s.leftVal.data.Type.Id) ||
		!juletype.IsNumeric(s.rightVal.data.Type.Id) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.data.Type.Kind, s.leftVal.data.Type.Kind)
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
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.add(&v)
	case tokens.MINUS:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.sub(&v)
	case tokens.STAR:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.mul(&v)
	case tokens.SOLIDUS:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.div(&v)
	case tokens.PERCENT:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.mod(&v)
	case tokens.AMPER:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.bitwiseAnd(&v)
	case tokens.VLINE:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.bitwiseOr(&v)
	case tokens.CARET:
		v.data.Type = s.leftVal.data.Type
		if juletype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
		s.bitwiseXor(&v)
	case tokens.RSHIFT:
		v.data.Type.Id = juletype.U64
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		if !okForShifting(s.rightVal) {
			s.p.pusherrtok(s.operator, "bitshift_must_unsigned")
		}
		s.rshift(&v)
	case tokens.LSHIFT:
		v.data.Type.Id = juletype.U64
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		if !okForShifting(s.rightVal) {
			s.p.pusherrtok(s.operator, "bitshift_must_unsigned")
		}
		s.lshift(&v)
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_uint", s.operator.Kind)
	}
	return
}

func (s *solver) logical() (v value) {
	v.data.Tok = s.operator
	v.data.Type.Id = juletype.Bool
	v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	if s.leftVal.data.Type.Id != juletype.Bool || s.rightVal.data.Type.Id != juletype.Bool {
		s.p.pusherrtok(s.operator, "logical_not_bool")
		return
	}
	if !s.isConstExpr() {
		return
	}
	switch s.operator.Kind {
	case tokens.AND:
		s.and(&v)
	case tokens.OR:
		s.or(&v)
	}
	return
}

func (s *solver) array() (v value) {
	v.data.Tok = s.operator
	if !typesAreCompatible(s.leftVal.data.Type, s.rightVal.data.Type, true) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.data.Type.Kind, s.leftVal.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_juletype", s.operator.Kind, s.leftVal.data.Type.Kind)
	}
	return
}

func (s *solver) slice() (v value) {
	v.data.Tok = s.operator
	if !typesAreCompatible(s.leftVal.data.Type, s.rightVal.data.Type, true) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.data.Type.Kind, s.leftVal.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_juletype",
			s.operator.Kind, s.leftVal.data.Type.Kind)
	}
	return
}

func (s *solver) nil() (v value) {
	v.data.Tok = s.operator
	if !typesAreCompatible(s.leftVal.data.Type, s.rightVal.data.Type, false) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.data.Type.Kind, s.leftVal.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.NOT_EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		if s.isConstExpr() {
			v.expr = s.leftVal.expr != nil && s.rightVal.expr != nil
		}
	case tokens.EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		if s.isConstExpr() {
			v.expr = s.leftVal.expr == nil && s.rightVal.expr == nil
		}
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_juletype",
			s.operator.Kind, tokens.NIL)
	}
	return
}

func (s *solver) structure() (v value) {
	v.data.Tok = s.operator
	if s.leftVal.data.Type.Kind != s.rightVal.data.Type.Kind {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.data.Type.Kind, s.leftVal.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.NOT_EQUALS, tokens.EQUALS:
		v.data.Type.Id = juletype.Bool
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_juletype",
			s.operator.Kind, tokens.STRUCT)
	}
	return
}

func (s *solver) function() (v value) {
	v.data.Tok = s.operator
	if (!typeIsPure(s.leftVal.data.Type) || s.leftVal.data.Type.Id != juletype.Nil) &&
		(!typeIsPure(s.rightVal.data.Type) || s.rightVal.data.Type.Id != juletype.Nil) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.data.Type.Kind, s.leftVal.data.Type.Kind)
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
		s.p.pusherrtok(s.operator, "operator_notfor_juletype",
			s.operator.Kind, tokens.NIL)
	}
	return
}

func (s *solver) check() bool {
	switch s.operator.Kind {
	case tokens.PLUS, tokens.MINUS, tokens.STAR, tokens.SOLIDUS, tokens.PERCENT, tokens.RSHIFT,
		tokens.LSHIFT, tokens.AMPER, tokens.VLINE, tokens.CARET, tokens.EQUALS, tokens.NOT_EQUALS,
		tokens.GREAT, tokens.LESS, tokens.GREAT_EQUAL, tokens.LESS_EQUAL:
	case tokens.AND, tokens.OR:
	default:
		s.p.pusherrtok(s.operator, "invalid_operator")
		return false
	}
	return true
}

func (s *solver) isConstExpr() bool {
	return s.leftVal.constExpr && s.rightVal.constExpr
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
	if !s.check() {
		return
	}
	switch s.operator.Kind {
	case tokens.AND, tokens.OR:
		return s.logical()
	}
	switch {
	case typeIsFunc(s.leftVal.data.Type), typeIsFunc(s.rightVal.data.Type):
		return s.function()
	case typeIsArray(s.leftVal.data.Type), typeIsArray(s.rightVal.data.Type):
		return s.array()
	case typeIsSlice(s.leftVal.data.Type), typeIsSlice(s.rightVal.data.Type):
		return s.slice()
	case typeIsPtr(s.leftVal.data.Type), typeIsPtr(s.rightVal.data.Type):
		return s.ptr()
	case typeIsEnum(s.leftVal.data.Type), typeIsEnum(s.rightVal.data.Type):
		return s.enum()
	case typeIsStruct(s.leftVal.data.Type), typeIsStruct(s.rightVal.data.Type):
		return s.structure()
	case s.leftVal.data.Type.Id == juletype.Nil, s.rightVal.data.Type.Id == juletype.Nil:
		return s.nil()
	case s.leftVal.data.Type.Id == juletype.Any, s.rightVal.data.Type.Id == juletype.Any:
		return s.any()
	case s.leftVal.data.Type.Id == juletype.Bool, s.rightVal.data.Type.Id == juletype.Bool:
		return s.bool()
	case s.leftVal.data.Type.Id == juletype.Str, s.rightVal.data.Type.Id == juletype.Str:
		return s.str()
	case juletype.IsFloat(s.leftVal.data.Type.Id),
		juletype.IsFloat(s.rightVal.data.Type.Id):
		return s.float()
	case juletype.IsUnsignedInteger(s.leftVal.data.Type.Id),
		juletype.IsUnsignedInteger(s.rightVal.data.Type.Id):
		return s.unsigned()
	case juletype.IsSignedNumeric(s.leftVal.data.Type.Id),
		juletype.IsSignedNumeric(s.rightVal.data.Type.Id):
		return s.signed()
	}
	return
}
