package models

import (
	"strings"

	"github.com/jule-lang/jule/pkg/juleapi"
	"github.com/jule-lang/jule/pkg/juletype"
)

// IterForeach is foreach iteration profile.
type IterForeach struct {
	KeyA     Var
	KeyB     Var
	InTok    Tok
	Expr     Expr
	ExprType DataType
}

func (f IterForeach) String(iter Iter) string {
	if !juleapi.IsIgnoreId(f.KeyA.Id) {
		return f.ForeachString(iter)
	}
	return f.IterationString(iter)
}

func (f *IterForeach) ClassicString(iter Iter) string {
	var cpp strings.Builder
	cpp.WriteString("foreach<")
	cpp.WriteString(f.ExprType.String())
	cpp.WriteByte(',')
	cpp.WriteString(f.KeyA.Type.String())
	if !juleapi.IsIgnoreId(f.KeyB.Id) {
		cpp.WriteByte(',')
		cpp.WriteString(f.KeyB.Type.String())
	}
	cpp.WriteString(">(")
	cpp.WriteString(f.Expr.String())
	cpp.WriteString(", [&](")
	cpp.WriteString(f.KeyA.Type.String())
	cpp.WriteByte(' ')
	cpp.WriteString(f.KeyA.OutId())
	if !juleapi.IsIgnoreId(f.KeyB.Id) {
		cpp.WriteByte(',')
		cpp.WriteString(f.KeyB.Type.String())
		cpp.WriteByte(' ')
		cpp.WriteString(f.KeyB.OutId())
	}
	cpp.WriteString(") -> void ")
	cpp.WriteString(iter.Block.String())
	cpp.WriteString(");")
	return cpp.String()
}

func (f *IterForeach) MapString(iter Iter) string {
	var cpp strings.Builder
	cpp.WriteString("foreach<")
	types := f.ExprType.Tag.([]DataType)
	cpp.WriteString(types[0].String())
	cpp.WriteByte(',')
	cpp.WriteString(types[1].String())
	cpp.WriteString(">(")
	cpp.WriteString(f.Expr.String())
	cpp.WriteString(", [&](")
	cpp.WriteString(f.KeyA.Type.String())
	cpp.WriteByte(' ')
	cpp.WriteString(f.KeyA.OutId())
	if !juleapi.IsIgnoreId(f.KeyB.Id) {
		cpp.WriteByte(',')
		cpp.WriteString(f.KeyB.Type.String())
		cpp.WriteByte(' ')
		cpp.WriteString(f.KeyB.OutId())
	}
	cpp.WriteString(") -> void ")
	cpp.WriteString(iter.Block.String())
	cpp.WriteString(");")
	return cpp.String()
}

func (f *IterForeach) ForeachString(iter Iter) string {
	switch f.ExprType.Id {
	case juletype.Str, juletype.Slice, juletype.Array:
		return f.ClassicString(iter)
	case juletype.Map:
		return f.MapString(iter)
	}
	return ""
}

func (f *IterForeach) IterationString(iter Iter) string {
	var cpp strings.Builder
	cpp.WriteString("for (auto ")
	cpp.WriteString(f.KeyB.OutId())
	cpp.WriteString(" : ")
	cpp.WriteString(f.Expr.String())
	cpp.WriteString(") ")
	cpp.WriteString(iter.Block.String())
	return cpp.String()
}
