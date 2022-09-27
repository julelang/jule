package transpiler

import (
	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/lex"
)

func make_slice(p *Transpiler, m *exprModel, t models.Type, args *models.Args, errtok lex.Token) (v value) {
	v.data.Type = t
	v.data.Value = " "
	if len(args.Src) < 2 {
		p.pusherrtok(errtok, "missing_expr_for", "len")
		return
	} else if len(args.Src) > 2 {
		p.pusherrtok(errtok, "argument_overflow")
	}
	len_expr := args.Src[1].Expr
	len_v, len_expr_model := p.evalExpr(len_expr, nil)
	err_key := check_value_for_indexing(len_v)
	if err_key != "" {
		p.pusherrtok(errtok, err_key)
	} else if typeIsRef(*t.ComponentType) {
		p.pusherrtok(errtok, "reference_not_initialized")
	}
	// Remove function identifier from model.
	m.nodes[m.index].nodes[0] = nil
	m.appendSubNode(exprNode{t.String()})
	m.appendSubNode(exprNode{"("})
	m.appendSubNode(len_expr_model)
	m.appendSubNode(exprNode{")"})
	return
}
