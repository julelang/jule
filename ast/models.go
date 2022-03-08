package ast

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
	"github.com/the-xlang/x/pkg/xapi"
)

// Object is an element of AST.
type Object struct {
	Token lex.Token
	Value interface{}
}

// StatementAST is statement.
type StatementAST struct {
	Token          lex.Token
	Value          interface{}
	WithTerminator bool
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
var Indent int32 = 0

func (b BlockAST) String() string {
	atomic.SwapInt32(&Indent, Indent+1)
	defer func() { atomic.SwapInt32(&Indent, Indent-1) }()
	return ParseBlock(b, int(Indent))
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
		return xapi.AsId(dt.Token.Kind) + cxx.String()
	case x.Func:
		return dt.FunctionString() + cxx.String()
	default:
		return x.CxxTypeNameFromType(dt.Code) + cxx.String()
	}
}

func (dt DataTypeAST) FunctionString() string {
	var cxx strings.Builder
	cxx.WriteString("std::function<")
	fun := dt.Tag.(FuncAST)
	cxx.WriteString(fun.RetType.String())
	cxx.WriteByte('(')
	if len(fun.Params) > 0 {
		for _, param := range fun.Params {
			cxx.WriteString(param.Type.String())
			cxx.WriteString(", ")
		}
		cxxStr := cxx.String()[:cxx.Len()-1]
		cxx.Reset()
		cxx.WriteString(cxxStr)
	} else {
		cxx.WriteString("void")
	}
	cxx.WriteString(")>")
	return cxx.String()
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
	Id    string
	Type  DataTypeAST
}

func (t TypeAST) String() string {
	var cxx strings.Builder
	cxx.WriteString("typedef ")
	cxx.WriteString(t.Type.String())
	cxx.WriteByte(' ')
	cxx.WriteString(xapi.AsId(t.Id))
	cxx.WriteByte(';')
	return cxx.String()
}

// FuncAST is function declaration AST model.
type FuncAST struct {
	Token   lex.Token
	Id      string
	Params  []ParameterAST
	RetType DataTypeAST
	Block   BlockAST
}

// DataTypeString returns data type string of function.
func (fc FuncAST) DataTypeString() string {
	var cxx strings.Builder
	cxx.WriteByte('(')
	if len(fc.Params) > 0 {
		for _, param := range fc.Params {
			cxx.WriteString(param.Type.String())
			cxx.WriteString(", ")
		}
		cxxStr := cxx.String()[:cxx.Len()-2]
		cxx.Reset()
		cxx.WriteString(cxxStr)
	}
	cxx.WriteByte(')')
	if fc.RetType.Code != x.Void {
		cxx.WriteString(fc.RetType.String())
	}
	return cxx.String()
}

// ParameterAST is function parameter AST model.
type ParameterAST struct {
	Token    lex.Token
	Id       string
	Const    bool
	Volatile bool
	Variadic bool
	Type     DataTypeAST
}

func (p ParameterAST) String() string {
	var cxx strings.Builder
	cxx.WriteString(p.Prototype())
	if p.Id != "" {
		cxx.WriteByte(' ')
		cxx.WriteString(xapi.AsId(p.Id))
	}
	if p.Variadic {
		cxx.WriteString(" =array<")
		cxx.WriteString(p.Type.String())
		cxx.WriteString(">()")
	}
	return cxx.String()
}

// Prototype returns prototype cxx of parameter.
func (p ParameterAST) Prototype() string {
	var cxx strings.Builder
	if p.Volatile {
		cxx.WriteString("volatile ")
	}
	if p.Const {
		cxx.WriteString("const ")
	}
	if p.Variadic {
		cxx.WriteString("array<")
		cxx.WriteString(p.Type.String())
		cxx.WriteByte('>')
	} else {
		cxx.WriteString(p.Type.String())
	}
	return cxx.String()
}

// ArgAST is AST model of argument.
type ArgAST struct {
	Token lex.Token
	Expr  ExprAST
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
	var expr strings.Builder
	for _, process := range e.Processes {
		for _, token := range process {
			switch token.Id {
			case lex.Id:
				expr.WriteString(xapi.AsId(token.Kind))
			default:
				expr.WriteString(token.Kind)
			}
		}
	}
	return expr.String()
}

// ExprStatementAST is AST model of expression statement in block.
type ExprStatementAST struct {
	Expr ExprAST
}

func (be ExprStatementAST) String() string {
	var cxx strings.Builder
	cxx.WriteString(be.Expr.String())
	cxx.WriteByte(';')
	return cxx.String()
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

// ReturnAST is return statement AST model.
type ReturnAST struct {
	Token lex.Token
	Expr  ExprAST
}

func (r ReturnAST) String() string {
	var cxx strings.Builder
	cxx.WriteString("return ")
	cxx.WriteString(r.Expr.String())
	cxx.WriteByte(';')
	return cxx.String()
}

// AttributeAST is attribtue AST model.
type AttributeAST struct {
	Token lex.Token
	Tag   lex.Token
}

func (a AttributeAST) String() string {
	return a.Tag.Kind
}

// VariableAST is variable declaration AST model.
type VariableAST struct {
	DefToken    lex.Token
	IdToken     lex.Token
	SetterToken lex.Token
	Id          string
	Type        DataTypeAST
	Value       ExprAST
	Const       bool
	Volatile    bool
	New         bool
	Tag         interface{}
}

func (v VariableAST) String() string {
	var sb strings.Builder
	if v.Volatile {
		sb.WriteString("volatile ")
	}
	if v.Const {
		sb.WriteString("const ")
	}
	sb.WriteString(v.Type.String())
	sb.WriteByte(' ')
	sb.WriteString(xapi.AsId(v.Id))
	if v.Value.Processes != nil {
		sb.WriteString(" = ")
		sb.WriteString(v.Value.String())
	}
	sb.WriteByte(';')
	return sb.String()
}

// AssignSelector is selector for assignment operation.
type AssignSelector struct {
	NewVariable bool
	Var         VariableAST
	Expr        ExprAST
	Ignore      bool
}

func (vs AssignSelector) String() string {
	if vs.NewVariable {
		// Returns variable name.
		return xapi.AsId(vs.Expr.Tokens[0].Kind)
	}
	return vs.Expr.String()
}

// AssignAST is assignment AST model.
type AssignAST struct {
	Setter         lex.Token
	SelectExprs    []AssignSelector
	ValueExprs     []ExprAST
	IsExpr         bool
	JustDeclare    bool
	MultipleReturn bool
}

func (vs AssignAST) cxxSingleAssign() string {
	var cxx strings.Builder
	expr := vs.SelectExprs[0]
	if len(expr.Expr.Tokens) != 1 ||
		!xapi.IsIgnoreId(expr.Expr.Tokens[0].Kind) {
		cxx.WriteString(vs.SelectExprs[0].String())
		cxx.WriteString(vs.Setter.Kind)
	}
	cxx.WriteString(vs.ValueExprs[0].String())
	return cxx.String()
}

func (vs AssignAST) cxxMultipleAssign() string {
	var cxx strings.Builder
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
	return cxx.String()
}

func (vs AssignAST) cxxMultipleReturn() string {
	var cxx strings.Builder
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
	return cxx.String()
}

func (vs AssignAST) cxxNewDefines() string {
	var cxx strings.Builder
	for _, selector := range vs.SelectExprs {
		if selector.Ignore || !selector.NewVariable {
			continue
		}
		cxx.WriteString(selector.Var.String() + " ")
	}
	return cxx.String()
}

func (vs AssignAST) String() string {
	var cxx strings.Builder
	cxx.WriteString(vs.cxxNewDefines())
	if vs.JustDeclare {
		return cxx.String()[:cxx.Len()-1] /* Remove unnecesarry whitespace. */
	}
	switch {
	case vs.MultipleReturn:
		cxx.WriteString(vs.cxxMultipleReturn())
	case len(vs.SelectExprs) == 1:
		cxx.WriteString(vs.cxxSingleAssign())
	default:
		cxx.WriteString(vs.cxxMultipleAssign())
	}
	if !vs.IsExpr {
		cxx.WriteByte(';')
	}
	return cxx.String()
}

type FreeAST struct {
	Token lex.Token
	Expr  ExprAST
}

func (f FreeAST) String() string {
	var cxx strings.Builder
	cxx.WriteString("delete ")
	cxx.WriteString(f.Expr.String())
	cxx.WriteByte(';')
	return cxx.String()
}

// IterProfile interface for iteration profiles.
type IterProfile interface {
	String(iter IterAST) string
}

// WhileProfile is while iteration profile.
type WhileProfile struct {
	Expr ExprAST
}

func (wp WhileProfile) String(iter IterAST) string {
	var cxx strings.Builder
	cxx.WriteString("while (")
	cxx.WriteString(wp.Expr.String())
	cxx.WriteString(") ")
	cxx.WriteString(iter.Block.String())
	return cxx.String()
}

// ForeachProfile is foreach iteration profile.
type ForeachProfile struct {
	KeyA     VariableAST
	KeyB     VariableAST
	InToken  lex.Token
	Expr     ExprAST
	ExprType DataTypeAST
}

func (fp ForeachProfile) String(iter IterAST) string {
	if !xapi.IsIgnoreId(fp.KeyA.Id) {
		return fp.ForeachString(iter)
	}
	return fp.IterationSring(iter)
}

func (fp ForeachProfile) ForeachString(iter IterAST) string {
	var cxx strings.Builder
	cxx.WriteString("foreach<")
	cxx.WriteString(fp.ExprType.String())
	cxx.WriteByte(',')
	cxx.WriteString(fp.KeyA.Type.String())
	if !xapi.IsIgnoreId(fp.KeyB.Id) {
		cxx.WriteByte(',')
		cxx.WriteString(fp.KeyB.Type.String())
	}
	cxx.WriteString(">(")
	cxx.WriteString(fp.Expr.String())
	cxx.WriteString(", [&](")
	cxx.WriteString(fp.KeyA.Type.String())
	cxx.WriteByte(' ')
	cxx.WriteString(xapi.AsId(fp.KeyA.Id))
	if !xapi.IsIgnoreId(fp.KeyB.Id) {
		cxx.WriteByte(',')
		cxx.WriteString(fp.KeyB.Type.String())
		cxx.WriteByte(' ')
		cxx.WriteString(xapi.AsId(fp.KeyB.Id))
	}
	cxx.WriteString(") -> void ")
	cxx.WriteString(iter.Block.String())
	cxx.WriteString(");")
	return cxx.String()
}

func (fp ForeachProfile) IterationSring(iter IterAST) string {
	var cxx strings.Builder
	cxx.WriteString("for (auto ")
	cxx.WriteString(xapi.AsId(fp.KeyB.Id))
	cxx.WriteString(" : ")
	cxx.WriteString(fp.Expr.String())
	cxx.WriteString(") ")
	cxx.WriteString(iter.Block.String())
	return cxx.String()
}

// IterAST is the AST model of iterations.
type IterAST struct {
	Token   lex.Token
	Block   BlockAST
	Profile IterProfile
}

func (iter IterAST) String() string {
	if iter.Profile == nil {
		var cxx strings.Builder
		cxx.WriteString("while (true) ")
		cxx.WriteString(iter.Block.String())
		return cxx.String()
	}
	return iter.Profile.String(iter)
}

// BreakAST is the AST model of break statement.
type BreakAST struct {
	Token lex.Token
}

func (b BreakAST) String() string {
	return "break;"
}

// ContinueAST is the AST model of break statement.
type ContinueAST struct {
	Token lex.Token
}

func (c ContinueAST) String() string {
	return "continue;"
}

// IfAST is the AST model of if expression.
type IfAST struct {
	Token lex.Token
	Expr  ExprAST
	Block BlockAST
}

func (ifast IfAST) String() string {
	var cxx strings.Builder
	cxx.WriteString("if (")
	cxx.WriteString(ifast.Expr.String())
	cxx.WriteString(") ")
	cxx.WriteString(ifast.Block.String())
	return cxx.String()
}

// ElseIfAST is the AST model of else if expression.
type ElseIfAST struct {
	Token lex.Token
	Expr  ExprAST
	Block BlockAST
}

func (elif ElseIfAST) String() string {
	var cxx strings.Builder
	cxx.WriteString("else if (")
	cxx.WriteString(elif.Expr.String())
	cxx.WriteString(") ")
	cxx.WriteString(elif.Block.String())
	return cxx.String()
}

// ElseAST is the AST model of else blocks.
type ElseAST struct {
	Token lex.Token
	Block BlockAST
}

func (elseast ElseAST) String() string {
	var cxx strings.Builder
	cxx.WriteString("else ")
	cxx.WriteString(elseast.Block.String())
	return cxx.String()
}

// CommentAST is the AST model of just comment lines.
type CommentAST struct {
	Token   lex.Token
	Content string
}

func (c CommentAST) String() string {
	var cxx strings.Builder
	cxx.WriteString("// ")
	cxx.WriteString(c.Content)
	return cxx.String()
}
