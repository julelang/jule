package parser

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// BuildFn builds AST model of function.
func BuildFn(tokens []lex.Token, method bool, anon bool, prototype bool) (f ast.Fn, errors []build.Log) {
	p := parser{}
	f = p.build_fn(tokens, method, anon, prototype)
	return f, p.errors
}

// BuildArgs builds AST model of arguments.
func BuildArgs(tokens []lex.Token, targeting bool) (*ast.Args, []build.Log) {
	args := &ast.Args{}
	var errors []build.Log

	last := 0
	brace_n := 0
	for i, tok := range tokens {
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
		push_arg(args, targeting, tokens[last:i], tok, &errors)
		last = i + 1
	}
	if last < len(tokens) {
		if last == 0 {
			if len(tokens) > 0 {
				push_arg(args, targeting, tokens[last:], tokens[last], &errors)
			}
		} else {
			push_arg(args, targeting, tokens[last:], tokens[last-1], &errors)
		}
	}
	return args, errors
}

func push_arg(args *ast.Args, targeting bool, toks []lex.Token, err_tok lex.Token, errors *[]build.Log) {
	if len(toks) == 0 {
		*errors = append(*errors, compiler_err(err_tok, "invalid_syntax"))
		return
	}
	var arg ast.Arg
	arg.Token = toks[0]
	if targeting && arg.Token.Id == lex.ID_IDENT {
		if len(toks) > 1 {
			tok := toks[1]
			if tok.Id == lex.ID_COLON {
				args.Targeted = true
				arg.TargetId = arg.Token.Kind
				toks = toks[2:]
			}
		}
	}
	arg.Expr = BuildExpr(toks)
	args.Src = append(args.Src, arg)
}

func BuildType(tokens []lex.Token, i *int) (ast.Type, []build.Log) {
	p := parser{}
	t, _ := p.build_type(tokens, i, true)
	return t, p.errors
}
