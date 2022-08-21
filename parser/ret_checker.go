package parser

import (
	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/juleapi"
)

type retChecker struct {
	p        *Parser
	retAST   *models.Ret
	f        *Func
	expModel retExpr
}

func (rc *retChecker) pushval(last, current int, errTok Tok) {
	if current-last == 0 {
		rc.p.pusherrtok(errTok, "missing_expr")
		return
	}
	toks := rc.retAST.Expr.Toks[last:current]
	val, model := rc.p.evalToks(toks)
	rc.expModel.models = append(rc.expModel.models, model)
	rc.expModel.values = append(rc.expModel.values, val)
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
		rc.retAST.Expr.Model = rc.expModel
	}
}

func (rc *retChecker) single() {
	rc.expModel.models = append(rc.expModel.models, rc.expModel.models[0])
	if len(rc.expModel.values) > 1 {
		rc.p.pusherrtok(rc.retAST.Tok, "overflow_return")
	}
	assignChecker{
		p:      rc.p,
		t:      rc.f.RetType.Type,
		v:      rc.expModel.values[0],
		errtok: rc.retAST.Tok,
	}.checkAssignType()
}

func (rc *retChecker) multi() {
	types := rc.f.RetType.Type.Tag.([]DataType)
	valLength := len(rc.expModel.values)
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
		assignChecker{
			p:      rc.p,
			t:      t,
			v:      rc.expModel.values[i],
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
	val := rc.expModel.values[0]
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
	rc.expModel.models = append(rc.expModel.models, rc.expModel.models[0])
	for i, rt := range retTypes {
		vt := valTypes[i]
		val := value{data: models.Data{Type: vt}}
		assignChecker{
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
			if !juleapi.IsIgnoreId(v.Kind) {
				model := new(exprModel)
				model.index = 0
				model.nodes = make([]exprBuildNode, 1)
				val, _ := rc.p.eval.single(v, model)
				rc.expModel.models = append(rc.expModel.models, model)
				rc.expModel.values = append(rc.expModel.values, val)
				break
			}
		}
		rc.retAST.Expr.Model = rc.expModel
		return
	}
	types := rc.f.RetType.Type.Tag.([]DataType)
	for i, v := range rc.f.RetType.Identifiers {
		if juleapi.IsIgnoreId(v.Kind) {
			node := exprNode{}
			node.value = types[i].String()
			node.value += juleapi.DefaultExpr
			rc.expModel.models = append(rc.expModel.models, node)
			continue
		}
		model := new(exprModel)
		model.index = 0
		model.nodes = make([]exprBuildNode, 1)
		val, _ := rc.p.eval.single(v, model)
		rc.expModel.models = append(rc.expModel.models, model)
		rc.expModel.values = append(rc.expModel.values, val)
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
