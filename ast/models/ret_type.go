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
	if !rt.Type.MultiTyped {
		return nil
	}
	types := rt.Type.Tag.([]DataType)
	var vars []*Var
	for i, tok := range rt.Identifiers {
		if xapi.IsIgnoreId(tok.Kind) {
			continue
		}
		variable := new(Var)
		variable.IdTok = tok
		variable.Id = tok.Kind
		variable.Type = types[i]
		vars = append(vars, variable)
	}
	return vars
}
