package documenter

import (
	"encoding/json"

	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/parser"
	"github.com/the-xlang/xxc/pkg/xtype"
)

type Defmap = parser.Defmap

type generic struct {
	Id string
}

type use struct {
	Path   string `json:"path"`
	Stdlib bool   `json:"stdlib"`
}

type xstruct struct {
	Id     string     `json:"id"`
	Desc   string     `json:"description"`
	Fields []global   `json:"fields"`
	Funcs  []function `json:"functions"`
}

type enum struct {
	Id    string   `json:"id"`
	Desc  string   `json:"description"`
	Items []string `json:"items"`
}

type type_alias struct {
	Id    string `json:"id"`
	Alias string `json:"alias"`
	Desc  string `json:"description"`
}

type global struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Constant bool   `json:"constant"`
	Desc     string `json:"description"`
}

type function struct {
	Id         string      `json:"id"`
	Ret        string      `json:"ret"`
	Generics   []generic   `json:"generics"`
	Params     []parameter `json:"parameters"`
	Desc       string      `json:"description"`
	Attributes []string    `json:"attributes"`
}

type parameter struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

type document struct {
	Uses    []use        `json:"uses"`
	Enums   []enum       `json:"enums"`
	Structs []xstruct    `json:"structs"`
	Types   []type_alias `json:"types"`
	Globals []global     `json:"globals"`
	Funcs   []function   `json:"functions"`
}

func ttoa(t models.DataType) string {
	if t.Kind == xtype.TypeMap[xtype.Void] {
		return ""
	}
	return t.Kind
}

func uses(p *parser.Parser) []use {
	uses := make([]use, len(p.Uses))
	for i, u := range p.Uses {
		uses[i] = use{
			Path:   u.LinkString,
			Stdlib: u.LinkString[0] != '"',
		}
	}
	return uses
}

func enums(dm *Defmap) []enum {
	enums := make([]enum, len(dm.Enums))
	for i, e := range dm.Enums {
		var conv enum
		conv.Id = e.Id
		conv.Desc = Descriptize(e.Desc)
		conv.Items = make([]string, len(e.Items))
		for i, item := range e.Items {
			conv.Items[i] = item.Id
		}
		enums[i] = conv
	}
	return enums
}

func structs(dm *Defmap) []xstruct {
	structs := make([]xstruct, len(dm.Structs))
	for i, s := range dm.Structs {
		var xs xstruct
		xs.Id = s.Ast.Id
		xs.Desc = Descriptize(s.Desc)
		xs.Fields = globals(s.Defs)
		xs.Funcs = funcs(s.Defs)
		structs[i] = xs
	}
	return structs
}

func types(dm *Defmap) []type_alias {
	types := make([]type_alias, len(dm.Types))
	for i, t := range dm.Types {
		types[i] = type_alias{
			Id:    t.Id,
			Alias: ttoa(t.Type),
			Desc:  Descriptize(t.Desc),
		}
	}
	return types
}

func globals(dm *Defmap) []global {
	globals := make([]global, len(dm.Globals))
	for i, v := range dm.Globals {
		globals[i] = global{
			Id:       v.Id,
			Type:     ttoa(v.Type),
			Constant: v.Const,
			Desc:     Descriptize(v.Desc),
		}
	}
	return globals
}

func params(parameters []models.Param) []parameter {
	params := make([]parameter, len(parameters))
	for i, p := range parameters {
		params[i] = parameter{
			Id:   p.Id,
			Type: ttoa(p.Type),
		}
	}
	return params
}

func attributes(attributes []models.Attribute) []string {
	attrs := make([]string, len(attributes))
	for i, attr := range attributes {
		attrs[i] = attr.String()
	}
	return attrs
}

func generics(genericTypes []*models.GenericType) []generic {
	generics := make([]generic, len(genericTypes))
	for i, gt := range genericTypes {
		var g generic
		g.Id = gt.Id
		generics[i] = g
	}
	return generics
}

func funcs(dm *Defmap) []function {
	funcs := make([]function, len(dm.Funcs))
	for i, f := range dm.Funcs {
		fun := function{
			Id:         f.Ast.Id,
			Ret:        ttoa(f.Ast.RetType.Type),
			Generics:   generics(f.Ast.Generics),
			Params:     params(f.Ast.Params),
			Desc:       Descriptize(f.Desc),
			Attributes: attributes(f.Ast.Attributes),
		}
		funcs[i] = fun
	}
	return funcs
}

// Doc returns documentation of code into JSON format.
func Doc(p *parser.Parser) (string, error) {
	doc := document{
		uses(p),
		enums(p.Defs),
		structs(p.Defs),
		types(p.Defs),
		globals(p.Defs),
		funcs(p.Defs),
	}
	bytes, err := json.MarshalIndent(doc, "", "\t")
	if err != nil {
		return "", err
	}
	docjson := string(bytes)
	return docjson, nil
}
