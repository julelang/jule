package documenter

import (
	"encoding/json"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/parser"
	"github.com/the-xlang/x/pkg/x"
)

type use struct {
	Path         string `json:"path"`
	StandardPath bool   `json:"standard_path"`
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

type document struct {
	Uses    []use      `json:"uses"`
	Types   []xtype    `json:"types"`
	Globals []global   `json:"globals"`
	Funcs   []function `json:"functions"`
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

func types(p *parser.Parser) []xtype {
	types := make([]xtype, len(p.Defs.Types))
	for i, t := range p.Defs.Types {
		types[i] = xtype{
			Id:    t.Id,
			Alias: t.Type.Val,
			Desc:  descriptize(t.Desc),
		}
	}
	return types
}

func globals(p *parser.Parser) []global {
	globals := make([]global, len(p.Defs.Globals))
	for i, v := range p.Defs.Globals {
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

func params(parameters []ast.Parameter) []parameter {
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

func funcs(p *parser.Parser) []function {
	funcs := make([]function, len(p.Defs.Funcs))
	for i, f := range p.Defs.Funcs {
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

// Documentize Parser defines with JSON format.
func Documentize(p *parser.Parser) (string, error) {
	doc := document{uses(p), types(p), globals(p), funcs(p)}
	bytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	docjson := string(bytes)
	return docjson, nil
}
