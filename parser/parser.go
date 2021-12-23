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
	Types                  []ast.TypeAST
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
		"%s:%d:%d %s", token.File.Path, token.Row, token.Column, message))
}

// AppendErrors appends specified errors.
func (p *Parser) AppendErrors(errors ...string) {
	p.PFI.Errors = append(p.PFI.Errors, errors...)
}

// PushError appends new error.
func (p *Parser) PushError(err string) {
	p.PFI.Errors = append(p.PFI.Errors, x.Errors[err])
}

// String returns full C++ code of parsed objects.
func (p Parser) String() string {
	return p.Cxx()
}

// CxxTypes returns C++ code developer-defined types.
func (p *Parser) CxxTypes() string {
	var cxx strings.Builder
	cxx.WriteString("#pragma region TYPES\n")
	for _, t := range p.Types {
		cxx.WriteString(t.String())
		cxx.WriteByte('\n')
	}
	cxx.WriteString("#pragma endregion TYPES")
	return cxx.String()
}

// CxxPrototypes returns C++ code of prototypes of C++ code.
func (p *Parser) CxxPrototypes() string {
	var cxx strings.Builder
	cxx.WriteString("#pragma region PROTOTYPES\n")
	for _, fun := range p.Functions {
		cxx.WriteString(fun.Prototype())
		cxx.WriteByte('\n')
	}
	cxx.WriteString("#pragma endregion PROTOTYPES")
	return cxx.String()
}

// CxxGlobalVariables returns C++ code of global variables.
func (p *Parser) CxxGlobalVariables() string {
	var cxx strings.Builder
	cxx.WriteString("#pragma region GLOBAL_VARIABLES\n")
	for _, va := range p.GlobalVariables {
		cxx.WriteString(va.String())
		cxx.WriteByte('\n')
	}
	cxx.WriteString("#pragma endregion GLOBAL_VARIABLES")
	return cxx.String()
}

// CxxFunctions returns C++ code of functions.
func (p *Parser) CxxFunctions() string {
	var cxx strings.Builder
	cxx.WriteString("#pragma region FUNCTIONS")
	cxx.WriteString("\n\n")
	for _, fun := range p.Functions {
		cxx.WriteString(fun.String())
		cxx.WriteString("\n\n")
	}
	cxx.WriteString("#pragma endregion FUNCTIONS")
	return cxx.String()
}

// Cxx returns full C++ code of parsed objects.
func (p *Parser) Cxx() string {
	var cxx strings.Builder
	cxx.WriteString(p.CxxTypes() + "\n\n")
	cxx.WriteString(p.CxxPrototypes() + "\n\n")
	cxx.WriteString(p.CxxGlobalVariables() + "\n\n")
	cxx.WriteString(p.CxxFunctions())
	return cxx.String()
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
		switch t := model.Value.(type) {
		case ast.AttributeAST:
			p.PushAttribute(t)
		case ast.StatementAST:
			p.ParseStatement(t)
		case ast.TypeAST:
			p.ParseType(t)
		default:
			p.PushErrorToken(model.Token, "invalid_syntax")
		}
	}
	p.finalCheck()
}

// ParseType parse X statement.
func (p *Parser) ParseType(t ast.TypeAST) {
	if p.existName(t.Name).Id != lex.NA {
		p.PushErrorToken(t.Token, "exist_name")
		return
	}
	p.Types = append(p.Types, t)
}

// PushAttribute processes and appends to attribute list.
func (p *Parser) PushAttribute(attribute ast.AttributeAST) {
	switch attribute.Tag.Kind {
	case "_inline":
	default:
		p.PushErrorToken(attribute.Tag, "undefined_tag")
	}
	for _, attr := range p.attributes {
		if attr.Tag.Kind == attribute.Tag.Kind {
			p.PushErrorToken(attribute.Tag, "attribute_repeat")
			return
		}
	}
	p.attributes = append(p.attributes, attribute)
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
	if p.existName(funAst.Name).Id != lex.NA {
		p.PushErrorToken(funAst.Token, "exist_name")
		return
	}
	fun := new(function)
	fun.ast = funAst
	fun.attributes = p.attributes
	p.attributes = nil
	p.checkFunctionAttributes(fun.attributes)
	p.Functions = append(p.Functions, fun)
}

// ParseVariable parse X global variable.
func (p *Parser) ParseGlobalVariable(varAST ast.VariableAST) {
	if p.existName(varAST.Name).Id != lex.NA {
		p.PushErrorToken(varAST.NameToken, "exist_name")
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
	value, model := p.computeExpression(varAST.Value)
	varAST.Value.Model = model
	if varAST.Type.Code != x.Void {
		if varAST.SetterToken.Id != lex.NA { // Pass default value.
			p.checkType(varAST.Type, value.ast.Type, false, varAST.NameToken)
		} else {
			var valueToken lex.Token
			valueToken.Id = lex.Value
			dt, ok := p.readyType(varAST.Type)
			if ok {
				valueToken.Kind = p.defaultValueOfType(dt)
				valueTokens := []lex.Token{valueToken}
				varAST.Value = ast.ExpressionAST{
					Tokens:    valueTokens,
					Processes: [][]lex.Token{valueTokens},
				}
			}
		}
	} else {
		varAST.Type = value.ast.Type
		p.checkValidityForAutoType(varAST.Type, varAST.SetterToken)
	}
	if varAST.DefineToken.Kind == "const" {
		if varAST.SetterToken.Id == lex.NA {
			p.PushErrorToken(varAST.NameToken, "missing_const_value")
		} else if !checkValidityConstantDataType(varAST.Type) {
			p.PushErrorToken(varAST.NameToken, "invalid_const_data_type")
		}
	}
	return varAST
}

func (p *Parser) checkFunctionAttributes(attributes []ast.AttributeAST) {
	for _, attribute := range attributes {
		switch attribute.Tag.Kind {
		case "_inline":
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
		variable.NameToken = param.Token
		variable.Type = param.Type
		if param.Const {
			variable.DefineToken.Id = lex.Const
		}
		vars = append(vars, variable)
	}
	return vars
}

func (p *Parser) checkFunctionReturn(fun ast.FunctionAST) {
	missed := true
	for index, s := range fun.Block.Statements {
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
				value, model := p.computeExpression(t.Expression)
				t.Expression.Model = model
				fun.Block.Statements[index].Value = t
				p.checkType(fun.ReturnType, value.ast.Type, true, t.Token)
			}
			missed = false
		}
	}
	if missed && fun.ReturnType.Code != x.Void {
		p.PushErrorToken(fun.Token, "missing_return")
	}
}

func (p *Parser) typeByName(name string) *ast.TypeAST {
	for _, t := range p.Types {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

func (p *Parser) functionByName(name string) *function {
	for _, fun := range builtinFunctions {
		if fun.ast.Name == name {
			return fun
		}
	}
	for _, fun := range p.Functions {
		if fun.ast.Name == name {
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
	t := p.typeByName(name)
	if t != nil {
		return t.Token
	}
	fun := p.functionByName(name)
	if fun != nil {
		return fun.ast.Token
	}
	variable := p.variableByName(name)
	if variable != nil {
		return variable.NameToken
	}
	for _, varAST := range p.WaitingGlobalVariables {
		if varAST.Name == name {
			return varAST.NameToken
		}
	}
	return lex.Token{}
}

func (p *Parser) finalCheck() {
	if p.functionByName("_"+x.EntryPoint) == nil {
		p.PushError("no_entry_point")
	}
	p.checkTypes()
	p.ParseWaitingGlobalVariables()
	p.WaitingGlobalVariables = nil
	p.checkFunctions()
}

func (p *Parser) checkTypes() {
	for _, t := range p.Types {
		_, ok := p.readyType(t.Type)
		if !ok {
			p.PushErrorToken(t.Token, "invalid_type_source")
		}
	}
}

func (p *Parser) checkFunctions() {
	for _, fun := range p.Functions {
		p.BlockVariables = variablesFromParameters(fun.ast.Params)
		p.checkFunctionSpecialCases(fun)
		p.checkFunction(fun.ast)
	}
}

type value struct {
	ast      ast.ValueAST
	constant bool
}

func (p *Parser) computeProcesses(processes [][]lex.Token) (v value, e expressionModel) {
	if processes == nil {
		return
	}
	builder := newExpBuilder()
	if len(processes) == 1 {
		builder.setIndex(0)
		v = p.processValuePart(processes[0], builder)
		e = builder.build()
		return
	}
	var process arithmeticProcess
	process.p = p
	j := p.nextOperator(processes)
	boolean := false
	for j != -1 {
		if !boolean {
			boolean = v.ast.Type.Code == x.Bool
		}
		if boolean {
			v.ast.Type.Code = x.Bool
		}
		if j == 0 {
			process.leftVal = v.ast
			process.operator = processes[j][0]
			builder.setIndex(j + 1)
			builder.appendNode(tokenExpNode{process.operator})
			process.right = processes[j+1]
			builder.setIndex(j + 1)
			process.rightVal = p.processValuePart(process.right, builder).ast
			v.ast = process.solve()
			processes = processes[2:]
			goto end
		} else if j == len(processes)-1 {
			process.operator = processes[j][0]
			process.left = processes[j-1]
			builder.setIndex(j - 1)
			process.leftVal = p.processValuePart(process.left, builder).ast
			process.rightVal = v.ast
			builder.setIndex(j)
			builder.appendNode(tokenExpNode{process.operator})
			v.ast = process.solve()
			processes = processes[:j-1]
			goto end
		} else if prev := processes[j-1]; prev[0].Id == lex.Operator &&
			len(prev) == 1 {
			process.leftVal = v.ast
			process.operator = processes[j][0]
			builder.setIndex(j)
			builder.appendNode(tokenExpNode{process.operator})
			process.right = processes[j+1]
			builder.setIndex(j + 1)
			process.rightVal = p.processValuePart(process.right, builder).ast
			v.ast = process.solve()
			processes = append(processes[:j], processes[j+2:]...)
			goto end
		}
		process.left = processes[j-1]
		builder.setIndex(j - 1)
		process.leftVal = p.processValuePart(process.left, builder).ast
		process.operator = processes[j][0]
		builder.setIndex(j)
		builder.appendNode(tokenExpNode{process.operator})
		process.right = processes[j+1]
		builder.setIndex(j + 1)
		process.rightVal = p.processValuePart(process.right, builder).ast
		{
			solvedValue := process.solve()
			if v.ast.Type.Code != x.Void {
				process.operator.Kind = "+"
				process.leftVal = v.ast
				process.right = processes[j+1]
				process.rightVal = solvedValue
				v.ast = process.solve()
			} else {
				v.ast = solvedValue
			}
		}
		// Remove computed processes.
		processes = append(processes[:j-1], processes[j+2:]...)
		if len(processes) == 1 {
			break
		}
	end:
		// Find next operator.
		j = p.nextOperator(processes)
	}
	e = builder.build()
	return
}

func (p *Parser) computeTokens(tokens []lex.Token) (value, expressionModel) {
	return p.computeProcesses(new(ast.AST).BuildExpression(tokens).Processes)
}

func (p *Parser) computeExpression(ex ast.ExpressionAST) (value, expressionModel) {
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
		} else if part[0].Id != lex.Operator {
			continue
		}
		switch part[0].Kind {
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
	p        *Parser
	left     []lex.Token
	leftVal  ast.ValueAST
	right    []lex.Token
	rightVal ast.ValueAST
	operator lex.Token
}

func (ap arithmeticProcess) solvePointer() (v ast.ValueAST) {
	if ap.leftVal.Type.Value != ap.rightVal.Type.Value {
		ap.p.PushErrorToken(ap.operator, "incompatible_type")
		return
	}
	switch ap.operator.Kind {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		ap.p.PushErrorToken(ap.operator, "operator_notfor_pointer")
	}
	return
}

func (ap arithmeticProcess) solveString() (v ast.ValueAST) {
	// Not both string?
	if ap.leftVal.Type.Code != ap.rightVal.Type.Code {
		ap.p.PushErrorToken(ap.operator, "incompatible_datatype")
		return
	}
	switch ap.operator.Kind {
	case "+":
		v.Type.Code = x.Str
	case "==", "!=":
		v.Type.Code = x.Bool
	default:
		ap.p.PushErrorToken(ap.operator, "operator_notfor_string")
	}
	return
}

func (ap arithmeticProcess) solveAny() (v ast.ValueAST) {
	switch ap.operator.Kind {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		ap.p.PushErrorToken(ap.operator, "operator_notfor_any")
	}
	return
}

func (ap arithmeticProcess) solveBool() (v ast.ValueAST) {
	if !typesAreCompatible(ap.leftVal.Type, ap.rightVal.Type, true) {
		ap.p.PushErrorToken(ap.operator, "incompatible_type")
		return
	}
	switch ap.operator.Kind {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		ap.p.PushErrorToken(ap.operator, "operator_notfor_bool")
	}
	return
}

func (ap arithmeticProcess) solveFloat() (v ast.ValueAST) {
	if !typesAreCompatible(ap.leftVal.Type, ap.rightVal.Type, true) {
		if !isConstantNumeric(ap.leftVal.Value) &&
			!isConstantNumeric(ap.rightVal.Value) {
			ap.p.PushErrorToken(ap.operator, "incompatible_type")
			return
		}
	}
	switch ap.operator.Kind {
	case "!=", "==", "<", ">", ">=", "<=":
		v.Type.Code = x.Bool
	case "+", "-", "*", "/":
		v.Type.Code = x.Float32
		if ap.leftVal.Type.Code == x.Float64 || ap.rightVal.Type.Code == x.Float64 {
			v.Type.Code = x.Float64
		}
	default:
		ap.p.PushErrorToken(ap.operator, "operator_notfor_float")
	}
	return
}

func (ap arithmeticProcess) solveSigned() (v ast.ValueAST) {
	if !typesAreCompatible(ap.leftVal.Type, ap.rightVal.Type, true) {
		if !isConstantNumeric(ap.leftVal.Value) &&
			!isConstantNumeric(ap.rightVal.Value) {
			ap.p.PushErrorToken(ap.operator, "incompatible_type")
			return
		}
	}
	switch ap.operator.Kind {
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
			ap.p.PushErrorToken(ap.rightVal.Token, "bitshift_must_unsigned")
		}
	default:
		ap.p.PushErrorToken(ap.operator, "operator_notfor_int")
	}
	return
}

func (ap arithmeticProcess) solveUnsigned() (v ast.ValueAST) {
	if !typesAreCompatible(ap.leftVal.Type, ap.rightVal.Type, true) {
		if !isConstantNumeric(ap.leftVal.Value) &&
			!isConstantNumeric(ap.rightVal.Value) {
			ap.p.PushErrorToken(ap.operator, "incompatible_type")
			return
		}
		return
	}
	switch ap.operator.Kind {
	case "!=", "==", "<", ">", ">=", "<=":
		v.Type.Code = x.Bool
	case "+", "-", "*", "/", "%", "&", "|", "^":
		v.Type = ap.leftVal.Type
		if x.TypeGreaterThan(ap.rightVal.Type.Code, v.Type.Code) {
			v.Type = ap.rightVal.Type
		}
	default:
		ap.p.PushErrorToken(ap.operator, "operator_notfor_uint")
	}
	return
}

func (ap arithmeticProcess) solveLogical() (v ast.ValueAST) {
	v.Type.Code = x.Bool
	if ap.leftVal.Type.Code != x.Bool {
		ap.p.PushErrorToken(ap.leftVal.Token, "logical_not_bool")
	}
	if ap.rightVal.Type.Code != x.Bool {
		ap.p.PushErrorToken(ap.rightVal.Token, "logical_not_bool")
	}
	return
}

func (ap arithmeticProcess) solveRune() (v ast.ValueAST) {
	if !typesAreCompatible(ap.leftVal.Type, ap.rightVal.Type, true) {
		ap.p.PushErrorToken(ap.operator, "incompatible_type")
		return
	}
	switch ap.operator.Kind {
	case "!=", "==", ">", "<", ">=", "<=":
		v.Type.Code = x.Bool
	case "+", "-", "*", "/", "^", "&", "%", "|":
		v.Type.Code = x.Rune
	default:
		ap.p.PushErrorToken(ap.operator, "operator_notfor_rune")
	}
	return
}

func (ap arithmeticProcess) solveArray() (v ast.ValueAST) {
	if !typesAreCompatible(ap.leftVal.Type, ap.rightVal.Type, true) {
		ap.p.PushErrorToken(ap.operator, "incompatible_type")
		return
	}
	switch ap.operator.Kind {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		ap.p.PushErrorToken(ap.operator, "operator_notfor_array")
	}
	return
}

func (ap arithmeticProcess) solveNil() (v ast.ValueAST) {
	if !typesAreCompatible(ap.leftVal.Type, ap.rightVal.Type, false) {
		ap.p.PushErrorToken(ap.operator, "incompatible_type")
		return
	}
	switch ap.operator.Kind {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		ap.p.PushErrorToken(ap.operator, "operator_notfor_nil")
	}
	return
}

func (ap arithmeticProcess) solve() (v ast.ValueAST) {
	switch ap.operator.Kind {
	case "+", "-", "*", "/", "%", ">>",
		"<<", "&", "|", "^", "==", "!=",
		">=", "<=", ">", "<":
	case "&&", "||":
		return ap.solveLogical()
	default:
		ap.p.PushErrorToken(ap.operator, "invalid_operator")
	}
	switch {
	case typeIsArray(ap.leftVal.Type) || typeIsArray(ap.rightVal.Type):
		return ap.solveArray()
	case typeIsPointer(ap.leftVal.Type) || typeIsPointer(ap.rightVal.Type):
		return ap.solvePointer()
	case ap.leftVal.Type.Code == x.Nil || ap.rightVal.Type.Code == x.Nil:
		return ap.solveNil()
	case ap.leftVal.Type.Code == x.Rune || ap.rightVal.Type.Code == x.Rune:
		return ap.solveRune()
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

type singleValueProcessor struct {
	token   lex.Token
	builder *expressionModelBuilder
	parser  *Parser
}

func (p *singleValueProcessor) string() value {
	var v value
	v.ast.Value = p.token.Kind
	v.ast.Type.Code = x.Str
	v.ast.Type.Value = "str"
	p.builder.appendNode(strExpNode{p.token})
	return v
}

func (p *singleValueProcessor) rune() value {
	var v value
	v.ast.Value = p.token.Kind
	v.ast.Type.Code = x.Rune
	v.ast.Type.Value = "rune"
	p.builder.appendNode(runeExpNode{p.token})
	return v
}

func (p *singleValueProcessor) boolean() value {
	var v value
	v.ast.Value = p.token.Kind
	v.ast.Type.Code = x.Bool
	v.ast.Type.Value = "bool"
	p.builder.appendNode(tokenExpNode{p.token})
	return v
}

func (p *singleValueProcessor) nil() value {
	var v value
	v.ast.Value = p.token.Kind
	v.ast.Type.Code = x.Nil
	p.builder.appendNode(tokenExpNode{p.token})
	return v
}

func (p *singleValueProcessor) numeric() value {
	var v value
	if strings.Contains(p.token.Kind, ".") ||
		strings.ContainsAny(p.token.Kind, "eE") {
		v.ast.Type.Code = x.Float64
		v.ast.Type.Value = "float64"
	} else {
		v.ast.Type.Code = x.Int32
		v.ast.Type.Value = "int32"
		ok := xbits.CheckBitInt(p.token.Kind, 32)
		if !ok {
			v.ast.Type.Code = x.Int64
			v.ast.Type.Value = "int64"
		}
	}
	v.ast.Value = p.token.Kind
	p.builder.appendNode(tokenExpNode{p.token})
	return v
}

func (p *singleValueProcessor) name() (v value, ok bool) {
	if variable := p.parser.variableByName(p.token.Kind); variable != nil {
		v.ast.Value = p.token.Kind
		v.ast.Type = variable.Type
		v.constant = variable.DefineToken.Id == lex.Const
		v.ast.Token = variable.NameToken
		p.builder.appendNode(tokenExpNode{p.token})
		ok = true
	} else if fun := p.parser.functionByName(p.token.Kind); fun != nil {
		v.ast.Value = p.token.Kind
		v.ast.Type.Code = x.Function
		v.ast.Type.Tag = fun.ast
		v.ast.Token = fun.ast.Token
		p.builder.appendNode(tokenExpNode{p.token})
		ok = true
	} else {
		p.parser.PushErrorToken(p.token, "name_not_defined")
	}
	return
}

func (p *Parser) processSingleValuePart(token lex.Token, builder *expressionModelBuilder) (v value, ok bool) {
	processor := singleValueProcessor{
		token:   token,
		builder: builder,
		parser:  p,
	}
	v.ast.Type.Code = x.Void
	v.ast.Token = token
	switch token.Id {
	case lex.Value:
		ok = true
		switch {
		case IsString(token.Kind):
			v = processor.string()
		case IsRune(token.Kind):
			v = processor.rune()
		case IsBoolean(token.Kind):
			v = processor.boolean()
		case IsNil(token.Kind):
			v = processor.nil()
		default:
			v = processor.numeric()
		}
	case lex.Name:
		v, ok = processor.name()
	default:
		p.PushErrorToken(token, "invalid_syntax")
	}
	return
}

type singleOperatorProcessor struct {
	token   lex.Token
	tokens  []lex.Token
	builder *expressionModelBuilder
	parser  *Parser
}

func (p *singleOperatorProcessor) unary() value {
	v := p.parser.processValuePart(p.tokens, p.builder)
	if !typeIsSingle(v.ast.Type) {
		p.parser.PushErrorToken(p.token, "invalid_data_unary")
	} else if !x.IsNumericType(v.ast.Type.Code) {
		p.parser.PushErrorToken(p.token, "invalid_data_unary")
	}
	return v
}

func (p *singleOperatorProcessor) plus() value {
	v := p.parser.processValuePart(p.tokens, p.builder)
	if !typeIsSingle(v.ast.Type) {
		p.parser.PushErrorToken(p.token, "invalid_data_plus")
	} else if !x.IsNumericType(v.ast.Type.Code) {
		p.parser.PushErrorToken(p.token, "invalid_data_plus")
	}
	return v
}

func (p *singleOperatorProcessor) tilde() value {
	v := p.parser.processValuePart(p.tokens, p.builder)
	if !typeIsSingle(v.ast.Type) {
		p.parser.PushErrorToken(p.token, "invalid_data_tilde")
	} else if !x.IsIntegerType(v.ast.Type.Code) {
		p.parser.PushErrorToken(p.token, "invalid_data_tilde")
	}
	return v
}

func (p *singleOperatorProcessor) logicalNot() value {
	v := p.parser.processValuePart(p.tokens, p.builder)
	if !typeIsSingle(v.ast.Type) {
		p.parser.PushErrorToken(p.token, "invalid_data_logical_not")
	} else if v.ast.Type.Code != x.Bool {
		p.parser.PushErrorToken(p.token, "invalid_data_logical_not")
	}
	return v
}

func (p *singleOperatorProcessor) star() value {
	v := p.parser.processValuePart(p.tokens, p.builder)
	if !typeIsPointer(v.ast.Type) {
		p.parser.PushErrorToken(p.token, "invalid_data_star")
	} else {
		v.ast.Type.Value = v.ast.Type.Value[1:]
	}
	return v
}

func (p *singleOperatorProcessor) amper() value {
	nodeLen := len(p.builder.current.nodes)
	v := p.parser.processValuePart(p.tokens, p.builder)
	if !canGetPointer(v) {
		p.parser.PushErrorToken(p.token, "invalid_data_amper")
	}
	if typeIsArray(v.ast.Type) {
		p.builder.current.nodes = append(
			p.builder.current.nodes[:nodeLen-1], /* -1 for remove amper operator */
			arrayPointerExp{p.builder.current.nodes[nodeLen:]})
	}
	v.ast.Type.Value = "*" + v.ast.Type.Value
	return v
}

func (p *Parser) processSingleOperatorPart(tokens []lex.Token, builder *expressionModelBuilder) value {
	var v value
	processor := singleOperatorProcessor{
		token: tokens[0],
		//? Length is 1 cause all length of operator tokens is 1.
		//? Change "1" with length of token's value
		//? if all operators length is not 1.
		tokens:  tokens[1:],
		builder: builder,
		parser:  p,
	}
	builder.appendNode(tokenExpNode{processor.token})
	if processor.tokens == nil {
		p.PushErrorToken(processor.token, "invalid_syntax")
		return v
	}
	switch processor.token.Kind {
	case "-":
		v = processor.unary()
	case "+":
		v = processor.plus()
	case "~":
		v = processor.tilde()
	case "!":
		v = processor.logicalNot()
	case "*":
		v = processor.star()
	case "&":
		v = processor.amper()
	default:
		p.PushErrorToken(processor.token, "invalid_syntax")
	}
	v.ast.Token = processor.token
	return v
}

func canGetPointer(v value) bool {
	if v.ast.Type.Code == x.Function {
		return false
	}
	return v.ast.Token.Id == lex.Name ||
		typeIsArray(v.ast.Type)
}

func (p *Parser) processValuePart(tokens []lex.Token, builder *expressionModelBuilder) (v value) {
	if tokens[0].Id == lex.Operator {
		return p.processSingleOperatorPart(tokens, builder)
	} else if len(tokens) == 1 {
		value, ok := p.processSingleValuePart(tokens[0], builder)
		if ok {
			v = value
			goto end
		}
	}
	switch token := tokens[len(tokens)-1]; token.Id {
	case lex.Brace:
		switch token.Kind {
		case ")":
			return p.processParenthesesValuePart(tokens, builder)
		case "}":
			return p.processBraceValuePart(tokens, builder)
		case "]":
			return p.processBracketValuePart(tokens, builder)
		}
	default:
		p.PushErrorToken(tokens[0], "invalid_syntax")
	}
end:
	return
}

func (p *Parser) processParenthesesValuePart(tokens []lex.Token, builder *expressionModelBuilder) (v value) {
	var valueTokens []lex.Token
	j := len(tokens) - 1
	braceCount := 0
	for ; j >= 0; j-- {
		token := tokens[j]
		if token.Id != lex.Brace {
			continue
		}
		switch token.Kind {
		case ")", "}", "]":
			braceCount++
		case "(", "{", "[":
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		valueTokens = tokens[:j]
		break
	}
	if len(valueTokens) == 0 && braceCount == 0 {
		// Write parentheses.
		builder.appendNode(tokenExpNode{lex.Token{Kind: "("}})
		defer builder.appendNode(tokenExpNode{lex.Token{Kind: ")"}})

		tk := tokens[0]
		tokens = tokens[1 : len(tokens)-1]
		if len(tokens) == 0 {
			p.PushErrorToken(tk, "invalid_syntax")
		}
		value, model := p.computeTokens(tokens)
		v = value
		builder.appendNode(model)
		return
	}
	v = p.processValuePart(valueTokens, builder)

	// Write parentheses.
	builder.appendNode(tokenExpNode{lex.Token{Kind: "("}})
	defer builder.appendNode(tokenExpNode{lex.Token{Kind: ")"}})

	switch v.ast.Type.Code {
	case x.Function:
		fun := v.ast.Type.Tag.(ast.FunctionAST)
		p.parseFunctionCallStatement(fun, tokens[len(valueTokens):], builder)
		v.ast.Type = fun.ReturnType
	default:
		p.PushErrorToken(tokens[len(valueTokens)], "invalid_syntax")
	}
	return
}

func (p *Parser) processBraceValuePart(tokens []lex.Token, builder *expressionModelBuilder) (v value) {
	var valueTokens []lex.Token
	j := len(tokens) - 1
	braceCount := 0
	for ; j >= 0; j-- {
		token := tokens[j]
		if token.Id != lex.Brace {
			continue
		}
		switch token.Kind {
		case "}", "]", ")":
			braceCount++
		case "{", "(", "[":
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		valueTokens = tokens[:j]
		break
	}
	valTokensLen := len(valueTokens)
	if valTokensLen == 0 || braceCount > 0 {
		p.PushErrorToken(tokens[0], "invalid_syntax")
		return
	}
	switch valueTokens[0].Id {
	case lex.Brace:
		switch valueTokens[0].Kind {
		case "[":
			ast := ast.New(nil)
			dt, ok := ast.BuildDataType(valueTokens, new(int), true)
			if !ok {
				p.AppendErrors(ast.Errors...)
				return
			}
			valueTokens = tokens[len(valueTokens):]
			var model expressionNode
			v, model = p.buildArray(p.buildEnumerableParts(valueTokens), dt, valueTokens[0])
			builder.appendNode(model)
			return
		case "(":
			astBuilder := ast.New(tokens)
			funAST := astBuilder.BuildFunction(true)
			if len(astBuilder.Errors) > 0 {
				p.AppendErrors(astBuilder.Errors...)
				return
			}
			p.checkAnonymousFunction(funAST)
			v.ast.Type.Tag = funAST
			v.ast.Type.Code = x.Function
			builder.appendNode(anonymousFunctionExp{funAST})
			return
		default:
			p.PushErrorToken(valueTokens[0], "invalid_syntax")
		}
	default:
		p.PushErrorToken(valueTokens[0], "invalid_syntax")
	}
	return
}

func (p *Parser) processBracketValuePart(tokens []lex.Token, builder *expressionModelBuilder) (v value) {
	var valueTokens []lex.Token
	j := len(tokens) - 1
	braceCount := 0
	for ; j >= 0; j-- {
		token := tokens[j]
		if token.Id != lex.Brace {
			continue
		}
		switch token.Kind {
		case "}", "]", ")":
			braceCount++
		case "{", "(", "[":
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		valueTokens = tokens[:j]
		break
	}
	valTokensLen := len(valueTokens)
	if valTokensLen == 0 || braceCount > 0 {
		p.PushErrorToken(tokens[0], "invalid_syntax")
		return
	}
	var model expressionNode
	v, model = p.computeTokens(valueTokens)
	builder.appendNode(model)
	tokens = tokens[len(valueTokens)+1 : len(tokens)-1] // Removed array syntax "["..."]"
	builder.appendNode(tokenExpNode{lex.Token{Kind: "["}})
	selectv, model := p.computeTokens(tokens)
	builder.appendNode(model)
	builder.appendNode(tokenExpNode{lex.Token{Kind: "]"}})
	return p.processEnumerableSelect(v, selectv, tokens[0])
}

func (p *Parser) processEnumerableSelect(enumv, selectv value, err lex.Token) (v value) {
	switch {
	case typeIsArray(enumv.ast.Type):
		return p.processArraySelect(enumv, selectv, err)
	case typeIsSingle(enumv.ast.Type):
		return p.processStringSelect(enumv, selectv, err)
	}
	p.PushErrorToken(err, "not_enumerable")
	return
}

func (p *Parser) processArraySelect(arrv, selectv value, err lex.Token) value {
	arrv.ast.Type = typeOfArrayElements(arrv.ast.Type)
	if !typeIsSingle(selectv.ast.Type) || !x.IsIntegerType(selectv.ast.Type.Code) {
		p.PushErrorToken(err, "notint_array_select")
	}
	return arrv
}

func (p *Parser) processStringSelect(strv, selectv value, err lex.Token) value {
	strv.ast.Type.Code = x.Rune
	if !typeIsSingle(selectv.ast.Type) || !x.IsIntegerType(selectv.ast.Type.Code) {
		p.PushErrorToken(err, "notint_string_select")
	}
	return strv
}

type enumPart struct {
	tokens []lex.Token
}

//! IMPORTANT: Tokens is should be store enumerable parentheses.
func (p *Parser) buildEnumerableParts(tokens []lex.Token) []enumPart {
	tokens = tokens[1 : len(tokens)-1]
	braceCount := 0
	lastComma := -1
	var parts []enumPart
	for index, token := range tokens {
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
		if token.Id == lex.Comma {
			if index-lastComma-1 == 0 {
				p.PushErrorToken(token, "missing_expression")
				lastComma = index
				continue
			}
			parts = append(parts, enumPart{tokens[lastComma+1 : index]})
			lastComma = index
		}
	}
	if lastComma+1 < len(tokens) {
		parts = append(parts, enumPart{tokens[lastComma+1:]})
	}
	return parts
}

func (p *Parser) buildArray(parts []enumPart, dt ast.DataTypeAST, err lex.Token) (value, expressionNode) {
	var v value
	v.ast.Type = dt
	model := arrayExp{dataType: dt}
	elementType := typeOfArrayElements(dt)
	for _, part := range parts {
		partValue, expModel := p.computeTokens(part.tokens)
		model.expressions = append(model.expressions, expModel)
		p.checkType(elementType, partValue.ast.Type, false, part.tokens[0])
	}
	return v, model
}

func (p *Parser) checkAnonymousFunction(fun ast.FunctionAST) {
	globalVariables := p.GlobalVariables
	blockVariables := p.BlockVariables
	p.GlobalVariables = append(blockVariables, p.GlobalVariables...)
	p.BlockVariables = variablesFromParameters(fun.Params)
	p.checkFunction(fun)
	p.GlobalVariables = globalVariables
	p.BlockVariables = blockVariables
}

func (p *Parser) parseFunctionCallStatement(fun ast.FunctionAST, tokens []lex.Token, builder *expressionModelBuilder) {
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
	p.parseArgs(fun, args, errToken, builder)
	if builder != nil {
		builder.appendNode(argsExp{args})
	}
}

func (p *Parser) parseArgs(fun ast.FunctionAST, args []ast.ArgAST, errToken lex.Token, builder *expressionModelBuilder) {
	if len(args) < len(fun.Params) {
		p.PushErrorToken(errToken, "missing_argument")
	}
	for index, arg := range args {
		p.parseArg(fun, index, &arg)
		args[index] = arg
	}
}

func (p *Parser) parseArg(fun ast.FunctionAST, index int, arg *ast.ArgAST) {
	if index >= len(fun.Params) {
		p.PushErrorToken(arg.Token, "argument_overflow")
		return
	}
	value, model := p.computeExpression(arg.Expression)
	arg.Expression.Model = model
	param := fun.Params[index]
	p.checkType(param.Type, value.ast.Type, false, arg.Token)
}

func (p *Parser) getRangeTokens(open, close string, tokens []lex.Token) []lex.Token {
	braceCount := 0
	start := 1
	for index, token := range tokens {
		if token.Id != lex.Brace {
			continue
		}
		if token.Kind == open {
			braceCount++
		} else if token.Kind == close {
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

func (p *Parser) checkFunctionSpecialCases(fun *function) {
	switch fun.ast.Name {
	case "_" + x.EntryPoint:
		if len(fun.ast.Params) > 0 {
			p.PushErrorToken(fun.ast.Token, "entrypoint_have_parameters")
		}
		if fun.ast.ReturnType.Code != x.Void {
			p.PushErrorToken(fun.ast.ReturnType.Token, "entrypoint_have_return")
		}
		if fun.attributes != nil {
			p.PushErrorToken(fun.ast.Token, "entrypoint_have_attributes")
		}
	}
}

func (p *Parser) checkBlock(b ast.BlockAST) {
	for index, model := range b.Statements {
		switch t := model.Value.(type) {
		case ast.BlockExpressionAST:
			_, t.Expression.Model = p.computeExpression(t.Expression)
			model.Value = t
			b.Statements[index] = model
		case ast.VariableAST:
			p.checkVariableStatement(&t)
			model.Value = t
			b.Statements[index] = model
		case ast.VariableSetAST:
			p.checkVariableSetStatement(&t)
			model.Value = t
			b.Statements[index] = model
		case ast.ReturnAST:
		default:
			p.PushErrorToken(model.Token, "invalid_syntax")
		}
	}
}

func (p *Parser) checkParameters(params []ast.ParameterAST) {
	for _, param := range params {
		if !param.Const {
			continue
		}
		if !checkValidityConstantDataType(param.Type) {
			p.PushErrorToken(param.Type.Token, "invalid_const_data_type")
		}
	}
}

func (p *Parser) checkFunction(fun ast.FunctionAST) {
	p.checkBlock(fun.Block)
	p.checkFunctionReturn(fun)
	p.checkParameters(fun.Params)
}

func (p *Parser) checkVariableStatement(varAST *ast.VariableAST) {
	for _, t := range p.Types {
		if varAST.Name == t.Name {
			p.PushErrorToken(varAST.NameToken, "exist_name")
			break
		}
	}
	for _, variable := range p.BlockVariables {
		if varAST.Name == variable.Name {
			p.PushErrorToken(varAST.NameToken, "exist_name")
			break
		}
	}
	*varAST = p.ParseVariable(*varAST)
	p.BlockVariables = append(p.BlockVariables, *varAST)
}

func (p *Parser) checkVariableSetStatement(vsAST *ast.VariableSetAST) {
	selected, _ := p.computeProcesses(vsAST.SelectExpression.Processes)
	if selected.constant {
		p.PushErrorToken(vsAST.Setter, "const_value_update")
		return
	}
	switch selected.ast.Type.Tag.(type) {
	case ast.FunctionAST:
		if p.functionByName(selected.ast.Token.Kind) != nil {
			p.PushErrorToken(vsAST.Setter, "type_not_support_value_update")
			return
		}
	}
	value, model := p.computeProcesses(vsAST.ValueExpression.Processes)
	vsAST.ValueExpression = model.ExpressionAST()
	if vsAST.Setter.Kind != "=" {
		vsAST.Setter.Kind = vsAST.Setter.Kind[:len(vsAST.Setter.Kind)-1]
		value.ast = arithmeticProcess{
			p:        p,
			left:     vsAST.SelectExpression.Tokens,
			leftVal:  selected.ast,
			right:    vsAST.ValueExpression.Tokens,
			rightVal: value.ast,
			operator: vsAST.Setter,
		}.solve()
		vsAST.Setter.Kind += "="
	}
	p.checkType(selected.ast.Type, value.ast.Type, false, vsAST.Setter)
}
