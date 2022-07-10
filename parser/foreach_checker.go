package parser

import (
	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xtype"
)

type foreachChecker struct {
	p       *Parser
	profile *models.IterForeach
	val     value
}

func (fc *foreachChecker) array() {
	fc.checkKeyASize()
	if xapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	elementType := fc.profile.ExprType
	elementType.Kind = elementType.Kind[2:]
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == xtype.Void {
		keyB.Type = elementType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkType(elementType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) xmap() {
	fc.checkKeyAMapKey()
	fc.checkKeyBMapVal()
}

func (fc *foreachChecker) checkKeyASize() {
	if xapi.IsIgnoreId(fc.profile.KeyA.Id) {
		return
	}
	keyA := &fc.profile.KeyA
	if keyA.Type.Id == xtype.Void {
		keyA.Type.Id = xtype.UInt
		keyA.Type.Kind = xtype.CxxTypeIdFromType(keyA.Type.Id)
		return
	}
	var ok bool
	keyA.Type, ok = fc.p.realType(keyA.Type, true)
	if ok {
		if !typeIsPure(keyA.Type) || !xtype.IsNumericType(keyA.Type.Id) {
			fc.p.pusherrtok(keyA.IdTok, "incompatible_datatype",
				keyA.Type.Kind, xtype.NumericTypeStr)
		}
	}
}

func (fc *foreachChecker) checkKeyAMapKey() {
	if xapi.IsIgnoreId(fc.profile.KeyA.Id) {
		return
	}
	keyType := fc.val.data.Type.Tag.([]DataType)[0]
	keyA := &fc.profile.KeyA
	if keyA.Type.Id == xtype.Void {
		keyA.Type = keyType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkType(keyType, keyA.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) checkKeyBMapVal() {
	if xapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	valType := fc.val.data.Type.Tag.([]DataType)[1]
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == xtype.Void {
		keyB.Type = valType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkType(valType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) str() {
	fc.checkKeyASize()
	if xapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	runeType := DataType{
		Id:   xtype.U8,
		Kind: xtype.CxxTypeIdFromType(xtype.U8),
	}
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == xtype.Void {
		keyB.Type = runeType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkType(runeType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) check() {
	switch {
	case typeIsArray(fc.val.data.Type):
		fc.array()
	case typeIsMap(fc.val.data.Type):
		fc.xmap()
	case fc.val.data.Type.Id == xtype.Str:
		fc.str()
	}
}
