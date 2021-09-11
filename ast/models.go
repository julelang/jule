package ast

import (
	"fmt"
	"strings"

	"github.com/the-xlang/x/lex"
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
	ReturnType TypeAST
	Block      BlockAST
}

// ExpressionAST is AST model of expression.
type ExpressionAST struct {
	Content []ExpressionNode
	Type    uint8
}

func (e ExpressionAST) String() string {
	var sb strings.Builder
	for _, node := range e.Content {
		sb.WriteString(node.String() + " ")
	}
	return sb.String()[:sb.Len()-1]
}

// ExpressionNode is value model.
type ExpressionNode struct {
	Content interface{}
	Type    uint8
}

func (n ExpressionNode) String() string {
	return fmt.Sprint(n.Content)
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
	if r.Expression.Type != NA {
		return r.Token.Value + " " + r.Expression.String()
	}
	return r.Token.Value
}
