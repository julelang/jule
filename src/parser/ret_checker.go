package parser

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

type ret_checker struct {
	p         *Parser
	ret_ast   *ast.Ret
	f         *Fn
	exp_model retExpr
	values    []value
}

func (rc *ret_checker) pushval(last, current int, errTok lex.Token) {
	if current-last == 0 {
		rc.p.pusherrtok(errTok, "missing_expr")
		return
	}
	toks := rc.ret_ast.Expr.Tokens[last:current]
	var prefix Type
	i := len(rc.values)
	if rc.f.RetType.DataType.MultiTyped {
		types := rc.f.RetType.DataType.Tag.([]Type)
		if i < len(types) {
			prefix = types[i]
		}
	} else if i == 0 {
		prefix = rc.f.RetType.DataType
	}
	v, model := rc.p.evalToks(toks, &prefix)
	rc.exp_model.models = append(rc.exp_model.models, model)
	rc.values = append(rc.values, v)
}

func (rc *ret_checker) check_expressions() {
	brace_n := 0
	last := 0
	for i, tok := range rc.ret_ast.Expr.Tokens {
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				brace_n++
			default:
				brace_n--
			}
		}
		if brace_n > 0 || tok.Id != lex.ID_COMMA {
			continue
		}
		rc.pushval(last, i, tok)
		last = i + 1
	}
	n := len(rc.ret_ast.Expr.Tokens)
	if last < n {
		if last == 0 {
			rc.pushval(0, n, rc.ret_ast.Token)
		} else {
			rc.pushval(last, n, rc.ret_ast.Expr.Tokens[last-1])
		}
	}
	if !types.IsVoid(rc.f.RetType.DataType) {
		rc.check_type_safety()
		rc.ret_ast.Expr.Model = rc.exp_model
	}
}

func (rc *ret_checker) check_for_ret_expr(v value) {
	if rc.p.unsafe_allowed() || !lex.IsIdentifierRune(v.data.Value) {
		return
	}
	if !v.mutable && types.IsMut(v.data.DataType) {
		rc.p.pusherrtok(rc.ret_ast.Token, "ret_with_mut_typed_non_mut")
		return
	}
}

func (rc *ret_checker) single() {
	if len(rc.values) > 1 {
		rc.p.pusherrtok(rc.ret_ast.Token, "overflow_return")
	}
	v := rc.values[0]
	rc.check_for_ret_expr(v)
	assign_checker{
		p:      rc.p,
		t: rc.f.RetType.DataType,
		v:      v,
		errtok: rc.ret_ast.Token,
	}.check()
}

func (rc *ret_checker) multi() {
	types := rc.f.RetType.DataType.Tag.([]Type)
	n := len(rc.values)
	if n == 1 {
		rc.checkMultiRetAsMutliRet()
		return
	} else if n > len(types) {
		rc.p.pusherrtok(rc.ret_ast.Token, "overflow_return")
	}
	for i, t := range types {
		if i >= n {
			break
		}
		v := rc.values[i]
		rc.check_for_ret_expr(v)
		assign_checker{
			p:      rc.p,
			t: t,
			v:      v,
			errtok: rc.ret_ast.Token,
		}.check()
	}
}

func (rc *ret_checker) check_type_safety() {
	if !rc.f.RetType.DataType.MultiTyped { // Single return
		rc.single()
		return
	}
	// Multi return
	rc.multi()
}

func (rc *ret_checker) checkMultiRetAsMutliRet() {
	v := rc.values[0]
	if !v.data.DataType.MultiTyped {
		rc.p.pusherrtok(rc.ret_ast.Token, "missing_multi_return")
		return
	}
	val_types := v.data.DataType.Tag.([]Type)
	ret_types := rc.f.RetType.DataType.Tag.([]Type)
	if len(val_types) < len(ret_types) {
		rc.p.pusherrtok(rc.ret_ast.Token, "missing_multi_return")
		return
	} else if len(val_types) < len(ret_types) {
		rc.p.pusherrtok(rc.ret_ast.Token, "overflow_return")
		return
	}
	for i, rt := range ret_types {
		vt := val_types[i]
		v := value{data: ast.Data{DataType: vt}}
		v.data.Value = " " // Ignore eval error.
		assign_checker{
			p:                rc.p,
			t:                rt,
			v:                v,
			ignoreAny:        false,
			not_allow_assign: false,
			errtok:           rc.ret_ast.Token,
		}.check()
	}
}

func (rc *ret_checker) retsVars() {
	if !rc.f.RetType.DataType.MultiTyped {
		for _, v := range rc.f.RetType.Identifiers {
			if !lex.IsIgnoreId(v.Kind) {
				model := new(exprModel)
				model.index = 0
				model.nodes = make([]exprBuildNode, 1)
				val, _ := rc.p.eval.single(v, model)
				rc.exp_model.models = append(rc.exp_model.models, model)
				rc.values = append(rc.values, val)
				break
			}
		}
		rc.ret_ast.Expr.Model = rc.exp_model
		return
	}
	types := rc.f.RetType.DataType.Tag.([]Type)
	for i, v := range rc.f.RetType.Identifiers {
		if lex.IsIgnoreId(v.Kind) {
			node := exprNode{}
			node.value = types[i].String()
			node.value += types[i].InitValue()
			rc.exp_model.models = append(rc.exp_model.models, node)
			continue
		}
		model := new(exprModel)
		model.index = 0
		model.nodes = make([]exprBuildNode, 1)
		val, _ := rc.p.eval.single(v, model)
		rc.exp_model.models = append(rc.exp_model.models, model)
		rc.values = append(rc.values, val)
	}
	rc.ret_ast.Expr.Model = rc.exp_model
}

func (rc *ret_checker) check() {
	n := len(rc.ret_ast.Expr.Tokens)
	if n == 0 && !types.IsVoid(rc.f.RetType.DataType) {
		if !rc.f.RetType.AnyVar() {
			rc.p.pusherrtok(rc.ret_ast.Token, "require_return_value")
		}
		rc.retsVars()
		return
	}
	if n > 0 && types.IsVoid(rc.f.RetType.DataType) {
		rc.p.pusherrtok(rc.ret_ast.Token, "void_function_return_value")
	}
	rc.exp_model.vars = rc.f.RetType.Vars(rc.p.nodeBlock)
	rc.check_expressions()
}
