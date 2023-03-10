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
	return strings.HasPrefix(c.Text, lex.DIRECTIVE_PREFIX)
}

// Directive AST.
type Directive struct {
	Token lex.Token
	Tag   string
}

// Kind type of data types.
type TypeKind = any

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
// Reports whether type is primitive i8.
func (t *Type) IsI8() bool { return t.is_primitive(lex.KND_I8) }
// Reports whether type is primitive i16.
func (t *Type) IsI16() bool { return t.is_primitive(lex.KND_I16) }
// Reports whether type is primitive i32.
func (t *Type) IsI32() bool { return t.is_primitive(lex.KND_I32) }
// Reports whether type is primitive i64.
func (t *Type) IsI64() bool { return t.is_primitive(lex.KND_I64) }
// Reports whether type is primitive u8.
func (t *Type) IsU8() bool { return t.is_primitive(lex.KND_U8) }
// Reports whether type is primitive u16.
func (t *Type) IsU16() bool { return t.is_primitive(lex.KND_U16) }
// Reports whether type is primitive u32.
func (t *Type) IsU32() bool { return t.is_primitive(lex.KND_U32) }
// Reports whether type is primitive u64.
func (t *Type) IsU64() bool { return t.is_primitive(lex.KND_U64) }
// Reports whether type is primitive f32.
func (t *Type) IsF32() bool { return t.is_primitive(lex.KND_F32) }
// Reports whether type is primitive f64.
func (t *Type) IsF64() bool { return t.is_primitive(lex.KND_F64) }
// Reports whether type is primitive int.
func (t *Type) IsInt() bool { return t.is_primitive(lex.KND_INT) }
// Reports whether type is primitive uint.
func (t *Type) IsUint() bool { return t.is_primitive(lex.KND_UINT) }
// Reports whether type is primitive uintptr.
func (t *Type) IsUintptr() bool { return t.is_primitive(lex.KND_UINTPTR) }
// Reports whether type is primitive bool.
func (t *Type) IsBool() bool { return t.is_primitive(lex.KND_BOOL) }
// Reports whether type is primitive str.
func (t *Type) IsStr() bool { return t.is_primitive(lex.KND_STR) }
// Reports whether type is primitive any.
func (t *Type) IsAny() bool { return t.is_primitive(lex.KND_ANY) }
// Reports whether type is void.
func (t *Type) IsVoid() bool { return t.Kind == nil && t.Token.Id == lex.ID_NA }
// Reports whether type is reference.
func (t *Type) IsRef() bool {
	if t.Kind == nil {
		return true
	}
	switch t.Kind.(type) {
	case *RefType:
		return true
	default:
		return false
	}
}
// Reports whether type is pointer.
func (t *Type) IsPtr() bool {
	if t.Kind == nil {
		return true
	}
	switch t.Kind.(type) {
	case *PtrType:
		return true
	default:
		return false
	}
}
// Reports whether type is slice.
func (t *Type) IsSlice() bool {
	if t.Kind == nil {
		return true
	}
	switch t.Kind.(type) {
	case *SliceType:
		return true
	default:
		return false
	}
}
// Reports whether type is array.
func (t *Type) IsArray() bool {
	if t.Kind == nil {
		return true
	}
	switch t.Kind.(type) {
	case *ArrayType:
		return true
	default:
		return false
	}
}

// Identifier type.
type IdentType struct {
	Ident     string
	CppLinked bool
	Generics  []*Type
}

// Namespace chain type.
type NamespaceType struct {
	Idents []string   // Namespace chain.
	Kind   *IdentType // Type of identifier.
}

type RefType struct { Elem *Type }   // Reference type.
type PtrType struct { Elem *Type }   // Pointer type.
type SliceType struct { Elem *Type } // Slice type.
type TupleType struct { Types []*Type } // Tuple type.
type FnType struct { Decl *FnDecl }     // Function type.

// Reports whether pointer is unsafe pointer (*unsafe).
func (pt *PtrType) IsUnsafe() bool { return pt.Elem == nil }

// Array type.
type ArrayType struct {
	Elem *Type
	Size *Expr
}

// Map type.
type MapType struct {
	Key *Type
	Val *Type
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

// Reports whether expression kind is function call.
func (e *Expr) IsFnCall() bool {
	if e.Kind == nil {
		return false
	}
	switch e.Kind.(type) {
	case *FnCallExpr:
		return true
	default:
		return false
	}
}

// Tuple expression.
type TupleExpr struct {
	Expr []ExprData
}

// Literal expression.
type LitExpr struct {
	Token lex.Token
	Value string
}

// Reports whether literal is nil value.
func (le *LitExpr) IsNil() bool { return le.Value == lex.KND_NIL }

// Identifier expression.
type IdentExpr struct {
	Token     lex.Token
	Ident     string
	CppLinked bool
}

// Reports whether identifier is self keyword.
func (ie *IdentExpr) IsSelf() bool { return ie.Ident == lex.KND_SELF }

// Unary expression.
type UnaryExpr struct {
	Op   lex.Token
	Expr ExprData
}

// Primitive type expression.
type TypeExpr struct {
	Token lex.Token // Token of type.
}

// Variadiced expression.
type VariadicExpr struct {
	Token lex.Token
	Expr  ExprData
}

// Casting expression.
type CastExpr struct {
	Kind *Type
	Expr ExprData
}

// Namespace identifier selection expression.
type NsSelectionExpr struct {
	Ns    []lex.Token // Tokens of selected namespace identifier chain.
	Ident lex.Token   // Token of selected identifier.
}

// Object sub identifier selection expression.
type SubIdentExpr struct {
	Expr  ExprData  // Selected object.
	Ident lex.Token // TOken of selected identifier.
}

// Binary operation.
type BinopExpr struct {
	L  ExprData
	R  ExprData
	Op lex.Token
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
	Kind       *Type
	Ident      string
}

// Reports whether parameter is self (receiver) parameter.
func (p *Param) IsSelf() bool { return strings.HasSuffix(p.Ident, lex.KND_SELF) }
// Reports whether self (receiver) parameter is reference.
func (p *Param) IsRef() bool { return p.Ident != "" && p.Ident[0] == '&'}

// Function declaration AST.
type FnDecl struct {
	Token       lex.Token
	IsUnsafe    bool
	IsPub       bool
	CppLinked   bool
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
	CppLinked   bool
	IsPub       bool
	IsMut       bool
	IsConst     bool
	DocComments *CommentGroup
	Kind        *Type
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
	CppLinked   bool
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

// Use declaration statement AST.
type UseDecl struct {
	Token      lex.Token
	LinkString string      // Use declaration path string
	FullUse    bool
	Selected   []lex.Token
	Cpp        bool
}

// Enum item.
type EnumItem struct {
	Token lex.Token
	Ident string
	Expr *Expr
}

// Enum declaration AST.
type EnumDecl struct {
	Token       lex.Token
	IsPub       bool
	Ident       string
	Kind        *Type
	Items       []*EnumItem
	DocComments *CommentGroup
}

// Field AST.
type Field struct {
	Token       lex.Token
	IsPub       bool
	InteriorMut bool
	Ident       string
	Kind        *Type
}

// Structure declaration AST.
type StructDecl struct {
	Token       lex.Token
	Ident       string
	Fields      []*Field
	IsPub       bool
	CppLinked   bool
	Directives  []*Directive
	DocComments *CommentGroup
	Generics    []*Generic
}

// Trait declaration AST.
type TraitDecl struct {
	Token       lex.Token
	Ident       string
	IsPub       bool
	DocComments *CommentGroup
	Methods     []*FnDecl
}

// Implementation AST.
type Impl struct {
	Base    lex.Token
	Dest    lex.Token
	Methods []*FnDecl
}

// Reports whether implementation type is trait to structure.
func (i *Impl) IsTraitImpl() bool { return i.Dest.Id != lex.ID_NA }
// Reports whether implementation type is append to destination structure.
func (i *Impl) IsStructImpl() bool { return i.Dest.Id == lex.ID_NA }
