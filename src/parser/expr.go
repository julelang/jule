package parser

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

// Expr builds AST model of expression.
func BuildExpr(tokens []lex.Token) ast.Expr {
	return ast.Expr{
		Tokens: tokens,
		Op:     build_expr_op(tokens),
	}
}

func build_binop_expr(toks []lex.Token) any {
	i := find_lowest_precedenced_operator(toks)
	if i != -1 {
		return build_binop(toks)
	}
	return ast.BinopExpr{Tokens: toks}
}

func build_binop(toks []lex.Token) ast.Binop {
	op := ast.Binop{}
	i := find_lowest_precedenced_operator(toks)
	op.L = build_binop_expr(toks[:i])
	op.R = build_binop_expr(toks[i+1:])
	op.Op = toks[i]
	return op
}

func eliminate_comments(toks []lex.Token) []lex.Token {
	cutted := []lex.Token{}
	for _, token := range toks {
		if token.Id != lex.ID_COMMENT {
			cutted = append(cutted, token)
		}
	}
	return cutted
}

// Returns BinopExpr or Binop instance for expression Op.
func build_expr_op(toks []lex.Token) any {
	toks = eliminate_comments(toks)
	i := find_lowest_precedenced_operator(toks)
	if i == -1 {
		return build_binop_expr(toks)
	}
	return build_binop(toks)
}

// Finds index of priority operator and returns index of operator
// if found, returns -1 if not.
func find_lowest_precedenced_operator(toks []lex.Token) int {
	prec := precedencer{}
	brace_n := 0
	for i, tok := range toks {
		switch {
		case tok.Id == lex.ID_BRACE:
			switch tok.Kind {
			case lex.KND_LBRACE, lex.KND_LPAREN, lex.KND_LBRACKET:
				brace_n++
			default:
				brace_n--
			}
			continue
		case i == 0:
			continue
		case tok.Id != lex.ID_OP:
			continue
		case brace_n > 0:
			continue
		}
		// Skip unary operator.
		if toks[i-1].Id == lex.ID_OP {
			continue
		}
		p := tok.Prec()
		if p != -1 {
			prec.set(p, i)
		}
	}
	data := prec.get_lower()
	if data == nil {
		return -1
	}
	return data.(int)
}
