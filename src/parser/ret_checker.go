package parser

import (
	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

type retChecker struct {
	t         *Parser
	ret_ast   *models.Ret
	f         *Fn
	exp_model retExpr
	values    []value
}

func (rc *retChecker) pushval(last, current int, errTok lex.Token) {
	if current-last == 0 {
		rc.t.pusherrtok(errTok, "missing_expr")
		return
	}
	toks := rc.ret_ast.Expr.Tokens[last:current]
	var prefix Type
	i := len(rc.values)
	if rc.f.RetType.Type.MultiTyped {
		types := rc.f.RetType.Type.Tag.([]Type)
		if i < len(types) {
			prefix = types[i]
		}
	} else if i == 0 {
		prefix = rc.f.RetType.Type
	}
	v, model := rc.t.evalToks(toks, &prefix)
	rc.exp_model.models = append(rc.exp_model.models, model)
	rc.values = append(rc.values, v)
}

func (rc *retChecker) checkepxrs() {
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
	if !types.IsVoid(rc.f.RetType.Type) {
		rc.checkExprTypes()
		rc.ret_ast.Expr.Model = rc.exp_model
	}
}

func (rc *retChecker) check_for_ret_expr(v value) {
	if rc.t.unsafe_allowed() || !lex.IsIdentifierRune(v.data.Value) {
		return
	}
	if !v.mutable && types.IsMut(v.data.Type) {
		rc.t.pusherrtok(rc.ret_ast.Token, "ret_with_mut_typed_non_mut")
		return
	}
}

func (rc *retChecker) single() {
	if len(rc.values) > 1 {
		rc.t.pusherrtok(rc.ret_ast.Token, "overflow_return")
	}
	v := rc.values[0]
	rc.check_for_ret_expr(v)
	assign_checker{
		p:      rc.t,
		expr_t: rc.f.RetType.Type,
		v:      v,
		errtok: rc.ret_ast.Token,
	}.check()
}

func (rc *retChecker) multi() {
	types := rc.f.RetType.Type.Tag.([]Type)
	n := len(rc.values)
	if n == 1 {
		rc.checkMultiRetAsMutliRet()
		return
	} else if n > len(types) {
		rc.t.pusherrtok(rc.ret_ast.Token, "overflow_return")
	}
	for i, t := range types {
		if i >= n {
			break
		}
		v := rc.values[i]
		rc.check_for_ret_expr(v)
		assign_checker{
			p:      rc.t,
			expr_t: t,
			v:      v,
			errtok: rc.ret_ast.Token,
		}.check()
	}
}

func (rc *retChecker) checkExprTypes() {
	if !rc.f.RetType.Type.MultiTyped { // Single return
		rc.single()
		return
	}
	// Multi return
	rc.multi()
}

func (rc *retChecker) checkMultiRetAsMutliRet() {
	v := rc.values[0]
	if !v.data.Type.MultiTyped {
		rc.t.pusherrtok(rc.ret_ast.Token, "missing_multi_return")
		return
	}
	valTypes := v.data.Type.Tag.([]Type)
	retTypes := rc.f.RetType.Type.Tag.([]Type)
	if len(valTypes) < len(retTypes) {
		rc.t.pusherrtok(rc.ret_ast.Token, "missing_multi_return")
		return
	} else if len(valTypes) < len(retTypes) {
		rc.t.pusherrtok(rc.ret_ast.Token, "overflow_return")
		return
	}
	for i, rt := range retTypes {
		vt := valTypes[i]
		val := value{data: models.Data{Type: vt}}
		assign_checker{
			p:      rc.t,
			expr_t: rt,
			v:      val,
			errtok: rc.ret_ast.Token,
		}.check()
	}
}

func (rc *retChecker) retsVars() {
	if !rc.f.RetType.Type.MultiTyped {
		for _, v := range rc.f.RetType.Identifiers {
			if !lex.IsIgnoreId(v.Kind) {
				model := new(exprModel)
				model.index = 0
				model.nodes = make([]exprBuildNode, 1)
				val, _ := rc.t.eval.single(v, model)
				rc.exp_model.models = append(rc.exp_model.models, model)
				rc.values = append(rc.values, val)
				break
			}
		}
		rc.ret_ast.Expr.Model = rc.exp_model
		return
	}
	types := rc.f.RetType.Type.Tag.([]Type)
	for i, v := range rc.f.RetType.Identifiers {
		if lex.IsIgnoreId(v.Kind) {
			node := exprNode{}
			node.value = types[i].String()
			node.value += build.CPP_DEFAULT_EXPR
			rc.exp_model.models = append(rc.exp_model.models, node)
			continue
		}
		model := new(exprModel)
		model.index = 0
		model.nodes = make([]exprBuildNode, 1)
		val, _ := rc.t.eval.single(v, model)
		rc.exp_model.models = append(rc.exp_model.models, model)
		rc.values = append(rc.values, val)
	}
	rc.ret_ast.Expr.Model = rc.exp_model
}

func (rc *retChecker) check() {
	n := len(rc.ret_ast.Expr.Tokens)
	if n == 0 && !types.IsVoid(rc.f.RetType.Type) {
		if !rc.f.RetType.AnyVar() {
			rc.t.pusherrtok(rc.ret_ast.Token, "require_return_value")
		}
		rc.retsVars()
		return
	}
	if n > 0 && types.IsVoid(rc.f.RetType.Type) {
		rc.t.pusherrtok(rc.ret_ast.Token, "void_function_return_value")
	}
	rc.exp_model.vars = rc.f.RetType.Vars(rc.t.nodeBlock)
	rc.checkepxrs()
}
