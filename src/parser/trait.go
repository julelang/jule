package parser

import (
	"strings"

	"github.com/julelang/jule/ast/models"
)

func trait_has_reference_receiver(t *models.Trait) bool {
	for _, f := range t.Defines.Funcs {
		if type_is_ref(f.Receiver.Type) {
			return true
		}
	}
	return false
}

func traitToString(t *models.Trait) string {
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
	for _, f := range t.Funcs {
		cpp.WriteString(is)
		cpp.WriteString("virtual ")
		cpp.WriteString(f.RetType.String())
		cpp.WriteByte(' ')
		cpp.WriteString(f.Id)
		cpp.WriteString(models.ParamsToCpp(f.Params))
		cpp.WriteString(" {")
		if !type_is_void(f.RetType.Type) {
			cpp.WriteString(" return {}; ")
		}
		cpp.WriteString("}\n")
	}
	models.DoneIndent()
	cpp.WriteString("};")
	return cpp.String()
}
