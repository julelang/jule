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

func (f *function) getTracePointStatements() []models.Statement {
	var trace strings.Builder
	trace.WriteString(`___trace.push(`)
	var tracepoint strings.Builder
	tracepoint.WriteString(f.Ast.Id)
	tracepoint.WriteString(f.Ast.DataTypeString())
	tracepoint.WriteString("\n\t")
	tracepoint.WriteString(f.Ast.Tok.File.Path())
	trace.WriteString(xapi.ToStr([]byte(tracepoint.String())))
	trace.WriteByte(')')
	statements := []models.Statement{{}, {}}
	statements[0].Data = models.ExprStatement{
		Expr: models.Expr{Model: exprNode{trace.String()}},
	}
	statements[1].Data = models.ExprStatement{
		Expr: models.Expr{Model: exprNode{"DEFER(___trace.ok())"}},
	}
	return statements
}

func (f function) String() string {
	var cxx strings.Builder
	cxx.WriteString(f.Head())
	cxx.WriteByte(' ')
	block := f.Ast.Block
	vars := f.Ast.RetType.Vars()
	if vars != nil {
		statements := make([]models.Statement, len(vars))
		for i, v := range vars {
			statements[i] = models.Statement{Tok: v.IdTok, Data: *v}
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
	block.Tree = append(f.getTracePointStatements(), block.Tree...)
	cxx.WriteString(block.String())
	return cxx.String()
}

// Head returns declaration head of function.
func (f *function) Head() string {
	var cxx strings.Builder
	cxx.WriteString(f.declHead())
	cxx.WriteString(paramsToCxx(f.Ast.Params))
	return cxx.String()
}

func (f *function) declHead() string {
	var cxx strings.Builder
	cxx.WriteString(genericsToCxx(f.Ast.Generics))
	if cxx.Len() > 0 {
		cxx.WriteByte('\n')
		cxx.WriteString(models.IndentString())
	}
	cxx.WriteString(attributesToString(f.Ast.Attributes))
	cxx.WriteString(f.Ast.RetType.String())
	cxx.WriteByte(' ')
	cxx.WriteString(f.outId())
	return cxx.String()
}

// Prototype returns prototype cxx code of function.
func (f *function) Prototype() string {
	var cxx strings.Builder
	cxx.WriteString(f.declHead())
	cxx.WriteString(f.PrototypeParams())
	cxx.WriteByte(';')
	return cxx.String()
}

// PrototypeParams returns prototype cxx code of function parameters.
func (f *function) PrototypeParams() string {
	if len(f.Ast.Params) == 0 {
		return "(void)"
	}
	var cxx strings.Builder
	cxx.WriteByte('(')
	for _, p := range f.Ast.Params {
		cxx.WriteString(p.Prototype())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ")"
}

func isOutableAttribute(kind string) bool {
	return kind == x.Attribute_Inline
}

func attributesToString(attributes []models.Attribute) string {
	var cxx strings.Builder
	for _, attr := range attributes {
		if isOutableAttribute(attr.Tag.Kind) {
			cxx.WriteString(attr.String())
			cxx.WriteByte(' ')
		}
	}
	return cxx.String()
}

func paramsToCxx(params []Param) string {
	if len(params) == 0 {
		return "(void)"
	}
	var cxx strings.Builder
	cxx.WriteByte('(')
	for _, p := range params {
		cxx.WriteString(p.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ")"
}
