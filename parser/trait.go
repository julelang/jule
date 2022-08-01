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
func (t *trait) FindFunc(id string) *function {
	for _, f := range t.Defs.Funcs {
		if f.Ast.Id == id {
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
	var cpp strings.Builder
	cpp.WriteString("struct ")
	cpp.WriteString(t.OutId())
	cpp.WriteString(" {\n")
	models.AddIndent()
	is := models.IndentString()
	for _, f := range t.Ast.Funcs {
		cpp.WriteString(is)
		cpp.WriteString("virtual ")
		cpp.WriteString(f.RetType.String())
		cpp.WriteByte(' ')
		cpp.WriteString(f.Id)
		cpp.WriteString(paramsToCpp(f.Params))
		cpp.WriteString(" = 0;\n")
	}
	models.DoneIndent()
	cpp.WriteString("};")
	return cpp.String()
}
