package models

import (
	"strings"

	"github.com/jule-lang/jule/pkg/juleapi"
	"github.com/jule-lang/jule/pkg/juletype"
)

type foreach_setter interface {
	setup_vars(key_a, key_b Var) string
	next_steps(ket_a, key_b Var, begin string) string
}

type index_setter struct {}

func (index_setter) setup_vars(key_a, key_b Var) string {
	var cpp strings.Builder
	indent := IndentString()
	if !juleapi.IsIgnoreId(key_a.Id) {
		if key_a.New {
			cpp.WriteString(key_a.String())
			cpp.WriteByte(' ')
		}
		cpp.WriteString(key_a.OutId())
		cpp.WriteString(" = 0;\n")
		cpp.WriteString(indent)
	}
	if !juleapi.IsIgnoreId(key_b.Id) {
		if key_b.New {
			cpp.WriteString(key_b.String())
			cpp.WriteByte(' ')
		}
		cpp.WriteString(key_b.OutId())
		cpp.WriteString(" = *__julec_foreach_begin;\n")
		cpp.WriteString(indent)
	}
	return cpp.String()
}

func (index_setter) next_steps(key_a, key_b Var, begin string) string {
	var cpp strings.Builder
	indent := IndentString()
	cpp.WriteString("++__julec_foreach_begin;\n")
	cpp.WriteString(indent)
	cpp.WriteString("if (__julec_foreach_begin != __julec_foreach_end) { ")
	if !juleapi.IsIgnoreId(key_a.Id) {
		cpp.WriteString("++")
		cpp.WriteString(key_a.OutId())
		cpp.WriteString("; ")
	}
	if !juleapi.IsIgnoreId(key_b.Id) {
		cpp.WriteString(key_b.OutId())
		cpp.WriteString(" = *__julec_foreach_begin; ")
	}
	cpp.WriteString("goto ")
	cpp.WriteString(begin)
	cpp.WriteString("; }\n")
	return cpp.String()
}

type map_setter struct {}

func (map_setter) setup_vars(key_a, key_b Var) string {
	var cpp strings.Builder
	indent := IndentString()
	if !juleapi.IsIgnoreId(key_a.Id) {
		if key_a.New {
			cpp.WriteString(key_a.String())
			cpp.WriteByte(' ')
		}
		cpp.WriteString(key_a.OutId())
		cpp.WriteString(" = __julec_foreach_begin->first;\n")
		cpp.WriteString(indent)
	}
	if !juleapi.IsIgnoreId(key_b.Id) {
		if key_b.New {
			cpp.WriteString(key_b.String())
			cpp.WriteByte(' ')
		}
		cpp.WriteString(key_b.OutId())
		cpp.WriteString(" = __julec_foreach_begin->second;\n")
		cpp.WriteString(indent)
	}
	return cpp.String()
}

func (map_setter) next_steps(key_a, key_b Var, begin string) string {
	var cpp strings.Builder
	indent := IndentString()
	cpp.WriteString("++__julec_foreach_begin;\n")
	cpp.WriteString(indent)
	cpp.WriteString("if (__julec_foreach_begin != __julec_foreach_end) { ")
	if !juleapi.IsIgnoreId(key_a.Id) {
		cpp.WriteString(key_a.OutId())
		cpp.WriteString(" = __julec_foreach_begin->first; ")
	}
	if !juleapi.IsIgnoreId(key_b.Id) {
		cpp.WriteString(key_b.OutId())
		cpp.WriteString(" = __julec_foreach_begin->second; ")
	}
	cpp.WriteString("goto ")
	cpp.WriteString(begin)
	cpp.WriteString("; }\n")
	return cpp.String()
}

// IterForeach is foreach iteration profile.
type IterForeach struct {
	KeyA     Var
	KeyB     Var
	InTok    Tok
	Expr     Expr
	ExprType DataType
}

func (f IterForeach) String(i *Iter) string {
	switch f.ExprType.Id {
	case juletype.Str, juletype.Slice, juletype.Array:
		return f.IterationString(i, index_setter{})
	case juletype.Map:
		return f.IterationString(i, map_setter{})
	}
	return ""
}

func (f *IterForeach) IterationString(i *Iter, setter foreach_setter) string {
	var cpp strings.Builder
	cpp.WriteString("{\n")
	AddIndent()
	indent := IndentString()
	cpp.WriteString(indent)
	cpp.WriteString("auto __julec_foreach_expr = ")
	cpp.WriteString(f.Expr.String())
	cpp.WriteString(";\n")
	cpp.WriteString(indent)
	cpp.WriteString("auto __julec_foreach_begin = __julec_foreach_expr.begin();\n")
	cpp.WriteString(indent)
	cpp.WriteString("const auto __julec_foreach_end = __julec_foreach_expr.end();\n")
	cpp.WriteString(indent)
	cpp.WriteString(setter.setup_vars(f.KeyA, f.KeyB))
	begin := i.BeginLabel()
	cpp.WriteString(begin)
	cpp.WriteString(":;\n")
	cpp.WriteString(indent)
	cpp.WriteString(i.Block.String())
	cpp.WriteByte('\n')
	cpp.WriteString(indent)
	cpp.WriteString(i.NextLabel())
	cpp.WriteString(":;\n")
	cpp.WriteString(indent)
	cpp.WriteString(setter.next_steps(f.KeyA, f.KeyB, begin))
	cpp.WriteString(indent)
	cpp.WriteString(i.EndLabel())
	cpp.WriteString(":;")
	cpp.WriteByte('\n')
	DoneIndent()
	cpp.WriteString(IndentString())
	cpp.WriteByte('}')
	return cpp.String()
}
