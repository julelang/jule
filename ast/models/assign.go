package models

import (
	"strings"

	"github.com/the-xlang/xxc/pkg/xapi"
)

// AssignLeft is selector for assignment operation.
type AssignLeft struct {
	Var    Var
	Expr   Expr
	Ignore bool
}

func (as AssignLeft) String() string {
	switch {
	case as.Var.New:
		// Returns variable identifier.
		tok := as.Expr.Toks[0]
		return xapi.OutId(tok.Kind, tok.File)
	case as.Ignore:
		return xapi.CxxIgnore
	}
	return as.Expr.String()
}

// Assign is assignment AST model.
type Assign struct {
	Setter      Tok
	Left        []AssignLeft
	Right       []Expr
	IsExpr      bool
	MultipleRet bool
}

func (a *Assign) cxxSingleAssign() string {
	expr := a.Left[0]
	if expr.Var.New {
		expr.Var.Expr = a.Right[0]
		s := expr.Var.String()
		return s[:len(s)-1] // Remove statement terminator
	}
	var cxx strings.Builder
	if len(expr.Expr.Toks) != 1 ||
		!xapi.IsIgnoreId(expr.Expr.Toks[0].Kind) {
		cxx.WriteString(expr.String())
		cxx.WriteString(a.Setter.Kind)
	}
	cxx.WriteString(a.Right[0].String())
	return cxx.String()
}

func (a *Assign) hasLeft() bool {
	for _, s := range a.Left {
		if !s.Ignore {
			return true
		}
	}
	return false
}

func (a *Assign) cxxMultipleAssign() string {
	var cxx strings.Builder
	if !a.hasLeft() {
		for _, right := range a.Right {
			cxx.WriteString(right.String())
			cxx.WriteByte(';')
		}
		return cxx.String()[:cxx.Len()-1] // Remove last semicolon
	}
	cxx.WriteString(a.cxxNewDefines())
	cxx.WriteString("std::tie(")
	var expCxx strings.Builder
	expCxx.WriteString("std::make_tuple(")
	for i, left := range a.Left {
		cxx.WriteString(left.String())
		cxx.WriteByte(',')
		expCxx.WriteString(a.Right[i].String())
		expCxx.WriteByte(',')
	}
	str := cxx.String()[:cxx.Len()-1] + ")"
	cxx.Reset()
	cxx.WriteString(str)
	cxx.WriteString(a.Setter.Kind)
	cxx.WriteString(expCxx.String()[:expCxx.Len()-1] + ")")
	return cxx.String()
}

func (a *Assign) cxxMultiRet() string {
	var cxx strings.Builder
	cxx.WriteString(a.cxxNewDefines())
	cxx.WriteString("std::tie(")
	for _, left := range a.Left {
		if left.Ignore {
			cxx.WriteString(xapi.CxxIgnore)
			cxx.WriteByte(',')
			continue
		}
		cxx.WriteString(left.String())
		cxx.WriteByte(',')
	}
	str := cxx.String()[:cxx.Len()-1]
	cxx.Reset()
	cxx.WriteString(str)
	cxx.WriteByte(')')
	cxx.WriteString(a.Setter.Kind)
	cxx.WriteString(a.Right[0].String())
	return cxx.String()
}

func (a *Assign) cxxNewDefines() string {
	var cxx strings.Builder
	for _, left := range a.Left {
		if left.Ignore || !left.Var.New {
			continue
		}
		cxx.WriteString(left.Var.String() + " ")
	}
	return cxx.String()
}

func (a *Assign) cxxSuffix() string {
	var cxx strings.Builder
	cxx.WriteString(a.Left[0].Expr.String())
	cxx.WriteString(a.Setter.Kind)
	return cxx.String()
}

func (a Assign) String() string {
	var cxx strings.Builder
	switch {
	case len(a.Right) == 0:
		cxx.WriteString(a.cxxSuffix())
	case a.MultipleRet:
		cxx.WriteString(a.cxxMultiRet())
	case len(a.Left) == 1:
		cxx.WriteString(a.cxxSingleAssign())
	default:
		cxx.WriteString(a.cxxMultipleAssign())
	}
	if !a.IsExpr {
		cxx.WriteByte(';')
	}
	return cxx.String()
}
