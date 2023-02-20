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

func (e *eval) pusherrtok(tok lex.Token, err string, args ...any) {
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
	model.value += lex.KND_LPAREN
	model.value += lm.String()
	model.value += " " + bop.Op.Kind + " "
	model.value += rm.String()
	model.value += lex.KND_RPARENT
	return model
}

func (e *eval) eval_op(op any) (v value, model ast.ExprModel) {
	switch t := op.(type) {
	case ast.BinopExpr:
		m := newExprModel(1)
		model = m
		v = e.process(t.Tokens, m)
		if v.constant {
			model = v.model
		} else if v.is_type {
			e.pusherrtok(v.data.Token, "invalid_expr")
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
		} else if v.constant && types.IsPure(v.data.DataType) && lex.IsLiteral(v.data.Value) {
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

func (e *eval) single(tok lex.Token, m *exprModel) (v value, ok bool) {
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
		case lex.IsChar(tok.Kind):
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
		e.pusherrtok(tok, "invalid_syntax")
	}
	return
}

func (e *eval) unary(toks []lex.Token, m *exprModel) value {
	var v value
	// Length is 1 cause all length of operator tokens is 1.
	// Change "1" with length of token's value
	// if all operators length is not 1.
	exprToks := toks[1:]
	processor := unary{toks[0], exprToks, m, e.p}
	if processor.toks == nil {
		e.pusherrtok(processor.token, "invalid_syntax")
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
		v = processor.logicalNot()
	case lex.KND_STAR:
		m.append_sub(exprNode{processor.token.Kind})
		v = processor.star()
	case lex.KND_AMPER:
		m.append_sub(exprNode{processor.token.Kind})
		v = processor.amper()
	default:
		e.pusherrtok(processor.token, "invalid_syntax")
	}
	v.data.Token = processor.token
	model := add_casting_to_model(v, m)
	m.nodes[m.index].nodes = nil
	m.append_sub(model)
	return v
}

func (e *eval) betweenParentheses(toks []lex.Token, m *exprModel) value {
	m.append_sub(exprNode{lex.KND_LPAREN})
	tk := toks[0]
	toks = toks[1 : len(toks)-1]
	if len(toks) == 0 {
		e.pusherrtok(tk, "invalid_syntax")
	}
	val, model := e.eval_toks(toks)
	m.append_sub(model)
	m.append_sub(exprNode{lex.KND_RPARENT})
	return val
}

func (e *eval) dataTypeFunc(expr lex.Token, callRange []lex.Token, m *exprModel) (v value, isret bool) {
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
			v = e.castExpr(dt, callRange, m, expr)
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
			v = e.castExpr(dt, callRange, m, expr)
		}
	}
	return
}

type callData struct {
	expr     []lex.Token
	args     []lex.Token
	generics []lex.Token
}

func getCallData(toks []lex.Token, m *exprModel) (data callData) {
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

func (e *eval) callFunc(f *Fn, data callData, m *exprModel) value {
	if !e.unsafe_allowed() && f.IsUnsafe {
		e.pusherrtok(data.expr[0], "unsafe_behavior_at_out_of_unsafe_scope")
	}
	if f.BuiltinCaller != nil {
		return f.BuiltinCaller.(BuiltinCaller)(e.p, f, data, m)
	}
	return e.p.call_fn(f, data, m)
}

func (e *eval) parenthesesRange(toks []lex.Token, m *exprModel) (v value) {
	tok := toks[0]
	switch tok.Id {
	case lex.ID_BRACE:
		switch tok.Kind {
		case lex.KND_LPAREN:
			val, ok := e.tryCast(toks, m)
			if ok {
				v = val
				return
			}
		}
	}
	data := getCallData(toks, m)
	if len(data.expr) == 0 {
		return e.betweenParentheses(data.args, m)
	}
	switch tok := data.expr[0]; tok.Id {
	case lex.ID_DT, lex.ID_IDENT:
		if len(data.expr) == 1 && len(data.generics) == 0 {
			v, isret := e.dataTypeFunc(data.expr[0], data.args, m)
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
			e.pusherrtok(data.expr[len(data.expr)-1], "mutable_operation_on_immutable")
		}
		return e.callFunc(f, data, m)
	}
	e.pusherrtok(data.expr[len(data.expr)-1], "invalid_syntax")
	return
}

func (e *eval) try_cpp_linked_var(toks []lex.Token, m *exprModel) (v value, ok bool) {
	if toks[0].Id != lex.ID_CPP {
		return
	} else if toks[1].Id != lex.ID_DOT {
		e.pusherrtok(toks[1], "invalid_syntax")
		return
	}
	tok := toks[2]
	if tok.Id != lex.ID_IDENT {
		e.pusherrtok(toks[2], "invalid_syntax")
		return
	}
	def, def_t := e.p.linkById(tok.Kind)
	if def_t == ' ' {
		e.pusherrtok(tok, "id_not_exist", tok.Kind)
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

func (e *eval) process(toks []lex.Token, m *exprModel) (v value) {
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
		return e.operatorRight(toks, m)
	case lex.ID_BRACE:
		switch tok.Kind {
		case lex.KND_RPARENT:
			return e.parenthesesRange(toks, m)
		case lex.KND_RBRACE:
			return e.braceRange(toks, m)
		case lex.KND_RBRACKET:
			return e.bracketRange(toks, m)
		}
	}
	e.pusherrtok(toks[0], "invalid_syntax")
	return
}

func (e *eval) subId(toks []lex.Token, m *exprModel) (v value) {
	i := len(toks) - 1
	idTok := toks[i]
	i--
	dotTok := toks[i]
	toks = toks[:i]
	switch len(toks) {
	case 0:
		e.pusherrtok(dotTok, "invalid_syntax")
		return
	case 1:
		tok := toks[0]
		if tok.Id == lex.ID_DT {
			return e.typeSubId(tok, idTok, m)
		} else if tok.Id == lex.ID_IDENT {
			t, _, _ := e.p.type_by_id(tok.Kind)
			if t != nil && !e.p.is_shadowed(tok.Kind) {
				return e.typeSubId(t.TargetType.Token, idTok, m)
			}
		}
	}
	val := e.process(toks, m)
	checkType := val.data.DataType
	if types.IsExplicitPtr(checkType) {
		if toks[0].Id != lex.ID_SELF && !e.unsafe_allowed() {
			e.pusherrtok(idTok, "unsafe_behavior_at_out_of_unsafe_scope")
		}
		checkType = types.DerefPtrOrRef(checkType)
	} else if types.IsRef(checkType) {
		checkType = types.DerefPtrOrRef(checkType)
	}
	switch {
	case types.IsPure(checkType):
		switch {
		case checkType.Id == types.STR:
			return e.strObjSubId(val, idTok, m)
		case valIsEnumType(val):
			return e.enumSubId(val, idTok, m)
		case valIsStructIns(val):
			return e.structObjSubId(val, idTok, m)
		case valIsTraitIns(val):
			return e.traitObjSubId(val, idTok, m)
		}
	case types.IsSlice(checkType):
		return e.sliceObjSubId(val, idTok, m)
	case types.IsArray(checkType):
		return e.arrayObjSubId(val, idTok, m)
	case types.IsMap(checkType):
		return e.mapObjSubId(val, idTok, m)
	}
	e.pusherrtok(dotTok, "obj_not_support_sub_fields", val.data.DataType.Kind)
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

func (e *eval) castExpr(dt Type, exprToks []lex.Token, m *exprModel, errTok lex.Token) value {
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

func (e *eval) tryCast(toks []lex.Token, m *exprModel) (v value, _ bool) {
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
		val := e.castExpr(dt, exprToks, m, errTok)
		return val, true
	}
	return
}

func (e *eval) cast(v value, t Type, errtok lex.Token) value {
	switch {
	case types.IsPtr(t):
		e.castPtr(t, &v, errtok)
	case types.IsRef(t):
		e.cast_ref(t, &v, errtok)
	case types.IsSlice(t):
		e.castSlice(t, v.data.DataType, errtok)
	case types.IsStruct(t):
		e.castStruct(t, &v, errtok)
	case types.IsPure(t):
		if v.data.DataType.Id == types.ANY {
			// The any type supports casting to any data type.
			break
		}
		e.castPure(t, &v, errtok)
	default:
		e.pusherrtok(errtok, "type_not_supports_casting", t.Kind)
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

func (e *eval) castStruct(t Type, v *value, errtok lex.Token) {
	if !types.IsTrait(v.data.DataType) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.DataType.Kind, t.Kind)
		return
	}
	s := t.Tag.(*ast.Struct)
	tr := v.data.DataType.Tag.(*ast.Trait)
	if !s.HasTrait(tr) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.DataType.Kind, t.Kind)
	}
}

func (e *eval) castPtr(t Type, v *value, errtok lex.Token) {
	if !e.unsafe_allowed() {
		e.pusherrtok(errtok, "unsafe_behavior_at_out_of_unsafe_scope")
		return
	}
	if !types.IsPtr(v.data.DataType) &&
		!types.IsPure(v.data.DataType) &&
		!types.IsInteger(v.data.DataType.Id) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.DataType.Kind, t.Kind)
	}
	v.constant = false
}

func (e *eval) cast_ref(t Type, v *value, errtok lex.Token) {
	if types.IsStruct(types.DerefPtrOrRef(t)) {
		e.castStruct(t, v, errtok)
		return
	}
	e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.DataType.Kind, t.Kind)
}

func (e *eval) castPure(t Type, v *value, errtok lex.Token) {
	switch t.Id {
	case types.ANY:
		return
	case types.STR:
		e.castStr(v.data.DataType, errtok)
		return
	}
	switch {
	case types.IsInteger(t.Id):
		e.castInteger(t, v, errtok)
	case types.IsNumeric(t.Id):
		e.castNumeric(t, v, errtok)
	default:
		e.pusherrtok(errtok, "type_not_supports_casting", t.Kind)
	}
}

func (e *eval) castStr(t Type, errtok lex.Token) {
	if types.IsPure(t) || (t.Id != types.U8 && t.Id != types.I32) {
		return
	}
	if !types.IsSlice(t) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", types.TYPE_MAP[types.STR], t.Kind)
		return
	}
	t = *t.ComponentType
	if !types.IsPure(t) || (t.Id != types.U8 && t.Id != types.I32) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", types.TYPE_MAP[types.STR], t.Kind)
	}
}

func (e *eval) castInteger(t Type, v *value, errtok lex.Token) {
	if v.constant {
		switch {
		case types.IsSignedInteger(t.Id):
			v.expr = tonums(v)
		default:
			v.expr = tonumu(v)
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
			e.pusherrtok(errtok, "unsafe_behavior_at_out_of_unsafe_scope")
			return
		} else if t.Id != types.I32 && t.Id != types.I64 &&
			t.Id != types.U16 && t.Id != types.U32 && t.Id != types.U64 {
			e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.DataType.Kind, t.Kind)
		}
		return
	}
	if types.IsPure(v.data.DataType) && types.IsNumeric(v.data.DataType.Id) {
		return
	}
	e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.DataType.Kind, t.Kind)
}

func (e *eval) castNumeric(t Type, v *value, errtok lex.Token) {
	if v.constant {
		switch {
		case types.IsFloat(t.Id):
			v.expr = tonumf(v)
		case types.IsSignedInteger(t.Id):
			v.expr = tonums(v)
		default:
			v.expr = tonumu(v)
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
	e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.DataType.Kind, t.Kind)
}

func (e *eval) castSlice(t, vt Type, errtok lex.Token) {
	if !types.IsPure(vt) || vt.Id != types.STR {
		e.pusherrtok(errtok, "type_not_supports_casting_to", vt.Kind, t.Kind)
		return
	}
	t = *t.ComponentType
	if !types.IsPure(t) || (t.Id != types.U8 && t.Id != types.I32) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", vt.Kind, t.Kind)
	}
}

func (e *eval) juletypeSubId(dm *ast.Defmap, idTok lex.Token, m *exprModel) (v value) {
	i, dm, t := dm.FindById(idTok.Kind, nil)
	if i == -1 {
		e.pusherrtok(idTok, "obj_have_not_id", idTok.Kind)
		return
	}
	v.lvalue = false
	v.data.Value = idTok.Kind
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

func (e *eval) i8SubId(idTok lex.Token, m *exprModel) value {
	return e.juletypeSubId(i8statics, idTok, m)
}

func (e *eval) i16SubId(idTok lex.Token, m *exprModel) value {
	return e.juletypeSubId(i16statics, idTok, m)
}

func (e *eval) i32SubId(idTok lex.Token, m *exprModel) value {
	return e.juletypeSubId(i32statics, idTok, m)
}

func (e *eval) i64SubId(idTok lex.Token, m *exprModel) value {
	return e.juletypeSubId(i64statics, idTok, m)
}

func (e *eval) u8SubId(idTok lex.Token, m *exprModel) value {
	return e.juletypeSubId(u8statics, idTok, m)
}

func (e *eval) u16SubId(idTok lex.Token, m *exprModel) value {
	return e.juletypeSubId(u16statics, idTok, m)
}

func (e *eval) u32SubId(idTok lex.Token, m *exprModel) value {
	return e.juletypeSubId(u32statics, idTok, m)
}

func (e *eval) u64SubId(idTok lex.Token, m *exprModel) value {
	return e.juletypeSubId(u64statics, idTok, m)
}

func (e *eval) uintSubId(idTok lex.Token, m *exprModel) value {
	return e.juletypeSubId(uintStatics, idTok, m)
}

func (e *eval) intSubId(idTok lex.Token, m *exprModel) value {
	return e.juletypeSubId(intStatics, idTok, m)
}

func (e *eval) f32SubId(idTok lex.Token, m *exprModel) value {
	return e.juletypeSubId(f32statics, idTok, m)
}

func (e *eval) f64SubId(idTok lex.Token, m *exprModel) value {
	return e.juletypeSubId(f64statics, idTok, m)
}

func (e *eval) typeSubId(typeTok, idTok lex.Token, m *exprModel) (v value) {
	switch typeTok.Kind {
	case lex.KND_I8:
		return e.i8SubId(idTok, m)
	case lex.KND_I16:
		return e.i16SubId(idTok, m)
	case lex.KND_I32:
		return e.i32SubId(idTok, m)
	case lex.KND_I64:
		return e.i64SubId(idTok, m)
	case lex.KND_U8:
		return e.u8SubId(idTok, m)
	case lex.KND_U16:
		return e.u16SubId(idTok, m)
	case lex.KND_U32:
		return e.u32SubId(idTok, m)
	case lex.KND_U64:
		return e.u64SubId(idTok, m)
	case lex.KND_UINT:
		return e.uintSubId(idTok, m)
	case lex.KND_INT:
		return e.intSubId(idTok, m)
	case lex.KND_F32:
		return e.f32SubId(idTok, m)
	case lex.KND_F64:
		return e.f64SubId(idTok, m)
	}
	e.pusherrtok(idTok, "obj_not_support_sub_fields", typeTok.Kind)
	return
}

func (e *eval) typeId(toks []lex.Token, m *exprModel) (v value) {
	v.data.DataType.Id = types.VOID
	v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	r := new_builder(nil)
	i := 0
	t, ok := r.DataType(toks, &i, true)
	if !ok {
		e.p.pusherrs(r.Errors...)
		return
	} else if i+1 >= len(toks) {
		e.pusherrtok(toks[0], "invalid_syntax")
		return
	}
	t, ok = e.p.realType(t, true)
	if !ok {
		return
	}
	toks = toks[i+1:]
	if types.IsPure(t) && types.IsStruct(t) {
		if toks[0].Id != lex.ID_BRACE || toks[0].Kind != lex.KND_LBRACE {
			e.pusherrtok(toks[0], "invalid_syntax")
			return
		}
		s := t.Tag.(*ast.Struct)
		return e.p.callStructConstructor(s, toks, m)
	}
	if toks[0].Id != lex.ID_BRACE || toks[0].Kind != lex.KND_LBRACKET {
		e.pusherrtok(toks[0], "invalid_syntax")
		return
	}
	return e.enumerable(toks, t, m)
}

func (e *eval) xObjSubId(dm *ast.Defmap, val value, interior_mutability bool, idTok lex.Token, m *exprModel) (v value) {
	i, dm, t := dm.FindById(idTok.Kind, idTok.File)
	if i == -1 {
		e.pusherrtok(idTok, "obj_have_not_id", idTok.Kind)
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
		v.mutable = v.mutable || (g.Mutable && interior_mutability)
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

func (e *eval) strObjSubId(val value, idTok lex.Token, m *exprModel) value {
	readyStrDefines(val)
	v := e.xObjSubId(strDefines, val, false, idTok, m)
	v.lvalue = false
	return v
}

func (e *eval) sliceObjSubId(val value, idTok lex.Token, m *exprModel) value {
	v := e.xObjSubId(sliceDefines, val, false, idTok, m)
	v.lvalue = false
	return v
}

func (e *eval) arrayObjSubId(val value, idTok lex.Token, m *exprModel) value {
	readyArrayDefines(val)
	v := e.xObjSubId(arrayDefines, val, false, idTok, m)
	v.lvalue = false
	return v
}

func (e *eval) mapObjSubId(val value, idTok lex.Token, m *exprModel) value {
	readyMapDefines(val.data.DataType)
	v := e.xObjSubId(mapDefines, val, false, idTok, m)
	v.lvalue = false
	return v
}

func (e *eval) enumSubId(val value, idTok lex.Token, m *exprModel) (v value) {
	enum := val.data.DataType.Tag.(*Enum)
	v = val
	v.lvalue = false
	v.is_type = false
	item := enum.ItemById(idTok.Kind)
	if item == nil {
		e.pusherrtok(idTok, "obj_have_not_id", idTok.Kind)
	} else {
		v.expr = item.ExprTag
		v.model = getModel(v)
	}
	nodes := m.nodes[m.index]
	nodes.nodes[len(nodes.nodes)-1] = v.model
	return
}

func (e *eval) structObjSubId(val value, idTok lex.Token, m *exprModel) value {
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
	val = e.xObjSubId(s.Defines, val, interior_mutability, idTok, m)
	if types.IsFn(val.data.DataType) {
		f := val.data.DataType.Tag.(*Fn)
		if f.Receiver != nil && types.IsRef(f.Receiver.DataType) && !types.IsRef(parent_type) {
			e.p.pusherrtok(idTok, "ref_method_used_with_not_ref_instance")
		}
	}
	return val
}

func (e *eval) traitObjSubId(val value, idTok lex.Token, m *exprModel) value {
	m.append_sub(exprNode{"._get()"})
	t := val.data.DataType.Tag.(*ast.Trait)
	val.constant = false
	val.lvalue = false
	val.is_type = false
	val = e.xObjSubId(t.Defines, val, false, idTok, m)
	val.constant = false
	return val
}

type ns_find interface {
	NsById(string) *ast.Namespace
}

func (e *eval) getNs(toks *[]lex.Token) *ast.Defmap {
	var prev ns_find = e.p
	var ns *ast.Namespace
	for i, tok := range *toks {
		if (i+1)%2 != 0 {
			if tok.Id != lex.ID_IDENT {
				e.pusherrtok(tok, "invalid_syntax")
				continue
			}
			src := prev.NsById(tok.Kind)
			if src == nil {
				if ns != nil {
					*toks = (*toks)[i:]
					return ns.Defines
				}
				e.pusherrtok(tok, "namespace_not_exist", tok.Kind)
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

func (e *eval) nsSubId(toks []lex.Token, m *exprModel) (v value) {
	defs := e.getNs(&toks)
	if defs == nil {
		return
	}
	// Temporary clean of local defines
	// Because this defines has high priority
	// So shadows defines of namespace
	blockTypes := e.p.blockTypes
	blockVars := e.p.blockVars
	e.p.blockTypes = nil
	e.p.blockVars = nil
	pdefs := e.p.Defines
	e.p.Defines = defs
	e.p.allowBuiltin = false
	v, _ = e.single(toks[0], m)
	e.p.allowBuiltin = true
	e.p.blockTypes = blockTypes
	e.p.blockVars = blockVars
	e.p.Defines = pdefs
	return
}

func (e *eval) id(toks []lex.Token, m *exprModel) (v value) {
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
		return e.subId(toks, m)
	case lex.ID_DBLCOLON:
		return e.nsSubId(toks, m)
	}
	e.pusherrtok(toks[i], "invalid_syntax")
	return
}

func (e *eval) operatorRight(toks []lex.Token, m *exprModel) (v value) {
	tok := toks[len(toks)-1]
	switch tok.Kind {
	case lex.KND_TRIPLE_DOT:
		toks = toks[:len(toks)-1]
		return e.variadic(toks, m, tok)
	default:
		e.pusherrtok(tok, "invalid_syntax")
	}
	return
}

func readyToVariadic(v *value) {
	if v.data.DataType.Id != types.STR || !types.IsPure(v.data.DataType) {
		return
	}
	v.data.DataType.Id = types.SLICE
	v.data.DataType.ComponentType = new(Type)
	v.data.DataType.ComponentType.Id = types.U8
	v.data.DataType.ComponentType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	v.data.DataType.Kind = lex.PREFIX_SLICE + v.data.DataType.ComponentType.Kind
}

func (e *eval) variadic(toks []lex.Token, m *exprModel, errtok lex.Token) (v value) {
	v = e.process(toks, m)
	readyToVariadic(&v)
	if !types.IsVariadicable(v.data.DataType) {
		e.pusherrtok(errtok, "variadic_with_non_variadicable", v.data.DataType.Kind)
		return
	}
	v.data.DataType = *v.data.DataType.ComponentType
	v.variadic = true
	v.constant = false
	return
}

func (e *eval) bracketRange(toks []lex.Token, m *exprModel) (v value) {
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
		v, model = e.build_slice_implicit(e.enumerableParts(toks), toks[0])
		m.append_sub(model)
		return v
	case len(exprToks) == 0 || brace_n > 0:
		e.pusherrtok(errTok, "invalid_syntax")
		return
	}
	var model ast.ExprModel
	v, model = e.eval_toks(exprToks)
	m.append_sub(model)
	toks = toks[len(exprToks):] // lex.Tokenens of [...]
	i := 0
	if toks, colon := ast.SplitColon(toks, &i); colon != -1 {
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
			m.append_sub(indexingExprModel(model))
			e.checkIntegerIndexing(leftv, errTok)
		} else {
			leftv.expr = int64(0)
			leftv.model = numericModel(leftv)
			m.append_sub(exprNode{"0"})
		}
		if len(rightToks) > 0 {
			m.append_sub(exprNode{","})
			var model ast.ExprModel
			rightv, model = e.p.evalToks(rightToks, nil)
			m.append_sub(indexingExprModel(model))
			e.checkIntegerIndexing(rightv, errTok)
		}
		m.append_sub(exprNode{")"})
		v = e.slicing(v, leftv, rightv, errTok)
		if !types.IsMut(v.data.DataType) {
			v.data.Value = " "
		}
		return v
	}
	m.append_sub(exprNode{lex.KND_LBRACKET})
	indexv, model := e.eval_toks(toks[1 : len(toks)-1])
	m.append_sub(indexingExprModel(model))
	m.append_sub(exprNode{lex.KND_RBRACKET})
	v = e.indexing(v, indexv, errTok)
	if !types.IsMut(v.data.DataType) {
		v.data.Value = " "
	}
	// Ignore indexed type from original
	v.data.DataType.Pure = true
	v.data.DataType.Original = nil
	return v
}

func (e *eval) checkIntegerIndexing(v value, errtok lex.Token) {
	err_key := check_value_for_indexing(v)
	if err_key != "" {
		e.pusherrtok(errtok, err_key)
	}
}

func (e *eval) indexing(enumv, indexv value, errtok lex.Token) (v value) {
	switch {
	case types.IsExplicitPtr(enumv.data.DataType):
		return e.indexing_explicit_ptr(enumv, indexv, errtok)
	case types.IsArray(enumv.data.DataType):
		return e.indexingArray(enumv, indexv, errtok)
	case types.IsSlice(enumv.data.DataType):
		return e.indexingSlice(enumv, indexv, errtok)
	case types.IsMap(enumv.data.DataType):
		return e.indexingMap(enumv, indexv, errtok)
	case types.IsPure(enumv.data.DataType):
		switch enumv.data.DataType.Id {
		case types.STR:
			return e.indexingStr(enumv, indexv, errtok)
		}
	}
	e.pusherrtok(errtok, "not_supports_indexing", enumv.data.DataType.Kind)
	return
}

func (e *eval) indexingSlice(slicev, index value, errtok lex.Token) value {
	slicev.data.DataType = *slicev.data.DataType.ComponentType
	e.checkIntegerIndexing(index, errtok)
	return slicev
}

func (e *eval) indexing_explicit_ptr(ptrv, index value, errtok lex.Token) value {
	if !e.unsafe_allowed() {
		e.pusherrtok(errtok, "unsafe_behavior_at_out_of_unsafe_scope")
	}
	ptrv.data.DataType = types.DerefPtrOrRef(ptrv.data.DataType)
	e.checkIntegerIndexing(index, errtok)
	return ptrv
}

func (e *eval) indexingArray(arrv, index value, errtok lex.Token) value {
	arrv.data.DataType = *arrv.data.DataType.ComponentType
	e.checkIntegerIndexing(index, errtok)
	return arrv
}

func (e *eval) indexingMap(mapv, leftv value, errtok lex.Token) value {
	types := mapv.data.DataType.Tag.([]Type)
	keyType := types[0]
	valType := types[1]
	mapv.data.DataType = valType
	e.p.check_type(keyType, leftv.data.DataType, false, true, errtok)
	return mapv
}

func (e *eval) indexingStr(strv, index value, errtok lex.Token) value {
	strv.data.DataType.Id = types.U8
	strv.data.DataType.Kind = types.TYPE_MAP[strv.data.DataType.Id]
	e.checkIntegerIndexing(index, errtok)
	if !index.constant {
		strv.constant = false
		return strv
	}
	if strv.constant {
		i := tonums(index.expr)
		s := strv.expr.(string)
		if int(i) >= len(s) {
			e.p.pusherrtok(errtok, "overflow_limits")
		} else {
			strv.expr = uint64(s[i])
			strv.model = numericModel(strv)
		}
	}
	return strv
}

func (e *eval) slicing(enumv, leftv, rightv value, errtok lex.Token) (v value) {
	switch {
	case types.IsArray(enumv.data.DataType):
		return e.slicingArray(enumv, errtok)
	case types.IsSlice(enumv.data.DataType):
		return e.slicingSlice(enumv, errtok)
	case types.IsPure(enumv.data.DataType):
		switch enumv.data.DataType.Id {
		case types.STR:
			return e.slicingStr(enumv, leftv, rightv, errtok)
		}
	}
	e.pusherrtok(errtok, "not_supports_slicing", enumv.data.DataType.Kind)
	return
}

func (e *eval) slicingSlice(v value, errtok lex.Token) value {
	v.lvalue = false
	return v
}

func (e *eval) slicingArray(v value, errtok lex.Token) value {
	v.lvalue = false
	v.data.DataType.Id = types.SLICE
	v.data.DataType.Kind = lex.PREFIX_SLICE + v.data.DataType.ComponentType.Kind
	return v
}

func (e *eval) slicingStr(v, leftv, rightv value, errtok lex.Token) value {
	v.lvalue = false
	v.data.DataType.Id = types.STR
	v.data.DataType.Kind = types.TYPE_MAP[v.data.DataType.Id]
	if !v.constant {
		return v
	}
	if rightv.constant {
		right := tonums(rightv.expr)
		s := v.expr.(string)
		if int(right) >= len(s) {
			e.p.pusherrtok(errtok, "overflow_limits")
		}
	}
	if leftv.constant && rightv.constant {
		left := tonums(leftv.expr)
		if left < 0 {
			return v
		}
		s := v.expr.(string)
		var right int64
		if rightv.expr == nil {
			right = int64(len(s))
		} else {
			right = tonums(rightv.expr)
		}
		if left > right {
			return v
		}
		v.expr = s[left:right]
		v.model = strModel(v)
	} else {
		v.constant = false
	}
	return v
}

// ! IMPORTANT: lex.Tokenens is should be store enumerable parentheses.
func (e *eval) enumerableParts(toks []lex.Token) [][]lex.Token {
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
			expr_t: *t.ComponentType,
			v:      partVal,
			errtok: part[0],
		}.check()
	}
	e.type_prefix = old_type
	return v, model
}

func (e *eval) build_slice_implicit(parts [][]lex.Token, errtok lex.Token) (value, ast.ExprModel) {
	if len(parts) == 0 {
		e.pusherrtok(errtok, "dynamic_type_annotation_failed")
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
			expr_t: *model.dataType.ComponentType,
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
			expr_t: *t.ComponentType,
			v:      partVal,
			errtok: part[0],
		}.check()
	}
	e.type_prefix = old_type
	return v, model
}

func (e *eval) buildMap(parts [][]lex.Token, t Type, errtok lex.Token) (value, ast.ExprModel) {
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
			e.pusherrtok(errtok, "missing_expr")
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
			expr_t: keyType,
			v:      key,
			errtok: colonTok,
		}.check()
		assign_checker{
			p:      e.p,
			expr_t: valType,
			v:      val,
			errtok: colonTok,
		}.check()
	}
	return v, model
}

func (e *eval) enumerable(exprToks []lex.Token, t Type, m *exprModel) (v value) {
	var model ast.ExprModel
	t, ok := e.p.realType(t, true)
	if !ok {
		return
	}
	errtok := exprToks[0]
	switch {
	case types.IsArray(t):
		v, model = e.build_array(e.enumerableParts(exprToks), t, errtok)
	case types.IsSlice(t):
		v, model = e.build_slice_explicit(e.enumerableParts(exprToks), t, errtok)
	case types.IsMap(t):
		v, model = e.buildMap(e.enumerableParts(exprToks), t, errtok)
	default:
		e.pusherrtok(errtok, "invalid_type_source")
		return
	}
	m.append_sub(model)
	return
}

func (e *eval) anonymousFn(toks []lex.Token, m *exprModel) (v value) {
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

func (e *eval) unsafeEval(toks []lex.Token, m *exprModel) (v value) {
	i := 0
	rang := ast.Range(&i, lex.KND_LBRACE, lex.KND_RBRACE, toks)
	if len(rang) == 0 {
		e.pusherrtok(toks[0], "missing_expr")
		return
	}
	old := e.allow_unsafe
	e.allow_unsafe = true
	v = e.process(rang, m)
	e.allow_unsafe = old
	return v
}

func (e *eval) braceRange(toks []lex.Token, m *exprModel) (v value) {
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
		e.pusherrtok(toks[0], "invalid_syntax")
		return
	}
	switch exprToks[0].Id {
	case lex.ID_UNSAFE:
		if len(toks) == 0 {
			e.pusherrtok(toks[0], "invalid_syntax")
			return
		} else if toks[1].Id != lex.ID_FN {
			return e.unsafeEval(toks[1:], m)
		}
		fallthrough
	case lex.ID_FN:
		return e.anonymousFn(toks, m)
	case lex.ID_IDENT, lex.ID_CPP:
		return e.typeId(toks, m)
	default:
		e.pusherrtok(exprToks[0], "invalid_syntax")
	}
	return
}
