package transpiler

import (
	"strconv"
	"strings"

	"github.com/jule-lang/jule/ast"
	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/jule"
	"github.com/jule-lang/jule/pkg/juletype"
)

type value struct {
	data      models.Data
	model     iExpr
	expr      any
	constExpr bool
	lvalue    bool
	variadic  bool
	is_type   bool
	mutable   bool
}

type eval struct {
	t            *Transpiler
	has_error    bool
	type_prefix  *Type
	allow_unsafe bool
}

func (e *eval) pusherrtok(tok lex.Token, err string, args ...any) {
	if e.has_error {
		return
	}
	e.has_error = true
	e.t.pusherrtok(tok, err, args...)
}

func (e *eval) eval_toks(toks []lex.Token) (value, iExpr) {
	builder := ast.Parser{}
	return e.eval_expr(builder.Expr(toks))
}

func (e *eval) eval_expr(expr Expr) (value, iExpr) {
	return e.eval(expr.Op)
}

func (e *eval) eval_op(op any) (v value, model iExpr) {
	switch t := op.(type) {
	case models.BinopExpr:
		m := newExprModel(1)
		model = m
		v = e.process(t.Tokens, m)
		if v.constExpr {
			model = v.model
		} else if v.is_type {
			e.pusherrtok(v.data.Token, "invalid_expr")
		}
		return
	case models.Binop:
	default:
		return
	}
	bop := op.(models.Binop)
	l, lm := e.eval_op(bop.L)
	if e.has_error {
		return
	}
	r, rm := e.eval_op(bop.R)
	if e.has_error {
		return
	}
	process := solver{
		t: e.t,
		op: bop.Op,
		l: l,
		r: r,
	}
	v = process.solve()
	v.lvalue = typeIsLvalue(v.data.Type)
	if v.constExpr {
		model = v.model
	} else {
		m := newExprModel(1)
		model = m
		m.appendSubNode(exprNode{tokens.LPARENTHESES})
		m.appendSubNode(lm)
		m.appendSubNode(exprNode{" " + bop.Op.Kind + " "})
		m.appendSubNode(rm)
		m.appendSubNode(exprNode{tokens.RPARENTHESES})
	}
	return
}

func (e *eval) eval(op any) (v value, model iExpr) {
	defer func() {
		if typeIsVoid(v.data.Type) {
			v.data.Type.Id = juletype.Void
			v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		} else if v.constExpr && typeIsPure(v.data.Type) && isConstExpression(v.data.Value) {
			switch v.expr.(type) {
			case int64:
				dt := Type{
					Id:   juletype.Int,
					Kind: juletype.TypeMap[juletype.Int],
				}
				if integerAssignable(dt.Id, v) {
					v.data.Type = dt
				}
			case uint64:
				dt := Type{
					Id:   juletype.UInt,
					Kind: juletype.TypeMap[juletype.UInt],
				}
				if integerAssignable(dt.Id, v) {
					v.data.Type = dt
				}
			}
		}
	}()
	if op == nil || e.has_error {
		return
	}
	return e.eval_op(op)
}

func (e *eval) single(tok lex.Token, m *exprModel) (v value, ok bool) {
	eval := valueEvaluator{tok, m, e.t}
	v.data.Type.Id = juletype.Void
	v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	v.data.Token = tok
	switch tok.Id {
	case tokens.Value:
		ok = true
		switch {
		case isstr(tok.Kind):
			v = eval.str()
		case ischar(tok.Kind):
			v = eval.char()
		case isbool(tok.Kind):
			v = eval.bool()
		case isnil(tok.Kind):
			v = eval.nil()
		default:
			v = eval.numeric()
		}
	case tokens.Id, tokens.Self:
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
	processor := unary{toks[0], exprToks, m, e.t}
	if processor.toks == nil {
		e.pusherrtok(processor.token, "invalid_syntax")
		return v
	}
	switch processor.token.Kind {
	case tokens.MINUS:
		m.appendSubNode(exprNode{processor.token.Kind})
		v = processor.minus()
	case tokens.PLUS:
		m.appendSubNode(exprNode{processor.token.Kind})
		v = processor.plus()
	case tokens.CARET:
		m.appendSubNode(exprNode{"~"})
		v = processor.caret()
	case tokens.EXCLAMATION:
		m.appendSubNode(exprNode{processor.token.Kind})
		v = processor.logicalNot()
	case tokens.STAR:
		m.appendSubNode(exprNode{processor.token.Kind})
		v = processor.star()
	case tokens.AMPER:
		m.appendSubNode(exprNode{processor.token.Kind})
		v = processor.amper()
	default:
		e.pusherrtok(processor.token, "invalid_syntax")
	}
	v.data.Token = processor.token
	return v
}

func (e *eval) betweenParentheses(toks []lex.Token, m *exprModel) value {
	// Write parentheses.
	m.appendSubNode(exprNode{tokens.LPARENTHESES})
	defer m.appendSubNode(exprNode{tokens.RPARENTHESES})

	tk := toks[0]
	toks = toks[1 : len(toks)-1]
	if len(toks) == 0 {
		e.pusherrtok(tk, "invalid_syntax")
	}
	val, model := e.eval_toks(toks)
	m.appendSubNode(model)
	return val
}

func (e *eval) dataTypeFunc(expr lex.Token, callRange []lex.Token, m *exprModel) (v value, isret bool) {
	switch expr.Id {
	case tokens.DataType:
		switch expr.Kind {
		case tokens.STR:
			m.appendSubNode(exprNode{"__julec_to_str("})
			_, vm := e.t.evalToks(callRange)
			m.appendSubNode(vm)
			m.appendSubNode(exprNode{tokens.RPARENTHESES})
			v.data.Type = Type{
				Id:   juletype.Str,
				Kind: juletype.TypeMap[juletype.Str],
			}
			isret = true
		default:
			dt := Type{
				Token: expr,
				Id:    juletype.TypeFromId(expr.Kind),
				Kind:  expr.Kind,
			}
			isret = true
			v = e.castExpr(dt, callRange, m, expr)
		}
	case tokens.Id:
		def, _, _ := e.t.defById(expr.Kind)
		if def == nil {
			break
		}
		switch t := def.(type) {
		case *TypeAlias:
			dt, ok := e.t.realType(t.Type, true)
			if !ok || typeIsStruct(dt) {
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
	if tok.Id == tokens.Brace && tok.Kind == tokens.RBRACKET {
		data.expr, data.generics = ast.RangeLast(data.expr)
	}
	return
}

func (e *eval) unsafe_allowed() bool {
	return e.allow_unsafe || e.t.unsafe_allowed()
}

func (e *eval) callFunc(f *Func, data callData, m *exprModel) value {
	if !e.unsafe_allowed() && f.IsUnsafe {
		e.pusherrtok(data.expr[0], "unsafe_behavior_at_out_of_unsafe_scope")
	}
	if f.BuiltinCaller != nil {
		return f.BuiltinCaller.(BuiltinCaller)(e.t, f, data, m)
	}
	return e.t.callFunc(f, data, m)
}

func (e *eval) parenthesesRange(toks []lex.Token, m *exprModel) (v value) {
	tok := toks[0]
	switch tok.Id {
	case tokens.Brace:
		switch tok.Kind {
		case tokens.LPARENTHESES:
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
	case tokens.DataType, tokens.Id:
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
	case typeIsFunc(v.data.Type):
		f := v.data.Type.Tag.(*Func)
		if f.Receiver != nil && f.Receiver.Mutable && !v.mutable {
			e.pusherrtok(data.expr[len(data.expr)-1], "mutable_operation_on_immutable")
		}
		return e.callFunc(f, data, m)
	}
	e.pusherrtok(data.expr[len(data.expr)-1], "invalid_syntax")
	return
}

func (e *eval) try_cpp_linked_var(toks []lex.Token, m *exprModel) (v value, ok bool) {
	if toks[0].Id != tokens.Cpp {
		return
	} else if toks[1].Id != tokens.Dot {
		e.pusherrtok(toks[1], "invalid_syntax")
		return
	}
	tok := toks[2]
	if tok.Id != tokens.Id {
		e.pusherrtok(toks[2], "invalid_syntax")
		return
	}
	def, def_t := e.t.linkById(tok.Kind)
	if def_t == ' ' {
		e.pusherrtok(tok, "id_not_exist", tok.Kind)
		return
	}
	m.appendSubNode(exprNode{tok.Kind})
	ok = true
	switch def_t {
	case 'f':
		v = make_value_from_fn(def.(*models.Fn))
	case 'v':
		v = make_value_from_var(def.(*models.Var))
	case 's':
		v = make_value_from_struct(def.(*structure))
	}
	return
}

func (e *eval) process(toks []lex.Token, m *exprModel) (v value) {
	v.constExpr = true
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
	case tokens.Operator:
		return e.unary(toks, m)
	}
	tok = toks[len(toks)-1]
	switch tok.Id {
	case tokens.Id:
		return e.id(toks, m)
	case tokens.Operator:
		return e.operatorRight(toks, m)
	case tokens.Brace:
		switch tok.Kind {
		case tokens.RPARENTHESES:
			return e.parenthesesRange(toks, m)
		case tokens.RBRACE:
			return e.braceRange(toks, m)
		case tokens.RBRACKET:
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
		if tok.Id == tokens.DataType {
			return e.typeSubId(tok, idTok, m)
		} else if tok.Id == tokens.Id {
			t, _, _ := e.t.typeById(tok.Kind)
			if t != nil {
				return e.typeSubId(t.Type.Token, idTok, m)
			}
		}
	}
	val := e.process(toks, m)
	checkType := val.data.Type
	if typeIsExplicitPtr(checkType) {
		if toks[0].Id != tokens.Self && !e.unsafe_allowed() {
			e.pusherrtok(idTok, "unsafe_behavior_at_out_of_unsafe_scope")
		}
		checkType = un_ptr_or_ref_type(checkType)
	} else if typeIsRef(checkType) {
		checkType = un_ptr_or_ref_type(checkType)
	}
	switch {
	case typeIsPure(checkType):
		switch {
		case checkType.Id == juletype.Str:
			return e.strObjSubId(val, idTok, m)
		case valIsEnumType(val):
			return e.enumSubId(val, idTok, m)
		case valIsStructIns(val):
			return e.structObjSubId(val, idTok, m)
		case valIsTraitIns(val):
			return e.traitObjSubId(val, idTok, m)
		}
	case typeIsSlice(checkType):
		return e.sliceObjSubId(val, idTok, m)
	case typeIsArray(checkType):
		return e.arrayObjSubId(val, idTok, m)
	case typeIsMap(checkType):
		return e.mapObjSubId(val, idTok, m)
	}
	e.pusherrtok(dotTok, "obj_not_support_sub_fields", val.data.Type.Kind)
	return
}

func (e *eval) get_cast_expr_model(t, vt Type, expr_model iExpr) iExpr {
	var model strings.Builder
	switch {
	case typeIsPtr(vt) || typeIsPtr(t):
		model.WriteString("((")
		model.WriteString(t.String())
		model.WriteString(")(")
		model.WriteString(expr_model.String())
		model.WriteString("))")
		goto end
	case typeIsPure(vt):
		switch {
		case typeIsTrait(vt):
			model.WriteString(expr_model.String())
			model.WriteString(subIdAccessorOfType(vt))
			model.WriteString("operator ")
			model.WriteString(t.String())
			model.WriteString("()")
			goto end
		}
	}
	model.WriteString("static_cast<")
	model.WriteString(t.String())
	model.WriteString(">(")
	model.WriteString(expr_model.String())
	model.WriteByte(')')
end:
	return exprNode{model.String()}
}

func (e *eval) castExpr(dt Type, exprToks []lex.Token, m *exprModel, errTok lex.Token) value {
	val, model := e.eval_toks(exprToks)
	m.appendSubNode(e.get_cast_expr_model(dt, val.data.Type, model))
	val = e.cast(val, dt, errTok)
	return val
}

func (e *eval) tryCast(toks []lex.Token, m *exprModel) (v value, _ bool) {
	brace_n := 0
	errTok := toks[0]
	for i, tok := range toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
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
		b := ast.NewBuilder(nil)
		dtindex := 0
		typeToks := toks[1:i]
		dt, ok := b.DataType(typeToks, &dtindex, false, false)
		b.Wait()
		if !ok {
			return
		}
		dt, ok = e.t.realType(dt, false)
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
		if tok.Id != tokens.Brace || tok.Kind != tokens.LPARENTHESES {
			return
		}
		exprToks, ok = e.t.getrange(tokens.LPARENTHESES, tokens.RPARENTHESES, exprToks)
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
	case typeIsPtr(t):
		e.castPtr(t, &v, errtok)
	case typeIsRef(t):
		e.cast_ref(t, &v, errtok)
	case typeIsSlice(t):
		e.castSlice(t, v.data.Type, errtok)
	case typeIsStruct(t):
		e.castStruct(t, &v, errtok)
	case typeIsPure(t):
		if v.data.Type.Id == juletype.Any {
			// The any type supports casting to any data type.
			break
		}
		e.castPure(t, &v, errtok)
	default:
		e.pusherrtok(errtok, "type_not_supports_casting", t.Kind)
	}
	v.data.Value = t.Kind
	v.data.Type = t
	v.lvalue = typeIsLvalue(t)
	v.mutable = typeIsRef(t) || type_is_mutable(t)
	if v.constExpr {
		var model exprNode
		model.value = v.data.Type.String()
		model.value += tokens.LPARENTHESES
		model.value += v.model.String()
		model.value += tokens.RPARENTHESES
		v.model = model
	}
	return v
}

func (e *eval) castStruct(t Type, v *value, errtok lex.Token) {
	if !typeIsTrait(v.data.Type) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.Type.Kind, t.Kind)
		return
	}
	s := t.Tag.(*structure)
	tr := v.data.Type.Tag.(*trait)
	if !s.hasTrait(tr) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.Type.Kind, t.Kind)
	}
}

func (e *eval) castPtr(t Type, v *value, errtok lex.Token) {
	if !e.unsafe_allowed() {
		e.pusherrtok(errtok, "unsafe_behavior_at_out_of_unsafe_scope")
		return
	}
	if !typeIsPtr(v.data.Type) &&
		!typeIsPure(v.data.Type) &&
		!juletype.IsInteger(v.data.Type.Id) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.Type.Kind, t.Kind)
	}
	v.constExpr = false
}

func (e *eval) cast_ref(t Type, v *value, errtok lex.Token) {
	if typeIsStruct(un_ptr_or_ref_type(t)) {
		e.castStruct(t, v, errtok)
		return
	}
	e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.Type.Kind, t.Kind)
}

func (e *eval) castPure(t Type, v *value, errtok lex.Token) {
	switch t.Id {
	case juletype.Any:
		return
	case juletype.Str:
		e.castStr(v.data.Type, errtok)
		return
	}
	switch {
	case juletype.IsInteger(t.Id):
		e.castInteger(t, v, errtok)
	case juletype.IsNumeric(t.Id):
		e.castNumeric(t, v, errtok)
	default:
		e.pusherrtok(errtok, "type_not_supports_casting", t.Kind)
	}
}

func (e *eval) castStr(t Type, errtok lex.Token) {
	if typeIsPure(t) || (t.Id != juletype.U8 && t.Id != juletype.I32) {
		return
	}
	if !typeIsSlice(t) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", juletype.TypeMap[juletype.Str], t.Kind)
		return
	}
	t = *t.ComponentType
	if !typeIsPure(t) || (t.Id != juletype.U8 && t.Id != juletype.I32) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", juletype.TypeMap[juletype.Str], t.Kind)
	}
}

func (e *eval) castInteger(t Type, v *value, errtok lex.Token) {
	if v.constExpr {
		switch {
		case juletype.IsSignedInteger(t.Id):
			v.expr = tonums(v)
		default:
			v.expr = tonumu(v)
		}
	}
	if typeIsEnum(v.data.Type) {
		e := v.data.Type.Tag.(*Enum)
		if juletype.IsNumeric(e.Type.Id) {
			return
		}
	}
	if typeIsPtr(v.data.Type) {
		if t.Id == juletype.UIntptr {
			return
		} else if !e.unsafe_allowed() {
			e.pusherrtok(errtok, "unsafe_behavior_at_out_of_unsafe_scope")
			return
		} else if t.Id != juletype.I32 && t.Id != juletype.I64 &&
			t.Id != juletype.U16 && t.Id != juletype.U32 && t.Id != juletype.U64 {
			e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.Type.Kind, t.Kind)
		}
		return
	}
	if typeIsPure(v.data.Type) && juletype.IsNumeric(v.data.Type.Id) {
		return
	}
	e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.Type.Kind, t.Kind)
}

func (e *eval) castNumeric(t Type, v *value, errtok lex.Token) {
	if v.constExpr {
		switch {
		case juletype.IsFloat(t.Id):
			v.expr = tonumf(v)
		case juletype.IsSignedInteger(t.Id):
			v.expr = tonums(v)
		default:
			v.expr = tonumu(v)
		}
	}
	if typeIsEnum(v.data.Type) {
		e := v.data.Type.Tag.(*Enum)
		if juletype.IsNumeric(e.Type.Id) {
			return
		}
	}
	if typeIsPure(v.data.Type) && juletype.IsNumeric(v.data.Type.Id) {
		return
	}
	e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.Type.Kind, t.Kind)
}

func (e *eval) castSlice(t, vt Type, errtok lex.Token) {
	if !typeIsPure(vt) || vt.Id != juletype.Str {
		e.pusherrtok(errtok, "type_not_supports_casting_to", vt.Kind, t.Kind)
		return
	}
	t = *t.ComponentType
	if !typeIsPure(t) || (t.Id != juletype.U8 && t.Id != juletype.I32) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", vt.Kind, t.Kind)
	}
}

func (e *eval) juletypeSubId(dm *DefineMap, idTok lex.Token, m *exprModel) (v value) {
	i, dm, t := dm.findById(idTok.Kind, nil)
	if i == -1 {
		e.pusherrtok(idTok, "obj_have_not_id", idTok.Kind)
		return
	}
	v.lvalue = false
	v.data.Value = idTok.Kind
	switch t {
	case 'g':
		g := dm.Globals[i]
		v.data.Type = g.Type
		v.constExpr = g.Const
		if v.constExpr {
			v.expr = g.ExprTag
			v.model = g.Expr.Model
			m.appendSubNode(v.model)
		} else {
			m.appendSubNode(exprNode{g.Tag.(string)})
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
	case tokens.I8:
		return e.i8SubId(idTok, m)
	case tokens.I16:
		return e.i16SubId(idTok, m)
	case tokens.I32:
		return e.i32SubId(idTok, m)
	case tokens.I64:
		return e.i64SubId(idTok, m)
	case tokens.U8:
		return e.u8SubId(idTok, m)
	case tokens.U16:
		return e.u16SubId(idTok, m)
	case tokens.U32:
		return e.u32SubId(idTok, m)
	case tokens.U64:
		return e.u64SubId(idTok, m)
	case tokens.UINT:
		return e.uintSubId(idTok, m)
	case tokens.INT:
		return e.intSubId(idTok, m)
	case tokens.F32:
		return e.f32SubId(idTok, m)
	case tokens.F64:
		return e.f64SubId(idTok, m)
	}
	e.pusherrtok(typeTok, "obj_not_support_sub_fields", typeTok.Kind)
	return
}

func (e *eval) typeId(toks []lex.Token, m *exprModel) (v value) {
	v.data.Type.Id = juletype.Void
	v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	b := ast.NewBuilder(nil)
	i := 0
	t, ok := b.DataType(toks, &i, true, true)
	b.Wait()
	if !ok {
		e.t.pusherrs(b.Errors...)
		return
	} else if i+1 >= len(toks) {
		e.pusherrtok(toks[0], "invalid_syntax")
		return
	}
	t, ok = e.t.realType(t, true)
	if !ok {
		return
	}
	toks = toks[i+1:]
	if typeIsPure(t) && typeIsStruct(t) {
		if toks[0].Id != tokens.Brace || toks[0].Kind != tokens.LBRACE {
			e.pusherrtok(toks[0], "invalid_syntax")
			return
		}
		s := t.Tag.(*structure)
		return e.t.callStructConstructor(s, toks, m)
	}
	if toks[0].Id != tokens.Brace || toks[0].Kind != tokens.LBRACKET {
		e.pusherrtok(toks[0], "invalid_syntax")
		return
	}
	return e.enumerable(toks, t, m)
}

func (e *eval) xObjSubId(dm *DefineMap, val value, interior_mutability bool, idTok lex.Token, m *exprModel) (v value) {
	i, dm, t := dm.findById(idTok.Kind, idTok.File)
	if i == -1 {
		e.pusherrtok(idTok, "obj_have_not_id", idTok.Kind)
		return
	}
	v = val
	m.appendSubNode(exprNode{subIdAccessorOfType(val.data.Type)})
	switch t {
	case 'g':
		g := dm.Globals[i]
		g.Used = true
		v.data.Type = g.Type
		v.lvalue = val.lvalue || typeIsLvalue(g.Type)
		v.mutable = v.mutable || (g.Mutable && interior_mutability)
		v.constExpr = g.Const
		if g.Const {
			v.expr = g.ExprTag
			v.model = g.Expr.Model
		}
		if g.Tag != nil {
			m.appendSubNode(exprNode{g.Tag.(string)})
		} else {
			m.appendSubNode(exprNode{g.OutId()})
		}
	case 'f':
		f := dm.Funcs[i]
		f.used = true
		v.data.Type.Id = juletype.Fn
		v.data.Type.Tag = f.Ast
		v.data.Type.Kind = f.Ast.DataTypeString()
		v.data.Token = f.Ast.Token
		m.appendSubNode(exprNode{f.Ast.Id})
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
	v := e.xObjSubId(arrayDefines, val, false, idTok, m)
	v.lvalue = false
	return v
}

func (e *eval) mapObjSubId(val value, idTok lex.Token, m *exprModel) value {
	readyMapDefines(val.data.Type)
	v := e.xObjSubId(mapDefines, val, false, idTok, m)
	v.lvalue = false
	return v
}

func (e *eval) enumSubId(val value, idTok lex.Token, m *exprModel) (v value) {
	enum := val.data.Type.Tag.(*Enum)
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
	parent_type := val.data.Type
	s := val.data.Type.Tag.(*structure)
	val.constExpr = false
	val.is_type = false
	if val.data.Value == tokens.SELF {
		nodes := &m.nodes[m.index].nodes
		n := len(*nodes)
		defer func() {
			// Save unary
			if ast.IsUnaryOperator((*nodes)[0].String()) {
				*nodes = append([]iExpr{(*nodes)[0], exprNode{"this->"}}, (*nodes)[n+1:]...)
				return
			}
			*nodes = append([]iExpr{exprNode{"this->"}}, (*nodes)[n+1:]...)
		}()
	}
	interior_mutability := false
	if e.t.rootBlock.Func.Receiver != nil {
		switch t := e.t.rootBlock.Func.Receiver.Tag.(type) {
		case *structure:
			interior_mutability = structure_instances_is_uses_same_base(t, s)
		}
	}
	val = e.xObjSubId(s.Defines, val, interior_mutability, idTok, m)
	if typeIsFunc(val.data.Type) {
		f := val.data.Type.Tag.(*Func)
		if f.Receiver != nil && typeIsRef(f.Receiver.Type) && !typeIsRef(parent_type) {
			e.t.pusherrtok(idTok, "ref_method_used_with_not_ref_instance")
		}
	}
	return val
}

func (e *eval) traitObjSubId(val value, idTok lex.Token, m *exprModel) value {
	m.appendSubNode(exprNode{".get()"})
	t := val.data.Type.Tag.(*trait)
	val.constExpr = false
	val.lvalue = false
	val.is_type = false
	val = e.xObjSubId(t.Defines, val, false, idTok, m)
	val.constExpr = false
	return val
}

type nsFind interface {
	nsById(string) *namespace
}

func (e *eval) getNs(toks *[]lex.Token) *DefineMap {
	var prev nsFind = e.t
	var ns *namespace
	for i, tok := range *toks {
		if (i+1)%2 != 0 {
			if tok.Id != tokens.Id {
				e.pusherrtok(tok, "invalid_syntax")
				continue
			}
			src := prev.nsById(tok.Kind)
			if src == nil {
				if ns != nil {
					*toks = (*toks)[i:]
					return ns.defines
				}
				e.pusherrtok(tok, "namespace_not_exist", tok.Kind)
				return nil
			}
			prev = src.defines
			ns = src
			continue
		}
		if tok.Id != tokens.DoubleColon {
			return ns.defines
		}
	}
	return ns.defines
}

func (e *eval) nsSubId(toks []lex.Token, m *exprModel) (v value) {
	defs := e.getNs(&toks)
	if defs == nil {
		return
	}
	// Temporary clean of local defines
	// Because this defines has high priority
	// So shadows defines of namespace
	blockTypes := e.t.blockTypes
	blockVars := e.t.blockVars
	e.t.blockTypes = nil
	e.t.blockVars = nil
	pdefs := e.t.Defines
	e.t.Defines = defs
	e.t.allowBuiltin = false
	v, _ = e.single(toks[0], m)
	e.t.allowBuiltin = true
	e.t.blockTypes = blockTypes
	e.t.blockVars = blockVars
	e.t.Defines = pdefs
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
	case tokens.Dot:
		return e.subId(toks, m)
	case tokens.DoubleColon:
		return e.nsSubId(toks, m)
	}
	e.pusherrtok(toks[i], "invalid_syntax")
	return
}

func (e *eval) operatorRight(toks []lex.Token, m *exprModel) (v value) {
	tok := toks[len(toks)-1]
	switch tok.Kind {
	case tokens.TRIPLE_DOT:
		toks = toks[:len(toks)-1]
		return e.variadic(toks, m, tok)
	default:
		e.pusherrtok(tok, "invalid_syntax")
	}
	return
}

func readyToVariadic(v *value) {
	if v.data.Type.Id != juletype.Str || !typeIsPure(v.data.Type) {
		return
	}
	v.data.Type.Id = juletype.Slice
	v.data.Type.ComponentType = new(Type)
	v.data.Type.ComponentType.Id = juletype.U8
	v.data.Type.ComponentType.Kind = juletype.TypeMap[v.data.Type.Id]
	v.data.Type.Kind = jule.Prefix_Slice + v.data.Type.ComponentType.Kind
}

func (e *eval) variadic(toks []lex.Token, m *exprModel, errtok lex.Token) (v value) {
	v = e.process(toks, m)
	readyToVariadic(&v)
	if !typeIsVariadicable(v.data.Type) {
		e.pusherrtok(errtok, "variadic_with_non_variadicable", v.data.Type.Kind)
		return
	}
	v.data.Type = *v.data.Type.ComponentType
	v.variadic = true
	v.constExpr = false
	return
}

func (e *eval) bracketRange(toks []lex.Token, m *exprModel) (v value) {
	errTok := toks[0]
	var exprToks []lex.Token
	brace_n := 0
	for i := len(toks) - 1; i >= 0; i-- {
		tok := toks[i]
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.RBRACE, tokens.RBRACKET, tokens.RPARENTHESES:
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
			case typeIsArray(*e.type_prefix) || typeIsSlice(*e.type_prefix):
				return e.enumerable(toks, *e.type_prefix, m)
			}
		}
		var model iExpr
		v, model = e.build_slice_implicit(e.enumerableParts(toks), toks[0])
		m.appendSubNode(model)
		return v
	case len(exprToks) == 0 || brace_n > 0:
		e.pusherrtok(errTok, "invalid_syntax")
		return
	}
	var model iExpr
	v, model = e.eval_toks(exprToks)
	m.appendSubNode(model)
	toks = toks[len(exprToks):] // lex.Tokenens of [...]
	i := 0
	if toks, colon := ast.SplitColon(toks, &i); colon != -1 {
		i = 0
		var leftv, rightv value
		leftv.constExpr = true
		rightv.constExpr = true
		leftToks := toks[:colon]
		rightToks := toks[colon+1:]
		m.appendSubNode(exprNode{".___slice("})
		if len(leftToks) > 0 {
			var model iExpr
			leftv, model = e.t.evalToks(leftToks)
			m.appendSubNode(indexingExprModel(model))
			e.checkIntegerIndexing(leftv, errTok)
		} else {
			leftv.expr = int64(0)
			leftv.model = numericModel(leftv)
			m.appendSubNode(exprNode{"0"})
		}
		if len(rightToks) > 0 {
			m.appendSubNode(exprNode{","})
			var model iExpr
			rightv, model = e.t.evalToks(rightToks)
			m.appendSubNode(indexingExprModel(model))
			e.checkIntegerIndexing(rightv, errTok)
		}
		m.appendSubNode(exprNode{")"})
		v = e.slicing(v, leftv, rightv, errTok)
		if !type_is_mutable(v.data.Type) {
			v.data.Value = " "
		}
		return v
	}
	m.appendSubNode(exprNode{tokens.LBRACKET})
	indexv, model := e.eval_toks(toks[1 : len(toks)-1])
	m.appendSubNode(indexingExprModel(model))
	m.appendSubNode(exprNode{tokens.RBRACKET})
	v = e.indexing(v, indexv, errTok)
	if !type_is_mutable(v.data.Type) {
		v.data.Value = " "
	}
	// Ignore indexed type from original
	v.data.Type.Pure = true
	v.data.Type.Original = nil
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
	case typeIsExplicitPtr(enumv.data.Type):
		return e.indexing_explicit_ptr(enumv, indexv, errtok)
	case typeIsArray(enumv.data.Type):
		return e.indexingArray(enumv, indexv, errtok)
	case typeIsSlice(enumv.data.Type):
		return e.indexingSlice(enumv, indexv, errtok)
	case typeIsMap(enumv.data.Type):
		return e.indexingMap(enumv, indexv, errtok)
	case typeIsPure(enumv.data.Type):
		switch enumv.data.Type.Id {
		case juletype.Str:
			return e.indexingStr(enumv, indexv, errtok)
		}
	}
	e.pusherrtok(errtok, "not_supports_indexing", enumv.data.Type.Kind)
	return
}

func (e *eval) indexingSlice(slicev, index value, errtok lex.Token) value {
	slicev.data.Type = *slicev.data.Type.ComponentType
	e.checkIntegerIndexing(index, errtok)
	return slicev
}

func (e *eval) indexing_explicit_ptr(ptrv, index value, errtok lex.Token) value {
	if !e.unsafe_allowed() {
		e.pusherrtok(errtok, "unsafe_behavior_at_out_of_unsafe_scope")
	}
	ptrv.data.Type = un_ptr_or_ref_type(ptrv.data.Type)
	e.checkIntegerIndexing(index, errtok)
	return ptrv
}

func (e *eval) indexingArray(arrv, index value, errtok lex.Token) value {
	arrv.data.Type = *arrv.data.Type.ComponentType
	e.checkIntegerIndexing(index, errtok)
	return arrv
}

func (e *eval) indexingMap(mapv, leftv value, errtok lex.Token) value {
	types := mapv.data.Type.Tag.([]Type)
	keyType := types[0]
	valType := types[1]
	mapv.data.Type = valType
	e.t.checkType(keyType, leftv.data.Type, false, true, errtok)
	return mapv
}

func (e *eval) indexingStr(strv, index value, errtok lex.Token) value {
	strv.data.Type.Id = juletype.U8
	strv.data.Type.Kind = juletype.TypeMap[strv.data.Type.Id]
	e.checkIntegerIndexing(index, errtok)
	if !index.constExpr {
		strv.constExpr = false
		return strv
	}
	if strv.constExpr {
		i := tonums(index.expr)
		s := strv.expr.(string)
		if int(i) >= len(s) {
			e.t.pusherrtok(errtok, "overflow_limits")
		} else {
			strv.expr = uint64(s[i])
			strv.model = numericModel(strv)
		}
	}
	return strv
}

func (e *eval) slicing(enumv, leftv, rightv value, errtok lex.Token) (v value) {
	switch {
	case typeIsArray(enumv.data.Type):
		return e.slicingArray(enumv, errtok)
	case typeIsSlice(enumv.data.Type):
		return e.slicingSlice(enumv, errtok)
	case typeIsPure(enumv.data.Type):
		switch enumv.data.Type.Id {
		case juletype.Str:
			return e.slicingStr(enumv, leftv, rightv, errtok)
		}
	}
	e.pusherrtok(errtok, "not_supports_slicing", enumv.data.Type.Kind)
	return
}

func (e *eval) slicingSlice(v value, errtok lex.Token) value {
	v.lvalue = false
	return v
}

func (e *eval) slicingArray(v value, errtok lex.Token) value {
	v.lvalue = false
	v.data.Type.Id = juletype.Slice
	v.data.Type.Kind = jule.Prefix_Slice + v.data.Type.ComponentType.Kind
	return v
}

func (e *eval) slicingStr(v, leftv, rightv value, errtok lex.Token) value {
	v.lvalue = false
	v.data.Type.Id = juletype.Str
	v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	if !v.constExpr {
		return v
	}
	if rightv.constExpr {
		right := tonums(rightv.expr)
		s := v.expr.(string)
		if int(right) >= len(s) {
			e.t.pusherrtok(errtok, "overflow_limits")
		}
	}
	if leftv.constExpr && rightv.constExpr {
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
		v.constExpr = false
	}
	return v
}

// ! IMPORTANT: lex.Tokenens is should be store enumerable parentheses.
func (e *eval) enumerableParts(toks []lex.Token) [][]lex.Token {
	toks = toks[1 : len(toks)-1]
	parts, errs := ast.Parts(toks, tokens.Comma, true)
	e.t.pusherrs(errs...)
	return parts
}

func (e *eval) build_array(parts [][]lex.Token, t Type, errtok lex.Token) (value, iExpr) {
	if !t.Size.AutoSized {
		n := models.Size(len(parts))
		if n > t.Size.N {
			e.t.pusherrtok(errtok, "overflow_limits")
		} else if typeIsRef(*t.ComponentType) && n < t.Size.N {
			e.t.pusherrtok(errtok, "reference_not_initialized")
		}
	} else {
		t.Size.N = models.Size(len(parts))
		t.Size.Expr = models.Expr{
			Model: exprNode{
				value: juletype.CppId(juletype.UInt) + "(" + strconv.FormatUint(uint64(t.Size.N), 10) + ")",
			},
		}
	}
	old_type := e.type_prefix
	e.type_prefix = t.ComponentType
	defer func() { e.type_prefix = old_type }()
	var v value
	v.data.Value = t.Kind
	v.data.Type = t
	model := sliceExpr{dataType: t}
	for _, part := range parts {
		partVal, expModel := e.eval_toks(part)
		model.expr = append(model.expr, expModel)
		assign_checker{
			t:      e.t,
			expr_t:      *t.ComponentType,
			v:      partVal,
			errtok: part[0],
		}.check()
	}
	return v, model
}

func (e *eval) build_slice_implicit(parts [][]lex.Token, errtok lex.Token) (value, iExpr) {
	if len(parts) == 0 {
		e.pusherrtok(errtok, "dynamic_type_annotation_failed")
		return value{}, nil
	}
	v := value{}
	model := sliceExpr{}
	partVal, expModel := e.eval_toks(parts[0])
	model.expr = append(model.expr, expModel)
	model.dataType = Type{
		Id: juletype.Slice,
		Kind: jule.Prefix_Slice + partVal.data.Type.Kind,
	}
	model.dataType.ComponentType = new(Type)
	*model.dataType.ComponentType = partVal.data.Type
	v.data.Type = model.dataType
	for _, part := range parts[1:] {
		partVal, expModel := e.eval_toks(part)
		model.expr = append(model.expr, expModel)
		assign_checker{
			t:      e.t,
			expr_t:      *model.dataType.ComponentType,
			v:      partVal,
			errtok: part[0],
		}.check()
	}
	return v, model
}

func (e *eval) build_slice_explicit(parts [][]lex.Token, t Type, errtok lex.Token) (value, iExpr) {
	old_type := e.type_prefix
	e.type_prefix = t.ComponentType
	defer func() { e.type_prefix = old_type }()
	var v value
	v.data.Value = t.Kind
	v.data.Type = t
	model := sliceExpr{dataType: t}
	for _, part := range parts {
		partVal, expModel := e.eval_toks(part)
		model.expr = append(model.expr, expModel)
		assign_checker{
			t:      e.t,
			expr_t:      *t.ComponentType,
			v:      partVal,
			errtok: part[0],
		}.check()
	}
	return v, model
}

func (e *eval) buildMap(parts [][]lex.Token, t Type, errtok lex.Token) (value, iExpr) {
	var v value
	v.data.Value = t.Kind
	v.data.Type = t
	model := mapExpr{dataType: t}
	types := t.Tag.([]Type)
	keyType := types[0]
	valType := types[1]
	for _, part := range parts {
		brace_n := 0
		colon := -1
		for i, tok := range part {
			if tok.Id == tokens.Brace {
				switch tok.Kind {
				case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
					brace_n++
				default:
					brace_n--
				}
			}
			if brace_n != 0 {
				continue
			}
			if tok.Id == tokens.Colon {
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
			t:      e.t,
			expr_t:      keyType,
			v:      key,
			errtok: colonTok,
		}.check()
		assign_checker{
			t:      e.t,
			expr_t:      valType,
			v:      val,
			errtok: colonTok,
		}.check()
	}
	return v, model
}

func (e *eval) enumerable(exprToks []lex.Token, t Type, m *exprModel) (v value) {
	var model iExpr
	t, ok := e.t.realType(t, true)
	if !ok {
		return
	}
	errtok := exprToks[0]
	switch {
	case typeIsArray(t):
		v, model = e.build_array(e.enumerableParts(exprToks), t, errtok)
	case typeIsSlice(t):
		v, model = e.build_slice_explicit(e.enumerableParts(exprToks), t, errtok)
	case typeIsMap(t):
		v, model = e.buildMap(e.enumerableParts(exprToks), t, errtok)
	default:
		e.pusherrtok(errtok, "invalid_type_source")
		return
	}
	m.appendSubNode(model)
	return
}

func (e *eval) anonymousFn(toks []lex.Token, m *exprModel) (v value) {
	b := ast.NewBuilder(toks)
	f := b.Func(b.Tokens, false, true, false)
	b.Wait()
	if len(b.Errors) > 0 {
		e.t.pusherrs(b.Errors...)
		return
	}
	e.t.checkAnonFunc(&f)
	f.Owner = e.t
	v.data.Value = f.Id
	v.data.Type.Tag = &f
	v.data.Type.Id = juletype.Fn
	v.data.Type.Kind = f.DataTypeString()
	m.appendSubNode(anonFuncExpr{&f})
	return
}

func (e *eval) unsafeEval(toks []lex.Token, m *exprModel) (v value) {
	i := 0
	rang := ast.Range(&i, tokens.LBRACE, tokens.RBRACE, toks)
	if len(rang) == 0 {
		e.pusherrtok(toks[0], "missing_expr")
		return
	}
	old := e.allow_unsafe
	defer func() { e.allow_unsafe = old }()
	e.allow_unsafe = true
	v = e.process(rang, m)
	return v
}

func (e *eval) braceRange(toks []lex.Token, m *exprModel) (v value) {
	var exprToks []lex.Token
	brace_n := 0
	for i := len(toks) - 1; i >= 0; i-- {
		tok := toks[i]
		if tok.Id != tokens.Brace {
			continue
		}
		switch tok.Kind {
		case tokens.RBRACE, tokens.RBRACKET, tokens.RPARENTHESES:
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
			case typeIsStruct(*e.type_prefix):
				prefix := e.type_prefix
				s := e.type_prefix.Tag.(*structure)
				v = e.t.callStructConstructor(s, toks, m)
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
	case tokens.Unsafe:
		if len(toks) == 0 {
			e.pusherrtok(toks[0], "invalid_syntax")
			return
		} else if toks[1].Id != tokens.Fn {
			return e.unsafeEval(toks[1:], m)
		}
		fallthrough
	case tokens.Fn:
		return e.anonymousFn(toks, m)
	case tokens.Id, tokens.Cpp:
		return e.typeId(toks, m)
	default:
		e.pusherrtok(exprToks[0], "invalid_syntax")
	}
	return
}
