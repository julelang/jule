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
type TypeKind = any

// Type.
type Type struct {
	Decl *ast.TypeDecl // Never changed by semantic analyzer.
	Kind TypeKind
}

// Reports whether type is checked already.
func (t *Type) checked() bool { return t.Kind != nil }
// Removes kind and ready to check.
// checked() reports false after this function.
func (t *Type) remove_kind() { t.Kind = nil }

// Primitive type.
type PrimType struct { kind string }
// Reports whether type is primitive i8.
func (pt *PrimType) Is_i8() bool { return pt.kind == lex.KND_I8 }
// Reports whether type is primitive i16.
func (pt *PrimType) Is_i16() bool { return pt.kind == lex.KND_I16 }
// Reports whether type is primitive i32.
func (pt *PrimType) Is_i32() bool { return pt.kind == lex.KND_I32 }
// Reports whether type is primitive i64.
func (pt *PrimType) Is_i64() bool { return pt.kind == lex.KND_I64 }
// Reports whether type is primitive u8.
func (pt *PrimType) Is_u8() bool { return pt.kind == lex.KND_U8 }
// Reports whether type is primitive u16.
func (pt *PrimType) Is_u16() bool { return pt.kind == lex.KND_U16 }
// Reports whether type is primitive u32.
func (pt *PrimType) Is_u32() bool { return pt.kind == lex.KND_U32 }
// Reports whether type is primitive u64.
func (pt *PrimType) Is_u64() bool { return pt.kind == lex.KND_U64 }
// Reports whether type is primitive f32.
func (pt *PrimType) Is_f32() bool { return pt.kind == lex.KND_F32 }
// Reports whether type is primitive f64.
func (pt *PrimType) Is_f64() bool { return pt.kind == lex.KND_F64 }
// Reports whether type is primitive int.
func (pt *PrimType) Is_int() bool { return pt.kind == lex.KND_INT }
// Reports whether type is primitive uint.
func (pt *PrimType) Is_uint() bool { return pt.kind == lex.KND_UINT }
// Reports whether type is primitive uintptr.
func (pt *PrimType) Is_uintptr() bool { return pt.kind == lex.KND_UINTPTR }
// Reports whether type is primitive bool.
func (pt *PrimType) Is_bool() bool { return pt.kind == lex.KND_BOOL }
// Reports whether type is primitive str.
func (pt *PrimType) Is_str() bool { return pt.kind == lex.KND_STR }
// Reports whether type is primitive any.
func (pt *PrimType) Is_any() bool { return pt.kind == lex.KND_ANY }

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

func (tc *_TypeChecker) build_prim(kind *ast.IdentType) *PrimType {
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

	return &PrimType{
		kind: kind.Ident,
	}
}

func (tc *_TypeChecker) build_ident_kind(kind *ast.IdentType) TypeKind {
	if kind.IsPrim() {
		return tc.build_prim(kind)
	}
	return nil
}

func (tc *_TypeChecker) build_kind(kind ast.TypeDeclKind) TypeKind {
	switch kind.(type) {
	case *ast.IdentType:
		return tc.build_ident_kind(kind.(*ast.IdentType))

	default:
		tc.push_err(tc.error_token, "invalid_type")
		return nil
	}
}

func (tc *_TypeChecker) check(t *Type) {
	tc.error_token = t.Decl.Token

	// TODO: Detect cycles.
	// TODO: Check type validity.
	kind := tc.build_kind(t.Decl.Kind)
	if kind == nil {
		t.remove_kind()
		return
	}
	t.Kind = kind
}
