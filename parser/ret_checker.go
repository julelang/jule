package parser

import (
	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/xapi"
)

type retChecker struct {
	p        *Parser
	retAST   *models.Ret
	f        *Func
	expModel multiRetExpr
	values   []value
}

func (rc *retChecker) pushval(last, current int, errTk Tok) {
	if current-last == 0 {
		rc.p.pusherrtok(errTk, "missing_expr")
		return
	}
	toks := rc.retAST.Expr.Toks[last:current]
	val, model := rc.p.evalToks(toks)
	rc.expModel.models = append(rc.expModel.models, model)
	rc.values = append(rc.values, val)
}

func (rc *retChecker) checkepxrs() {
	braceCount := 0
	last := 0
	for i, tok := range rc.retAST.Expr.Toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 || tok.Id != tokens.Comma {
			continue
		}
		rc.pushval(last, i, tok)
		last = i + 1
	}
	length := len(rc.retAST.Expr.Toks)
	if last < length {
		if last == 0 {
			rc.pushval(0, length, rc.retAST.Tok)
		} else {
			rc.pushval(last, length, rc.retAST.Expr.Toks[last-1])
		}
	}
	if !typeIsVoid(rc.f.RetType.Type) {
		rc.checkExprTypes()
	}
}

func (rc *retChecker) single() {
	rc.retAST.Expr.Model = rc.expModel.models[0]
	if len(rc.values) > 1 {
		rc.p.pusherrtok(rc.retAST.Tok, "overflow_return")
	}
	rc.p.wg.Add(1)
	go assignChecker{
		p:      rc.p,
		t:      rc.f.RetType.Type,
		v:      rc.values[0],
		errtok: rc.retAST.Tok,
	}.checkAssignType()
}

func (rc *retChecker) multi() {
	rc.retAST.Expr.Model = rc.expModel
	types := rc.f.RetType.Type.Tag.([]DataType)
	valLength := len(rc.values)
	if valLength == 1 {
		rc.checkMultiRetAsMutliRet()
		return
	} else if valLength > len(types) {
		rc.p.pusherrtok(rc.retAST.Tok, "overflow_return")
	}
	for i, t := range types {
		if i >= valLength {
			break
		}
		rc.p.wg.Add(1)
		go assignChecker{
			p:      rc.p,
			t:      t,
			v:      rc.values[i],
			errtok: rc.retAST.Tok,
		}.checkAssignType()
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
	val := rc.values[0]
	if !val.data.Type.MultiTyped {
		rc.p.pusherrtok(rc.retAST.Tok, "missing_multi_return")
		return
	}
	valTypes := val.data.Type.Tag.([]DataType)
	retTypes := rc.f.RetType.Type.Tag.([]DataType)
	if len(valTypes) < len(retTypes) {
		rc.p.pusherrtok(rc.retAST.Tok, "missing_multi_return")
		return
	} else if len(valTypes) < len(retTypes) {
		rc.p.pusherrtok(rc.retAST.Tok, "overflow_return")
		return
	}
	// Set model for just signle return
	rc.retAST.Expr.Model = rc.expModel.models[0]
	for i, rt := range retTypes {
		vt := valTypes[i]
		val := value{data: models.Data{Type: vt}}
		rc.p.wg.Add(1)
		go assignChecker{
			p:      rc.p,
			t:      rt,
			v:      val,
			errtok: rc.retAST.Tok,
		}.checkAssignType()
	}
}

func (rc *retChecker) retsVars() {
	if !rc.f.RetType.Type.MultiTyped {
		for _, v := range rc.f.RetType.Identifiers {
			if !xapi.IsIgnoreId(v.Kind) {
				model := new(exprModel)
				model.index = 0
				model.nodes = make([]exprBuildNode, 1)
				_, _ = rc.p.evalSingleExpr(v, model)
				rc.retAST.Expr.Model = model
				break
			}
		}
		return
	}
	types := rc.f.RetType.Type.Tag.([]DataType)
	for i, v := range rc.f.RetType.Identifiers {
		if xapi.IsIgnoreId(v.Kind) {
			node := exprNode{}
			node.value = types[i].String()
			node.value += xapi.DefaultExpr
			rc.expModel.models = append(rc.expModel.models, node)
			continue
		}
		model := new(exprModel)
		model.index = 0
		model.nodes = make([]exprBuildNode, 1)
		_, _ = rc.p.evalSingleExpr(v, model)
		rc.expModel.models = append(rc.expModel.models, model)
	}
	rc.retAST.Expr.Model = rc.expModel
}

func (rc *retChecker) check() {
	exprToksLen := len(rc.retAST.Expr.Toks)
	if exprToksLen == 0 && !typeIsVoid(rc.f.RetType.Type) {
		if !rc.f.RetType.AnyVar() {
			rc.p.pusherrtok(rc.retAST.Tok, "require_return_value")
		}
		rc.retsVars()
		return
	}
	if exprToksLen > 0 && typeIsVoid(rc.f.RetType.Type) {
		rc.p.pusherrtok(rc.retAST.Tok, "void_function_return_value")
	}
	rc.checkepxrs()
}
