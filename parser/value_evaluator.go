package parser

import (
	"strings"

	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xbits"
	"github.com/the-xlang/xxc/pkg/xtype"
)

func toRawStrLiteral(literal string) string {
	literal = literal[1 : len(literal)-1] // Remove bounds
	literal = `"(` + literal + `)"`
	literal = xapi.ToRawStr(literal)
	return literal
}

func toCharLiteral(kind string) (string, bool) {
	kind = kind[1 : len(kind)-1]
	isByte := false
	switch {
	case len(kind) == 1 && kind[0] <= 255:
		isByte = true
	case kind[0] == '\\' && kind[1] == 'x':
		isByte = true
	case kind[0] == '\\' && kind[1] >= '0' && kind[1] <= '7':
		isByte = true
	}
	kind = "'" + kind + "'"
	return xapi.ToChar(kind), isByte
}

type valueEvaluator struct {
	tok   Tok
	model *exprModel
	p     *Parser
}

func (p *valueEvaluator) str() value {
	var v value
	v.ast.Data = p.tok.Kind
	v.ast.Type.Id = xtype.Str
	v.ast.Type.Val = tokens.STR
	if israwstr(p.tok.Kind) {
		p.model.appendSubNode(exprNode{toRawStrLiteral(p.tok.Kind)})
	} else {
		p.model.appendSubNode(exprNode{xapi.ToStr(p.tok.Kind)})
	}
	return v
}

func (ve *valueEvaluator) char() value {
	var v value
	v.ast.Data = ve.tok.Kind
	literal, _ := toCharLiteral(ve.tok.Kind)
	v.ast.Type.Id = xtype.Char
	v.ast.Type.Val = tokens.CHAR
	ve.model.appendSubNode(exprNode{literal})
	return v
}

func (ve *valueEvaluator) bool() value {
	var v value
	v.ast.Data = ve.tok.Kind
	v.ast.Type.Id = xtype.Bool
	v.ast.Type.Val = tokens.BOOL
	ve.model.appendSubNode(exprNode{ve.tok.Kind})
	return v
}

func (ve *valueEvaluator) nil() value {
	var v value
	v.ast.Data = ve.tok.Kind
	v.ast.Type.Id = xtype.Nil
	v.ast.Type.Val = xtype.NilTypeStr
	ve.model.appendSubNode(exprNode{ve.tok.Kind})
	return v
}

func (ve *valueEvaluator) num() value {
	var v value
	v.ast.Data = ve.tok.Kind
	if strings.Contains(ve.tok.Kind, tokens.DOT) ||
		strings.ContainsAny(ve.tok.Kind, "eE") {
		v.ast.Type.Id = xtype.F64
		v.ast.Type.Val = tokens.F64
	} else {
		intbit := xbits.BitsizeType(xtype.Int)
		switch {
		case xbits.CheckBitInt(ve.tok.Kind, intbit):
			v.ast.Type.Id = xtype.Int
			v.ast.Type.Val = tokens.INT
		case intbit < xbits.MaxInt && xbits.CheckBitInt(ve.tok.Kind, xbits.MaxInt):
			v.ast.Type.Id = xtype.I64
			v.ast.Type.Val = tokens.I64
		default:
			v.ast.Type.Id = xtype.U64
			v.ast.Type.Val = tokens.U64
		}
	}
	node := exprNode{xtype.CxxTypeIdFromType(v.ast.Type.Id) + "{" + ve.tok.Kind + "}"}
	ve.model.appendSubNode(node)
	return v
}

func (ve *valueEvaluator) varId(id string, variable *Var) (v value) {
	variable.Used = true
	v.ast.Data = id
	v.ast.Type = variable.Type
	v.constant = variable.Const
	v.volatile = variable.Volatile
	v.ast.Tok = variable.IdTok
	v.lvalue = true
	// If built-in.
	if variable.IdTok.Id == tokens.NA {
		ve.model.appendSubNode(exprNode{xapi.OutId(id, nil)})
	} else {
		ve.model.appendSubNode(exprNode{xapi.OutId(id, variable.IdTok.File)})
	}
	return
}

func (ve *valueEvaluator) funcId(id string, f *function) (v value) {
	f.used = true
	v.ast.Data = id
	v.ast.Type.Id = xtype.Func
	v.ast.Type.Tag = f.Ast
	v.ast.Type.Val = f.Ast.DataTypeString()
	v.ast.Tok = f.Ast.Tok
	ve.model.appendSubNode(exprNode{f.outId()})
	return
}

func (ve *valueEvaluator) enumId(id string, e *Enum) (v value) {
	e.Used = true
	v.ast.Data = id
	v.ast.Type.Id = xtype.Enum
	v.ast.Type.Tag = e
	v.ast.Type.Val = e.Id
	v.ast.Tok = e.Tok
	v.constant = true
	v.isType = true
	// If built-in.
	if e.Tok.Id == tokens.NA {
		ve.model.appendSubNode(exprNode{xapi.OutId(id, nil)})
	} else {
		ve.model.appendSubNode(exprNode{xapi.OutId(id, e.Tok.File)})
	}
	return
}

func (ve *valueEvaluator) structId(id string, s *xstruct) (v value) {
	s.Used = true
	v.ast.Data = id
	v.ast.Type.Id = xtype.Struct
	v.ast.Type.Tag = s
	v.ast.Type.Val = s.Ast.Id
	v.ast.Type.Tok = s.Ast.Tok
	v.ast.Tok = s.Ast.Tok
	v.isType = true
	// If built-in.
	if s.Ast.Tok.Id == tokens.NA {
		ve.model.appendSubNode(exprNode{xapi.OutId(id, nil)})
	} else {
		ve.model.appendSubNode(exprNode{xapi.OutId(id, s.Ast.Tok.File)})
	}
	return
}

func (ve *valueEvaluator) id() (_ value, ok bool) {
	id := ve.tok.Kind
	if variable, _ := ve.p.varById(id); variable != nil {
		return ve.varId(id, variable), true
	} else if f, _, _ := ve.p.FuncById(id); f != nil {
		return ve.funcId(id, f), true
	} else if e, _, _ := ve.p.enumById(id); e != nil {
		return ve.enumId(id, e), true
	} else if s, _, _ := ve.p.structById(id); s != nil {
		return ve.structId(id, s), true
	} else {
		ve.p.pusherrtok(ve.tok, "id_noexist", id)
	}
	return
}
