package sema

import (
	"strconv"
	"strings"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/constant"
	"github.com/julelang/jule/constant/lit"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

func is_instanced_struct(s *StructIns) bool {
	return len(s.Decl.Generics) == len(s.Generics)
}

func normalize_bitsize(d *Data) {
	if !d.Is_const() {
		return
	}

	var kind string
	switch {
	case d.Constant.Is_f64():
		x := d.Constant.Read_f64()
		kind = types.Float_from_bits(types.Bitsize_of_float(x))
	
	case d.Constant.Is_i64():
		x := d.Constant.Read_i64()
		kind = types.Int_from_bits(types.Bitsize_of_int(x))
	
	case d.Constant.Is_u64():
		x := d.Constant.Read_u64()
		kind = types.Uint_from_bits(types.Bitsize_of_uint(x))
	
	default:
		return
	}

	d.Kind.kind = build_prim_type(kind)
}

func apply_cast_kind(d *Data) {
	if d.Cast_kind == nil {
		return
	}

	d.Model = &CastingExprModel{
		Expr:     d.Model,
		Kind:     d.Cast_kind,
		ExprKind: d.Kind,
	}
	d.Kind = d.Cast_kind
	d.Cast_kind = nil // Ignore, because model added.
}

// Value data.
type Data struct {
	Kind       *TypeKind
	Cast_kind  *TypeKind // This expression should be cast to this kind.
	Mutable    bool
	Lvalue     bool
	Variadiced bool
	Is_rune    bool
	Model      ExprModel

	// True if kind is declaration such as:
	//  - *Enum
	//  - *Struct
	//  - int type
	//  - bool type
	Decl       bool

	// Constant expression data.
	Constant   *constant.Const
}

// Reports whether Data is nil literal.
func (d *Data) Is_nil() bool { return d.Kind.Is_nil() }
// Reports whether Data is void.
func (d *Data) Is_void() bool { return d.Kind.Is_void() }
// Reports whether Data is constant expression.
func (d *Data) Is_const() bool { return d.Constant != nil }

func build_void_data() *Data {
	return &Data{
		Mutable:  false,
		Lvalue:   false,
		Decl:     false,
		Constant: nil,
		Kind:     &TypeKind{
			kind: build_prim_type("void"),
		},
	}
}

// Value.
type Value struct {
	Expr  *ast.Expr
	Data  *Data
}

func kind_by_bitsize(expr any) string {
	switch expr.(type) {
	case float64:
		x := expr.(float64)
		return types.Float_from_bits(types.Bitsize_of_float(x))

	case int64:
		x := expr.(int64)
		return types.Int_from_bits(types.Bitsize_of_int(x))

	case uint64:
		x := expr.(uint64)
		return types.Uint_from_bits(types.Bitsize_of_uint(x))

	default:
		return ""
	}
}

func check_data_for_integer_indexing(d *Data) (err_key string) {
	switch {
	case d == nil:
		return ""

	case d.Kind.Prim() == nil:
		return "invalid_expr"

	case !types.Is_int(d.Kind.Prim().To_str()):
		return "invalid_expr"

	case d.Is_const() && d.Constant.As_i64() < 0:
		return "overflow_limits"

	default:
		d.Cast_kind = &TypeKind{kind: build_prim_type(types.TypeKind_INT)}
		apply_cast_kind(d)
		return ""
	}
}

// Evaluator.
type _Eval struct {
	s        *_Sema  // Used for error logging.
	lookup   Lookup
	prefix   *TypeKind
	unsafety bool
	owner    *Var
}

func (e *_Eval) push_err(token lex.Token, key string, args ...any) {
	e.s.errors = append(e.s.errors, compiler_err(token, key, args...))
}

// Reports whether evaluation in unsafe scope.
func (e *_Eval) is_unsafe() bool { return e.unsafety }

// Reports whether evaluated expression is in global scope.
func (e *_Eval) is_global() bool {
	switch e.lookup.(type) {
	case *_Sema:
		return true

	default:
		return false
	}
}

func (e *_Eval) lit_nil() *Data {
	// Return new Data with nil kind.
	// Nil kind represents "nil" literal.

	constant := constant.New_nil()
	return &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: constant,
		Decl:     false,
		Kind:     &TypeKind{kind: nil},
		Model:    constant,
	}
}

func (e *_Eval) lit_str(lt *ast.LitExpr) *Data {
	s := lt.Value[1:len(lt.Value)-1] // Remove quotes.
	s = lit.To_str([]byte(s))
	constant := constant.New_str(s)

	return &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: constant,
		Decl:     false,
		Kind:     &TypeKind{
			kind: build_prim_type(types.TypeKind_STR),
		},
		Model:    constant,
	}
}

func (e *_Eval) lit_bool(lit *ast.LitExpr) *Data {
	constant := constant.New_bool(lit.Value == lex.KND_TRUE)
	return &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: constant,
		Decl:     false,
		Kind:     &TypeKind{
			kind: build_prim_type(types.TypeKind_BOOL),
		},
		Model:    constant,
	}
}

func (e *_Eval) lit_rune(l *ast.LitExpr) *Data {
	const BYTE_KIND = types.TypeKind_U8
	const RUNE_KIND = types.TypeKind_I32
	
	lt := l.Value[1:len(l.Value)-1] // Remove quotes.
	r := lit.To_rune([]byte(lt))
	data := &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: constant.New_i64(int64(r)),
		Decl:     false,
	}

	_, is_byte := lit.Is_byte_lit(l.Value)
	if is_byte {
		data.Kind = &TypeKind{
			kind: build_prim_type(BYTE_KIND),
		}
	} else {
		data.Kind = &TypeKind{
			kind: build_prim_type(RUNE_KIND),
		}
	}

	data.Model = &RuneExprModel{Code: r}
	data.Is_rune = true
	return data
}

func (e *_Eval) lit_float(l *ast.LitExpr) *Data {
	const FLOAT_KIND = types.TypeKind_F64

	f, _ := strconv.ParseFloat(l.Value, 64)
	constant := constant.New_f64(f)

	return &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: constant,
		Decl:     false,
		Kind:     &TypeKind{
			kind: build_prim_type(FLOAT_KIND),
		},
		Model:    constant,
	}
}

func (e *_Eval) lit_int(l *ast.LitExpr) *Data {
	const BIT_SIZE = 0b01000000

	lit := l.Value
	base := 0

	switch {
	case strings.HasPrefix(lit, "0x"): // Hexadecimal
		lit = lit[2:]
		base = 0b00010000

	case strings.HasPrefix(lit, "0b"): // Binary
		lit = lit[2:]
		base = 0b10

	case lit[0] == '0' && len(lit) > 1: // Octal
		lit = lit[1:]
		base = 0b1000

	default: // Decimal
		base = 0b1010
	}

	d := &Data{
		Lvalue:  false,
		Mutable: false,
		Decl:    false,
	}

	var value any = nil
	sig, err := strconv.ParseInt(lit, base, BIT_SIZE)
	if err == nil {
		value = sig
		d.Constant = constant.New_i64(sig)
	} else {
		unsig, _ := strconv.ParseUint(lit, base, BIT_SIZE)
		d.Constant = constant.New_u64(unsig)
		value = unsig
	}

	d.Kind = &TypeKind{
		kind: build_prim_type(kind_by_bitsize(value)),
	}

	normalize_bitsize(d)
	d.Model = d.Constant

	return d
}

func (e *_Eval) lit_num(l *ast.LitExpr) *Data {
	switch {
	case lex.Is_float(l.Value):
		return e.lit_float(l)

	default:
		return e.lit_int(l)
	}
}

func (e *_Eval) eval_lit(lit *ast.LitExpr) *Data {
	switch {
	case lit.Is_nil():
		return e.lit_nil()

	case lex.Is_str(lit.Value):
		return e.lit_str(lit)

	case lex.Is_bool(lit.Value):
		return e.lit_bool(lit)

	case lex.Is_rune(lit.Value):
		return e.lit_rune(lit)

	case lex.Is_num(lit.Value):
		return e.lit_num(lit)

	default:
		return nil
	}
}

func find_builtins_import(ident string, imp *ImportInfo) any {
	return find_package_builtin_def(imp.Link_path, ident)
}

func find_builtins_sema(ident string, s *_Sema) any {
	for _, imp := range s.file.Imports {
		if imp.Import_all || imp.exist_ident(ident) {
			def := find_builtins_import(ident, imp)
			if def != nil {
				return def
			}
		} 
	}
	return nil
}

func (e *_Eval) find_builtins(ident string) any {
	switch e.lookup.(type) {
	case *ImportInfo:
		def := find_builtins_import(ident, e.lookup.(*ImportInfo))
		if def != nil {
			return def
		}

	case *_Sema:
		def := find_builtins_sema(ident, e.lookup.(*_Sema))
		if def != nil {
			return def
		}

	case *_ScopeChecker:
		def := find_builtins_sema(ident, e.lookup.(*_ScopeChecker).s)
		if def != nil {
			return def
		}
	}

	return find_builtin_def(ident)
}

func (e *_Eval) get_def(ident string, cpp_linked bool) any {
	if !cpp_linked {
		enm := e.lookup.Find_enum(ident)
		if enm != nil {
			return enm
		}
	}

	v := e.lookup.Find_var(ident, cpp_linked)
	if v != nil {
		return v
	}

	f := e.lookup.Find_fn(ident, cpp_linked)
	if f != nil {
		return f
	}

	s := e.lookup.Find_struct(ident, cpp_linked)
	if s != nil {
		return s
	}

	ta := e.lookup.Find_type_alias(ident, cpp_linked)
	if ta != nil {
		return ta
	}

	return e.find_builtins(ident)
}

func (e *_Eval) eval_enum(enm *Enum, error_token lex.Token) *Data {
	if !e.s.is_accessible_define(enm.Public, enm.Token) {
		e.push_err(error_token, "ident_is_not_accessible", enm.Ident)
		return nil
	}

	return &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: nil,
		Decl:     true,
		Kind:     &TypeKind{
			kind: enm,
		},
	}
}

func (e *_Eval) eval_struct(s *StructIns, error_token lex.Token) *Data {
	if !e.s.is_accessible_define(s.Decl.Public, s.Decl.Token) {
		e.push_err(error_token, "ident_is_not_accessible", s.Decl.Ident)
		return nil
	}

	return &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: nil,
		Decl:     true,
		Kind:     &TypeKind{
			kind: s,
		},
		Model:    s,
	}
}

func (e *_Eval) eval_fn_ins(f *FnIns) *Data {
	return &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: nil,
		Decl:     false,
		Kind:     &TypeKind{
			kind: f,
		},
		Model:    f,
	}
}

func (e *_Eval) eval_fn(f *Fn, error_token lex.Token) *Data {
	if !e.s.is_accessible_define(f.Public, f.Token) {
		e.push_err(error_token, "ident_is_not_accessible", f.Ident)
		return nil
	}

	ins := f.instance()
	return e.eval_fn_ins(ins)
}

// Checks owner illegal cycles.
// Appends depend to depends if there is no illegal cycle.
// Returns true if e.owner is nil.
func (e *_Eval) check_illegal_cycles(v *Var, decl_token lex.Token) (ok bool) {
	if e.owner == nil {
		return true
	}

	// Check illegal cycle for itself.
	// Because refers's owner is ta.
	if e.owner == v {
		e.push_err(e.owner.Token, "illegal_cycle_refers_itself", e.owner.Ident)
		return false
	}

	const PADDING = 4

	message := ""

	push := func(v1 *Var, v2 *Var) {
		refers_to := build.Errorf("refers_to", v1.Ident, v2.Ident)
		message = strings.Repeat(" ", PADDING) + refers_to + "\n" + message
	}

	// Check cross illegal cycle.
	var check_cross func(v *Var) bool
	check_cross = func(v *Var) bool {
		for _, d := range v.Depends {
			if d == e.owner {
				push(v, d)
				return false
			}

			if !check_cross(d) {
				push(v, d)
				return false
			}
		}

		return true
	}

	if !check_cross(v) {
		err_msg := message
		message = ""
		push(e.owner, v)
		err_msg = err_msg + message
		e.push_err(decl_token, "illegal_cross_cycle", err_msg)
		return false
	}

	e.owner.Depends = append(e.owner.Depends, v)
	return true
}

func (e *_Eval) eval_var(v *Var, error_token lex.Token) *Data {
	if !e.s.is_accessible_define(v.Public, v.Token) {
		e.push_err(error_token, "ident_is_not_accessible", v.Ident)
		return nil
	}

	v.Used = true

	ok := e.check_illegal_cycles(v, error_token)
	if !ok {
		return nil
	}

	if !v.Cpp_linked && (v.Value == nil || v.Value.Data == nil) {
		if v.Constant {
			return nil
		}
	}

	d := &Data{
		Lvalue:   !v.Constant,
		Mutable:  v.Mutable,
		Decl:     false,
		Kind:     v.Kind.Kind.clone(),
		Model:    v,
	}

	if !v.Cpp_linked && v.Is_initialized() {
		d.Is_rune = v.Value.Data.Is_rune
	}

	if v.Constant {
		d.Constant = new(constant.Const)
		*d.Constant = *v.Value.Data.Constant
		d.Model = d.Constant
	}

	if d.Kind.Fnc() != nil {
		f := d.Kind.Fnc()
		if f.Decl != nil {
			// Ignore identifier for non-anonymous (because has an identifier via variable).
			f.Decl.Ident = v.Ident
		}
	}

	return d
}

func (e *_Eval) eval_type_alias(ta *TypeAlias, error_token lex.Token) *Data {
	if !e.s.is_accessible_define(ta.Public, ta.Token) {
		e.push_err(error_token, "ident_is_not_accessible", ta.Ident)
		return nil
	}

	ta.Used = true

	kind := ta.Kind.Kind.kind
	switch kind.(type) {
	case *StructIns:
		return e.eval_struct(kind.(*StructIns), error_token)

	case *Enum:
		return e.eval_enum(kind.(*Enum), error_token)

	case *Prim, *Slc:
		return &Data{
			Decl: true,
			Kind: ta.Kind.Kind.clone(),
		}
	
	default:
		e.push_err(error_token, "invalid_expr")
		return nil
	}
}

func (e *_Eval) eval_def(def any, ident lex.Token) *Data {
	switch def.(type) {
	case *Var:
		return e.eval_var(def.(*Var), ident)

	case *Enum:
		return e.eval_enum(def.(*Enum), ident)

	case *Struct:
		return e.eval_struct(def.(*Struct).instance(), ident)

	case *Fn:
		return e.eval_fn(def.(*Fn), ident)

	case *FnIns:
		return e.eval_fn_ins(def.(*FnIns))

	case *TypeAlias:
		return e.eval_type_alias(def.(*TypeAlias), ident)

	default:
		e.push_err(ident, "ident_not_exist", ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_ident(ident *ast.IdentExpr) *Data {
	def := e.get_def(ident.Ident, ident.Cpp_linked)
	return e.eval_def(def, ident.Token)
}

func (e *_Eval) eval_unary_minus(d *Data) *Data {
	t := d.Kind.Prim()
	if t == nil || !types.Is_num(t.To_str()) {
		return nil
	}
	
	if d.Is_const() {
		switch {
		case d.Constant.Is_f64():
			d.Constant.Set_f64(-d.Constant.Read_f64())

		case d.Constant.Is_i64():
			d.Constant.Set_i64(-d.Constant.Read_i64())

		case d.Constant.Is_u64():
			d.Constant.Set_u64(-d.Constant.Read_u64())
		}
	}

	d.Lvalue = false
	d.Model = &UnaryExprModel{
		Expr: d.Model,
		Op:   lex.KND_MINUS,
	}
	return d
}

func (e *_Eval) eval_unary_plus(d *Data) *Data {
	t := d.Kind.Prim()
	if t == nil || !types.Is_num(t.To_str()) {
		return nil
	}
	
	if d.Is_const() {
		switch {
		case d.Constant.Is_f64():
			d.Constant.Set_f64(+d.Constant.Read_f64())

		case d.Constant.Is_i64():
			d.Constant.Set_i64(+d.Constant.Read_i64())

		case d.Constant.Is_u64():
			d.Constant.Set_u64(+d.Constant.Read_u64())
		}
	}
	
	d.Lvalue = false
	d.Model = &UnaryExprModel{
		Expr: d.Model,
		Op:   lex.KND_PLUS,
	}
	return d
}

func (e *_Eval) eval_unary_caret(d *Data) *Data {
	t := d.Kind.Prim()
	if t == nil || !types.Is_int(t.To_str()) {
		return nil
	}

	if d.Is_const() {
		switch {
		case d.Constant.Is_i64():
			d.Constant.Set_i64(^d.Constant.Read_i64())

		case d.Constant.Is_u64():
			d.Constant.Set_u64(^d.Constant.Read_u64())
		}
	}

	d.Lvalue = false
	d.Model = &UnaryExprModel{
		Expr: d.Model,
		Op:   lex.KND_CARET,
	}
	return d
}

func (e *_Eval) eval_unary_excl(d *Data) *Data {
	t := d.Kind.Prim()
	if t == nil || !t.Is_bool() {
		return nil
	}
	
	if d.Is_const() {
		switch {
		case d.Constant.Is_bool():
			d.Constant.Set_bool(!d.Constant.Read_bool())
		}
	}
	
	d.Lvalue = false
	d.Model = &UnaryExprModel{
		Expr: d.Model,
		Op:   lex.KND_EXCL,
	}
	return d
}

func (e *_Eval) eval_unary_star(d *Data, op lex.Token) *Data {
	if !e.is_unsafe() {
		e.push_err(op, "unsafe_behavior_at_out_of_unsafe_scope")
	}

	t := d.Kind.Ptr()
	if t == nil || t.Is_unsafe() {
		return nil
	}
	d.Constant = nil
	d.Lvalue = true
	d.Kind = t.Elem
	d.Model = &UnaryExprModel{
		Expr: d.Model,
		Op:   lex.KND_STAR,
	}
	return d
}

func (e *_Eval) eval_unary_amper(d *Data) *Data {
	switch d.Model.(type) {
	case *StructLitExprModel:
		lit := d.Model.(*StructLitExprModel)
		d.Kind = &TypeKind{
			kind: &Ref{
				Elem: &TypeKind{kind: lit.Strct},
			},
		}

		d.Model = &AllocStructLitExprModel{
			Lit: lit,
		}

	default:
		switch {
		case d.Kind.Ref() != nil:
			d.Kind = &TypeKind{
				kind: &Ptr{Elem: d.Kind.Ref().Elem.clone()},
			}
			d.Model = &GetRefPtrExprModel{
				Expr: d.Model,
			}

		case can_get_ptr(d):
			d.Kind = &TypeKind{
				kind: &Ptr{Elem: d.Kind.clone()},
			}
			d.Model = &UnaryExprModel{
				Expr: d.Model,
				Op:   lex.KND_AMPER,
			}

		default:
			d = nil
		}
	}

	if d != nil {
		d.Constant = nil
		d.Lvalue = true
		d.Mutable = true
	}

	return d
}

func (e *_Eval) eval_unary(u *ast.UnaryExpr) *Data {
	d := e.eval_expr_kind(u.Expr)
	if d == nil {
		return nil
	}

	cast_kind := d.Cast_kind
	switch u.Op.Kind {
	case lex.KND_MINUS:
		d = e.eval_unary_minus(d)

	case lex.KND_PLUS:
		d = e.eval_unary_plus(d)

	case lex.KND_CARET:
		d = e.eval_unary_caret(d)

	case lex.KND_EXCL:
		d = e.eval_unary_excl(d)

	case lex.KND_STAR:
		d = e.eval_unary_star(d, u.Op)

	case lex.KND_AMPER:
		d = e.eval_unary_amper(d)

	default:
		d = nil
	}

	if d == nil {
		e.push_err(u.Op, "invalid_expr_unary_operator", u.Op.Kind)
	} else if d.Is_const() {
		d.Model = d.Constant
	} else if cast_kind != nil {
		d.Cast_kind = cast_kind
		apply_cast_kind(d)
	}

	return d
}

func (e *_Eval) eval_variadic(v *ast.VariadicExpr) *Data {
	d := e.eval_expr_kind(v.Expr)
	if d == nil {
		return nil
	}

	if !is_variadicable(d.Kind) {
		e.push_err(v.Token, "variadic_with_non_variadicable", d.Kind.To_str())
		return nil
	}

	d.Variadiced = true
	d.Kind = d.Kind.Slc().Elem
	return d
}

func (e *_Eval) eval_unsafe(u *ast.UnsafeExpr) *Data {
	unsafety := e.unsafety
	e.unsafety = true

	d := e.eval_expr_kind(u.Expr)

	e.unsafety = unsafety

	return d
}

func (e *_Eval) eval_arr(s *ast.SliceExpr) *Data {
	// Arrays always has type prefixes.
	pt := e.prefix.Arr()

	arr := &Arr{
		Auto: false,
		N:    0,
		Elem: pt.Elem,
	}

	if pt.Auto {
		arr.N = len(s.Elems)
	} else {
		arr.N = len(s.Elems)
		if arr.N > pt.N {
			e.push_err(s.Token, "overflow_limits")
		} else if arr.N < pt.N {
			arr.N = pt.N
		}
	}

	model := &ArrayExprModel{
		Kind:  arr,
		Elems: make([]ExprModel, len(s.Elems)),
	}

	prefix := e.prefix
	e.prefix = arr.Elem
	for i, elem := range s.Elems {
		d := e.eval_expr_kind(elem)
		if d == nil {
			continue
		}

		e.s.check_assign_type(arr.Elem, d, s.Token, true)
		model.Elems[i] = d.Model
	}
	e.prefix = prefix

	return &Data{
		Kind:  &TypeKind{kind: arr},
		Model: model,
	}
}

func (e *_Eval) eval_exp_slc(s *ast.SliceExpr, elem_type *TypeKind) *Data {
	slc := &Slc{
		Elem: elem_type,
	}

	model := &SliceExprModel{
		Elem_kind:  elem_type,
		Elems: make([]ExprModel, len(s.Elems)),
	}

	prefix := e.prefix
	e.prefix = slc.Elem
	for i, elem := range s.Elems {
		d := e.eval_expr_kind(elem)
		if d == nil {
			continue
		}

		e.s.check_assign_type(slc.Elem, d, s.Token, true)
		model.Elems[i] = d.Model
	}
	e.prefix = prefix

	return &Data{
		Kind:  &TypeKind{kind: slc},
		Model: model,
	}
}

func (e *_Eval) eval_slice_expr(s *ast.SliceExpr) *Data {
	if e.prefix != nil {
		switch {
		case e.prefix.Arr() != nil:
			return e.eval_arr(s)

		case e.prefix.Slc() != nil:
			pt := e.prefix.Slc()
			return e.eval_exp_slc(s, pt.Elem)
		}
	}

	prefix := e.prefix
	e.prefix = nil

	if len(s.Elems) == 0 {
		e.push_err(s.Token, "dynamic_type_annotation_failed")
		return nil
	}

	first_elem := e.eval_expr_kind(s.Elems[0])
	if first_elem == nil {
		return nil
	}

	d := e.eval_exp_slc(s, first_elem.Kind)

	e.prefix = prefix
	return d
}

func (e *_Eval) check_integer_indexing_by_data(d *Data, token lex.Token) {
	err_key := check_data_for_integer_indexing(d)
	if err_key != "" {
		e.push_err(token, err_key)
	}
}

func (e *_Eval) indexing_ptr(d *Data, index *Data, i *ast.IndexingExpr) {
	e.check_integer_indexing_by_data(index, i.Token)

	ptr := d.Kind.Ptr()
	switch {
	case ptr.Is_unsafe():
		e.push_err(i.Token, "unsafe_ptr_indexing")
		return

	case !e.is_unsafe():
		e.push_err(i.Token, "unsafe_behavior_at_out_of_unsafe_scope")
	}

	d.Kind = ptr.Elem.clone()
}

func (e *_Eval) indexing_arr(d *Data, index *Data, i *ast.IndexingExpr) {
	arr := d.Kind.Arr()
	d.Kind = arr.Elem.clone()
	e.check_integer_indexing_by_data(index, i.Token)
}

func (e *_Eval) indexing_slc(d *Data, index *Data, i *ast.IndexingExpr) {
	slc := d.Kind.Slc()
	d.Kind = slc.Elem.clone()
	e.check_integer_indexing_by_data(index, i.Token)
}

func (e *_Eval) indexing_map(d *Data, index *Data, i *ast.IndexingExpr) {
	if index == nil {
		return
	}

	m := d.Kind.Map()
	e.s.check_type_compatibility(m.Key, index.Kind, i.Token, true)

	d.Kind = m.Val.clone()
}

func (e *_Eval) indexing_str(d *Data, index *Data, i *ast.IndexingExpr) {
	const BYTE_KIND = types.TypeKind_U8
	d.Kind.kind = build_prim_type(BYTE_KIND)

	if index == nil {
		return
	}

	e.check_integer_indexing_by_data(index, i.Token)

	if !index.Is_const() {
		d.Constant = nil
		return
	}

	if d.Is_const() {
		error_token := i.Token
		i := index.Constant.As_i64()
		s := d.Constant.Read_str()
		if int(i) >= len(s) {
			e.push_err(error_token, "overflow_limits")
		} else {
			d.Constant.Set_u64(uint64(s[i]))
		}
	}
}

func (e *_Eval) to_indexing(d *Data, index *Data, i *ast.IndexingExpr) {
	switch {
	case d.Kind.Ptr() != nil:
		e.indexing_ptr(d, index, i)
		return

	case d.Kind.Arr() != nil:
		e.indexing_arr(d, index, i)
		return

	case d.Kind.Slc() != nil:
		e.indexing_slc(d, index, i)
		return

	case d.Kind.Map() != nil:
		e.indexing_map(d, index, i)
		return

	case d.Kind.Prim() != nil:
		prim := d.Kind.Prim()
		switch {
		case prim.Is_str():
			e.indexing_str(d, index, i)
			return
		}
	}

	e.push_err(i.Token, "not_supports_indexing", d.Kind.To_str())
}

func (e *_Eval) eval_indexing(i *ast.IndexingExpr) *Data {
	d := e.eval_expr_kind(i.Expr)
	if d == nil {
		return nil
	}

	expr_model := d.Model
	index := e.eval_expr_kind(i.Index)
	e.to_indexing(d, index, i)

	if index != nil {
		if d.Is_const() {
			d.Model = d.Constant
		} else {
			d.Model = &IndexigExprModel{
				Expr:  expr_model,
				Index: index.Model,
			}
		}
	}

	return d
}

// Returns left and right index values.
// Returns zero integer expression if slicing have not left index.
// So, left index always represents an expression.
// Left data is nil if expression eval failed.
func (e *_Eval) eval_slicing_exprs(s *ast.SlicingExpr) (*Data, *Data) {
	var l *Data = nil
	var r *Data = nil

	if s.Start != nil {
		l = e.eval_expr_kind(s.Start)
		if l != nil {
			e.check_integer_indexing_by_data(l, s.Token)
		} else {
			return nil, nil
		}
	} else {
		l = &Data{
			Constant: constant.New_i64(0),
			Kind: &TypeKind{kind: build_prim_type(types.SYS_INT)},
		}
		l.Model = l.Constant
	}

	if s.To != nil {
		r = e.eval_expr_kind(s.To)
		if r != nil {
			e.check_integer_indexing_by_data(r, s.Token)
		} else {
			return nil, nil
		}
	}

	return l, r
}

func (e *_Eval) slicing_arr(d *Data) {
	d.Lvalue = false
	d.Kind.kind = &Slc{Elem: d.Kind.Arr().Elem.clone()}
}

func (e *_Eval) slicing_slc(d *Data) {
	d.Lvalue = false
}

func (e *_Eval) slicing_str(d *Data, l *Data, r *Data) {
	d.Lvalue = false
	if !d.Is_const() {
		return
	}

	if l == nil || r == nil {
		d.Constant = nil
		return
	}
	
	if l.Is_const() && r.Is_const() {
		left := l.Constant.As_i64()
		if left < 0 {
			return
		}

		s := d.Constant.Read_str()
		var right int64
		if r == nil {
			right = int64(len(s))
		} else {
			right = r.Constant.As_i64()
		}

		if left > right {
			return
		}
		d.Constant.Set_str(s[left:right])
	} else {
		d.Constant = nil
	}
}

func (e *_Eval) check_slicing(d *Data, l *Data, r *Data, s *ast.SlicingExpr) {
	switch {
	case d.Kind.Arr() != nil:
		e.slicing_arr(d)
		return

	case d.Kind.Slc() != nil:
		e.slicing_slc(d)
		return

	case d.Kind.Prim() != nil:
		prim := d.Kind.Prim()
		switch {
		case prim.Is_str():
			e.slicing_str(d, l, r)
			return
		}
	}

	e.push_err(s.Token, "not_supports_slicing", d.Kind.To_str())
}

func (e *_Eval) eval_slicing(s *ast.SlicingExpr) *Data {
	d := e.eval_expr_kind(s.Expr)
	if d == nil {
		return nil
	}

	l, r := e.eval_slicing_exprs(s)
	if l == nil {
		return d
	}

	e.check_slicing(d, l, r, s)
	d.Cast_kind = nil

	model := &SlicingExprModel{
		Expr: d.Model,
		L:    l.Model,
	}

	if r != nil {
		model.R = r.Model
	}

	d.Model = model
	return d
}

func (e *_Eval) cast_ptr(t *TypeKind, d *Data, error_token lex.Token) {
	if !e.is_unsafe() {
		e.push_err(error_token, "unsafe_behavior_at_out_of_unsafe_scope")
		return
	}

	prim := d.Kind.Prim()
	if d.Kind.Ptr() == nil && (prim == nil || !types.Is_int(prim.To_str())) {
		e.push_err(error_token, "type_not_supports_casting_to", d.Kind.To_str(), t.To_str())
	}

	d.Constant = nil
}

func (e *_Eval) cast_struct(t *TypeKind, d *Data, error_token lex.Token) {
	tr := d.Kind.Trt()
	if tr == nil {
		e.push_err(error_token, "type_not_supports_casting_to", d.Kind.To_str(), t.To_str())
		return
	}

	s := t.Strct()
	if !s.Decl.Is_implements(tr) {
		e.push_err(error_token, "type_not_supports_casting_to", d.Kind.To_str(), t.To_str())
	}
}

func (e *_Eval) cast_ref(t *TypeKind, d *Data, error_token lex.Token) {
	ref := t.Ref()
	if ref.Elem.Strct() != nil {
		e.cast_struct(t, d, error_token)
		return
	}

	e.push_err(error_token, "type_not_supports_casting_to", d.Kind.To_str(), t.To_str())
}

func (e *_Eval) cast_slc(t *TypeKind, d *Data, error_token lex.Token) {
	if d.Kind.Prim() == nil || !d.Kind.Prim().Is_str() {
		e.push_err(error_token, "type_not_supports_casting_to", d.Kind.To_str(), t.To_str())
		return
	}

	t = t.Slc().Elem
	prim := t.Prim()
	if prim == nil || (!prim.Is_u8() && !prim.Is_i32()) {
		e.push_err(error_token, "type_not_supports_casting_to", d.Kind.To_str(), t.To_str())
	}
}

func (e *_Eval) cast_str(d *Data, error_token lex.Token) {
	if d.Kind.Prim() != nil {
		prim := d.Kind.Prim()
		if !prim.Is_u8() && !prim.Is_i32() {
			e.push_err(error_token, "type_not_supports_casting_to", types.TypeKind_STR, d.Kind.To_str())
		}
		return
	}

	if d.Kind.Slc() == nil {
		e.push_err(error_token, "type_not_supports_casting_to", types.TypeKind_STR, d.Kind.To_str())
		return
	}

	t := d.Kind.Slc().Elem
	prim := t.Prim()
	if prim == nil || (!prim.Is_u8() && !prim.Is_i32()) {
		e.push_err(error_token, "type_not_supports_casting_to", types.TypeKind_STR, d.Kind.To_str())
	}
}

func (e *_Eval) cast_int(t *TypeKind, d *Data, error_token lex.Token) {
	if d.Is_const() {
		prim := t.Prim()
		switch {
		case types.Is_sig_int(prim.kind):
			d.Constant.Set_i64(d.Constant.As_i64())

		case types.Is_unsig_int(prim.kind):
			d.Constant.Set_u64(d.Constant.As_u64())
		}
	}

	if d.Kind.Enm() != nil {
		e := d.Kind.Enm()
		if types.Is_num(e.Kind.Kind.To_str()) {
			return
		}
	}

	if d.Kind.Ptr() != nil {
		prim := t.Prim()
		if prim.Is_uintptr() {
			// Ignore case.
		} else if !e.is_unsafe() {
			e.push_err(error_token, "unsafe_behavior_at_out_of_unsafe_scope")
		} else if !prim.Is_i32() && !prim.Is_i64() && !prim.Is_u16() && !prim.Is_u32() && !prim.Is_u64() {
			e.push_err(error_token, "type_not_supports_casting_to", d.Kind.To_str(), t.To_str())
		}
		return
	}

	prim := d.Kind.Prim()
	if prim != nil && types.Is_num(prim.To_str()) {
		return
	}

	e.push_err(error_token, "type_not_supports_casting_to", d.Kind.To_str(), t.To_str())
}

func (e *_Eval) cast_num(t *TypeKind, d *Data, error_token lex.Token) {
	if d.Is_const() {
		prim := t.Prim()
		switch {
		case types.Is_float(prim.kind):
			d.Constant.Set_f64(d.Constant.As_f64())

		case types.Is_sig_int(prim.kind):
			d.Constant.Set_i64(d.Constant.As_i64())

		case types.Is_unsig_int(prim.kind):
			d.Constant.Set_u64(d.Constant.As_u64())
		}
	}

	if d.Kind.Enm() != nil {
		e := d.Kind.Enm()
		if types.Is_num(e.Kind.Kind.To_str()) {
			return
		}
	}

	prim := d.Kind.Prim()
	if prim != nil && types.Is_num(prim.To_str()) {
		return
	}

	e.push_err(error_token, "type_not_supports_casting_to", d.Kind.To_str(), t.To_str())
}

func (e *_Eval) cast_prim(t *TypeKind, d *Data, error_token lex.Token) {
	prim := t.Prim()
	switch {
	case prim.Is_any():
		// The any type supports casting to any data type.

	case prim.Is_str():
		e.cast_str(d, error_token)

	case types.Is_int(prim.To_str()):
		e.cast_int(t, d, error_token)

	case types.Is_num(prim.To_str()):
		e.cast_num(t, d, error_token)

	default:
		e.push_err(error_token, "type_not_supports_casting", t.To_str())
	}
}

func (e *_Eval) eval_cast_by_type_n_data(t *TypeKind, d *Data, error_token lex.Token) *Data {
	switch {
	case t.Ptr() != nil:
		e.cast_ptr(t, d, error_token)

	case t.Ref() != nil:
		e.cast_ref(t, d, error_token)

	case t.Slc() != nil:
		e.cast_slc(t, d, error_token)

	case t.Strct() != nil:
		e.cast_struct(t, d, error_token)

	case t.Prim() != nil:
		e.cast_prim(t, d, error_token)

	default:
		e.push_err(error_token, "type_not_supports_casting", t.To_str())
		d = nil
	}

	if d == nil {
		return nil
	}

	d.Lvalue = is_lvalue(t)
	d.Mutable = is_mut(t)
	d.Decl = false
	if t.Prim() != nil && d.Is_const() {
		d.Model = d.Constant
	}
	d.Cast_kind = t
	//d.Kind = t // Do not this, will be set automatically end of the eval.

	return d
}

func (e *_Eval) eval_cast(c *ast.CastExpr) *Data {
	t := build_type(c.Kind)
	ok := e.s.check_type(t, e.lookup)
	if !ok {
		return nil
	}
	
	d := e.eval_expr_kind(c.Expr)
	if d == nil {
		return nil
	}

	d = e.eval_cast_by_type_n_data(t.Kind, d, c.Kind.Token)

	return d
}

func (e *_Eval) eval_ns_selection(s *ast.NsSelectionExpr) *Data {
	if s.Ns[0].Id == lex.ID_PRIM {
		return e.eval_prim_type_sub_ident(s.Ns[0], s.Ident)
	}

	path := build_link_path_by_tokens(s.Ns)
	imp := e.lookup.Select_package(func(p *ImportInfo) bool {
		return p.Link_path == path
	})

	if imp == nil || !imp.is_lookupable(lex.KND_SELF) {
		e.push_err(s.Ident, "namespace_not_exist", s.Ident.Kind)
		return nil
	}

	lookup := e.lookup
	e.lookup = imp

	const CPP_LINKED = false
	def := e.get_def(s.Ident.Kind, CPP_LINKED)
	d := e.eval_def(def, s.Ident)

	e.lookup = lookup

	return d
}

func (e *_Eval) eval_struct_lit_explicit(s *StructIns, exprs []ast.ExprData, error_token lex.Token) *Data {
	ok := e.s.check_generic_quantity(len(s.Decl.Generics), len(s.Generics), error_token)
	if !ok {
		return nil
	}
	// NOTE: Instance already checked (just fields) if generic quantity passes.

	slc := _StructLitChecker{
		e:           e,
		error_token: error_token,
		s:           s,
	}
	slc.check(exprs)

	return &Data{
		Mutable: true,
		Kind:    &TypeKind{kind: s},
		Model:   &StructLitExprModel{
			Strct: s,
			Args:  slc.args,
		},
	}
}

func (e *_Eval) eval_struct_lit(lit *ast.StructLit) *Data {
	t := build_type(lit.Kind)
	ok := e.s.check_type(t, e.lookup)
	if !ok {
		return nil
	}

	s := t.Kind.Strct()
	if s == nil {
		e.push_err(lit.Kind.Token, "invalid_syntax")
		return nil
	}

	return e.eval_struct_lit_explicit(s, lit.Exprs, lit.Kind.Token)
}

func (e *_Eval) eval_type(t *ast.Type) *Data {
	tk := build_type(t)
	ok := e.s.check_type(tk, e.lookup)
	if !ok {
		return nil
	}

	return &Data{
		Decl:  true,
		Kind:  tk.Kind,
		Model: tk.Kind,
	}
}

func (e *_Eval) call_type_fn(fc *ast.FnCallExpr, d *Data) *Data {
	if len(fc.Generics) > 0 {
		e.push_err(fc.Token, "type_not_supports_generics", d.Kind.To_str())
	} else if len(fc.Args) < 1 {
		e.push_err(fc.Token, "missing_expr_for", "v")
	} else if len(fc.Args) > 1 {
		e.push_err(fc.Args[1].Token, "argument_overflow")
	}

	if len(fc.Args) > 0 {
		arg := e.eval_expr_kind(fc.Args[0].Kind)

		// Skip strings beceause string constructor
		// takes any type.
		prim := d.Kind.Prim()
		if prim != nil && prim.Is_str() {
			d.Model = &StrConstructorcallExprModel{
				Expr: arg.Model,
			}
			goto _ret
		}

		if arg != nil {
			d = e.eval_cast_by_type_n_data(d.Kind, arg, fc.Args[0].Token)
		}
	}

_ret:
	d.Decl = false
	return d
}

func (e *_Eval) check_fn_call_generics(f *FnIns,
	fc *ast.FnCallExpr) (ok bool, dynamic_annotation bool) {
	switch {
	case len(f.Decl.Generics) > 0 && len(fc.Generics) == 0 && len(f.Params) > 0:
		dynamic_annotation = true
		// Make empty types to generics for ordering.
		f.Generics = make([]*TypeKind, len(f.Decl.Generics))
		return true, true

	case !e.s.check_generic_quantity(len(f.Decl.Generics), len(fc.Generics), fc.Token):
		return false, false

	default:
		// Build real kinds of generic types.
		f.Generics = make([]*TypeKind, len(f.Decl.Generics))
		for i, g := range fc.Generics {
			k := build_type(g)
			ok := e.s.check_type(k, e.lookup)
			if !ok {
				return false, false
			}
			f.Generics[i] = k.Kind
		}

		return true, false
	}
}

func (e *_Eval) call_builtin_fn(fc *ast.FnCallExpr, d *Data) *Data {
	f := d.Kind.Fnc()
	
	d = f.Caller(e, fc, d)
	if d == nil {
		return d
	}

	d.Mutable = true
	return d
}

func (e *_Eval) call_fn(fc *ast.FnCallExpr, d *Data) *Data {
	f := d.Kind.Fnc()
	if f.Is_builtin() {
		return e.call_builtin_fn(fc, d)
	}

	old := e.s
	if f.Decl.Owner != nil {
		e.s = f.Decl.Owner.sema
	}

	defer func() {
		if old != e.s {
			old.errors = append(old.errors, e.s.errors...)
		}
		e.s = old
	}()

	if !d.Mutable && f.Decl.Is_method() && f.Decl.Params[0].Mutable {
		e.push_err(fc.Token, "mutable_operation_on_immutable")
	} else if !e.is_unsafe() && f.Decl.Unsafety {
		e.push_err(fc.Token, "unsafe_behavior_at_out_of_unsafe_scope")
	}

	ok, dynamic_annotation := e.check_fn_call_generics(f, fc)
	if !ok {
		return nil
	}

	if !dynamic_annotation {
		ok = e.s.reload_fn_ins_types(f)
		if !ok {
			return nil
		}
	} else {
		e.s.build_fn_non_generic_type_kinds(f, false)
	}

	fcac := _FnCallArgChecker{
		e:                  e,
		f:                  f,
		args:               fc.Args,
		dynamic_annotation: dynamic_annotation,
		error_token:        fc.Token,
	}
	ok = fcac.check()
	if !ok && dynamic_annotation {
		return nil
	}

	is_unique_ins, pos := f.Decl.append_instance(f)
	if pos != -1 {
		f = f.Decl.Instances[pos]
	}

	call_model := d.Model

	if f.Decl.Is_void() {
		d = build_void_data()
	} else {
		if dynamic_annotation {
			ok = e.s.reload_fn_ins_types(f)
			if !ok {
				return nil
			}
		}

		d.Kind = f.Result
		d.Lvalue = is_lvalue(f.Result)
	}

	d.Mutable = true
	d.Model = &FnCallExprModel{
		Func: f,
		IsCo: fc.Concurrent,
		Expr: call_model,
		Args: fcac.arg_models,
	}

	if len(f.Generics) > 0 && is_unique_ins {
		// Check generic function instance instantly.
		e.s.check_fn_ins(f)
	}

	return d
}

func (e *_Eval) eval_fn_call(fc *ast.FnCallExpr) *Data {
	d := e.eval_expr_kind(fc.Expr.Kind)
	if d == nil {
		return nil
	}

	if d.Decl {
		return e.call_type_fn(fc, d)
	}

	if d.Kind.Fnc() == nil {
		e.push_err(fc.Token, "invalid_syntax")
		return nil
	}

	return e.call_fn(fc, d)
}

func (e *_Eval) eval_enum_sub_ident(enm *Enum, ident lex.Token) *Data {
	d := &Data{
		Lvalue:   false,
		Decl:     false,
		Mutable:  false,
		Kind:     &TypeKind{kind: enm},
	}

	item := enm.Find_item(ident.Kind)
	if item == nil {
		e.push_err(ident, "obj_have_not_ident", ident.Kind)
	} else {
		d.Constant = new(constant.Const)
		*d.Constant = *item.Value.Data.Constant
		d.Model = d.Constant
	}

	return d
}

func (e *_Eval) eval_trait_sub_ident(d *Data, trt *Trait, ident lex.Token) *Data {
	f := trt.Find_method(ident.Kind)
	if f == nil {
		e.push_err(ident, "obj_have_not_ident", ident.Kind)
		return nil
	}

	return &Data{
		Lvalue:   false,
		Decl:     false,
		Mutable:  false,
		Constant: nil,
		Kind:     &TypeKind{kind: f.instance()},
		Model:    &TraitSubIdentExprModel{
			Expr:  d.Model,
			Ident: ident.Kind,
		},
	}
}

func (e *_Eval) eval_struct_sub_ident(d *Data, s *StructIns, si *ast.SubIdentExpr, ref bool) *Data {
	f := s.Find_field(si.Ident.Kind)
	if f != nil {
		model := &StrctSubIdentExprModel{
			ExprKind: d.Kind,
			Expr:     d.Model,
			Field:    f,
		}
		d.Model = model
		d.Kind = f.Kind.clone()

		if f.Decl.Mutable && !d.Mutable {
			// Interior mutability.
			switch e.lookup.(type) {
			case *_ScopeChecker:
				scope := e.lookup.(*_ScopeChecker)
				d.Mutable = scope.owner != nil && scope.owner.Owner == s
				if d.Mutable {
					v := new(Var)
					*v = *model.Expr.(*Var)
					v.Mutable = true
					model.Expr = v
				}
			}
		}

		return d
	}

	m := s.Find_method(si.Ident.Kind)
	if m == nil {
		e.push_err(si.Ident, "obj_have_not_ident", si.Ident.Kind)
		return nil
	}

	if m.Params[0].Is_ref() && !ref {
		e.push_err(si.Ident, "ref_method_used_with_not_ref_instance")
	}

	ins := m.instance()
	ins.Owner = s
	d.Model = &StrctSubIdentExprModel{
		ExprKind: d.Kind,
		Expr:     d.Model,
		Method:   ins,
	}
	d.Kind = &TypeKind{kind: ins}
	return d
}

func (e *_Eval) eval_slice_sub_ident(d *Data, ident lex.Token) *Data {
	switch ident.Kind {
	case "len":
		return &Data{
			Mutable: false,
			Kind:    &TypeKind{kind: build_prim_type(types.SYS_INT)},
			Model:   &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_len()",
			},
		}

	case "cap":
		return &Data{
			Mutable: false,
			Kind:    &TypeKind{kind: build_prim_type(types.SYS_INT)},
			Model:   &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_cap()",
			},
		}

	default:
		e.push_err(ident, "obj_have_not_ident", ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_array_sub_ident(d *Data, ident lex.Token) *Data {
	switch ident.Kind {
	case "len":
		c := constant.New_i64(int64(d.Kind.Arr().N))
		return &Data{
			Constant: c,
			Mutable:  false,
			Kind:     &TypeKind{kind: build_prim_type(types.SYS_INT)},
			Model:    c,
		}

	default:
		e.push_err(ident, "obj_have_not_ident", ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_map_sub_ident(d *Data, ident lex.Token) *Data {
	switch ident.Kind {
	case "len":
		return &Data{
			Mutable: false,
			Kind:    &TypeKind{kind: build_prim_type(types.SYS_INT)},
			Model:   &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_len()",
			},
		}

	case "clear":
		return &Data{
			Kind: &TypeKind{
				kind:  &FnIns{
					Caller: builtin_caller_common,
				},
			},
			Model: &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_clear",
			},
		}

	case "keys":
		m := d.Kind.Map()
		return &Data{
			Kind: &TypeKind{
				kind: &FnIns{
					Caller: builtin_caller_common,
					Result: &TypeKind{
						kind: &Slc{
							Elem: m.Key,
						},
					},
				},
			},
			Model: &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_keys",
			},
		}

	case "values":
		m := d.Kind.Map()
		return &Data{
			Kind: &TypeKind{
				kind: &FnIns{
					Caller: builtin_caller_common,
					Result: &TypeKind{
						kind: &Slc{
							Elem: m.Val,
						},
					},
				},
			},
			Model: &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_values",
			},
		}

	case "has":
		m := d.Kind.Map()
		return &Data{
			Kind: &TypeKind{
				kind: &FnIns{
					Caller: builtin_caller_common,
					Params: []*ParamIns{
						{
							Decl: &Param{
								Ident: "key",
							},
							Kind: m.Key,
						},
					},
					Result: &TypeKind{kind: build_prim_type(types.TypeKind_BOOL)},
				},
			},
			Model: &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_has",
			},
		}

	case "del":
		m := d.Kind.Map()
		return &Data{
			Kind: &TypeKind{
				kind: &FnIns{
					Caller: builtin_caller_common,
					Params: []*ParamIns{
						{
							Decl: &Param{
								Ident: "key",
							},
							Kind: m.Key,
						},
					},
				},
			},
			Model: &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_del",
			},
		}

	default:
		e.push_err(ident, "obj_have_not_ident", ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_str_sub_ident(d *Data, ident lex.Token) *Data {
	switch ident.Kind {
	case "len":
		return &Data{
			Mutable: false,
			Kind:    &TypeKind{kind: build_prim_type(types.SYS_INT)},
			Model:   &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_len()",
			},
		}

	case "has_prefix":
		return &Data{
			Kind: &TypeKind{
				kind: &FnIns{
					Caller: builtin_caller_common,
					Params: []*ParamIns{
						{
							Decl: &Param{
								Ident: "sub",
							},
							Kind: d.Kind,
						},
					},
					Result: &TypeKind{kind: build_prim_type(types.TypeKind_BOOL)},
				},
			},
			Model: &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_has_prefix",
			},
		}

	case "has_suffix":
		return &Data{
			Kind: &TypeKind{
				kind: &FnIns{
					Caller: builtin_caller_common,
					Params: []*ParamIns{
						{
							Decl: &Param{
								Ident: "sub",
							},
							Kind: d.Kind,
						},
					},
					Result: &TypeKind{kind: build_prim_type(types.TypeKind_BOOL)},
				},
			},
			Model: &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_has_suffix",
			},
		}

	case "find":
		return &Data{
			Kind: &TypeKind{
				kind: &FnIns{
					Caller: builtin_caller_common,
					Params: []*ParamIns{
						{
							Decl: &Param{
								Ident: "sub",
							},
							Kind: d.Kind,
						},
					},
					Result: &TypeKind{kind: build_prim_type(types.TypeKind_INT)},
				},
			},
			Model: &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_find",
			},
		}

	case "rfind":
		return &Data{
			Kind: &TypeKind{
				kind: &FnIns{
					Caller: builtin_caller_common,
					Params: []*ParamIns{
						{
							Decl: &Param{
								Ident: "sub",
							},
							Kind: d.Kind,
						},
					},
					Result: &TypeKind{kind: build_prim_type(types.TypeKind_INT)},
				},
			},
			Model: &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_rfind",
			},
		}

	case "trim":
		return &Data{
			Kind: &TypeKind{
				kind: &FnIns{
					Caller: builtin_caller_common,
					Params: []*ParamIns{
						{
							Decl: &Param{
								Ident: "bytes",
							},
							Kind: d.Kind,
						},
					},
					Result: d.Kind,
				},
			},
			Model: &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_trim",
			},
		}

	case "rtrim":
		return &Data{
			Kind: &TypeKind{
				kind: &FnIns{
					Caller: builtin_caller_common,
					Params: []*ParamIns{
						{
							Decl: &Param{
								Ident: "bytes",
							},
							Kind: d.Kind,
						},
					},
					Result: d.Kind,
				},
			},
			Model: &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_rtrim",
			},
		}

	case "split":
		return &Data{
			Kind: &TypeKind{
				kind: &FnIns{
					Caller: builtin_caller_common,
					Params: []*ParamIns{
						{
							Decl: &Param{
								Ident: "sub",
							},
							Kind: d.Kind,
						},
						{
							Decl: &Param{
								Ident: "n",
							},
							Kind: &TypeKind{kind: build_prim_type(types.TypeKind_INT)},
						},
					},
					Result: &TypeKind{
						kind: &Slc{
							Elem: d.Kind,
						},
					},
				},
			},
			Model: &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_split",
			},
		}

	case "replace":
		return &Data{
			Kind: &TypeKind{
				kind: &FnIns{
					Caller: builtin_caller_common,
					Params: []*ParamIns{
						{
							Decl: &Param{
								Ident: "sub",
							},
							Kind: d.Kind,
						},
						{
							Decl: &Param{
								Ident: "new",
							},
							Kind: d.Kind,
						},
						{
							Decl: &Param{
								Ident: "n",
							},
							Kind: &TypeKind{kind: build_prim_type(types.TypeKind_INT)},
						},
					},
					Result: d.Kind,
				},
			},
			Model: &CommonSubIdentExprModel{
				Expr:  d.Model,
				Ident: "_replace",
			},
		}

	default:
		e.push_err(ident, "obj_have_not_ident", ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_int_type_sub_ident(ident lex.Token) *Data {
	const kind = types.TypeKind_INT

	switch ident.Kind {
	case "MAX":
		c := constant.New_i64(int64(types.Max_of(kind)))
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	case "MIN":
		c := constant.New_i64(int64(types.Min_of(kind)))
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(ident, "type_have_not_ident", kind, ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_uint_type_sub_ident(ident lex.Token) *Data {
	const kind = types.TypeKind_UINT

	switch ident.Kind {
	case "MAX":
		c := constant.New_u64(uint64(types.Max_of(kind)))
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(ident, "type_have_not_ident", kind, ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_i8_type_sub_ident(ident lex.Token) *Data {
	const kind = types.TypeKind_I8
	const min = types.MIN_I8
	const max = types.MAX_I8

	switch ident.Kind {
	case "MAX":
		c := constant.New_i64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	case "MIN":
		c := constant.New_i64(min)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(ident, "type_have_not_ident", kind, ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_i16_type_sub_ident(ident lex.Token) *Data {
	const kind = types.TypeKind_I16
	const min = types.MIN_I16
	const max = types.MAX_I16

	switch ident.Kind {
	case "MAX":
		c := constant.New_i64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	case "MIN":
		c := constant.New_i64(min)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(ident, "type_have_not_ident", kind, ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_i32_type_sub_ident(ident lex.Token) *Data {
	const kind = types.TypeKind_I32
	const min = types.MIN_I32
	const max = types.MAX_I32

	switch ident.Kind {
	case "MAX":
		c := constant.New_i64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	case "MIN":
		c := constant.New_i64(min)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(ident, "type_have_not_ident", kind, ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_i64_type_sub_ident(ident lex.Token) *Data {
	const kind = types.TypeKind_I64
	const min = types.MIN_I64
	const max = types.MAX_I64

	switch ident.Kind {
	case "MAX":
		c := constant.New_i64(min)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	case "MIN":
		c := constant.New_i64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(ident, "type_have_not_ident", kind, ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_u8_type_sub_ident(ident lex.Token) *Data {
	const kind = types.TypeKind_U8
	const max = types.MAX_U8

	switch ident.Kind {
	case "MAX":
		c := constant.New_u64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(ident, "type_have_not_ident", kind, ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_u16_type_sub_ident(ident lex.Token) *Data {
	const kind = types.TypeKind_U16
	const max = types.MAX_U16

	switch ident.Kind {
	case "MAX":
		c := constant.New_u64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(ident, "type_have_not_ident", kind, ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_u32_type_sub_ident(ident lex.Token) *Data {
	const kind = types.TypeKind_U32
	const max = types.MAX_U32

	switch ident.Kind {
	case "MAX":
		c := constant.New_u64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(ident, "type_have_not_ident", kind, ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_u64_type_sub_ident(ident lex.Token) *Data {
	const kind = types.TypeKind_U64
	const max = types.MAX_U64

	switch ident.Kind {
	case "MAX":
		c := constant.New_u64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(ident, "type_have_not_ident", kind, ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_f32_type_sub_ident(ident lex.Token) *Data {
	const kind = types.TypeKind_F32
	const max = types.MAX_F32
	const min = types.MIN_F32

	switch ident.Kind {
	case "MAX":
		c := constant.New_f64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	case "MIN":
		c := constant.New_f64(min)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(ident, "type_have_not_ident", kind, ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_f64_type_sub_ident(ident lex.Token) *Data {
	const kind = types.TypeKind_F64
	const max = types.MAX_F64
	const min = types.MIN_F64

	switch ident.Kind {
	case "MAX":
		c := constant.New_f64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	case "MIN":
		c := constant.New_f64(min)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(ident, "type_have_not_ident", kind, ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_prim_type_sub_ident(kind lex.Token, ident lex.Token) *Data {
	switch kind.Kind {
	case types.TypeKind_INT:
		return e.eval_int_type_sub_ident(ident)

	case types.TypeKind_UINT:
		return e.eval_uint_type_sub_ident(ident)

	case types.TypeKind_I8:
		return e.eval_i8_type_sub_ident(ident)

	case types.TypeKind_I16:
		return e.eval_i16_type_sub_ident(ident)

	case types.TypeKind_I32:
		return e.eval_i32_type_sub_ident(ident)

	case types.TypeKind_I64:
		return e.eval_i64_type_sub_ident(ident)

	case types.TypeKind_U8:
		return e.eval_u8_type_sub_ident(ident)

	case types.TypeKind_U16:
		return e.eval_u16_type_sub_ident(ident)

	case types.TypeKind_U32:
		return e.eval_u32_type_sub_ident(ident)

	case types.TypeKind_U64:
		return e.eval_u64_type_sub_ident(ident)

	case types.TypeKind_F32:
		return e.eval_f32_type_sub_ident(ident)

	case types.TypeKind_F64:
		return e.eval_f64_type_sub_ident(ident)

	default:
		e.push_err(ident, "type_have_not_ident", kind.Kind, ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_type_sub_ident(d *Data, si *ast.SubIdentExpr) *Data {
	switch {
	case d.Kind.Enm() != nil:
		return e.eval_enum_sub_ident(d.Kind.Enm(), si.Ident)

	default:
		e.push_err(si.Ident, "type_not_support_sub_fields", d.Kind.To_str())
		return nil
	}
}

func (e *_Eval) eval_obj_sub_ident(d *Data, si *ast.SubIdentExpr) *Data {
	kind := d.Kind
	if d.Kind.Ptr() != nil {
		ptr := d.Kind.Ptr()
		if !ptr.Is_unsafe() {
			if !si.Is_self && !e.is_unsafe() {
				e.push_err(si.Ident, "unsafe_behavior_at_out_of_unsafe_scope")
			}
			kind = d.Kind.Ptr().Elem
		}
	} else if d.Kind.Ref() != nil {
		kind = d.Kind.Ref().Elem
	}

	switch {
	case d.Kind.Trt() != nil:
		return e.eval_trait_sub_ident(d, d.Kind.Trt(), si.Ident)

	case kind.Strct() != nil:
		s := kind.Strct()
		if is_instanced_struct(s) {
			used_reference_elem := d.Kind.Ref() != nil
			return e.eval_struct_sub_ident(d, kind.Strct(), si, used_reference_elem)
		}

	case kind.Slc() != nil:
		return e.eval_slice_sub_ident(d, si.Ident)

	case kind.Arr() != nil:
		return e.eval_array_sub_ident(d, si.Ident)

	case kind.Map() != nil:
		return e.eval_map_sub_ident(d, si.Ident)

	case kind.Prim() != nil:
		p := kind.Prim()
		switch p.kind {
		case types.TypeKind_STR:
			return e.eval_str_sub_ident(d, si.Ident)
		}
	}

	e.push_err(si.Ident, "obj_not_support_sub_fields", d.Kind.To_str())
	return nil
}

func (e *_Eval) eval_sub_ident(si *ast.SubIdentExpr) *Data {
	d := e.eval_expr_kind(si.Expr)
	if d == nil {
		return nil
	}

	if d.Decl {
		return e.eval_type_sub_ident(d, si)
	}

	return e.eval_obj_sub_ident(d, si)
}

func (e *_Eval) eval_tuple(tup *ast.TupleExpr) *Data {
	tup_t := &Tuple{}
	tup_t.Types = make([]*TypeKind, len(tup.Expr))
	
	model := &TupleExprModel{
		Datas: make([]*Data, len(tup.Expr)),
	}

	ok := true
	for i, expr := range tup.Expr {
		d := e.eval_expr_kind(expr)
		if d == nil {
			ok = false
			continue
		}
		tup_t.Types[i] = d.Kind
		model.Datas[i] = d
	}

	if !ok {
		return nil
	}

	return &Data{
		Kind:  &TypeKind{kind: tup_t},
		Model: model,
	}
}

func (e *_Eval) eval_map(m *Map, lit *ast.BraceLit) *Data {
	model := &MapExprModel{
		Key_kind: m.Key,
		Val_kind: m.Val,
	}

	for _, expr := range lit.Exprs {
		switch expr.(type) {
		case *ast.KeyValPair:
			// Ok.

		default:
			e.push_err(lit.Token, "invalid_syntax")
			return nil
		}

		pair := expr.(*ast.KeyValPair)

		key := e.eval_expr_kind(pair.Key)
		if key == nil {
			return nil
		}

		val := e.eval_expr_kind(pair.Val)
		if val == nil {
			return nil
		}

		e.s.check_assign_type(m.Key, key, pair.Colon, true)
		e.s.check_assign_type(m.Val, val, pair.Colon, true)

		model.Entries = append(model.Entries, &KeyValPairExprModel{
			Key: key.Model,
			Val: val.Model,
		})
	}

	return &Data{
		Mutable:    true,
		Lvalue:     false,
		Variadiced: false,
		Constant:   nil,
		Decl:       false,
		Kind:       &TypeKind{kind: m},
		Model:      model,
	}
}

func (e *_Eval) eval_brace_lit(lit *ast.BraceLit) *Data {
	switch {
	case e.prefix == nil:
		e.push_err(lit.Token, "invalid_syntax")
		return nil

	case e.prefix.Map() != nil:
		return e.eval_map(e.prefix.Map(), lit)

	case e.prefix.Strct() != nil:
		return e.eval_struct_lit_explicit(e.prefix.Strct(), lit.Exprs, lit.Token)
	
	default:
		e.push_err(lit.Token, "invalid_syntax")
		return nil
	}
}

func (e *_Eval) eval_anon_fn(decl *ast.FnDecl) *Data {
	tc := _TypeChecker{
		s:      e.s,
		lookup: e.lookup,
	}
	ins := tc.build_fn(decl)

	switch e.lookup.(type) {
	case *_ScopeChecker:
		sc := e.lookup.(*_ScopeChecker)
		scc := sc.new_child_checker()
		scc.labels = new([]*_ScopeLabel)
		scc.gotos =  new([]*_ScopeGoto)
		scc.owner = nil
		scc.child_index = 0
		scc.it = 0
		scc.cse = 0
		scc.owner = ins
		e.s.check_fn_ins_sc(ins, scc)

	default:
		e.s.check_fn_ins(ins)
	}

	return &Data{
		Kind:  &TypeKind{kind: ins},
		Model: &AnonFnExprModel{
			Func:   ins,
			Global: e.is_global(),
		},
	}
}

func (e *_Eval) eval_binop(op *ast.BinopExpr) *Data {
	bs := _BinopSolver{
		e: e,
	}
	return bs.solve(op)
}

func (e *_Eval) eval_expr_kind(kind ast.ExprData) *Data {
	var d *Data

	switch kind.(type) {
	case *ast.LitExpr:
		d = e.eval_lit(kind.(*ast.LitExpr))

	case *ast.IdentExpr:
		d = e.eval_ident(kind.(*ast.IdentExpr))

	case *ast.UnaryExpr:
		d = e.eval_unary(kind.(*ast.UnaryExpr))

	case *ast.VariadicExpr:
		d = e.eval_variadic(kind.(*ast.VariadicExpr))

	case *ast.UnsafeExpr:
		d = e.eval_unsafe(kind.(*ast.UnsafeExpr))

	case *ast.SliceExpr:
		d = e.eval_slice_expr(kind.(*ast.SliceExpr))

	case *ast.IndexingExpr:
		d = e.eval_indexing(kind.(*ast.IndexingExpr))

	case *ast.SlicingExpr:
		d = e.eval_slicing(kind.(*ast.SlicingExpr))

	case *ast.CastExpr:
		d = e.eval_cast(kind.(*ast.CastExpr))

	case *ast.NsSelectionExpr:
		d = e.eval_ns_selection(kind.(*ast.NsSelectionExpr))

	case *ast.StructLit:
		d = e.eval_struct_lit(kind.(*ast.StructLit))
	
	case *ast.Type:
		d = e.eval_type(kind.(*ast.Type))

	case *ast.FnCallExpr:
		d = e.eval_fn_call(kind.(*ast.FnCallExpr))

	case *ast.SubIdentExpr:
		d = e.eval_sub_ident(kind.(*ast.SubIdentExpr))

	case *ast.TupleExpr:
		d = e.eval_tuple(kind.(*ast.TupleExpr))

	case *ast.BraceLit:
		d = e.eval_brace_lit(kind.(*ast.BraceLit))

	case *ast.FnDecl:
		d = e.eval_anon_fn(kind.(*ast.FnDecl))

	case *ast.BinopExpr:
		d = e.eval_binop(kind.(*ast.BinopExpr))

	default:
		d = nil
	}

	if d == nil {
		return nil
	}

	if d.Kind == nil {
		return d
	}

	if d.Cast_kind == nil && d.Is_const() && !d.Is_rune && d.Kind.Prim() != nil {
		switch {
		case d.Constant.Is_i64():
			if int_assignable(types.TypeKind_INT, d) {
				d.Kind.kind = build_prim_type(types.TypeKind_INT)
			}

		case d.Constant.Is_u64():
			if int_assignable(types.TypeKind_UINT, d) {
				d.Kind.kind = build_prim_type(types.TypeKind_UINT)
			}
		}
	}

	if d.Cast_kind == nil && !d.Variadiced && !d.Lvalue && !d.Is_const() && d.Kind.Prim() != nil && types.Is_num(d.Kind.Prim().kind) {
		d.Cast_kind = d.Kind
	}

	apply_cast_kind(d)
	return d
}

// Returns value data of evaluated expression.
// Returns nil if error occurs.
func (e *_Eval) eval(expr *ast.Expr) *Data {
	d := e.eval_expr_kind(expr.Kind)
	if d == nil {
		return nil
	}

	switch {
	case d.Kind.Fnc() != nil:
		f := d.Kind.Fnc()
		if f.Is_builtin() {
			break
		}

		if len(f.Generics) != len(f.Decl.Generics) {
			e.s.push_err(expr.Token, "has_generics")
		}

		if f.Decl.Is_method() {
			e.s.push_err(expr.Token, "method_not_invoked")
		}
	}

	return d
}

// Returns value data of evaluated expression.
// Returns nil if error occurs.
// Accepts decls as invalid expression.
func (e *_Eval) eval_expr(expr *ast.Expr) *Data {
	d := e.eval(expr)
	switch {
	case d == nil:
		return nil

	case d.Decl:
		e.push_err(expr.Token, "invalid_expr")
		return nil

	default:
		return d
	}
}

func is_ok_for_shifting(d *Data) bool {
	prim := d.Kind.Prim()
	if prim == nil || !types.Is_int(prim.To_str()) {
		return false
	}

	if !d.Is_const() {
		return true
	}

	switch {
	case d.Constant.Is_i64():
		return d.Constant.Read_i64() >= 0

	case d.Constant.Is_u64():
		return true

	default:
		return false
	}
}

type _BinopSolver struct {
	e  *_Eval
	l  *Data
	r  *Data
	op lex.Token
}

func (bs *_BinopSolver) check_type_compatibility() bool {
	tcc := _TypeCompatibilityChecker{
		s:           bs.e.s,
		error_token: bs.op,
		dest:        bs.l.Kind,
		src:         bs.r.Kind,
		deref:       true,
	}
	return tcc.check()
}

func (bs *_BinopSolver) eval_nil() *Data {
	if !is_nil_compatible(bs.r.Kind) {
		bs.e.push_err(bs.op, "incompatible_types", lex.KND_NIL, bs.r.Kind.To_str())
		return nil
	}

	switch bs.op.Kind {
	case lex.KND_EQS, lex.KND_NOT_EQ:
		return &Data{
			Kind: &TypeKind{
				kind: build_prim_type(types.TypeKind_BOOL),
			},
		}

	default:
		bs.e.push_err(bs.op, "operator_not_for_juletype", bs.op.Kind, lex.KND_NIL)
		return nil
	}
}

func (bs *_BinopSolver) eval_ptr() *Data {
	if !bs.check_type_compatibility() {
		bs.e.push_err(bs.op, "incompatible_types", bs.l.Kind.To_str(), bs.r.Kind.To_str())
		return nil
	}

	switch bs.op.Kind {
	case lex.KND_EQS,
		lex.KND_NOT_EQ,
		lex.KND_LT,
		lex.KND_GT,
		lex.KND_LESS_EQ,
		lex.KND_GREAT_EQ:
		return &Data{
			Kind: &TypeKind{
				kind: build_prim_type(types.TypeKind_BOOL),
			},
		}

	default:
		bs.e.push_err(bs.op, "operator_not_for_juletype", bs.op.Kind, bs.l.Kind.To_str())
		return nil
	}
}

func (bs *_BinopSolver) eval_arr() *Data {
	if !bs.check_type_compatibility() {
		bs.e.push_err(bs.op, "incompatible_types", bs.l.Kind.To_str(), bs.r.Kind.To_str())
		return nil
	}

	switch bs.op.Kind {
	case lex.KND_EQS, lex.KND_NOT_EQ:
		return &Data{
			Kind: &TypeKind{
				kind: build_prim_type(types.TypeKind_BOOL),
			},
		}

	default:
		bs.e.push_err(bs.op, "operator_not_for_juletype", bs.op.Kind, bs.l.Kind.To_str())
		return nil
	}
}

func (bs *_BinopSolver) eval_slc() *Data {
	if !bs.check_type_compatibility() {
		bs.e.push_err(bs.op, "incompatible_types", bs.l.Kind.To_str(), bs.r.Kind.To_str())
		return nil
	}

	switch bs.op.Kind {
	case lex.KND_EQS, lex.KND_NOT_EQ:
		return &Data{
			Kind: &TypeKind{
				kind: build_prim_type(types.TypeKind_BOOL),
			},
		}

	default:
		bs.e.push_err(bs.op, "operator_not_for_juletype", bs.op.Kind, bs.l.Kind.To_str())
		return nil
	}
}

func (bs *_BinopSolver) eval_fn() *Data {
	if !bs.r.Kind.Is_nil() {
		bs.e.push_err(bs.op, "incompatible_types", bs.l.Kind.To_str(), bs.r.Kind.To_str())
		return nil
	}

	switch bs.op.Kind {
	case lex.KND_EQS, lex.KND_NOT_EQ:
		return &Data{
			Kind: &TypeKind{
				kind: build_prim_type(types.TypeKind_BOOL),
			},
		}

	default:
		bs.e.push_err(bs.op, "operator_not_for_juletype", bs.op.Kind, bs.l.Kind.To_str())
		return nil
	}
}

func (bs *_BinopSolver) eval_struct() *Data {
	if !bs.check_type_compatibility() {
		bs.e.push_err(bs.op, "incompatible_types", bs.l.Kind.To_str(), bs.r.Kind.To_str())
		return nil
	}

	switch bs.op.Kind {
	case lex.KND_EQS, lex.KND_NOT_EQ:
		return &Data{
			Kind: &TypeKind{
				kind: build_prim_type(types.TypeKind_BOOL),
			},
		}

	default:
		bs.e.push_err(bs.op, "operator_not_for_juletype", bs.op.Kind, bs.l.Kind.To_str())
		return nil
	}
}

func (bs *_BinopSolver) eval_trait() *Data {
	if !bs.check_type_compatibility() {
		bs.e.push_err(bs.op, "incompatible_types", bs.l.Kind.To_str(), bs.r.Kind.To_str())
		return nil
	}

	switch bs.op.Kind {
	case lex.KND_EQS, lex.KND_NOT_EQ:
		return &Data{
			Kind: &TypeKind{
				kind: build_prim_type(types.TypeKind_BOOL),
			},
		}

	default:
		bs.e.push_err(bs.op, "operator_not_for_juletype", bs.op.Kind, bs.l.Kind.To_str())
		return nil
	}
}

func (bs *_BinopSolver) eval_any() *Data {
	switch bs.op.Kind {
	case lex.KND_EQS, lex.KND_NOT_EQ:
		return &Data{
			Kind: &TypeKind{
				kind: build_prim_type(types.TypeKind_BOOL),
			},
		}

	default:
		bs.e.push_err(bs.op, "operator_not_for_juletype", bs.op.Kind, types.TypeKind_ANY)
		return nil
	}
}

func (bs *_BinopSolver) eval_bool() *Data {
	if !bs.check_type_compatibility() {
		bs.e.push_err(bs.op, "incompatible_types", bs.l.Kind.To_str(), bs.r.Kind.To_str())
		return nil
	}

	switch bs.op.Kind {
	case lex.KND_EQS, lex.KND_NOT_EQ, lex.KND_DBL_AMPER, lex.KND_DBL_VLINE:
		return bs.l

	default:
		bs.e.push_err(bs.op, "operator_not_for_juletype", bs.op.Kind, bs.l.Kind.To_str())
		return nil
	}
}

func (bs *_BinopSolver) eval_str() *Data {
	rk := bs.r.Kind.To_str()
	if rk != types.TypeKind_STR {
		bs.e.push_err(bs.op, "incompatible_types", bs.l.Kind.To_str(), rk)
		return nil
	}

	switch bs.op.Kind {
	case lex.KND_PLUS:
		return bs.l

	case lex.KND_EQS, lex.KND_NOT_EQ:
		return &Data{
			Kind: &TypeKind{
				kind: build_prim_type(types.TypeKind_BOOL),
			},
		}

	default:
		bs.e.push_err(bs.op, "operator_not_for_juletype", bs.op.Kind, types.TypeKind_ANY)
		return nil
	}
}

func (bs *_BinopSolver) set_type_to_greater() {
	if bs.l.Is_const() && bs.r.Is_const() || !bs.l.Is_const() && !bs.r.Is_const() {
		lk := bs.l.Kind.To_str()
		rk := bs.r.Kind.To_str()
		if types.Is_greater(rk, lk) {
			bs.l.Kind = bs.r.Kind
		}
		return
	}

	if bs.l.Is_const() {
		bs.l.Kind = bs.r.Kind
		return
	}

	bs.r.Kind = bs.l.Kind
}

func (bs *_BinopSolver) mod() {
	check := func(d *Data) {
		if !d.Is_const() {
			if d.Kind.Prim() == nil || !types.Is_int(d.Kind.Prim().kind) {
				bs.e.push_err(bs.op, "modulo_with_not_int")
			}
			return
		}

		switch {
		case sig_assignable(types.TypeKind_I64, d):
			d.Constant.Set_i64(d.Constant.As_i64())

		case unsig_assignable(types.TypeKind_U64, d):
			d.Constant.Set_u64(d.Constant.As_u64())

		default:
			bs.e.push_err(bs.op, "modulo_with_not_int")
		}
	}

	check(bs.l)
	check(bs.r)
}

func (bs *_BinopSolver) eval_float() *Data {
	lk := bs.l.Kind.To_str()
	rk := bs.r.Kind.To_str()
	if !types.Is_num(lk) || !types.Is_num(rk) {
		bs.e.push_err(bs.op, "incompatible_types", lk, rk)
		return nil
	}

	// Logicals.
	switch bs.op.Kind {
	case lex.KND_EQS,
		lex.KND_NOT_EQ,
		lex.KND_LT,
		lex.KND_GT,
		lex.KND_GREAT_EQ,
		lex.KND_LESS_EQ:
		return &Data{
			Kind: &TypeKind{
				kind: build_prim_type(types.TypeKind_BOOL),
			},
		}
	}

	// Arithmetics.
	switch bs.op.Kind {
	case lex.KND_PLUS,
		lex.KND_MINUS,
		lex.KND_STAR,
		lex.KND_SOLIDUS:
		bs.set_type_to_greater()
		return bs.l

	case lex.KND_PERCENT:
		if !types.Is_int(rk) {
			bs.e.push_err(bs.op, "incompatible_types", lk, rk)
			return nil
		}
		bs.mod()
		return bs.r

	default:
		bs.e.push_err(bs.op, "operator_not_for_float", bs.op.Kind)
		return nil
	}
}

func (bs *_BinopSolver) eval_unsig_int() *Data {
	lk := bs.l.Kind.To_str()
	rk := bs.r.Kind.To_str()
	if !types.Is_num(lk) || !types.Is_num(rk) {
		bs.e.push_err(bs.op, "incompatible_types", lk, rk)
		return nil
	}

	// Logicals.
	switch bs.op.Kind {
	case lex.KND_EQS,
		lex.KND_NOT_EQ,
		lex.KND_LT,
		lex.KND_GT,
		lex.KND_GREAT_EQ,
		lex.KND_LESS_EQ:
		return &Data{
			Kind: &TypeKind{
				kind: build_prim_type(types.TypeKind_BOOL),
			},
		}
	}

	// Arithmetics.
	switch bs.op.Kind {
	case lex.KND_PLUS,
		lex.KND_MINUS,
		lex.KND_STAR,
		lex.KND_SOLIDUS,
		lex.KND_AMPER,
		lex.KND_VLINE,
		lex.KND_CARET:
		bs.set_type_to_greater()
		return bs.l

	case lex.KND_PERCENT:
		bs.mod()
		bs.set_type_to_greater()
		return bs.l

	case lex.KND_LSHIFT, lex.KND_RSHIFT:
		if !is_ok_for_shifting(bs.r) {
			bs.e.push_err(bs.op, "bitshift_must_unsigned")
			return nil
		}

		if bs.l.Is_const() {
			bs.l.Kind = bs.r.Kind
		}

		return bs.l

	default:
		bs.e.push_err(bs.op, "operator_not_for_uint", bs.op.Kind)
		return nil
	}
}

func (bs *_BinopSolver) eval_sig_int() *Data {
	lk := bs.l.Kind.To_str()
	rk := bs.r.Kind.To_str()
	if !types.Is_num(lk) || !types.Is_num(rk) {
		bs.e.push_err(bs.op, "incompatible_types", lk, rk)
		return nil
	}

	// Logicals.
	switch bs.op.Kind {
	case lex.KND_EQS,
		lex.KND_NOT_EQ,
		lex.KND_LT,
		lex.KND_GT,
		lex.KND_GREAT_EQ,
		lex.KND_LESS_EQ:
		return &Data{
			Kind: &TypeKind{
				kind: build_prim_type(types.TypeKind_BOOL),
			},
		}
	}

	// Arithmetics.
	switch bs.op.Kind {
	case lex.KND_PLUS,
		lex.KND_MINUS,
		lex.KND_STAR,
		lex.KND_SOLIDUS,
		lex.KND_AMPER,
		lex.KND_VLINE,
		lex.KND_CARET:
		bs.set_type_to_greater()
		return bs.l

	case lex.KND_PERCENT:
		bs.mod()
		bs.set_type_to_greater()
		return bs.l

	case lex.KND_LSHIFT, lex.KND_RSHIFT:
		if !is_ok_for_shifting(bs.r) {
			bs.e.push_err(bs.op, "bitshift_must_unsigned")
			return nil
		}
		
		return bs.l

	default:
		bs.e.push_err(bs.op, "operator_not_for_int", bs.op.Kind)
		return nil
	}
}

func (bs *_BinopSolver) eval_prim() *Data {
	prim := bs.l.Kind.Prim()
	switch {
	case prim.Is_any():
		return bs.eval_any()
	
	case prim.Is_bool():
		return bs.eval_bool()

	case prim.Is_str():
		return bs.eval_str()
	}

	rprim := bs.r.Kind.Prim()
	if rprim == nil {
		bs.e.push_err(bs.op, "incompatible_types", prim.To_str(), bs.r.Kind.To_str())
		return nil
	}

	lk := prim.To_str()
	switch {
	case types.Is_float(lk):
		return bs.eval_float()

	case types.Is_unsig_int(lk):
		return bs.eval_unsig_int()

	case types.Is_sig_int(lk):
		return bs.eval_sig_int()

	default:
		return nil
	}
}

func (bs *_BinopSolver) eval() *Data {
	if bs.l.Kind.Enm() != nil {
		bs.l.Kind = bs.l.Kind.Enm().Kind.Kind
	}
	if bs.r.Kind.Enm() != nil {
		bs.r.Kind = bs.r.Kind.Enm().Kind.Kind
	}

	switch {
	case bs.l.Kind.Is_void():
		bs.e.push_err(bs.op, "operator_not_for_juletype", bs.op.Kind, "void")
		return nil

	case bs.l.Kind.Is_nil():
		return bs.eval_nil()

	case bs.l.Kind.Ptr() != nil:
		return bs.eval_ptr()

	case bs.l.Kind.Arr() != nil:
		return bs.eval_arr()

	case bs.l.Kind.Slc() != nil:
		return bs.eval_slc()

	case bs.l.Kind.Fnc() != nil:
		return bs.eval_fn()
		
	case bs.l.Kind.Trt() != nil || bs.r.Kind.Trt() != nil:
		if bs.r.Kind.Trt() != nil {
			bs.l, bs.r = bs.r, bs.l
		}
		return bs.eval_trait()

	case bs.l.Kind.Strct() != nil:
		return bs.eval_struct()


	case bs.l.Kind.Prim() != nil:
		return bs.eval_prim()

	default:
		return nil
	}
}

func (bs *_BinopSolver) assign_shift(d *Data, r float64) {
	switch {
	case r <= 6:
		d.Kind.Prim().kind = types.TypeKind_I8
		d.Constant.Set_i64(d.Constant.As_i64())

	case r <= 7:
		d.Kind.Prim().kind = types.TypeKind_U8
		d.Constant.Set_u64(d.Constant.As_u64())

	case r <= 14:
		d.Kind.Prim().kind = types.TypeKind_I16
		d.Constant.Set_i64(d.Constant.As_i64())

	case r <= 15:
		d.Kind.Prim().kind = types.TypeKind_U16
		d.Constant.Set_u64(d.Constant.As_u64())

	case r <= 30:
		d.Kind.Prim().kind = types.TypeKind_I32
		d.Constant.Set_i64(d.Constant.As_i64())

	case r <= 31:
		d.Kind.Prim().kind = types.TypeKind_U32
		d.Constant.Set_u64(d.Constant.As_u64())

	case r <= 62:
		d.Kind.Prim().kind = types.TypeKind_I64
		d.Constant.Set_i64(d.Constant.As_i64())

	case r <= 63:
		d.Kind.Prim().kind = types.TypeKind_U64
		d.Constant.Set_u64(d.Constant.As_u64())

	case r <= 127:
		d.Kind.Prim().kind = types.TypeKind_F32
		d.Constant.Set_f64(d.Constant.As_f64())

	default:
		d.Kind.Prim().kind = types.TypeKind_F64
		d.Constant.Set_f64(d.Constant.As_f64())
	}
}

func (bs *_BinopSolver) solve_const(d *Data) {
	switch {
	case d == nil:
		return
		
	case !bs.l.Is_const() || !bs.r.Is_const():
		d.Constant = nil
		return
	}

	switch bs.op.Kind {
	case lex.KND_EQS:
		d.Constant = constant.New_bool(bs.l.Constant.Eqs(*bs.r.Constant))

	case lex.KND_NOT_EQ:
		d.Constant = constant.New_bool(!bs.l.Constant.Eqs(*bs.r.Constant))

	case lex.KND_DBLCOLON:
		d.Constant = constant.New_bool(bs.l.Constant.Or(*bs.r.Constant))

	case lex.KND_DBL_AMPER:
		d.Constant = constant.New_bool(bs.l.Constant.And(*bs.r.Constant))

	case lex.KND_GT:
		d.Constant = constant.New_bool(bs.l.Constant.Gt(*bs.r.Constant))

	case lex.KND_LT:
		d.Constant = constant.New_bool(bs.l.Constant.Lt(*bs.r.Constant))

	case lex.KND_GREAT_EQ:
		d.Constant = constant.New_bool(bs.l.Constant.Gt(*bs.r.Constant) || bs.l.Constant.Eqs(*bs.r.Constant))

	case lex.KND_LESS_EQ:
		d.Constant = constant.New_bool(bs.l.Constant.Lt(*bs.r.Constant) || bs.l.Constant.Eqs(*bs.r.Constant))

	case lex.KND_PLUS:
		_ = bs.l.Constant.Add(*bs.r.Constant)
		d.Constant = bs.l.Constant

	case lex.KND_MINUS:
		_ = bs.l.Constant.Sub(*bs.r.Constant)
		d.Constant = bs.l.Constant

	case lex.KND_STAR:
		_ = bs.l.Constant.Mul(*bs.r.Constant)
		d.Constant = bs.l.Constant

	case lex.KND_SOLIDUS:
		ok := bs.l.Constant.Div(*bs.r.Constant)
		if !ok && bs.r.Constant.As_f64() == 0 {
			bs.e.push_err(bs.op, "divide_by_zero")
		}
		d.Constant = bs.l.Constant

	case lex.KND_PERCENT:
		ok := bs.l.Constant.Mod(*bs.r.Constant)
		if !ok && bs.r.Constant.As_f64() == 0 {
			bs.e.push_err(bs.op, "divide_by_zero")
		}
		d.Constant = bs.l.Constant

	case lex.KND_COLON:
		_ = bs.l.Constant.Bitwise_or(*bs.r.Constant)
		d.Constant = bs.l.Constant

	case lex.KND_AMPER:
		_ = bs.l.Constant.Bitwise_and(*bs.r.Constant)
		d.Constant = bs.l.Constant

	case lex.KND_CARET:
		_ = bs.l.Constant.Xor(*bs.r.Constant)
		d.Constant = bs.l.Constant

	case lex.KND_LSHIFT:
		_ = bs.l.Constant.Lshift(*bs.r.Constant)
		d.Constant = bs.l.Constant
		bs.assign_shift(d, bs.r.Constant.As_f64())

	case lex.KND_RSHIFT:
		_ = bs.l.Constant.Rshift(*bs.r.Constant)
		d.Constant = bs.l.Constant
		bs.assign_shift(d, bs.r.Constant.As_f64())
	}

	d.Model = d.Constant
}

func (bs *_BinopSolver) post_const(d *Data) {
	if d == nil {
		return
	}
	if !d.Is_const() {
		return
	}

	normalize_bitsize(d)
}

func (bs *_BinopSolver) solve_explicit(l *Data, r *Data) *Data {
	bs.l, bs.r = l, r
	d := bs.eval()
	bs.l, bs.r = l, r // Save normal order

	bs.solve_const(d)
	bs.post_const(d)

	if d != nil && !d.Is_const() {
		d.Model = &BinopExprModel{
			L: l.Model,
			R: r.Model,
			Op: bs.op.Kind,
		}
	}

	if l.Cast_kind != nil && r.Cast_kind == nil {
		d.Cast_kind = l.Cast_kind
	} else if r.Cast_kind != nil && l.Cast_kind == nil {
		d.Cast_kind = r.Cast_kind
	}

	return d
}

func (bs *_BinopSolver) solve(op *ast.BinopExpr) *Data {
	l := bs.e.eval_expr_kind(op.L)
	if l == nil {
		return nil
	}

	r := bs.e.eval_expr_kind(op.R)
	if r == nil {
		return nil
	}

	bs.op = op.Op

	d := bs.solve_explicit(l, r)

	// Save rune type.
	if d != nil && l.Is_rune && r.Is_rune {
		d.Is_rune = true
	}

	return d
}
