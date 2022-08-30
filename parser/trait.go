package parser

import (
	"strings"

	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/pkg/juleapi"
)

type trait struct {
	Ast  *models.Trait
	Defines *DefineMap
	Used bool
	Desc string
}

// FindFunc returns function by id.
// Returns nil if not exist.
func (t *trait) FindFunc(id string) *Fn {
	for _, f := range t.Defines.Funcs {
		if f.Ast.Id == id {
			return f
		}
	}
	return nil
}

// OutId returns juleapi.OutId result of trait.
func (t *trait) OutId() string {
	return juleapi.OutId(t.Ast.Id, t.Ast.Token.File)
}

func (t *trait) String() string {
	var cpp strings.Builder
	cpp.WriteString("struct ")
	outid := t.OutId()
	cpp.WriteString(outid)
	cpp.WriteString(" {\n")
	models.AddIndent()
	is := models.IndentString()
	cpp.WriteString(is)
	cpp.WriteString("virtual ~")
	cpp.WriteString(outid)
	cpp.WriteString("(void) noexcept {}\n\n")
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
