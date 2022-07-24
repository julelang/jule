package parser

import (
	"strings"

	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/pkg/xapi"
)

type trait struct {
	Ast  *models.Trait
	Defs *Defmap
	Used bool
	Desc string
}

// FindFunc returns function by id.
// Returns nil if not exist.
func (t *trait) FindFunc(id string) *Func {
	for _, f := range t.Ast.Funcs {
		if f.Id == id {
			return f
		}
	}
	return nil
}

// OutId returns xapi.OutId result of trait.
func (t *trait) OutId() string {
	return xapi.OutId(t.Ast.Id, t.Ast.Tok.File)
}

func (t *trait) String() string {
	var cxx strings.Builder
	cxx.WriteString("struct ")
	cxx.WriteString(t.OutId())
	cxx.WriteString(" {\n")
	models.AddIndent()
	is := models.IndentString()
	for _, f := range t.Ast.Funcs {
		cxx.WriteString(is)
		cxx.WriteString("virtual ")
		cxx.WriteString(f.RetType.String())
		cxx.WriteByte(' ')
		cxx.WriteString(f.Id)
		cxx.WriteString(paramsToCxx(f.Params))
		cxx.WriteString(" = 0;\n")
	}
	models.DoneIndent()
	cxx.WriteString("};")
	return cxx.String()
}
