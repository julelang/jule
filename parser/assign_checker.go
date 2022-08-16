package parser

import (
	"strconv"

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
		return t >= float64(min) && t <= float64(max)
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
		return t <= float64(max)
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
	t         DataType
	v         value
	ignoreAny bool
	errtok    Tok
}

func (ac assignChecker) checkAssignType() {
	if ac.p.eval.hasError || ac.v.data.Value == "" {
		return
	}
	if typeIsPure(ac.t) && ac.v.constExpr && typeIsPure(ac.v.data.Type) {
		switch {
		case juletype.IsFloat(ac.t.Id):
			if !floatAssignable(ac.t.Id, ac.v) {
				ac.p.pusherrtok(ac.errtok, "overflow_limits")
			}
			return
		case juletype.IsInteger(ac.t.Id) && juletype.IsInteger(ac.v.data.Type.Id):
			if !integerAssignable(ac.t.Id, ac.v) {
				ac.p.pusherrtok(ac.errtok, "overflow_limits")
			}
			return
		}
	}
	ac.p.checkType(ac.t, ac.v.data.Type, ac.ignoreAny, ac.errtok)
}
