package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/types"
)

// Type alias for built-in function callers.
//
// Parameters;
//  e: Caller owner Eval instance.
//  fc: Function call expression.
//  d: Data instance for evaluated expression of function.
type _BuiltinCaller = func(e *_Eval, fc *ast.FnCallExpr, d *Data) *Data

var builtin_fn_out = &FnIns{}
var builtin_fn_outln = &FnIns{}
var builtin_fn_new = &FnIns{}
var builtin_fn_drop = &FnIns{}
var builtin_fn_panic = &FnIns{}
var builtin_fn_make = &FnIns{}
var builtin_fn_append = &FnIns{}
var builtin_fn_recover = &FnIns{}
var builtin_fn_std_mem_size_of = &FnIns{}

var builtin_fn_real = &FnIns{
	Result: &TypeKind{kind: build_prim_type(types.TypeKind_BOOL)},
}

var builtin_fn_copy = &FnIns{
	Result: &TypeKind{kind: build_prim_type(types.TypeKind_INT)},
}

var builtin_type_alias_byte = &TypeAlias{
	Public: true,
	Kind:   &TypeSymbol{
		Kind: &TypeKind{kind: build_prim_type(types.TypeKind_U8)},
	},
}

var builtin_type_alias_rune = &TypeAlias{
	Public: true,
	Kind:   &TypeSymbol{
		Kind: &TypeKind{kind: build_prim_type(types.TypeKind_I32)},
	},
}

var builtin_trait_error = &Trait{
	Public:  true,
	Ident:   "Error",
	Methods: []*Fn{
		{
			Public: true,
			Ident:  "error",
			Params: []*Param{
				{
					Ident: "self",
				},
			},
			Result: &RetType{
				Kind: &TypeSymbol{
					Decl: &ast.Type{Kind: &ast.IdentType{Ident: "str"}},
				},
			},
		},
	},
}

func init() {
	builtin_fn_out.Caller = builtin_caller_out
	builtin_fn_outln.Caller = builtin_caller_outln
	builtin_fn_new.Caller = builtin_caller_new
	builtin_fn_real.Caller = builtin_caller_real
	builtin_fn_drop.Caller = builtin_caller_drop
	builtin_fn_panic.Caller = builtin_caller_panic
	builtin_fn_make.Caller = builtin_caller_make
	builtin_fn_append.Caller = builtin_caller_append
	builtin_fn_copy.Caller = builtin_caller_copy
	builtin_fn_recover.Caller = builtin_caller_recover

	builtin_fn_std_mem_size_of.Caller = builtin_caller_std_mem_size_of
}

func find_builtin_fn(ident string) *FnIns {
	switch ident {
	case "out":
		return builtin_fn_out

	case "outln":
		return builtin_fn_outln

	case "new":
		return builtin_fn_new

	case "real":
		return builtin_fn_real

	case "drop":
		return builtin_fn_drop

	case "panic":
		return builtin_fn_panic

	case "make":
		return builtin_fn_make

	case "append":
		return builtin_fn_append

	case "copy":
		return builtin_fn_copy

	case "recover":
		return builtin_fn_recover
	
	default:
		return nil
	}
}

func find_builtin_type_alias(ident string) *TypeAlias {
	switch ident {
	case "byte":
		return builtin_type_alias_byte

	case "rune":
		return builtin_type_alias_rune

	default:
		return nil
	}
}

func find_builtin_trait(ident string) *Trait {
	switch ident {
	case "Error":
		return builtin_trait_error

	default:
		return nil
	}
}

func find_builtin_def(ident string) any {
	f := find_builtin_fn(ident)
	if f != nil {
		return f
	}

	ta := find_builtin_type_alias(ident)
	if ta != nil {
		return ta
	}

	t := find_builtin_trait(ident)
	if t != nil {
		return t
	}

	return nil
}

func find_builtin_def_std_mem(ident string) any {
	switch ident {
	case "size_of":
		return builtin_fn_std_mem_size_of

	default:
		return nil
	}
}

func find_package_builtin_def(link_path string, ident string) any {
	switch link_path {
	case "std::mem":
		return find_builtin_def_std_mem(ident)

	default:
		return nil
	}
}

func builtin_caller_common(e *_Eval, fc *ast.FnCallExpr, d *Data) *Data {
	f := d.Kind.Fnc()

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

func builtin_caller_out(e *_Eval, fc *ast.FnCallExpr, _ *Data) *Data {
	if len(fc.Args) < 1 {
		e.push_err(fc.Token, "missing_expr_for", "v")
		return nil
	}
	if len(fc.Args) > 1 {
		e.push_err(fc.Args[2].Token, "argument_overflow")
	}

	expr := e.eval_expr(fc.Args[0])
	if expr == nil {
		return nil
	}

	if expr.Kind.Fnc() != nil {
		e.push_err(fc.Args[0].Token, "invalid_expr")
		return nil
	}

	d := build_void_data()
	d.Model = &BuiltinOutCallExprModel{Expr: expr.Model}
	return d
}

func builtin_caller_outln(e *_Eval, fc *ast.FnCallExpr, _ *Data) *Data {
	d := builtin_caller_out(e, fc, nil)
	if d == nil {
		return nil
	}

	d.Model = &BuiltinOutlnCallExprModel{
		Expr: d.Model.(*BuiltinOutCallExprModel).Expr,
	}
	return d
}

func builtin_caller_new(e *_Eval, fc *ast.FnCallExpr, d *Data) *Data {
	if len(fc.Args) < 1 {
		e.push_err(fc.Token, "missing_expr_for", "type")
		return nil
	}
	if len(fc.Args) > 2 {
		e.push_err(fc.Args[2].Token, "argument_overflow")
	}

	t := e.eval_expr_kind(fc.Args[0].Kind)
	if t == nil {
		return nil
	}

	if !t.Decl {
		e.push_err(fc.Args[0].Token, "invalid_type")
		return nil
	}

	if !is_valid_for_ref(t.Kind) {
		e.push_err(fc.Args[0].Token, "invalid_type")
		return nil
	}

	d.Kind = &TypeKind{kind: &Ref{Elem: t.Kind}}


	if len(fc.Args) == 2 { // Initialize expression.
		init := e.s.evalp(fc.Args[1], e.lookup, &TypeSymbol{Kind: t.Kind})
		if init != nil {
			e.s.check_assign_type(t.Kind, init, fc.Args[1].Token, false)
			d.Model = &BuiltinNewCallExprModel{
				Kind: t.Kind,
				Init: init.Model,
			}
		}
	} else {
		d.Model = &BuiltinNewCallExprModel{Kind: t.Kind}
	}

	return d
}

func builtin_caller_real(e *_Eval, fc *ast.FnCallExpr, d *Data) *Data {
	if len(fc.Args) < 1 {
		e.push_err(fc.Token, "missing_expr_for", "ref")
		return nil
	}
	if len(fc.Args) > 1 {
		e.push_err(fc.Args[2].Token, "argument_overflow")
	}

	ref := e.eval_expr(fc.Args[0])
	if ref == nil {
		return nil
	}

	if ref.Kind.Ref() == nil {
		e.push_err(fc.Args[0].Token, "invalid_expr")
		return nil
	}

	d.Kind = builtin_fn_real.Result
	d.Model = &BuiltinRealCallExprModel{Expr: ref.Model}
	return d
}

func builtin_caller_drop(e *_Eval, fc *ast.FnCallExpr, _ *Data) *Data {
	if len(fc.Args) < 1 {
		e.push_err(fc.Token, "missing_expr_for", "ref")
		return nil
	}
	if len(fc.Args) > 1 {
		e.push_err(fc.Args[2].Token, "argument_overflow")
	}

	ref := e.eval_expr(fc.Args[0])
	if ref == nil {
		return nil
	}

	if ref.Kind.Ref() == nil {
		e.push_err(fc.Args[0].Token, "invalid_expr")
		return nil
	}

	d := build_void_data()
	d.Model = &BuiltinDropCallExprModel{Expr: ref.Model}
	return d
}

func builtin_caller_panic(e *_Eval, fc *ast.FnCallExpr, _ *Data) *Data {
	if len(fc.Args) < 1 {
		e.push_err(fc.Token, "missing_expr_for", "error")
		return nil
	}
	if len(fc.Args) > 1 {
		e.push_err(fc.Args[2].Token, "argument_overflow")
	}

	expr := e.eval_expr(fc.Args[0])
	if expr == nil {
		return nil
	}

	d := build_void_data()
	d.Model = &BuiltinPanicCallExprModel{Expr: expr.Model}
	return d
}

func builtin_caller_make(e *_Eval, fc *ast.FnCallExpr, d *Data) *Data {
	if len(fc.Args) < 2 {
		if len(fc.Args) == 1 {
			e.push_err(fc.Token, "missing_expr_for", "size")
			return nil
		}
		e.push_err(fc.Token, "missing_expr_for", "type, and size")
		return nil
	}
	if len(fc.Args) > 2 {
		e.push_err(fc.Args[2].Token, "argument_overflow")
	}

	t := e.eval_expr_kind(fc.Args[0].Kind)
	if t == nil {
		return nil
	}

	if !t.Decl || t.Kind.Slc() == nil {
		e.push_err(fc.Args[0].Token, "invalid_type")
		return nil
	}

	d.Kind = t.Kind
	
	size := e.s.evalp(fc.Args[1], e.lookup, &TypeSymbol{Kind: t.Kind})
	if size == nil {
		return d
	}
	
	e.check_integer_indexing_by_data(size, fc.Args[1].Token)

	// Ignore size expression if size is constant zero.
	if size.Is_const() && size.Constant.As_i64() == 0 {
		size.Model = nil
	}

	d.Model = &BuiltinMakeCallExprModel{
		Kind: t.Kind,
		Size: size.Model,
	}

	return d
}

func builtin_caller_append(e *_Eval, fc *ast.FnCallExpr, d *Data) *Data {
	if len(fc.Args) < 2 {
		if len(fc.Args) == 1 {
			e.push_err(fc.Token, "missing_expr_for", "src")
			return nil
		}
		e.push_err(fc.Token, "missing_expr_for", "src, and values")
		return nil
	}

	t := e.eval_expr(fc.Args[0])
	if t == nil {
		return nil
	}

	if t.Kind.Slc() == nil {
		e.push_err(fc.Args[0].Token, "invalid_expr")
		return nil
	}

	f := &FnIns{
		Params: []*ParamIns{
			{
				Decl: &Param{},
				Kind: t.Kind,
			},
			{
				Decl: &Param{
					Variadic: true,
				},
				Kind: t.Kind.Slc().Elem,
			},
		},
		Result: t.Kind,
	}
	d.Kind = &TypeKind{kind: f}
	d.Model = &CommonIdentExprModel{Ident: "_append"}

	d = builtin_caller_common(e, fc, d)
	return d
}

func builtin_caller_copy(e *_Eval, fc *ast.FnCallExpr, d *Data) *Data {
	if len(fc.Args) < 2 {
		if len(fc.Args) == 1 {
			e.push_err(fc.Token, "missing_expr_for", "src")
			return nil
		}
		e.push_err(fc.Token, "missing_expr_for", "src, values")
		return nil
	}
	if len(fc.Args) > 2 {
		e.push_err(fc.Args[2].Token, "argument_overflow")
	}

	t := e.eval_expr(fc.Args[0])
	if t == nil {
		return nil
	}

	if t.Kind.Slc() == nil {
		e.push_err(fc.Args[0].Token, "invalid_expr")
		return nil
	}

	if !t.Mutable {
		e.push_err(fc.Args[0].Token, "mutable_operation_on_immutable")
	}

	f := &FnIns{
		Params: []*ParamIns{
			{
				Decl: &Param{},
				Kind: t.Kind,
			},
			{
				Decl: &Param{},
				Kind: t.Kind,
			},
		},
		Result: builtin_fn_copy.Result,
	}

	d.Kind = &TypeKind{kind: f}
	d.Model = &CommonIdentExprModel{Ident: "_copy"}

	d = builtin_caller_common(e, fc, d)
	return d
}

func builtin_caller_recover(e *_Eval, fc *ast.FnCallExpr, _ *Data) *Data {
	const HANDLER_KIND = "fn(Error)"

	if len(fc.Args) < 1 {
		e.push_err(fc.Token, "missing_expr_for", "handler")
		return nil
	}
	if len(fc.Args) > 1 {
		e.push_err(fc.Args[1].Token, "argument_overflow")
	}

	t := e.eval_expr(fc.Args[0])
	if t == nil {
		return nil
	}

	if t.Kind.Fnc() == nil {
		e.push_err(fc.Args[0].Token, "invalid_expr")
		return nil
	}

	tkind := t.Kind.Fnc().To_str()
	if tkind !=  HANDLER_KIND {
		e.push_err(fc.Args[0].Token, "incompatible_types", tkind, HANDLER_KIND)
	}

	d := build_void_data()
	d.Kind = t.Kind
	d.Model = &Recover{
		Handler_expr: t.Model,
	}
	return d
}

func builtin_caller_std_mem_size_of(e *_Eval, fc *ast.FnCallExpr, _ *Data) *Data {
	result := &Data{
		Kind:  &TypeKind{kind: build_prim_type(types.TypeKind_UINT)},
	}

	if len(fc.Args) < 1 {
		e.push_err(fc.Token, "missing_expr_for", "type|expr")
		return result
	}
	if len(fc.Args) > 1 {
		e.push_err(fc.Args[1].Token, "argument_overflow")
	}

	d := e.eval_expr_kind(fc.Args[0].Kind)
	if d == nil {
		return result
	}

	result.Model = &SizeofExprModel{Expr: d.Model}
	return result
}
