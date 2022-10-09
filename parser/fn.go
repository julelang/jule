package parser

import (
	"strings"

	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/pkg/juleapi"
)

type Fn struct {
	Ast            *Func
	Desc           string
	used           bool
	checked        bool
	isEntryPoint   bool
}

func (f *Fn) outId() string {
	if f.isEntryPoint {
		return juleapi.OutId(f.Ast.Id, nil)
	}
	return f.Ast.OutId()
}

func fnBlockToString(vars []*Var, b *models.Block) string {
	var cpp strings.Builder
	if vars != nil {
		statements := make([]models.Statement, len(vars))
		for i, v := range vars {
			statements[i] = models.Statement{Token: v.Token, Data: *v}
		}
		b.Tree = append(statements, b.Tree...)
	}
	cpp.WriteString(b.String())
	return cpp.String()
}

func (f Fn) stringOwner(owner string) string {
	var cpp strings.Builder
	cpp.WriteString(f.Head(owner))
	cpp.WriteByte(' ')
	vars := f.Ast.RetType.Vars(f.Ast.Block)
	cpp.WriteString(fnBlockToString(vars, f.Ast.Block))
	return cpp.String()
}

func (f Fn) String() string {
	return f.stringOwner("")
}

// Head returns declaration head of function.
func (f *Fn) Head(owner string) string {
	var cpp strings.Builder
	cpp.WriteString(f.declHead(owner))
	cpp.WriteString(paramsToCpp(f.Ast.Params))
	return cpp.String()
}

func (f *Fn) declHead(owner string) string {
	var cpp strings.Builder
	cpp.WriteString(genericsToCpp(f.Ast.Generics))
	if cpp.Len() > 0 {
		cpp.WriteByte('\n')
		cpp.WriteString(models.IndentString())
	}
	if !f.isEntryPoint {
		cpp.WriteString("inline ")
	}
	cpp.WriteString(attributesToString(f.Ast.Attributes))
	cpp.WriteString(f.Ast.RetType.String())
	cpp.WriteByte(' ')
	if owner != "" {
		cpp.WriteString(owner)
		cpp.WriteString(lex.KND_DBLCOLON)
	}
	cpp.WriteString(f.outId())
	return cpp.String()
}

// Prototype returns prototype cpp code of function.
func (f *Fn) Prototype(owner string) string {
	var cpp strings.Builder
	cpp.WriteString(f.declHead(owner))
	cpp.WriteString(f.PrototypeParams())
	cpp.WriteByte(';')
	return cpp.String()
}

// PrototypeParams returns prototype cpp code of function parameters.
func (f *Fn) PrototypeParams() string {
	if len(f.Ast.Params) == 0 {
		return "(void)"
	}
	var cpp strings.Builder
	cpp.WriteByte('(')
	for _, p := range f.Ast.Params {
		cpp.WriteString(p.Prototype())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ")"
}

func isOutableAttribute(kind string) bool {
	return false
}

func attributesToString(attributes []models.Attribute) string {
	var cpp strings.Builder
	for _, attr := range attributes {
		if isOutableAttribute(attr.Tag) {
			cpp.WriteString(attr.String())
			cpp.WriteByte(' ')
		}
	}
	return cpp.String()
}

func paramsToCpp(params []Param) string {
	if len(params) == 0 {
		return "(void)"
	}
	var cpp strings.Builder
	cpp.WriteByte('(')
	for _, p := range params {
		cpp.WriteString(p.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ")"
}

func is_constructor(f *Func) bool {
	if !typeIsStruct(f.RetType.Type) {
		return false
	}
	s := f.RetType.Type.Tag.(*structure)
	return s.Ast.Id == f.Id
}
