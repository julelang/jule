// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// This file reserved for types, type kinds and type build algorithms.
// This file haven't type compatibility checking algorithm or something else.

package sema

import (
	"strconv"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

// Type alias.
type TypeAlias struct {
	Scope      *ast.ScopeTree
	Public     bool
	Cpp_linked bool
	Used       bool
	Token      lex.Token
	Ident      string
	Kind       *TypeSymbol
	Doc        string
	Refers     []uintptr // Addresses of referred identifiers.
}

type _Kind interface {
	To_str() string
}

// Type's kind's type.
type TypeKind struct {
	Cpp_linked bool
	Cpp_ident  string
	kind      _Kind
}
// Returns clone.
func (tk *TypeKind) clone() *TypeKind {
	kind := new(TypeKind)
	kind.Cpp_ident = tk.Cpp_ident
	kind.Cpp_linked = tk.Cpp_linked
	kind.kind = tk.kind
	return kind
}
// Returns kind as string.
func (tk TypeKind) To_str() string {
	if tk.Is_nil() {
		return "nil"
	}
	return tk.kind.To_str()
}
// Reports whether kind is "nil".
func (tk *TypeKind) Is_nil() bool { return tk.kind == nil }
// Reports whether kind is "void".
func (tk *TypeKind) Is_void() bool {
	prim := tk.Prim()
	return prim != nil && prim.To_str() == "void"
}
// Returns primitive type if kind is primitive type, nil if not.
func (tk *TypeKind) Prim() *Prim {
	switch tk.kind.(type) {
	case *Prim:
		return tk.kind.(*Prim)

	default:
		return nil
	}
}
// Returns reference type if kind is reference, nil if not.
func (tk *TypeKind) Ref() *Ref {
	switch tk.kind.(type) {
	case *Ref:
		return tk.kind.(*Ref)

	default:
		return nil
	}
}
// Returns pointer type if kind is pointer, nil if not.
func (tk *TypeKind) Ptr() *Ptr {
	switch tk.kind.(type) {
	case *Ptr:
		return tk.kind.(*Ptr)

	default:
		return nil
	}
}
// Returns enum type if kind is enum, nil if not.
func (tk *TypeKind) Enm() *Enum {
	switch tk.kind.(type) {
	case *Enum:
		return tk.kind.(*Enum)

	default:
		return nil
	}
}
// Returns array type if kind is array, nil if not.
func (tk *TypeKind) Arr() *Arr {
	switch tk.kind.(type) {
	case *Arr:
		return tk.kind.(*Arr)

	default:
		return nil
	}
}
// Returns slice type if kind is slice, nil if not.
func (tk *TypeKind) Slc() *Slc {
	switch tk.kind.(type) {
	case *Slc:
		return tk.kind.(*Slc)

	default:
		return nil
	}
}
// Returns fn type if kind is function, nil if not.
func (tk *TypeKind) Fnc() *FnIns {
	switch tk.kind.(type) {
	case *FnIns:
		return tk.kind.(*FnIns)

	default:
		return nil
	}
}
// Returns struct type if kind is structure, nil if not.
func (tk *TypeKind) Strct() *StructIns {
	switch tk.kind.(type) {
	case *StructIns:
		return tk.kind.(*StructIns)

	default:
		return nil
	}
}
// Returns trait type if kind is trait, nil if not.
func (tk *TypeKind) Trt() *Trait {
	switch tk.kind.(type) {
	case *Trait:
		return tk.kind.(*Trait)

	default:
		return nil
	}
}
// Returns map type if kind is map, nil if not.
func (tk *TypeKind) Map() *Map {
	switch tk.kind.(type) {
	case *Map:
		return tk.kind.(*Map)

	default:
		return nil
	}
}
// Returns tuple type if kind is tuple, nil if not.
func (tk *TypeKind) Tup() *Tuple {
	switch tk.kind.(type) {
	case *Tuple:
		return tk.kind.(*Tuple)

	default:
		return nil
	}
}

// Type.
type TypeSymbol struct {
	Decl *ast.Type // Never changed by semantic analyzer.
	Kind *TypeKind
}

// Reports whether type is checked already.
func (ts *TypeSymbol) checked() bool { return ts.Kind != nil }
// Removes kind and ready to check.
// checked() reports false after this function.
func (ts *TypeSymbol) remove_kind() { ts.Kind = nil }

// Primitive type.
type Prim struct { kind string }
// Returns kind.
func (p Prim) To_str() string { return p.kind }
// Reports whether type is primitive i8.
func (p *Prim) Is_i8() bool { return p.kind == types.TypeKind_I8 }
// Reports whether type is primitive i16.
func (p *Prim) Is_i16() bool { return p.kind == types.TypeKind_I16 }
// Reports whether type is primitive i32.
func (p *Prim) Is_i32() bool { return p.kind == types.TypeKind_I32 }
// Reports whether type is primitive i64.
func (p *Prim) Is_i64() bool { return p.kind == types.TypeKind_I64 }
// Reports whether type is primitive u8.
func (p *Prim) Is_u8() bool { return p.kind == types.TypeKind_U8 }
// Reports whether type is primitive u16.
func (p *Prim) Is_u16() bool { return p.kind == types.TypeKind_U16 }
// Reports whether type is primitive u32.
func (p *Prim) Is_u32() bool { return p.kind == types.TypeKind_U32 }
// Reports whether type is primitive u64.
func (p *Prim) Is_u64() bool { return p.kind == types.TypeKind_U64 }
// Reports whether type is primitive f32.
func (p *Prim) Is_f32() bool { return p.kind == types.TypeKind_F32 }
// Reports whether type is primitive f64.
func (p *Prim) Is_f64() bool { return p.kind == types.TypeKind_F64 }
// Reports whether type is primitive int.
func (p *Prim) Is_int() bool { return p.kind == types.TypeKind_INT }
// Reports whether type is primitive uint.
func (p *Prim) Is_uint() bool { return p.kind == types.TypeKind_UINT }
// Reports whether type is primitive uintptr.
func (p *Prim) Is_uintptr() bool { return p.kind == types.TypeKind_UINTPTR }
// Reports whether type is primitive bool.
func (p *Prim) Is_bool() bool { return p.kind == types.TypeKind_BOOL }
// Reports whether type is primitive str.
func (p *Prim) Is_str() bool { return p.kind == types.TypeKind_STR }
// Reports whether type is primitive any.
func (p *Prim) Is_any() bool { return p.kind == types.TypeKind_ANY }

// Reference type.
type Ref struct { Elem *TypeKind }
// Returns reference kind as string.
func (r Ref) To_str() string { return "&" + r.Elem.To_str() }

// Slice type.
type Slc struct { Elem *TypeKind }
// Returns slice kind as string.
func (s Slc) To_str() string { return "[]" + s.Elem.To_str() }

// Tuple type.
type Tuple struct { Types []*TypeKind }
// Returns tuple kind as string.
func (t Tuple) To_str() string {
	s := "("
	s += t.Types[0].To_str()
	for _, t := range t.Types[1:] {
		s += ","
		s += t.To_str()
	}
	s += ")"
	return s
}

// Map type.
type Map struct {
	Key *TypeKind
	Val *TypeKind
}
// Returns map kind as string.
func (m Map) To_str() string {
	s := "["
	s += m.Key.To_str()
	s += ":"
	s += m.Val.To_str()
	s += "]"
	return s
}

// Array type.
type Arr struct {
	Auto bool       // Auto-sized array.
	N    int
	Elem *TypeKind
}
// Returns array kind as string.
func (a Arr) To_str() string {
	s := "["
	s += strconv.Itoa(a.N)
	s += "]"
	s += a.Elem.To_str()
	return s
}

// Pointer type.
type Ptr struct { Elem *TypeKind }
// Returns pointer kind as string.
func (p Ptr) To_str() string {
	if p.Is_unsafe() {
		return "*unsafe"
	}
	return "*" + p.Elem.To_str()
}
// Reports whether pointer is unsafe pointer (*unsafe).
func (p *Ptr) Is_unsafe() bool { return p.Elem == nil }

func can_get_ptr(d *Data) bool {
	if !d.Lvalue || d.Is_const() {
		return false
	}

	switch {
	case d.Kind.Fnc() != nil || d.Kind.Enm() != nil:
		return false

	default:
		return true
	}
}

func is_lvalue(t *TypeKind) bool {
	return t.Ref() != nil || t.Ptr() != nil || t.Slc() != nil || t.Map() != nil
}

func is_mut(t *TypeKind) bool {
	return t.Slc() != nil || t.Ptr() != nil || t.Ref() != nil
}

func is_nil_compatible(t *TypeKind) bool {
	prim := t.Prim()
	if prim != nil && prim.Is_any() {
		return true
	}

	return (t.Is_nil() ||
		t.Fnc() != nil ||
		t.Ptr() != nil ||
		t.Slc() != nil ||
		t.Trt() != nil ||
		t.Map() != nil)
}

func is_valid_for_ref(t *TypeKind) bool {
	return !(t.Enm() != nil || t.Ptr() != nil || t.Ref() != nil|| t.Arr() != nil)
}

func is_variadicable(tk *TypeKind) bool { return tk.Slc() != nil }

func build_link_path_by_tokens(tokens []lex.Token) string {
	s := tokens[0].Kind
	for _, token := range tokens[1:] {
		s += lex.KND_DBLCOLON
		s += token.Kind
	}
	return s
}

func build_prim_type(kind string) *Prim {
	return &Prim{
		kind: kind,
	}
}

// Reports whether kind is primitive type.
func is_prim(kind string) bool {
	return kind == lex.KND_I8 ||
		kind == lex.KND_I16 ||
		kind == lex.KND_I32 ||
		kind == lex.KND_I64 ||
		kind == lex.KND_U8 ||
		kind == lex.KND_U16 ||
		kind == lex.KND_U32 ||
		kind == lex.KND_U64 ||
		kind == lex.KND_F32 ||
		kind == lex.KND_F64 ||
		kind == lex.KND_INT ||
		kind == lex.KND_UINT ||
		kind == lex.KND_UINTPTR ||
		kind == lex.KND_BOOL ||
		kind == lex.KND_STR ||
		kind == lex.KND_ANY
}

type _Referencer struct {
	ident  string
	owner  uintptr
	refers *[]uintptr
	strct  *Struct
}

func (r *_Referencer) is_struct_mode() bool { return r.strct != nil }

// Checks type and builds result as kind.
// Removes kind if error occurs,
// so type is not reports true for checked state.
type _TypeChecker struct {
	// Uses Sema for:
	//  - Push errors.
	s *_Sema

	// Uses Lookup for:
	//  - Lookup symbol tables.
	lookup Lookup

	// If this is not nil, appends referred ident types.
	// Also used as checker owner.
	referencer *_Referencer

	error_token lex.Token

	// This identifiers ignored and
	// appends as primitive type.
	//
	// Each dimension 2 array accepted as identifier group.
	ignore_generics []*ast.Generic

	// Ignores with trait check pattern.
	// Uses to_trait_kind_str's representation.
	ignore_with_trait_pattern bool

	// This generics used as type alias for real kind.
	use_generics []*TypeAlias

	// Current checked type is not plain type.
	// Type is pointer, reference, slice or similar.
	not_plain bool
}

func (tc *_TypeChecker) push_err(token lex.Token, key string, args ...any) {
	tc.s.push_err(token, key, args...)
}

func (tc *_TypeChecker) build_prim(decl *ast.IdentType) *Prim {
	if !is_prim(decl.Ident) {
		tc.push_err(tc.error_token, "invalid_type")
		return nil 
	}

	if len(decl.Generics) > 0 {
		tc.push_err(decl.Token, "type_not_supports_generics", decl.Ident)
		return nil
	}

	return build_prim_type(decl.Ident)
}

// Checks illegal cycles.
// Appends reference to reference if there is no illegal cycle.
// Returns true if tc.referencer is nil.
// Returns true if refers is nil.
func (tc *_TypeChecker) check_illegal_cycles(decl uintptr, refers *[]uintptr, decl_token lex.Token) (ok bool) {
	if tc.referencer == nil || tc.referencer.refers == nil || refers == nil {
		return true
	}

	// Check illegal cycle for itself.
	// Because refers's owner is ta.
	if tc.referencer.refers == refers {
		tc.push_err(decl_token, "illegal_cycle_refers_itself", tc.referencer.ident)
		return false
	}

	// Check cross illegal cycle.
	for _, r := range *refers {
		if r == tc.referencer.owner {
			tc.push_err(decl_token, "illegal_cross_cycle", tc.referencer.ident, decl_token.Kind)
			return false
		}
	}

	*tc.referencer.refers = append(*tc.referencer.refers, decl)
	return true
}

// Checks structure illegal cycles.
// Appends depend to depends if there is no illegal cycle.
// Returns true if tc.referencer is nil.
// Returns true if tc.referencer.is_struct_mode() is false.
// Returns true if tc.not_plain is true.
func (tc *_TypeChecker) check_struct_illegal_cycles(decl *ast.IdentType, s *Struct) (ok bool) {
	switch {
	case tc.referencer == nil:
		return true

	case !tc.referencer.is_struct_mode():
		return true

	case tc.not_plain:
		return true
	}

	// Check illegal cycle for itself.
	// Because refers's owner is ta.
	if tc.referencer.strct == s {
		tc.push_err(decl.Token, "illegal_cycle_refers_itself", tc.referencer.ident)
		return false
	}

	// Check cross illegal cycle.
	for _, d := range s.Depends {
		if d == tc.referencer.strct {
			tc.push_err(decl.Token, "illegal_cross_cycle", tc.referencer.ident, decl.Ident)
			return false
		}
	}

	tc.referencer.strct.Depends = append(tc.referencer.strct.Depends, s)
	return true
}

func (tc *_TypeChecker) from_type_alias(decl *ast.IdentType, ta *TypeAlias) _Kind {
	if !tc.s.is_accessible_define(ta.Public, ta.Token) {
		tc.push_err(decl.Token, "ident_not_exist", decl.Ident)
		return nil
	}

	ta.Used = true

	if len(decl.Generics) > 0 {
		tc.push_err(decl.Token, "type_not_supports_generics", decl.Ident)
		return nil
	}

	ok := tc.check_illegal_cycles(_uintptr(ta), &ta.Refers, decl.Token)
	if !ok {
		return nil
	}

	// Build kind if not builded already.
	ok = tc.s.check_type_alias_decl_kind(ta, tc.lookup)
	if !ok {
		return nil
	}

	kind := ta.Kind.Kind.clone()

	if ta.Cpp_linked {
		kind.Cpp_linked = true
		kind.Cpp_ident = ta.Ident
	}

	return kind
}

func (tc *_TypeChecker) from_enum(decl *ast.IdentType, e *Enum) *Enum {
	if !tc.s.is_accessible_define(e.Public, e.Token) {
		tc.push_err(decl.Token, "ident_not_exist", decl.Ident)
		return nil
	}

	if len(decl.Generics) > 0 {
		tc.push_err(decl.Token, "type_not_supports_generics", decl.Ident)
		return nil
	}

	ok := tc.check_illegal_cycles(_uintptr(e), &e.Refers, decl.Token)
	if !ok {
		return nil
	}

	return e
}

func (tc *_TypeChecker) check_struct_ins(ins *StructIns, error_token lex.Token) (ok bool) {
	ok = tc.s.check_generic_quantity(len(ins.Decl.Generics), len(ins.Generics), error_token)
	if !ok {
		return false
	}

	if tc.referencer != nil && tc.referencer.strct == ins.Decl {
		// Break algorithm cycle.
		return true
	} else if ins.Decl.sema != nil && len(ins.Decl.Generics) == 0 {
		// Break algorithm cycle.
		return true
	}

	referencer := &_Referencer{
		ident: ins.Decl.Ident,
		strct: ins.Decl,
	}

	generics := make([]*TypeAlias, len(ins.Generics))
	for i, g := range ins.Generics {
		generics[i] = &TypeAlias{
			Ident: ins.Decl.Generics[i].Ident,
			Kind:  &TypeSymbol{
				Kind: g,
			},
		}
	}

	// Check field types.
	for _, f := range ins.Fields {
		tc := _TypeChecker{
			s:            tc.s,
			lookup:       tc.s,
			referencer:   referencer,
			use_generics: generics,
		}
		kind := tc.check_decl(f.Decl.Kind.Decl)
		ok := kind != nil

		if ins.Decl.sema != nil && tc.s != ins.Decl.sema && len(ins.Decl.sema.errors) > 0 {
			tc.s.errors = append(tc.s.errors, ins.Decl.sema.errors...)
		}

		if !ok {
			return false
		}

		f.Kind = kind
	}

	return true
}

func (tc *_TypeChecker) from_struct(decl *ast.IdentType, s *Struct) *StructIns {
	if !tc.s.is_accessible_define(s.Public, s.Token) {
		tc.push_err(decl.Token, "ident_not_exist", decl.Ident)
		return nil
	}

	ok := tc.check_struct_illegal_cycles(decl, s)
	if !ok {
		return nil
	}

	ins := s.instance()
	ins.Generics = make([]*TypeKind, len(decl.Generics))
	referencer := tc.referencer
	tc.referencer = nil
	for i, g := range decl.Generics {
		kind := tc.build(g.Kind)
		if kind == nil {
			ok = false
			continue
		}
		ins.Generics[i] = kind
	}

	tc.referencer = referencer

	if !ok {
		return nil
	}

	ok = tc.check_struct_ins(ins, decl.Token)
	if !ok {
		return nil
	}

	s.append_instance(ins)
	return ins
}

func (tc *_TypeChecker) get_def(decl *ast.IdentType) _Kind {
	for i, g := range tc.ignore_generics {
		if g.Ident == decl.Ident {
			if tc.ignore_with_trait_pattern {
				return build_prim_type(strconv.Itoa(i))
			} else {
				return build_prim_type(g.Ident)
			}
		}
	}

	for _, g := range tc.use_generics {
		if g.Ident == decl.Ident {
			st := g.Kind.Kind.Strct()
			if st != nil {
				ok := tc.check_struct_illegal_cycles(decl, st.Decl)
				if !ok {
					return nil
				}
			}
			return g.Kind.Kind.kind
		}
	}

	if !decl.Cpp_linked {
		e := tc.lookup.Find_enum(decl.Ident)
		if e != nil {
			return tc.from_enum(decl, e)
		}

		t := tc.lookup.Find_trait(decl.Ident)
		if t == nil {
			t = find_builtin_trait(decl.Ident)
		}
		if t != nil {
			if !tc.s.is_accessible_define(t.Public, t.Token) {
				tc.push_err(decl.Token, "ident_not_exist", decl.Ident)
				return nil
			}

			if len(decl.Generics) > 0 {
				tc.push_err(decl.Token, "type_not_supports_generics", decl.Ident)
				return nil
			}
			return t
		}
	}

	s := tc.lookup.Find_struct(decl.Ident, decl.Cpp_linked)
	if s != nil {
		return tc.from_struct(decl, s)
	}

	ta := tc.lookup.Find_type_alias(decl.Ident, decl.Cpp_linked)
	if ta == nil {
		ta = find_builtin_type_alias(decl.Ident)
	}
	if ta != nil {
		return tc.from_type_alias(decl, ta)
	}

	tc.push_err(decl.Token, "ident_not_exist", decl.Ident)
	return nil
}

func (tc *_TypeChecker) build_ident(decl *ast.IdentType) _Kind {
	switch {
	case is_prim(decl.Ident):
		return tc.build_prim(decl)
	
	default:
		return tc.get_def(decl)
	}
}

func (tc *_TypeChecker) build_ref(decl *ast.RefType) *Ref {
	elem := tc.check_decl(decl.Elem)

	// Check special cases.
	switch {
	case elem == nil:
		return nil

	case elem.Ref() != nil:
		tc.push_err(tc.error_token, "ref_refs_ref")
		return nil

	case elem.Ptr() != nil:
		tc.push_err(tc.error_token, "ref_refs_ptr")
		return nil

	case elem.Enm() != nil:
		tc.push_err(tc.error_token, "ref_refs_enum")
		return nil

	case elem.Arr() != nil:
		tc.push_err(tc.error_token, "ref_refs_array")
		return nil
	}

	return &Ref{
		Elem: elem,
	}
}

func (tc *_TypeChecker) build_ptr(decl *ast.PtrType) *Ptr {
	not_plain := tc.not_plain
	tc.not_plain = true
	defer func() { tc.not_plain = not_plain }()

	var elem *TypeKind = nil

	if !decl.Is_unsafe() {
		elem = tc.check_decl(decl.Elem)

		// Check special cases.
		switch {
		case elem == nil:
			return nil

		case elem.Ref() != nil:
			tc.push_err(tc.error_token, "ptr_points_ref")
			return nil
	
		case elem.Enm() != nil:
			tc.push_err(tc.error_token, "ptr_points_enum")
			return nil

		case elem.Arr() != nil && elem.Arr().Auto:
			tc.push_err(decl.Elem.Token, "array_auto_sized")
			return nil
		}
	}

	return &Ptr{
		Elem: elem,
	}
}

func (tc *_TypeChecker) build_slc(decl *ast.SlcType) *Slc {
	elem := tc.check_decl(decl.Elem)

	// Check special cases.
	switch {
	case elem == nil:
		return nil
	
	case elem.Arr() != nil && elem.Arr().Auto:
		tc.push_err(decl.Elem.Token, "array_auto_sized")
		return nil
	}

	return &Slc{
		Elem: elem,
	}
}

func (tc *_TypeChecker) build_arr(decl *ast.ArrType) *Arr {
	var n int = 0

	if !decl.Auto_sized() {
		size := tc.s.eval(decl.Size, tc.lookup)
		if size == nil {
			return nil
		}

		if !size.Is_const() {
			tc.push_err(decl.Elem.Token, "expr_not_const")
			return nil
		} else if !types.Is_int(size.Kind.Prim().kind) {
			tc.push_err(decl.Elem.Token, "array_size_is_not_int")
			return nil
		}

		n = int(size.Constant.As_i64())
		if n < 0 {
			tc.push_err(decl.Elem.Token, "array_size_is_negative")
			return nil
		}
	}

	elem := tc.check_decl(decl.Elem)

	// Check special cases.
	switch {
	case elem == nil:
		return nil
	
	case elem.Arr() != nil && elem.Arr().Auto:
		tc.push_err(decl.Elem.Token, "array_auto_sized")
		return nil
	}

	return &Arr{
		Auto: decl.Auto_sized(),
		N:    n,
		Elem: elem,
	}
}

func (tc *_TypeChecker) build_map(decl *ast.MapType) *Map {
	key := tc.check_decl(decl.Key)
	if key == nil {
		return nil
	}

	val := tc.check_decl(decl.Val)
	if val == nil {
		return nil
	}

	return &Map{
		Key: key,
		Val: val,
	}
}

func (tc *_TypeChecker) build_tuple(decl *ast.TupleType) *Tuple {
	types := make([]*TypeKind, len(decl.Types))
	for i, t := range decl.Types {
		kind := tc.check_decl(t)
		if kind == nil {
			return nil
		}
		types[i] = kind
	}

	return &Tuple{Types: types}
}

func (tc *_TypeChecker) check_fn_types(f *FnIns) (ok bool) {
	for _, p := range f.Params {
		p.Kind = tc.build(p.Decl.Kind.Decl.Kind)
		ok = p.Kind != nil
		if !ok {
			return false
		}
	}

	if !f.Decl.Is_void() {
		f.Result = tc.build(f.Decl.Result.Kind.Decl.Kind)
		return f.Result != nil
	}

	return true
}

func (tc *_TypeChecker) build_fn(decl *ast.FnDecl) *FnIns {
	if len(decl.Generics) > 0 {
		tc.push_err(decl.Token, "genericed_fn_as_anonymous_fn")
		return nil
	}

	f := build_fn(decl)
	ins := f.instance_force()

	ok := tc.check_fn_types(ins)
	if !ok {
		return nil
	}

	return ins
}

func (tc *_TypeChecker) build_by_std_namespace(decl *ast.NamespaceType) _Kind {
	path := build_link_path_by_tokens(decl.Idents)
	imp := tc.lookup.Select_package(func(imp *ImportInfo) bool {
		return imp.Std && imp.Link_path == path
	})

	if imp == nil || !imp.is_lookupable(lex.KND_SELF) {
		tc.push_err(decl.Idents[0], "namespace_not_exist", path)
		return nil
	}

	lookup := tc.lookup
	tc.lookup = imp

	kind := tc.build_ident(decl.Kind)

	tc.lookup = lookup

	return kind
}

func (tc *_TypeChecker) build_by_namespace(decl *ast.NamespaceType) _Kind {
	token := decl.Idents[0]
	if token.Kind == "std" {
		return tc.build_by_std_namespace(decl)
	}

	tc.push_err(token, "ident_not_exist", token.Kind)
	return nil
}

func (tc *_TypeChecker) build(decl_kind ast.TypeDeclKind) *TypeKind {
	var kind _Kind = nil

	switch decl_kind.(type) {
	case *ast.IdentType:
		kind = tc.build_ident(decl_kind.(*ast.IdentType))

	case *ast.RefType:
		kind = tc.build_ref(decl_kind.(*ast.RefType))

	case *ast.PtrType:
		kind = tc.build_ptr(decl_kind.(*ast.PtrType))

	case *ast.SlcType:
		kind = tc.build_slc(decl_kind.(*ast.SlcType))

	case *ast.ArrType:
		kind = tc.build_arr(decl_kind.(*ast.ArrType))

	case *ast.MapType:
		kind = tc.build_map(decl_kind.(*ast.MapType))

	case *ast.TupleType:
		kind = tc.build_tuple(decl_kind.(*ast.TupleType))

	case *ast.FnDecl:
		kind = tc.build_fn(decl_kind.(*ast.FnDecl))

	case *ast.NamespaceType:
		kind = tc.build_by_namespace(decl_kind.(*ast.NamespaceType))

	default:
		tc.push_err(tc.error_token, "invalid_type")
		return nil
	}

	if kind == nil {
		return nil
	}

	switch kind.(type) {
	case *TypeKind:
		return kind.(*TypeKind)

	default:
		return &TypeKind{
			kind: kind,
		}
	}
}

func (tc *_TypeChecker) check_decl(decl *ast.Type) *TypeKind {
	// Save current token.
	error_token := tc.error_token

	tc.error_token = decl.Token
	kind := tc.build(decl.Kind)
	tc.error_token = error_token

	return kind
}

func (tc *_TypeChecker) check(t *TypeSymbol) {
	if t.Decl == nil {
		return
	}

	kind := tc.check_decl(t.Decl)
	if kind == nil {
		t.remove_kind()
		return
	}
	t.Kind = kind
}
