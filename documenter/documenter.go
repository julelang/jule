package documenter

import (
	"encoding/json"

	"github.com/the-xlang/xxc/ast"
	"github.com/the-xlang/xxc/parser"
	"github.com/the-xlang/xxc/pkg/x"
)

type Defmap = parser.Defmap

type use struct {
	Path         string `json:"path"`
	StandardPath bool   `json:"standard_path"`
}

type enum struct {
	Id    string   `json:"id"`
	Desc  string   `json:"description"`
	Items []string `json:"items"`
}

type xtype struct {
	Id    string `json:"id"`
	Alias string `json:"alias"`
	Desc  string `json:"description"`
}

type global struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Constant bool   `json:"constant"`
	Volatile bool   `json:"volatile"`
	Desc     string `json:"description"`
}

type function struct {
	Id         string      `json:"id"`
	Ret        string      `json:"ret"`
	Params     []parameter `json:"parameters"`
	Desc       string      `json:"description"`
	Attributes []string    `json:"attributes"`
}

type parameter struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Constant bool   `json:"constant"`
	Volatile bool   `json:"volatile"`
}

type namespace struct {
	Id         string      `json:"id"`
	Enums      []enum      `json:"enums"`
	Types      []xtype     `json:"types"`
	Globals    []global    `json:"globals"`
	Funcs      []function  `json:"functions"`
	Namespaces []namespace `json:"namespaces"`
}

type document struct {
	Uses       []use       `json:"uses"`
	Enums      []enum      `json:"enums"`
	Types      []xtype     `json:"types"`
	Globals    []global    `json:"globals"`
	Funcs      []function  `json:"functions"`
	Namespaces []namespace `json:"namespaces"`
}

func uses(p *parser.Parser) []use {
	uses := make([]use, len(p.Uses))
	for i, u := range p.Uses {
		path := u.Path
		path = path[len(x.StdlibPath)+1:]
		uses[i] = use{
			Path:         path,
			StandardPath: true,
		}
	}
	return uses
}

func enums(dm *Defmap) []enum {
	enums := make([]enum, len(dm.Enums))
	for i, e := range dm.Enums {
		var conv enum
		conv.Id = e.Id
		conv.Desc = e.Desc
		conv.Items = make([]string, len(e.Items))
		for i, item := range e.Items {
			conv.Items[i] = item.Id
		}
		enums[i] = conv
	}
	return enums
}

func types(dm *Defmap) []xtype {
	types := make([]xtype, len(dm.Types))
	for i, t := range dm.Types {
		types[i] = xtype{
			Id:    t.Id,
			Alias: t.Type.Val,
			Desc:  descriptize(t.Desc),
		}
	}
	return types
}

func globals(dm *Defmap) []global {
	globals := make([]global, len(dm.Globals))
	for i, v := range dm.Globals {
		globals[i] = global{
			Id:       v.Id,
			Type:     v.Type.Val,
			Constant: v.Const,
			Volatile: v.Volatile,
			Desc:     descriptize(v.Desc),
		}
	}
	return globals
}

func params(parameters []ast.Param) []parameter {
	params := make([]parameter, len(parameters))
	for i, p := range parameters {
		params[i] = parameter{
			Id:       p.Id,
			Type:     p.Type.Val,
			Constant: p.Const,
			Volatile: p.Volatile,
		}
	}
	return params
}

func attributes(attributes []ast.Attribute) []string {
	attrs := make([]string, len(attributes))
	for i, attr := range attributes {
		attrs[i] = attr.String()
	}
	return attrs
}

func funcs(dm *Defmap) []function {
	funcs := make([]function, len(dm.Funcs))
	for i, f := range dm.Funcs {
		fun := function{
			Id:         f.Ast.Id,
			Ret:        f.Ast.RetType.Val,
			Params:     params(f.Ast.Params),
			Desc:       descriptize(f.Desc),
			Attributes: attributes(f.Attributes),
		}
		funcs[i] = fun
	}
	return funcs
}

func makeNamespace(dm *Defmap) namespace {
	var ns namespace
	ns.Enums = enums(dm)
	ns.Types = types(dm)
	ns.Globals = globals(dm)
	ns.Funcs = funcs(dm)
	ns.Namespaces = namespaces(dm)
	return ns
}

func namespaces(dm *Defmap) []namespace {
	namespaces := make([]namespace, len(dm.Namespaces))
	for i, ns := range dm.Namespaces {
		nspace := makeNamespace(ns.Defs)
		nspace.Id = ns.Id
		namespaces[i] = nspace
	}
	return namespaces
}

// Documentize Parser defines with JSON format.
func Documentize(p *parser.Parser) (string, error) {
	doc := document{
		uses(p),
		enums(p.Defs),
		types(p.Defs),
		globals(p.Defs),
		funcs(p.Defs),
		namespaces(p.Defs),
	}
	bytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	docjson := string(bytes)
	return docjson, nil
}
