package parser

import (
	"strconv"
	"strings"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/cmd/julec/gen"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

type value struct {
	data      ast.Data
	model     ast.ExprModel
	expr      any
	constant  bool
	lvalue    bool
	variadic  bool
	is_type   bool
	mutable   bool
	cast_type *Type
}

type eval struct {
	p            *Parser
	has_error    bool
	type_prefix  *Type
	allow_unsafe bool
}

func (e *eval) push_err_tok(tok lex.Token, err string, args ...any) {
	if e.has_error {
		return
	}
	e.has_error = true
	e.p.pusherrtok(tok, err, args...)
}

func (e *eval) eval_toks(toks []lex.Token) (value, ast.ExprModel) {
	resolver := builder{}
	return e.eval_expr(resolver.Expr(toks))
}

func (e *eval) eval_expr(expr Expr) (value, ast.ExprModel) { return e.eval(expr.Op) }

func get_bop_model(v value, bop ast.Binop, lm ast.ExprModel, rm ast.ExprModel) ast.ExprModel {
	if v.constant {
		return v.model
	}
	model := exprNode{}
	if bop.Op.Kind == lex.KND_SOLIDUS {
		model.value += "__julec_div("
		model.value += lm.String()
		model.value += ","
		model.value += rm.String()
	} else {
		model.value += lex.KND_LPAREN
		model.value += lm.String()
		model.value += " " + bop.Op.Kind + " "
		model.value += rm.String()
	}
	model.value += lex.KND_RPARENT
	return model
}

func (e *eval) eval_op(op any) (v value, model ast.ExprModel) {
	switch t := op.(type) {
	case ast.BinopExpr:
		m := new_expr_model(1)
		model = m
		v = e.process(t.Tokens, m)
		if v.constant {
			model = v.model
		} else if v.is_type {
			e.push_err_tok(v.data.Token, "invalid_expr")
		}
		return
	case ast.Binop:
	default:
		return
	}

	bop := op.(ast.Binop)
	l, lm := e.eval_op(bop.L)
	if e.has_error {
		return
	}

	r, rm := e.eval_op(bop.R)
	if e.has_error {
		return
	}

	process := solver{
		p:  e.p,
		op: bop.Op,
		l:  l,
		r:  r,
	}
	v = process.solve()
	v.lvalue = types.IsLvalue(v.data.DataType)
	model = get_bop_model(v, bop, lm, rm)
	return
}

func is_invalid_prefix_type(t *Type) bool {
	switch {
	case t.Id == types.ID:
		return true
	case types.IsSlice(*t):
		return is_invalid_prefix_type(t.ComponentType)
	case types.IsMap(*t):
		types := t.Tag.([]Type)
		return is_invalid_prefix_type(&types[0]) || is_invalid_prefix_type(&types[1])
	}
	return false
}

func add_casting_to_model(v value, m ast.ExprModel) ast.ExprModel {
	if v.cast_type == nil {
		return exprNode{m.String()}
	}
	return get_cast_expr_model(*v.cast_type, v.data.DataType, m)
}

func (e *eval) eval(op any) (v value, model ast.ExprModel) {
	if e.type_prefix != nil && is_invalid_prefix_type(e.type_prefix) {
		e.type_prefix = nil
	}

	defer func() {
		if types.IsVoid(v.data.DataType) {
			v.data.DataType.Id = types.VOID
			v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
		} else if v.constant && types.IsPure(v.data.DataType) && lex.IsLiteral(v.data.Value) && !lex.IsRune(v.data.Value) {
			switch v.expr.(type) {
			case int64:
				dt := Type{
					Id:   types.INT,
					Kind: types.TYPE_MAP[types.INT],
				}
				if int_assignable(dt.Id, v) {
					v.data.DataType = dt
				}
			case uint64:
				dt := Type{
					Id:   types.UINT,
					Kind: types.TYPE_MAP[types.UINT],
				}
				if int_assignable(dt.Id, v) {
					v.data.DataType = dt
				}
			}
		}
		if v.cast_type != nil {
			model = add_casting_to_model(v, model)
		}
	}()

	if op == nil || e.has_error {
		return
	}

	return e.eval_op(op)
}

func (e *eval) single(tok lex.Token, m *expr_model) (v value, ok bool) {
	eval := literal_eval{tok, m, e.p}
	v.data.DataType.Id = types.VOID
	v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	v.data.Token = tok
	switch tok.Id {
	case lex.ID_LITERAL:
		ok = true
		switch {
		case lex.IsStr(tok.Kind):
			v = eval.str()
		case lex.IsRune(tok.Kind):
			v = eval.rune()
		case lex.IsBool(tok.Kind):
			v = eval.bool()
		case lex.IsNil(tok.Kind):
			v = eval.nil()
		default:
			v = eval.numeric()
		}
	case lex.ID_IDENT, lex.ID_SELF:
		v, ok = eval.id()
	default:
		e.push_err_tok(tok, "invalid_syntax")
	}
	return
}

func (e *eval) unary(toks []lex.Token, m *expr_model) value {
	var v value
	// Length is 1 cause all length of operator tokens is 1.
	// Change "1" with length of token's value
	// if all operators length is not 1.
	exprToks := toks[1:]
	processor := unary{toks[0], exprToks, m, e.p}
	if processor.toks == nil {
		e.push_err_tok(processor.token, "invalid_syntax")
		return v
	}

	// NOTICE: The first expression model must be unary operator!

	switch processor.token.Kind {
	case lex.KND_MINUS:
		m.append_sub(exprNode{processor.token.Kind})
		v = processor.minus()
	case lex.KND_PLUS:
		m.append_sub(exprNode{processor.token.Kind})
		v = processor.plus()
	case lex.KND_CARET:
		m.append_sub(exprNode{"~"})
		v = processor.caret()
	case lex.KND_EXCL:
		m.append_sub(exprNode{processor.token.Kind})
		v = processor.logical_not()
	case lex.KND_STAR:
		m.append_sub(exprNode{processor.token.Kind})
		v = processor.star()
	case lex.KND_AMPER:
		m.append_sub(exprNode{processor.token.Kind})
		v = processor.amper()
	default:
		e.push_err_tok(processor.token, "invalid_syntax")
	}
	v.data.Token = processor.token
	model := add_casting_to_model(v, m)
	m.nodes[m.index].nodes = nil
	m.append_sub(model)
	return v
}

func (e *eval) between_parentheses(toks []lex.Token, m *expr_model) value {
	m.append_sub(exprNode{lex.KND_LPAREN})
	tk := toks[0]
	toks = toks[1 : len(toks)-1]
	if len(toks) == 0 {
		e.push_err_tok(tk, "invalid_syntax")
	}
	val, model := e.eval_toks(toks)
	m.append_sub(model)
	m.append_sub(exprNode{lex.KND_RPARENT})
	return val
}

func (e *eval) data_type_fn(expr lex.Token, callRange []lex.Token, m *expr_model) (v value, isret bool) {
	switch expr.Id {
	case lex.ID_DT:
		switch expr.Kind {
		case lex.KND_STR:
			m.append_sub(exprNode{"__julec_to_str("})
			_, vm := e.p.evalToks(callRange, nil)
			m.append_sub(vm)
			m.append_sub(exprNode{lex.KND_RPARENT})
			v.data.DataType = Type{
				Id:   types.STR,
				Kind: types.TYPE_MAP[types.STR],
			}
			isret = true
		default:
			dt := Type{
				Token: expr,
				Id:    types.TypeFromId(expr.Kind),
				Kind:  expr.Kind,
			}
			isret = true
			v = e.cast_expr(dt, callRange, m, expr)
		}
	case lex.ID_IDENT:
		def, _, _ := e.p.defined_by_id(expr.Kind)
		if def == nil {
			break
		}
		switch t := def.(type) {
		case *TypeAlias:
			dt, ok := e.p.realType(t.TargetType, true)
			if !ok || types.IsStruct(dt) {
				return
			}
			isret = true
			v = e.cast_expr(dt, callRange, m, expr)
		}
	}
	return
}

type call_data struct {
	expr     []lex.Token
	args     []lex.Token
	generics []lex.Token
}

func get_call_data(toks []lex.Token, m *expr_model) (data call_data) {
	data.expr, data.args = ast.RangeLast(toks)
	if len(data.expr) == 0 {
		return
	}
	// Below is call expression
	tok := data.expr[len(data.expr)-1]
	if tok.Id == lex.ID_BRACE && tok.Kind == lex.KND_RBRACKET {
		data.expr, data.generics = ast.RangeLast(data.expr)
	}
	return
}

func (e *eval) unsafe_allowed() bool {
	return e.allow_unsafe || e.p.unsafe_allowed()
}

func (e *eval) call_fn(f *Fn, data call_data, m *expr_model) value {
	if !e.unsafe_allowed() && f.IsUnsafe {
		e.push_err_tok(data.expr[0], "unsafe_behavior_at_out_of_unsafe_scope")
	}
	if f.BuiltinCaller != nil {
		return f.BuiltinCaller.(builtin_caller)(e.p, f, data, m)
	}
	return e.p.call_fn(f, data, m)
}

func (e *eval) parentheses_range(toks []lex.Token, m *expr_model) (v value) {
	tok := toks[0]
	switch tok.Id {
	case lex.ID_BRACE:
		switch tok.Kind {
		case lex.KND_LPAREN:
			val, ok := e.try_cast(toks, m)
			if ok {
				v = val
				return
			}
		}
	}
	data := get_call_data(toks, m)
	if len(data.expr) == 0 {
		return e.between_parentheses(data.args, m)
	}
	switch tok := data.expr[0]; tok.Id {
	case lex.ID_DT, lex.ID_IDENT:
		if len(data.expr) == 1 && len(data.generics) == 0 {
			v, isret := e.data_type_fn(data.expr[0], data.args, m)
			if isret {
				return v
			}
		}
		fallthrough
	default:
		v = e.process(data.expr, m)
	}
	switch {
	case types.IsFn(v.data.DataType):
		f := v.data.DataType.Tag.(*Fn)
		if f.Receiver != nil && f.Receiver.Mutable && !v.mutable {
			e.push_err_tok(data.expr[len(data.expr)-1], "mutable_operation_on_immutable")
		}
		return e.call_fn(f, data, m)
	}
	e.push_err_tok(data.expr[len(data.expr)-1], "invalid_syntax")
	return
}

func (e *eval) try_cpp_linked_var(toks []lex.Token, m *expr_model) (v value, ok bool) {
	if toks[0].Id != lex.ID_CPP {
		return
	} else if toks[1].Id != lex.ID_DOT {
		e.push_err_tok(toks[1], "invalid_syntax")
		return
	}
	tok := toks[2]
	if tok.Id != lex.ID_IDENT {
		e.push_err_tok(toks[2], "invalid_syntax")
		return
	}
	def, def_t := e.p.linkById(tok.Kind)
	if def_t == ' ' {
		e.push_err_tok(tok, "id_not_exist", tok.Kind)
		return
	}
	m.append_sub(exprNode{tok.Kind})
	ok = true
	switch def_t {
	case 'f':
		v = make_value_from_fn(def.(*ast.Fn))
	case 'v':
		v = make_value_from_var(def.(*ast.Var))
	case 's':
		v = make_value_from_struct(def.(*ast.Struct))
		// Cpp linkage not supports type aliases in expressions
	}
	return
}

func (e *eval) process(toks []lex.Token, m *expr_model) (v value) {
	v.constant = true
	if len(toks) == 1 {
		v, _ = e.single(toks[0], m)
		return
	} else if len(toks) == 3 {
		ok := false
		v, ok = e.try_cpp_linked_var(toks, m)
		if ok {
			return v
		}
	}
	tok := toks[0]
	switch tok.Id {
	case lex.ID_OP:
		return e.unary(toks, m)
	}
	tok = toks[len(toks)-1]
	switch tok.Id {
	case lex.ID_IDENT:
		return e.id(toks, m)
	case lex.ID_OP:
		return e.operator_right(toks, m)
	case lex.ID_BRACE:
		switch tok.Kind {
		case lex.KND_RPARENT:
			return e.parentheses_range(toks, m)
		case lex.KND_RBRACE:
			return e.brace_range(toks, m)
		case lex.KND_RBRACKET:
			return e.bracket_range(toks, m)
		}
	}
	e.push_err_tok(toks[0], "invalid_syntax")
	return
}

func (e *eval) sub_id(toks []lex.Token, m *expr_model) (v value) {
	i := len(toks) - 1
	idTok := toks[i]
	i--
	dotTok := toks[i]
	toks = toks[:i]
	switch len(toks) {
	case 0:
		e.push_err_tok(dotTok, "invalid_syntax")
		return
	case 1:
		tok := toks[0]
		if tok.Id == lex.ID_DT {
			return e.type_sub_id(tok, idTok, m)
		} else if tok.Id == lex.ID_IDENT {
			t, _, _ := e.p.type_by_id(tok.Kind)
			if t != nil && !e.p.is_shadowed(tok.Kind) {
				return e.type_sub_id(t.TargetType.Token, idTok, m)
			}
		}
	}
	val := e.process(toks, m)
	checkType := val.data.DataType
	if types.IsExplicitPtr(checkType) {
		if toks[0].Id != lex.ID_SELF && !e.unsafe_allowed() {
			e.push_err_tok(idTok, "unsafe_behavior_at_out_of_unsafe_scope")
		}
		checkType = types.Elem(checkType)
	} else if types.IsRef(checkType) {
		checkType = types.Elem(checkType)
	}
	switch {
	case types.IsPure(checkType):
		switch {
		case checkType.Id == types.STR:
			return e.str_obj_sub_id(val, idTok, m)
		case is_enum_type(val):
			return e.enum_sub_id(val, idTok, m)
		case is_struct_ins(val):
			return e.struct_obj_sub_id(val, idTok, m)
		case is_trait_ins(val):
			return e.trait_obj_sub_id(val, idTok, m)
		}
	case types.IsSlice(checkType):
		return e.slice_obj_sub_id(val, idTok, m)
	case types.IsArray(checkType):
		return e.array_obj_sub_id(val, idTok, m)
	case types.IsMap(checkType):
		return e.map_obj_sub_id(val, idTok, m)
	}
	e.push_err_tok(dotTok, "obj_not_support_sub_fields", val.data.DataType.Kind)
	return
}

func get_cast_expr_model(t, vt Type, expr ast.ExprModel) ast.ExprModel {
	var model strings.Builder
	switch {
	case types.IsPtr(vt) || types.IsPtr(t):
		model.WriteString("((")
		model.WriteString(t.String())
		model.WriteString(")(")
		model.WriteString(expr.String())
		model.WriteString("))")
		goto end
	case types.IsPure(vt):
		switch {
		case types.IsTrait(vt), vt.Id == types.ANY:
			model.WriteString(expr.String())
			model.WriteString(types.GetAccessor(vt))
			model.WriteString("operator ")
			model.WriteString(t.String())
			model.WriteString("()")
			goto end
		}
	}
	model.WriteString("static_cast<")
	model.WriteString(t.String())
	model.WriteString(">(")
	model.WriteString(expr.String())
	model.WriteByte(')')
end:
	return exprNode{model.String()}
}

func (e *eval) cast_expr(dt Type, exprToks []lex.Token, m *expr_model, errTok lex.Token) value {
	val, model := e.eval_toks(exprToks)
	m.append_sub(get_cast_expr_model(dt, val.data.DataType, model))
	val = e.cast(val, dt, errTok)
	if types.IsPure(val.data.DataType) && types.IsNumeric(val.data.DataType.Id) {
		val.cast_type = new(Type)
		*val.cast_type = dt.Copy()
		val.cast_type.Original = nil
		val.cast_type.Pure = true
	}
	return val
}

func (e *eval) try_cast(toks []lex.Token, m *expr_model) (v value, _ bool) {
	brace_n := 0
	errTok := toks[0]
	for i, tok := range toks {
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
				brace_n++
				continue
			default:
				brace_n--
			}
		}
		if brace_n > 0 {
			continue
		} else if i+1 == len(toks) {
			return
		}
		r := new_builder(nil)
		dtindex := 0
		typeToks := toks[1:i]
		dt, ok := r.DataType(typeToks, &dtindex, false)
		if !ok {
			return
		}
		dt, ok = e.p.realType(dt, false)
		if !ok {
			return
		}
		if dtindex+1 < len(typeToks) {
			return
		}
		exprToks := toks[i+1:]
		if len(exprToks) == 0 {
			return
		}
		tok = exprToks[0]
		if tok.Id != lex.ID_BRACE || tok.Kind != lex.KND_LPAREN {
			return
		}
		exprToks, ok = e.p.get_range(lex.KND_LPAREN, lex.KND_RPARENT, exprToks)
		if !ok {
			return
		}
		val := e.cast_expr(dt, exprToks, m, errTok)
		return val, true
	}
	return
}

func (e *eval) cast(v value, t Type, errtok lex.Token) value {
	switch {
	case types.IsPtr(t):
		e.cast_ptr(t, &v, errtok)
	case types.IsRef(t):
		e.cast_ref(t, &v, errtok)
	case types.IsSlice(t):
		e.cast_slice(t, v.data.DataType, errtok)
	case types.IsStruct(t):
		e.cast_struct(t, &v, errtok)
	case types.IsPure(t):
		if v.data.DataType.Id == types.ANY {
			// The any type supports casting to any data type.
			break
		}
		e.cast_pure(t, &v, errtok)
	default:
		e.push_err_tok(errtok, "type_not_supports_casting", t.Kind)
	}
	v.data.Value = t.Kind
	v.data.DataType = t
	v.lvalue = types.IsLvalue(t)
	v.mutable = types.IsRef(t) || types.IsMut(t)
	if v.constant {
		var model exprNode
		model.value = v.data.DataType.String()
		model.value += lex.KND_LPAREN
		model.value += v.model.String()
		model.value += lex.KND_RPARENT
		v.model = model
	}
	return v
}

func (e *eval) cast_struct(t Type, v *value, errtok lex.Token) {
	if !types.IsTrait(v.data.DataType) {
		e.push_err_tok(errtok, "type_not_supports_casting_to", v.data.DataType.Kind, t.Kind)
		return
	}
	s := t.Tag.(*ast.Struct)
	tr := v.data.DataType.Tag.(*ast.Trait)
	if !s.HasTrait(tr) {
		e.push_err_tok(errtok, "type_not_supports_casting_to", v.data.DataType.Kind, t.Kind)
	}
}

func (e *eval) cast_ptr(t Type, v *value, errtok lex.Token) {
	if !e.unsafe_allowed() {
		e.push_err_tok(errtok, "unsafe_behavior_at_out_of_unsafe_scope")
		return
	}
	if !types.IsPtr(v.data.DataType) &&
		!types.IsPure(v.data.DataType) &&
		!types.IsInteger(v.data.DataType.Id) {
		e.push_err_tok(errtok, "type_not_supports_casting_to", v.data.DataType.Kind, t.Kind)
	}
	v.constant = false
}

func (e *eval) cast_ref(t Type, v *value, errtok lex.Token) {
	if types.IsStruct(types.Elem(t)) {
		e.cast_struct(t, v, errtok)
		return
	}
	e.push_err_tok(errtok, "type_not_supports_casting_to", v.data.DataType.Kind, t.Kind)
}

func (e *eval) cast_pure(t Type, v *value, errtok lex.Token) {
	switch t.Id {
	case types.ANY:
		return
	case types.STR:
		e.cast_str(v.data.DataType, errtok)
		return
	}
	switch {
	case types.IsInteger(t.Id):
		e.cast_int(t, v, errtok)
	case types.IsNumeric(t.Id):
		e.cast_num(t, v, errtok)
	default:
		e.push_err_tok(errtok, "type_not_supports_casting", t.Kind)
	}
}

func (e *eval) cast_str(t Type, errtok lex.Token) {
	if types.IsPure(t) {
		if t.Id != types.U8 && t.Id != types.I32 {
			e.push_err_tok(errtok, "type_not_supports_casting_to", types.TYPE_MAP[types.STR], t.Kind)
		}
		return
	}

	if !types.IsSlice(t) {
		e.push_err_tok(errtok, "type_not_supports_casting_to", types.TYPE_MAP[types.STR], t.Kind)
		return
	}
	t = *t.ComponentType
	if !types.IsPure(t) || (t.Id != types.U8 && t.Id != types.I32) {
		e.push_err_tok(errtok, "type_not_supports_casting_to", types.TYPE_MAP[types.STR], t.Kind)
	}
}

func (e *eval) cast_int(t Type, v *value, errtok lex.Token) {
	if v.constant {
		switch {
		case types.IsSignedInteger(t.Id):
			v.expr = to_num_signed(v)
		default:
			v.expr = to_num_unsigned(v)
		}
	}
	if types.IsEnum(v.data.DataType) {
		e := v.data.DataType.Tag.(*Enum)
		if types.IsNumeric(e.DataType.Id) {
			return
		}
	}
	if types.IsPtr(v.data.DataType) {
		if t.Id == types.UINTPTR {
			return
		} else if !e.unsafe_allowed() {
			e.push_err_tok(errtok, "unsafe_behavior_at_out_of_unsafe_scope")
			return
		} else if t.Id != types.I32 && t.Id != types.I64 &&
			t.Id != types.U16 && t.Id != types.U32 && t.Id != types.U64 {
			e.push_err_tok(errtok, "type_not_supports_casting_to", v.data.DataType.Kind, t.Kind)
		}
		return
	}
	if types.IsPure(v.data.DataType) && types.IsNumeric(v.data.DataType.Id) {
		return
	}
	e.push_err_tok(errtok, "type_not_supports_casting_to", v.data.DataType.Kind, t.Kind)
}

func (e *eval) cast_num(t Type, v *value, errtok lex.Token) {
	if v.constant {
		switch {
		case types.IsFloat(t.Id):
			v.expr = to_num_float(v)
		case types.IsSignedInteger(t.Id):
			v.expr = to_num_signed(v)
		default:
			v.expr = to_num_unsigned(v)
		}
	}
	if types.IsEnum(v.data.DataType) {
		e := v.data.DataType.Tag.(*Enum)
		if types.IsNumeric(e.DataType.Id) {
			return
		}
	}
	if types.IsPure(v.data.DataType) && types.IsNumeric(v.data.DataType.Id) {
		return
	}
	e.push_err_tok(errtok, "type_not_supports_casting_to", v.data.DataType.Kind, t.Kind)
}

func (e *eval) cast_slice(t Type, vt Type, errtok lex.Token) {
	if !types.IsPure(vt) || vt.Id != types.STR {
		e.push_err_tok(errtok, "type_not_supports_casting_to", vt.Kind, t.Kind)
		return
	}
	t = *t.ComponentType
	if !types.IsPure(t) || (t.Id != types.U8 && t.Id != types.I32) {
		e.push_err_tok(errtok, "type_not_supports_casting_to", vt.Kind, t.Kind)
	}
}

func (e *eval) jt_sub_id(dm *ast.Defmap, id_tok lex.Token, m *expr_model) (v value) {
	i, dm, t := dm.FindById(id_tok.Kind, nil)
	if i == -1 {
		e.push_err_tok(id_tok, "obj_have_not_id", id_tok.Kind)
		return
	}
	v.lvalue = false
	v.data.Value = id_tok.Kind
	switch t {
	case 'g':
		g := dm.Globals[i]
		v.data.DataType = g.DataType
		v.constant = g.Constant
		if v.constant {
			v.expr = g.ExprTag
			v.model = g.Expr.Model
			m.append_sub(v.model)
		} else {
			m.append_sub(exprNode{g.Tag.(string)})
		}
	}
	return
}

func (e *eval) i8_sub_id(id_tok lex.Token, m *expr_model) value {
	return e.jt_sub_id(i8statics, id_tok, m)
}
func (e *eval) i16_sub_id(id_tok lex.Token, m *expr_model) value {
	return e.jt_sub_id(i16statics, id_tok, m)
}
func (e *eval) i32_sub_id(id_tok lex.Token, m *expr_model) value {
	return e.jt_sub_id(i32statics, id_tok, m)
}
func (e *eval) i64_sub_id(id_tok lex.Token, m *expr_model) value {
	return e.jt_sub_id(i64statics, id_tok, m)
}
func (e *eval) u8_sub_id(id_tok lex.Token, m *expr_model) value {
	return e.jt_sub_id(u8statics, id_tok, m)
}
func (e *eval) u16_sub_id(id_tok lex.Token, m *expr_model) value {
	return e.jt_sub_id(u16statics, id_tok, m)
}
func (e *eval) u32_sub_id(id_tok lex.Token, m *expr_model) value {
	return e.jt_sub_id(u32statics, id_tok, m)
}
func (e *eval) u64_sub_id(id_tok lex.Token, m *expr_model) value {
	return e.jt_sub_id(u64statics, id_tok, m)
}
func (e *eval) uint_sub_id(id_tok lex.Token, m *expr_model) value {
	return e.jt_sub_id(uintStatics, id_tok, m)
}
func (e *eval) int_sub_id(id_tok lex.Token, m *expr_model) value {
	return e.jt_sub_id(intStatics, id_tok, m)
}
func (e *eval) f32_sub_id(id_tok lex.Token, m *expr_model) value {
	return e.jt_sub_id(f32statics, id_tok, m)
}
func (e *eval) f64_sub_id(id_tok lex.Token, m *expr_model) value {
	return e.jt_sub_id(f64statics, id_tok, m)
}

func (e *eval) type_sub_id(type_tok, id_tok lex.Token, m *expr_model) (v value) {
	switch type_tok.Kind {
	case lex.KND_I8:
		return e.i8_sub_id(id_tok, m)
	case lex.KND_I16:
		return e.i16_sub_id(id_tok, m)
	case lex.KND_I32:
		return e.i32_sub_id(id_tok, m)
	case lex.KND_I64:
		return e.i64_sub_id(id_tok, m)
	case lex.KND_U8:
		return e.u8_sub_id(id_tok, m)
	case lex.KND_U16:
		return e.u16_sub_id(id_tok, m)
	case lex.KND_U32:
		return e.u32_sub_id(id_tok, m)
	case lex.KND_U64:
		return e.u64_sub_id(id_tok, m)
	case lex.KND_UINT:
		return e.uint_sub_id(id_tok, m)
	case lex.KND_INT:
		return e.int_sub_id(id_tok, m)
	case lex.KND_F32:
		return e.f32_sub_id(id_tok, m)
	case lex.KND_F64:
		return e.f64_sub_id(id_tok, m)
	}
	e.push_err_tok(id_tok, "obj_not_support_sub_fields", type_tok.Kind)
	return
}

func (e *eval) type_id(toks []lex.Token, m *expr_model) (v value) {
	v.data.DataType.Id = types.VOID
	v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	r := new_builder(nil)
	i := 0
	t, ok := r.DataType(toks, &i, true)
	if !ok {
		e.p.pusherrs(r.Errors...)
		return
	} else if i+1 >= len(toks) {
		e.push_err_tok(toks[0], "invalid_syntax")
		return
	}
	t, ok = e.p.realType(t, true)
	if !ok {
		return
	}
	toks = toks[i+1:]
	if types.IsPure(t) && types.IsStruct(t) {
		if toks[0].Id != lex.ID_BRACE || toks[0].Kind != lex.KND_LBRACE {
			e.push_err_tok(toks[0], "invalid_syntax")
			return
		}
		s := t.Tag.(*ast.Struct)
		return e.p.callStructConstructor(s, toks, m)
	}
	if toks[0].Id != lex.ID_BRACE || toks[0].Kind != lex.KND_LBRACKET {
		e.push_err_tok(toks[0], "invalid_syntax")
		return
	}
	return e.enumerable(toks, t, m)
}

func (e *eval) obj_sub_id(dm *ast.Defmap, val value, interior_mut bool, id_tok lex.Token, m *expr_model) (v value) {
	i, dm, t := dm.FindById(id_tok.Kind, id_tok.File)
	if i == -1 {
		e.push_err_tok(id_tok, "obj_have_not_id", id_tok.Kind)
		return
	}
	v = val
	m.append_sub(exprNode{types.GetAccessor(val.data.DataType)})
	switch t {
	case 'g':
		g := dm.Globals[i]
		g.Used = true
		v.data.DataType = g.DataType
		v.lvalue = val.lvalue || types.IsLvalue(g.DataType)
		v.mutable = v.mutable || (g.Mutable && interior_mut)
		v.constant = g.Constant
		if g.Constant {
			v.expr = g.ExprTag
			v.model = g.Expr.Model
		}
		if g.Tag != nil {
			m.append_sub(exprNode{g.Tag.(string)})
		} else {
			m.append_sub(exprNode{g.OutId()})
		}
	case 'f':
		f := dm.Fns[i]
		f.Used = true
		v.data.DataType.Id = types.FN
		v.data.DataType.Tag = f
		v.data.DataType.Kind = f.TypeKind()
		v.data.Token = f.Token
		m.append_sub(exprNode{f.OutId()})
	}
	return
}

func (e *eval) str_obj_sub_id(val value, idTok lex.Token, m *expr_model) value {
	readyStrDefines(val)
	v := e.obj_sub_id(strDefines, val, false, idTok, m)
	v.lvalue = false
	return v
}

func (e *eval) slice_obj_sub_id(val value, idTok lex.Token, m *expr_model) value {
	v := e.obj_sub_id(sliceDefines, val, false, idTok, m)
	v.lvalue = false
	return v
}

func (e *eval) array_obj_sub_id(val value, idTok lex.Token, m *expr_model) value {
	readyArrayDefines(val)
	v := e.obj_sub_id(arrayDefines, val, false, idTok, m)
	v.lvalue = false
	return v
}

func (e *eval) map_obj_sub_id(val value, idTok lex.Token, m *expr_model) value {
	readyMapDefines(val.data.DataType)
	v := e.obj_sub_id(mapDefines, val, false, idTok, m)
	v.lvalue = false
	return v
}

func (e *eval) enum_sub_id(val value, idTok lex.Token, m *expr_model) (v value) {
	enum := val.data.DataType.Tag.(*Enum)
	v = val
	v.lvalue = false
	v.is_type = false
	item := enum.ItemById(idTok.Kind)
	if item == nil {
		e.push_err_tok(idTok, "obj_have_not_id", idTok.Kind)
	} else {
		v.expr = item.ExprTag
		v.model = get_const_expr_model(v)
	}
	nodes := m.nodes[m.index]
	nodes.nodes[len(nodes.nodes)-1] = v.model
	return
}

func (e *eval) struct_obj_sub_id(val value, idTok lex.Token, m *expr_model) value {
	parent_type := val.data.DataType
	s := val.data.DataType.Tag.(*ast.Struct)
	val.constant = false
	val.is_type = false
	if val.data.Value == lex.KND_SELF {
		nodes := &m.nodes[m.index].nodes
		n := len(*nodes)
		defer func() {
			// Save unary
			if ast.IsUnaryOp((*nodes)[0].String()) {
				*nodes = append([]ast.ExprModel{(*nodes)[0], exprNode{"this->"}}, (*nodes)[n+1:]...)
				return
			}
			*nodes = append([]ast.ExprModel{exprNode{"this->"}}, (*nodes)[n+1:]...)
		}()

		// Remove "self" from value for correct algorithm.
		val.data.Value = ""
	}
	interior_mutability := false
	if e.p.rootBlock.Func.Receiver != nil {
		switch t := e.p.rootBlock.Func.Receiver.Tag.(type) {
		case *ast.Struct:
			interior_mutability = t.IsSameBase(s)
		}
	}
	val = e.obj_sub_id(s.Defines, val, interior_mutability, idTok, m)
	if types.IsFn(val.data.DataType) {
		f := val.data.DataType.Tag.(*Fn)
		if f.Receiver != nil && types.IsRef(f.Receiver.DataType) && !types.IsRef(parent_type) {
			e.p.pusherrtok(idTok, "ref_method_used_with_not_ref_instance")
		}
	}
	return val
}

func (e *eval) trait_obj_sub_id(val value, idTok lex.Token, m *expr_model) value {
	m.append_sub(exprNode{"._get()"})
	t := val.data.DataType.Tag.(*ast.Trait)
	val.constant = false
	val.lvalue = false
	val.is_type = false
	val = e.obj_sub_id(t.Defines, val, false, idTok, m)
	val.constant = false
	return val
}

type ns_find interface {
	NsById(string) *ast.Namespace
}

func (e *eval) get_ns(toks *[]lex.Token) *ast.Defmap {
	var prev ns_find = e.p
	var ns *ast.Namespace
	for i, tok := range *toks {
		if (i+1)%2 != 0 {
			if tok.Id != lex.ID_IDENT {
				e.push_err_tok(tok, "invalid_syntax")
				continue
			}
			src := prev.NsById(tok.Kind)
			if src == nil {
				if ns != nil {
					*toks = (*toks)[i:]
					return ns.Defines
				}
				e.push_err_tok(tok, "namespace_not_exist", tok.Kind)
				return nil
			}
			prev = src.Defines
			ns = src
			continue
		}
		if tok.Id != lex.ID_DBLCOLON {
			return ns.Defines
		}
	}
	return ns.Defines
}

func (e *eval) ns_sub_id(toks []lex.Token, m *expr_model) (v value) {
	defs := e.get_ns(&toks)
	if defs == nil {
		return
	}
	// Temporary clean of local defines
	// Because this defines has high priority
	// So shadows defines of namespace
	blockTypes := e.p.blockTypes
	blockVars := e.p.block_vars
	package_files := e.p.package_files
	e.p.blockTypes = nil
	e.p.block_vars = nil
	e.p.setup_package() // Create new package
	pdefs := e.p.Defines
	e.p.Defines = defs
	e.p.allowBuiltin = false
	v, _ = e.single(toks[0], m)
	e.p.allowBuiltin = true
	e.p.package_files = package_files
	e.p.blockTypes = blockTypes
	e.p.block_vars = blockVars
	e.p.Defines = pdefs
	return
}

func (e *eval) id(toks []lex.Token, m *expr_model) (v value) {
	i := len(toks) - 1
	tok := toks[i]
	if i <= 0 {
		v, _ = e.single(tok, m)
		return
	}
	i--
	tok = toks[i]
	switch tok.Id {
	case lex.ID_DOT:
		return e.sub_id(toks, m)
	case lex.ID_DBLCOLON:
		return e.ns_sub_id(toks, m)
	}
	e.push_err_tok(toks[i], "invalid_syntax")
	return
}

func (e *eval) operator_right(toks []lex.Token, m *expr_model) (v value) {
	tok := toks[len(toks)-1]
	switch tok.Kind {
	case lex.KND_TRIPLE_DOT:
		toks = toks[:len(toks)-1]
		return e.variadic(toks, m, tok)
	default:
		e.push_err_tok(tok, "invalid_syntax")
	}
	return
}

func ready_to_variadic(v *value) {
	if v.data.DataType.Id != types.STR || !types.IsPure(v.data.DataType) {
		return
	}
	v.data.DataType.Id = types.SLICE
	v.data.DataType.ComponentType = new(Type)
	v.data.DataType.ComponentType.Id = types.U8
	v.data.DataType.ComponentType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	v.data.DataType.Kind = lex.PREFIX_SLICE + v.data.DataType.ComponentType.Kind
}

func (e *eval) variadic(toks []lex.Token, m *expr_model, errtok lex.Token) (v value) {
	v = e.process(toks, m)
	ready_to_variadic(&v)
	if !types.IsVariadicable(v.data.DataType) {
		e.push_err_tok(errtok, "variadic_with_non_variadicable", v.data.DataType.Kind)
		return
	}
	v.data.DataType = *v.data.DataType.ComponentType
	v.variadic = true
	v.constant = false
	return
}

func (e *eval) try_slicing(v *value, toks []lex.Token, m *expr_model, errTok lex.Token) bool {
	i := 0
	toks, colon := ast.SplitColon(toks, &i)
	if colon == -1 {
		return false
	}
	i = 0
	var leftv, rightv value
	leftv.constant = true
	rightv.constant = true
	leftToks := toks[:colon]
	rightToks := toks[colon+1:]
	m.append_sub(exprNode{".___slice("})
	if len(leftToks) > 0 {
		var model ast.ExprModel
		leftv, model = e.p.evalToks(leftToks, nil)
		m.append_sub(get_indexing_expr_model(leftv, model))
		e.check_integer_indexing(leftv, errTok)
	} else {
		leftv.expr = int64(0)
		leftv.model = get_num_model(leftv)
		m.append_sub(exprNode{"0"})
	}
	if len(rightToks) > 0 {
		m.append_sub(exprNode{","})
		var model ast.ExprModel
		rightv, model = e.p.evalToks(rightToks, nil)
		m.append_sub(get_indexing_expr_model(rightv, model))
		e.check_integer_indexing(rightv, errTok)
	}
	m.append_sub(exprNode{")"})
	*v = e.slicing(*v, leftv, rightv, errTok)
	if !types.IsMut(v.data.DataType) {
		v.data.Value = " "
	}
	return true
}

func (e *eval) indexing(v *value, toks []lex.Token, m *expr_model, err_tok lex.Token) {
	m.append_sub(exprNode{lex.KND_LBRACKET})
	indexv, model := e.eval_toks(toks[1 : len(toks)-1])
	if types.IsMap(v.data.DataType) {
		m.append_sub(model)
	} else {
		m.append_sub(get_indexing_expr_model(indexv, model))
	}
	m.append_sub(exprNode{lex.KND_RBRACKET})

	*v = e.check_indexing_type(*v, indexv, err_tok)
	if !types.IsMut(v.data.DataType) {
		v.data.Value = " "
	}
	// Ignore indexed type from original
	v.data.DataType.Pure = true
	v.data.DataType.Original = nil
}

func (e *eval) bracket_range(toks []lex.Token, m *expr_model) (v value) {
	errTok := toks[0]
	var exprToks []lex.Token
	brace_n := 0
	for i := len(toks) - 1; i >= 0; i-- {
		tok := toks[i]
		if tok.Id == lex.ID_BRACE {
			switch tok.Kind {
			case lex.KND_RBRACE, lex.KND_RBRACKET, lex.KND_RPARENT:
				brace_n++
			default:
				brace_n--
			}
		}
		if brace_n == 0 {
			exprToks = toks[:i]
			break
		}
	}

	switch {
	case len(exprToks) == 0:
		if e.type_prefix != nil {
			switch {
			case types.IsArray(*e.type_prefix) || types.IsSlice(*e.type_prefix):
				return e.enumerable(toks, *e.type_prefix, m)
			}
		}
		var model ast.ExprModel
		v, model = e.build_slice_implicit(e.enumerable_parts(toks), toks[0])
		m.append_sub(model)
		return v
	case len(exprToks) == 0 || brace_n > 0:
		e.push_err_tok(errTok, "invalid_syntax")
		return
	}

	var model ast.ExprModel
	v, model = e.eval_toks(exprToks)
	m.append_sub(model)
	toks = toks[len(exprToks):] // lex.Tokenens of [...]
	
	if e.try_slicing(&v, toks, m, errTok) {
		return
	}
	e.indexing(&v, toks, m, errTok)

	return
}

func (e *eval) check_integer_indexing(v value, err_tok lex.Token) {
	err_key := check_value_for_indexing(v)
	if err_key != "" {
		e.push_err_tok(err_tok, err_key)
	}
}

func (e *eval) check_indexing_type(enumv value, indexv value, err_tok lex.Token) (v value) {
	switch {
	case types.IsExplicitPtr(enumv.data.DataType):
		return e.indexing_explicit_ptr(enumv, indexv, err_tok)
	case types.IsArray(enumv.data.DataType):
		return e.indexing_array(enumv, indexv, err_tok)
	case types.IsSlice(enumv.data.DataType):
		return e.indexing_slice(enumv, indexv, err_tok)
	case types.IsMap(enumv.data.DataType):
		return e.indexing_map(enumv, indexv, err_tok)
	case types.IsPure(enumv.data.DataType):
		switch enumv.data.DataType.Id {
		case types.STR:
			return e.indexing_str(enumv, indexv, err_tok)
		}
	}
	e.push_err_tok(err_tok, "not_supports_indexing", enumv.data.DataType.Kind)
	return
}

func (e *eval) indexing_slice(slicev, index value, errtok lex.Token) value {
	slicev.data.DataType = *slicev.data.DataType.ComponentType
	e.check_integer_indexing(index, errtok)
	return slicev
}

func (e *eval) indexing_explicit_ptr(ptrv, index value, errtok lex.Token) value {
	if !e.unsafe_allowed() {
		e.push_err_tok(errtok, "unsafe_behavior_at_out_of_unsafe_scope")
	}
	ptrv.data.DataType = types.Elem(ptrv.data.DataType)
	e.check_integer_indexing(index, errtok)
	return ptrv
}

func (e *eval) indexing_array(arrv, index value, errtok lex.Token) value {
	arrv.data.DataType = *arrv.data.DataType.ComponentType
	e.check_integer_indexing(index, errtok)
	return arrv
}

func (e *eval) indexing_map(mapv, leftv value, errtok lex.Token) value {
	types := mapv.data.DataType.Tag.([]Type)
	keyType := types[0]
	valType := types[1]
	mapv.data.DataType = valType
	e.p.check_type(keyType, leftv.data.DataType, false, true, errtok)
	return mapv
}

func (e *eval) indexing_str(strv, index value, errtok lex.Token) value {
	strv.data.DataType.Id = types.U8
	strv.data.DataType.Kind = types.TYPE_MAP[strv.data.DataType.Id]
	e.check_integer_indexing(index, errtok)
	if !index.constant {
		strv.constant = false
		return strv
	}
	if strv.constant {
		i := to_num_signed(index.expr)
		s := strv.expr.(string)
		if int(i) >= len(s) {
			e.p.pusherrtok(errtok, "overflow_limits")
		} else {
			strv.expr = uint64(s[i])
			strv.model = get_num_model(strv)
		}
	}
	return strv
}

func (e *eval) slicing(enumv, leftv, rightv value, errtok lex.Token) (v value) {
	switch {
	case types.IsArray(enumv.data.DataType):
		return e.slicing_array(enumv, errtok)
	case types.IsSlice(enumv.data.DataType):
		return e.slicing_slice(enumv, errtok)
	case types.IsPure(enumv.data.DataType):
		switch enumv.data.DataType.Id {
		case types.STR:
			return e.slicing_str(enumv, leftv, rightv, errtok)
		}
	}
	e.push_err_tok(errtok, "not_supports_slicing", enumv.data.DataType.Kind)
	return
}

func (e *eval) slicing_slice(v value, errtok lex.Token) value {
	v.lvalue = false
	return v
}

func (e *eval) slicing_array(v value, errtok lex.Token) value {
	v.lvalue = false
	v.data.DataType.Id = types.SLICE
	v.data.DataType.Kind = lex.PREFIX_SLICE + v.data.DataType.ComponentType.Kind
	return v
}

func (e *eval) slicing_str(v, leftv, rightv value, errtok lex.Token) value {
	v.lvalue = false
	v.data.DataType.Id = types.STR
	v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	if !v.constant {
		return v
	}
	if rightv.constant {
		right := to_num_signed(rightv.expr)
		s := v.expr.(string)
		if int(right) >= len(s) {
			e.p.pusherrtok(errtok, "overflow_limits")
		}
	}
	if leftv.constant && rightv.constant {
		left := to_num_signed(leftv.expr)
		if left < 0 {
			return v
		}
		s := v.expr.(string)
		var right int64
		if rightv.expr == nil {
			right = int64(len(s))
		} else {
			right = to_num_signed(rightv.expr)
		}
		if left > right {
			return v
		}
		v.expr = s[left:right]
		v.model = get_str_model(v)
	} else {
		v.constant = false
	}
	return v
}

// ! IMPORTANT: lex.Tokenens is should be store enumerable parentheses.
func (e *eval) enumerable_parts(toks []lex.Token) [][]lex.Token {
	toks = toks[1 : len(toks)-1]
	parts, errs := ast.Parts(toks, lex.ID_COMMA, true)
	e.p.pusherrs(errs...)
	return parts
}

func (e *eval) build_array(parts [][]lex.Token, t Type, errtok lex.Token) (value, ast.ExprModel) {
	if !t.Size.AutoSized {
		n := ast.Size(len(parts))
		if n > t.Size.N {
			e.p.pusherrtok(errtok, "overflow_limits")
		}
	} else {
		t.Size.N = ast.Size(len(parts))
		t.Size.Expr = ast.Expr{
			Model: exprNode{
				value: types.CppId(types.UINT) + "(" + strconv.FormatUint(uint64(t.Size.N), 10) + ")",
			},
		}
	}
	old_type := e.type_prefix
	e.type_prefix = t.ComponentType
	var v value
	v.data.Value = t.Kind
	v.data.DataType = t
	model := sliceExpr{dataType: t}
	for _, part := range parts {
		partVal, expModel := e.eval_toks(part)
		model.expr = append(model.expr, expModel)
		assign_checker{
			p:      e.p,
			t:      *t.ComponentType,
			v:      partVal,
			errtok: part[0],
		}.check()
	}
	e.type_prefix = old_type
	return v, model
}

func (e *eval) build_slice_implicit(parts [][]lex.Token, errtok lex.Token) (value, ast.ExprModel) {
	if len(parts) == 0 {
		e.push_err_tok(errtok, "dynamic_type_annotation_failed")
		return value{}, nil
	}
	v := value{}
	model := sliceExpr{}
	partVal, expModel := e.eval_toks(parts[0])
	model.expr = append(model.expr, expModel)
	model.dataType = Type{
		Id:   types.SLICE,
		Kind: lex.PREFIX_SLICE + partVal.data.DataType.Kind,
	}
	model.dataType.ComponentType = new(Type)
	*model.dataType.ComponentType = partVal.data.DataType
	v.data.DataType = model.dataType
	v.data.Value = model.dataType.Kind
	for _, part := range parts[1:] {
		partVal, expModel := e.eval_toks(part)
		model.expr = append(model.expr, expModel)
		assign_checker{
			p:      e.p,
			t:      *model.dataType.ComponentType,
			v:      partVal,
			errtok: part[0],
		}.check()
	}
	return v, model
}

func (e *eval) build_slice_explicit(parts [][]lex.Token, t Type, errtok lex.Token) (value, ast.ExprModel) {
	old_type := e.type_prefix
	e.type_prefix = t.ComponentType
	var v value
	v.data.Value = t.Kind
	v.data.DataType = t
	model := sliceExpr{dataType: t}
	for _, part := range parts {
		partVal, expModel := e.eval_toks(part)
		model.expr = append(model.expr, expModel)
		assign_checker{
			p:      e.p,
			t:      *t.ComponentType,
			v:      partVal,
			errtok: part[0],
		}.check()
	}
	e.type_prefix = old_type
	return v, model
}

func (e *eval) build_map(parts [][]lex.Token, t Type, errtok lex.Token) (value, ast.ExprModel) {
	var v value
	v.data.Value = t.Kind
	v.data.DataType = t
	model := mapExpr{dataType: t}
	types := t.Tag.([]Type)
	keyType := types[0]
	valType := types[1]
	for _, part := range parts {
		brace_n := 0
		colon := -1
		for i, tok := range part {
			if tok.Id == lex.ID_BRACE {
				switch tok.Kind {
				case lex.KND_LBRACE, lex.KND_LBRACKET, lex.KND_LPAREN:
					brace_n++
				default:
					brace_n--
				}
			}
			if brace_n != 0 {
				continue
			}
			if tok.Id == lex.ID_COLON {
				colon = i
				break
			}
		}
		if colon < 1 || colon+1 >= len(part) {
			e.push_err_tok(errtok, "missing_expr")
			continue
		}
		colonTok := part[colon]
		keyToks := part[:colon]
		valToks := part[colon+1:]
		key, keyModel := e.eval_toks(keyToks)
		model.keyExprs = append(model.keyExprs, keyModel)
		val, valModel := e.eval_toks(valToks)
		model.valExprs = append(model.valExprs, valModel)
		assign_checker{
			p:      e.p,
			t:      keyType,
			v:      key,
			errtok: colonTok,
		}.check()
		assign_checker{
			p:      e.p,
			t:      valType,
			v:      val,
			errtok: colonTok,
		}.check()
	}
	return v, model
}

func (e *eval) enumerable(exprToks []lex.Token, t Type, m *expr_model) (v value) {
	var model ast.ExprModel
	t, ok := e.p.realType(t, true)
	if !ok {
		return
	}
	errtok := exprToks[0]
	switch {
	case types.IsArray(t):
		v, model = e.build_array(e.enumerable_parts(exprToks), t, errtok)
	case types.IsSlice(t):
		v, model = e.build_slice_explicit(e.enumerable_parts(exprToks), t, errtok)
	case types.IsMap(t):
		v, model = e.build_map(e.enumerable_parts(exprToks), t, errtok)
	default:
		e.push_err_tok(errtok, "invalid_type_source")
		return
	}
	m.append_sub(model)
	return
}

func (e *eval) anon_fn(toks []lex.Token, m *expr_model) (v value) {
	r := new_builder(toks)
	f := r.Func(r.Tokens, false, true, false)
	if len(r.Errors) > 0 {
		e.p.pusherrs(r.Errors...)
		return
	}
	e.p.check_anon_fn(&f)
	f.Owner = e.p
	v.data.Value = f.Id
	v.data.DataType.Tag = &f
	v.data.DataType.Id = types.FN
	v.data.DataType.Kind = f.TypeKind()
	m.append_sub(gen.AnonFuncExpr{Ast: &f})
	return
}

func (e *eval) unsafe_eval(toks []lex.Token, m *expr_model) (v value) {
	i := 0
	rang := ast.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, toks)
	if len(rang) == 0 {
		e.push_err_tok(toks[0], "missing_expr")
		return
	}
	old := e.allow_unsafe
	e.allow_unsafe = true
	v = e.process(rang, m)
	e.allow_unsafe = old
	return v
}

func (e *eval) brace_range(toks []lex.Token, m *expr_model) (v value) {
	var exprToks []lex.Token
	brace_n := 0
	for i := len(toks) - 1; i >= 0; i-- {
		tok := toks[i]
		if tok.Id != lex.ID_BRACE {
			continue
		}
		switch tok.Kind {
		case lex.KND_RBRACE, lex.KND_RBRACKET, lex.KND_RPARENT:
			brace_n++
		default:
			brace_n--
		}
		if brace_n != 0 {
			continue
		}
		exprToks = toks[:i]
		break
	}
	switch {
	case len(exprToks) == 0:
		if e.type_prefix != nil {
			switch {
			case types.IsMap(*e.type_prefix):
				return e.enumerable(toks, *e.type_prefix, m)
			case types.IsStruct(*e.type_prefix):
				prefix := e.type_prefix
				s := e.type_prefix.Tag.(*ast.Struct)
				v = e.p.callStructConstructor(s, toks, m)
				e.type_prefix = prefix
				return
			}
		}
		fallthrough
	case brace_n > 0:
		e.push_err_tok(toks[0], "invalid_syntax")
		return
	}
	switch exprToks[0].Id {
	case lex.ID_UNSAFE:
		if len(toks) == 0 {
			e.push_err_tok(toks[0], "invalid_syntax")
			return
		} else if toks[1].Id != lex.ID_FN {
			return e.unsafe_eval(toks[1:], m)
		}
		fallthrough
	case lex.ID_FN:
		return e.anon_fn(toks, m)
	case lex.ID_IDENT, lex.ID_CPP:
		return e.type_id(toks, m)
	default:
		e.push_err_tok(exprToks[0], "invalid_syntax")
	}
	return
}
