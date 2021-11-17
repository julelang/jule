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
func (ast *AST) PushError(err string) { ast.PushErrorToken(ast.Tokens[ast.Position], err) }

// Ended reports position is at end of tokens or not.
func (ast *AST) Ended() bool { return ast.Position >= len(ast.Tokens) }

// Build builds AST tree.
//
//! This function is main point of AST build.
func (ast *AST) Build() {
	for ast.Position != -1 && !ast.Ended() {
		firstToken := ast.Tokens[ast.Position]
		switch firstToken.Type {
		case lex.Brace:
			ast.BuildBrace()
		case lex.Name:
			ast.BuildName()
		case lex.Var, lex.Const:
			ast.BuildGlobalVariable()
		case lex.Type:
			ast.BuildType()
		default:
			ast.PushError("invalid_syntax")
			ast.Position++
		}
	}
}

// BuildType builds AST model of type defination statement.
func (ast *AST) BuildType() {
	position := 1 // Initialize value is 1 for skip keyword.
	tokens := ast.skipStatement()
	if position >= len(tokens) {
		ast.PushErrorToken(tokens[position-1], "invalid_syntax")
		return
	}
	token := tokens[position]
	if token.Type != lex.Name {
		ast.PushErrorToken(token, "invalid_syntax")
	}
	position++
	if position >= len(tokens) {
		ast.PushErrorToken(tokens[position-1], "invalid_syntax")
		return
	}
	destinationType, _ := ast.BuildDataType(tokens[position:], new(int), true)
	ast.Tree = append(ast.Tree, Object{
		Token: tokens[1],
		Value: TypeAST{
			Token: tokens[1],
			Name:  tokens[1].Value,
			Type:  destinationType,
		},
	})
}

// BuildName builds AST model of global name statement.
func (ast *AST) BuildName() {
	ast.Position++
	if ast.Ended() {
		ast.PushErrorToken(ast.Tokens[ast.Position-1], "invalid_syntax")
		return
	}
	token := ast.Tokens[ast.Position]
	ast.Position--
	switch token.Type {
	case lex.Brace:
		switch token.Value {
		case "(":
			funAST := ast.BuildFunction(false)
			ast.Tree = append(ast.Tree, Object{
				Token: funAST.Token,
				Value: StatementAST{
					Token: funAST.Token,
					Value: funAST,
				},
			})
			return
		}
	}
	ast.Position++
	ast.PushErrorToken(token, "invalid_syntax")
}

// BuildBrace builds AST model by global brace statement.
func (ast *AST) BuildBrace() {
	token := ast.Tokens[ast.Position]
	switch token.Value {
	case "[":
		ast.BuildAttribute()
	default:
		ast.PushErrorToken(token, "invalid_syntax")
	}
}

// BuildAttribute builds AST model of attribute.
func (ast *AST) BuildAttribute() {
	var attribute AttributeAST
	ast.Position++
	if ast.Ended() {
		ast.PushErrorToken(ast.Tokens[ast.Position-1], "invalid_syntax")
		return
	}
	ast.Position++
	if ast.Ended() {
		ast.PushErrorToken(ast.Tokens[ast.Position-1], "invalid_syntax")
		return
	}
	attribute.Token = ast.Tokens[ast.Position]
	if attribute.Token.Type != lex.Brace || attribute.Token.Value != "]" {
		ast.PushErrorToken(attribute.Token, "invalid_syntax")
		ast.Position = -1 // Stop modelling.
		return
	}
	attribute.Token = ast.Tokens[ast.Position-1]
	attribute.Value = attribute.Token.Value
	ast.Tree = append(ast.Tree, Object{
		Token: attribute.Token,
		Value: attribute,
	})
	ast.Position++
}

// BuildFunction builds AST model of function.
func (ast *AST) BuildFunction(anonymous bool) (funAST FunctionAST) {
	funAST.Token = ast.Tokens[ast.Position]
	if anonymous {
		funAST.Name = "anonymous"
	} else {
		if funAST.Token.Type != lex.Name {
			ast.PushErrorToken(funAST.Token, "invalid_syntax")
		}
		funAST.Name = funAST.Token.Value
		ast.Position++
		if ast.Ended() {
			ast.Position--
			ast.PushError("function_body_not_exist")
			ast.Position = -1 // Stop modelling.
			return
		}
	}
	funAST.ReturnType.Code = x.Void
	tokens := ast.getRange("(", ")")
	if tokens == nil {
		return
	} else if len(tokens) > 0 {
		ast.BuildParameters(&funAST, tokens)
	}
	if ast.Ended() {
		ast.Position--
		ast.PushError("function_body_not_exist")
		ast.Position = -1 // Stop modelling.
		return
	}
	token := ast.Tokens[ast.Position]
	t, ok := ast.BuildDataType(ast.Tokens, &ast.Position, false)
	if ok {
		funAST.ReturnType = t
		ast.Position++
		if ast.Ended() {
			ast.Position--
			ast.PushError("function_body_not_exist")
			ast.Position = -1 // Stop modelling.
			return
		}
		token = ast.Tokens[ast.Position]
	}
	if token.Type != lex.Brace || token.Value != "{" {
		ast.PushError("invalid_syntax")
		ast.Position = -1 // Stop modelling.
		return
	}
	blockTokens := ast.getRange("{", "}")
	if blockTokens == nil {
		ast.PushError("function_body_not_exist")
		ast.Position = -1 // Stop modelling.
		return
	}
	funAST.Block = ast.BuildBlock(blockTokens)
	return
}

// BuildGlobalVariable builds AST model of global variable.
func (ast *AST) BuildGlobalVariable() {
	statementTokens := ast.skipStatement()
	if statementTokens == nil {
		return
	}
	statement := ast.BuildVariableStatement(statementTokens)
	ast.Tree = append(ast.Tree, Object{
		Token: statement.Token,
		Value: statement,
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
	paramAST := ParameterAST{
		Token: tokens[0],
	}
	index := 0
	if t, ok := ast.BuildDataType(tokens, &index, true); ok {
		if index+1 == len(tokens) {
			paramAST.Type = t
			goto end
		}
	}
	{
		if len(tokens) < 2 {
			ast.PushErrorToken(paramAST.Token, "missing_type")
			return
		}
		nameToken := tokens[0]
		if nameToken.Type != lex.Name {
			ast.PushErrorToken(nameToken, "invalid_syntax")
		}
		if !x.IsIgnoreName(nameToken.Value) {
			for _, param := range fn.Params {
				if param.Name == nameToken.Value {
					ast.PushErrorToken(nameToken, "parameter_exist")
					break
				}
			}
			paramAST.Name = nameToken.Value
		}
		index := 1
		paramAST.Type, _ = ast.BuildDataType(tokens, &index, true)
		if index+1 < len(tokens) {
			ast.PushErrorToken(tokens[index+1], "invalid_syntax")
		}
	}
end:
	fn.Params = append(fn.Params, paramAST)
}

// IsStatement reports token is
// statement finish point or not.
func IsStatement(token lex.Token) bool { return token.Type == lex.SemiColon }

// BuildDataType builds AST model of data type.
func (ast *AST) BuildDataType(tokens []lex.Token, index *int, err bool) (t DataTypeAST, _ bool) {
	first := *index
	for ; *index < len(tokens); *index++ {
		token := tokens[*index]
		switch token.Type {
		case lex.DataType:
			t.Token = token
			t.Code = x.TypeFromName(t.Token.Value)
			t.Value += t.Token.Value
			return t, true
		case lex.Name:
			t.Token = token
			t.Code = x.Name
			t.Value += t.Token.Value
			return t, true
		case lex.Operator:
			if token.Value == "*" {
				t.Value += token.Value
				break
			}
			fallthrough
		case lex.Brace:
			switch token.Value {
			case "(":
				t.Token = token
				t.Code = x.Function
				value, funAST := ast.buildFunctionDataType(tokens, index)
				funAST.ReturnType, _ = ast.BuildDataType(tokens, index, false)
				t.Value += value
				t.Tag = funAST
				return t, true
			case "[":
				*index++
				if *index > len(tokens) {
					if err {
						ast.PushErrorToken(token, "invalid_syntax")
					}
					return t, false
				}
				token = tokens[*index]
				if token.Type != lex.Brace || token.Value != "]" {
					if err {
						ast.PushErrorToken(token, "invalid_syntax")
					}
					return t, false
				}
				t.Value += "[]"
				continue
			}
			/*if err {
				ast.PushErrorToken(token, "invalid_syntax")
			}*/
			return t, false
		default:
			if err {
				ast.PushErrorToken(token, "invalid_syntax")
			}
			return t, false
		}
	}
	if err {
		ast.PushErrorToken(tokens[first], "invalid_type")
	}
	return t, false
}

func (ast *AST) buildFunctionDataType(tokens []lex.Token, index *int) (string, FunctionAST) {
	var funAST FunctionAST
	var typeValue strings.Builder
	typeValue.WriteByte('(')
	brace := 1
	firstIndex := *index
	for *index++; *index < len(tokens); *index++ {
		token := tokens[*index]
		typeValue.WriteString(token.Value)
		switch token.Type {
		case lex.Brace:
			switch token.Value {
			case "{", "[", "(":
				brace++
			default:
				brace--
			}
		}
		if brace == 0 {
			ast.BuildParameters(&funAST, tokens[firstIndex+1:*index])
			return typeValue.String(), funAST
		}
	}
	ast.PushErrorToken(tokens[firstIndex], "invalid_type")
	return "", funAST
}

// IsSigleOperator is returns true
// if operator is unary or smilar to unary,
// returns false if not.
func IsSingleOperator(operator string) bool {
	return operator == "-" ||
		operator == "+" ||
		operator == "~" ||
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
			switch token.Value {
			case "{":
				braceCount++
			case "}":
				braceCount--
			}
		}
		if braceCount > 0 ||
			!IsStatement(token) ||
			index-oldStatementPoint == 0 {
			continue
		}
		b.Statements = append(b.Statements,
			ast.BuildStatement(tokens[oldStatementPoint:index]))
		if ast.Position == -1 {
			break
		}
		oldStatementPoint = index + 1
	}
	if oldStatementPoint < len(tokens) {
		ast.PushErrorToken(tokens[len(tokens)-1], "missing_semicolon")
	}
	return
}

// BuildStatement builds AST model of statement.
func (ast *AST) BuildStatement(tokens []lex.Token) (s StatementAST) {
	s, ok := ast.BuildVariableSetStatement(tokens)
	if ok {
		return s
	}
	firstToken := tokens[0]
	switch firstToken.Type {
	case lex.Name:
		return ast.BuildNameStatement(tokens)
	case lex.Var, lex.Const:
		return ast.BuildVariableStatement(tokens)
	case lex.Return:
		return ast.BuildReturnStatement(tokens)
	case lex.Brace:
		if firstToken.Value == "(" {
			return ast.BuildExpressionStatement(tokens)
		}
	case lex.Operator:
		if firstToken.Value == "<" {
			return ast.BuildReturnStatement(tokens)
		}
	}
	ast.PushErrorToken(firstToken, "invalid_syntax")
	return
}

// BuildVariableSetStatement builds AST model of variable set statement.
func (ast *AST) BuildVariableSetStatement(tokens []lex.Token) (s StatementAST, _ bool) {
	switch tokens[0].Type {
	case lex.Name, lex.Brace, lex.Operator:
	default:
		return
	}
	braceCount := 0
	for index, token := range tokens {
		switch token.Type {
		case lex.Brace:
			switch token.Value {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount != 0 {
			continue
		}
		switch token.Value {
		case "=":
			if index == len(tokens)-1 {
				ast.PushErrorToken(token, "missing_value")
				return s, true /* true for not give another errors */
			}
			s = StatementAST{
				Token: token,
				Value: VariableSetAST{
					Setter:           token,
					SelectExpression: ast.BuildExpression(tokens[:index]),
					ValueExpression:  ast.BuildExpression(tokens[index+1:]),
				},
			}
			return s, true
		}
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

// BuildFunctionCallStatement builds AST model of function call statement.
func (ast *AST) BuildFunctionCallStatement(tokens []lex.Token) StatementAST {
	return ast.BuildExpressionStatement(tokens)
}

// BuildExpressionStatement builds AST model of expression.
func (ast *AST) BuildExpressionStatement(tokens []lex.Token) StatementAST {
	return StatementAST{
		Token: tokens[0],
		Value: BlockExpressionAST{
			Expression: ast.BuildExpression(tokens),
		},
	}
}

// BuildArgs builds AST model of arguments.
func (ast *AST) BuildArgs(tokens []lex.Token) []ArgAST {
	var args []ArgAST
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
		ast.pushArg(&args, tokens[last:index], token)
		last = index + 1
	}
	if last < len(tokens) {
		if last == 0 {
			ast.pushArg(&args, tokens[last:], tokens[last])
		} else {
			ast.pushArg(&args, tokens[last:], tokens[last-1])
		}
	}
	return args
}

func (ast *AST) pushArg(args *[]ArgAST, tokens []lex.Token, err lex.Token) {
	if len(tokens) == 0 {
		ast.PushErrorToken(err, "invalid_syntax")
		return
	}
	var arg ArgAST
	arg.Token = tokens[0]
	arg.Tokens = tokens
	arg.Expression = ast.BuildExpression(arg.Tokens)
	*args = append(*args, arg)
}

// BuildVariableStatement builds AST model of variable declaration statement.
func (ast *AST) BuildVariableStatement(tokens []lex.Token) (s StatementAST) {
	var varAST VariableAST
	position := 0
	varAST.DefineToken = tokens[position]
	position++
	varAST.NameToken = tokens[position]
	if varAST.NameToken.Type != lex.Name {
		ast.PushErrorToken(varAST.NameToken, "invalid_syntax")
	}
	varAST.Name = varAST.NameToken.Value
	varAST.Type = DataTypeAST{Code: x.Void}
	position++
	if position >= len(tokens) {
		ast.PushErrorToken(tokens[position-1], "missing_autotype_value")
		return
	}
	token := tokens[position]
	t, ok := ast.BuildDataType(tokens, &position, false)
	if ok {
		varAST.Type = t
		position++
		if position >= len(tokens) {
			if varAST.Type.Code == x.Void {
				ast.PushErrorToken(token, "missing_autotype_value")
				return
			}
			goto ret
		}
		token = tokens[position]
	}
	if token.Type == lex.Operator {
		if token.Value != "=" {
			ast.PushErrorToken(token, "invalid_syntax")
			return
		}
		valueTokens := tokens[position+1:]
		if len(valueTokens) == 0 {
			ast.PushErrorToken(token, "missing_value")
			return
		}
		varAST.Value = ast.BuildExpression(valueTokens)
		varAST.SetterToken = token
	}
ret:
	return StatementAST{
		Token: varAST.NameToken,
		Value: varAST,
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
	singleOperatored := false
	for index := 0; index < len(tokens); index++ {
		token := tokens[index]
		switch token.Type {
		case lex.Operator:
			if !operator {
				if IsSingleOperator(token.Value) && !singleOperatored {
					part = append(part, token)
					singleOperatored = true
					continue
				}
				if braceCount == 0 {
					ast.PushErrorToken(token, "operator_overflow")
				}
			}
			singleOperatored = false
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
				if token.Value == "[" {
					oldIndex := index
					if _, ok := ast.BuildDataType(tokens, &index, false); ok {
						part = append(part, tokens[oldIndex:index+1]...)
						continue
					}
					index = oldIndex
				}
				singleOperatored = false
				braceCount++
			default:
				braceCount--
			}
		}
		if index > 0 && braceCount == 0 {
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
		if token.Value == "(" ||
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
			ast.Position = -1 // Stop modelling.
			return nil
		}
		return ast.Tokens[start : ast.Position-1]
	}
	return nil
}

func (ast *AST) skipStatement() []lex.Token {
	start := ast.Position
	for ; !ast.Ended(); ast.Position++ {
		token := ast.Tokens[ast.Position]
		if token.Type == lex.SemiColon {
			ast.Position++
			return ast.Tokens[start : ast.Position-1]
		}
	}
	ast.Position--
	ast.PushError("missing_semicolon")
	ast.Position = -1 // Stop modelling.
	return nil
}
