package ast

import "github.com/the-xlang/x/lex"

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
	Content []Object
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
