package parser

import (
	"github.com/the-xlang/xxc/pkg/xbits"
	"github.com/the-xlang/xxc/pkg/xtype"
)

type assignChecker struct {
	p         *Parser
	constant  bool
	t         DataType
	v         value
	ignoreAny bool
	errtok    Tok
}

func (ac assignChecker) checkAssignTypeAsync() {
	defer func() { ac.p.wg.Done() }()
	ac.p.checkAssignConst(ac.constant, ac.t, ac.v, ac.errtok)
	if typeIsPure(ac.t) && isConstNumeric(ac.v.data.Value) {
		switch {
		case xtype.IsSignedIntegerType(ac.t.Id):
			if xbits.CheckBitInt(ac.v.data.Value, xbits.BitsizeType(ac.t.Id)) {
				return
			}
		case xtype.IsFloatType(ac.t.Id):
			if checkFloatBit(ac.v.data, xbits.BitsizeType(ac.t.Id)) {
				return
			}
		case xtype.IsUnsignedNumericType(ac.t.Id):
			if xbits.CheckBitUInt(ac.v.data.Value, xbits.BitsizeType(ac.t.Id)) {
				return
			}
		}
	}
	ac.p.wg.Add(1)
	go ac.p.checkTypeAsync(ac.t, ac.v.data.Type, ac.ignoreAny, ac.errtok)
}
