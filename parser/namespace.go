package parser

import (
	"strings"

	"github.com/the-xlang/xxc/ast"
	"github.com/the-xlang/xxc/pkg/xapi"
)

type namespace struct {
	Id   string
	Defs *defmap
}

func (ns *namespace) cxxFuncPrototypes() string {
	var cxx strings.Builder
	for _, f := range ns.Defs.Funcs {
		if f.used {
			cxx.WriteString(ast.IndentString())
			cxx.WriteString(f.Prototype())
			cxx.WriteByte('\n')
		}
	}
	return cxx.String()
}

func (ns *namespace) cxxGlobals() string {
	var cxx strings.Builder
	for _, g := range ns.Defs.Globals {
		if g.Used {
			cxx.WriteByte('\n')
			cxx.WriteString(ast.IndentString())
			cxx.WriteString(g.String())
		}
	}
	return cxx.String()
}

func (ns *namespace) cxxFuncs() string {
	var cxx strings.Builder
	for _, f := range ns.Defs.Funcs {
		if f.used {
			cxx.WriteByte('\n')
			cxx.WriteString(ast.IndentString())
			cxx.WriteString(f.String())
		}
	}
	return cxx.String()
}

func (ns *namespace) cxxNamespaces() string {
	var cxx strings.Builder
	for _, n := range ns.Defs.Namespaces {
		cxx.WriteByte('\n')
		cxx.WriteString(ast.IndentString())
		cxx.WriteString(n.String())
	}
	return cxx.String()
}

func (ns namespace) String() string {
	var cxx strings.Builder
	cxx.WriteString("namespace ")
	cxx.WriteString(xapi.AsId(ns.Id))
	cxx.WriteString(" {\n")
	ast.AddIndent()
	cxx.WriteString(ns.cxxFuncPrototypes())
	cxx.WriteString(ns.cxxGlobals())
	cxx.WriteByte('\n')
	cxx.WriteString(ns.cxxFuncs())
	cxx.WriteByte('\n')
	cxx.WriteString(ns.cxxNamespaces())
	ast.DoneIndent()
	cxx.WriteByte('\n')
	cxx.WriteString(ast.IndentString())
	cxx.WriteByte('}')
	return cxx.String()
}
