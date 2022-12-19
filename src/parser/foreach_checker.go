package parser

import (
	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

type foreachChecker struct {
	p       *Parser
	profile *models.IterForeach
	val     value
}

func (fc *foreachChecker) array() {
	fc.checkKeyASize()
	if lex.IsIgnoreId(fc.profile.KeyB.Id) {
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
	if lex.IsIgnoreId(fc.profile.KeyB.Id) {
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
	if lex.IsIgnoreId(fc.profile.KeyA.Id) {
		return
	}
	a := &fc.profile.KeyA
	a.Type.Id = types.INT
	a.Type.Kind = types.TYPE_MAP[a.Type.Id]
}

func (fc *foreachChecker) checkKeyAMapKey() {
	if lex.IsIgnoreId(fc.profile.KeyA.Id) {
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
	if lex.IsIgnoreId(fc.profile.KeyB.Id) {
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
	if lex.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	runeType := Type{
		Id:   types.U8,
		Kind: types.TYPE_MAP[types.U8],
	}
	b := &fc.profile.KeyB
	b.Type = runeType
}

func (fc *foreachChecker) check() {
	switch {
	case types.IsSlice(fc.val.data.Type):
		fc.slice()
	case types.IsArray(fc.val.data.Type):
		fc.array()
	case types.IsMap(fc.val.data.Type):
		fc.hashmap()
	case fc.val.data.Type.Id == types.STR:
		fc.str()
	}
}
