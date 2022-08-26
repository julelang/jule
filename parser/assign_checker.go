package parser

import (
	"math"
	"strconv"

	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/pkg/julebits"
	"github.com/jule-lang/jule/pkg/juletype"
)

func floatAssignable(dt uint8, v value) bool {
	switch t := v.expr.(type) {
	case float64:
		v.data.Value = strconv.FormatFloat(t, 'e', -1, 64)
	case int64:
		v.data.Value = strconv.FormatFloat(float64(t), 'e', -1, 64)
	case uint64:
		v.data.Value = strconv.FormatFloat(float64(t), 'e', -1, 64)
	}
	return checkFloatBit(v.data, julebits.BitsizeType(dt))
}

func signedAssignable(dt uint8, v value) bool {
	min := juletype.MinOfType(dt)
	max := int64(juletype.MaxOfType(dt))
	switch t := v.expr.(type) {
	case float64:
		i, frac := math.Modf(t)
		if frac != 0 {
			return false
		}
		return i >= float64(min) && i <= float64(max)
	case uint64:
		if t <= uint64(max) {
			return true
		}
	case int64:
		return t >= min && t <= max
	}
	return false
}

func unsignedAssignable(dt uint8, v value) bool {
	max := juletype.MaxOfType(dt)
	switch t := v.expr.(type) {
	case float64:
		if t < 0 {
			return false
		}
		i, frac := math.Modf(t)
		if frac != 0 {
			return false
		}
		return i <= float64(max)
	case uint64:
		if t <= max {
			return true
		}
	case int64:
		if t < 0 {
			return false
		}
		return uint64(t) <= max
	}
	return false
}

func integerAssignable(dt uint8, v value) bool {
	switch {
	case juletype.IsSignedInteger(dt):
		return signedAssignable(dt, v)
	case juletype.IsUnsignedInteger(dt):
		return unsignedAssignable(dt, v)
	}
	return false
}

type assignChecker struct {
	p         *Parser
	t         Type
	v         value
	ignoreAny bool
	errtok    lex.Token
}

func (ac assignChecker) checkAssignType() {
	if ac.p.eval.has_error || ac.v.data.Value == "" {
		return
	}
	if ac.v.constExpr &&
		typeIsPure(ac.t) &&
		typeIsPure(ac.v.data.Type) &&
		juletype.IsNumeric(ac.v.data.Type.Id) {
		switch {
		case juletype.IsFloat(ac.t.Id):
			if !floatAssignable(ac.t.Id, ac.v) {
				ac.p.pusherrtok(ac.errtok, "overflow_limits")
			}
			return
		case juletype.IsInteger(ac.t.Id):
			if !integerAssignable(ac.t.Id, ac.v) {
				ac.p.pusherrtok(ac.errtok, "overflow_limits")
			}
			return
		}
	}
	ac.p.checkType(ac.t, ac.v.data.Type, ac.ignoreAny, ac.errtok)
}
