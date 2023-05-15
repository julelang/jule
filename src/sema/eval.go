package sema

import (
	"strconv"
	"strings"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/constant"
	"github.com/julelang/jule/constant/lit"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

// Value data.
type Data struct {
	Kind       *TypeKind
	Mutable    bool
	Lvalue     bool
	Variadiced bool
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

func (e *_Eval) lit_str(lit *ast.LitExpr) *Data {
	constant := constant.New_str(lit.Value[1:len(lit.Value)-1])
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
	
	rs := lit.To_rune([]byte(l.Value))
	rs = rs[2:] // Skip hexadecimal prefix.
	r, _ := strconv.ParseInt(rs, 16, 64)

	data := &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: constant.New_i64(r),
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

	data.Model = data.Constant
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


	data := &Data{
		Lvalue:  false,
		Mutable: false,
		Decl:    false,
	}

	var value any = nil
	const BIT_SIZE = 0b01000000
	sig, err := strconv.ParseInt(lit, base, BIT_SIZE)
	if err == nil {
		value = sig
		data.Constant = constant.New_i64(sig)
	} else {
		unsig, _ := strconv.ParseUint(lit, base, BIT_SIZE)
		data.Constant = constant.New_u64(unsig)
		value = unsig
	}

	data.Kind = &TypeKind{
		kind: build_prim_type(kind_by_bitsize(value)),
	}

	// TODO: Implement normalization.

	data.Model = data.Constant
	return data
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

	return nil
}

func (e *_Eval) eval_enum(enm *Enum, error_token lex.Token) *Data {
	if !e.s.is_accessible_define(enm.Public, enm.Token) {
		e.push_err(error_token, "ident_not_exist", enm.Ident)
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
		e.push_err(error_token, "ident_not_exist", s.Decl.Ident)
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

func (e *_Eval) eval_fn(f *Fn, error_token lex.Token) *Data {
	if !e.s.is_accessible_define(f.Public, f.Token) {
		e.push_err(error_token, "ident_not_exist", f.Ident)
		return nil
	}

	ins := f.instance()
	return &Data{
		Lvalue:   false,
		Mutable:  false,
		Constant: nil,
		Decl:     false,
		Kind:     &TypeKind{
			kind: ins,
		},
		Model:    ins,
	}
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

	// Check cross illegal cycle.
	for _, d := range v.Depends {
		if d == e.owner {
			e.push_err(decl_token, "illegal_cross_cycle", e.owner.Ident, decl_token.Kind)
			return false
		}
	}

	e.owner.Depends = append(e.owner.Depends, v)
	return true
}

func (e *_Eval) eval_var(v *Var, error_token lex.Token) *Data {
	if !e.s.is_accessible_define(v.Public, v.Token) {
		e.push_err(error_token, "ident_not_exist", v.Ident)
		return nil
	}

	ok := e.check_illegal_cycles(v, error_token)
	if !ok {
		return nil
	}

	if v.Value.Data == nil {
		return nil
	}

	d := &Data{
		Lvalue:   !v.Constant,
		Mutable:  v.Mutable,
		Decl:     false,
		Constant: v.Value.Data.Constant,
		Kind:     v.Kind.Kind,
		Model:    v,
	}

	return d
}

func (e *_Eval) eval_type_alias(ta *TypeAlias, error_token lex.Token) *Data {
	if !e.s.is_accessible_define(ta.Public, ta.Token) {
		e.push_err(error_token, "ident_not_exist", ta.Ident)
		return nil
	}

	kind := ta.Kind.Kind.kind
	switch kind.(type) {
	case *StructIns:
		return e.eval_struct(kind.(*StructIns), error_token)

	case *Enum:
		return e.eval_enum(kind.(*Enum), error_token)

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
	d.Model = &UnaryExprModel{
		Expr: d.Model,
		Op:   lex.KND_STAR,
	}
	return d
}

func (e *_Eval) eval_unary_amper(d *Data) *Data {
	switch d.Model.(type) {
	case *StructLitExprModel:
		d.Model = &AllocStructLitExprModel{
			Lit: d.Model.(*StructLitExprModel),
		}

	default:
		switch {
		case d.Kind.Ref() != nil:
			d.Model = &GetRefPtrExprModel{
				Expr: d.Model,
			}

		case can_get_ptr(d):
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
	data := e.eval_expr_kind(u.Expr)
	if data == nil {
		return nil
	}

	switch u.Op.Kind {
	case lex.KND_MINUS:
		data = e.eval_unary_minus(data)

	case lex.KND_PLUS:
		data = e.eval_unary_plus(data)

	case lex.KND_CARET:
		data = e.eval_unary_caret(data)

	case lex.KND_EXCL:
		data = e.eval_unary_excl(data)

	case lex.KND_STAR:
		data = e.eval_unary_star(data, u.Op)

	case lex.KND_AMPER:
		data = e.eval_unary_amper(data)

	default:
		data = nil
	}

	if data == nil {
		e.push_err(u.Op, "invalid_expr_unary_operator", u.Op.Kind)
	} else if data.Is_const() {
		data.Model = data.Constant
	}

	return data
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

	for i, elem := range s.Elems {
		d := e.eval_expr_kind(elem)
		if d == nil {
			continue
		}

		e.s.check_assign_type(arr.Elem, d, s.Token, true)
		model.Elems[i] = d.Model
	}

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

	for i, elem := range s.Elems {
		d := e.eval_expr_kind(elem)
		if d == nil {
			continue
		}

		e.s.check_assign_type(slc.Elem, d, s.Token, true)
		model.Elems[i] = d.Model
	}

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

	d.Kind = ptr.Elem
}

func (e *_Eval) indexing_arr(d *Data, index *Data, i *ast.IndexingExpr) {
	arr := d.Kind.Arr()
	d.Kind = arr.Elem
	e.check_integer_indexing_by_data(index, i.Token)
}

func (e *_Eval) indexing_slc(d *Data, index *Data, i *ast.IndexingExpr) {
	slc := d.Kind.Slc()
	d.Kind = slc.Elem
	e.check_integer_indexing_by_data(index, i.Token)
}

func (e *_Eval) indexing_map(d *Data, index *Data, i *ast.IndexingExpr) {
	if index == nil {
		return
	}

	m := d.Kind.Map()
	e.s.check_type_compatibility(m.Key, index.Kind, i.Token, true)

	d.Kind = m.Val
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
	d.Kind.kind = &Slc{Elem: d.Kind.Slc().Elem}
}

func (e *_Eval) slicing_slc(d *Data) {
	d.Lvalue = false
}

func (e *_Eval) slicing_str(d *Data, l *Data, r *Data) {
	d.Lvalue = false
	if !d.Is_const() {
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
		prim := d.Kind.Prim()
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
		prim := d.Kind.Prim()
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

	if d != nil {
		d.Lvalue = is_lvalue(t)
		d.Mutable = is_mut(t)
		d.Decl = false
		d.Model = &CastingExprModel{
			Kind:     t,
			Expr:     d.Model,
			ExprKind: d.Kind,
		}
		d.Kind = t
	}
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

	return e.eval_cast_by_type_n_data(t.Kind, d, c.Kind.Token)
}

func (e *_Eval) eval_ns_selection(s *ast.NsSelectionExpr) *Data {
	path := build_link_path_by_tokens(s.Ns)
	pkg := e.lookup.Select_package(func(p *Package) bool {
		return p.Link_path == path
	})

	if pkg == nil {
		e.push_err(s.Ident, "namespace_not_exist", s.Ident.Kind)
		return nil
	}

	lookup := e.lookup
	e.lookup = pkg

	const CPP_LINKED = false
	def := e.get_def(s.Ident.Kind, CPP_LINKED)
	d := e.eval_def(def, s.Ident)

	e.lookup = lookup

	return d
}

func (e *_Eval) is_instanced_struct(s *StructIns) bool {
	return len(s.Decl.Generics) == len(s.Generics)
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

	ok = e.s.check_generic_quantity(len(s.Decl.Generics), len(s.Generics), lit.Kind.Token)
	if !ok {
		return nil
	}
	// NOTE: Instance already checked (just fields) if generic quantity passes.

	slc := _StructLitChecker{
		e:           e,
		error_token: lit.Kind.Token,
		s:           s,
	}
	slc.check(lit.Exprs)

	return &Data{
		Mutable: true,
		Kind:    t.Kind,
		Model:   &StructLitExprModel{
			Strct: s,
			Args:  slc.args,
		},
	}
}

func (e *_Eval) eval_type(t *ast.Type) *Data {
	tk := build_type(t)
	ok := e.s.check_type(tk, e.lookup)
	if !ok {
		return nil
	}

	return &Data{
		Decl: true,
		Kind: tk.Kind,
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
			goto _ret
		}

		if arg != nil {
			_ = e.eval_cast_by_type_n_data(d.Kind, arg, fc.Args[0].Token)
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

func (e *_Eval) call_fn(fc *ast.FnCallExpr, d *Data) *Data {
	f := d.Kind.Fnc()
	if f.Is_builtin() {
		fcac := _FnCallArgChecker{
			e:                  e,
			f:                  f,
			args:               fc.Args,
			dynamic_annotation: false,
			error_token:        fc.Token,
		}
		_ = fcac.check()

		model := &FnCallExprModel{
			Func: f,
			IsCo: fc.Concurrent,
			Expr: d.Model,
			Args: fcac.arg_models,
		}

		if f.Result == nil {
			d = build_void_data()
		} else {
			d = &Data{
				Kind: f.Result,
			}
		}

		d.Model = model
		return d
	}

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

	f.Decl.append_instance(f)

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

	d.Model = &FnCallExprModel{
		Func: f,
		IsCo: fc.Concurrent,
		Expr: call_model,
		Args: fcac.arg_models,
	}

	return d
}

func (e *_Eval) eval_fn_call(fc *ast.FnCallExpr) *Data {
	d := e.eval_expr_kind(fc.Expr.Kind)
	if d == nil {
		return nil
	}

	if d.Decl {
		if d.Kind.Prim() != nil {
			return e.call_type_fn(fc, d)
		}
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
		Kind:     enm.Kind.Kind,
	}

	item := enm.Find_item(ident.Kind)
	if item == nil {
		e.push_err(ident, "obj_have_not_ident", ident.Kind)
	} else {
		d.Constant = item.Value.Data.Constant
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
		Kind:     &TypeKind{f.instance()},
		Model:    &TraitSubIdentExprModel{
			Expr:  d.Model,
			Ident: ident.Kind,
		},
	}
}

func (e *_Eval) eval_struct_sub_ident(d *Data, si *ast.SubIdentExpr, ref bool) *Data {
	s := d.Kind.Strct()

	// TODO: Apply interior mutability.
	f := s.Find_field(si.Ident.Kind)
	if f != nil {
		d.Model = &StrctSubIdentExprModel{
			ExprKind: d.Kind,
			Expr:     d.Model,
			Field:    f,
		}
		d.Kind = f.Kind
		return d
	}

	m := s.Find_method(si.Ident.Kind)
	if m == nil {
		e.push_err(si.Ident, "obj_have_not_ident", si.Ident.Kind)
		return nil
	}

	if m.Decl.Params[0].Is_ref() && !ref {
		e.push_err(si.Ident, "ref_method_used_with_not_ref_instance")
	}

	d.Model = &StrctSubIdentExprModel{
		ExprKind: d.Kind,
		Expr:     d.Model,
		Method:   m,
	}
	d.Kind.kind = m
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
				kind:  &FnIns{},
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

func (e *_Eval) eval_int_type_sub_ident(si *ast.SubIdentExpr) *Data {
	const kind = types.TypeKind_INT
	switch si.Ident.Kind {
	case "max":
		c := constant.New_i64(int64(types.Max_of(kind)))
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	case "min":
		c := constant.New_i64(int64(types.Min_of(kind)))
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(si.Ident, "type_have_not_ident", kind, si.Ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_uint_type_sub_ident(si *ast.SubIdentExpr) *Data {
	const kind = types.TypeKind_UINT
	switch si.Ident.Kind {
	case "max":
		c := constant.New_u64(uint64(types.Max_of(kind)))
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(si.Ident, "type_have_not_ident", kind, si.Ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_i8_type_sub_ident(si *ast.SubIdentExpr) *Data {
	const kind = types.TypeKind_I8
	const min = types.MIN_I8
	const max = types.MAX_I8
	switch si.Ident.Kind {
	case "max":
		c := constant.New_i64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	case "min":
		c := constant.New_i64(min)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(si.Ident, "type_have_not_ident", kind, si.Ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_i16_type_sub_ident(si *ast.SubIdentExpr) *Data {
	const kind = types.TypeKind_I16
	const min = types.MIN_I16
	const max = types.MAX_I16
	switch si.Ident.Kind {
	case "max":
		c := constant.New_i64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	case "min":
		c := constant.New_i64(min)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(si.Ident, "type_have_not_ident", kind, si.Ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_i32_type_sub_ident(si *ast.SubIdentExpr) *Data {
	const kind = types.TypeKind_I32
	const min = types.MIN_I32
	const max = types.MAX_I32
	switch si.Ident.Kind {
	case "max":
		c := constant.New_i64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	case "min":
		c := constant.New_i64(min)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(si.Ident, "type_have_not_ident", kind, si.Ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_i64_type_sub_ident(si *ast.SubIdentExpr) *Data {
	const kind = types.TypeKind_I64
	const min = types.MIN_I64
	const max = types.MAX_I64
	switch si.Ident.Kind {
	case "max":
		c := constant.New_i64(min)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	case "min":
		c := constant.New_i64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(si.Ident, "type_have_not_ident", kind, si.Ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_u8_type_sub_ident(si *ast.SubIdentExpr) *Data {
	const kind = types.TypeKind_U8
	const max = types.MAX_U8
	switch si.Ident.Kind {
	case "max":
		c := constant.New_u64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(si.Ident, "type_have_not_ident", kind, si.Ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_u16_type_sub_ident(si *ast.SubIdentExpr) *Data {
	const kind = types.TypeKind_U16
	const max = types.MAX_U16
	switch si.Ident.Kind {
	case "max":
		c := constant.New_u64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(si.Ident, "type_have_not_ident", kind, si.Ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_u32_type_sub_ident(si *ast.SubIdentExpr) *Data {
	const kind = types.TypeKind_U32
	const max = types.MAX_U32
	switch si.Ident.Kind {
	case "max":
		c := constant.New_u64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(si.Ident, "type_have_not_ident", kind, si.Ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_u64_type_sub_ident(si *ast.SubIdentExpr) *Data {
	const kind = types.TypeKind_U64
	const max = types.MAX_U64
	switch si.Ident.Kind {
	case "max":
		c := constant.New_u64(max)
		return &Data{
			Constant: c,
			Model:    c,
			Kind:     &TypeKind{kind: build_prim_type(kind)},
		}

	default:
		e.push_err(si.Ident, "type_have_not_ident", kind, si.Ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_prim_type_sub_ident(p *Prim, si *ast.SubIdentExpr) *Data {
	kind := p.To_str()
	switch kind {
	case types.TypeKind_INT:
		return e.eval_int_type_sub_ident(si)

	case types.TypeKind_UINT:
		return e.eval_uint_type_sub_ident(si)

	case types.TypeKind_I8:
		return e.eval_i8_type_sub_ident(si)

	case types.TypeKind_I16:
		return e.eval_i16_type_sub_ident(si)

	case types.TypeKind_I32:
		return e.eval_i32_type_sub_ident(si)

	case types.TypeKind_I64:
		return e.eval_i64_type_sub_ident(si)

	case types.TypeKind_U8:
		return e.eval_u8_type_sub_ident(si)

	case types.TypeKind_U16:
		return e.eval_u16_type_sub_ident(si)

	case types.TypeKind_U32:
		return e.eval_u32_type_sub_ident(si)

	case types.TypeKind_U64:
		return e.eval_u64_type_sub_ident(si)

	default:
		e.push_err(si.Ident, "type_have_not_ident", kind, si.Ident.Kind)
		return nil
	}
}

func (e *_Eval) eval_type_sub_ident(d *Data, si *ast.SubIdentExpr) *Data {
	switch {
	case d.Kind.Prim() != nil:
		return e.eval_prim_type_sub_ident(d.Kind.Prim(), si)

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
	case kind.Enm() != nil:
		return e.eval_enum_sub_ident(kind.Enm(), si.Ident)
	
	case kind.Trt() != nil:
		return e.eval_trait_sub_ident(d, kind.Trt(), si.Ident)
	
	case kind.Strct() != nil:
		s := kind.Strct()
		if e.is_instanced_struct(s) {
			used_reference_elem := kind != d.Kind
			return e.eval_struct_sub_ident(d, si, used_reference_elem)
		}

	case kind.Slc() != nil:
		return e.eval_slice_sub_ident(d, si.Ident)

	case kind.Arr() != nil:
		return e.eval_array_sub_ident(d, si.Ident)

	case kind.Map() != nil:
		return e.eval_map_sub_ident(d, si.Ident)
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
	ok := true
	for i, expr := range tup.Expr {
		d := e.eval_expr_kind(expr)
		if d == nil {
			ok = false
			continue
		}
		tup_t.Types[i] = d.Kind
	}

	if !ok {
		return nil
	}

	return &Data{
		Kind: &TypeKind{kind: tup_t},
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
	// TODO: Add support for structure prefix.
	switch {
	case e.prefix == nil:
		e.push_err(lit.Token, "invalid_syntax")
		return nil

	case e.prefix.Map() != nil:
		return e.eval_map(e.prefix.Map(), lit)
	
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

	// TODO: Check scope.

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
	switch kind.(type) {
	case *ast.LitExpr:
		return e.eval_lit(kind.(*ast.LitExpr))
		
	case *ast.IdentExpr:
		return e.eval_ident(kind.(*ast.IdentExpr))

	case *ast.UnaryExpr:
		return e.eval_unary(kind.(*ast.UnaryExpr))

	case *ast.VariadicExpr:
		return e.eval_variadic(kind.(*ast.VariadicExpr))

	case *ast.UnsafeExpr:
		return e.eval_unsafe(kind.(*ast.UnsafeExpr))

	case *ast.SliceExpr:
		return e.eval_slice_expr(kind.(*ast.SliceExpr))

	case *ast.IndexingExpr:
		return e.eval_indexing(kind.(*ast.IndexingExpr))

	case *ast.SlicingExpr:
		return e.eval_slicing(kind.(*ast.SlicingExpr))

	case *ast.CastExpr:
		return e.eval_cast(kind.(*ast.CastExpr))

	case *ast.NsSelectionExpr:
		return e.eval_ns_selection(kind.(*ast.NsSelectionExpr))

	case *ast.StructLit:
		return e.eval_struct_lit(kind.(*ast.StructLit))
	
	case *ast.Type:
		return e.eval_type(kind.(*ast.Type))

	case *ast.FnCallExpr:
		return e.eval_fn_call(kind.(*ast.FnCallExpr))

	case *ast.SubIdentExpr:
		return e.eval_sub_ident(kind.(*ast.SubIdentExpr))

	case *ast.TupleExpr:
		return e.eval_tuple(kind.(*ast.TupleExpr))

	case *ast.BraceLit:
		return e.eval_brace_lit(kind.(*ast.BraceLit))

	case *ast.FnDecl:
		return e.eval_anon_fn(kind.(*ast.FnDecl))

	case *ast.BinopExpr:
		return e.eval_binop(kind.(*ast.BinopExpr))

	default:
		return nil
	}
}

// Returns value data of evaluated expression.
// Returns nil if error occurs.
func (e *_Eval) eval(expr *ast.Expr) *Data {
	d := e.eval_expr_kind(expr.Kind)
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
	} else if !d.Is_const() {
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

func (bs *_BinopSolver) eval_float() *Data {
	rk := bs.r.Kind.To_str()
	if !types.Is_num(rk) {
		bs.e.push_err(bs.op, "incompatible_types", bs.l.Kind.To_str(), rk)
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
		if types.Is_greater(rk, bs.l.Kind.To_str()) {
			bs.l.Kind = bs.r.Kind
		}
		return bs.l

	case lex.KND_PERCENT:
		if !types.Is_int(rk) {
			bs.e.push_err(bs.op, "incompatible_types", bs.l.Kind.To_str(), rk)
			return nil
		}
		return bs.r

	default:
		bs.e.push_err(bs.op, "operator_not_for_float", bs.op.Kind)
		return nil
	}
}

func (bs *_BinopSolver) eval_unsig_int() *Data {
	rk := bs.r.Kind.To_str()
	if !types.Is_num(rk) {
		bs.e.push_err(bs.op, "incompatible_types", bs.l.Kind.To_str(), rk)
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
		lex.KND_PERCENT,
		lex.KND_AMPER,
		lex.KND_VLINE,
		lex.KND_CARET:
		if types.Is_greater(rk, bs.l.Kind.To_str()) {
			bs.l.Kind = bs.r.Kind
		}
		return bs.l

	case lex.KND_LSHIFT, lex.KND_RSHIFT:
		if !is_ok_for_shifting(bs.r) {
			bs.e.push_err(bs.op, "bitshift_must_unsigned")
			return nil
		}
		bs.l.Kind.kind = build_prim_type(types.TypeKind_U64)
		return bs.l

	default:
		bs.e.push_err(bs.op, "operator_not_for_uint", bs.op.Kind)
		return nil
	}
}

func (bs *_BinopSolver) eval_sig_int() *Data {
	rk := bs.r.Kind.To_str()
	if !types.Is_num(rk) {
		bs.e.push_err(bs.op, "incompatible_types", bs.l.Kind.To_str(), rk)
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
		lex.KND_PERCENT,
		lex.KND_AMPER,
		lex.KND_VLINE,
		lex.KND_CARET:
		if types.Is_greater(rk, bs.l.Kind.To_str()) {
			bs.l.Kind = bs.r.Kind
		}
		return bs.l

	case lex.KND_LSHIFT, lex.KND_RSHIFT:
		if !is_ok_for_shifting(bs.r) {
			bs.e.push_err(bs.op, "bitshift_must_unsigned")
			return nil
		}
		bs.l.Kind.kind = build_prim_type(types.TypeKind_U64)
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
	rk := rprim.To_str()
	switch {
	case types.Is_float(lk) || types.Is_float(rk):
		if types.Is_float(rk) {
			bs.l, bs.r = bs.r, bs.l
		}
		return bs.eval_float()

	case types.Is_unsig_int(lk) || types.Is_unsig_int(rk):
		if types.Is_unsig_int(rk) {
			bs.l, bs.r = bs.r, bs.l
		}
		return bs.eval_unsig_int()

	case types.Is_sig_int(lk) || types.Is_sig_int(rk):
		if types.Is_sig_int(rk) {
			bs.l, bs.r = bs.r, bs.l
		}
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

func (bs *_BinopSolver) solve_const(d *Data) {
	switch {
	case d == nil:
		return
		
	case !bs.l.Is_const() || !bs.r.Is_const():
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

	case lex.KND_RSHIFT:
		_ = bs.l.Constant.Rshift(*bs.r.Constant)
		d.Constant = bs.l.Constant
	}
}

func (bs *_BinopSolver) normalize_bitsize(d *Data) {
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

func (bs *_BinopSolver) post_const(d *Data) {
	if d == nil || !d.Is_const() {
		return
	}

	bs.normalize_bitsize(d)
}

func (bs *_BinopSolver) solve_explicit(l *Data, r *Data) *Data {
	bs.l, bs.r = l, r
	data := bs.eval()
	bs.l, bs.r = l, r // Save normal order
	
	bs.solve_const(data)
	bs.post_const(data)
	
	if data != nil {
		data.Model = &BinopExprModel{
			L: l.Model,
			R: r.Model,
			Op: bs.op.Kind,
		}
	}
	
	return data
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
	return bs.solve_explicit(bs.l, bs.r)
}
