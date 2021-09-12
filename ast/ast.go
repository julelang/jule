package ast

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
	"github.com/the-xlang/x/pkg/xbits"
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
		"%s:%d:%d %s", token.File.Path, token.Line, token.Column, message))
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
		case lex.Fun:
			ast.BuildFunction()
		default:
			ast.PushError("invalid_syntax")
			ast.Position++
		}
	}
}

// BuildFunction builds AST model of function.
func (ast *AST) BuildFunction() {
	ast.Position++ // Skip function keyword.
	var fn FunctionAST
	fn.Token = ast.Tokens[ast.Position]
	fn.Name = fn.Token.Value
	fn.ReturnType.Type = x.Void
	ast.Position++
	tokens := ast.getRange("(", ")")
	if tokens == nil {
		return
	} else if len(tokens) > 0 {
		ast.BuildParameters(&fn, tokens)
	}
	if ast.Ended() {
		ast.Position--
		ast.PushError("function_body_not_exist")
		ast.Position = -1 // Stop parsing.
		return
	}
	token := ast.Tokens[ast.Position]
	if token.Type == lex.Type {
		fn.ReturnType = ast.BuildType(token)
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
	fn.Block = ast.BuildBlock(blockTokens)
	ast.Tree = append(ast.Tree, Object{
		Token: fn.Token,
		Type:  Statement,
		Value: StatementAST{
			Token: fn.Token,
			Type:  StatementFunction,
			Value: fn,
		},
	})
}

// BuildParameters builds AST model of function parameters.
func (ast *AST) BuildParameters(fn *FunctionAST, tokens []lex.Token) {
	last := 0
	braceCount := 0
	for index, token := range tokens {
		if token.Type == lex.Brace {
			switch token.Value {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 || token.Type != lex.Comma {
			continue
		}
		ast.pushParameter(fn, tokens[last:index], token)
		last = index + 1
	}
	if last < len(tokens) {
		if last == 0 {
			ast.pushParameter(fn, tokens[last:], tokens[last])
		} else {
			ast.pushParameter(fn, tokens[last:], tokens[last-1])
		}
	}
}

func (ast *AST) pushParameter(fn *FunctionAST, tokens []lex.Token, err lex.Token) {
	if len(tokens) == 0 {
		ast.PushErrorToken(err, "invalid_syntax")
		return
	}
	nameToken := tokens[0]
	if nameToken.Type != lex.Name {
		ast.PushErrorToken(nameToken, "invalid_syntax")
	}
	if len(tokens) < 2 {
		ast.PushErrorToken(nameToken, "type_missing")
		return
	}
	for _, param := range fn.Params {
		if param.Name == nameToken.Value {
			ast.PushErrorToken(nameToken, "parameter_exist")
			break
		}
	}
	fn.Params = append(fn.Params, ParameterAST{
		Token: nameToken,
		Name:  nameToken.Value,
		Type:  ast.BuildType(tokens[1]),
	})
}

// IsStatement reports token is
// statement finish point or not.
func IsStatement(token lex.Token) bool {
	return token.Type == lex.SemiColon
}

// BuildType builds AST model of type.
func (ast *AST) BuildType(token lex.Token) (t TypeAST) {
	if token.Type != lex.Type {
		ast.PushErrorToken(token, "invalid_type")
		return
	}
	t.Type = x.TypeFromName(token.Value)
	t.Value = token.Value
	return t
}

// IsSigleOperator is returns true
// if operator is unary or smilar to unary,
// returns false if not.
func IsSingleOperator(operator string) bool {
	return operator == "-" ||
		operator == "!" ||
		operator == "*" ||
		operator == "&"
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
		if braceCount > 0 || !IsStatement(token) {
			continue
		}
		if index-oldStatementPoint == 0 {
			continue
		}
		b.Content = append(b.Content,
			ast.BuildStatement(tokens[oldStatementPoint:index]))
		if ast.Position == -1 {
			break
		}
		oldStatementPoint = index + 1
	}
	return
}

// BuildStatement builds AST model of statement.
func (ast *AST) BuildStatement(tokens []lex.Token) (s StatementAST) {
	firstToken := tokens[0]
	switch firstToken.Type {
	case lex.Name:
		return ast.BuildNameStatement(tokens)
	case lex.Return:
		return ast.BuildReturnStatement(tokens)
	default:
		ast.PushErrorToken(firstToken, "invalid_syntax")
	}
	return
}

// BuildReturnStatement builds AST model of return statement.
func (ast *AST) BuildNameStatement(tokens []lex.Token) (s StatementAST) {
	if len(tokens) == 1 {
		ast.PushErrorToken(tokens[0], "invalid_syntax")
		return
	}
	switch tokens[1].Type {
	case lex.Brace:
		switch tokens[1].Value {
		case "(":
			return ast.BuildFunctionCallStatement(tokens)
		}
	}
	ast.PushErrorToken(tokens[0], "invalid_syntax")
	return
}

func (ast *AST) BuildFunctionCallStatement(tokens []lex.Token) StatementAST {
	var fnCall FunctionCallAST
	fnCall.Token = tokens[0]
	fnCall.Name = fnCall.Token.Value
	tokens = tokens[1:]
	fnCall.Args = ast.getRangeTokens("(", ")", tokens)
	if fnCall.Args == nil {
		ast.Position = -1
		return StatementAST{}
	} else if len(fnCall.Args) != len(tokens)-2 {
		ast.PushErrorToken(tokens[len(tokens)-2], "invalid_syntax")
		ast.Position = -1
		return StatementAST{}
	}
	return StatementAST{
		Token: fnCall.Token,
		Value: fnCall,
		Type:  StatementFunctionCall,
	}
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
	e.Processes = ast.getExpressionProcesses(tokens)
	e.Tokens = tokens
	return
}

func (ast *AST) getExpressionProcesses(tokens []lex.Token) [][]lex.Token {
	var processes [][]lex.Token
	var part []lex.Token
	operator := false
	value := false
	braceCount := 0
	pushedError := false
	for index, token := range tokens {
		switch token.Type {
		case lex.Operator:
			if !operator {
				if IsSingleOperator(token.Value) {
					part = append(part, token)
					continue
				}
				ast.PushErrorToken(token, "operator_overflow")
			}
			operator = false
			value = true
			if braceCount > 0 {
				part = append(part, token)
				continue
			}
			processes = append(processes, part)
			processes = append(processes, []lex.Token{token})
			part = []lex.Token{}
			continue
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
		operator = requireOperatorForProcess(token, index, len(tokens))
		value = false
	}
	if len(part) > 0 {
		processes = append(processes, part)
	}
	if value {
		ast.PushErrorToken(processes[len(processes)-1][0], "operator_overflow")
		pushedError = true
	}
	if pushedError {
		return nil
	}
	return processes
}

func requireOperatorForProcess(token lex.Token, index, tokensLen int) bool {
	switch token.Type {
	case lex.Comma:
		return false
	case lex.Brace:
		if token.Value == "[" ||
			token.Value == "(" ||
			token.Value == "{" {
			return false
		}
	}
	return index < tokensLen-1
}

func (ast *AST) checkExpressionToken(token lex.Token) {
	if token.Value[0] >= '0' && token.Value[0] <= '9' {
		var result bool
		if strings.IndexByte(token.Value, '.') != -1 {
			_, result = new(big.Float).SetString(token.Value)
		} else {
			result = xbits.CheckBitInt(token.Value, 64)
		}
		if !result {
			ast.PushErrorToken(token, "invalid_numeric_range")
		}
	}
}

func (ast *AST) getRangeTokens(open, close string, tokens []lex.Token) []lex.Token {
	braceCount := 0
	start := 1
	for index, token := range tokens {
		if token.Type != lex.Brace {
			continue
		}
		if token.Value == open {
			braceCount++
		} else if token.Value == close {
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		return tokens[start:index]
	}
	ast.PushErrorToken(tokens[0], "brace_not_closed")
	return nil
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
