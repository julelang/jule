package sema

import "github.com/julelang/jule/ast"

// Type alias for built-in function callers.
//
// Parameters;
//  e: Caller owner Eval instance.
//  fc: Function call expression.
//  d: Data instance for evaluated expression of function.
type _BuiltinCaller = func(e *_Eval, fc *ast.FnCallExpr, d *Data) *Data

func common_builtin_caller(e *_Eval, fc *ast.FnCallExpr, d *Data) *Data {
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
