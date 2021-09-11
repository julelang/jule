package ast

import (
	"fmt"

	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
)

// AST processor.
type AST struct {
	Tree     []Object
	Errors   []string
	Tokens   []lex.Token
	Position int
}

// New AST instance.
func New(tokens []lex.Token) *AST {
	ast := new(AST)
	ast.Tokens = tokens
	ast.Position = 0
	return ast
}

// PushErrorToken appends error by specified token.
func (ast *AST) PushErrorToken(token lex.Token, err string) {
	message := x.Errors[err]
	ast.Errors = append(ast.Errors, fmt.Sprintf(
		"%s:%d %s", token.File.Path, token.Line, message))
}

// PushError appends error by current token.
func (ast *AST) PushError(err string) {
	ast.PushErrorToken(ast.Tokens[ast.Position], err)
}

// Ended reports position is at end of tokens or not.
func (ast *AST) Ended() bool {
	return ast.Position >= len(ast.Tokens)
}

// Build builds AST tree.
//
//! This function is main point of AST build.
func (ast *AST) Build() {
	for ast.Position != -1 && !ast.Ended() {
		firstToken := ast.Tokens[ast.Position]
		switch firstToken.Type {
		case lex.Name:
			ast.processName()
		default:
			ast.PushError("invalid_syntax")
		}
	}
}

// ParseFunction parse X function to C++ code.
func (ast *AST) BuildFunction() {
	var function FunctionAST
	function.Token = ast.Tokens[ast.Position]
	function.Name = function.Token.Value
	function.ReturnType.Type = x.Void
	// Skip function parentheses.
	//! Fix here at after.
	ast.Position++
	parameters := ast.getRange("(", ")")
	if parameters == nil {
		return
	} else if len(parameters) > 0 {
		ast.PushError("parameters_not_supported")
	}
	if ast.Ended() {
		ast.Position--
		ast.PushError("function_body_not_exist")
		ast.Position = -1 // Stop parsing.
		return
	}
	token := ast.Tokens[ast.Position]
	if token.Type == lex.Type {
		function.ReturnType.Type = x.TypeFromName(token.Value)
		function.ReturnType.Value = token.Value
		ast.Position++
		if ast.Ended() {
			ast.Position--
			ast.PushError("function_body_not_exist")
			ast.Position = -1 // Stop parsing.
			return
		}
		token = ast.Tokens[ast.Position]
	}
	if token.Type != lex.Brace || token.Value != "{" {
		ast.PushError("invalid_syntax")
		ast.Position = -1 // Stop parsing.
		return
	}
	blockTokens := ast.getRange("{", "}")
	if blockTokens == nil {
		ast.PushError("function_body_not_exist")
		ast.Position = -1
		return
	}
	function.Block = ast.BuildBlock(blockTokens)
	ast.Tree = append(ast.Tree, Object{
		Token: function.Token,
		Type:  Statement,
		Value: StatementAST{
			Token: function.Token,
			Type:  StatementFunction,
			Value: function,
		},
	})
}

// IsStatement reports token is
// statement finish point or not.
func IsStatement(before, current lex.Token) bool {
	return current.Type == lex.SemiColon || before.Line < current.Line
}

// BuildBlock builds AST model of statements of code block.
func (ast *AST) BuildBlock(tokens []lex.Token) (b BlockAST) {
	braceCount := 0
	oldStatementPoint := 0
	for index, token := range tokens {
		if token.Type == lex.Brace {
			if token.Value == "{" {
				braceCount++
			} else {
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		}
		if index < len(tokens)-1 {
			if index == 0 && !IsStatement(token, token) {
				continue
			} else if index > 0 && !IsStatement(tokens[index-1], token) {
				continue
			}
		}
		if token.Type != lex.SemiColon {
			index++
		}
		if index-oldStatementPoint == 0 {
			continue
		}
		b.Content = append(b.Content,
			ast.BuildStatement(tokens[oldStatementPoint:index]))
		oldStatementPoint = index + 1
	}
	return
}

// BuildStatement builds AST model of statement.
func (ast *AST) BuildStatement(tokens []lex.Token) (s StatementAST) {
	firstToken := tokens[0]
	switch firstToken.Type {
	case lex.Return:
		return ast.BuildReturnStatement(tokens)
	default:
		ast.PushErrorToken(firstToken, "invalid_syntax")
	}
	return
}

// BuildReturnStatement builds AST model of return statement.
func (ast *AST) BuildReturnStatement(tokens []lex.Token) StatementAST {
	var returnModel ReturnAST
	returnModel.Token = tokens[0]
	if len(tokens) > 1 {
		ast.PushErrorToken(tokens[1], "not_support_expression")
	}
	return StatementAST{
		Token: returnModel.Token,
		Type:  StatementReturn,
		Value: returnModel,
	}
}

func (ast *AST) processName() {
	ast.Position++
	if ast.Ended() {
		ast.Position--
		ast.PushError("invalid_syntax")
		return
	}
	ast.Position--
	secondToken := ast.Tokens[ast.Position+1]
	switch secondToken.Type {
	case lex.Brace:
		switch secondToken.Value {
		case "(":
			ast.BuildFunction()
		default:
			ast.PushError("invalid_syntax")
		}
	}
}

func (ast *AST) getRange(open, close string) []lex.Token {
	token := ast.Tokens[ast.Position]
	if token.Type == lex.Brace && token.Value == open {
		ast.Position++
		braceCount := 1
		start := ast.Position
		for ; braceCount > 0 && !ast.Ended(); ast.Position++ {
			token := ast.Tokens[ast.Position]
			if token.Type != lex.Brace {
				continue
			}
			if token.Value == open {
				braceCount++
			} else if token.Value == close {
				braceCount--
			}
		}
		if braceCount > 0 {
			ast.Position--
			ast.PushError("brace_not_closed")
			ast.Position = -1
			return nil
		}
		return ast.Tokens[start : ast.Position-1]
	}
	return nil
}
