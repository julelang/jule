package parser

import (
	"strings"

	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xapi"
)

type function struct {
	Ast          *Func
	Desc         string
	used         bool
	checked      bool
	isEntryPoint bool
}

func (f *function) outId() string {
	if f.isEntryPoint {
		return xapi.OutId(f.Ast.Id, nil)
	}
	return f.Ast.OutId()
}

func (f function) String() string {
	var cpp strings.Builder
	cpp.WriteString(f.Head())
	cpp.WriteByte(' ')
	block := f.Ast.Block
	vars := f.Ast.RetType.Vars()
	if vars != nil {
		statements := make([]models.Statement, len(vars))
		for i, v := range vars {
			statements[i] = models.Statement{Tok: v.Token, Data: *v}
		}
		block.Tree = append(statements, block.Tree...)
	}
	if f.Ast.Receiver != nil && !typeIsPtr(*f.Ast.Receiver) {
		s := f.Ast.Receiver.Tag.(*xstruct)
		self := s.selfVar(*f.Ast.Receiver)
		statements := make([]models.Statement, 1)
		statements[0] = models.Statement{Tok: s.Ast.Tok, Data: self}
		block.Tree = append(statements, block.Tree...)
	}
	cpp.WriteString(block.String())
	return cpp.String()
}

// Head returns declaration head of function.
func (f *function) Head() string {
	var cpp strings.Builder
	cpp.WriteString(f.declHead())
	cpp.WriteString(paramsToCpp(f.Ast.Params))
	return cpp.String()
}

func (f *function) declHead() string {
	var cpp strings.Builder
	cpp.WriteString(genericsToCpp(f.Ast.Generics))
	if cpp.Len() > 0 {
		cpp.WriteByte('\n')
		cpp.WriteString(models.IndentString())
	}
	cpp.WriteString(attributesToString(f.Ast.Attributes))
	cpp.WriteString(f.Ast.RetType.String())
	cpp.WriteByte(' ')
	cpp.WriteString(f.outId())
	return cpp.String()
}

// Prototype returns prototype cpp code of function.
func (f *function) Prototype() string {
	var cpp strings.Builder
	cpp.WriteString(f.declHead())
	cpp.WriteString(f.PrototypeParams())
	cpp.WriteByte(';')
	return cpp.String()
}

// PrototypeParams returns prototype cpp code of function parameters.
func (f *function) PrototypeParams() string {
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
	return kind == x.Attribute_Inline
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
