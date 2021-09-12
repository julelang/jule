package ast

import (
	"strings"

	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
)

// Object is an element of AST.
type Object struct {
	Token lex.Token
	Value interface{}
	Type  uint8
}

// IdentifierAST is identifier.
type IdentifierAST struct {
	Type  uint8
	Value string
}

// StatementAST is statement.
type StatementAST struct {
	Token lex.Token
	Type  uint8
	Value interface{}
}

// RangeAST represents block range or etc.
type RangeAST struct {
	Type    uint8
	Content []Object
}

// BlockAST is code block.
type BlockAST struct {
	Content []StatementAST
}

// TypeAST is data type identifier.
type TypeAST struct {
	Type  uint8
	Value string
}

// FunctionAST is function declaration AST model.
type FunctionAST struct {
	Token      lex.Token
	Name       string
	Params     []ParameterAST
	ReturnType TypeAST
	Block      BlockAST
}

// ParameterAST is function parameter AST model.
type ParameterAST struct {
	Token lex.Token
	Name  string
	Type  TypeAST
}

func (p ParameterAST) String() string {
	return x.CxxTypeNameFromType(p.Type.Type) + " " + p.Name
}

// FunctionAST is function declaration AST model.
type FunctionCallAST struct {
	Token lex.Token
	Name  string
	Args  []lex.Token
}

func (fc FunctionCallAST) String() string {
	var sb strings.Builder
	sb.WriteString(fc.Name)
	sb.WriteByte('(')
	sb.WriteString(tokensToString(fc.Args))
	sb.WriteByte(')')
	return sb.String()
}

// ExpressionAST is AST model of expression.
type ExpressionAST struct {
	Tokens    []lex.Token
	Processes [][]lex.Token
}

func (e ExpressionAST) String() string {
	return tokensToString(e.Tokens)
}

// ValueAST is AST model of constant value.
type ValueAST struct {
	Token lex.Token
	Value string
	Type  uint8
}

func (v ValueAST) String() string {
	return v.Value
}

// BraceAST is AST model of brace.
type BraceAST struct {
	Token lex.Token
	Value string
}

func (b BraceAST) String() string {
	return b.Value
}

// OperatorAST is AST model of operator.
type OperatorAST struct {
	Token lex.Token
	Value string
}

func (o OperatorAST) String() string {
	return o.Value
}

// ReturnAST is return statement AST model.
type ReturnAST struct {
	Token      lex.Token
	Expression ExpressionAST
}

func (r ReturnAST) String() string {
	return r.Token.Value + " " + r.Expression.String()
}

func tokensToString(tokens []lex.Token) string {
	var sb strings.Builder
	for _, token := range tokens {
		sb.WriteString(token.Value)
		if token.Type != lex.Brace &&
			token.Type != lex.Name {
			sb.WriteByte(' ')
		}
	}
	return sb.String()
}
