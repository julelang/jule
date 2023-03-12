// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Type alias.
type TypeAlias struct {
	Public     bool
	Cpp_linked bool
	Token      lex.Token
	Ident      string
	Kind       *Type
	Doc        string
	Refers     []*ast.IdentType
}

// Type's kind's type.
type TypeKind struct {
	Kind any
}

// Type.
type Type struct {
	Decl *ast.TypeDecl // Never changed by semantic analyzer.
	Kind *TypeKind
}

// Reports whether type is checked already.
func (t *Type) checked() bool { return t.Kind != nil }
// Removes kind and ready to check.
// checked() reports false after this function.
func (t *Type) remove_kind() { t.Kind = nil }

// Primitive type.
type Prim struct { kind string }
// Reports whether type is primitive i8.
func (p *Prim) Is_i8() bool { return p.kind == lex.KND_I8 }
// Reports whether type is primitive i16.
func (p *Prim) Is_i16() bool { return p.kind == lex.KND_I16 }
// Reports whether type is primitive i32.
func (p *Prim) Is_i32() bool { return p.kind == lex.KND_I32 }
// Reports whether type is primitive i64.
func (p *Prim) Is_i64() bool { return p.kind == lex.KND_I64 }
// Reports whether type is primitive u8.
func (p *Prim) Is_u8() bool { return p.kind == lex.KND_U8 }
// Reports whether type is primitive u16.
func (p *Prim) Is_u16() bool { return p.kind == lex.KND_U16 }
// Reports whether type is primitive u32.
func (p *Prim) Is_u32() bool { return p.kind == lex.KND_U32 }
// Reports whether type is primitive u64.
func (p *Prim) Is_u64() bool { return p.kind == lex.KND_U64 }
// Reports whether type is primitive f32.
func (p *Prim) Is_f32() bool { return p.kind == lex.KND_F32 }
// Reports whether type is primitive f64.
func (p *Prim) Is_f64() bool { return p.kind == lex.KND_F64 }
// Reports whether type is primitive int.
func (p *Prim) Is_int() bool { return p.kind == lex.KND_INT }
// Reports whether type is primitive uint.
func (p *Prim) Is_uint() bool { return p.kind == lex.KND_UINT }
// Reports whether type is primitive uintptr.
func (p *Prim) Is_uintptr() bool { return p.kind == lex.KND_UINTPTR }
// Reports whether type is primitive bool.
func (p *Prim) Is_bool() bool { return p.kind == lex.KND_BOOL }
// Reports whether type is primitive str.
func (p *Prim) Is_str() bool { return p.kind == lex.KND_STR }
// Reports whether type is primitive any.
func (p *Prim) Is_any() bool { return p.kind == lex.KND_ANY }

// Reference type.
type Ref struct { Elem *TypeKind }

// Checks type and builds result as kind.
// Removes kind if error occurs,
// so type is not reports true for checked state.
type _TypeChecker struct {
	// Uses Sema for:
	//  - Push errors.
	s *_Sema

	// Uses Lookup for:
	//  - Lookup symbol tables.
	lookup _Lookup

	error_token lex.Token
}

func (tc *_TypeChecker) push_err(token lex.Token, key string, args ...any) {
	tc.s.push_err(token, key, args...)
}

func (tc *_TypeChecker) build_prim(kind *ast.IdentType) *Prim {
	switch kind.Ident {
	case lex.KND_I8,
		lex.KND_I16,
		lex.KND_I32,
		lex.KND_I64,
		lex.KND_U8,
		lex.KND_U16,
		lex.KND_U32,
		lex.KND_U64,
		lex.KND_F32,
		lex.KND_F64,
		lex.KND_INT,
		lex.KND_UINT,
		lex.KND_UINTPTR,
		lex.KND_BOOL,
		lex.KND_STR,
		lex.KND_ANY:
		// Ignore.
	default:
		tc.push_err(tc.error_token, "invalid_type")
		return nil 
	}

	if len(kind.Generics) > 0 {
		tc.push_err(kind.Token, "type_not_supports_generics", kind.Ident)
	}

	return &Prim{
		kind: kind.Ident,
	}
}

func (tc *_TypeChecker) build_ident_kind(kind *ast.IdentType) *TypeKind {
	if kind.IsPrim() {
		return &TypeKind{Kind: tc.build_prim(kind)}
	}
	return nil
}

func (tc *_TypeChecker) build_ref(kind *ast.RefType) *Ref {
	elem := tc.check_decl(kind.Elem)
	if elem == nil {
		return nil
	}

	// TODO: check cases:
	//         - ref_refs_ref
	//         - ref_refs_ptr
	//         - ref_refs_array
	//         - ref_refs_enum

	return &Ref{
		Elem: elem,
	}
}

func (tc *_TypeChecker) build_kind(kind ast.TypeDeclKind) *TypeKind {
	switch kind.(type) {
	case *ast.IdentType:
		return tc.build_ident_kind(kind.(*ast.IdentType))

	case *ast.RefType:
		return &TypeKind{Kind: tc.build_ref(kind.(*ast.RefType))}

	default:
		tc.push_err(tc.error_token, "invalid_type")
		return nil
	}
}

func (tc *_TypeChecker) check_decl(decl *ast.TypeDecl) *TypeKind {
	// Save current token.
	error_token := tc.error_token

	tc.error_token = decl.Token
	kind := tc.build_kind(decl.Kind)
	tc.error_token = error_token
	return kind
}

func (tc *_TypeChecker) check(t *Type) {
	// TODO: Detect cycles.
	// TODO: Check type validity.
	kind := tc.check_decl(t.Decl)
	if kind == nil {
		t.remove_kind()
		return
	}
	t.Kind = kind
}
