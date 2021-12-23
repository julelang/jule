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
		"%s:%d:%d %s", token.File.Path, token.Row, token.Column, message))
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
		switch firstToken.Id {
		case lex.At:
			ast.BuildAttribute()
		case lex.Name:
			ast.BuildName()
		case lex.Const:
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
	if token.Id != lex.Name {
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
			Name:  tokens[1].Kind,
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
	switch token.Id {
	case lex.Colon:
		ast.BuildGlobalVariable()
	case lex.Brace:
		switch token.Kind {
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

// BuildAttribute builds AST model of attribute.
func (ast *AST) BuildAttribute() {
	var attribute AttributeAST
	attribute.Token = ast.Tokens[ast.Position]
	ast.Position++
	if ast.Ended() {
		ast.PushErrorToken(ast.Tokens[ast.Position-1], "invalid_syntax")
		return
	}
	attribute.Tag = ast.Tokens[ast.Position]
	if attribute.Tag.Id != lex.Name ||
		attribute.Token.Column+1 != attribute.Tag.Column {
		ast.PushErrorToken(attribute.Tag, "invalid_syntax")
		ast.Position = -1 // Stop modelling.
		return
	}
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
		if funAST.Token.Id != lex.Name {
			ast.PushErrorToken(funAST.Token, "invalid_syntax")
		}
		funAST.Name = funAST.Token.Kind
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
	if token.Id != lex.Brace || token.Kind != "{" {
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
		if token.Id == lex.Brace {
			switch token.Kind {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 || token.Id != lex.Comma {
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
	for index, token := range tokens {
		switch token.Id {
		case lex.Const:
			if paramAST.Const {
				ast.PushErrorToken(token, "already_constant")
				continue
			}
			paramAST.Const = true
		case lex.Name:
			tokens = tokens[index:]
			if len(tokens) < 2 {
				ast.PushErrorToken(paramAST.Token, "missing_type")
				return
			}
			if !x.IsIgnoreName(token.Kind) {
				for _, param := range fn.Params {
					if param.Name == token.Kind {
						ast.PushErrorToken(token, "parameter_exist")
						break
					}
				}
				paramAST.Name = token.Kind
			}
			index := 1
			paramAST.Type, _ = ast.BuildDataType(tokens, &index, true)
			if index+1 < len(tokens) {
				ast.PushErrorToken(tokens[index+1], "invalid_syntax")
			}
			goto end
		default:
			if t, ok := ast.BuildDataType(tokens, &index, true); ok {
				if index+1 == len(tokens) {
					paramAST.Type = t
					goto end
				}
			}
			ast.PushErrorToken(token, "invalid_syntax")
			goto end
		}
	}
end:
	if paramAST.Type.Code == x.Void {
		ast.PushErrorToken(paramAST.Token, "invalid_syntax")
	}
	fn.Params = append(fn.Params, paramAST)
}

// IsStatement reports token is
// statement finish point or not.
func IsStatement(token lex.Token) bool { return token.Id == lex.SemiColon }

// BuildDataType builds AST model of data type.
func (ast *AST) BuildDataType(tokens []lex.Token, index *int, err bool) (dt DataTypeAST, _ bool) {
	first := *index
	for ; *index < len(tokens); *index++ {
		token := tokens[*index]
		switch token.Id {
		case lex.DataType:
			buildDataType(token, &dt)
			return dt, true
		case lex.Name:
			buildNameType(token, &dt)
			return dt, true
		case lex.Operator:
			if token.Kind == "*" {
				dt.Value += token.Kind
				break
			}
			fallthrough
		case lex.Brace:
			switch token.Kind {
			case "(":
				ast.buildFunctionType(token, tokens, index, &dt)
				return dt, true
			case "[":
				*index++
				if *index > len(tokens) {
					if err {
						ast.PushErrorToken(token, "invalid_syntax")
					}
					return dt, false
				}
				token = tokens[*index]
				if token.Id != lex.Brace || token.Kind != "]" {
					if err {
						ast.PushErrorToken(token, "invalid_syntax")
					}
					return dt, false
				}
				dt.Value += "[]"
				continue
			}
			/*if err {
				ast.PushErrorToken(token, "invalid_syntax")
			}*/
			return dt, false
		default:
			if err {
				ast.PushErrorToken(token, "invalid_syntax")
			}
			return dt, false
		}
	}
	if err {
		ast.PushErrorToken(tokens[first], "invalid_type")
	}
	return dt, false
}

func buildDataType(token lex.Token, dt *DataTypeAST) {
	dt.Token = token
	dt.Code = x.TypeFromName(dt.Token.Kind)
	dt.Value += dt.Token.Kind
}

func buildNameType(token lex.Token, dt *DataTypeAST) {
	dt.Token = token
	dt.Code = x.Name
	dt.Value += dt.Token.Kind
}

func (ast *AST) buildFunctionType(token lex.Token, tokens []lex.Token, index *int, dt *DataTypeAST) {
	dt.Token = token
	dt.Code = x.Function
	value, funAST := ast.buildFunctionDataType(tokens, index)
	funAST.ReturnType, _ = ast.BuildDataType(tokens, index, false)
	dt.Value += value
	dt.Tag = funAST
}

func (ast *AST) buildFunctionDataType(tokens []lex.Token, index *int) (string, FunctionAST) {
	var funAST FunctionAST
	var typeValue strings.Builder
	typeValue.WriteByte('(')
	brace := 1
	firstIndex := *index
	for *index++; *index < len(tokens); *index++ {
		token := tokens[*index]
		typeValue.WriteString(token.Kind)
		switch token.Id {
		case lex.Brace:
			switch token.Kind {
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
		if token.Id == lex.Brace {
			switch token.Kind {
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
	switch firstToken.Id {
	case lex.Name:
		return ast.BuildNameStatement(tokens)
	case lex.Const:
		return ast.BuildVariableStatement(tokens)
	case lex.Return:
		return ast.BuildReturnStatement(tokens)
	case lex.Brace:
		if firstToken.Kind == "(" {
			return ast.BuildExpressionStatement(tokens)
		}
	case lex.Operator:
		if firstToken.Kind == "<" {
			return ast.BuildReturnStatement(tokens)
		}
	}
	ast.PushErrorToken(firstToken, "invalid_syntax")
	return
}

// BuildVariableSetStatement builds AST model of variable set statement.
func (ast *AST) BuildVariableSetStatement(tokens []lex.Token) (s StatementAST, _ bool) {
	switch tokens[0].Id {
	case lex.Name, lex.Brace, lex.Operator:
		if len(tokens) > 1 && tokens[1].Id == lex.Colon {
			return
		}
	default:
		return
	}
	braceCount := 0
	for index, token := range tokens {
		switch token.Id {
		case lex.Brace:
			switch token.Kind {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount != 0 {
			continue
		}
		if strings.HasSuffix(token.Kind, "=") {
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
	switch tokens[1].Id {
	case lex.Colon:
		return ast.BuildVariableStatement(tokens)
	case lex.Brace:
		switch tokens[1].Kind {
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
		if token.Id == lex.Brace {
			switch token.Kind {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 || token.Id != lex.Comma {
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
	if tokens[position].Id != lex.Name {
		varAST.DefineToken = tokens[position]
		position++
	}
	varAST.NameToken = tokens[position]
	if varAST.NameToken.Id != lex.Name {
		ast.PushErrorToken(varAST.NameToken, "invalid_syntax")
	}
	varAST.Name = varAST.NameToken.Kind
	varAST.Type = DataTypeAST{Code: x.Void}
	// Skip type definer operator(':')
	position++
	if varAST.DefineToken.File != nil {
		if tokens[position].Id != lex.Colon {
			ast.PushErrorToken(tokens[position], "invalid_syntax")
			return
		}
		position++
	} else {
		position++
	}
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
	if token.Id == lex.Operator {
		if token.Kind != "=" {
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
		switch token.Id {
		case lex.Operator:
			if !operator {
				if IsSingleOperator(token.Kind) && !singleOperatored {
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
			switch token.Kind {
			case "(", "[", "{":
				if token.Kind == "[" {
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
			if (lt.Id == lex.Name || lt.Id == lex.Value) &&
				(token.Id == lex.Name || token.Id == lex.Value) {
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
	switch token.Id {
	case lex.Comma:
		return false
	case lex.Brace:
		if token.Kind == "(" ||
			token.Kind == "{" {
			return false
		}
	}
	return index < tokensLen-1
}

func (ast *AST) checkExpressionToken(token lex.Token) {
	if token.Kind[0] >= '0' && token.Kind[0] <= '9' {
		var result bool
		if strings.IndexByte(token.Kind, '.') != -1 {
			_, result = new(big.Float).SetString(token.Kind)
		} else {
			result = xbits.CheckBitInt(token.Kind, 64)
		}
		if !result {
			ast.PushErrorToken(token, "invalid_numeric_range")
		}
	}
}

func (ast *AST) getRange(open, close string) []lex.Token {
	token := ast.Tokens[ast.Position]
	if token.Id == lex.Brace && token.Kind == open {
		ast.Position++
		braceCount := 1
		start := ast.Position
		for ; braceCount > 0 && !ast.Ended(); ast.Position++ {
			token := ast.Tokens[ast.Position]
			if token.Id != lex.Brace {
				continue
			}
			if token.Kind == open {
				braceCount++
			} else if token.Kind == close {
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
		if token.Id == lex.SemiColon {
			ast.Position++
			return ast.Tokens[start : ast.Position-1]
		}
	}
	ast.Position--
	ast.PushError("missing_semicolon")
	ast.Position = -1 // Stop modelling.
	return nil
}
