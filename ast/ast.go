package ast

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

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
		returnModel.Expression = ast.BuildExpression(tokens[1:])
	}
	return StatementAST{
		Token: returnModel.Token,
		Type:  StatementReturn,
		Value: returnModel,
	}
}

// BuildExpression builds AST model of expression.
func (ast *AST) BuildExpression(tokens []lex.Token) (e ExpressionAST) {
	processes := ast.getExpressionProcesses(tokens)
	if len(processes) == 1 {
		value := ast.processExpression(tokens)
		e.Content = append(e.Content, ExpressionNode{
			Content: value,
			Type:    ExpressionNodeValue,
		})
		e.Type = value.Type
		return
	}
	return
}

// IsString reports vaule is string representation or not.
func IsString(value string) bool {
	return value[0] == '"'
}

// IsBoolean reports vaule is boolean representation or not.
func IsBoolean(value string) bool {
	return value == "true" || value == "false"
}

func (ast *AST) processSingleValuePart(token lex.Token) (result ValueAST) {
	result.Type = NA
	result.Token = token
	switch token.Type {
	case lex.Value:
		if IsString(token.Value) {
			result.Data = token.Value[1 : len(token.Value)-1]
			result.Type = x.String
		} else if IsBoolean(token.Value) {
			result.Data = token.Value
			result.Type = x.Boolean
		}
		// Numeric.
		if strings.Contains(token.Value, ".") ||
			strings.ContainsAny(token.Value, "eE") {
			result.Type = x.Float64
		} else {
			result.Type = x.Int32
			ok := CheckBitInt(token.Value, 32)
			if !ok {
				result.Type = x.Int64
			}
		}
		result.Data = token.Value
	}
	return
}

func (ast *AST) processExpression(tokens []lex.Token) (result ValueAST) {
	if len(tokens) == 1 {
		result = ast.processSingleValuePart(tokens[0])
		if result.Type != NA {
			goto end
		}
	}
	ast.PushErrorToken(tokens[0], "invalid_syntax")
end:
	return
}

func (ast *AST) getExpressionProcesses(tokens []lex.Token) [][]lex.Token {
	var processes [][]lex.Token
	var part []lex.Token
	braceCount := 0
	pushedError := false
	for index, token := range tokens {
		switch token.Type {
		case lex.Brace:
			switch token.Value {
			case "(", "[", "{":
				braceCount++
			default:
				braceCount--
			}
		}
		if index > 0 {
			lt := tokens[index-1]
			if (lt.Type == lex.Name || lt.Type == lex.Value) &&
				(token.Type == lex.Name || token.Type == lex.Value) {
				ast.PushErrorToken(token, "invalid_syntax")
				pushedError = true
			}
		}
		ast.checkExpressionToken(token)
		part = append(part, token)
	}
	if len(part) != 0 {
		processes = append(processes, part)
	}
	if pushedError {
		return nil
	}
	return processes
}

// CheckBitInt reports integer is compatible this bit-size or not.
func CheckBitInt(value string, bit int) bool {
	_, err := strconv.ParseInt(value, 10, bit)
	return err == nil
}

func (ast *AST) checkExpressionToken(token lex.Token) {
	if token.Value[0] >= '0' && token.Value[0] <= '9' {
		var result bool
		if strings.IndexByte(token.Value, '.') != -1 {
			_, result = new(big.Float).SetString(token.Value)
		} else {
			result = CheckBitInt(token.Value, 64)
		}
		if !result {
			ast.PushErrorToken(token, "invalid_numeric_range")
		}
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
