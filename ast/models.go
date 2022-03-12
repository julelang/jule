package ast

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
	"github.com/the-xlang/x/pkg/xapi"
)

// Obj is an element of AST.
type Obj struct {
	Token lex.Token
	Value interface{}
}

// Statement is statement.
type Statement struct {
	Token          lex.Token
	Value          interface{}
	WithTerminator bool
}

func (s Statement) String() string {
	return fmt.Sprint(s.Value)
}

// BlockAST is code block.
type BlockAST struct {
	Statements []Statement
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

// DataType is data type identifier.
type DataType struct {
	Token      lex.Token
	Code       uint8
	Value      string
	MultiTyped bool
	Tag        interface{}
}

func (dt DataType) String() string {
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
	case x.Id:
		return xapi.AsId(dt.Token.Kind) + cxx.String()
	case x.Func:
		return dt.FunctionString() + cxx.String()
	default:
		return x.CxxTypeIdFromType(dt.Code) + cxx.String()
	}
}

func (dt DataType) FunctionString() string {
	var cxx strings.Builder
	cxx.WriteString("std::function<")
	fun := dt.Tag.(Func)
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

func (dt DataType) MultiTypeString() string {
	types := dt.Tag.([]DataType)
	var cxx strings.Builder
	cxx.WriteString("std::tuple<")
	for _, t := range types {
		cxx.WriteString(t.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ">"
}

// Type is type declaration.
type Type struct {
	Pub         bool
	Token       lex.Token
	Id          string
	Type        DataType
	Description string
}

func (t Type) String() string {
	var cxx strings.Builder
	cxx.WriteString("typedef ")
	cxx.WriteString(t.Type.String())
	cxx.WriteByte(' ')
	cxx.WriteString(xapi.AsId(t.Id))
	cxx.WriteByte(';')
	return cxx.String()
}

// Func is function declaration AST model.
type Func struct {
	Pub     bool
	Token   lex.Token
	Id      string
	Params  []Parameter
	RetType DataType
	Block   BlockAST
}

// DataTypeString returns data type string of function.
func (fc Func) DataTypeString() string {
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

// Parameter is function parameter AST model.
type Parameter struct {
	Token    lex.Token
	Id       string
	Const    bool
	Volatile bool
	Variadic bool
	Type     DataType
}

func (p Parameter) String() string {
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
func (p Parameter) Prototype() string {
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

// Arg is AST model of argument.
type Arg struct {
	Token lex.Token
	Expr  Expr
}

func (a Arg) String() string {
	return a.Expr.String()
}

// Expr is AST model of expression.
type Expr struct {
	Tokens    []lex.Token
	Processes [][]lex.Token
	Model     IExprModel
}

// IExprModel for special expression model to Cxx string.
type IExprModel interface {
	String() string
}

func (e Expr) String() string {
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

// ExprStatement is AST model of expression statement in block.
type ExprStatement struct {
	Expr Expr
}

func (be ExprStatement) String() string {
	var cxx strings.Builder
	cxx.WriteString(be.Expr.String())
	cxx.WriteByte(';')
	return cxx.String()
}

// Value is AST model of constant value.
type Value struct {
	Token lex.Token
	Data  string
	Type  DataType
}

func (v Value) String() string {
	return v.Data
}

// Ret is return statement AST model.
type Ret struct {
	Token lex.Token
	Expr  Expr
}

func (r Ret) String() string {
	var cxx strings.Builder
	cxx.WriteString("return ")
	cxx.WriteString(r.Expr.String())
	cxx.WriteByte(';')
	return cxx.String()
}

// Attribute is attribtue AST model.
type Attribute struct {
	Token lex.Token
	Tag   lex.Token
}

func (a Attribute) String() string {
	return a.Tag.Kind
}

// Var is variable declaration AST model.
type Var struct {
	Pub         bool
	DefToken    lex.Token
	IdToken     lex.Token
	SetterToken lex.Token
	Id          string
	Type        DataType
	Value       Expr
	Const       bool
	Volatile    bool
	New         bool
	Tag         interface{}
	Description string
}

func (v Var) String() string {
	var cxx strings.Builder
	if v.Volatile {
		cxx.WriteString("volatile ")
	}
	if v.Const {
		cxx.WriteString("const ")
	}
	cxx.WriteString(v.Type.String())
	cxx.WriteByte(' ')
	cxx.WriteString(xapi.AsId(v.Id))
	if v.Value.Processes != nil {
		cxx.WriteString(" = ")
		cxx.WriteString(v.Value.String())
	}
	cxx.WriteByte(';')
	return cxx.String()
}

// AssignSelector is selector for assignment operation.
type AssignSelector struct {
	NewVariable bool
	Var         Var
	Expr        Expr
	Ignore      bool
}

func (vs AssignSelector) String() string {
	if vs.NewVariable {
		// Returns variable identifier.
		return xapi.AsId(vs.Expr.Tokens[0].Kind)
	}
	return vs.Expr.String()
}

// Assign is assignment AST model.
type Assign struct {
	Setter         lex.Token
	SelectExprs    []AssignSelector
	ValueExprs     []Expr
	IsExpr         bool
	JustDeclare    bool
	MultipleReturn bool
}

func (vs Assign) cxxSingleAssign() string {
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

func (vs Assign) cxxMultipleAssign() string {
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

func (vs Assign) cxxMultipleReturn() string {
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

func (vs Assign) cxxNewDefines() string {
	var cxx strings.Builder
	for _, selector := range vs.SelectExprs {
		if selector.Ignore || !selector.NewVariable {
			continue
		}
		cxx.WriteString(selector.Var.String() + " ")
	}
	return cxx.String()
}

func (vs Assign) String() string {
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

type Free struct {
	Token lex.Token
	Expr  Expr
}

func (f Free) String() string {
	var cxx strings.Builder
	cxx.WriteString("delete ")
	cxx.WriteString(f.Expr.String())
	cxx.WriteByte(';')
	return cxx.String()
}

// IterProfile interface for iteration profiles.
type IterProfile interface {
	String(iter Iter) string
}

// WhileProfile is while iteration profile.
type WhileProfile struct {
	Expr Expr
}

func (wp WhileProfile) String(iter Iter) string {
	var cxx strings.Builder
	cxx.WriteString("while (")
	cxx.WriteString(wp.Expr.String())
	cxx.WriteString(") ")
	cxx.WriteString(iter.Block.String())
	return cxx.String()
}

// ForeachProfile is foreach iteration profile.
type ForeachProfile struct {
	KeyA     Var
	KeyB     Var
	InToken  lex.Token
	Expr     Expr
	ExprType DataType
}

func (fp ForeachProfile) String(iter Iter) string {
	if !xapi.IsIgnoreId(fp.KeyA.Id) {
		return fp.ForeachString(iter)
	}
	return fp.IterationSring(iter)
}

func (fp ForeachProfile) ForeachString(iter Iter) string {
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

func (fp ForeachProfile) IterationSring(iter Iter) string {
	var cxx strings.Builder
	cxx.WriteString("for (auto ")
	cxx.WriteString(xapi.AsId(fp.KeyB.Id))
	cxx.WriteString(" : ")
	cxx.WriteString(fp.Expr.String())
	cxx.WriteString(") ")
	cxx.WriteString(iter.Block.String())
	return cxx.String()
}

// Iter is the AST model of iterations.
type Iter struct {
	Token   lex.Token
	Block   BlockAST
	Profile IterProfile
}

func (iter Iter) String() string {
	if iter.Profile == nil {
		var cxx strings.Builder
		cxx.WriteString("while (true) ")
		cxx.WriteString(iter.Block.String())
		return cxx.String()
	}
	return iter.Profile.String(iter)
}

// Break is the AST model of break statement.
type Break struct {
	Token lex.Token
}

func (b Break) String() string {
	return "break;"
}

// Continue is the AST model of break statement.
type Continue struct {
	Token lex.Token
}

func (c Continue) String() string {
	return "continue;"
}

// If is the AST model of if expression.
type If struct {
	Token lex.Token
	Expr  Expr
	Block BlockAST
}

func (ifast If) String() string {
	var cxx strings.Builder
	cxx.WriteString("if (")
	cxx.WriteString(ifast.Expr.String())
	cxx.WriteString(") ")
	cxx.WriteString(ifast.Block.String())
	return cxx.String()
}

// ElseIf is the AST model of else if expression.
type ElseIf struct {
	Token lex.Token
	Expr  Expr
	Block BlockAST
}

func (elif ElseIf) String() string {
	var cxx strings.Builder
	cxx.WriteString("else if (")
	cxx.WriteString(elif.Expr.String())
	cxx.WriteString(") ")
	cxx.WriteString(elif.Block.String())
	return cxx.String()
}

// Else is the AST model of else blocks.
type Else struct {
	Token lex.Token
	Block BlockAST
}

func (elseast Else) String() string {
	var cxx strings.Builder
	cxx.WriteString("else ")
	cxx.WriteString(elseast.Block.String())
	return cxx.String()
}

// Comment is the AST model of just comment lines.
type Comment struct {
	Token   lex.Token
	Content string
}

func (c Comment) String() string {
	var cxx strings.Builder
	cxx.WriteString("// ")
	cxx.WriteString(c.Content)
	return cxx.String()
}

// Use is the AST model of use statement.
type Use struct {
	Token lex.Token
	Path  string
}
