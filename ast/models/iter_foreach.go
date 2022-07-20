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
	var cxx strings.Builder
	cxx.WriteString("foreach<")
	cxx.WriteString(f.ExprType.String())
	cxx.WriteByte(',')
	cxx.WriteString(f.KeyA.Type.String())
	if !xapi.IsIgnoreId(f.KeyB.Id) {
		cxx.WriteByte(',')
		cxx.WriteString(f.KeyB.Type.String())
	}
	cxx.WriteString(">(")
	cxx.WriteString(f.Expr.String())
	cxx.WriteString(", [&](")
	cxx.WriteString(f.KeyA.Type.String())
	cxx.WriteByte(' ')
	cxx.WriteString(xapi.OutId(f.KeyA.Id, f.KeyA.IdTok.File))
	if !xapi.IsIgnoreId(f.KeyB.Id) {
		cxx.WriteByte(',')
		cxx.WriteString(f.KeyB.Type.String())
		cxx.WriteByte(' ')
		cxx.WriteString(xapi.OutId(f.KeyB.Id, f.KeyB.IdTok.File))
	}
	cxx.WriteString(") -> void ")
	cxx.WriteString(iter.Block.String())
	cxx.WriteString(");")
	return cxx.String()
}

func (f *IterForeach) MapString(iter Iter) string {
	var cxx strings.Builder
	cxx.WriteString("foreach<")
	types := f.ExprType.Tag.([]DataType)
	cxx.WriteString(types[0].String())
	cxx.WriteByte(',')
	cxx.WriteString(types[1].String())
	cxx.WriteString(">(")
	cxx.WriteString(f.Expr.String())
	cxx.WriteString(", [&](")
	cxx.WriteString(f.KeyA.Type.String())
	cxx.WriteByte(' ')
	cxx.WriteString(xapi.OutId(f.KeyA.Id, f.KeyA.IdTok.File))
	if !xapi.IsIgnoreId(f.KeyB.Id) {
		cxx.WriteByte(',')
		cxx.WriteString(f.KeyB.Type.String())
		cxx.WriteByte(' ')
		cxx.WriteString(xapi.OutId(f.KeyB.Id, f.KeyB.IdTok.File))
	}
	cxx.WriteString(") -> void ")
	cxx.WriteString(iter.Block.String())
	cxx.WriteString(");")
	return cxx.String()
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
	var cxx strings.Builder
	cxx.WriteString("for (auto ")
	cxx.WriteString(xapi.OutId(f.KeyB.Id, f.KeyB.IdTok.File))
	cxx.WriteString(" : ")
	cxx.WriteString(f.Expr.String())
	cxx.WriteString(") ")
	cxx.WriteString(iter.Block.String())
	return cxx.String()
}
