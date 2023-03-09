package ast

import (
	"strings"

	"github.com/julelang/jule/lex"
)

type NodeData = any // Type of AST Node's data.

// AST Node.
type Node struct {
	Token lex.Token
	Data  any
}

// Group for AST model of comments.
type CommentGroup struct {
	Comments []*Comment
}

// AST model of just comment lines.
type Comment struct {
	Token lex.Token
	Text  string
}

// Reports whether comment is directive.
func (c *Comment) IsDirective() bool {
	return strings.HasPrefix(c.Text, lex.DIRECTIVE_COMMENT_PREFIX)
}

// Directive AST.
type Directive struct {
	Token lex.Token
	Tag   string
}

// Kind type of data types.
type TypeKind interface {
	As_text() string
}

// Type AST.
type Type struct {
	Token lex.Token
	Kind  TypeKind
}

func (t *Type) is_primitive(kind string) bool {
	if t.Kind != nil {
		return false
	}
	return t.Token.Id == lex.ID_DT && t.Token.Kind == kind
}
func (t *Type) IsI8() bool { return t.is_primitive(lex.KND_I8) }
func (t *Type) IsI16() bool { return t.is_primitive(lex.KND_I16) }
func (t *Type) IsI32() bool { return t.is_primitive(lex.KND_I32) }
func (t *Type) IsI64() bool { return t.is_primitive(lex.KND_I64) }
func (t *Type) IsU8() bool { return t.is_primitive(lex.KND_U8) }
func (t *Type) IsU16() bool { return t.is_primitive(lex.KND_U16) }
func (t *Type) IsU32() bool { return t.is_primitive(lex.KND_U32) }
func (t *Type) IsU64() bool { return t.is_primitive(lex.KND_U64) }
func (t *Type) IsF32() bool { return t.is_primitive(lex.KND_F32) }
func (t *Type) IsF64() bool { return t.is_primitive(lex.KND_F64) }
func (t *Type) IsInt() bool { return t.is_primitive(lex.KND_INT) }
func (t *Type) IsUint() bool { return t.is_primitive(lex.KND_UINT) }
func (t *Type) IsUintptr() bool { return t.is_primitive(lex.KND_UINTPTR) }
func (t *Type) IsBool() bool { return t.is_primitive(lex.KND_BOOL) }
func (t *Type) IsStr() bool { return t.is_primitive(lex.KND_STR) }
func (t *Type) IsAny() bool { return t.is_primitive(lex.KND_ANY) }
func (t *Type) IsVoid() bool { return t.Kind == nil && t.Token.Id == lex.ID_NA }

type IdentType struct {
	Ident string
}

type RefType struct {
	Elem *Type
}

type MultiRetType struct {
	Types []*Type
}

// Returns type kind as text.
// Returns empty string kind is nil.
func (t *Type) As_text() string {
	if t.Kind == nil {
		return ""
	}
	return t.Kind.As_text()
}
func (itk *IdentType) As_text() string { return itk.Ident }
func (rtk *RefType) As_text() string {
	if rtk.Elem == nil {
		return ""
	}
	return rtk.Elem.As_text()
}
func (mtk *MultiRetType) As_text() string {
	kind := "("
	i := 0
	for ; i < len(mtk.Types); i++ {
		kind += mtk.Types[i].As_text()
		if i+1 < len(mtk.Types) {
			kind += ","
		}
	}
	kind += ")"
	return kind
}

// Return type AST model.
type RetType struct {
	Kind   *Type
	Idents []lex.Token
}

type ExprData = any // Type of AST Expr's data.

// Expression AST.
type Expr struct {
	Token lex.Token
	Kind  ExprData
}

// Function call expression kind.
type FnCallExpr struct {
	Token    lex.Token
	Expr     *Expr
	Generics []*Type
	Args     []*Expr
	IsCo     bool
}

// Generic type AST.
type Generic struct {
	Token lex.Token
	Ident string
}

// Label statement AST.
type LabelSt struct {
	Token lex.Token
	Ident string
}

// Goto statement AST.
type GotoSt struct {
	Token lex.Token
	Label lex.Token
}

// Fall statement AST.
type FallSt struct {
	Token lex.Token
}

// Left expression of assign statement.
type AssignLeft struct {
	Token lex.Token
	IsMut bool
	Ident string
	Expr  *Expr
}

// Assign statement.
type AssignSt struct {
	Setter lex.Token
	L      []*AssignLeft
	R      *Expr
}

// Scope AST.
type Scope struct {
	Parent     *Scope // nil if scope is root
	IsUnsafe   bool
	IsDeferred bool
	Tree       []NodeData
}

// Param AST.
type Param struct {
	Token      lex.Token
	IsMut      bool
	IsVariadic bool
	DataType   *Type
	Ident      string
}

// Function declaration AST.
type FnDecl struct {
	Token       lex.Token
	IsUnsafe    bool
	IsPub       bool
	Ident       string
	Directives  []*Directive
	DocComments *CommentGroup
	Scope       *Scope
	Generics    []*Generic
	RetType     *RetType
	Params      []*Param
}

// Variable declaration AST.
type VarDecl struct {
	Scope       *Scope    // nil for global scopes
	Token       lex.Token
	Ident       string
	IsPub       bool
	IsMut       bool
	IsConst     bool
	DocComments *CommentGroup
	DataType    *Type
	Expr        *Expr
}

// Return statement AST.
type RetSt struct {
	Token lex.Token
	Expr  *Expr
}

type IterKind = any // Type of AST Iter's kind.

// Iteration.
type Iter struct {
	Token lex.Token
	Kind  IterKind
	Scope *Scope
}

// While iteration kind.
type WhileKind struct {
	Expr *Expr
}

// Range iteration kind.
type RangeKind struct {
	InToken lex.Token // Token of "in" keyword
	Expr    *Expr
	KeyA    *VarDecl  // first key of range
	KeyB    *VarDecl  // second key of range
}

// While-next iteration kind.
type WhileNextKind struct {
	Expr *Expr
	Next NodeData
}

// Break statement AST.
type BreakSt struct {
	Token lex.Token
	Label lex.Token
}

// Continue statement AST.
type ContSt struct {
	Token lex.Token
	Label lex.Token
}

// If condition.
type If struct {
	Token lex.Token
	Expr  *Expr
	Scope *Scope
}

// Else condition.
type Else struct {
	Token lex.Token
	Scope *Scope
}

// Condition chain.
type Conditional struct {
	If      *If
	Elifs   []*If
	Default *Else
}

// Type alias declration AST.
type TypeAliasDecl struct {
	IsPub       bool
	Token       lex.Token
	Ident       string
	Kind        *Type
	DocComments *CommentGroup
}

// Case of match-case.
type Case struct {
	Token lex.Token
	// Holds expression.
	// Expressions holds *Type if If type matching.
	Exprs []*Expr
	Scope *Scope
}

// Match-Case AST.
type MatchCase struct {
	Token     lex.Token
	TypeMatch bool
	Expr      *Expr
	Cases     []*Case
	Default   *Else
}
