package parser

import (
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/xbits"
	"github.com/the-xlang/xxc/pkg/xtype"
)

type solver struct {
	p        *Parser
	left     Toks
	leftVal  value
	right    Toks
	rightVal value
	operator Tok
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
		v.data.Type.Id = xtype.Bool
		v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype", s.operator.Kind, "pointer")
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
		v.data.Type.Id = xtype.Str
		v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.data.Type.Id = xtype.Bool
		v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype",
			s.operator.Kind, tokens.STR)
	}
	return
}

func (s *solver) any() (v value) {
	v.data.Tok = s.operator
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.data.Type.Id = xtype.Bool
		v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype", s.operator.Kind, tokens.ANY)
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
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.data.Type.Id = xtype.Bool
		v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype",
			s.operator.Kind, tokens.BOOL)
	}
	return
}

func (s *solver) float() (v value) {
	v.data.Tok = s.operator
	if !xtype.IsNumericType(s.leftVal.data.Type.Id) ||
		!xtype.IsNumericType(s.rightVal.data.Type.Id) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.data.Type.Kind, s.leftVal.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS, tokens.LESS, tokens.GREAT,
		tokens.GREAT_EQUAL, tokens.LESS_EQUAL:
		v.data.Type.Id = xtype.Bool
		v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
	case tokens.PLUS, tokens.MINUS, tokens.STAR, tokens.SOLIDUS:
		v.data.Type = s.leftVal.data.Type
		if xtype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_float", s.operator.Kind)
	}
	return
}

func (s *solver) signed() (v value) {
	v.data.Tok = s.operator
	if !xtype.IsNumericType(s.leftVal.data.Type.Id) ||
		!xtype.IsNumericType(s.rightVal.data.Type.Id) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.data.Type.Kind, s.leftVal.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS, tokens.LESS,
		tokens.GREAT, tokens.GREAT_EQUAL, tokens.LESS_EQUAL:
		v.data.Type.Id = xtype.Bool
		v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
	case tokens.PLUS, tokens.MINUS, tokens.STAR, tokens.SOLIDUS,
		tokens.PERCENT, tokens.AMPER, tokens.VLINE, tokens.CARET:
		v.data.Type = s.leftVal.data.Type
		if xtype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
	case tokens.RSHIFT, tokens.LSHIFT:
		v.data.Type = s.leftVal.data.Type
		if !xtype.IsUnsignedNumericType(s.rightVal.data.Type.Id) &&
			!checkIntBit(s.rightVal.data, xbits.BitsizeType(xtype.U64)) {
			s.p.pusherrtok(s.rightVal.data.Tok, "bitshift_must_unsigned")
		}
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_int", s.operator.Kind)
	}
	return
}

func (s *solver) unsigned() (v value) {
	v.data.Tok = s.operator
	if !xtype.IsNumericType(s.leftVal.data.Type.Id) ||
		!xtype.IsNumericType(s.rightVal.data.Type.Id) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.data.Type.Kind, s.leftVal.data.Type.Kind)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS, tokens.LESS,
		tokens.GREAT, tokens.GREAT_EQUAL, tokens.LESS_EQUAL:
		v.data.Type.Id = xtype.Bool
		v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
	case tokens.PLUS, tokens.MINUS, tokens.STAR, tokens.SOLIDUS,
		tokens.PERCENT, tokens.AMPER, tokens.VLINE, tokens.CARET:
		v.data.Type = s.leftVal.data.Type
		if xtype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
	case tokens.RSHIFT, tokens.LSHIFT:
		v.data.Type = s.leftVal.data.Type
		if xtype.TypeGreaterThan(s.rightVal.data.Type.Id, v.data.Type.Id) {
			v.data.Type = s.rightVal.data.Type
		}
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_uint", s.operator.Kind)
	}
	return
}

func (s *solver) logical() (v value) {
	v.data.Tok = s.operator
	v.data.Type.Id = xtype.Bool
	v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
	if s.leftVal.data.Type.Id != xtype.Bool || s.rightVal.data.Type.Id != xtype.Bool {
		s.p.pusherrtok(s.operator, "logical_not_bool")
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
		v.data.Type.Id = xtype.Bool
		v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype", s.operator.Kind, s.leftVal.data.Type.Kind)
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
		v.data.Type.Id = xtype.Bool
		v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype",
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
	case tokens.NOT_EQUALS, tokens.EQUALS:
		v.data.Type.Id = xtype.Bool
		v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype",
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
		v.data.Type.Id = xtype.Bool
		v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype",
			s.operator.Kind, tokens.STRUCT)
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

func (s *solver) solve() (v value) {
	defer func() {
		if typeIsVoid(v.data.Type) {
			v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
		} else {
			v.constExpr = s.leftVal.constExpr && s.rightVal.constExpr
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
	case s.leftVal.data.Type.Id == xtype.Nil, s.rightVal.data.Type.Id == xtype.Nil:
		return s.nil()
	case s.leftVal.data.Type.Id == xtype.Any, s.rightVal.data.Type.Id == xtype.Any:
		return s.any()
	case s.leftVal.data.Type.Id == xtype.Bool, s.rightVal.data.Type.Id == xtype.Bool:
		return s.bool()
	case s.leftVal.data.Type.Id == xtype.Str, s.rightVal.data.Type.Id == xtype.Str:
		return s.str()
	case xtype.IsFloatType(s.leftVal.data.Type.Id),
		xtype.IsFloatType(s.rightVal.data.Type.Id):
		return s.float()
	case xtype.IsUnsignedNumericType(s.leftVal.data.Type.Id),
		xtype.IsUnsignedNumericType(s.rightVal.data.Type.Id):
		return s.unsigned()
	case xtype.IsSignedNumericType(s.leftVal.data.Type.Id),
		xtype.IsSignedNumericType(s.rightVal.data.Type.Id):
		return s.signed()
	}
	return
}
