package documenter

import (
	"encoding/json"
	"strings"
	"unicode"

	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/parser"
	"github.com/jule-lang/jule/pkg/juletype"
)

type Defmap = parser.DefineMap

type generic struct {
	Id string
}

type use struct {
	Path   string `json:"path"`
	Stdlib bool   `json:"stdlib"`
}

type jtrait struct {
	Id     string     `json:"id"`
	Desc   string     `json:"description"`
	Funcs  []function `json:"functions"`
}

type structure struct {
	Id     string              `json:"id"`
	Desc   string              `json:"description"`
	Fields []global            `json:"fields"`
	Funcs  []function          `json:"functions"`
	ImplementedTraits []string `json:"implemented_traits"`
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
	Traits  []jtrait     `json:"traits"`
	Structs []structure  `json:"structs"`
	Types   []type_alias `json:"types"`
	Globals []global     `json:"globals"`
	Funcs   []function   `json:"functions"`
}

func fmt_json_ttoa(t models.Type) string {
	if t.Kind == juletype.TypeMap[juletype.Void] {
		return ""
	}
	return t.Kind
}

func fmt_json_uses(p *parser.Parser) []use {
	uses := make([]use, len(p.Uses))
	for i, u := range p.Uses {
		uses[i] = use{
			Path:   u.LinkString,
			Stdlib: u.LinkString[0] != '"',
		}
	}
	return uses
}

func fmt_json_enums(dm *Defmap) []enum {
	enums := make([]enum, len(dm.Enums))
	for i, e := range dm.Enums {
		var conv enum
		conv.Id = e.Id
		conv.Desc = fmt_json_doc_comment(e.Desc)
		conv.Items = make([]string, len(e.Items))
		for i, item := range e.Items {
			conv.Items[i] = item.Id
		}
		enums[i] = conv
	}
	return enums
}

func fmt_json_traits(dm *Defmap) []jtrait {
	traits := make([]jtrait, len(dm.Traits))
	for i, e := range dm.Traits {
		var t jtrait
		t.Id = e.Ast.Id
		t.Desc = fmt_json_doc_comment(e.Desc)
		t.Funcs = make([]function, len(e.Ast.Funcs))
		for i, f := range e.Ast.Funcs {
			t.Funcs[i] = function{
				Id: f.Id,
				Attributes: fmt_json_attributes(f.Attributes),
				Ret: fmt_json_ttoa(f.RetType.Type),
				Generics: fmt_json_generics(f.Generics),
				Params: fmt_json_params(f.Params),
				
			}
		}
		traits[i] = t
	}
	return traits
}

func fmt_json_structs(dm *Defmap) []structure {
	structs := make([]structure, len(dm.Structs))
	for i, s := range dm.Structs {
		var ss structure
		ss.Id = s.Ast.Id
		ss.Desc = fmt_json_doc_comment(s.Description)
		ss.Fields = fmt_json_globals(s.Defines)
		ss.Funcs = fmt_json_funcs(s.Defines)
		ss.ImplementedTraits = make([]string, len(*s.Traits))
		for i, t := range *s.Traits {
			ss.ImplementedTraits[i] = t.Ast.Id
		}
		structs[i] = ss
	}
	return structs
}

func fmt_json_type_aliases(dm *Defmap) []type_alias {
	types := make([]type_alias, len(dm.Types))
	for i, t := range dm.Types {
		types[i] = type_alias{
			Id:    t.Id,
			Alias: fmt_json_ttoa(t.Type),
			Desc:  fmt_json_doc_comment(t.Desc),
		}
	}
	return types
}

func fmt_json_globals(dm *Defmap) []global {
	globals := make([]global, len(dm.Globals))
	for i, v := range dm.Globals {
		globals[i] = global{
			Id:       v.Id,
			Type:     fmt_json_ttoa(v.Type),
			Constant: v.Const,
			Desc:     fmt_json_doc_comment(v.Desc),
		}
	}
	return globals
}

func fmt_json_params(parameters []models.Param) []parameter {
	params := make([]parameter, len(parameters))
	for i, p := range parameters {
		params[i] = parameter{
			Id:   p.Id,
			Type: fmt_json_ttoa(p.Type),
		}
	}
	return params
}

func fmt_json_attributes(attributes []models.Attribute) []string {
	attrs := make([]string, len(attributes))
	for i, attr := range attributes {
		attrs[i] = attr.String()
	}
	return attrs
}

func fmt_json_generics(genericTypes []*models.GenericType) []generic {
	generics := make([]generic, len(genericTypes))
	for i, gt := range genericTypes {
		var g generic
		g.Id = gt.Id
		generics[i] = g
	}
	return generics
}

func fmt_json_funcs(dm *Defmap) []function {
	funcs := make([]function, len(dm.Funcs))
	for i, f := range dm.Funcs {
		fun := function{
			Id:         f.Ast.Id,
			Ret:        fmt_json_ttoa(f.Ast.RetType.Type),
			Generics:   fmt_json_generics(f.Ast.Generics),
			Params:     fmt_json_params(f.Ast.Params),
			Desc:       fmt_json_doc_comment(f.Desc),
			Attributes: fmt_json_attributes(f.Ast.Attributes),
		}
		funcs[i] = fun
	}
	return funcs
}

func doc_fmt_json(p *parser.Parser) (string, error) {
	doc := document{
		fmt_json_uses(p),
		fmt_json_enums(p.Defines),
		fmt_json_traits(p.Defines),
		fmt_json_structs(p.Defines),
		fmt_json_type_aliases(p.Defines),
		fmt_json_globals(p.Defines),
		fmt_json_funcs(p.Defines),
	}
	bytes, err := json.MarshalIndent(doc, "", "\t")
	if err != nil {
		return "", err
	}
	docjson := string(bytes)
	return docjson, nil
}

// fmt_json_doc_comment is ready decription string to process for json format.
func fmt_json_doc_comment(s string) string {
	var doc strings.Builder
	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	s = strings.ReplaceAll(s, "\n", " ")
	doc.WriteString(s)
	return doc.String()
}
