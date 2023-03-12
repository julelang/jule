// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package ast

import (
	"strings"

	"github.com/julelang/jule/lex"
)

// Type of AST Node's data.
type NodeData = any

// AST Node.
type Node struct {
	Token lex.Token
	Data  any
}

// Reports whether node data is declaration.
func (n *Node) Is_decl() bool {
	switch n.Data.(type) {
		case *EnumDecl,
			*FnDecl,
			*StructDecl,
			*TraitDecl,
			*TypeAliasDecl,
			*FieldDecl,
			*UseDecl,
			*VarDecl,
			*TypeDecl:
		return true

	default:
		return false
	}
}
// Reports whether node data is comment or comment group.
func (n *Node) Is_comment() bool {
	switch n.Data.(type) {
		case *Comment, *CommentGroup:
		return true

	default:
		return false
	}
}
// Reports whether node data is impl.
func (n *Node) Is_impl() bool {
	switch n.Data.(type) {
	case *Impl:
		return true

	default:
		return false
	}
}
// Reports whether node data is use declaration.
func (n *Node) Is_use_decl() bool {
	switch n.Data.(type) {
	case *UseDecl:
		return true

	default:
		return false
	}
}

// Comment group.
type CommentGroup struct {
	Comments []*Comment
}

// Comment line.
type Comment struct {
	Token lex.Token
	Text  string
}

// Reports whether comment is directive.
func (c *Comment) Is_directive() bool {
	return strings.HasPrefix(c.Text, lex.DIRECTIVE_PREFIX)
}

// Directive.
type Directive struct {
	Token lex.Token
	Tag   string
}

// Kind type of type declarations.
type TypeDeclKind = any

// TypeDecl declaration.
// Also represents type expression.
//
// For primitive types:
//  - Represented by IdentType.
//  - Token's identity is data type.
//  - Primitive type kind is Ident.
//
// For function types:
//  - Function types represented by *FnDecl.
type TypeDecl struct {
	Token lex.Token
	Kind  TypeDeclKind
}

// Identifier type.
type IdentType struct {
	Token      lex.Token
	Ident      string
	Cpp_linked bool
	Generics   []*TypeDecl
}

// Reports whether identifier is primitive type.
func (it *IdentType) Is_prim() bool { return it.Token.Id == lex.ID_PRIM }

// Namespace chain type.
type NamespaceType struct {
	Idents []string   // Namespace chain.
	Kind   *IdentType // Type of identifier.
}

type RefType struct { Elem *TypeDecl }      // Reference type.
type PtrType struct { Elem *TypeDecl }      // Pointer type.
type SlcType struct { Elem *TypeDecl }      // Slice type.
type TupleType struct { Types []*TypeDecl } // Tuple type.

// Reports whether pointer is unsafe pointer (*unsafe).
func (pt *PtrType) Is_unsafe() bool { return pt.Elem == nil }

// Array type.
type ArrayType struct {
	Elem *TypeDecl
	Size *Expr
}

// Map type.
type MapType struct {
	Key *TypeDecl
	Val *TypeDecl
}

// Return type.
type RetType struct {
	Kind   *TypeDecl
	Idents []lex.Token
}

// Reports whether return type is void.
func (rt *RetType) Is_void() bool { return rt.Kind == nil }

type ExprData = any // Type of Expr's data.

// Expression.
type Expr struct {
	Token lex.Token
	Kind  ExprData
}

// Reports whether expression kind is function call.
func (e *Expr) Is_fn_call() bool {
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

// Unsafe expression.
type UnsafeExpr struct {
	Token lex.Token // Token of unsafe keyword.
	Expr  ExprData
}

// Reports whether literal is nil value.
func (le *LitExpr) Is_nil() bool { return le.Value == lex.KND_NIL }

// Identifier expression.
type IdentExpr struct {
	Token      lex.Token
	Ident      string
	Cpp_linked bool
}

// Reports whether identifier is self keyword.
func (ie *IdentExpr) Is_self() bool { return ie.Ident == lex.KND_SELF }

// Unary expression.
type UnaryExpr struct {
	Op   lex.Token
	Expr ExprData
}

// Variadiced expression.
type VariadicExpr struct {
	Token lex.Token
	Expr  ExprData
}

// Casting expression.
type CastExpr struct {
	Kind *TypeDecl
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
	Token      lex.Token
	Expr       *Expr
	Generics   []*TypeDecl
	Args       []*Expr
	Concurrent bool
}

// Field-Expression pair.
type FieldExprPair struct {
	Field lex.Token // Field identifier token.
	Expr  ExprData
}

// Reports whether pair targeted field.
func (fep *FieldExprPair) Is_targeted() bool { return fep.Field.Id != lex.ID_NA }

// Struct literal instiating expression.
type StructLit struct {
	Kind  *TypeDecl
	Pairs []*FieldExprPair
}

// Anonymous brace instiating expression.
// Empty braces ( {} ).
type BraceLit struct {
	Exprs []ExprData
}

// Reports literal is empty ( {} ).
func (bl *BraceLit) Is_empty() bool { return len(bl.Exprs) == 0 }

// Key-value pair expression.
type KeyValPair struct {
	Key   ExprData
	Val   ExprData
	Colon lex.Token
}

// Slice initiating expression.
// Also represents array initiating expression.
type SliceExpr struct {
	Elems []ExprData
}

// Reports whether slice is empty.
func (se *SliceExpr) Is_empty() bool { return len(se.Elems) == 0 }

// Indexing expression.
type IndexingExpr struct {
	Expr  ExprData // Value expression to indexing.
	Index ExprData // Index value expression.
}

// Slicing expression.
type SlicingExpr struct {
	Expr  ExprData // Value expression to slicing.
	Start ExprData // Start index value expression.
	To    ExprData // To index value expression.
}

// Generic type.
type Generic struct {
	Token lex.Token
	Ident string
}

// Label statement.
type LabelSt struct {
	Token lex.Token
	Ident string
}

// Goto statement.
type GotoSt struct {
	Token lex.Token
	Label lex.Token
}

// Fall statement.
type FallSt struct {
	Token lex.Token
}

// Left expression of assign statement.
type AssignLeft struct {
	Token   lex.Token
	Mutable bool
	Ident   string
	Expr    *Expr
}

// Assign statement.
type AssignSt struct {
	Setter lex.Token
	L      []*AssignLeft
	R      *Expr
}

// Scope.
type Scope struct {
	Parent   *Scope     // Nil if scope is root.
	Unsafety bool
	Deferred bool
	Stmts    []NodeData // Statements.
}

// Param.
type Param struct {
	Token    lex.Token
	Mutable  bool
	Variadic bool
	Kind     *TypeDecl
	Ident    string
}

// Reports whether parameter is self (receiver) parameter.
func (p *Param) Is_self() bool { return strings.HasSuffix(p.Ident, lex.KND_SELF) }
// Reports whether self (receiver) parameter is reference.
func (p *Param) Is_ref() bool { return p.Ident != "" && p.Ident[0] == '&'}

// Function declaration.
// Also represents anonymous function expression.
type FnDecl struct {
	Token        lex.Token
	Unsafety     bool
	Public       bool
	Cpp_linked   bool
	Ident        string
	Directives   []*Directive
	Doc_comments *CommentGroup
	Scope        *Scope
	Generics     []*Generic
	Result       *RetType
	Params       []*Param
}

// Variable declaration.
type VarDecl struct {
	Scope        *Scope    // nil for global scopes
	Token        lex.Token
	Ident        string
	Cpp_linked   bool
	Public       bool
	Mutable      bool
	Constant     bool
	Doc_comments *CommentGroup
	Kind         *TypeDecl
	Expr         *Expr
}

// Return statement.
type RetSt struct {
	Token lex.Token
	Expr  *Expr
}

type IterKind = any // Type of Iter's kind.

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
	In_token lex.Token // Token of "in" keyword
	Expr     *Expr
	Key_a    *VarDecl  // first key of range
	Key_b    *VarDecl  // second key of range
}

// While-next iteration kind.
type WhileNextKind struct {
	Expr *Expr
	Next NodeData
}

// Break statement.
type BreakSt struct {
	Token lex.Token
	Label lex.Token
}

// Continue statement.
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

// Type alias declration.
type TypeAliasDecl struct {
	Public       bool
	Cpp_linked   bool
	Token        lex.Token
	Ident        string
	Kind         *TypeDecl
	Doc_comments *CommentGroup
}

// Case of match-case.
type Case struct {
	Token lex.Token
	Scope *Scope

	// Holds expression.
	// Expressions holds *Type if If type matching.
	Exprs []*Expr
}

// Match-Case.
type MatchCase struct {
	Token      lex.Token
	Type_match bool
	Expr       *Expr
	Cases      []*Case
	Default    *Else
}

// Use declaration statement.
type UseDecl struct {
	Token     lex.Token
	Link_path string      // Use declaration path string.
	Full      bool        // Full implicit import.
	Selected  []lex.Token
	Cpp       bool        // Cpp header use declaration.
	Std       bool        // Standard package use declaration.
}

// Enum item.
type EnumItem struct {
	Token lex.Token
	Ident string
	Expr *Expr
}

// Enum declaration.
type EnumDecl struct {
	Token        lex.Token
	Public       bool
	Ident        string
	Kind         *TypeDecl
	Items        []*EnumItem
	Doc_comments *CommentGroup
}

// Reports enum's type is default.
func (ed *EnumDecl) Default_typed() bool { return ed.Kind == nil }

// Field declaration.
type FieldDecl struct {
	Token   lex.Token
	Public  bool
	Mutable bool       // Interior mutability.
	Ident   string
	Kind    *TypeDecl
}

// Structure declaration.
type StructDecl struct {
	Token        lex.Token
	Ident        string
	Fields       []*FieldDecl
	Public       bool
	Cpp_linked   bool
	Directives   []*Directive
	Doc_comments *CommentGroup
	Generics     []*Generic
}

// Trait declaration.
type TraitDecl struct {
	Token        lex.Token
	Ident        string
	Public       bool
	Doc_comments *CommentGroup
	Methods      []*FnDecl
}

// Implementation.
type Impl struct {
	Base    lex.Token
	Dest    lex.Token
	Methods []*FnDecl
}

// Reports whether implementation type is trait to structure.
func (i *Impl) Is_trait_impl() bool { return i.Dest.Id != lex.ID_NA }
// Reports whether implementation type is append to destination structure.
func (i *Impl) Is_struct_impl() bool { return i.Dest.Id == lex.ID_NA }
