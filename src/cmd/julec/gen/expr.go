package gen

import (
	"strings"

	"github.com/julelang/jule/ast"
)

type AnonFuncExpr struct {
	Ast *ast.Fn
}

func (af AnonFuncExpr) String() string {
	var cpp strings.Builder
	t := ast.Type{
		Token: af.Ast.Token,
		Kind:  af.Ast.TypeKind(),
		Tag:   af.Ast,
	}
	cpp.WriteString(t.FnString())
	cpp.WriteString("([=]")
	cpp.WriteString(gen_params(af.Ast.Params))
	cpp.WriteString(" mutable -> ")
	cpp.WriteString(af.Ast.RetType.String())
	cpp.WriteByte(' ')
	vars := af.Ast.RetType.Vars(af.Ast.Block)
	cpp.WriteString(gen_fn_block(vars, af.Ast.Block))
	cpp.WriteByte(')')
	return cpp.String()
}
