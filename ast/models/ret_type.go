package models

import "github.com/the-xlang/xxc/pkg/xapi"

// RetType is function return type AST model.
type RetType struct {
	Type        DataType
	Identifiers Toks
}

func (rt RetType) String() string {
	return rt.Type.String()
}

// AnyVar reports exist any variable or not.
func (rt *RetType) AnyVar() bool {
	for _, tok := range rt.Identifiers {
		if !xapi.IsIgnoreId(tok.Kind) {
			return true
		}
	}
	return false
}

// Vars returns variables of ret type if exist, nil if not.
func (rt *RetType) Vars() []*Var {
	get := func(tok Tok, t DataType) *Var {
		if xapi.IsIgnoreId(tok.Kind) {
			return nil
		}
		v := new(Var)
		v.IdTok = tok
		v.Id = tok.Kind
		v.Type = t
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
	types := rt.Type.Tag.([]DataType)
	for i, tok := range rt.Identifiers {
		v := get(tok, types[i])
		if v != nil {
			vars = append(vars, v)
		}
	}
	return vars
}
