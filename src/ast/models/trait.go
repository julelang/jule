package models

import (
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/pkg/juleapi"
)

// Trait is the AST model of traits.
type Trait struct {
	Pub     bool
	Token   lex.Token
	Id      string
	Desc    string
	Used    bool
	Funcs   []*Fn
	Defines *Defmap
}

// FindFunc returns function by id.
// Returns nil if not exist.
func (t *Trait) FindFunc(id string) *Fn {
	for _, f := range t.Defines.Fns {
		if f.Id == id {
			return f
		}
	}
	return nil
}

// OutId returns juleapi.OutId result of trait.
func (t *Trait) OutId() string {
	return juleapi.OutId(t.Id, t.Token.File.Addr())
}
