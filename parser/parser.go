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
	waitingGlobalVariables []ast.VariableAST
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
	astModel := ast.NewBuilder(p.Tokens)
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
			p.Statement(t)
		case ast.TypeAST:
			p.Type(t)
		default:
			p.PushErrorToken(model.Token, "invalid_syntax")
		}
	}
	p.check()
}

// Type parses X type define statement.
func (p *Parser) Type(t ast.TypeAST) {
	if p.existName(t.Name).Id != lex.NA {
		p.PushErrorToken(t.Token, "exist_name")
		return
	} else if x.IsIgnoreName(t.Name) {
		p.PushErrorToken(t.Token, "ignore_name_identifier")
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

// Statement parse X statement.
func (p *Parser) Statement(s ast.StatementAST) {
	switch t := s.Value.(type) {
	case ast.FunctionAST:
		p.Function(t)
	case ast.VariableAST:
		p.GlobalVariable(t)
	default:
		p.PushErrorToken(s.Token, "invalid_syntax")
	}
}

// Function parse X function.
func (p *Parser) Function(funAST ast.FunctionAST) {
	if p.existName(funAST.Name).Id != lex.NA {
		p.PushErrorToken(funAST.Token, "exist_name")
	} else if x.IsIgnoreName(funAST.Name) {
		p.PushErrorToken(funAST.Token, "ignore_name_identifier")
	}
	fun := new(function)
	fun.Ast = funAST
	fun.Attributes = p.attributes
	p.attributes = nil
	p.checkFunctionAttributes(fun.Attributes)
	p.Functions = append(p.Functions, fun)
}

// ParseVariable parse X global variable.
func (p *Parser) GlobalVariable(varAST ast.VariableAST) {
	if p.existName(varAST.Name).Id != lex.NA {
		p.PushErrorToken(varAST.NameToken, "exist_name")
		return
	}
	p.waitingGlobalVariables = append(p.waitingGlobalVariables, varAST)
}

// WaitingGlobalVariables parse X global variables for waiting parsing.
func (p *Parser) WaitingGlobalVariables() {
	for _, varAST := range p.waitingGlobalVariables {
		p.GlobalVariables = append(p.GlobalVariables, p.Variable(varAST))
	}
}

// Variable parse X variable.
func (p *Parser) Variable(varAST ast.VariableAST) ast.VariableAST {
	if x.IsIgnoreName(varAST.Name) {
		p.PushErrorToken(varAST.NameToken, "ignore_name_identifier")
	}
	var dt ast.DataTypeAST
	switch t := varAST.Tag.(type) {
	case ast.DataTypeAST:
		dt = t
	default:
		if varAST.SetterToken.Id != lex.NA {
			var val value
			val, varAST.Value.Model = p.computeExpr(varAST.Value)
			dt = val.ast.Type
		}
	}
	if varAST.Type.Code != x.Void {
		if varAST.SetterToken.Id != lex.NA { // Pass default value.
			p.checkType(varAST.Type, dt, false, varAST.NameToken)
		} else {
			var valueToken lex.Token
			valueToken.Id = lex.Value
			dt, ok := p.readyType(varAST.Type)
			if ok {
				valueToken.Kind = p.defaultValueOfType(dt)
				valueTokens := []lex.Token{valueToken}
				processes := [][]lex.Token{valueTokens}
				varAST.Value = ast.ExprAST{
					Tokens:    valueTokens,
					Processes: processes,
				}
			}
		}
	} else {
		if varAST.SetterToken.Id == lex.NA {
			p.PushErrorToken(varAST.NameToken, "missing_autotype_value")
		} else {
			varAST.Type = dt
			p.checkValidityForAutoType(varAST.Type, varAST.SetterToken)
		}
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

func (p *Parser) typeByName(name string) *ast.TypeAST {
	for _, t := range p.Types {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

// FunctionByName returns function by specified name.
//
// Special case:
//  FunctionByName(name) -> nil: if function is not exist.
func (p *Parser) FunctionByName(name string) *function {
	for _, fun := range builtinFunctions {
		if fun.Ast.Name == name {
			return fun
		}
	}
	for _, fun := range p.Functions {
		if fun.Ast.Name == name {
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
	fun := p.FunctionByName(name)
	if fun != nil {
		return fun.Ast.Token
	}
	variable := p.variableByName(name)
	if variable != nil {
		return variable.NameToken
	}
	for _, varAST := range p.waitingGlobalVariables {
		if varAST.Name == name {
			return varAST.NameToken
		}
	}
	return lex.Token{}
}

func (p *Parser) check() {
	if p.FunctionByName("_"+x.EntryPoint) == nil {
		p.PushError("no_entry_point")
	}
	p.checkTypes()
	p.WaitingGlobalVariables()
	p.waitingGlobalVariables = nil
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
		p.BlockVariables = variablesFromParameters(fun.Ast.Params)
		p.checkFunctionSpecialCases(fun)
		p.checkFunction(&fun.Ast)
	}
}

type value struct {
	ast      ast.ValueAST
	constant bool
}

func (p *Parser) computeProcesses(processes [][]lex.Token) (v value, e exprModel) {
	if processes == nil {
		return
	}
	builder := newExprBuilder()
	if len(processes) == 1 {
		builder.setIndex(0)
		v = p.computeValPart(processes[0], builder)
		e = builder.build()
		return
	}
	process := solver{p: p}
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
			builder.appendNode(tokenExprNode{process.operator})
			process.right = processes[j+1]
			builder.setIndex(j + 1)
			process.rightVal = p.computeValPart(process.right, builder).ast
			v.ast = process.Solve()
			processes = processes[2:]
			goto end
		} else if j == len(processes)-1 {
			process.operator = processes[j][0]
			process.left = processes[j-1]
			builder.setIndex(j - 1)
			process.leftVal = p.computeValPart(process.left, builder).ast
			process.rightVal = v.ast
			builder.setIndex(j)
			builder.appendNode(tokenExprNode{process.operator})
			v.ast = process.Solve()
			processes = processes[:j-1]
			goto end
		} else if prev := processes[j-1]; prev[0].Id == lex.Operator &&
			len(prev) == 1 {
			process.leftVal = v.ast
			process.operator = processes[j][0]
			builder.setIndex(j)
			builder.appendNode(tokenExprNode{process.operator})
			process.right = processes[j+1]
			builder.setIndex(j + 1)
			process.rightVal = p.computeValPart(process.right, builder).ast
			v.ast = process.Solve()
			processes = append(processes[:j], processes[j+2:]...)
			goto end
		}
		process.left = processes[j-1]
		builder.setIndex(j - 1)
		process.leftVal = p.computeValPart(process.left, builder).ast
		process.operator = processes[j][0]
		builder.setIndex(j)
		builder.appendNode(tokenExprNode{process.operator})
		process.right = processes[j+1]
		builder.setIndex(j + 1)
		process.rightVal = p.computeValPart(process.right, builder).ast
		{
			solvedValue := process.Solve()
			if v.ast.Type.Code != x.Void {
				process.operator.Kind = "+"
				process.leftVal = v.ast
				process.right = processes[j+1]
				process.rightVal = solvedValue
				v.ast = process.Solve()
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

func (p *Parser) computeTokens(tokens []lex.Token) (value, exprModel) {
	return p.computeProcesses(new(ast.Builder).Expr(tokens).Processes)
}

func (p *Parser) computeExpr(ex ast.ExprAST) (value, exprModel) {
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

type valueProcessor struct {
	token   lex.Token
	builder *exprModelBuilder
	parser  *Parser
}

func (p *valueProcessor) string() value {
	var v value
	v.ast.Value = p.token.Kind
	v.ast.Type.Code = x.Str
	v.ast.Type.Value = "str"
	p.builder.appendNode(strExprNode{p.token})
	return v
}

func (p *valueProcessor) rune() value {
	var v value
	v.ast.Value = p.token.Kind
	v.ast.Type.Code = x.Rune
	v.ast.Type.Value = "rune"
	p.builder.appendNode(runeExprNode{p.token})
	return v
}

func (p *valueProcessor) boolean() value {
	var v value
	v.ast.Value = p.token.Kind
	v.ast.Type.Code = x.Bool
	v.ast.Type.Value = "bool"
	p.builder.appendNode(tokenExprNode{p.token})
	return v
}

func (p *valueProcessor) nil() value {
	var v value
	v.ast.Value = p.token.Kind
	v.ast.Type.Code = x.Nil
	p.builder.appendNode(tokenExprNode{p.token})
	return v
}

func (p *valueProcessor) numeric() value {
	var v value
	if strings.Contains(p.token.Kind, ".") ||
		strings.ContainsAny(p.token.Kind, "eE") {
		v.ast.Type.Code = x.F64
		v.ast.Type.Value = "f64"
	} else {
		v.ast.Type.Code = x.I32
		v.ast.Type.Value = "i32"
		ok := xbits.CheckBitInt(p.token.Kind, 32)
		if !ok {
			v.ast.Type.Code = x.I64
			v.ast.Type.Value = "i64"
		}
	}
	v.ast.Value = p.token.Kind
	p.builder.appendNode(tokenExprNode{p.token})
	return v
}

func (p *valueProcessor) name() (v value, ok bool) {
	if variable := p.parser.variableByName(p.token.Kind); variable != nil {
		v.ast.Value = p.token.Kind
		v.ast.Type = variable.Type
		v.constant = variable.DefineToken.Id == lex.Const
		v.ast.Token = variable.NameToken
		p.builder.appendNode(tokenExprNode{p.token})
		ok = true
	} else if fun := p.parser.FunctionByName(p.token.Kind); fun != nil {
		v.ast.Value = p.token.Kind
		v.ast.Type.Code = x.Function
		v.ast.Type.Tag = fun.Ast
		v.ast.Type.Value = fun.Ast.DataTypeString()
		v.ast.Token = fun.Ast.Token
		p.builder.appendNode(tokenExprNode{p.token})
		ok = true
	} else {
		p.parser.PushErrorToken(p.token, "name_not_defined")
	}
	return
}

type solver struct {
	p        *Parser
	left     []lex.Token
	leftVal  ast.ValueAST
	right    []lex.Token
	rightVal ast.ValueAST
	operator lex.Token
}

func (s solver) pointer() (v ast.ValueAST) {
	if s.leftVal.Type.Value != s.rightVal.Type.Value {
		s.p.PushErrorToken(s.operator, "incompatible_type")
		return
	}
	switch s.operator.Kind {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		s.p.PushErrorToken(s.operator, "operator_notfor_pointer")
	}
	return
}

func (s solver) string() (v ast.ValueAST) {
	// Not both string?
	if s.leftVal.Type.Code != s.rightVal.Type.Code {
		s.p.PushErrorToken(s.operator, "incompatible_datatype")
		return
	}
	switch s.operator.Kind {
	case "+":
		v.Type.Code = x.Str
	case "==", "!=":
		v.Type.Code = x.Bool
	default:
		s.p.PushErrorToken(s.operator, "operator_notfor_string")
	}
	return
}

func (s solver) any() (v ast.ValueAST) {
	switch s.operator.Kind {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		s.p.PushErrorToken(s.operator, "operator_notfor_any")
	}
	return
}

func (s solver) bool() (v ast.ValueAST) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		s.p.PushErrorToken(s.operator, "incompatible_type")
		return
	}
	switch s.operator.Kind {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		s.p.PushErrorToken(s.operator, "operator_notfor_bool")
	}
	return
}

func (s solver) float() (v ast.ValueAST) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		if !isConstantNumeric(s.leftVal.Value) &&
			!isConstantNumeric(s.rightVal.Value) {
			s.p.PushErrorToken(s.operator, "incompatible_type")
			return
		}
	}
	switch s.operator.Kind {
	case "!=", "==", "<", ">", ">=", "<=":
		v.Type.Code = x.Bool
	case "+", "-", "*", "/":
		v.Type.Code = x.F32
		if s.leftVal.Type.Code == x.F64 || s.rightVal.Type.Code == x.F64 {
			v.Type.Code = x.F64
		}
	default:
		s.p.PushErrorToken(s.operator, "operator_notfor_float")
	}
	return
}

func (s solver) signed() (v ast.ValueAST) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		if !isConstantNumeric(s.leftVal.Value) &&
			!isConstantNumeric(s.rightVal.Value) {
			s.p.PushErrorToken(s.operator, "incompatible_type")
			return
		}
	}
	switch s.operator.Kind {
	case "!=", "==", "<", ">", ">=", "<=":
		v.Type.Code = x.Bool
	case "+", "-", "*", "/", "%", "&", "|", "^":
		v.Type = s.leftVal.Type
		if x.TypeGreaterThan(s.rightVal.Type.Code, v.Type.Code) {
			v.Type = s.rightVal.Type
		}
	case ">>", "<<":
		v.Type = s.leftVal.Type
		if !x.IsUnsignedNumericType(s.rightVal.Type.Code) &&
			!checkIntBit(s.rightVal, xbits.BitsizeOfType(x.U64)) {
			s.p.PushErrorToken(s.rightVal.Token, "bitshift_must_unsigned")
		}
	default:
		s.p.PushErrorToken(s.operator, "operator_notfor_int")
	}
	return
}

func (s solver) unsigned() (v ast.ValueAST) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		if !isConstantNumeric(s.leftVal.Value) &&
			!isConstantNumeric(s.rightVal.Value) {
			s.p.PushErrorToken(s.operator, "incompatible_type")
			return
		}
		return
	}
	switch s.operator.Kind {
	case "!=", "==", "<", ">", ">=", "<=":
		v.Type.Code = x.Bool
	case "+", "-", "*", "/", "%", "&", "|", "^":
		v.Type = s.leftVal.Type
		if x.TypeGreaterThan(s.rightVal.Type.Code, v.Type.Code) {
			v.Type = s.rightVal.Type
		}
	default:
		s.p.PushErrorToken(s.operator, "operator_notfor_uint")
	}
	return
}

func (s solver) logical() (v ast.ValueAST) {
	v.Type.Code = x.Bool
	if s.leftVal.Type.Code != x.Bool {
		s.p.PushErrorToken(s.leftVal.Token, "logical_not_bool")
	}
	if s.rightVal.Type.Code != x.Bool {
		s.p.PushErrorToken(s.rightVal.Token, "logical_not_bool")
	}
	return
}

func (s solver) rune() (v ast.ValueAST) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		s.p.PushErrorToken(s.operator, "incompatible_type")
		return
	}
	switch s.operator.Kind {
	case "!=", "==", ">", "<", ">=", "<=":
		v.Type.Code = x.Bool
	case "+", "-", "*", "/", "^", "&", "%", "|":
		v.Type.Code = x.Rune
	default:
		s.p.PushErrorToken(s.operator, "operator_notfor_rune")
	}
	return
}

func (s solver) array() (v ast.ValueAST) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		s.p.PushErrorToken(s.operator, "incompatible_type")
		return
	}
	switch s.operator.Kind {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		s.p.PushErrorToken(s.operator, "operator_notfor_array")
	}
	return
}

func (s solver) nil() (v ast.ValueAST) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, false) {
		s.p.PushErrorToken(s.operator, "incompatible_type")
		return
	}
	switch s.operator.Kind {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		s.p.PushErrorToken(s.operator, "operator_notfor_nil")
	}
	return
}

func (s solver) Solve() (v ast.ValueAST) {
	switch s.operator.Kind {
	case "+", "-", "*", "/", "%", ">>",
		"<<", "&", "|", "^", "==", "!=",
		">=", "<=", ">", "<":
	case "&&", "||":
		return s.logical()
	default:
		s.p.PushErrorToken(s.operator, "invalid_operator")
	}
	switch {
	case typeIsArray(s.leftVal.Type) || typeIsArray(s.rightVal.Type):
		return s.array()
	case typeIsPointer(s.leftVal.Type) || typeIsPointer(s.rightVal.Type):
		return s.pointer()
	case s.leftVal.Type.Code == x.Nil || s.rightVal.Type.Code == x.Nil:
		return s.nil()
	case s.leftVal.Type.Code == x.Rune || s.rightVal.Type.Code == x.Rune:
		return s.rune()
	case s.leftVal.Type.Code == x.Any || s.rightVal.Type.Code == x.Any:
		return s.any()
	case s.leftVal.Type.Code == x.Bool || s.rightVal.Type.Code == x.Bool:
		return s.bool()
	case s.leftVal.Type.Code == x.Str || s.rightVal.Type.Code == x.Str:
		return s.string()
	case x.IsFloatType(s.leftVal.Type.Code) ||
		x.IsFloatType(s.rightVal.Type.Code):
		return s.float()
	case x.IsSignedNumericType(s.leftVal.Type.Code) ||
		x.IsSignedNumericType(s.rightVal.Type.Code):
		return s.signed()
	case x.IsUnsignedNumericType(s.leftVal.Type.Code) ||
		x.IsUnsignedNumericType(s.rightVal.Type.Code):
		return s.unsigned()
	}
	return
}

func (p *Parser) computeVal(token lex.Token, builder *exprModelBuilder) (v value, ok bool) {
	processor := valueProcessor{token, builder, p}
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

type operatorProcessor struct {
	token   lex.Token
	tokens  []lex.Token
	builder *exprModelBuilder
	parser  *Parser
}

func (p *operatorProcessor) unary() value {
	v := p.parser.computeValPart(p.tokens, p.builder)
	if !typeIsSingle(v.ast.Type) {
		p.parser.PushErrorToken(p.token, "invalid_data_unary")
	} else if !x.IsNumericType(v.ast.Type.Code) {
		p.parser.PushErrorToken(p.token, "invalid_data_unary")
	}
	return v
}

func (p *operatorProcessor) plus() value {
	v := p.parser.computeValPart(p.tokens, p.builder)
	if !typeIsSingle(v.ast.Type) {
		p.parser.PushErrorToken(p.token, "invalid_data_plus")
	} else if !x.IsNumericType(v.ast.Type.Code) {
		p.parser.PushErrorToken(p.token, "invalid_data_plus")
	}
	return v
}

func (p *operatorProcessor) tilde() value {
	v := p.parser.computeValPart(p.tokens, p.builder)
	if !typeIsSingle(v.ast.Type) {
		p.parser.PushErrorToken(p.token, "invalid_data_tilde")
	} else if !x.IsIntegerType(v.ast.Type.Code) {
		p.parser.PushErrorToken(p.token, "invalid_data_tilde")
	}
	return v
}

func (p *operatorProcessor) logicalNot() value {
	v := p.parser.computeValPart(p.tokens, p.builder)
	if !typeIsSingle(v.ast.Type) {
		p.parser.PushErrorToken(p.token, "invalid_data_logical_not")
	} else if v.ast.Type.Code != x.Bool {
		p.parser.PushErrorToken(p.token, "invalid_data_logical_not")
	}
	return v
}

func (p *operatorProcessor) star() value {
	v := p.parser.computeValPart(p.tokens, p.builder)
	if !typeIsPointer(v.ast.Type) {
		p.parser.PushErrorToken(p.token, "invalid_data_star")
	} else {
		v.ast.Type.Value = v.ast.Type.Value[1:]
	}
	return v
}

func (p *operatorProcessor) amper() value {
	v := p.parser.computeValPart(p.tokens, p.builder)
	if !canGetPointer(v) {
		p.parser.PushErrorToken(p.token, "invalid_data_amper")
	}
	v.ast.Type.Value = "*" + v.ast.Type.Value
	return v
}

func (p *Parser) computeOperatorPart(tokens []lex.Token, builder *exprModelBuilder) value {
	var v value
	//? Length is 1 cause all length of operator tokens is 1.
	//? Change "1" with length of token's value
	//? if all operators length is not 1.
	exprTokens := tokens[1:]
	processor := operatorProcessor{tokens[0], exprTokens, builder, p}
	builder.appendNode(tokenExprNode{processor.token})
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
	return v.ast.Token.Id == lex.Name
}

func (p *Parser) computeHeapAlloc(tokens []lex.Token, builder *exprModelBuilder) (v value) {
	if len(tokens) == 1 {
		p.PushErrorToken(tokens[0], "invalid_syntax_keyword_new")
		return
	}
	v.ast.Token = tokens[0]
	tokens = tokens[1:]
	astb := new(ast.Builder)
	index := new(int)
	dt, ok := astb.DataType(tokens, index, true)
	builder.appendNode(newHeapAllocationExprModel{dt})
	dt.Value = "*" + dt.Value
	v.ast.Type = dt
	if !ok {
		p.PushErrorToken(tokens[0], "fail_build_heap_allocation_type")
		return
	}
	if *index < len(tokens)-1 {
		p.PushErrorToken(tokens[*index+1], "invalid_syntax")
	}
	return
}

func (p *Parser) computeValPart(tokens []lex.Token, builder *exprModelBuilder) (v value) {
	if len(tokens) == 1 {
		value, ok := p.computeVal(tokens[0], builder)
		if ok {
			v = value
			return
		}
	}
	firstTok := tokens[0]
	switch firstTok.Id {
	case lex.Operator:
		return p.computeOperatorPart(tokens, builder)
	case lex.New:
		return p.computeHeapAlloc(tokens, builder)
	}
	switch token := tokens[len(tokens)-1]; token.Id {
	case lex.Brace:
		switch token.Kind {
		case ")":
			return p.computeParenthesesRange(tokens, builder)
		case "}":
			return p.computeBraceRange(tokens, builder)
		case "]":
			return p.computeBracketRange(tokens, builder)
		}
	default:
		p.PushErrorToken(tokens[0], "invalid_syntax")
	}
	return
}

func (p *Parser) computeParenthesesRange(tokens []lex.Token, builder *exprModelBuilder) (v value) {
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
		builder.appendNode(tokenExprNode{lex.Token{Kind: "("}})
		defer builder.appendNode(tokenExprNode{lex.Token{Kind: ")"}})

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
	v = p.computeValPart(valueTokens, builder)

	// Write parentheses.
	builder.appendNode(tokenExprNode{lex.Token{Kind: "("}})
	defer builder.appendNode(tokenExprNode{lex.Token{Kind: ")"}})

	switch v.ast.Type.Code {
	case x.Function:
		fun := v.ast.Type.Tag.(ast.FunctionAST)
		p.parseFunctionCall(fun, tokens[len(valueTokens):], builder)
		v.ast.Type = fun.ReturnType
	default:
		p.PushErrorToken(tokens[len(valueTokens)], "invalid_syntax")
	}
	return
}

func (p *Parser) computeBraceRange(tokens []lex.Token, builder *exprModelBuilder) (v value) {
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
			ast := ast.NewBuilder(nil)
			dt, ok := ast.DataType(valueTokens, new(int), true)
			if !ok {
				p.AppendErrors(ast.Errors...)
				return
			}
			valueTokens = tokens[len(valueTokens):]
			var model exprNode
			v, model = p.buildArray(p.buildEnumerableParts(valueTokens), dt, valueTokens[0])
			builder.appendNode(model)
			return
		case "(":
			astBuilder := ast.NewBuilder(tokens)
			funAST := astBuilder.Function(true)
			if len(astBuilder.Errors) > 0 {
				p.AppendErrors(astBuilder.Errors...)
				return
			}
			p.checkAnonymousFunction(&funAST)
			v.ast.Type.Tag = funAST
			v.ast.Type.Code = x.Function
			v.ast.Type.Value = funAST.DataTypeString()
			builder.appendNode(anonymousFunctionExpr{funAST})
			return
		default:
			p.PushErrorToken(valueTokens[0], "invalid_syntax")
		}
	default:
		p.PushErrorToken(valueTokens[0], "invalid_syntax")
	}
	return
}

func (p *Parser) computeBracketRange(tokens []lex.Token, builder *exprModelBuilder) (v value) {
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
	var model exprNode
	v, model = p.computeTokens(valueTokens)
	builder.appendNode(model)
	tokens = tokens[len(valueTokens)+1 : len(tokens)-1] // Removed array syntax "["..."]"
	builder.appendNode(tokenExprNode{lex.Token{Kind: "["}})
	selectv, model := p.computeTokens(tokens)
	builder.appendNode(model)
	builder.appendNode(tokenExprNode{lex.Token{Kind: "]"}})
	return p.computeEnumerableSelect(v, selectv, tokens[0])
}

func (p *Parser) computeEnumerableSelect(enumv, selectv value, err lex.Token) (v value) {
	switch {
	case typeIsArray(enumv.ast.Type):
		return p.computeArraySelect(enumv, selectv, err)
	case typeIsSingle(enumv.ast.Type):
		return p.computeStringSelect(enumv, selectv, err)
	}
	p.PushErrorToken(err, "not_enumerable")
	return
}

func (p *Parser) computeArraySelect(arrv, selectv value, err lex.Token) value {
	arrv.ast.Type = typeOfArrayElements(arrv.ast.Type)
	if !typeIsSingle(selectv.ast.Type) || !x.IsIntegerType(selectv.ast.Type.Code) {
		p.PushErrorToken(err, "notint_array_select")
	}
	return arrv
}

func (p *Parser) computeStringSelect(strv, selectv value, err lex.Token) value {
	strv.ast.Type.Code = x.Rune
	if !typeIsSingle(selectv.ast.Type) || !x.IsIntegerType(selectv.ast.Type.Code) {
		p.PushErrorToken(err, "notint_string_select")
	}
	return strv
}

//! IMPORTANT: Tokens is should be store enumerable parentheses.
func (p *Parser) buildEnumerableParts(tokens []lex.Token) [][]lex.Token {
	tokens = tokens[1 : len(tokens)-1]
	braceCount := 0
	lastComma := -1
	var parts [][]lex.Token
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
			parts = append(parts, tokens[lastComma+1:index])
			lastComma = index
		}
	}
	if lastComma+1 < len(tokens) {
		parts = append(parts, tokens[lastComma+1:])
	}
	return parts
}

func (p *Parser) buildArray(parts [][]lex.Token, dt ast.DataTypeAST, err lex.Token) (value, exprNode) {
	var v value
	v.ast.Type = dt
	model := arrayExpr{dataType: dt}
	elementType := typeOfArrayElements(dt)
	for _, part := range parts {
		partValue, expModel := p.computeTokens(part)
		model.expr = append(model.expr, expModel)
		p.checkType(elementType, partValue.ast.Type, false, part[0])
	}
	return v, model
}

func (p *Parser) checkAnonymousFunction(fun *ast.FunctionAST) {
	globalVariables := p.GlobalVariables
	blockVariables := p.BlockVariables
	p.GlobalVariables = append(blockVariables, p.GlobalVariables...)
	p.BlockVariables = variablesFromParameters(fun.Params)
	p.checkFunction(fun)
	p.GlobalVariables = globalVariables
	p.BlockVariables = blockVariables
}

func (p *Parser) parseFunctionCall(fun ast.FunctionAST, tokens []lex.Token, builder *exprModelBuilder) {
	errToken := tokens[0]
	tokens, _ = p.getRange("(", ")", tokens)
	if tokens == nil {
		tokens = make([]lex.Token, 0)
	}
	ast := new(ast.Builder)
	args := ast.Args(tokens)
	if len(ast.Errors) > 0 {
		p.AppendErrors(ast.Errors...)
	}
	p.parseArgs(fun, args, errToken, builder)
	if builder != nil {
		builder.appendNode(argsExpr{args})
	}
}

func (p *Parser) parseArgs(fun ast.FunctionAST, args []ast.ArgAST, errToken lex.Token, builder *exprModelBuilder) {
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
	value, model := p.computeExpr(arg.Expr)
	arg.Expr.Model = model
	param := fun.Params[index]
	p.checkType(param.Type, value.ast.Type, false, arg.Token)
}

// Returns between of brackets.
//
// Special case is:
//  getRange(open, close, tokens) = nil, false if first token is not brace.
func (p *Parser) getRange(open, close string, tokens []lex.Token) (_ []lex.Token, ok bool) {
	braceCount := 0
	start := 1
	if tokens[0].Id != lex.Brace {
		return nil, false
	}
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
		return tokens[start:index], true
	}
	return nil, false
}

func (p *Parser) checkFunctionSpecialCases(fun *function) {
	switch fun.Ast.Name {
	case "_" + x.EntryPoint:
		p.checkEntryPointSpecialCases(fun)
	}
}

func (p *Parser) checkEntryPointSpecialCases(fun *function) {
	if len(fun.Ast.Params) > 0 {
		p.PushErrorToken(fun.Ast.Token, "entrypoint_have_parameters")
	}
	if fun.Ast.ReturnType.Code != x.Void {
		p.PushErrorToken(fun.Ast.ReturnType.Token, "entrypoint_have_return")
	}
	if fun.Attributes != nil {
		p.PushErrorToken(fun.Ast.Token, "entrypoint_have_attributes")
	}
}

func (p *Parser) checkBlock(b *ast.BlockAST) {
	for index := 0; index < len(b.Statements); index++ {
		model := &b.Statements[index]
		switch t := model.Value.(type) {
		case ast.BlockExprAST:
			_, t.Expr.Model = p.computeExpr(t.Expr)
			model.Value = t
		case ast.VariableAST:
			p.checkVariableStatement(&t, false)
			model.Value = t
		case ast.VariableSetAST:
			p.checkVarsetStatement(&t)
			model.Value = t
		case ast.FreeAST:
			p.checkFreeStatement(&t)
			model.Value = t
		case ast.IterAST:
			p.checkIterExpression(&t)
			model.Value = t
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

type returnChecker struct {
	p        *Parser
	retAST   *ast.ReturnAST
	fun      ast.FunctionAST
	expModel multiReturnExprModel
	values   []value
}

func (rc *returnChecker) pushValue(last, current int, errTk lex.Token) {
	if current-last == 0 {
		rc.p.PushErrorToken(errTk, "missing_value")
		return
	}
	tokens := rc.retAST.Expr.Tokens[last:current]
	value, model := rc.p.computeTokens(tokens)
	rc.expModel.models = append(rc.expModel.models, model)
	rc.values = append(rc.values, value)
}

func (rc *returnChecker) checkValues() {
	braceCount := 0
	last := 0
	for index, token := range rc.retAST.Expr.Tokens {
		if token.Id == lex.Brace {
			switch token.Kind {
			case "(", "{", "[":
			default:
				braceCount--
			}
		}
		if braceCount > 0 || token.Id != lex.Comma {
			continue
		}
		rc.pushValue(last, index, token)
		last = index + 1
	}
	length := len(rc.retAST.Expr.Tokens)
	if last < length {
		if last == 0 {
			rc.pushValue(0, length, rc.retAST.Token)
		} else {
			rc.pushValue(last, length, rc.retAST.Expr.Tokens[last-1])
		}
	}
	if !typeIsVoidReturn(rc.fun.ReturnType) {
		rc.checkValueTypes()
	}
}

func (rc *returnChecker) checkValueTypes() {
	valLength := len(rc.values)
	if !rc.fun.ReturnType.MultiTyped {
		rc.retAST.Expr.Model = rc.expModel.models[0]
		if valLength > 1 {
			rc.p.PushErrorToken(rc.retAST.Token, "overflow_return")
		}
		rc.p.checkType(rc.fun.ReturnType, rc.values[0].ast.Type, true, rc.retAST.Token)
		return
	}
	// Multi return
	rc.retAST.Expr.Model = rc.expModel
	types := rc.fun.ReturnType.Tag.([]ast.DataTypeAST)
	if valLength == 1 {
		rc.p.PushErrorToken(rc.retAST.Token, "missing_multi_return")
	} else if valLength > len(types) {
		rc.p.PushErrorToken(rc.retAST.Token, "overflow_return")
	}
	for index, t := range types {
		if index >= valLength {
			break
		}
		rc.p.checkType(t, rc.values[index].ast.Type, true, rc.retAST.Token)
	}
}

func (rc *returnChecker) check() {
	exprTokensLen := len(rc.retAST.Expr.Tokens)
	if exprTokensLen == 0 && !typeIsVoidReturn(rc.fun.ReturnType) {
		rc.p.PushErrorToken(rc.retAST.Token, "require_return_value")
		return
	}
	if exprTokensLen > 0 && typeIsVoidReturn(rc.fun.ReturnType) {
		rc.p.PushErrorToken(rc.retAST.Token, "void_function_return_value")
	}
	rc.checkValues()
}

func (p *Parser) checkReturns(fun ast.FunctionAST) {
	missed := true
	for index, s := range fun.Block.Statements {
		switch t := s.Value.(type) {
		case ast.ReturnAST:
			rc := returnChecker{p: p, retAST: &t, fun: fun}
			rc.check()
			fun.Block.Statements[index].Value = t
			missed = false
		}
	}
	if missed && !typeIsVoidReturn(fun.ReturnType) {
		p.PushErrorToken(fun.Token, "missing_return")
	}
}

func (p *Parser) checkFunction(fun *ast.FunctionAST) {
	p.checkBlock(&fun.Block)
	p.checkReturns(*fun)
	p.checkParameters(fun.Params)
}

func (p *Parser) checkVariableStatement(varAST *ast.VariableAST, noParse bool) {
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
	if !noParse {
		*varAST = p.Variable(*varAST)
	}
	p.BlockVariables = append(p.BlockVariables, *varAST)
}

func (p *Parser) checkVarsetOperation(selected value, err lex.Token) bool {
	if selected.constant {
		p.PushErrorToken(err, "const_value_update")
		return false
	}
	switch selected.ast.Type.Tag.(type) {
	case ast.FunctionAST:
		if p.FunctionByName(selected.ast.Token.Kind) != nil {
			p.PushErrorToken(err, "type_not_support_value_update")
			return false
		}
	}
	return true
}

func (p *Parser) checkOneVarset(vsAST *ast.VariableSetAST) {
	selected, _ := p.computeExpr(vsAST.SelectExprs[0].Expr)
	if !p.checkVarsetOperation(selected, vsAST.Setter) {
		return
	}
	value, model := p.computeExpr(vsAST.ValueExprs[0])
	vsAST.ValueExprs[0] = model.ExprAST()
	if vsAST.Setter.Kind != "=" {
		vsAST.Setter.Kind = vsAST.Setter.Kind[:len(vsAST.Setter.Kind)-1]
		solver := solver{
			p:        p,
			left:     vsAST.SelectExprs[0].Expr.Tokens,
			leftVal:  selected.ast,
			right:    vsAST.ValueExprs[0].Tokens,
			rightVal: value.ast,
			operator: vsAST.Setter,
		}
		value.ast = solver.Solve()
		vsAST.Setter.Kind += "="
	}
	p.checkType(selected.ast.Type, value.ast.Type, false, vsAST.Setter)
}

func (p *Parser) parseVarsetSelections(vsAST *ast.VariableSetAST) {
	for index, selector := range vsAST.SelectExprs {
		p.checkVariableStatement(&selector.Variable, false)
		vsAST.SelectExprs[index] = selector
	}
}

func (p *Parser) getVarsetTypes(vsAST *ast.VariableSetAST) []ast.DataTypeAST {
	values := make([]ast.DataTypeAST, len(vsAST.ValueExprs))
	for index, expr := range vsAST.ValueExprs {
		val, model := p.computeExpr(expr)
		vsAST.ValueExprs[index].Model = model
		values[index] = val.ast.Type
	}
	return values
}

func (p *Parser) processFuncMultiVarset(vsAST *ast.VariableSetAST, funcVal value) {
	types := funcVal.ast.Type.Tag.([]ast.DataTypeAST)
	if len(types) != len(vsAST.SelectExprs) {
		p.PushErrorToken(vsAST.Setter, "missing_multiassign_identifiers")
		return
	}
	p.processMultiVarset(vsAST, types)
}

func (p *Parser) processMultiVarset(vsAST *ast.VariableSetAST, types []ast.DataTypeAST) {
	for index := range vsAST.SelectExprs {
		selector := &vsAST.SelectExprs[index]
		selector.Ignore = x.IsIgnoreName(selector.Variable.Name)
		dt := types[index]
		if !selector.NewVariable {
			if selector.Ignore {
				continue
			}
			selected, _ := p.computeExpr(selector.Expr)
			if !p.checkVarsetOperation(selected, vsAST.Setter) {
				return
			}
			p.checkType(selected.ast.Type, dt, false, vsAST.Setter)
			continue
		}
		selector.Variable.Tag = dt
		p.checkVariableStatement(&selector.Variable, false)
	}
}

func (p *Parser) checkVarsetStatement(vsAST *ast.VariableSetAST) {
	selectLength := len(vsAST.SelectExprs)
	valueLength := len(vsAST.ValueExprs)
	if vsAST.JustDeclare {
		p.parseVarsetSelections(vsAST)
		return
	} else if selectLength == 1 && !vsAST.SelectExprs[0].NewVariable {
		p.checkOneVarset(vsAST)
		return
	} else if vsAST.Setter.Kind != "=" {
		p.PushErrorToken(vsAST.Setter, "invalid_syntax")
		return
	}
	if valueLength == 1 {
		firstVal, _ := p.computeExpr(vsAST.ValueExprs[0])
		if firstVal.ast.Type.MultiTyped {
			vsAST.MultipleReturn = true
			p.processFuncMultiVarset(vsAST, firstVal)
			return
		}
	}
	switch {
	case selectLength > valueLength:
		p.PushErrorToken(vsAST.Setter, "overflow_multiassign_identifiers")
		return
	case selectLength < valueLength:
		p.PushErrorToken(vsAST.Setter, "missing_multiassign_identifiers")
		return
	}
	p.processMultiVarset(vsAST, p.getVarsetTypes(vsAST))
}

func (p *Parser) checkFreeStatement(freeAST *ast.FreeAST) {
	val, model := p.computeExpr(freeAST.Expr)
	freeAST.Expr.Model = model
	if !typeIsPointer(val.ast.Type) {
		p.PushErrorToken(freeAST.Token, "free_nonpointer")
	}
}

func (p *Parser) checkIterExpression(iterAST *ast.IterAST) {
	p.checkBlock(&iterAST.Block)
}
