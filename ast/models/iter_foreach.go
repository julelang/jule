package models

import (
	"strings"

	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xapi"
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
	if !xapi.IsIgnoreId(f.KeyA.Id) {
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
	if !xapi.IsIgnoreId(f.KeyB.Id) {
		cpp.WriteByte(',')
		cpp.WriteString(f.KeyB.Type.String())
	}
	cpp.WriteString(">(")
	cpp.WriteString(f.Expr.String())
	cpp.WriteString(", [&](")
	cpp.WriteString(f.KeyA.Type.String())
	cpp.WriteByte(' ')
	cpp.WriteString(xapi.OutId(f.KeyA.Id, f.KeyA.IdTok.File))
	if !xapi.IsIgnoreId(f.KeyB.Id) {
		cpp.WriteByte(',')
		cpp.WriteString(f.KeyB.Type.String())
		cpp.WriteByte(' ')
		cpp.WriteString(xapi.OutId(f.KeyB.Id, f.KeyB.IdTok.File))
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
	cpp.WriteString(xapi.OutId(f.KeyA.Id, f.KeyA.IdTok.File))
	if !xapi.IsIgnoreId(f.KeyB.Id) {
		cpp.WriteByte(',')
		cpp.WriteString(f.KeyB.Type.String())
		cpp.WriteByte(' ')
		cpp.WriteString(xapi.OutId(f.KeyB.Id, f.KeyB.IdTok.File))
	}
	cpp.WriteString(") -> void ")
	cpp.WriteString(iter.Block.String())
	cpp.WriteString(");")
	return cpp.String()
}

func (f *IterForeach) ForeachString(iter Iter) string {
	switch {
	case f.ExprType.Kind == tokens.STR,
		strings.HasPrefix(f.ExprType.Kind, x.Prefix_Slice),
		strings.HasPrefix(f.ExprType.Kind, x.Prefix_Array):
		return f.ClassicString(iter)
	case f.ExprType.Kind[0] == '[':
		return f.MapString(iter)
	}
	return ""
}

func (f *IterForeach) IterationString(iter Iter) string {
	var cpp strings.Builder
	cpp.WriteString("for (auto ")
	cpp.WriteString(xapi.OutId(f.KeyB.Id, f.KeyB.IdTok.File))
	cpp.WriteString(" : ")
	cpp.WriteString(f.Expr.String())
	cpp.WriteString(") ")
	cpp.WriteString(iter.Block.String())
	return cpp.String()
}
