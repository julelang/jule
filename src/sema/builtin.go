package sema

import "github.com/julelang/jule/ast"

// Type alias for built-in function callers.
//
// Parameters;
//  e: Caller owner Eval instance.
//  fc: Function call expression.
//  d: Data instance for evaluated expression of function.
type _BuiltinCaller = func(e *_Eval, fc *ast.FnCallExpr, d *Data) *Data

var builtin_fn_new = &FnIns{}

func init() {
	builtin_fn_new.Caller = builtin_caller_new
}

func get_builtin_def(ident string) any {
	switch ident {
	case "new":
		return builtin_fn_new

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
		e.push_err(fc.Args[0].Token, "invalid_expr")
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
		d.Model = &BuiltinNewCallExprModel{
			Kind: t.Kind,
		}
	}

	return d
}
