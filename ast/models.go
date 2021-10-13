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

// IdentifierAST is identifier.
type IdentifierAST struct {
	Type  uint8
	Value string
}

// StatementAST is statement.
type StatementAST struct {
	Token lex.Token
	Value interface{}
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

func (b BlockAST) String() string {
	return parseBlock(b, 1)
}

func parseBlock(b BlockAST, indent int) string {
	// Space count per indent.
	const indentSpace = 2
	var cxx strings.Builder
	for _, s := range b.Statements {
		cxx.WriteByte('\n')
		cxx.WriteString(fullString(' ', indent*indentSpace))
		cxx.WriteString(fmt.Sprint(s.Value))
	}
	return cxx.String()
}

func fullString(b byte, count int) string {
	var sb strings.Builder
	for count > 0 {
		count--
		sb.WriteByte(b)
	}
	return sb.String()
}

// DataTypeAST is data type identifier.
type DataTypeAST struct {
	Token lex.Token
	Code  uint8
	Value string
}

func (dt DataTypeAST) String() string {
	var cxx strings.Builder
	for _, run := range dt.Value {
		if run == '*' {
			cxx.WriteRune(run)
			continue
		}
		break
	}
	if dt.Code == x.Name {
		return dt.Token.Value + cxx.String()
	}
	return x.CxxTypeNameFromType(dt.Code) + cxx.String()
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
	Type  DataTypeAST
}

func (p ParameterAST) String() string {
	return p.Type.String() + " " + p.Name
}

// FunctionAST is function declaration AST model.
type FunctionCallAST struct {
	Token lex.Token
	Name  string
	Args  []ArgAST
}

func (fc FunctionCallAST) String() string {
	var cxx string
	cxx += fc.Name
	cxx += "("
	if len(fc.Args) > 0 {
		for _, arg := range fc.Args {
			cxx += arg.String()
			cxx += ","
		}
		cxx = cxx[:len(cxx)-1]
	}
	cxx += ");"
	return cxx
}

// ArgAST is AST model of argument.
type ArgAST struct {
	Token      lex.Token
	Tokens     []lex.Token
	Expression ExpressionAST
}

func (a ArgAST) String() string {
	return a.Expression.String()
}

// ExpressionAST is AST model of expression.
type ExpressionAST struct {
	Tokens    []lex.Token
	Processes [][]lex.Token
	Model     ExpressionModel
}

// ExpressionModel for special expression model to Cxx string.
type ExpressionModel interface {
	String() string
}

func (e ExpressionAST) String() string {
	if e.Model != nil {
		return e.Model.String()
	}
	var sb strings.Builder
	for _, process := range e.Processes {
		if len(process) == 1 && process[0].Type == lex.Operator {
			sb.WriteByte(' ')
			sb.WriteString(process[0].Value)
			sb.WriteByte(' ')
			continue
		}
		for _, token := range process {
			sb.WriteString(token.Value)
		}
	}
	return sb.String()
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
	Value string
}

func (b BraceAST) String() string {
	return b.Value
}

// OperatorAST is AST model of operator.
type OperatorAST struct {
	Token lex.Token
	Value string
}

func (o OperatorAST) String() string {
	return o.Value
}

// ReturnAST is return statement AST model.
type ReturnAST struct {
	Token      lex.Token
	Expression ExpressionAST
}

func (r ReturnAST) String() string {
	if r.Token.Type == lex.Operator {
		return "return " + r.Expression.String() + ";"
	}
	return "return " + r.Expression.String() + ";"
}

// AttributeAST is attribtue AST model.
type AttributeAST struct {
	Token lex.Token
	Value string
}

func (a AttributeAST) String() string {
	return a.Value
}

// VariableAST is variable declaration AST model.
type VariableAST struct {
	DefineToken lex.Token
	NameToken   lex.Token
	SetterToken lex.Token
	Name        string
	Type        DataTypeAST
	Value       ExpressionAST
}

func (v VariableAST) String() string {
	var sb strings.Builder
	if v.DefineToken.Value == "const" {
		sb.WriteString("const ")
	}
	sb.WriteString(v.StringType())
	sb.WriteByte(' ')
	sb.WriteString(v.Name)
	sb.WriteString(" = ")
	sb.WriteString(v.Value.String())
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

// VariableSetAST is variable set AST model.
type VariableSetAST struct {
	Setter           lex.Token
	SelectExpression ExpressionAST
	ValueExpression  ExpressionAST
}

func (vs VariableSetAST) String() string {
	return vs.SelectExpression.String() + " = " + vs.ValueExpression.String() + ";"
}
