package parser

import (
	"strings"

	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/pkg/xapi"
)

type namespace struct {
	Id   string
	Tok  Tok
	Defs *Defmap
}

func (ns namespace) String() string {
	var cxx strings.Builder
	cxx.WriteString("namespace ")
	cxx.WriteString(xapi.OutId(ns.Id, ns.Tok.File))
	cxx.WriteString(" {\n")
	models.AddIndent()
	cxx.WriteString(cxxPrototypes(ns.Defs))
	cxx.WriteString(cxxTraits(ns.Defs))
	cxx.WriteString(cxxStructs(ns.Defs))
	cxx.WriteString(cxxGlobals(ns.Defs))
	cxx.WriteByte('\n')
	cxx.WriteString(cxxFuncs(ns.Defs))
	cxx.WriteByte('\n')
	cxx.WriteString(cxxNamespaces(ns.Defs))
	models.DoneIndent()
	cxx.WriteByte('\n')
	cxx.WriteString(models.IndentString())
	cxx.WriteByte('}')
	return cxx.String()
}
