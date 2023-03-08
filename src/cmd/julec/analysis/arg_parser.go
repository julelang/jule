package analysis

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

func get_param_map(params []Param) *paramMap {
	pmap := new(paramMap)
	*pmap = make(paramMap, len(params))
	for i := range params {
		param := &params[i]
		(*pmap)[param.Id] = &paramMapPair{param, nil}
	}
	return pmap
}

type pureArgParser struct {
	p       *Analyzer
	pmap    *paramMap
	f       *Fn
	args    *ast.Args
	i       int
	arg     Arg
	errTok  lex.Token
	m       *expr_model
	paramId string
}

func (pap *pureArgParser) build_args() {
	if pap.pmap == nil {
		return
	}
	pap.args.Src = make([]Arg, len(*pap.pmap))
	for i, p := range pap.f.Params {
		pair := (*pap.pmap)[p.Id]
		switch {
		case pair.arg != nil:
			pap.args.Src[i] = *pair.arg
		case pair.param.Variadic:
			arg := Arg{Expr: Expr{Model: exprNode{build.CPP_DEFAULT_EXPR}}}
			pap.args.Src[i] = arg
		}
	}
}

func (pap *pureArgParser) push_variadic_args(pair *paramMapPair) {
	// Used to build initializer list for slice
	var model serieExpr
	model.exprs = append(model.exprs, exprNode{lex.KND_LBRACE})
	variadiced := false
	pap.p.parse_arg(pap.f, pair, pap.args, &variadiced)
	model.exprs = append(model.exprs, exprNode{pair.arg.String()})
	once := false
	for pap.i++; pap.i < len(pap.args.Src); pap.i++ {
		pair.arg = &pap.args.Src[pap.i]
		once = true
		pap.p.parse_arg(pap.f, pair, pap.args, &variadiced)
		model.exprs = append(model.exprs, exprNode{lex.KND_COMMA})
		model.exprs = append(model.exprs, exprNode{pair.arg.String()})
	}
	model.exprs = append(model.exprs, exprNode{lex.KND_RBRACE})
	if !variadiced {
		pair.arg.Expr.Model = model
	}
	if !once {
		return
	}
	// Variadic argument must have only one expression for variadication
	if variadiced {
		pap.p.pusherrtok(pap.errTok, "more_args_with_variadiced")
	}
}

func (pap *pureArgParser) check_param_arg(pair *paramMapPair) {
	if pair.arg == nil && !pair.param.Variadic {
		pap.p.pusherrtok(pap.errTok, "missing_expr_for", pair.param.Id)
	}
}

func (pap *pureArgParser) check_passes_struct() {
	if len(pap.args.Src) == 0 {
		for _, pair := range *pap.pmap {
			if types.IsRef(pair.param.DataType) {
				pap.p.pusherrtok(pap.errTok, "reference_field_not_initialized", pair.param.Id)
			}
		}
		pap.pmap = nil
		return
	}
	for _, pair := range *pap.pmap {
		pap.check_param_arg(pair)
	}
}

func (pap *pureArgParser) check_passes_fn() {
	for _, pair := range *pap.pmap {
		pap.check_param_arg(pair)
	}
}

func (pap *pureArgParser) check_passes() {
	if pap.f.IsConstructor() {
		pap.check_passes_struct()
		return
	}
	pap.check_passes_fn()
}

func (pap *pureArgParser) push_arg() {
	pair := (*pap.pmap)[pap.paramId]
	arg := pap.arg
	pair.arg = &arg
	if pair.param.Variadic {
		pap.push_variadic_args(pair)
	} else {
		pap.p.parse_arg(pap.f, pair, pap.args, nil)
	}
	pap.i++
}

func is_multi_ret_as_args(f *Fn, nargs int) bool {
	return nargs < len(f.Params) && nargs == 1
}

func (pap *pureArgParser) parse() {
	if is_multi_ret_as_args(pap.f, len(pap.args.Src)) {
		if pap.try_fn_multi_ret_as_args() {
			return
		}
	}
	pap.pmap = get_param_map(pap.f.Params)
	argCount := 0
	for pap.i < len(pap.args.Src) {
		if argCount >= len(pap.f.Params) {
			pap.p.pusherrtok(pap.errTok, "argument_overflow")
			return
		}
		argCount++
		pap.arg = pap.args.Src[pap.i]
		pap.paramId = pap.f.Params[pap.i].Id
		pap.push_arg()
	}
	pap.check_passes()
	pap.build_args()
}

func (pap *pureArgParser) try_fn_multi_ret_as_args() bool {
	arg := pap.args.Src[0]
	val, model := pap.p.eval_expr(arg.Expr, nil)
	arg.Expr.Model = model
	if !val.data.DataType.MultiTyped {
		return false
	}
	types := val.data.DataType.Tag.([]Type)
	if len(types) < len(pap.f.Params) {
		return false
	} else if len(types) > len(pap.f.Params) {
		return false
	}
	pair := &paramMapPair{
		param: nil,
		arg:   &arg,
	}
	for i, param := range pap.f.Params {
		pair.param = &param
		rt := types[i]
		val := value{data: ast.Data{DataType: rt}}
		pap.p.check_arg(pap.f, pair, pap.args, nil, val)
		pap.p.checkArgType(&param, val, arg.Token)
	}
	if pap.m != nil {
		ready_to_parse_generic_fn(pap.f)
		model := exprNode{"__julec_tuple_as_args<"}
		model.value += pap.f.CppKind(true)
		model.value += ">"
		fname := pap.m.nodes[pap.m.index].nodes[0]
		pap.m.nodes[pap.m.index].nodes[0] = model
		arg.Expr.Model = exprNode{fname.String() + "," + arg.Expr.String()}
		pap.args.Src[0] = arg
	}
	return true
}
