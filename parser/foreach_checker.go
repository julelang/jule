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
	b := &fc.profile.KeyB
	if b.Type.Id == juletype.Void {
		b.Type = componentType
		return
	}
	fc.p.checkType(componentType, b.Type, true, true, fc.profile.InToken)
}

func (fc *foreachChecker) slice() {
	fc.checkKeyASize()
	if juleapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	componentType := *fc.profile.ExprType.ComponentType
	b := &fc.profile.KeyB
	if b.Type.Id == juletype.Void {
		b.Type = componentType
		return
	}
	fc.p.checkType(componentType, b.Type, true, true, fc.profile.InToken)
}

func (fc *foreachChecker) xmap() {
	fc.checkKeyAMapKey()
	fc.checkKeyBMapVal()
}

func (fc *foreachChecker) checkKeyASize() {
	if juleapi.IsIgnoreId(fc.profile.KeyA.Id) {
		return
	}
	a := &fc.profile.KeyA
	if a.Type.Id == juletype.Void {
		a.Type.Id = juletype.UInt
		a.Type.Kind = juletype.CppId(a.Type.Id)
		return
	}
	var ok bool
	a.Type, ok = fc.p.realType(a.Type, true)
	if ok {
		if !typeIsPure(a.Type) || !juletype.IsNumeric(a.Type.Id) {
			fc.p.pusherrtok(a.Token, "incompatible_types",
				a.Type.Kind, juletype.NumericTypeStr)
		}
	}
}

func (fc *foreachChecker) checkKeyAMapKey() {
	if juleapi.IsIgnoreId(fc.profile.KeyA.Id) {
		return
	}
	keyType := fc.val.data.Type.Tag.([]Type)[0]
	a := &fc.profile.KeyA
	if a.Type.Id == juletype.Void {
		a.Type = keyType
		return
	}
	fc.p.checkType(keyType, a.Type, true, true, fc.profile.InToken)
}

func (fc *foreachChecker) checkKeyBMapVal() {
	if juleapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	valType := fc.val.data.Type.Tag.([]Type)[1]
	b := &fc.profile.KeyB
	if b.Type.Id == juletype.Void {
		b.Type = valType
		return
	}
	fc.p.checkType(valType, b.Type, true, true, fc.profile.InToken)
}

func (fc *foreachChecker) str() {
	fc.checkKeyASize()
	if juleapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	runeType := Type{
		Id:   juletype.U8,
		Kind: juletype.CppId(juletype.U8),
	}
	b := &fc.profile.KeyB
	if b.Type.Id == juletype.Void {
		b.Type = runeType
		return
	}
	fc.p.checkType(runeType, b.Type, true, true, fc.profile.InToken)
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
