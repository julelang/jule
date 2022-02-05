package ast

import (
	"fmt"
	"strings"

	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
)

// Object is an element of AST.
type Object struct {
	Token lex.Token
	Value interface{}
}

// StatementAST is statement.
type StatementAST struct {
	Token lex.Token
	Value interface{}
}

func (s StatementAST) String() string {
	return fmt.Sprint(s.Value)
}

// RangeAST represents block range or etc.
type RangeAST struct {
	Type    uint8
	Content []Object
}

// BlockAST is code block.
type BlockAST struct {
	Statements []StatementAST
}

// Indent total of blocks.
var Indent = 1

func (b BlockAST) String() string {
	Indent = 1
	return ParseBlock(b, Indent)
}

// IndentSpace of blocks.
const IndentSpace = 2

// ParseBlock to cxx.
func ParseBlock(b BlockAST, indent int) string {
	// Space count per indent.
	var cxx strings.Builder
	cxx.WriteByte('{')
	for _, s := range b.Statements {
		cxx.WriteByte('\n')
		cxx.WriteString(strings.Repeat(" ", indent*IndentSpace))
		cxx.WriteString(s.String())
	}
	cxx.WriteByte('\n')
	cxx.WriteString(strings.Repeat(" ", (indent-1)*IndentSpace) + "}")
	return cxx.String()
}

// DataTypeAST is data type identifier.
type DataTypeAST struct {
	Token      lex.Token
	Code       uint8
	Value      string
	MultiTyped bool
	Tag        interface{}
}

func (dt DataTypeAST) String() string {
	var cxx strings.Builder
	for index, run := range dt.Value {
		if run == '*' {
			cxx.WriteRune(run)
			continue
		}
		dt.Value = dt.Value[index:]
		break
	}
	if dt.MultiTyped {
		return dt.MultiTypeString() + cxx.String()
	}
	if dt.Value != "" && dt.Value[0] == '[' {
		pointers := cxx.String()
		cxx.Reset()
		cxx.WriteString("array<")
		dt.Value = dt.Value[2:]
		cxx.WriteString(dt.String())
		cxx.WriteByte('>')
		cxx.WriteString(pointers)
		return cxx.String()
	}
	switch dt.Code {
	case x.Name:
		return dt.Token.Kind + cxx.String()
	case x.Function:
		return dt.FunctionString() + cxx.String()
	}
	return x.CxxTypeNameFromType(dt.Code) + cxx.String()
}

func (dt DataTypeAST) FunctionString() string {
	cxx := "function<"
	fun := dt.Tag.(FunctionAST)
	cxx += fun.ReturnType.String()
	cxx += "("
	if len(fun.Params) > 0 {
		for _, param := range fun.Params {
			cxx += param.Type.String() + ", "
		}
		cxx = cxx[:len(cxx)-2]
	}
	cxx += ")>"
	return cxx
}

func (dt DataTypeAST) MultiTypeString() string {
	types := dt.Tag.([]DataTypeAST)
	var cxx strings.Builder
	cxx.WriteString("std::tuple<")
	for _, t := range types {
		cxx.WriteString(t.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ">"
}

// TypeAST is type declaration.
type TypeAST struct {
	Token lex.Token
	Name  string
	Type  DataTypeAST
}

func (t TypeAST) String() string {
	return "typedef " + t.Type.String() + " " + t.Name + ";"
}

// FunctionAST is function declaration AST model.
type FunctionAST struct {
	Token      lex.Token
	Name       string
	Params     []ParameterAST
	ReturnType DataTypeAST
	Block      BlockAST
}

// ParameterAST is function parameter AST model.
type ParameterAST struct {
	Token lex.Token
	Name  string
	Const bool
	Type  DataTypeAST
}

func (p ParameterAST) String() string {
	var cxx strings.Builder
	if p.Const {
		cxx.WriteString("const ")
	}
	cxx.WriteString(p.Type.String())
	if p.Name != "" {
		return cxx.String() + " " + p.Name
	}
	return cxx.String()
}

// DataTypeString returns data type string of function.
func (fc FunctionAST) DataTypeString() string {
	dt := "("
	if len(fc.Params) > 0 {
		for _, param := range fc.Params {
			dt += param.Type.Value + ", "
		}
		dt = dt[:len(dt)-2]
	}
	dt += ")"
	if fc.ReturnType.Code != x.Void {
		dt += fc.ReturnType.Value
	}
	return dt
}

// ArgAST is AST model of argument.
type ArgAST struct {
	Token  lex.Token
	Tokens []lex.Token
	Expr   ExprAST
}

func (a ArgAST) String() string {
	return a.Expr.String()
}

// ExprAST is AST model of expression.
type ExprAST struct {
	Tokens    []lex.Token
	Processes [][]lex.Token
	Model     IExprModel
}

// IExprModel for special expression model to Cxx string.
type IExprModel interface {
	String() string
}

func (e ExprAST) String() string {
	if e.Model != nil {
		return e.Model.String()
	}
	var sb strings.Builder
	for _, process := range e.Processes {
		if len(process) == 1 && process[0].Id == lex.Operator {
			sb.WriteByte(' ')
			sb.WriteString(process[0].Kind)
			sb.WriteByte(' ')
			continue
		}
		for _, token := range process {
			sb.WriteString(token.Kind)
		}
	}
	return sb.String()
}

// BlockExprAST is AST model of expression statement in block.
type BlockExprAST struct {
	Expr ExprAST
}

func (be BlockExprAST) String() string {
	return be.Expr.String() + ";"
}

// ValueAST is AST model of constant value.
type ValueAST struct {
	Token lex.Token
	Value string
	Type  DataTypeAST
}

func (v ValueAST) String() string {
	return v.Value
}

// BraceAST is AST model of brace.
type BraceAST struct {
	Token lex.Token
}

func (b BraceAST) String() string {
	return b.Token.Kind
}

// OperatorAST is AST model of operator.
type OperatorAST struct {
	Token lex.Token
}

func (o OperatorAST) String() string {
	return o.Token.Kind
}

// ReturnAST is return statement AST model.
type ReturnAST struct {
	Token lex.Token
	Expr  ExprAST
}

func (r ReturnAST) String() string {
	switch r.Token.Id {
	case lex.Operator:
		return "return " + r.Expr.String() + ";"
	}
	return "return " + r.Expr.String() + ";"
}

// AttributeAST is attribtue AST model.
type AttributeAST struct {
	Token lex.Token
	Tag   lex.Token
}

func (a AttributeAST) String() string {
	return a.Tag.Kind[1:] // Remove name underscore at start
}

// VariableAST is variable declaration AST model.
type VariableAST struct {
	DefineToken lex.Token
	NameToken   lex.Token
	SetterToken lex.Token
	Name        string
	Type        DataTypeAST
	Value       ExprAST
	Tag         interface{}
}

func (v VariableAST) String() string {
	var sb strings.Builder
	switch v.DefineToken.Id {
	case lex.Const:
		sb.WriteString("const ")
	}
	sb.WriteString(v.StringType())
	sb.WriteByte(' ')
	sb.WriteString(v.Name)
	if v.Value.Processes != nil {
		sb.WriteString(" = ")
		sb.WriteString(v.Value.String())
	}
	sb.WriteByte(';')
	return sb.String()
}

// StringType parses type to cxx.
func (v VariableAST) StringType() string {
	if v.Type.Code == x.Void {
		return "auto"
	}
	return v.Type.String()
}

// VarsetSelector is selector for variable set operation.
type VarsetSelector struct {
	NewVariable bool
	Variable    VariableAST
	Expr        ExprAST
	Ignore      bool
}

func (vs VarsetSelector) String() string {
	if vs.NewVariable {
		return vs.Expr.Tokens[0].Kind // Returns variable name.
	}
	return vs.Expr.String()
}

// VariableSetAST is variable set AST model.
type VariableSetAST struct {
	Setter         lex.Token
	SelectExprs    []VarsetSelector
	ValueExprs     []ExprAST
	JustDeclare    bool
	MultipleReturn bool
}

func (vs VariableSetAST) cxxSingleSet(cxx *strings.Builder) string {
	cxx.WriteString(vs.SelectExprs[0].String())
	cxx.WriteString(vs.Setter.Kind)
	cxx.WriteString(vs.ValueExprs[0].String())
	cxx.WriteByte(';')
	return cxx.String()
}

func (vs VariableSetAST) cxxMultipleSet(cxx *strings.Builder) string {
	cxx.WriteString("std::tie(")
	var expCxx strings.Builder
	expCxx.WriteString("std::make_tuple(")
	for index, selector := range vs.SelectExprs {
		if selector.Ignore {
			continue
		}
		cxx.WriteString(selector.String())
		cxx.WriteByte(',')
		expCxx.WriteString(vs.ValueExprs[index].String())
		expCxx.WriteByte(',')
	}
	str := cxx.String()[:cxx.Len()-1] + ")"
	cxx.Reset()
	cxx.WriteString(str)
	cxx.WriteString(vs.Setter.Kind)
	cxx.WriteString(expCxx.String()[:expCxx.Len()-1] + ")")
	cxx.WriteByte(';')
	return cxx.String()
}

func (vs VariableSetAST) cxxMultipleReturn(cxx *strings.Builder) string {
	cxx.WriteString("std::tie(")
	for _, selector := range vs.SelectExprs {
		if selector.Ignore {
			cxx.WriteString("std::ignore,")
			continue
		}
		cxx.WriteString(selector.String())
		cxx.WriteByte(',')
	}
	str := cxx.String()[:cxx.Len()-1]
	cxx.Reset()
	cxx.WriteString(str)
	cxx.WriteByte(')')
	cxx.WriteString(vs.Setter.Kind)
	cxx.WriteString(vs.ValueExprs[0].String())
	cxx.WriteByte(';')
	return cxx.String()
}

func (vs VariableSetAST) cxxNewDefines(cxx *strings.Builder) {
	for _, selector := range vs.SelectExprs {
		if selector.Ignore || !selector.NewVariable {
			continue
		}
		cxx.WriteString(selector.Variable.String() + " ")
	}
}

func (vs VariableSetAST) String() string {
	var cxx strings.Builder
	vs.cxxNewDefines(&cxx)
	if vs.JustDeclare {
		return cxx.String()[:cxx.Len()-1] /* Remove unnecesarry whitespace. */
	}
	if vs.MultipleReturn {
		return vs.cxxMultipleReturn(&cxx)
	} else if len(vs.SelectExprs) == 1 {
		return vs.cxxSingleSet(&cxx)
	}
	return vs.cxxMultipleSet(&cxx)
}

type FreeAST struct {
	Token lex.Token
	Expr  ExprAST
}

func (f FreeAST) String() string {
	return "delete " + f.Expr.String() + ";"
}
