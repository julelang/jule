package documenter

import (
	"encoding/json"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/parser"
)

type xtype struct {
	Id          string `json:"id"`
	Alias       string `json:"alias"`
	Description string `json:"description"`
}

type global struct {
	Id          string `json:"id"`
	Type        string `json:"type"`
	Constant    bool   `json:"constant"`
	Volatile    bool   `json:"volatile"`
	Description string `json:"description"`
}

type function struct {
	Id          string      `json:"id"`
	Ret         string      `json:"ret"`
	Params      []parameter `json:"parameters"`
	Description string      `json:"description"`
	Attributes  []string    `json:"attributes"`
}

type parameter struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Constant bool   `json:"constant"`
	Volatile bool   `json:"volatile"`
}

type document struct {
	Types   []xtype    `json:"types"`
	Globals []global   `json:"globals"`
	Funcs   []function `json:"functions"`
}

func types(p *parser.Parser) []xtype {
	types := make([]xtype, len(p.Defs.Types))
	for i, t := range p.Defs.Types {
		types[i] = xtype{
			Id:          t.Id,
			Alias:       t.Type.Value,
			Description: descriptize(t.Description),
		}
	}
	return types
}

func globals(p *parser.Parser) []global {
	globals := make([]global, len(p.Defs.Globals))
	for i, v := range p.Defs.Globals {
		globals[i] = global{
			Id:          v.Id,
			Type:        v.Type.Value,
			Constant:    v.Const,
			Volatile:    v.Volatile,
			Description: descriptize(v.Description),
		}
	}
	return globals
}

func params(parameters []ast.Parameter) []parameter {
	params := make([]parameter, len(parameters))
	for i, p := range parameters {
		params[i] = parameter{
			Id:       p.Id,
			Type:     p.Type.Value,
			Constant: p.Const,
			Volatile: p.Volatile,
		}
	}
	return params
}

func attributes(attributes []ast.Attribute) []string {
	attrs := make([]string, len(attributes))
	for index, attribute := range attributes {
		attrs[index] = attribute.String()
	}
	return attrs
}

func funcs(p *parser.Parser) []function {
	funcs := make([]function, len(p.Defs.Funcs))
	for i, f := range p.Defs.Funcs {
		fun := function{
			Id:          f.Ast.Id,
			Ret:         f.Ast.RetType.Value,
			Params:      params(f.Ast.Params),
			Description: descriptize(f.Description),
			Attributes:  attributes(f.Attributes),
		}
		funcs[i] = fun
	}
	return funcs
}

// Documentize Parser defines with JSON format.
func Documentize(p *parser.Parser) (string, error) {
	doc := document{types(p), globals(p), funcs(p)}
	bytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	docjson := string(bytes)
	return docjson, nil
}
