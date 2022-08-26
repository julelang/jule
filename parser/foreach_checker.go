package parser

import (
	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/pkg/juleapi"
	"github.com/jule-lang/jule/pkg/juletype"
)

type foreachChecker struct {
	p       *Parser
	profile *models.IterForeach
	val     value
}

func (fc *foreachChecker) array() {
	fc.checkKeyASize()
	if juleapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	componentType := *fc.profile.ExprType.ComponentType
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == juletype.Void {
		keyB.Type = componentType
		return
	}
	fc.p.checkType(componentType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) slice() {
	fc.checkKeyASize()
	if juleapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	componentType := *fc.profile.ExprType.ComponentType
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == juletype.Void {
		keyB.Type = componentType
		return
	}
	fc.p.checkType(componentType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) xmap() {
	fc.checkKeyAMapKey()
	fc.checkKeyBMapVal()
}

func (fc *foreachChecker) checkKeyASize() {
	if juleapi.IsIgnoreId(fc.profile.KeyA.Id) {
		return
	}
	keyA := &fc.profile.KeyA
	if keyA.Type.Id == juletype.Void {
		keyA.Type.Id = juletype.UInt
		keyA.Type.Kind = juletype.CppId(keyA.Type.Id)
		return
	}
	var ok bool
	keyA.Type, ok = fc.p.realType(keyA.Type, true)
	if ok {
		if !typeIsPure(keyA.Type) || !juletype.IsNumeric(keyA.Type.Id) {
			fc.p.pusherrtok(keyA.Token, "incompatible_types",
				keyA.Type.Kind, juletype.NumericTypeStr)
		}
	}
}

func (fc *foreachChecker) checkKeyAMapKey() {
	if juleapi.IsIgnoreId(fc.profile.KeyA.Id) {
		return
	}
	keyType := fc.val.data.Type.Tag.([]DataType)[0]
	keyA := &fc.profile.KeyA
	if keyA.Type.Id == juletype.Void {
		keyA.Type = keyType
		return
	}
	fc.p.checkType(keyType, keyA.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) checkKeyBMapVal() {
	if juleapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	valType := fc.val.data.Type.Tag.([]DataType)[1]
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == juletype.Void {
		keyB.Type = valType
		return
	}
	fc.p.checkType(valType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) str() {
	fc.checkKeyASize()
	if juleapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	runeType := DataType{
		Id:   juletype.U8,
		Kind: juletype.CppId(juletype.U8),
	}
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == juletype.Void {
		keyB.Type = runeType
		return
	}
	fc.p.checkType(runeType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) check() {
	switch {
	case typeIsSlice(fc.val.data.Type):
		fc.slice()
	case typeIsArray(fc.val.data.Type):
		fc.array()
	case typeIsMap(fc.val.data.Type):
		fc.xmap()
	case fc.val.data.Type.Id == juletype.Str:
		fc.str()
	}
}
