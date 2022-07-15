package parser

import (
	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/pkg/x"
)

type targetedArgParser struct {
	p      *Parser
	pmap   *paramMap
	f      *Func
	args   *models.Args
	i      int
	arg    Arg
	errTok Tok
}

func (tap *targetedArgParser) buildArgs() {
	tap.args.Src = make([]models.Arg, 0)
	for _, p := range tap.f.Params {
		pair := (*tap.pmap)[p.Id]
		switch {
		case pair.arg != nil:
			tap.args.Src = append(tap.args.Src, *pair.arg)
		case paramHasDefaultArg(pair.param):
			arg := Arg{Expr: pair.param.Default}
			tap.args.Src = append(tap.args.Src, arg)
		case pair.param.Variadic:
			model := sliceExpr{pair.param.Type, nil}
			model.dataType.Kind = x.Prefix_Slice + model.dataType.Kind // For slice.
			arg := Arg{Expr: Expr{Model: model}}
			tap.args.Src = append(tap.args.Src, arg)
		}
	}
}

func (tap *targetedArgParser) pushVariadicArgs(pair *paramMapPair) {
	model := sliceExpr{pair.param.Type, nil}
	model.dataType.Kind = x.Prefix_Slice + model.dataType.Kind // For slice.
	variadiced := false
	tap.p.parseArg(*pair.param, pair.arg, &variadiced)
	model.expr = append(model.expr, pair.arg.Expr.Model.(iExpr))
	once := false
	for tap.i++; tap.i < len(tap.args.Src); tap.i++ {
		arg := tap.args.Src[tap.i]
		if arg.TargetId != "" {
			tap.i--
			break
		}
		once = true
		tap.p.parseArg(*pair.param, &arg, &variadiced)
		model.expr = append(model.expr, arg.Expr.Model.(iExpr))
	}
	if !once {
		return
	}
	// Variadic argument have one more variadiced expressions.
	if variadiced {
		tap.p.pusherrtok(tap.errTok, "more_args_with_variadiced")
	}
	pair.arg.Expr.Model = model
}

func (tap *targetedArgParser) pushArg() {
	defer func() { tap.i++ }()
	if tap.arg.TargetId == "" {
		tap.p.pusherrtok(tap.arg.Tok, "argument_must_target_to_parameter")
		return
	}
	pair, ok := (*tap.pmap)[tap.arg.TargetId]
	if !ok {
		tap.p.pusherrtok(tap.arg.Tok, "function_not_has_parameter", tap.arg.TargetId)
		return
	} else if pair.arg != nil {
		tap.p.pusherrtok(tap.arg.Tok, "parameter_already_has_argument", tap.arg.TargetId)
		return
	}
	arg := tap.arg
	pair.arg = &arg
	if pair.param.Variadic {
		tap.pushVariadicArgs(pair)
	} else {
		tap.p.parseArg(*pair.param, pair.arg, nil)
	}
}

func (tap *targetedArgParser) checkPasses() {
	for _, pair := range *tap.pmap {
		if pair.arg == nil &&
			!pair.param.Variadic &&
			!paramHasDefaultArg(pair.param) {
			tap.p.pusherrtok(tap.errTok, "missing_argument_for", pair.param.Id)
		}
	}
}

func (tap *targetedArgParser) parse() {
	tap.pmap = getParamMap(tap.f.Params)
	// Check non targeteds
	argCount := 0
	for tap.i, tap.arg = range tap.args.Src {
		if tap.arg.TargetId != "" { // Targeted?
			break
		}
		if argCount >= len(tap.f.Params) {
			tap.p.pusherrtok(tap.errTok, "argument_overflow")
			return
		}
		argCount++
		param := tap.f.Params[tap.i]
		arg := tap.arg
		(*tap.pmap)[param.Id].arg = &arg
		tap.p.parseArg(param, &arg, nil)
	}
	for tap.i < len(tap.args.Src) {
		tap.arg = tap.args.Src[tap.i]
		tap.pushArg()
	}
	tap.checkPasses()
	tap.buildArgs()
}
