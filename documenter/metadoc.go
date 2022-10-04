package documenter

import (
	"strings"
	"unicode"

	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/pkg/juleset"
	"github.com/jule-lang/jule/pkg/juletype"
	"github.com/jule-lang/jule/parser"
)

func fmt_meta_ttoa(t models.Type) string {
	if t.Kind == juletype.TypeMap[juletype.Void] {
		return ""
	}
	t.Pure = false
	t.Generic = false
	t.SetToOriginal()
	return ": " + t.Kind
}

func fmt_meta_uses(p *parser.Parser) string {
	var meta strings.Builder
	for _, u := range p.Uses {
		meta.WriteString("use ")
		meta.WriteString(u.LinkString)
		switch {
		case u.FullUse:
			meta.WriteString("::*")
		case len(u.Selectors) > 0:
			meta.WriteString("::{")
			n := len(u.Selectors)
			for i, s := range u.Selectors {
				meta.WriteString(s.Kind)
				if i+1 < n {
					meta.WriteByte(',')
				}
			}
			meta.WriteByte('}')
		}
		meta.WriteByte('\n')
	}
	return meta.String()
}

func fmt_meta_enums(dm *Defmap) string {
	var meta strings.Builder
	for _, e := range dm.Enums {
		meta.WriteString(fmt_meta_doc_comment(e.Desc))
		meta.WriteString(fmt_meta_pub_identifier(e.Pub))
		meta.WriteString("enum ")
		meta.WriteString(e.Id)
		meta.WriteString(fmt_meta_ttoa(e.Type))
		meta.WriteString(" {\n")
		indent := strings.Repeat(juleset.Default.Indent, juleset.Default.IndentCount)
		for _, item := range e.Items {
			meta.WriteString(indent)
			meta.WriteString(item.Id)
			meta.WriteString(fmt_meta_assign_expr(item.Expr.Tokens))
			meta.WriteString(",\n")
		}
		meta.WriteString("}\n\n")
	}
	return meta.String()
}

func fmt_meta_traits(dm *Defmap) string {
	var meta strings.Builder
	for _, t := range dm.Traits {
		meta.WriteString(fmt_meta_doc_comment(t.Desc))
		meta.WriteString(fmt_meta_pub_identifier(t.Ast.Pub))
		meta.WriteString("trait ")
		meta.WriteString(t.Ast.Id)
		meta.WriteString(" {\n")
		indent := strings.Repeat(juleset.Default.Indent, juleset.Default.IndentCount)
		for _, f := range t.Ast.Funcs {
			meta.WriteString(indent)
			ff := parser.Fn{
				Ast: f,
			}
			meta.WriteString(fmt_meta_func(&ff))
			meta.WriteString("\n")
		}
		meta.WriteString("}\n\n")
	}
	return meta.String()
}

func fmt_meta_structs(dm *Defmap) string {
	var meta strings.Builder
	for _, s := range dm.Structs {
		meta.WriteString(fmt_meta_doc_comment(s.Description))
		meta.WriteString(fmt_meta_generics(s.Ast.Generics))
		meta.WriteString(fmt_meta_pub_identifier(s.Ast.Pub))
		meta.WriteString("struct ")
		meta.WriteString(s.Ast.Id)
		meta.WriteString(" {\n")
		indent := strings.Repeat(juleset.Default.Indent, juleset.Default.IndentCount)
		for _, f := range s.Ast.Fields {
			meta.WriteString(indent)
			meta.WriteString(fmt_meta_pub_identifier(f.Pub))
			meta.WriteString(f.Id)
			meta.WriteString(fmt_meta_ttoa(f.Type))
			meta.WriteString(fmt_meta_assign_expr(f.Expr.Tokens))
			meta.WriteString("\n")
		}
		meta.WriteString("}\n\n")
		if len(s.Defines.Funcs) > 0 {
			meta.WriteString("impl ")
			meta.WriteString(s.Ast.Id)
			meta.WriteString(" {\n\n")
			meta.WriteString(fmt_meta_funcs(s.Defines))
			meta.WriteString("}\n\n")
		}
		if len(*s.Traits) > 0 {
			for _, t := range *s.Traits {
				meta.WriteString("impl ")
				meta.WriteString(t.Ast.Id)
				meta.WriteString(" for ")
				meta.WriteString(s.Ast.Id)
				meta.WriteByte('\n')
			}
			meta.WriteByte('\n')
		}
	}
	return meta.String()
}

func fmt_meta_type_aliases(dm *Defmap) string {
	var meta strings.Builder
	for _, t := range dm.Types {
		meta.WriteString(fmt_meta_doc_comment(t.Desc))
		meta.WriteString(fmt_meta_pub_identifier(t.Pub))
		meta.WriteString("type ")
		meta.WriteString(t.Id)
		meta.WriteString(fmt_meta_ttoa(t.Type))
		meta.WriteByte('\n')
	}
	return meta.String()
}

func fmt_meta_globals(dm *Defmap) string {
	var meta strings.Builder
	for _, g := range dm.Globals {
		meta.WriteString(fmt_meta_doc_comment(g.Desc))
		meta.WriteString(fmt_meta_pub_identifier(g.Pub))
		if g.Const {
			meta.WriteString("const ")
		} else {
			meta.WriteString("let ")
		}
		meta.WriteString(g.Id)
		meta.WriteString(fmt_meta_ttoa(g.Type))
		meta.WriteString(fmt_meta_assign_expr(g.Expr.Tokens))
		meta.WriteByte('\n')
	}
	return meta.String()
}

func fmt_meta_func(f *parser.Fn) string {
	var meta strings.Builder
	meta.WriteString(fmt_meta_doc_comment(f.Desc))
	meta.WriteString(fmt_meta_attributes(f.Ast.Attributes))
	meta.WriteString(fmt_meta_generics(f.Ast.Generics))
	meta.WriteString(fmt_meta_pub_identifier(f.Ast.Pub))
	meta.WriteString("fn ")
	n := len(f.Ast.Params)
	meta.WriteString(f.Ast.Id)
	meta.WriteByte('(')
	if f.Ast.Receiver != nil {
		meta.WriteString(f.Ast.Receiver.ReceiverTypeString())
		if n > 0 {
			meta.WriteByte(',')
		}
	}
	for i, p := range f.Ast.Params {
		meta.WriteString(p.Id)
		meta.WriteString(": ")
		if p.Variadic {
			meta.WriteString("...")
		}
		meta.WriteString(fmt_meta_ttoa(p.Type)[2:])
		if i+1 < n {
			meta.WriteByte(',')
		}
	}
	meta.WriteByte(')')
	rt := fmt_meta_ttoa(f.Ast.RetType.Type)
	if rt != "" {
		meta.WriteString(" " + rt[2:])
	}
	return meta.String()
}

func fmt_meta_funcs(dm *Defmap) string {
	var meta strings.Builder
	for _, f := range dm.Funcs {
		meta.WriteString(fmt_meta_func(f))
		meta.WriteString("\n\n")
	}
	return meta.String()
}

func doc_fmt_meta(p *parser.Parser) (string, error) {
	var meta strings.Builder
	meta.WriteString(fmt_meta_uses(p))
	meta.WriteByte('\n')
	meta.WriteString(fmt_meta_enums(p.Defines))
	meta.WriteString(fmt_meta_traits(p.Defines))
	meta.WriteString(fmt_meta_structs(p.Defines))
	meta.WriteString(fmt_meta_type_aliases(p.Defines))
	meta.WriteByte('\n')
	meta.WriteString(fmt_meta_funcs(p.Defines))
	meta.WriteString(fmt_meta_globals(p.Defines))
	return meta.String(), nil
}

// fmt_meta_doc_comment is ready decription string to process for meta format.
func fmt_meta_doc_comment(s string) string {
	if s == "" {
		return ""
	}
	var doc strings.Builder
	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	s = strings.ReplaceAll(s, "\n", "\n// ")
	doc.WriteString("// ")
	doc.WriteString(s[:len(s)-len("\n//")])
	return doc.String()
}

func fmt_meta_generics(generics []*models.GenericType) string {
	if len(generics) == 0 {
		return ""
	}
	var meta strings.Builder
	meta.WriteString("type[")
	n := len(generics)
	for i, g := range generics {
		meta.WriteString(g.Id)
		if i+1 < n {
			meta.WriteByte(',')
		}
	}
	meta.WriteString("]\n")
	return meta.String()
}

func fmt_meta_pub_identifier(is_pub bool) string {
	if is_pub {
		return "pub "
	}
	return ""
}

func fmt_meta_assign_expr(toks []lex.Token) string {
	if len(toks) == 0 {
		return ""
	}
	var meta strings.Builder
	meta.WriteString(" = ")
	for _, t := range toks {
		meta.WriteString(t.Kind)
	}
	return meta.String()
}

func fmt_meta_attributes(attrs []models.Attribute) string {
	if len(attrs) == 0 {
		return ""
	}
	var meta strings.Builder
	for _, attr := range attrs {
		meta.WriteString("//jule:")
		meta.WriteString(attr.Tag)
		meta.WriteByte('\n')
	}
	return meta.String()
}
