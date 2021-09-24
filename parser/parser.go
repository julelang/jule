package parser

import (
	"fmt"
	"strings"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
	"github.com/the-xlang/x/pkg/xbits"
)

// Parser is parser of X code.
type Parser struct {
	attributes []ast.AttributeAST

	Functions              []*function
	GlobalVariables        []ast.VariableAST
	WaitingGlobalVariables []ast.VariableAST
	BlockVariables         []ast.VariableAST
	Tokens                 []lex.Token
	PFI                    *ParseFileInfo
}

// NewParser returns new instance of Parser.
func NewParser(tokens []lex.Token, PFI *ParseFileInfo) *Parser {
	parser := new(Parser)
	parser.Tokens = tokens
	parser.PFI = PFI
	return parser
}

// PushErrorToken appends new error by token.
func (p *Parser) PushErrorToken(token lex.Token, err string) {
	message := x.Errors[err]
	p.PFI.Errors = append(p.PFI.Errors, fmt.Sprintf(
		"%s:%d:%d %s", token.File.Path, token.Line, token.Column, message))
}

// AppendErrors appends specified errors.
func (p *Parser) AppendErrors(errors ...string) {
	p.PFI.Errors = append(p.PFI.Errors, errors...)
}

// PushError appends new error.
func (p *Parser) PushError(err string) {
	p.PFI.Errors = append(p.PFI.Errors, x.Errors[err])
}

// String is returns full C++ code of parsed objects.
func (p Parser) String() string {
	return p.Cxx()
}

// Cxx is returns full C++ code of parsed objects.
func (p *Parser) Cxx() string {
	var sb strings.Builder
	sb.WriteString("#pragma region X_GLOBAL_VARIABLES")
	sb.WriteByte('\n')
	for _, va := range p.GlobalVariables {
		sb.WriteString(va.String())
		sb.WriteByte('\n')
	}
	sb.WriteString("#pragma endregion X_GLOBAL_VARIABLES")
	sb.WriteString("\n\n")
	sb.WriteString("#pragma region X_FUNCTIONS")
	sb.WriteString("\n\n")
	for _, fun := range p.Functions {
		sb.WriteString(fun.String())
		sb.WriteString("\n\n")
	}
	sb.WriteString("#pragma endregion X_FUNCTIONS")
	return sb.String()
}

// Parse is parse X code.
//
//! This function is main point of parsing.
func (p *Parser) Parse() {
	astModel := ast.New(p.Tokens)
	astModel.Build()
	if astModel.Errors != nil {
		p.PFI.Errors = append(p.PFI.Errors, astModel.Errors...)
		return
	}
	for _, model := range astModel.Tree {
		switch model.Type {
		case ast.Attribute:
			p.PushAttribute(model.Value.(ast.AttributeAST))
		case ast.Statement:
			p.ParseStatement(model.Value.(ast.StatementAST))
		default:
			p.PushErrorToken(model.Token, "invalid_syntax")
		}
	}
	p.finalCheck()
}

// PushAttribute processes and appends to attribute list.
func (p *Parser) PushAttribute(t ast.AttributeAST) {
	switch t.Token.Value {
	case "inline":
	default:
		p.PushErrorToken(t.Token, "invalid_syntax")
	}
	p.attributes = append(p.attributes, t)
}

// ParseStatement parse X statement.
func (p *Parser) ParseStatement(s ast.StatementAST) {
	switch t := s.Value.(type) {
	case ast.FunctionAST:
		p.ParseFunction(t)
	case ast.VariableAST:
		p.ParseGlobalVariable(t)
	default:
		p.PushErrorToken(s.Token, "invalid_syntax")
	}
}

// ParseFunction parse X function.
func (p *Parser) ParseFunction(funAst ast.FunctionAST) {
	if p.existName(funAst.Name).Type != ast.NA {
		p.PushErrorToken(funAst.Token, "exist_name")
		return
	}
	fun := new(function)
	fun.Token = funAst.Token
	fun.Name = funAst.Name
	fun.ReturnType = funAst.ReturnType
	fun.Block = funAst.Block
	fun.Params = funAst.Params
	fun.Attributes = p.attributes
	p.attributes = nil
	p.checkFunctionAttributes(fun.Attributes)
	p.Functions = append(p.Functions, fun)
}

// ParseVariable parse X global variable.
func (p *Parser) ParseGlobalVariable(varAST ast.VariableAST) {
	if p.existName(varAST.Name).Type != ast.NA {
		p.PushErrorToken(varAST.Token, "exist_name")
		return
	}
	p.WaitingGlobalVariables = append(p.WaitingGlobalVariables, varAST)
}

// ParseWaitingGlobalVariables parse X global variables for waiting parsing.
func (p *Parser) ParseWaitingGlobalVariables() {
	for _, varAST := range p.WaitingGlobalVariables {
		p.GlobalVariables = append(p.GlobalVariables, p.ParseVariable(varAST))
	}
}

// ParseVariable parse X variable.
func (p *Parser) ParseVariable(varAST ast.VariableAST) ast.VariableAST {
	value := p.computeExpression(varAST.Value)
	if varAST.Type.Code != x.Void {
		if typeIsSingle(varAST.Type) && typeIsSingle(value.Type) {
			if !x.TypesAreCompatible(varAST.Type.Code, value.Type.Code, true) {
				p.PushErrorToken(varAST.Token, "incompatible_datatype")
			}
		} else {
			if varAST.Type.Value != value.Type.Value {
				p.PushErrorToken(varAST.Token, "incompatible_datatype")
			}
		}
	} else {
		varAST.Type = value.Type
	}
	return varAST
}

func (p *Parser) checkFunctionAttributes(attributes []ast.AttributeAST) {
	for _, attribute := range attributes {
		switch attribute.Token.Value {
		case "inline":
		default:
			p.PushErrorToken(attribute.Token, "invalid_attribute")
		}
	}
}

func variablesFromParameters(params []ast.ParameterAST) []ast.VariableAST {
	var vars []ast.VariableAST
	for _, param := range params {
		var variable ast.VariableAST
		variable.Name = param.Name
		variable.Token = param.Token
		variable.Type = param.Type
		vars = append(vars, variable)
	}
	return vars
}

func (p *Parser) checkFunctionReturn(fun *function) {
	miss := true
	for _, s := range fun.Block.Content {
		switch t := s.Value.(type) {
		case ast.ReturnAST:
			if len(t.Expression.Tokens) == 0 {
				if fun.ReturnType.Code != x.Void {
					p.PushErrorToken(t.Token, "require_return_value")
				}
			} else {
				if fun.ReturnType.Code == x.Void {
					p.PushErrorToken(t.Token, "void_function_return_value")
				}
				value := p.computeExpression(t.Expression)
				if typeIsSingle(value.Type) && typeIsSingle(fun.ReturnType) {
					if !x.TypesAreCompatible(value.Type.Code, fun.ReturnType.Code, true) {
						p.PushErrorToken(t.Token, "incompatible_type")
					}
				} else if fun.ReturnType.Code != x.Any {
					if value.Type.Value != fun.ReturnType.Value {
						p.PushErrorToken(t.Token, "incompatible_type")
					}
				}
			}
			miss = false
		}
	}
	if miss && fun.ReturnType.Code != x.Void {
		p.PushErrorToken(fun.Token, "missing_return")
	}
}

func (p *Parser) functionByName(name string) *function {
	for _, fun := range builtinFunctions {
		if fun.Name == name {
			return fun
		}
	}
	for _, fun := range p.Functions {
		if fun.Name == name {
			return fun
		}
	}
	return nil
}

func (p *Parser) variableByName(name string) *ast.VariableAST {
	for _, variable := range p.BlockVariables {
		if variable.Name == name {
			return &variable
		}
	}
	for _, variable := range p.GlobalVariables {
		if variable.Name == name {
			return &variable
		}
	}
	return nil
}

func (p *Parser) existName(name string) lex.Token {
	fun := p.functionByName(name)
	if fun != nil {
		return fun.Token
	}
	variable := p.variableByName(name)
	if variable != nil {
		return variable.Token
	}
	for _, varAST := range p.WaitingGlobalVariables {
		if varAST.Name == name {
			return varAST.Token
		}
	}
	return lex.Token{}
}

func (p *Parser) finalCheck() {
	if p.functionByName(x.EntryPoint) == nil {
		p.PushError("no_entry_point")
	}
	p.ParseWaitingGlobalVariables()
	p.WaitingGlobalVariables = nil
	p.checkFunctions()
}

func (p *Parser) checkFunctions() {
	for _, fun := range p.Functions {
		p.BlockVariables = variablesFromParameters(fun.Params)
		p.checkFunction(fun)
		p.checkBlock(fun.Block)
		p.checkFunctionReturn(fun)
	}
}

func (p *Parser) computeProcesses(processes [][]lex.Token) ast.ValueAST {
	if processes == nil {
		return ast.ValueAST{}
	}
	if len(processes) == 1 {
		value := p.processValuePart(processes[0])
		return value
	}
	var process arithmeticProcess
	var value ast.ValueAST
	process.cp = p
	j := p.nextOperator(processes)
	boolean := false
	for j != -1 {
		if !boolean {
			boolean = value.Type.Code == x.Bool
		}
		if boolean {
			value.Type.Code = x.Bool
		}
		if j == 0 {
			process.leftVal = value
			process.operator = processes[j][0]
			process.right = processes[j+1]
			process.rightVal = p.processValuePart(process.right)
			value = process.solve()
			processes = processes[2:]
			j = p.nextOperator(processes)
			continue
		} else if j == len(processes)-1 {
			process.operator = processes[j][0]
			process.left = processes[j-1]
			process.leftVal = p.processValuePart(process.left)
			process.rightVal = value
			value = process.solve()
			processes = processes[:j-1]
			j = p.nextOperator(processes)
			continue
		} else if prev := processes[j-1]; prev[0].Type == lex.Operator &&
			len(prev) == 1 {
			process.leftVal = value
			process.operator = processes[j][0]
			process.right = processes[j+1]
			process.rightVal = p.processValuePart(process.right)
			value = process.solve()
			processes = append(processes[:j], processes[j+2:]...)
			j = p.nextOperator(processes)
			continue
		}
		process.left = processes[j-1]
		process.leftVal = p.processValuePart(process.left)
		process.operator = processes[j][0]
		process.right = processes[j+1]
		process.rightVal = p.processValuePart(process.right)
		solvedValue := process.solve()
		if value.Type.Code != ast.NA {
			process.operator.Value = "+"
			process.leftVal = value
			process.right = processes[j+1]
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
		j = p.nextOperator(processes)
	}
	return value
}

func (p *Parser) computeTokens(tokens []lex.Token) ast.ValueAST {
	return p.computeProcesses(new(ast.AST).BuildExpression(tokens).Processes)
}

func (p *Parser) computeExpression(ex ast.ExpressionAST) ast.ValueAST {
	processes := make([][]lex.Token, len(ex.Processes))
	copy(processes, ex.Processes)
	return p.computeProcesses(processes)
}

// nextOperator find index of priority operator and returns index of operator
// if found, returns -1 if not.
func (p *Parser) nextOperator(tokens [][]lex.Token) int {
	precedence5 := -1
	precedence4 := -1
	precedence3 := -1
	precedence2 := -1
	precedence1 := -1
	for index, part := range tokens {
		if len(part) != 1 {
			continue
		} else if part[0].Type != lex.Operator {
			continue
		}
		switch part[0].Value {
		case "*", "/", "%", "<<", ">>", "&":
			precedence5 = index
		case "+", "-", "|", "^":
			precedence4 = index
		case "==", "!=", "<", "<=", ">", ">=":
			precedence3 = index
		case "&&":
			precedence2 = index
		case "||":
			precedence1 = index
		default:
			p.PushErrorToken(part[0], "invalid_operator")
		}
	}
	if precedence5 != -1 {
		return precedence5
	} else if precedence4 != -1 {
		return precedence4
	} else if precedence3 != -1 {
		return precedence3
	} else if precedence2 != -1 {
		return precedence2
	}
	return precedence1
}

type arithmeticProcess struct {
	cp       *Parser
	left     []lex.Token
	leftVal  ast.ValueAST
	right    []lex.Token
	rightVal ast.ValueAST
	operator lex.Token
}

func (ap arithmeticProcess) solvePointer() (v ast.ValueAST) {
	if ap.leftVal.Type.Value != ap.rightVal.Type.Value {
		ap.cp.PushErrorToken(ap.operator, "incompatible_type")
		return
	}
	switch ap.operator.Value {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		ap.cp.PushErrorToken(ap.operator, "operator_notfor_pointer")
	}
	return
}

func (ap arithmeticProcess) solveString() (v ast.ValueAST) {
	// Not both string?
	if ap.leftVal.Type != ap.rightVal.Type {
		ap.cp.PushErrorToken(ap.operator, "incompatible_datatype")
		return
	}
	switch ap.operator.Value {
	case "+":
		v.Type.Code = x.Str
	case "==", "!=":
		v.Type.Code = x.Bool
	default:
		ap.cp.PushErrorToken(ap.operator, "operator_notfor_strings")
	}
	return
}

func (ap arithmeticProcess) solveAny() (v ast.ValueAST) {
	switch ap.operator.Value {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		ap.cp.PushErrorToken(ap.operator, "operator_notfor_any")
	}
	return
}

func (ap arithmeticProcess) solveBool() (v ast.ValueAST) {
	if !x.TypesAreCompatible(ap.leftVal.Type.Code, ap.rightVal.Type.Code, true) {
		ap.cp.PushErrorToken(ap.operator, "incompatible_type")
		return
	}
	switch ap.operator.Value {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		ap.cp.PushErrorToken(ap.operator, "operator_notfor_bool")
	}
	return
}

func (ap arithmeticProcess) solveFloat() (v ast.ValueAST) {
	if !x.TypesAreCompatible(ap.leftVal.Type.Code, ap.rightVal.Type.Code, true) {
		if !isConstantNumeric(ap.leftVal.Value) &&
			!isConstantNumeric(ap.rightVal.Value) {
			ap.cp.PushErrorToken(ap.operator, "incompatible_type")
			return
		}
	}
	switch ap.operator.Value {
	case "!=", "==", "<", ">", ">=", "<=":
		v.Type.Code = x.Bool
	case "+", "-", "*", "/":
		v.Type.Code = x.Float32
		if ap.leftVal.Type.Code == x.Float64 || ap.rightVal.Type.Code == x.Float64 {
			v.Type.Code = x.Float64
		}
	default:
		ap.cp.PushErrorToken(ap.operator, "operator_notfor_float")
	}
	return
}

func (ap arithmeticProcess) solveSigned() (v ast.ValueAST) {
	if !x.TypesAreCompatible(ap.leftVal.Type.Code, ap.rightVal.Type.Code, true) {
		if !isConstantNumeric(ap.leftVal.Value) &&
			!isConstantNumeric(ap.rightVal.Value) {
			ap.cp.PushErrorToken(ap.operator, "incompatible_type")
			return
		}
	}
	switch ap.operator.Value {
	case "!=", "==", "<", ">", ">=", "<=":
		v.Type.Code = x.Bool
	case "+", "-", "*", "/", "%", "&", "|", "^":
		v.Type = ap.leftVal.Type
		if x.TypeGreaterThan(ap.rightVal.Type.Code, v.Type.Code) {
			v.Type = ap.rightVal.Type
		}
	case ">>", "<<":
		v.Type = ap.leftVal.Type
		if !x.IsUnsignedNumericType(ap.rightVal.Type.Code) &&
			!checkIntBit(ap.rightVal, xbits.BitsizeOfType(x.UInt64)) {
			ap.cp.PushErrorToken(ap.rightVal.Token, "bitshift_must_unsigned")
		}
	default:
		ap.cp.PushErrorToken(ap.operator, "operator_notfor_int")
	}
	return
}

func (ap arithmeticProcess) solveUnsigned() (v ast.ValueAST) {
	if !x.TypesAreCompatible(ap.leftVal.Type.Code, ap.rightVal.Type.Code, true) {
		if !isConstantNumeric(ap.leftVal.Value) &&
			!isConstantNumeric(ap.rightVal.Value) {
			ap.cp.PushErrorToken(ap.operator, "incompatible_type")
			return
		}
		return
	}
	switch ap.operator.Value {
	case "!=", "==", "<", ">", ">=", "<=":
		v.Type.Code = x.Bool
	case "+", "-", "*", "/", "%", "&", "|", "^":
		v.Type = ap.leftVal.Type
		if x.TypeGreaterThan(ap.rightVal.Type.Code, v.Type.Code) {
			v.Type = ap.rightVal.Type
		}
	default:
		ap.cp.PushErrorToken(ap.operator, "operator_notfor_uint")
	}
	return
}

func (ap arithmeticProcess) solveLogical() (v ast.ValueAST) {
	v.Type.Code = x.Bool
	if ap.leftVal.Type.Code != x.Bool {
		ap.cp.PushErrorToken(ap.leftVal.Token, "logical_not_bool")
	}
	if ap.rightVal.Type.Code != x.Bool {
		ap.cp.PushErrorToken(ap.rightVal.Token, "logical_not_bool")
	}
	return
}

func (ap arithmeticProcess) solve() (v ast.ValueAST) {
	switch ap.operator.Value {
	case "+", "-", "*", "/", "%", ">>",
		"<<", "&", "|", "^", "==", "!=",
		">=", "<=", ">", "<":
	case "&&", "||":
		return ap.solveLogical()
	default:
		ap.cp.PushErrorToken(ap.operator, "invalid_operator")
	}
	switch {
	case typeIsPointer(ap.leftVal.Type) || typeIsPointer(ap.rightVal.Type):
		return ap.solvePointer()
	case ap.leftVal.Type.Code == x.Any || ap.rightVal.Type.Code == x.Any:
		return ap.solveAny()
	case ap.leftVal.Type.Code == x.Bool || ap.rightVal.Type.Code == x.Bool:
		return ap.solveBool()
	case ap.leftVal.Type.Code == x.Str || ap.rightVal.Type.Code == x.Str:
		return ap.solveString()
	case x.IsFloatType(ap.leftVal.Type.Code) ||
		x.IsFloatType(ap.rightVal.Type.Code):
		return ap.solveFloat()
	case x.IsSignedNumericType(ap.leftVal.Type.Code) ||
		x.IsSignedNumericType(ap.rightVal.Type.Code):
		return ap.solveSigned()
	case x.IsUnsignedNumericType(ap.leftVal.Type.Code) ||
		x.IsUnsignedNumericType(ap.rightVal.Type.Code):
		return ap.solveUnsigned()
	}
	return
}

const functionName = 0x0000A

func (p *Parser) processSingleValuePart(token lex.Token) (v ast.ValueAST, ok bool) {
	v.Type.Code = ast.NA
	v.Token = token
	switch token.Type {
	case lex.Value:
		if IsString(token.Value) {
			// result.Value = token.Value[1 : len(token.Value)-1]
			v.Value = "L" + token.Value
			v.Type.Code = x.Str
			v.Type.Value = "str"
			ok = true
		} else if IsBoolean(token.Value) {
			v.Value = token.Value
			v.Type.Code = x.Bool
			v.Type.Value = "bool"
			ok = true
		} else { // Numeric.
			if strings.Contains(token.Value, ".") ||
				strings.ContainsAny(token.Value, "eE") {
				v.Type.Code = x.Float64
				v.Type.Value = "float64"
			} else {
				v.Type.Code = x.Int32
				v.Type.Value = "int32"
				ok := xbits.CheckBitInt(token.Value, 32)
				if !ok {
					v.Type.Code = x.Int64
					v.Type.Value = "int64"
				}
			}
			v.Value = token.Value
			ok = true
		}
	case lex.Name:
		if variable := p.variableByName(token.Value); variable != nil {
			v.Value = token.Value
			v.Type = variable.Type
			ok = true
		} else if p.functionByName(token.Value) != nil {
			v.Value = token.Value
			v.Type.Code = functionName
			ok = true
		} else {
			p.PushErrorToken(token, "name_not_defined")
		}
	default:
		p.PushErrorToken(token, "invalid_syntax")
	}
	return
}

func typeIsPointer(t ast.TypeAST) bool {
	if t.Value == "" {
		return false
	}
	return t.Value[0] == '*'
}

func typeIsSingle(t ast.TypeAST) bool {
	return !typeIsPointer(t)
}

func (p *Parser) processSingleOperatorPart(tokens []lex.Token) ast.ValueAST {
	var v ast.ValueAST
	token := tokens[0]
	//? Length is 1 caouse all lengths of operators is 1,
	//? change "1" with length of token's valaue
	//? if all operators length is not 1.
	tokens = tokens[1:]
	if len(tokens) == 0 {
		p.PushErrorToken(token, "invalid_syntax")
		return v
	}
	switch token.Value {
	case "-":
		v = p.processValuePart(tokens)
		if !typeIsSingle(v.Type) {
			p.PushErrorToken(token, "invalid_data_unary")
		} else if !x.IsNumericType(v.Type.Code) {
			p.PushErrorToken(token, "invalid_data_unary")
		}
	case "+":
		v = p.processValuePart(tokens)
		if !typeIsSingle(v.Type) {
			p.PushErrorToken(token, "invalid_data_unary")
		} else if !x.IsNumericType(v.Type.Code) {
			p.PushErrorToken(token, "invalid_data_plus")
		}
	case "~":
		v = p.processValuePart(tokens)
		if !typeIsSingle(v.Type) {
			p.PushErrorToken(token, "invalid_data_unary")
		} else if !x.IsIntegerType(v.Type.Code) {
			p.PushErrorToken(token, "invalid_data_tilde")
		}
	case "!":
		v = p.processValuePart(tokens)
		if !typeIsSingle(v.Type) {
			p.PushErrorToken(token, "invalid_data_unary")
		} else if v.Type.Code != x.Bool {
			p.PushErrorToken(token, "invalid_data_logical_not")
		}
	case "*":
		v = p.processValuePart(tokens)
		if !typeIsPointer(v.Type) {
			p.PushErrorToken(token, "invalid_data_star")
		}
		v.Type.Value = v.Type.Value[1:]
	case "&":
		v = p.processValuePart(tokens)
		if v.Token.Type != lex.Name {
			p.PushErrorToken(token, "invalid_data_amper")
		}
		v.Type.Value = "*" + v.Type.Value
	default:
		p.PushErrorToken(token, "invalid_syntax")
	}
	v.Token = token
	return v
}

func (p *Parser) processValuePart(tokens []lex.Token) (v ast.ValueAST) {
	if tokens[0].Type == lex.Operator {
		return p.processSingleOperatorPart(tokens)
	} else if len(tokens) == 1 {
		value, ok := p.processSingleValuePart(tokens[0])
		if ok {
			v = value
			goto end
		}
	}
	switch token := tokens[len(tokens)-1]; token.Type {
	case lex.Brace:
		switch token.Value {
		case ")":
			return p.processParenthesesValuePart(tokens)
		}
	default:
		p.PushErrorToken(tokens[0], "invalid_syntax")
	}
end:
	return
}

func (p *Parser) processParenthesesValuePart(tokens []lex.Token) ast.ValueAST {
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
			p.PushErrorToken(tk, "invalid_syntax")
		}
		return p.computeTokens(tokens)
	}
	value := p.processValuePart(valueTokens)
	switch value.Type.Code {
	case functionName:
		fun := p.functionByName(value.Value)
		p.parseFunctionCallStatement(fun, tokens[len(valueTokens):])
		value.Type = fun.ReturnType
	default:
		p.PushErrorToken(tokens[len(valueTokens)], "invalid_syntax")
	}
	return value
}

func (p *Parser) parseFunctionCallStatement(fun *function, tokens []lex.Token) {
	errToken := tokens[0]
	tokens = p.getRangeTokens("(", ")", tokens)
	if tokens == nil {
		tokens = make([]lex.Token, 0)
	}
	ast := new(ast.AST)
	args := ast.BuildArgs(tokens)
	if len(ast.Errors) > 0 {
		p.AppendErrors(ast.Errors...)
	}
	p.parseArgs(fun, args, errToken)
}

func (p *Parser) parseArgs(fun *function, args []ast.ArgAST, errToken lex.Token) {
	if len(args) < len(fun.Params) {
		p.PushErrorToken(errToken, "missing_argument")
	}
	for index, arg := range args {
		p.parseArg(fun, index, arg)
	}
}

func (p *Parser) parseArg(fun *function, index int, arg ast.ArgAST) {
	if index >= len(fun.Params) {
		p.PushErrorToken(arg.Token, "argument_overflow")
		return
	}
	value := p.computeExpression(arg.Expression)
	param := fun.Params[index]
	if typeIsSingle(value.Type) && typeIsSingle(param.Type) {
		if !x.TypesAreCompatible(value.Type.Code, param.Type.Code, false) {
			value.Type = param.Type
			if !checkIntBit(value, xbits.BitsizeOfType(param.Type.Code)) {
				p.PushErrorToken(arg.Token, "incompatible_type")
			}
		}
	} else if param.Type.Code != x.Any {
		if value.Type.Value != param.Type.Value {
			p.PushErrorToken(arg.Token, "incompatible_type")
		}
	}
}

func (p *Parser) getRangeTokens(open, close string, tokens []lex.Token) []lex.Token {
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
	p.PushErrorToken(tokens[0], "brace_not_closed")
	return nil
}

func (p *Parser) checkFunction(fun *function) {
	switch fun.Name {
	case x.EntryPoint:
		if len(fun.Params) > 0 {
			p.PushErrorToken(fun.Token, "entrypoint_have_parameters")
		}
		if fun.ReturnType.Code != x.Void {
			p.PushErrorToken(fun.ReturnType.Token, "entrypoint_have_return")
		}
	}
}

func (p *Parser) checkBlock(b ast.BlockAST) {
	for index, model := range b.Content {
		switch t := model.Value.(type) {
		case ast.FunctionCallAST:
			p.checkFunctionCallStatement(t)
		case ast.VariableAST:
			p.checkVariableStatement(&t)
			model.Value = t
			b.Content[index] = model
		case ast.ReturnAST:
		default:
			p.PushErrorToken(model.Token, "invalid_syntax")
		}
	}
}

func (p *Parser) checkFunctionCallStatement(cs ast.FunctionCallAST) {
	fun := p.functionByName(cs.Name)
	if fun == nil {
		p.PushErrorToken(cs.Token, "name_not_defined")
		return
	}
	p.parseArgs(fun, cs.Args, cs.Token)
}

func (p *Parser) checkVariableStatement(varAST *ast.VariableAST) {
	for _, variable := range p.BlockVariables {
		if varAST.Name == variable.Name {
			p.PushErrorToken(varAST.Token, "exist_name")
			break
		}
	}
	*varAST = p.ParseVariable(*varAST)
	p.BlockVariables = append(p.BlockVariables, *varAST)
}

func isConstantNumeric(v string) bool {
	if v == "" {
		return false
	}
	return v[0] >= '0' && v[0] <= '9'
}

func checkIntBit(v ast.ValueAST, bit int) bool {
	if bit == 0 {
		return false
	}
	if x.IsSignedNumericType(v.Type.Code) {
		return xbits.CheckBitInt(v.Value, bit)
	}
	return xbits.CheckBitUInt(v.Value, bit)
}
