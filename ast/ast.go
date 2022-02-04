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
	t, ok := ast.BuildFunctionReturnDataType(ast.Tokens, &ast.Position)
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

// BuildDataType builds AST model of data type.
func (ast *AST) BuildDataType(tokens []lex.Token, index *int, err bool) (dt DataTypeAST, ok bool) {
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
			if err {
				ast.PushErrorToken(token, "invalid_syntax")
			}
			return dt, false
		case lex.Brace:
			switch token.Kind {
			case "(":
				ast.buildFunctionDataType(token, tokens, index, &dt)
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

func (ast *AST) buildFunctionDataType(token lex.Token, tokens []lex.Token, index *int, dt *DataTypeAST) {
	dt.Token = token
	dt.Code = x.Function
	value, funAST := ast.buildFunctionDataTypeHead(tokens, index)
	funAST.ReturnType, _ = ast.BuildFunctionReturnDataType(tokens, index)
	dt.Value += value
	dt.Tag = funAST
}

func (ast *AST) buildFunctionDataTypeHead(tokens []lex.Token, index *int) (string, FunctionAST) {
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

func (ast *AST) pushTypeToTypes(types *[]DataTypeAST, tokens []lex.Token, errToken lex.Token) {
	if len(tokens) == 0 {
		ast.PushErrorToken(errToken, "missing_value")
		return
	}
	currentDt, _ := ast.BuildDataType(tokens, new(int), false)
	*types = append(*types, currentDt)
}

func (ast *AST) BuildFunctionReturnDataType(tokens []lex.Token, index *int) (dt DataTypeAST, ok bool) {
	if *index >= len(tokens) {
		goto end
	}
	if tokens[*index].Id == lex.Brace &&
		tokens[*index].Kind == "[" {
		*index++
		if *index >= len(tokens) {
			*index--
			goto end
		}
		if tokens[*index].Id == lex.Brace &&
			tokens[*index].Kind == "]" {
			*index--
			goto end
		}
		var types []DataTypeAST
		braceCount := 1
		last := *index
		for ; *index < len(tokens); *index++ {
			token := tokens[*index]
			if token.Id == lex.Brace {
				switch token.Kind {
				case "(", "[", "{":
					braceCount++
				default:
					braceCount--
				}
			}
			if braceCount == 0 {
				ast.pushTypeToTypes(&types, tokens[last:*index], tokens[last-1])
				break
			} else if braceCount > 1 {
				continue
			}
			if token.Id != lex.Comma {
				continue
			}
			ast.pushTypeToTypes(&types, tokens[last:*index], tokens[*index-1])
			last = *index + 1
		}
		if len(types) > 1 {
			dt.MultiTyped = true
			dt.Tag = types
		} else {
			dt = types[0]
		}
		ok = true
		return
	}
end:
	return ast.BuildDataType(tokens, index, false)
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

// IsStatement reports token is
// statement finish point or not.
func IsStatement(current, prev lex.Token) (yes bool, withSemicolon bool) {
	yes = current.Id == lex.SemiColon || prev.Row < current.Row
	if yes {
		withSemicolon = current.Id == lex.SemiColon
	}
	return
}

func (ast *AST) pushStatementToBlock(b *BlockAST, tokens []lex.Token) {
	if len(tokens) == 0 {
		return
	}
	if tokens[len(tokens)-1].Id == lex.SemiColon {
		if len(tokens) == 1 {
			return
		}
		tokens = tokens[:len(tokens)-1]
	}
	b.Statements = append(b.Statements, ast.BuildStatement(tokens))
}

func nextStatementPos(tokens []lex.Token, start int) int {
	braceCount := 0
	index := start
	for ; index < len(tokens); index++ {
		token := tokens[index]
		if token.Id == lex.Brace {
			switch token.Kind {
			case "{", "[", "(":
				braceCount++
				continue
			default:
				braceCount--
				continue
			}
		}
		if braceCount > 0 {
			continue
		}
		var isStatement, withSemicolon bool
		if index > start {
			isStatement, withSemicolon = IsStatement(token, tokens[index-1])
		} else {
			isStatement, withSemicolon = IsStatement(token, token)
		}
		if !isStatement {
			continue
		}
		if withSemicolon {
			index++
		}
		return index
	}
	return index
}

// BuildBlock builds AST model of statements of code block.
func (ast *AST) BuildBlock(tokens []lex.Token) (b BlockAST) {
	var index, start int
	for {
		if ast.Position == -1 {
			return
		}
		index = nextStatementPos(tokens, index)
		ast.pushStatementToBlock(&b, tokens[start:index])
		if index >= len(tokens) {
			break
		}
		start = index
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

func checkVariableSetStatementTokens(tokens []lex.Token) bool {
	braceCount := 0
	for _, token := range tokens {
		if token.Id == lex.Brace {
			switch token.Kind {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		}
		if token.Id == lex.Operator &&
			token.Kind[len(token.Kind)-1] == '=' {
			return true
		}
	}
	return false
}

type varsetInfo struct {
	selectorTokens   []lex.Token
	expressionTokens []lex.Token
	setter           lex.Token
	ok               bool
	justDeclare      bool
}

func (ast *AST) variableSetInfo(tokens []lex.Token) (info varsetInfo) {
	info.ok = true
	braceCount := 0
	for index, token := range tokens {
		if token.Id == lex.Brace {
			switch token.Kind {
			case "(", "[", "{":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		}
		if token.Id == lex.Operator &&
			token.Kind[len(token.Kind)-1] == '=' {
			info.selectorTokens = tokens[:index]
			if info.selectorTokens == nil {
				ast.PushErrorToken(token, "invalid_syntax")
				info.ok = false
			}
			info.setter = token
			if index+1 >= len(tokens) {
				ast.PushErrorToken(token, "missing_value")
				info.ok = false
			} else {
				info.expressionTokens = tokens[index+1:]
			}
			return
		}
	}
	info.justDeclare = true
	info.selectorTokens = tokens
	return
}

func (ast *AST) pushVarsetSelector(selectors *[]VarsetSelector, last, current int, info varsetInfo) {
	var selector VarsetSelector
	selector.Expression.Tokens = info.selectorTokens[last:current]
	if last-current == 0 {
		ast.PushErrorToken(info.selectorTokens[current-1], "missing_value")
		return
	}
	// Variable is new?
	if selector.Expression.Tokens[0].Id == lex.Name &&
		current-last > 1 &&
		selector.Expression.Tokens[1].Id == lex.Colon {
		selector.NewVariable = true
		selector.Variable.NameToken = selector.Expression.Tokens[0]
		selector.Variable.Name = selector.Variable.NameToken.Kind
		selector.Variable.SetterToken = info.setter
		// Has specific data-type?
		if current-last > 2 {
			selector.Variable.Type, _ = ast.BuildDataType(
				selector.Expression.Tokens[2:], new(int), false)
		}
	} else {
		if selector.Expression.Tokens[0].Id == lex.Name {
			selector.Variable.NameToken = selector.Expression.Tokens[0]
			selector.Variable.Name = selector.Variable.NameToken.Kind
		}
		selector.Expression = ast.BuildExpression(selector.Expression.Tokens)
	}
	*selectors = append(*selectors, selector)
}

func (ast *AST) varsetSelectors(info varsetInfo) []VarsetSelector {
	var selectors []VarsetSelector
	braceCount := 0
	lastIndex := 0
	for index, token := range info.selectorTokens {
		if token.Id == lex.Brace {
			switch token.Kind {
			case "(", "[", "{":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		} else if token.Id != lex.Comma {
			continue
		}
		ast.pushVarsetSelector(&selectors, lastIndex, index, info)
		lastIndex = index + 1
	}
	if lastIndex < len(info.selectorTokens) {
		ast.pushVarsetSelector(&selectors, lastIndex,
			len(info.selectorTokens), info)
	}
	return selectors
}

func (ast *AST) pushVarsetExpression(exps *[]ExpressionAST, last, current int, info varsetInfo) {
	tokens := info.expressionTokens[last:current]
	if tokens == nil {
		ast.PushErrorToken(info.expressionTokens[current-1], "missing_value")
		return
	}
	*exps = append(*exps, ast.BuildExpression(tokens))
}

func (ast *AST) varsetExpressions(info varsetInfo) []ExpressionAST {
	var expressions []ExpressionAST
	braceCount := 0
	lastIndex := 0
	for index, token := range info.expressionTokens {
		if token.Id == lex.Brace {
			switch token.Kind {
			case "(", "[", "{":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		} else if token.Id != lex.Comma {
			continue
		}
		ast.pushVarsetExpression(&expressions, lastIndex, index, info)
		lastIndex = index + 1
	}
	if lastIndex < len(info.expressionTokens) {
		ast.pushVarsetExpression(&expressions, lastIndex,
			len(info.expressionTokens), info)
	}
	return expressions
}

// BuildVariableSetStatement builds AST model of variable set statement.
func (ast *AST) BuildVariableSetStatement(tokens []lex.Token) (s StatementAST, _ bool) {
	if !checkVariableSetStatementTokens(tokens) {
		return
	}
	info := ast.variableSetInfo(tokens)
	if !info.ok {
		return
	}
	var varAST VariableSetAST
	varAST.Setter = info.setter
	varAST.JustDeclare = info.justDeclare
	varAST.SelectExpressions = ast.varsetSelectors(info)
	if !info.justDeclare {
		varAST.ValueExpressions = ast.varsetExpressions(info)
	}
	s.Token = tokens[0]
	s.Value = varAST
	return s, true
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
	if position < len(tokens) {
		token := tokens[position]
		t, ok := ast.BuildDataType(tokens, &position, false)
		if ok {
			varAST.Type = t
			position++
			if position >= len(tokens) {
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
	ast.Position = nextStatementPos(ast.Tokens, start)
	return ast.Tokens[start:ast.Position]
}
