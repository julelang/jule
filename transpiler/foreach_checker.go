package transpiler

import (
	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/pkg/juleapi"
	"github.com/jule-lang/jule/pkg/juletype"
)

type foreachChecker struct {
	p       *Transpiler
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
	b.Type = componentType
	val := fc.val
	val.data.Type = componentType
	fc.p.check_valid_init_expr(b.Mutable, val, fc.profile.InToken)
}

func (fc *foreachChecker) slice() {
	fc.checkKeyASize()
	if juleapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	componentType := *fc.profile.ExprType.ComponentType
	b := &fc.profile.KeyB
	b.Type = componentType
	val := fc.val
	val.data.Type = componentType
	fc.p.check_valid_init_expr(b.Mutable, val, fc.profile.InToken)
}

func (fc *foreachChecker) hashmap() {
	fc.checkKeyAMapKey()
	fc.checkKeyBMapVal()
}

func (fc *foreachChecker) checkKeyASize() {
	if juleapi.IsIgnoreId(fc.profile.KeyA.Id) {
		return
	}
	a := &fc.profile.KeyA
	a.Type.Id = juletype.Int
	a.Type.Kind = juletype.TypeMap[a.Type.Id]
}

func (fc *foreachChecker) checkKeyAMapKey() {
	if juleapi.IsIgnoreId(fc.profile.KeyA.Id) {
		return
	}
	keyType := fc.val.data.Type.Tag.([]Type)[0]
	a := &fc.profile.KeyA
	a.Type = keyType
	val := fc.val
	val.data.Type = keyType
	fc.p.check_valid_init_expr(a.Mutable, val, fc.profile.InToken)
}

func (fc *foreachChecker) checkKeyBMapVal() {
	if juleapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	valType := fc.val.data.Type.Tag.([]Type)[1]
	b := &fc.profile.KeyB
	b.Type = valType
	val := fc.val
	val.data.Type = valType
	fc.p.check_valid_init_expr(b.Mutable, val, fc.profile.InToken)
}

func (fc *foreachChecker) str() {
	fc.checkKeyASize()
	if juleapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	runeType := Type{
		Id:   juletype.U8,
		Kind: juletype.TypeMap[juletype.U8],
	}
	b := &fc.profile.KeyB
	b.Type = runeType
}

func (fc *foreachChecker) check() {
	switch {
	case typeIsSlice(fc.val.data.Type):
		fc.slice()
	case typeIsArray(fc.val.data.Type):
		fc.array()
	case typeIsMap(fc.val.data.Type):
		fc.hashmap()
	case fc.val.data.Type.Id == juletype.Str:
		fc.str()
	}
}
