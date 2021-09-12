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

// IsString reports vaule is string representation or not.
func IsString(value string) bool {
	return value[0] == '"'
}

// IsBoolean reports vaule is boolean representation or not.
func IsBoolean(value string) bool {
	return value == "true" || value == "false"
}

// CheckBitInt reports integer is compatible this bit-size or not.
func CheckBitInt(value string, bit int) bool {
	_, err := strconv.ParseInt(value, 10, bit)
	return err == nil
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
	_, e = ast.processExpression(tokens)
	return
}

func (ast *AST) processSingleValuePart(token lex.Token) (result ValueAST) {
	result.Type = NA
	result.Token = token
	switch token.Type {
	case lex.Value:
		if IsString(token.Value) {
			result.Value = token.Value[1 : len(token.Value)-1]
			result.Type = x.String
		} else if IsBoolean(token.Value) {
			result.Value = token.Value
			result.Type = x.Boolean
		} else { // Numeric.
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
			result.Value = token.Value
		}
	}
	return
}

func (ast *AST) processValuePart(tokens []lex.Token) (result ValueAST) {
	if len(tokens) == 1 {
		result = ast.processSingleValuePart(tokens[0])
		if result.Type != NA {
			goto end
		}
	}
	switch token := tokens[len(tokens)-1]; token.Type {
	case lex.Brace:
		switch token.Value {
		case ")":
			return ast.processParenthesesValuePart(tokens)
		}
	default:
		ast.PushErrorToken(tokens[0], "invalid_syntax")
	}
end:
	return
}

func (ast *AST) processParenthesesValuePart(tokens []lex.Token) ValueAST {
	var valueTokens []lex.Token
	j := len(tokens) - 1
	braceCount := 0
	for ; j >= 0; j-- {
		token := tokens[j]
		if token.Type != lex.Brace {
			continue
		}
		switch token.Value {
		case ")":
			braceCount++
		case "(":
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		valueTokens = tokens[:j]
		break
	}
	if len(valueTokens) == 0 && braceCount == 0 {
		tk := tokens[0]
		tokens = tokens[1 : len(tokens)-1]
		if len(tokens) == 0 {
			ast.PushErrorToken(tk, "invalid_syntax")
		}
		value, _ := ast.processExpression(tokens)
		return value
	}
	val := ast.processValuePart(valueTokens)
	switch val.Type {
	default:
		ast.PushErrorToken(tokens[len(valueTokens)], "invalid_syntax")
	}
	return ValueAST{} // Unreachable return.
}

type arithmeticProcess struct {
	ast      *AST
	left     []lex.Token
	leftVal  ValueAST
	right    []lex.Token
	rightVal ValueAST
	operator lex.Token
}

func (p arithmeticProcess) solveString() (value ValueAST) {
	// Not both string?
	if p.leftVal.Type != p.rightVal.Type {
		p.ast.PushErrorToken(p.operator, "invalid_data_types")
		return
	}
	value.Type = x.String
	switch p.operator.Value {
	case "+":
		value.Value = p.leftVal.String() + p.rightVal.String()
	default:
		p.ast.PushErrorToken(p.operator, "operator_notfor_strings")
	}
	return
}

func (p arithmeticProcess) solve() (value ValueAST) {
	switch {
	case p.leftVal.Type == x.Boolean || p.rightVal.Type == x.Boolean:
		p.ast.PushErrorToken(p.operator, "operator_notfor_booleans")
		return
	case p.leftVal.Type == x.String || p.rightVal.Type == x.String:
		return p.solveString()
	}
	if x.IsSignedNumericType(p.leftVal.Type) !=
		x.IsSignedNumericType(p.rightVal.Type) {
		p.ast.PushErrorToken(p.operator, "operator_notfor_uint_and_int")
		return
	}
	// Numeric.
	value.Type = p.leftVal.Type
	if x.TypeGreaterThan(p.rightVal.Type, value.Type) {
		value.Type = p.rightVal.Type
	}
	return
}

func (ast *AST) processExpression(tokens []lex.Token) (ValueAST, ExpressionAST) {
	processes := ast.getExpressionProcesses(tokens)
	if processes == nil {
		return ValueAST{}, ExpressionAST{}
	}
	result := buildExpressionByProcesses(processes)
	if len(processes) == 1 {
		value := ast.processValuePart(processes[0])
		result.Type = value.Type
		return value, result
	}
	var process arithmeticProcess
	var value ValueAST
	process.ast = ast
	j := ast.nextOperator(processes)
	for j != -1 {
		if j == 0 {
			process.leftVal = value
			process.operator = processes[j][0]
			process.right = processes[j+1]
			process.rightVal = ast.processValuePart(process.right)
			value = process.solve()
			processes = processes[2:]
			j = ast.nextOperator(processes)
			continue
		} else if j == len(processes)-1 {
			process.operator = processes[j][0]
			process.left = processes[j-1]
			process.leftVal = ast.processValuePart(process.left)
			process.rightVal = value
			value = process.solve()
			processes = processes[:j-1]
			j = ast.nextOperator(processes)
			continue
		} else if prev := processes[j-1]; prev[0].Type == lex.Operator &&
			len(prev) == 1 {
			process.leftVal = value
			process.operator = processes[j][0]
			process.right = processes[j+1]
			process.rightVal = ast.processValuePart(process.right)
			value = process.solve()
			processes = append(processes[:j], processes[j+2:]...)
			j = ast.nextOperator(processes)
			continue
		}
		process.left = processes[j-1]
		process.leftVal = ast.processValuePart(process.left)
		process.operator = processes[j][0]
		process.right = processes[j+1]
		process.rightVal = ast.processValuePart(process.right)
		solvedValue := process.solve()
		if value.Type != NA {
			process.operator.Value = "+"
			process.right = processes[j+1]
			process.leftVal = value
			process.rightVal = solvedValue
			value = process.solve()
		} else {
			value = solvedValue
		}
		// Remove computed processes.
		processes = append(processes[:j-1], processes[j+2:]...)
		if len(processes) == 1 {
			break
		}
		// Find next operator.
		j = ast.nextOperator(processes)
	}
	result.Type = value.Type
	return value, result
}

func buildExpressionByProcesses(processes [][]lex.Token) ExpressionAST {
	var result ExpressionAST
	for _, part := range processes {
		for _, token := range part {
			switch token.Type {
			case lex.Operator:
				result.Content = append(result.Content, ExpressionNode{
					Content: OperatorAST{
						Token: token,
						Value: token.Value,
					},
					Type: ExpressionNodeOperator,
				})
			case lex.Value:
				result.Content = append(result.Content, ExpressionNode{
					Content: ValueAST{
						Token: token,
						Value: token.Value,
					},
					Type: ExpressionNodeValue,
				})
			case lex.Brace:
				result.Content = append(result.Content, ExpressionNode{
					Content: BraceAST{
						Token: token,
						Value: token.Value,
					},
					Type: ExpressionNodeBrace,
				})
			}
		}
	}
	return result
}

// nextOperator find index of priority operator and returns index of operator
// if found, returns -1 if not.
func (ast *AST) nextOperator(tokens [][]lex.Token) int {
	high, mid, low := -1, -1, -1
	for index, part := range tokens {
		if len(part) != 1 {
			continue
		} else if part[0].Type != lex.Operator {
			continue
		}
		switch part[0].Value {
		case "<<", ">>":
			return index
		case "&", "&^", "%":
			if high == -1 {
				high = index
			}
		case "*", "/", "\\", "|":
			if mid == -1 {
				mid = index
			}
		case "+", "-":
			if low == -1 {
				low = index
			}
		default:
			ast.PushErrorToken(part[0], "invalid_operator")
		}
	}
	if high != -1 {
		return high
	} else if mid != -1 {
		return mid
	}
	return low
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
