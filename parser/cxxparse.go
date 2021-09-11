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

// PushErrorToken appends new error by token.
func (cp *CxxParser) PushErrorToken(token lex.Token, err string) {
	message := x.Errors[err]
	cp.PFI.Errors = append(cp.PFI.Errors, fmt.Sprintf(
		"%s:%d %s", token.File.Path, token.Line, message))
}

// PushError appends new error.
func (cp *CxxParser) PushError(err string) {
	cp.PFI.Errors = append(cp.PFI.Errors, x.Errors[err])
}

// String is return full C++ code of parsed objects.
func (cp CxxParser) String() string {
	var sb strings.Builder
	for _, function := range cp.Functions {
		sb.WriteString(function.String())
		sb.WriteString("\n\n")
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
			cp.PushErrorToken(model.Token, "invalid_syntax")
		}
	}
	cp.finalCheck()
}

// ParseStatement parse X statement to C++ code.
func (cp *CxxParser) ParseStatement(s ast.StatementAST) {
	switch s.Type {
	case ast.StatementFunction:
		cp.ParseFunction(s.Value.(ast.FunctionAST))
	default:
		cp.PushErrorToken(s.Token, "invalid_syntax")
	}
}

// ParseFunction parse X function to C++ code.
func (cp *CxxParser) ParseFunction(fnAst ast.FunctionAST) {
	if function := cp.functionByBName(fnAst.Name); function != nil {
		cp.PushErrorToken(fnAst.Token, "exist_name")
		return
	}
	fn := new(Function)
	fn.Token = fnAst.Token
	fn.Name = fnAst.Name
	fn.ReturnType = fnAst.ReturnType.Type
	fn.Block = fnAst.Block
	cp.checkFunctionReturn(fn)
	cp.Functions = append(cp.Functions, fn)
}

func (cp *CxxParser) checkFunctionReturn(fn *Function) {
	if fn.ReturnType == x.Void {
		return
	}
	miss := true
	for _, s := range fn.Block.Content {
		if s.Type == ast.StatementReturn {
			if !x.TypesAreCompatible(
				s.Value.(ast.ReturnAST).Expression.Type, fn.ReturnType) {
				cp.PushErrorToken(s.Token, "incompatible_type")
			}
			miss = false
		}
	}
	if miss {
		cp.PushErrorToken(fn.Token, "missing_return")
	}
}

func (cp *CxxParser) functionByBName(name string) *Function {
	for _, function := range cp.Functions {
		if function.Name == name {
			return function
		}
	}
	return nil
}

func (cp *CxxParser) finalCheck() {
	if cp.functionByBName(x.EntryPoint) == nil {
		cp.PushError("no_entry_point")
	}
}
