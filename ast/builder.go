package ast

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
	"github.com/the-xlang/x/pkg/xbits"
)

// Builder is builds AST tree.
type Builder struct {
	Tree     []Object
	Errors   []string
	Tokens   []lex.Token
	Position int
}

// NewBuilder instance.
func NewBuilder(tokens []lex.Token) *Builder {
	ast := new(Builder)
	ast.Tokens = tokens
	ast.Position = 0
	return ast
}

// PushError appends error by specified token.
func (b *Builder) PushError(token lex.Token, err string) {
	message := x.Errors[err]
	b.Errors = append(b.Errors, fmt.Sprintf(
		"%s:%d:%d %s", token.File.Path, token.Row, token.Column, message))
}

// Ended reports position is at end of tokens or not.
func (ast *Builder) Ended() bool {
	return ast.Position >= len(ast.Tokens)
}

// Build builds AST tree.
func (b *Builder) Build() {
	for b.Position != -1 && !b.Ended() {
		statement := b.skipStatement()
		token := statement[0]
		switch token.Id {
		case lex.At:
			b.Attribute(statement)
		case lex.Name:
			b.Name(statement)
		case lex.Const:
			b.GlobalVariable(statement)
		case lex.Type:
			b.Type(statement)
		default:
			b.PushError(token, "invalid_syntax")
		}
	}
}

// Type builds AST model of type defination statement.
func (b *Builder) Type(tokens []lex.Token) {
	position := 1 // Initialize value is 1 for skip keyword.
	if position >= len(tokens) {
		b.PushError(tokens[position-1], "invalid_syntax")
		return
	}
	token := tokens[position]
	if token.Id != lex.Name {
		b.PushError(token, "invalid_syntax")
	}
	position++
	if position >= len(tokens) {
		b.PushError(tokens[position-1], "invalid_syntax")
		return
	}
	destType, _ := b.DataType(tokens[position:], new(int), true)
	token = tokens[1]
	typeAST := TypeAST{token, token.Kind, destType}
	b.Tree = append(b.Tree, Object{token, typeAST})
}

// Name builds AST model of global name statement.
func (b *Builder) Name(tokens []lex.Token) {
	if len(tokens) == 1 {
		b.PushError(tokens[0], "invalid_syntax")
		return
	}
	token := tokens[1]
	switch token.Id {
	case lex.Colon:
		b.GlobalVariable(tokens)
		return
	case lex.Brace:
		switch token.Kind {
		case "(":
			funAST := b.Function(tokens, false)
			statement := StatementAST{funAST.Token, funAST}
			b.Tree = append(b.Tree, Object{funAST.Token, statement})
			return
		}
	}
	b.PushError(token, "invalid_syntax")
}

// Attribute builds AST model of attribute.
func (b *Builder) Attribute(tokens []lex.Token) {
	var attribute AttributeAST
	index := 0
	attribute.Token = tokens[index]
	index++
	if b.Ended() {
		b.PushError(tokens[index-1], "invalid_syntax")
		return
	}
	attribute.Tag = tokens[index]
	if attribute.Tag.Id != lex.Name ||
		attribute.Token.Column+1 != attribute.Tag.Column {
		b.PushError(attribute.Tag, "invalid_syntax")
		return
	}
	b.Tree = append(b.Tree, Object{attribute.Token, attribute})
}

// Function builds AST model of function.
func (b *Builder) Function(tokens []lex.Token, anonymous bool) (funAST FunctionAST) {
	funAST.Token = tokens[0]
	index := 0
	if anonymous {
		funAST.Name = "anonymous"
	} else {
		if funAST.Token.Id != lex.Name {
			b.PushError(funAST.Token, "invalid_syntax")
		}
		funAST.Name = funAST.Token.Kind
		index++
	}
	funAST.ReturnType.Code = x.Void
	paramTokens := getRange(&index, "(", ")", tokens)
	if len(paramTokens) > 0 {
		b.Parameters(&funAST, paramTokens)
	}
	if index >= len(tokens) {
		b.PushError(funAST.Token, "function_body_not_exist")
		return
	}
	token := tokens[index]
	t, ok := b.FunctionReturnDataType(tokens, &index)
	if ok {
		funAST.ReturnType = t
		index++
		if index >= len(tokens) {
			b.PushError(funAST.Token, "function_body_not_exist")
			return
		}
		token = tokens[index]
	}
	if token.Id != lex.Brace || token.Kind != "{" {
		b.PushError(token, "invalid_syntax")
		return
	}
	blockTokens := getRange(&index, "{", "}", tokens)
	if blockTokens == nil {
		b.PushError(funAST.Token, "function_body_not_exist")
		return
	}
	if index < len(tokens) {
		b.PushError(tokens[index], "invalid_syntax")
	}
	funAST.Block = b.Block(blockTokens)
	return
}

// GlobalVariable builds AST model of global variable.
func (b *Builder) GlobalVariable(tokens []lex.Token) {
	if tokens == nil {
		return
	}
	statement := b.VariableStatement(tokens)
	b.Tree = append(b.Tree, Object{statement.Token, statement})
}

// Parameters builds AST model of function parameters.
func (b *Builder) Parameters(fn *FunctionAST, tokens []lex.Token) {
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
		b.pushParameter(fn, tokens[last:index], token)
		last = index + 1
	}
	if last < len(tokens) {
		if last == 0 {
			b.pushParameter(fn, tokens[last:], tokens[last])
		} else {
			b.pushParameter(fn, tokens[last:], tokens[last-1])
		}
	}
}

func (b *Builder) pushParameter(fn *FunctionAST, tokens []lex.Token, err lex.Token) {
	if len(tokens) == 0 {
		b.PushError(err, "invalid_syntax")
		return
	}
	paramAST := ParameterAST{
		Token: tokens[0],
	}
	for index, token := range tokens {
		switch token.Id {
		case lex.Const:
			if paramAST.Const {
				b.PushError(token, "already_constant")
				continue
			}
			paramAST.Const = true
		case lex.Name:
			tokens = tokens[index:]
			if len(tokens) < 2 {
				b.PushError(paramAST.Token, "missing_type")
				return
			}
			if !x.IsIgnoreName(token.Kind) {
				for _, param := range fn.Params {
					if param.Name == token.Kind {
						b.PushError(token, "parameter_exist")
						break
					}
				}
				paramAST.Name = token.Kind
			}
			index := 1
			paramAST.Type, _ = b.DataType(tokens, &index, true)
			if index+1 < len(tokens) {
				b.PushError(tokens[index+1], "invalid_syntax")
			}
			goto end
		default:
			if t, ok := b.DataType(tokens, &index, true); ok {
				if index+1 == len(tokens) {
					paramAST.Type = t
					goto end
				}
			}
			b.PushError(token, "invalid_syntax")
			goto end
		}
	}
end:
	if paramAST.Type.Code == x.Void {
		b.PushError(paramAST.Token, "invalid_syntax")
	}
	fn.Params = append(fn.Params, paramAST)
}

// DataType builds AST model of data type.
func (b *Builder) DataType(tokens []lex.Token, index *int, err bool) (dt DataTypeAST, ok bool) {
	first := *index
	for ; *index < len(tokens); *index++ {
		token := tokens[*index]
		switch token.Id {
		case lex.DataType:
			dataType(token, &dt)
			return dt, true
		case lex.Name:
			nameType(token, &dt)
			return dt, true
		case lex.Operator:
			if token.Kind == "*" {
				dt.Value += token.Kind
				break
			}
			if err {
				b.PushError(token, "invalid_syntax")
			}
			return dt, false
		case lex.Brace:
			switch token.Kind {
			case "(":
				b.functionDataType(token, tokens, index, &dt)
				return dt, true
			case "[":
				*index++
				if *index > len(tokens) {
					if err {
						b.PushError(token, "invalid_syntax")
					}
					return dt, false
				}
				token = tokens[*index]
				if token.Id != lex.Brace || token.Kind != "]" {
					if err {
						b.PushError(token, "invalid_syntax")
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
				b.PushError(token, "invalid_syntax")
			}
			return dt, false
		}
	}
	if err {
		b.PushError(tokens[first], "invalid_type")
	}
	return dt, false
}

func dataType(token lex.Token, dt *DataTypeAST) {
	dt.Token = token
	dt.Code = x.TypeFromName(dt.Token.Kind)
	dt.Value += dt.Token.Kind
}

func nameType(token lex.Token, dt *DataTypeAST) {
	dt.Token = token
	dt.Code = x.Name
	dt.Value += dt.Token.Kind
}

func (b *Builder) functionDataType(token lex.Token, tokens []lex.Token, index *int, dt *DataTypeAST) {
	dt.Token = token
	dt.Code = x.Function
	value, funAST := b.FunctionDataTypeHead(tokens, index)
	funAST.ReturnType, _ = b.FunctionReturnDataType(tokens, index)
	dt.Value += value
	dt.Tag = funAST
}

func (b *Builder) FunctionDataTypeHead(tokens []lex.Token, index *int) (string, FunctionAST) {
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
			b.Parameters(&funAST, tokens[firstIndex+1:*index])
			return typeValue.String(), funAST
		}
	}
	b.PushError(tokens[firstIndex], "invalid_type")
	return "", funAST
}

func (b *Builder) pushTypeToTypes(types *[]DataTypeAST, tokens []lex.Token, errToken lex.Token) {
	if len(tokens) == 0 {
		b.PushError(errToken, "missing_value")
		return
	}
	currentDt, _ := b.DataType(tokens, new(int), false)
	*types = append(*types, currentDt)
}

func (b *Builder) FunctionReturnDataType(tokens []lex.Token, index *int) (dt DataTypeAST, ok bool) {
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
				b.pushTypeToTypes(&types, tokens[last:*index], tokens[last-1])
				break
			} else if braceCount > 1 {
				continue
			}
			if token.Id != lex.Comma {
				continue
			}
			b.pushTypeToTypes(&types, tokens[last:*index], tokens[*index-1])
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
	return b.DataType(tokens, index, false)
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

func (b *Builder) pushStatementToBlock(block *BlockAST, tokens []lex.Token) {
	if len(tokens) == 0 {
		return
	}
	if tokens[len(tokens)-1].Id == lex.SemiColon {
		if len(tokens) == 1 {
			return
		}
		tokens = tokens[:len(tokens)-1]
	}
	statement := b.Statement(tokens)
	block.Statements = append(block.Statements, statement)
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

// Block builds AST model of statements of code block.
func (b *Builder) Block(tokens []lex.Token) (block BlockAST) {
	var index, start int
	for {
		if b.Position == -1 {
			return
		}
		index = nextStatementPos(tokens, index)
		b.pushStatementToBlock(&block, tokens[start:index])
		if index >= len(tokens) {
			break
		}
		start = index
	}
	return
}

// Statement builds AST model of statement.
func (b *Builder) Statement(tokens []lex.Token) (s StatementAST) {
	s, ok := b.VariableSetStatement(tokens)
	if ok {
		return s
	}
	firstToken := tokens[0]
	switch firstToken.Id {
	case lex.Name:
		return b.NameStatement(tokens)
	case lex.Const:
		return b.VariableStatement(tokens)
	case lex.Return:
		return b.ReturnStatement(tokens)
	case lex.Free:
		return b.FreeStatement(tokens)
	case lex.Iter:
		return b.IterStatement(tokens)
	case lex.Break:
		return b.BreakStatement(tokens)
	case lex.Continue:
		return b.ContinueStatement(tokens)
	case lex.Brace:
		if firstToken.Kind == "(" {
			return b.ExprStatement(tokens)
		}
	case lex.Operator:
		if firstToken.Kind == "<" {
			return b.ReturnStatement(tokens)
		}
	}
	b.PushError(firstToken, "invalid_syntax")
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
	selectorTokens []lex.Token
	exprTokens     []lex.Token
	setter         lex.Token
	ok             bool
	justDeclare    bool
}

func (b *Builder) variableSetInfo(tokens []lex.Token) (info varsetInfo) {
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
				b.PushError(token, "invalid_syntax")
				info.ok = false
			}
			info.setter = token
			if index+1 >= len(tokens) {
				b.PushError(token, "missing_value")
				info.ok = false
			} else {
				info.exprTokens = tokens[index+1:]
			}
			return
		}
	}
	info.justDeclare = true
	info.selectorTokens = tokens
	return
}

func (b *Builder) pushVarsetSelector(selectors *[]VarsetSelector, last, current int, info varsetInfo) {
	var selector VarsetSelector
	selector.Expr.Tokens = info.selectorTokens[last:current]
	if last-current == 0 {
		b.PushError(info.selectorTokens[current-1], "missing_value")
		return
	}
	// Variable is new?
	if selector.Expr.Tokens[0].Id == lex.Name &&
		current-last > 1 &&
		selector.Expr.Tokens[1].Id == lex.Colon {
		selector.NewVariable = true
		selector.Variable.NameToken = selector.Expr.Tokens[0]
		selector.Variable.Name = selector.Variable.NameToken.Kind
		selector.Variable.SetterToken = info.setter
		// Has specific data-type?
		if current-last > 2 {
			selector.Variable.Type, _ = b.DataType(
				selector.Expr.Tokens[2:], new(int), false)
		}
	} else {
		if selector.Expr.Tokens[0].Id == lex.Name {
			selector.Variable.NameToken = selector.Expr.Tokens[0]
			selector.Variable.Name = selector.Variable.NameToken.Kind
		}
		selector.Expr = b.Expr(selector.Expr.Tokens)
	}
	*selectors = append(*selectors, selector)
}

func (b *Builder) varsetSelectors(info varsetInfo) []VarsetSelector {
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
		b.pushVarsetSelector(&selectors, lastIndex, index, info)
		lastIndex = index + 1
	}
	if lastIndex < len(info.selectorTokens) {
		b.pushVarsetSelector(&selectors, lastIndex, len(info.selectorTokens), info)
	}
	return selectors
}

func (b *Builder) pushVarsetExpr(exps *[]ExprAST, last, current int, info varsetInfo) {
	tokens := info.exprTokens[last:current]
	if tokens == nil {
		b.PushError(info.exprTokens[current-1], "missing_value")
		return
	}
	*exps = append(*exps, b.Expr(tokens))
}

func (b *Builder) varsetExprs(info varsetInfo) []ExprAST {
	var exprs []ExprAST
	braceCount := 0
	lastIndex := 0
	for index, token := range info.exprTokens {
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
		b.pushVarsetExpr(&exprs, lastIndex, index, info)
		lastIndex = index + 1
	}
	if lastIndex < len(info.exprTokens) {
		b.pushVarsetExpr(&exprs, lastIndex, len(info.exprTokens), info)
	}
	return exprs
}

// VariableSetStatement builds AST model of variable set statement.
func (b *Builder) VariableSetStatement(tokens []lex.Token) (s StatementAST, _ bool) {
	if !checkVariableSetStatementTokens(tokens) {
		return
	}
	info := b.variableSetInfo(tokens)
	if !info.ok {
		return
	}
	var varAST VariableSetAST
	varAST.Setter = info.setter
	varAST.JustDeclare = info.justDeclare
	varAST.SelectExprs = b.varsetSelectors(info)
	if !info.justDeclare {
		varAST.ValueExprs = b.varsetExprs(info)
	}
	s.Token = tokens[0]
	s.Value = varAST
	return s, true
}

// BuildReturnStatement builds AST model of return statement.
func (b *Builder) NameStatement(tokens []lex.Token) (s StatementAST) {
	if len(tokens) == 1 {
		b.PushError(tokens[0], "invalid_syntax")
		return
	}
	switch tokens[1].Id {
	case lex.Colon:
		return b.VariableStatement(tokens)
	case lex.Brace:
		switch tokens[1].Kind {
		case "(":
			return b.FunctionCallStatement(tokens)
		}
	}
	b.PushError(tokens[0], "invalid_syntax")
	return
}

// FunctionCallStatement builds AST model of function call statement.
func (b *Builder) FunctionCallStatement(tokens []lex.Token) StatementAST {
	return b.ExprStatement(tokens)
}

// ExprStatement builds AST model of expression.
func (b *Builder) ExprStatement(tokens []lex.Token) StatementAST {
	block := BlockExprAST{b.Expr(tokens)}
	return StatementAST{tokens[0], block}
}

// Args builds AST model of arguments.
func (b *Builder) Args(tokens []lex.Token) []ArgAST {
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
		b.pushArg(&args, tokens[last:index], token)
		last = index + 1
	}
	if last < len(tokens) {
		if last == 0 {
			b.pushArg(&args, tokens[last:], tokens[last])
		} else {
			b.pushArg(&args, tokens[last:], tokens[last-1])
		}
	}
	return args
}

func (b *Builder) pushArg(args *[]ArgAST, tokens []lex.Token, err lex.Token) {
	if len(tokens) == 0 {
		b.PushError(err, "invalid_syntax")
		return
	}
	var arg ArgAST
	arg.Token = tokens[0]
	arg.Tokens = tokens
	arg.Expr = b.Expr(arg.Tokens)
	*args = append(*args, arg)
}

// VariableStatement builds AST model of variable declaration statement.
func (b *Builder) VariableStatement(tokens []lex.Token) (s StatementAST) {
	var varAST VariableAST
	position := 0
	if tokens[position].Id != lex.Name {
		varAST.DefineToken = tokens[position]
		position++
	}
	varAST.NameToken = tokens[position]
	if varAST.NameToken.Id != lex.Name {
		b.PushError(varAST.NameToken, "invalid_syntax")
	}
	varAST.Name = varAST.NameToken.Kind
	varAST.Type = DataTypeAST{Code: x.Void}
	// Skip type definer operator(':')
	position++
	if varAST.DefineToken.File != nil {
		if tokens[position].Id != lex.Colon {
			b.PushError(tokens[position], "invalid_syntax")
			return
		}
		position++
	} else {
		position++
	}
	if position < len(tokens) {
		token := tokens[position]
		t, ok := b.DataType(tokens, &position, false)
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
				b.PushError(token, "invalid_syntax")
				return
			}
			valueTokens := tokens[position+1:]
			if len(valueTokens) == 0 {
				b.PushError(token, "missing_value")
				return
			}
			varAST.Value = b.Expr(valueTokens)
			varAST.SetterToken = token
		}
	}
ret:
	return StatementAST{varAST.NameToken, varAST}
}

// ReturnStatement builds AST model of return statement.
func (b *Builder) ReturnStatement(tokens []lex.Token) StatementAST {
	var returnModel ReturnAST
	returnModel.Token = tokens[0]
	if len(tokens) > 1 {
		returnModel.Expr = b.Expr(tokens[1:])
	}
	return StatementAST{returnModel.Token, returnModel}
}

func (b *Builder) FreeStatement(tokens []lex.Token) StatementAST {
	var freeAST FreeAST
	freeAST.Token = tokens[0]
	tokens = tokens[1:]
	if len(tokens) == 0 {
		b.PushError(freeAST.Token, "missing_expression")
	} else {
		freeAST.Expr = b.Expr(tokens)
	}
	return StatementAST{freeAST.Token, freeAST}
}

func iterExprTokens(tokens []lex.Token) (expr []lex.Token) {
	braceCount := 0
	for index, token := range tokens {
		if token.Id == lex.Brace {
			switch token.Kind {
			case "{":
				if braceCount > 0 {
					braceCount++
					break
				}
				return tokens[:index]
			case "(", "[":
				braceCount++
			default:
				braceCount--
			}
		}
	}
	return nil
}

func (b *Builder) IterStatement(tokens []lex.Token) (s StatementAST) {
	var iter IterAST
	iter.Token = tokens[0]
	tokens = tokens[1:]
	if len(tokens) == 0 {
		b.PushError(iter.Token, "iter_body_not_exist")
		return
	}
	exprTokens := iterExprTokens(tokens)
	if len(exprTokens) > 0 {
		iter.While = true
		iter.Profile.Expr = b.Expr(exprTokens)
		tokens = tokens[len(exprTokens):] // Skip expr tokens
	}
	index := new(int)
	blockTokens := getRange(index, "{", "}", tokens)
	if blockTokens == nil {
		b.PushError(tokens[0], "invalid_syntax")
		return
	}
	if *index < len(tokens) {
		b.PushError(tokens[*index], "invalid_syntax")
	}
	iter.Block = b.Block(blockTokens)
	return StatementAST{iter.Token, iter}
}

func (b *Builder) BreakStatement(tokens []lex.Token) StatementAST {
	var breakAST BreakAST
	breakAST.Token = tokens[0]
	if len(tokens) > 1 {
		b.PushError(tokens[1], "invalid_syntax")
	}
	return StatementAST{breakAST.Token, breakAST}
}

func (b *Builder) ContinueStatement(tokens []lex.Token) StatementAST {
	var continueAST ContinueAST
	continueAST.Token = tokens[0]
	if len(tokens) > 1 {
		b.PushError(tokens[1], "invalid_syntax")
	}
	return StatementAST{continueAST.Token, continueAST}
}

// Expr builds AST model of expression.
func (b *Builder) Expr(tokens []lex.Token) (e ExprAST) {
	e.Processes = b.getExprProcesses(tokens)
	e.Tokens = tokens
	return
}

func (b *Builder) getExprProcesses(tokens []lex.Token) [][]lex.Token {
	var processes [][]lex.Token
	var part []lex.Token
	operator := false
	value := false
	braceCount := 0
	pushedError := false
	singleOperatored := false
	newKeyword := false
	for index := 0; index < len(tokens); index++ {
		token := tokens[index]
		switch token.Id {
		case lex.Operator:
			if newKeyword {
				part = append(part, token)
				continue
			}
			if !operator {
				if IsSingleOperator(token.Kind) && !singleOperatored {
					part = append(part, token)
					singleOperatored = true
					continue
				}
				if braceCount == 0 {
					b.PushError(token, "operator_overflow")
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
					if _, ok := b.DataType(tokens, &index, false); ok {
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
		case lex.New:
			newKeyword = true
		case lex.Name:
			if braceCount == 0 {
				newKeyword = false
			}
		}
		if index > 0 && braceCount == 0 {
			lt := tokens[index-1]
			if (lt.Id == lex.Name || lt.Id == lex.Value) &&
				(token.Id == lex.Name || token.Id == lex.Value) {
				b.PushError(token, "invalid_syntax")
				pushedError = true
			}
		}
		b.checkExprToken(token)
		part = append(part, token)
		operator = requireOperatorForProcess(token, index, len(tokens))
		value = false
	}
	if len(part) > 0 {
		processes = append(processes, part)
	}
	if value {
		b.PushError(processes[len(processes)-1][0], "operator_overflow")
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

func (b *Builder) checkExprToken(token lex.Token) {
	if token.Kind[0] >= '0' && token.Kind[0] <= '9' {
		var result bool
		if strings.IndexByte(token.Kind, '.') != -1 {
			_, result = new(big.Float).SetString(token.Kind)
		} else {
			result = xbits.CheckBitInt(token.Kind, 64)
		}
		if !result {
			b.PushError(token, "invalid_numeric_range")
		}
	}
}

func getRange(index *int, open, close string, tokens []lex.Token) []lex.Token {
	token := tokens[*index]
	if token.Id == lex.Brace && token.Kind == open {
		*index++
		braceCount := 1
		start := *index
		for ; braceCount > 0 && *index < len(tokens); *index++ {
			token := tokens[*index]
			if token.Id != lex.Brace {
				continue
			}
			if token.Kind == open {
				braceCount++
			} else if token.Kind == close {
				braceCount--
			}
		}
		return tokens[start : *index-1]
	}
	return nil
}

func (b *Builder) skipStatement() []lex.Token {
	start := b.Position
	b.Position = nextStatementPos(b.Tokens, start)
	return b.Tokens[start:b.Position]
}
