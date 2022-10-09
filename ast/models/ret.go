package models

import (
	"strings"

	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/pkg/juleapi"
)

// RetType is function return type AST model.
type RetType struct {
	Type        Type
	Identifiers []lex.Token
}

func (rt RetType) String() string {
	return rt.Type.String()
}

// AnyVar reports exist any variable or not.
func (rt *RetType) AnyVar() bool {
	for _, tok := range rt.Identifiers {
		if !juleapi.IsIgnoreId(tok.Kind) {
			return true
		}
	}
	return false
}

// Vars returns variables of ret type if exist, nil if not.
func (rt *RetType) Vars(owner *Block) []*Var {
	get := func(tok lex.Token, t Type) *Var {
		v := new(Var)
		v.Token = tok
		if juleapi.IsIgnoreId(tok.Kind) {
			v.Id = juleapi.IGNORE
		} else {
			v.Id = tok.Kind
		}
		v.Type = t
		v.Owner = owner
		v.Mutable = true
		return v
	}
	if !rt.Type.MultiTyped {
		if len(rt.Identifiers) > 0 {
			v := get(rt.Identifiers[0], rt.Type)
			if v == nil {
				return nil
			}
			return []*Var{v}
		}
		return nil
	}
	var vars []*Var
	types := rt.Type.Tag.([]Type)
	for i, tok := range rt.Identifiers {
		v := get(tok, types[i])
		if v != nil {
			vars = append(vars, v)
		}
	}
	return vars
}

// Ret is return statement AST model.
type Ret struct {
	Token lex.Token
	Expr  Expr
}

func (r Ret) String() string {
	if r.Expr.Model == nil {
		return "return;"
	}
	var cpp strings.Builder
	cpp.WriteString(r.Expr.String())
	cpp.WriteByte(';')
	return cpp.String()
}
