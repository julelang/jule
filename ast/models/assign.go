package models

import (
	"strings"

	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/pkg/juleapi"
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
		return as.Var.OutId()
	case as.Ignore:
		return juleapi.CPP_IGNORE
	}
	return as.Expr.String()
}

// Assign is assignment AST model.
type Assign struct {
	Setter      lex.Token
	Left        []AssignLeft
	Right       []Expr
	IsExpr      bool
	MultipleRet bool
}

func (a *Assign) cppSingleAssign() string {
	expr := a.Left[0]
	if expr.Var.New {
		expr.Var.Expr = a.Right[0]
		s := expr.Var.String()
		return s[:len(s)-1] // Remove statement terminator
	}
	var cpp strings.Builder
	if len(expr.Expr.Tokens) != 1 ||
		!juleapi.IsIgnoreId(expr.Expr.Tokens[0].Kind) {
		cpp.WriteString(expr.String())
		cpp.WriteString(a.Setter.Kind)
	}
	cpp.WriteString(a.Right[0].String())
	return cpp.String()
}

func (a *Assign) hasLeft() bool {
	for _, s := range a.Left {
		if !s.Ignore {
			return true
		}
	}
	return false
}

func (a *Assign) cppMultipleAssign() string {
	var cpp strings.Builder
	if !a.hasLeft() {
		for _, right := range a.Right {
			cpp.WriteString(right.String())
			cpp.WriteByte(';')
		}
		return cpp.String()[:cpp.Len()-1] // Remove last semicolon
	}
	cpp.WriteString(a.cppNewDefines())
	cpp.WriteString("std::tie(")
	var exprCpp strings.Builder
	exprCpp.WriteString("std::make_tuple(")
	for i, left := range a.Left {
		cpp.WriteString(left.String())
		cpp.WriteByte(',')
		exprCpp.WriteString(a.Right[i].String())
		exprCpp.WriteByte(',')
	}
	str := cpp.String()[:cpp.Len()-1] + ")"
	cpp.Reset()
	cpp.WriteString(str)
	cpp.WriteString(a.Setter.Kind)
	cpp.WriteString(exprCpp.String()[:exprCpp.Len()-1] + ")")
	return cpp.String()
}

func (a *Assign) cppMultiRet() string {
	var cpp strings.Builder
	cpp.WriteString(a.cppNewDefines())
	cpp.WriteString("std::tie(")
	for _, left := range a.Left {
		if left.Ignore {
			cpp.WriteString(juleapi.CPP_IGNORE)
			cpp.WriteByte(',')
			continue
		}
		cpp.WriteString(left.String())
		cpp.WriteByte(',')
	}
	str := cpp.String()[:cpp.Len()-1]
	cpp.Reset()
	cpp.WriteString(str)
	cpp.WriteByte(')')
	cpp.WriteString(a.Setter.Kind)
	cpp.WriteString(a.Right[0].String())
	return cpp.String()
}

func (a *Assign) cppNewDefines() string {
	var cpp strings.Builder
	for _, left := range a.Left {
		if left.Ignore || !left.Var.New {
			continue
		}
		cpp.WriteString(left.Var.String() + " ")
	}
	return cpp.String()
}

func (a *Assign) cppPostfix() string {
	var cpp strings.Builder
	cpp.WriteString(a.Left[0].Expr.String())
	cpp.WriteString(a.Setter.Kind)
	return cpp.String()
}

func (a Assign) String() string {
	var cpp strings.Builder
	switch {
	case len(a.Right) == 0:
		cpp.WriteString(a.cppPostfix())
	case a.MultipleRet:
		cpp.WriteString(a.cppMultiRet())
	case len(a.Left) == 1:
		cpp.WriteString(a.cppSingleAssign())
	default:
		cpp.WriteString(a.cppMultipleAssign())
	}
	if !a.IsExpr {
		cpp.WriteByte(';')
	}
	return cpp.String()
}
