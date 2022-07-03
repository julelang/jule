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
	elementType.Val = elementType.Val[2:]
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == xtype.Void {
		keyB.Type = elementType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkTypeAsync(elementType, keyB.Type, true, fc.profile.InTok)
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
		keyA.Type.Val = xtype.CxxTypeIdFromType(keyA.Type.Id)
		return
	}
	var ok bool
	keyA.Type, ok = fc.p.realType(keyA.Type, true)
	if ok {
		if !typeIsSingle(keyA.Type) || !xtype.IsNumericType(keyA.Type.Id) {
			fc.p.pusherrtok(keyA.IdTok, "incompatible_datatype",
				keyA.Type.Val, xtype.NumericTypeStr)
		}
	}
}

func (fc *foreachChecker) checkKeyAMapKey() {
	if xapi.IsIgnoreId(fc.profile.KeyA.Id) {
		return
	}
	keyType := fc.val.ast.Type.Tag.([]DataType)[0]
	keyA := &fc.profile.KeyA
	if keyA.Type.Id == xtype.Void {
		keyA.Type = keyType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkTypeAsync(keyType, keyA.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) checkKeyBMapVal() {
	if xapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	valType := fc.val.ast.Type.Tag.([]DataType)[1]
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == xtype.Void {
		keyB.Type = valType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkTypeAsync(valType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) str() {
	fc.checkKeyASize()
	if xapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	runeType := DataType{
		Id:  xtype.Char,
		Val: xtype.CxxTypeIdFromType(xtype.Char),
	}
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == xtype.Void {
		keyB.Type = runeType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkTypeAsync(runeType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) check() {
	switch {
	case typeIsArray(fc.val.ast.Type):
		fc.array()
	case typeIsMap(fc.val.ast.Type):
		fc.xmap()
	case fc.val.ast.Type.Id == xtype.Str:
		fc.str()
	}
}
