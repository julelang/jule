package ast

import (
	"strings"

	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

type type_builder struct {
	b      *Builder
	t      *models.Type
	tokens []lex.Token
	i      *int
	err    bool
	first  int
	kind   string
	ok     bool
}

func (tb *type_builder) dt(tok lex.Token) {
	tb.t.Token = tok
	tb.t.Id = types.TypeFromId(tb.t.Token.Kind)
	tb.kind += tb.t.Token.Kind
	tb.ok = true
}

func (tb *type_builder) unsafe_kw(tok lex.Token) {
	tb.t.Id = types.UNSAFE
	tb.t.Token = tok
	tb.kind += tok.Kind
	tb.ok = true
}

func (tb *type_builder) op(tok lex.Token) (imret bool) {
	switch tok.Kind {
	case lex.KND_STAR, lex.KND_AMPER, lex.KND_DBL_AMPER:
		tb.kind += tok.Kind
	default:
		if tb.err {
			tb.b.pusherr(tok, "invalid_syntax")
		}
		return true
	}
	return false
}

func (tb *type_builder) function(tok lex.Token) {
	tb.t.Token = tok
	tb.t.Id = types.FN
	f, proto_ok := tb.b.fn_prototype(tb.tokens, tb.i, false, true)
	if !proto_ok {
		tb.b.pusherr(tok, "invalid_type")
		return
	}
	*tb.i--
	tb.t.Tag = &f
	tb.kind += f.TypeKind()
	tb.ok = true
}

func (tb *type_builder) ident(tok lex.Token) {
	tb.kind += tok.Kind
	if *tb.i+1 < len(tb.tokens) && tb.tokens[*tb.i+1].Id == lex.ID_DBLCOLON {
		return
	}
	tb.t.Id = types.ID
	tb.t.Token = tok
	tb.ident_end()
	tb.ok = true
}

func (tb *type_builder) ident_end() {
	if *tb.i+1 >= len(tb.tokens) {
		return
	}
	*tb.i++
	tok := tb.tokens[*tb.i]
	if tok.Id != lex.ID_BRACE || tok.Kind != lex.KND_LBRACKET {
		*tb.i--
		return
	}
	tb.kind += "["
	var genericsStr strings.Builder
	parts := tb.ident_generics()
	generics := make([]models.Type, len(parts))
	for i, part := range parts {
		index := 0
		t, _ := tb.b.DataType(part, &index, true)
		if index+1 < len(part) {
			tb.b.pusherr(part[index+1], "invalid_syntax")
		}
		genericsStr.WriteString(t.String())
		genericsStr.WriteByte(',')
		generics[i] = t
	}
	tb.kind +=  genericsStr.String()[:genericsStr.Len()-1] + "]"
	tb.t.Tag = generics
}

func (tb *type_builder) ident_generics() [][]lex.Token {
	first := *tb.i
	brace_n := 0
	for ; *tb.i < len(tb.tokens); *tb.i++ {
		tok := tb.tokens[*tb.i]
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_LBRACKET:
				brace_n++
			case lex.KND_RBRACKET:
				brace_n--
			}
		}
		if brace_n == 0 {
			break
		}
	}
	tokens := tb.tokens[first+1 : *tb.i]
	parts, errs := Parts(tokens, lex.ID_COMMA, true)
	tb.b.Errors = append(tb.b.Errors, errs...)
	return parts
}

func (tb *type_builder) cpp_kw(tok lex.Token) (imret bool) {
	if *tb.i+1 >= len(tb.tokens) {
		if tb.err {
			tb.b.pusherr(tok, "invalid_syntax")
		}
		return true
	}
	*tb.i++
	if tb.tokens[*tb.i].Id != lex.ID_DOT {
		if tb.err {
			tb.b.pusherr(tb.tokens[*tb.i], "invalid_syntax")
		}
	}
	if *tb.i+1 >= len(tb.tokens) {
		if tb.err {
			tb.b.pusherr(tok, "invalid_syntax")
		}
		return true
	}
	*tb.i++
	if tb.tokens[*tb.i].Id != lex.ID_IDENT {
		if tb.err {
			tb.b.pusherr(tb.tokens[*tb.i], "invalid_syntax")
		}
	}
	tb.t.CppLinked = true
	tb.t.Id = types.ID
	tb.t.Token = tb.tokens[*tb.i]
	tb.kind += tb.t.Token.Kind
	tb.ident_end()
	tb.ok = true
	return false
}

func (tb *type_builder) enumerable(tok lex.Token) (imret bool) {
	*tb.i++
	if *tb.i >= len(tb.tokens) {
		if tb.err {
			tb.b.pusherr(tok, "invalid_syntax")
		}
		return
	}
	tok = tb.tokens[*tb.i]
	if tok.Id == lex.ID_BRACE && tok.Kind == lex.KND_RBRACKET {
		tb.kind += lex.PREFIX_SLICE
		tb.t.ComponentType = new(models.Type)
		tb.t.Id = types.SLICE
		tb.t.Token = tok
		*tb.i++
		if *tb.i >= len(tb.tokens) {
			if tb.err {
				tb.b.pusherr(tok, "invalid_syntax")
			}
			return
		}
		*tb.t.ComponentType, tb.ok = tb.b.DataType(tb.tokens, tb.i, tb.err)
		tb.kind += tb.t.ComponentType.Kind
		return false
	}
	*tb.i-- // Start from bracket
	tb.ok = tb.map_or_array()
	if tb.t.Id == types.VOID {
		if tb.err {
			tb.b.pusherr(tok, "invalid_syntax")
		}
		return true
	}
	tb.t.Token = tok
	return false
}

func (tb *type_builder) array() (ok bool) {
	defer func() { tb.t.Original = *tb.t }()
	if *tb.i+1 >= len(tb.tokens) {
		return
	}
	tb.t.Id = types.ARRAY
	*tb.i++
	exprI := *tb.i
	tb.t.ComponentType = new(models.Type)
	ok = tb.b.datatype(tb.t.ComponentType, tb.tokens, tb.i, tb.err)
	if !ok {
		return
	}
	_, exprToks := RangeLast(tb.tokens[:exprI])
	exprToks = exprToks[1 : len(exprToks)-1]
	tok := exprToks[0]
	if len(exprToks) == 1 && tok.Id == lex.ID_OP && tok.Kind == lex.KND_TRIPLE_DOT {
		tb.t.Size.AutoSized = true
		tb.t.Size.Expr.Tokens = exprToks
	} else {
		tb.t.Size.Expr = tb.b.Expr(exprToks)
	}
	tb.kind = tb.kind + lex.PREFIX_ARRAY + tb.t.ComponentType.Kind
	return
}

func (tb *type_builder) map_or_array() (ok bool) {
	ok = tb.map_t()
	if !ok {
		ok = tb.array()
	}
	return
}

// MapDataType builds map data-type.
func (tb *type_builder) map_t() (ok bool) {
	typeToks, colon := SplitColon(tb.tokens, tb.i)
	if typeToks == nil || colon == -1 {
		return
	}
	defer func() { tb.t.Original = *tb.t }()
	tb.t.Id = types.MAP
	tb.t.Token = tb.tokens[0]
	colonTok := tb.tokens[colon]
	if colon == 0 || colon+1 >= len(typeToks) {
		if tb.err {
			tb.b.pusherr(colonTok, "missing_expr")
		}
		return
	}
	keyTypeToks := typeToks[:colon]
	valueTypeToks := typeToks[colon+1:]
	types := make([]models.Type, 2)
	j := 0
	types[0], _ = tb.b.DataType(keyTypeToks, &j, tb.err)
	j = 0
	types[1], _ = tb.b.DataType(valueTypeToks, &j, tb.err)
	tb.t.Tag = types
	tb.kind = tb.kind + tb.t.MapKind()
	ok = true
	return
}

func (tb *type_builder) step() (imret bool) {
	tok := tb.tokens[*tb.i]
	switch tok.Id {
	case lex.ID_DT:
		tb.dt(tok)
		return
	case lex.ID_IDENT:
		tb.ident(tok)
		return
	case lex.ID_CPP:
		imret = tb.cpp_kw(tok)
		return
	case lex.ID_DBLCOLON:
		tb.kind += tok.Kind
		return
	case lex.ID_UNSAFE:
		if *tb.i+1 >= len(tb.tokens) || tb.tokens[*tb.i+1].Id != lex.ID_FN {
			tb.unsafe_kw(tok)
			return
		}
		fallthrough
	case lex.ID_FN:
		tb.function(tok)
		return
	case lex.ID_OP:
		imret = tb.op(tok)
		return
	case lex.ID_BRACE:
		switch tok.Kind {
		case lex.KND_LBRACKET:
			imret = tb.enumerable(tok)
			return
		}
		imret = true
		return
	default:
		if tb.err {
			tb.b.pusherr(tok, "invalid_syntax")
		}
		imret = true
		return
	}
}

func (tb *type_builder) build() bool {
	defer func() { tb.t.Original = *tb.t }()
	tb.first = *tb.i
	for ; *tb.i < len(tb.tokens); *tb.i++ {
		imret := tb.step()
		if tb.ok {
			break
		} else if imret {
			return tb.ok
		}
	}
	tb.t.Kind = tb.kind
	return tb.ok
}
