package parser

import (
	"strings"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// Reports whether token is statement finish point.
func is_st(current lex.Token, prev lex.Token) (ok bool, terminated bool) {
	ok = current.Id == lex.ID_SEMICOLON || prev.Row < current.Row
	terminated = current.Id == lex.ID_SEMICOLON
	return
}

// Reports position of the next statement if exist, len(toks) if not.
func next_st_pos(tokens []lex.Token, start int) (int, bool) {
	brace_n := 0
	i := start
	for ; i < len(tokens); i++ {
		var ok, terminated bool
		tok := tokens[i]
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				if brace_n == 0 && i > start {
					ok, terminated = is_st(tok, tokens[i-1])
					if ok {
						goto ret
					}
				}
				brace_n++
				continue
			default:
				brace_n--
				if brace_n == 0 && i+1 < len(tokens) {
					ok, terminated = is_st(tokens[i+1], tok)
					if ok {
						i++
						goto ret
					}
				}
				continue
			}
		}
		if brace_n != 0 {
			continue
		} else if i > start {
			ok, terminated = is_st(tok, tokens[i-1])
		} else {
			ok, terminated = is_st(tok, tok)
		}
		if !ok {
			continue
		}
	ret:
		if terminated {
			i++
		}
		return i, terminated
	}
	return i, false
}

// Returns current statement tokens.
// Starts selection at *i.
func skip_st(i *int, tokens []lex.Token) []lex.Token {
	start := *i
	*i, _ = next_st_pos(tokens, start)
	st_tokens := tokens[start:*i]
	if st_tokens[len(st_tokens)-1].Id == lex.ID_SEMICOLON {
		if len(st_tokens) == 1 {
			return skip_st(i, tokens)
		}
		// -1 for eliminate statement terminator.
		st_tokens = st_tokens[:len(st_tokens)-1]
	}
	return st_tokens
}

// Splits all statements.
func split_stms(tokens []lex.Token) [][]lex.Token {
	var stms [][]lex.Token = nil
	pos := 0
	for pos < len(tokens) {
		stms = append(stms, skip_st(&pos, tokens))
	}
	return stms
}

func compiler_err(token lex.Token, key string, args ...any) build.Log {
	return build.Log{
		Type:   build.ERR,
		Row:    token.Row,
		Column: token.Column,
		Path:   token.File.Path(),
		Text:   build.Errorf(key, args...),
	}
}

type parser struct {
	file          *lex.File
	directives    []*ast.Directive
	comment_group *ast.CommentGroup
	tree          []ast.Node
	errors        []build.Log
}

// Appends error by specified token, key and args.
func (p *parser) push_err(token lex.Token, key string, args ...any) {
	p.errors = append(p.errors, compiler_err(token, key, args...))
}

func (p *parser) push_directive(token lex.Token) {
	d := &ast.Directive{
		Token: token,
		Tag:   token.Kind[len(lex.DIRECTIVE_COMMENT_PREFIX):], // Remove directive prefix
	}

	// Don't append if directive kind is invalid.
	ok := false
	for _, kind := range build.ATTRS {
		if d.Tag == kind {
			ok = true
			break
		}
	}
	if !ok {
		return
	}

	// Don't append if already added this directive.
	for _, pd := range p.directives {
		if d.Tag == pd.Tag {
			return
		}
	}

	p.directives = append(p.directives, d)
}

func (p *parser) build_comment(token lex.Token) ast.NodeData {
	// Remove slashes and trim spaces.
	token.Kind = strings.TrimSpace(token.Kind[2:])

	if strings.HasPrefix(token.Kind, lex.DIRECTIVE_COMMENT_PREFIX) {
		p.push_directive(token)
	} else {
		if p.comment_group == nil {
			p.comment_group = &ast.CommentGroup{}
		}
		p.comment_group.Comments = append(p.comment_group.Comments, &ast.Comment{
			Token: token,
			Text:  token.Kind,
		})
	}

	return &ast.Comment{
		Token: token,
		Text:  token.Kind,
	}
}

func (p *parser) build_node_data(st []lex.Token) ast.NodeData {
	token := st[0]
	switch token.Id {
	case lex.ID_COMMENT:
		// Push first token because this is full text comment.
		// Comments are just single-line.
		// Range comments not accepts by lexer.
		return p.build_comment(token)
	default:
		p.push_err(token, "invalid_syntax")
		return nil
	}
}

func (p *parser) check_comment_group(node ast.Node) {
	if p.comment_group == nil {
		return
	}
	switch node.Data.(type) {
	case ast.Comment, ast.Directive:
		// Ignore
	default:
		p.comment_group = nil
	}
}

func (p *parser) check_directive(node ast.Node) {
	if p.directives == nil {
		return
	}
	switch node.Data.(type) {
	case ast.Directive, ast.Comment:
		// Ignore
	default:
		p.directives = nil
	}
}

func (p *parser) append_node(st []lex.Token) {
	if len(st) == 0 {
		return
	}

	node := ast.Node{
		Token: st[0],
		Data:  p.build_node_data(st),
	}

	if node.Data != nil {
		p.check_comment_group(node)
		p.check_directive(node)
		p.tree = append(p.tree, node)
	}
}

func (p *parser) parse() {
	stms := split_stms(p.file.Tokens())
	for _, st := range stms {
		p.append_node(st)
	}
}
