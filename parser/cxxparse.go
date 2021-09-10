package parser

import (
	"fmt"
	"strings"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
)

// CxxParser.
type CxxParser struct {
	Functions []*Function

	Tokens []lex.Token
	PFI    *ParseFileInfo
}

// NewParser returns new instance of CxxParser.
func NewParser(tokens []lex.Token, PFI *ParseFileInfo) *CxxParser {
	parser := new(CxxParser)
	parser.Tokens = tokens
	parser.PFI = PFI
	return parser
}

// PushError is appends new error by parser fields.
func (cp *CxxParser) PushError(token lex.Token, err string) {
	message := x.Errors[err]
	cp.PFI.Errors = append(cp.PFI.Errors, fmt.Sprintf(
		"%s %s: %d", token.File.Path, message, token.Line))
}

// String is return full C++ code of parsed objects.
func (cp CxxParser) String() string {
	var sb strings.Builder
	for _, function := range cp.Functions {
		sb.WriteString(function.String())
	}
	return sb.String()
}

// Parse is parse X code to C++ code.
//
//! This function is main point of parsing.
func (cp *CxxParser) Parse() {
	astModel := ast.New(cp.Tokens)
	astModel.Build()
	if astModel.Errors != nil {
		cp.PFI.Errors = append(cp.PFI.Errors, astModel.Errors...)
		return
	}
	for _, model := range astModel.Tree {
		switch model.Type {
		case ast.Statement:
			cp.ParseStatement(model.Value.(ast.StatementAST))
		default:
			cp.PushError(model.Token, "invalid_syntax")
		}
	}
}

// ParseStatement parse X statement to C++ code.
func (cp *CxxParser) ParseStatement(s ast.StatementAST) {
	switch s.Type {
	case ast.StatementFunction:
		cp.ParseFunction(s.Value.(ast.FunctionAST))
	default:
		cp.PushError(s.Token, "invalid_syntax")
	}
}

// ParseFunction parse X function to C++ code.
func (cp *CxxParser) ParseFunction(f ast.FunctionAST) {
	function := new(Function)
	function.Name = f.Name
	function.Line = f.Token.Line
	function.FILE = f.Token.File
	function.ReturnType = f.ReturnType.Type
	cp.Functions = append(cp.Functions, function)
}
