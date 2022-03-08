package documenter

import (
	"encoding/json"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/parser"
)

type xtype struct {
	Id    string `json:"id"`
	Alias string `json:"alias"`
}

type variable struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Constant bool   `json:"constant"`
	Volatile bool   `json:"volatile"`
}

type function struct {
	Id     string      `json:"id"`
	Ret    string      `json:"ret"`
	Params []parameter `json:"parameters"`
}

type parameter struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Constant bool   `json:"constant"`
	Volatile bool   `json:"volatile"`
}

type document struct {
	Types   []xtype    `json:"types"`
	Globals []variable `json:"globals"`
	Funcs   []function `json:"functions"`
}

func types(p *parser.Parser) []xtype {
	types := make([]xtype, len(p.Types))
	for i, t := range p.Types {
		types[i] = xtype{t.Id, t.Type.Value}
	}
	return types
}

func globals(p *parser.Parser) []variable {
	globals := make([]variable, len(p.GlobalVars))
	for i, v := range p.GlobalVars {
		globals[i] = variable{v.Id, v.Type.Value, v.Const, v.Volatile}
	}
	return globals
}

func params(parameters []ast.Parameter) []parameter {
	params := make([]parameter, len(parameters))
	for i, p := range parameters {
		params[i] = parameter{p.Id, p.Type.Value, p.Const, p.Volatile}
	}
	return params
}

func funcs(p *parser.Parser) []function {
	funcs := make([]function, len(p.Funcs))
	for i, f := range p.Funcs {
		fun := function{}
		fun.Id = f.Ast.Id
		fun.Ret = f.Ast.RetType.Value
		fun.Params = params(f.Ast.Params)
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
