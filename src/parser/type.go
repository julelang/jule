// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package parser

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
)

func build_void_type() *ast.Type { return &ast.Type{} }

func build_prim_type(token lex.Token) *ast.Type {
	return &ast.Type{
		Token: token,
		Kind: &ast.IdentType{
			Token:      token,
			Ident:      token.Kind,
			Cpp_linked: false,
			Generics:   nil,
		},
	}
}

// Returns colon index and range tokens.
// Returns nil slice and -1 if not found.
// Starts search at *i.
func split_map_range(tokens []lex.Token, i *int) (range_tokens []lex.Token, colon int) {
	colon = -1
	range_n := 0
	start := *i
	for ; *i < len(tokens); *i++ {
		token := tokens[*i]
		if token.Id == lex.ID_RANGE {
			switch token.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				range_n++
				continue
			default:
				range_n--
			}
		}
		if range_n == 0 {
			if start+1 > *i {
				return
			}
			range_tokens = tokens[start+1 : *i]
			break
		} else if range_n != 1 {
			continue
		}
		if colon == -1 && token.Id == lex.ID_COLON {
			colon = *i - start - 1
		}
	}
	return
}

type _TypeBuilder struct {
	p      *_Parser
	tokens []lex.Token
	i      *int
	err    bool
}

func (tb *_TypeBuilder) push_err(token lex.Token, key string) {
	if tb.err {
		tb.p.push_err(token, key)
	}
}

func (tb *_TypeBuilder) build_prim() *ast.Type {
	t := build_prim_type(tb.tokens[*tb.i])
	*tb.i++
	return t
}

func (tb *_TypeBuilder) build_namespace() *ast.Type {
	t := &ast.Type{
		Token: tb.tokens[*tb.i],
	}

	nst := &ast.NamespaceType{}
	n := 0
	for ; *tb.i < len(tb.tokens); *tb.i++ {
		token := tb.tokens[*tb.i]
		if n%2 == 0 {
			if token.Id != lex.ID_IDENT {
				tb.push_err(token, "invalid_syntax")
			}
			nst.Idents = append(nst.Idents, token)
		} else if token.Id != lex.ID_DBLCOLON {
			break
		}
		n++
	}

	// Remove selected identifier token.
	nst.Idents = nst.Idents[:len(nst.Idents)-1]

	*tb.i-- // Set offset to last identifier.
	nst.Kind = tb.build_ident().Kind.(*ast.IdentType)
	t.Kind = nst
	return t
}

func (tb *_TypeBuilder) build_generics() []*ast.Type {
	if *tb.i >= len(tb.tokens) {
		return nil
	}
	tokens := tb.tokens[*tb.i]
	if tokens.Id != lex.ID_RANGE || tokens.Kind != lex.KND_LBRACKET {
		return nil
	}

	parts := tb.ident_generics()
	types := make([]*ast.Type, len(parts))
	for i, part := range parts {
		j := 0
		t, _ := tb.p.build_type(part, &j, true)
		if j < len(part) {
			tb.push_err(part[j], "invalid_syntax")
		}
		types[i] = t
	}
	return types
}

func (tb *_TypeBuilder) ident_generics() [][]lex.Token {
	first := *tb.i
	range_n := 0
	for ; *tb.i < len(tb.tokens); *tb.i++ {
		token := tb.tokens[*tb.i]
		if token.Id == lex.ID_RANGE {
			switch token.Kind {
			case lex.KND_LBRACKET:
				range_n++

			case lex.KND_RBRACKET:
				range_n--
			}
		}
		if range_n == 0 {
			*tb.i++ // Skip right bracket
			break
		}
	}
	tokens := tb.tokens[first+1 : *tb.i-1] // Take range of brackets.
	parts, errors := lex.Parts(tokens, lex.ID_COMMA, true)
	if tb.err {
		tb.p.errors = append(tb.p.errors, errors...)
	}
	return parts
}

func (tb *_TypeBuilder) build_ident() *ast.Type {
	if *tb.i+1 < len(tb.tokens) && tb.tokens[*tb.i+1].Id == lex.ID_DBLCOLON {
		return tb.build_namespace()
	}
	token := tb.tokens[*tb.i]
	it := &ast.IdentType{
		Token:      token,
		Ident:      token.Kind,
		Cpp_linked: false,
	}
	*tb.i++
	it.Generics = tb.build_generics()
	return &ast.Type{
		Token: token,
		Kind:  it,
	}
}

func (tb *_TypeBuilder) build_cpp_link() *ast.Type {
	if *tb.i+1 >= len(tb.tokens) || tb.tokens[*tb.i+1].Id != lex.ID_DOT {
		tb.push_err(tb.tokens[*tb.i], "invalid_syntax")
		return nil
	}
	*tb.i += 2 // Skip cpp keyword and dot token.
	t := tb.build_ident()
	t.Kind.(*ast.IdentType).Cpp_linked = true
	return t
}

func (tb *_TypeBuilder) build_fn() *ast.Type {
	token := tb.tokens[*tb.i]
	f := tb.p.build_fn_prototype(tb.tokens, tb.i, false, true)
	if f == nil {
		return nil
	}
	return &ast.Type{
		Token: token,
		Kind:  f,
	}
}

func (tb *_TypeBuilder) build_ptr() *ast.Type {
	token := tb.tokens[*tb.i]
	if *tb.i+1 >= len(tb.tokens) {
		tb.push_err(token, "invalid_syntax")
		return nil
	}

	*tb.i++
	if tb.tokens[*tb.i].Id == lex.ID_UNSAFE {
		*tb.i++
		return &ast.Type{
			Token: token,
			Kind: &ast.PtrType{
				Elem: nil, // Set Elem as nil for unsafe pointer (*unsafe) type.
			},
		}
	}

	elem := tb.step()
	if elem == nil {
		return nil
	}

	return &ast.Type{
		Token: token,
		Kind: &ast.PtrType{
			Elem: elem,
		},
	}
}

func (tb *_TypeBuilder) build_ref() *ast.Type {
	token := tb.tokens[*tb.i]
	if *tb.i+1 >= len(tb.tokens) {
		tb.push_err(token, "invalid_syntax")
		return nil
	}

	*tb.i++
	elem := tb.step()
	if elem == nil {
		return nil
	}

	return &ast.Type{
		Token: token,
		Kind: &ast.RefType{
			Elem: elem,
		},
	}
}

func (tb *_TypeBuilder) build_op() *ast.Type {
	token := tb.tokens[*tb.i]
	switch token.Kind {
	case lex.KND_STAR:
		return tb.build_ptr()

	case lex.KND_AMPER:
		return tb.build_ref()

	case lex.KND_DBL_AMPER:
		tb.push_err(token, "ref_refs_ref")
		return tb.build_ref() // Skip tokens and many type error

	default:
		tb.push_err(token, "invalid_syntax")
	}
	return nil
}

func (tb *_TypeBuilder) build_slc() *ast.Type {
	token := tb.tokens[*tb.i]
	*tb.i++ // skip right bracket
	elem := tb.step()
	if elem == nil {
		return nil
	}
	return &ast.Type{
		Token: token,
		Kind: &ast.SlcType{
			Elem: elem,
		},
	}
}

func (tb *_TypeBuilder) build_arr() *ast.Type {
	// *tb.i points to element type of array.
	// Brackets places at ... < *tb.i offset.

	if *tb.i >= len(tb.tokens) {
		tb.push_err(tb.tokens[*tb.i-1], "missing_type")
		return nil
	}

	expr_delta := *tb.i

	elem := tb.step()
	if elem == nil {
		return nil
	}

	arrt := &ast.ArrType{
		Elem: elem,
	}

	_, expr_tokens := lex.Range_last(tb.tokens[:expr_delta])
	expr_tokens = expr_tokens[1 : len(expr_tokens)-1] // Remove brackets.
	token := expr_tokens[0]
	if len(expr_tokens) == 1 && token.Id == lex.ID_OP && token.Kind == lex.KND_TRIPLE_DOT {
		// Ignore.
	} else {
		arrt.Size = tb.p.build_expr(expr_tokens)
	}

	return &ast.Type{
		Token: token,
		Kind:  arrt,
	}
}

func (tb *_TypeBuilder) build_map(colon int, tokens []lex.Token) *ast.Type {
	colon_token := tb.tokens[colon]
	if colon == 0 || colon+1 >= len(tokens) {
		tb.push_err(colon_token, "missing_type")
		return nil
	}
	key_tokens := tokens[:colon]
	val_tokens := tokens[colon+1:]
	mapt := &ast.MapType{}

	j := 0
	keyt, ok := tb.p.build_type(key_tokens, &j, tb.err)
	if !ok {
		return nil
	} else if j < len(key_tokens) {
		tb.push_err(key_tokens[j], "invalid_syntax")
	}
	mapt.Key = keyt

	j = 0
	valt, ok := tb.p.build_type(val_tokens, &j, tb.err)
	if !ok {
		return nil
	} else if j < len(val_tokens) {
		tb.push_err(val_tokens[j], "invalid_syntax")
	}
	mapt.Val = valt

	return &ast.Type{
		Token: colon_token,
		Kind:  mapt,
	}
}

func (tb *_TypeBuilder) build_enumerable() *ast.Type {
	token := tb.tokens[*tb.i]
	if *tb.i+2 >= len(tb.tokens) || token.Id != lex.ID_RANGE || token.Kind != lex.KND_LBRACKET {
		tb.push_err(token, "invalid_syntax")
		return nil
	}
	*tb.i++
	token = tb.tokens[*tb.i]
	if token.Id == lex.ID_RANGE && token.Kind == lex.KND_RBRACKET {
		return tb.build_slc()
	}

	*tb.i-- // Point to left bracket for range parsing of split_colon.
	map_tokens, colon := split_map_range(tb.tokens, tb.i)
	*tb.i++
	if colon == -1 {
		return tb.build_arr()
	}
	return tb.build_map(colon, map_tokens)
}

func (tb *_TypeBuilder) step() *ast.Type {
	token := tb.tokens[*tb.i]
	switch token.Id {
	case lex.ID_PRIM:
		return tb.build_prim()

	case lex.ID_IDENT:
		return tb.build_ident()

	case lex.ID_CPP:
		return tb.build_cpp_link()

	case lex.ID_FN:
		return tb.build_fn()

	case lex.ID_OP:
		return tb.build_op()

	case lex.ID_RANGE:
		return tb.build_enumerable()

	default:
		*tb.i++
		tb.push_err(token, "invalid_syntax")
		return nil
	}
}

// Builds type.
// Returns void if error occurs.
func (tb *_TypeBuilder) build() (_ *ast.Type, ok bool) {
	root := tb.step()
	if root == nil {
		return build_void_type(), false
	}
	return root, true
}
