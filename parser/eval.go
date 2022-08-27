package parser

import (
	"strconv"

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
	heapMust  bool
	lvalue    bool
	variadic  bool
	isType    bool
	isField   bool
}

func isOperator(process []lex.Token) bool {
	return len(process) == 1 && process[0].Id == tokens.Operator
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

func (e *eval) toks(toks []lex.Token) (value, iExpr) {
	return e.expr(new(ast.Builder).Expr(toks))
}

func (e *eval) expr(expr Expr) (value, iExpr) {
	processes := make([][]lex.Token, len(expr.Processes))
	copy(processes, expr.Processes)
	return e.processes(processes)
}

func (e *eval) processes(processes [][]lex.Token) (v value, model iExpr) {
	defer func() {
		if typeIsVoid(v.data.Type) {
			v.data.Type.Id = juletype.Void
			v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		}
	}()
	if processes == nil || e.has_error {
		return
	}
	if len(processes) == 1 {
		m := newExprModel(processes)
		model = m
		v = e.process(processes[0], m)
		if v.constExpr {
			model = v.model
		}
		return
	}
	valProcesses := make([]any, len(processes))
	hasError := e.has_error
	for i, process := range processes {
		if isOperator(process) {
			valProcesses[i] = nil
			continue
		}
		val, model := e.p.evalToks(process)
		hasError = hasError || e.has_error || model == nil
		valProcesses[i] = []any{val, model}
	}
	if hasError {
		e.has_error = true
		return
	}
	return e.valProcesses(valProcesses, processes)
}

func (e *eval) valProcesses(exprs []any, processes [][]lex.Token) (v value, model iExpr) {
	switch len(exprs) {
	case 0:
		v.data.Type.Id = juletype.Void
		v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
		return
	case 1:
		expr := exprs[0].([]any)
		v, model = expr[0].(value), expr[1].(iExpr)
		v.lvalue = typeIsLvalue(v.data.Type)
		if v.constExpr {
			model = v.model
		}
		return
	}
	i := e.nextOperator(processes)
	process := solver{p: e.p}
	process.operator = processes[i][0]
	left := exprs[i-1].([]any)
	leftV, leftExpr := left[0].(value), left[1].(iExpr)
	right := exprs[i+1].([]any)
	rightV, rightExpr := right[0].(value), right[1].(iExpr)
	process.left = processes[i-1]
	process.left_val = leftV
	process.right = processes[i+1]
	process.right_val = rightV
	val := process.solve()
	var expr iExpr
	if val.constExpr {
		expr = val.model
	} else {
		sexpr := serieExpr{}
		// If processes has one more couple (see: [EXPR] [OPERATOR] [EXPR])
		if len(processes) > 3 {
			sexpr.exprs = make([]any, 5)
			sexpr.exprs[0] = exprNode{tokens.LPARENTHESES}
			sexpr.exprs[1] = leftExpr
			sexpr.exprs[2] = exprNode{process.operator.Kind}
			sexpr.exprs[3] = rightExpr
			sexpr.exprs[4] = exprNode{tokens.RPARENTHESES}
		} else {
			sexpr.exprs = make([]any, 3)
			sexpr.exprs[0] = leftExpr
			sexpr.exprs[1] = exprNode{process.operator.Kind}
			sexpr.exprs[2] = rightExpr
		}
		expr = sexpr
	}
	processes = append(processes[:i-1], append([][]lex.Token{{}}, processes[i+2:]...)...)
	exprs = append(exprs[:i-1], append([]any{[]any{val, expr}}, exprs[i+2:]...)...)
	return e.valProcesses(exprs, processes)
}

// nextOperator find index of priority operator and returns index of operator
// if found, returns -1 if not.
func (e *eval) nextOperator(processes [][]lex.Token) int {
	prec := precedencer{}
	for i, process := range processes {
		switch {
		case !isOperator(process),
			processes[i-1] == nil && processes[i+1] == nil:
			continue
		}
		switch process[0].Kind {
		case tokens.STAR, tokens.PERCENT, tokens.SOLIDUS,
			tokens.RSHIFT, tokens.LSHIFT, tokens.AMPER:
			prec.set(5, i)
		case tokens.PLUS, tokens.MINUS, tokens.VLINE, tokens.CARET:
			prec.set(4, i)
		case tokens.EQUALS, tokens.NOT_EQUALS, tokens.LESS,
			tokens.LESS_EQUAL, tokens.GREAT, tokens.GREAT_EQUAL:
			prec.set(3, i)
		case tokens.DOUBLE_AMPER:
			prec.set(2, i)
		case tokens.DOUBLE_VLINE:
			prec.set(1, i)
		default:
			e.pusherrtok(process[0], "invalid_operator")
		}
	}
	data := prec.get()
	if data == nil {
		return -1
	}
	return data.(int)
}

func (e *eval) single(tok lex.Token, m *exprModel) (v value, ok bool) {
	eval := valueEvaluator{tok, m, e.p}
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
	processor := unary{toks[0], exprToks, m, e.p}
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
	val, model := e.toks(toks)
	m.appendSubNode(model)
	return val
}

func (e *eval) dataTypeFunc(expr lex.Token, callRange []lex.Token, m *exprModel) (v value, isret bool) {
	switch expr.Id {
	case tokens.DataType:
		switch expr.Kind {
		case tokens.STR:
			m.appendSubNode(exprNode{"__julec_tostr("})
			_, vm := e.p.evalToks(callRange)
			m.appendSubNode(vm)
			m.appendSubNode(exprNode{tokens.RPARENTHESES})
			v.data.Type = Type{
				Id:   juletype.Fn,
				Kind: strDefaultFunc.DataTypeString(),
				Tag:  strDefaultFunc,
			}
			isret = true

		default:
			dt := Type{
				Token:  expr,
				Id:   juletype.TypeFromId(expr.Kind),
				Kind: expr.Kind,
			}
			isret = true
			v = e.castExpr(dt, callRange, m, expr)
		}
	case tokens.Id:
		def, _, _ := e.p.defById(expr.Kind)
		if def == nil {
			break
		}
		switch t := def.(type) {
		case *TypeAlias:
			dt, ok := e.p.realType(t.Type, true)
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

func (e *eval) callCppLink(data callData, m *exprModel) (v value) {
	v.data.Type.Id = juletype.Void
	v.data.Type.Kind = juletype.TypeMap[v.data.Type.Id]
	tok := data.expr[0]
	data.expr = data.expr[1:] // Remove cpp keyword
	if len(data.expr) == 0 {
		e.pusherrtok(tok, "invalid_syntax")
		return
	}
	tok = data.expr[0]
	if tok.Id != tokens.Dot {
		e.pusherrtok(tok, "invalid_syntax")
		return
	}
	data.expr = data.expr[1:] // Remove dot keyword
	if len(data.expr) == 0 {
		e.pusherrtok(tok, "invalid_syntax")
		return
	}
	tok = data.expr[0]
	if tok.Id != tokens.Id {
		e.pusherrtok(tok, "invalid_syntax")
		return
	}
	link := e.p.linkById(tok.Kind)
	if link == nil {
		e.pusherrtok(tok, "id_not_exist", tok.Kind)
		return
	}
	m.appendSubNode(exprNode{link.Link.Id})
	return e.callFunc(link.Link, data, m)
}

func (e *eval) unsafe_allowed() bool {
	return e.allow_unsafe ||
		(e.p.rootBlock != nil && e.p.rootBlock.IsUnsafe) ||
		(e.p.nodeBlock != nil && e.p.nodeBlock.IsUnsafe)
}

func (e *eval) callFunc(f *Func, data callData, m *exprModel) value {
	if !e.unsafe_allowed() && f.IsUnsafe {
		e.pusherrtok(data.expr[0], "unsafe_behavior_at_out_of_unsafe_scope")
	}
	if f.BuiltinCaller != nil {
		return f.BuiltinCaller.(BuiltinCaller)(e.p, f, data, m)
	}
	return e.p.callFunc(f, data, m)
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
	case tokens.Cpp:
		return e.callCppLink(data, m)
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
		return e.callFunc(f, data, m)
	}
	e.pusherrtok(data.expr[len(data.expr)-1], "invalid_syntax")
	return
}

func (e *eval) process(toks []lex.Token, m *exprModel) (v value) {
	defer func() {
		if typeIsVoid(v.data.Type) {
			v.data.Type.Kind = juletype.TypeMap[juletype.Void]
			v.constExpr = false
		}
	}()
	v.constExpr = true
	if len(toks) == 1 {
		v, _ = e.single(toks[0], m)
		return
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
		tok := dotTok
		tok.Id = tokens.Self
		tok.Kind = tokens.SELF
		toks = []lex.Token{tok}
	case 1:
		tok := toks[0]
		if tok.Id == tokens.DataType {
			return e.typeSubId(tok, idTok, m)
		} else if tok.Id == tokens.Id {
			t, _, _ := e.p.typeById(tok.Kind)
			if t != nil {
				return e.typeSubId(t.Type.Token, idTok, m)
			}
		}
	}
	val := e.process(toks, m)
	checkType := val.data.Type
	if typeIsExplicitPtr(checkType) {
		checkType = unptrType(checkType)
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

func (e *eval) castExpr(dt Type, exprToks []lex.Token, m *exprModel, errTok lex.Token) value {
	val, model := e.toks(exprToks)
	m.appendSubNode(exprNode{tokens.LPARENTHESES})
	m.appendSubNode(exprNode{tokens.LPARENTHESES + dt.String() + tokens.RPARENTHESES})
	m.appendSubNode(exprNode{tokens.LPARENTHESES})
	m.appendSubNode(model)
	m.appendSubNode(exprNode{tokens.RPARENTHESES})
	m.appendSubNode(exprNode{tokens.RPARENTHESES})
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
		if tok.Id != tokens.Brace || tok.Kind != tokens.LPARENTHESES {
			return
		}
		exprToks, ok = e.p.getrange(tokens.LPARENTHESES, tokens.RPARENTHESES, exprToks)
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
	if typeIsStruct(unptrType(t)) {
		e.castStruct(t, v, errtok)
		return
	} else if !e.unsafe_allowed() {
		e.pusherrtok(errtok, "unsafe_behavior_at_out_of_unsafe_scope")
		return
	}
	if !typeIsPtr(v.data.Type) &&
		!typeIsPure(v.data.Type) &&
		!juletype.IsInteger(v.data.Type.Id) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", v.data.Type.Kind, t.Kind)
	}
}

func (e *eval) castPure(t Type, v *value, errtok lex.Token) {
	switch t.Id {
	case juletype.Any:
		return
	case juletype.Str:
		e.castStr(v.data.Type, errtok)
		return
	case juletype.Enum:
		e.castEnum(t, v, errtok)
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
	if !typeIsSlice(t) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", juletype.TypeMap[juletype.Str], t.Kind)
		return
	}
	t = *t.ComponentType
	if !typeIsPure(t) || (t.Id != juletype.U8 && t.Id != juletype.I32) {
		e.pusherrtok(errtok, "type_not_supports_casting_to", juletype.TypeMap[juletype.Str], t.Kind)
	}
}

func (e *eval) castEnum(t Type, v *value, errtok lex.Token) {
	enum := t.Tag.(*Enum)
	t = enum.Type
	t.Kind = enum.Id
	e.castNumeric(t, v, errtok)
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
		e.p.pusherrs(b.Errors...)
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
	if typeIsPure(t) && typeIsStruct(t) {
		if toks[0].Id != tokens.Brace || toks[0].Kind != tokens.LBRACE {
			e.pusherrtok(toks[0], "invalid_syntax")
			return
		}
		s := t.Tag.(*structure)
		return e.p.callStructConstructor(s, toks, m)
	}
	if toks[0].Id != tokens.Brace || toks[0].Kind != tokens.LBRACKET {
		e.pusherrtok(toks[0], "invalid_syntax")
		return
	}
	return e.enumerable(toks, t, m)
}

func (e *eval) xObjSubId(dm *DefineMap, val value, idTok lex.Token, m *exprModel) (v value) {
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
		v.constExpr = g.Const
		if g.Const {
			v.expr = g.ExprTag
			v.model = g.Expr.Model
		}
		v.isField = true
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
	v := e.xObjSubId(strDefines, val, idTok, m)
	v.lvalue = false
	return v
}

func (e *eval) sliceObjSubId(val value, idTok lex.Token, m *exprModel) value {
	v := e.xObjSubId(sliceDefines, val, idTok, m)
	v.lvalue = false
	return v
}

func (e *eval) arrayObjSubId(val value, idTok lex.Token, m *exprModel) value {
	v := e.xObjSubId(arrayDefines, val, idTok, m)
	v.lvalue = false
	return v
}

func (e *eval) mapObjSubId(val value, idTok lex.Token, m *exprModel) value {
	readyMapDefines(val.data.Type)
	v := e.xObjSubId(mapDefines, val, idTok, m)
	v.lvalue = false
	return v
}

func (e *eval) enumSubId(val value, idTok lex.Token, m *exprModel) (v value) {
	enum := val.data.Type.Tag.(*Enum)
	v = val
	v.data.Type = enum.Type
	v.data.Type.Token = enum.Tok
	v.lvalue = false
	v.isType = false
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
	s := val.data.Type.Tag.(*structure)
	val.constExpr = false
	val.isType = false
	val = e.xObjSubId(s.Defines, val, idTok, m)
	return val
}

func (e *eval) traitObjSubId(val value, idTok lex.Token, m *exprModel) value {
	m.appendSubNode(exprNode{".get()"})
	t := val.data.Type.Tag.(*trait)
	val.constExpr = false
	val.lvalue = false
	val.isType = false
	val = e.xObjSubId(t.Defines, val, idTok, m)
	val.constExpr = false
	return val
}

type nsFind interface {
	nsById(string) *namespace
}

func (e *eval) getNs(toks *[]lex.Token) *DefineMap {
	var prev nsFind = e.p
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
	if len(exprToks) == 0 || brace_n > 0 {
		e.pusherrtok(errTok, "invalid_syntax")
		return
	}
	var model iExpr
	v, model = e.toks(exprToks)
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
			leftv, model = e.p.evalToks(leftToks)
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
			rightv, model = e.p.evalToks(rightToks)
			m.appendSubNode(indexingExprModel(model))
			e.checkIntegerIndexing(rightv, errTok)
		}
		m.appendSubNode(exprNode{")"})
		return e.slicing(v, leftv, rightv, errTok)
	}
	m.appendSubNode(exprNode{tokens.LBRACKET})
	indexv, model := e.toks(toks[1 : len(toks)-1])
	m.appendSubNode(indexingExprModel(model))
	m.appendSubNode(exprNode{tokens.RBRACKET})
	v = e.indexing(v, indexv, errTok)
	// Ignore indexed type from original
	v.data.Type.Pure = true
	v.data.Type.Original = nil
	return v
}

func (e *eval) checkIntegerIndexing(v value, errtok lex.Token) {
	switch {
	case !typeIsPure(v.data.Type):
		e.pusherrtok(errtok, "invalid_expr")
	case !juletype.IsInteger(v.data.Type.Id):
		e.pusherrtok(errtok, "invalid_expr")
	}
}

func (e *eval) indexing(enumv, indexv value, errtok lex.Token) (v value) {
	switch {
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
	if index.constExpr && tonums(index.expr) < 0 {
		e.p.pusherrtok(index.data.Token, "invalid_expr")
	}
	return slicev
}

func (e *eval) indexingArray(arrv, index value, errtok lex.Token) value {
	arrv.data.Type = *arrv.data.Type.ComponentType
	e.checkIntegerIndexing(index, errtok)
	if index.constExpr && tonums(index.expr) < 0 {
		e.p.pusherrtok(index.data.Token, "invalid_expr")
	}
	return arrv
}

func (e *eval) indexingMap(mapv, leftv value, errtok lex.Token) value {
	types := mapv.data.Type.Tag.([]Type)
	keyType := types[0]
	valType := types[1]
	mapv.data.Type = valType
	e.p.checkType(keyType, leftv.data.Type, false, errtok)
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
	i := tonums(index.expr)
	if i < 0 {
		e.p.pusherrtok(errtok, "overflow_limits")
	} else if strv.constExpr {
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
	if leftv.constExpr && tonums(leftv.expr) < 0 {
		e.p.pusherrtok(errtok, "overflow_limits")
	}
	if rightv.constExpr && tonums(rightv.expr) < 0 {
		e.p.pusherrtok(errtok, "overflow_limits")
	}
	if leftv.constExpr && rightv.constExpr && rightv.expr != nil {
		if tonums(leftv.expr) > tonums(rightv.expr) {
			e.p.pusherrtok(errtok, "overflow_limits")
		}
	}
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
			e.p.pusherrtok(errtok, "overflow_limits")
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

//! IMPORTANT: lex.Tokenens is should be store enumerable parentheses.
func (e *eval) enumerableParts(toks []lex.Token) [][]lex.Token {
	toks = toks[1 : len(toks)-1]
	parts, errs := ast.Parts(toks, tokens.Comma, true)
	e.p.pusherrs(errs...)
	return parts
}

func (e *eval) buildArray(parts [][]lex.Token, t Type, errtok lex.Token) (value, iExpr) {
	if !t.Size.AutoSized {
		if models.Size(len(parts)) > t.Size.N {
			e.p.pusherrtok(errtok, "overflow_limits")
		}
	} else {
		t.Size.N = models.Size(len(parts))
		t.Size.Expr = models.Expr{
			Model: exprNode{
				value: juletype.TypeMap[juletype.UInt] + "{" + strconv.FormatUint(uint64(t.Size.N), 10) + "}",
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
		partVal, expModel := e.toks(part)
		model.expr = append(model.expr, expModel)
		assignChecker{
			p:      e.p,
			t:      *t.ComponentType,
			v:      partVal,
			errtok: part[0],
		}.checkAssignType()
	}
	return v, model
}

func (e *eval) buildSlice(parts [][]lex.Token, t Type, errtok lex.Token) (value, iExpr) {
	old_type := e.type_prefix
	e.type_prefix = t.ComponentType
	defer func() { e.type_prefix = old_type }()
	var v value
	v.data.Value = t.Kind
	v.data.Type = t
	model := sliceExpr{dataType: t}
	for _, part := range parts {
		partVal, expModel := e.toks(part)
		model.expr = append(model.expr, expModel)
		assignChecker{
			p:      e.p,
			t:      *t.ComponentType,
			v:      partVal,
			errtok: part[0],
		}.checkAssignType()
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
		key, keyModel := e.toks(keyToks)
		model.keyExprs = append(model.keyExprs, keyModel)
		val, valModel := e.toks(valToks)
		model.valExprs = append(model.valExprs, valModel)
		assignChecker{
			p:      e.p,
			t:      keyType,
			v:      key,
			errtok: colonTok,
		}.checkAssignType()
		assignChecker{
			p:      e.p,
			t:      valType,
			v:      val,
			errtok: colonTok,
		}.checkAssignType()
	}
	return v, model
}

func (e *eval) enumerable(exprToks []lex.Token, t Type, m *exprModel) (v value) {
	var model iExpr
	t, ok := e.p.realType(t, true)
	if !ok {
		return
	}
	switch {
	case typeIsArray(t):
		v, model = e.buildArray(e.enumerableParts(exprToks), t, exprToks[0])
	case typeIsSlice(t):
		v, model = e.buildSlice(e.enumerableParts(exprToks), t, exprToks[0])
	case typeIsMap(t):
		v, model = e.buildMap(e.enumerableParts(exprToks), t, exprToks[0])
	default:
		e.pusherrtok(exprToks[0], "invalid_type_source")
		return
	}
	m.appendSubNode(model)
	return
}

func (e *eval) anonymousFn(toks []lex.Token, m *exprModel) (v value) {
	b := ast.NewBuilder(toks)
	f := b.Func(b.Tokens, true, false)
	b.Wait()
	if len(b.Errors) > 0 {
		e.p.pusherrs(b.Errors...)
		return
	}
	e.p.checkAnonFunc(&f)
	f.Owner = e.p
	v.data.Value = f.Id
	v.data.Type.Tag = &f
	v.data.Type.Id = juletype.Fn
	v.data.Type.Kind = f.DataTypeString()
	m.appendSubNode(anonFuncExpr{&f, e.p.blockVars})
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
			case typeIsArray(*e.type_prefix) || typeIsSlice(*e.type_prefix):
				return e.enumerable(toks, *e.type_prefix, m)
			case typeIsStruct(*e.type_prefix):
				s := e.type_prefix.Tag.(*structure)
				return e.p.callStructConstructor(s, toks, m)
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
	case tokens.Id:
		return e.typeId(toks, m)
	case tokens.Brace:
		switch exprToks[0].Kind {
		case tokens.LBRACKET:
			b := ast.NewBuilder(nil)
			i := 0
			t, ok := b.DataType(exprToks, &i, true, true)
			b.Wait()
			if !ok {
				e.p.pusherrs(b.Errors...)
				return
			} else if i+1 < len(exprToks) {
				e.pusherrtok(toks[i+1], "invalid_syntax")
			}
			exprToks = toks[len(exprToks):]
			return e.enumerable(exprToks, t, m)
		default:
			e.pusherrtok(exprToks[0], "invalid_syntax")
		}
	default:
		e.pusherrtok(exprToks[0], "invalid_syntax")
	}
	return
}
