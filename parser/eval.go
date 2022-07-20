package parser

import (
	"strconv"

	"github.com/the-xlang/xxc/ast"
	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xlog"
	"github.com/the-xlang/xxc/pkg/xtype"
)

func isOperator(process Toks) bool {
	return len(process) == 1 && process[0].Id == tokens.Operator
}

type eval struct {
	p        *Parser
	hasError bool
}

func (e *eval) pusherrtok(tok Tok, err string, args ...any) {
	if e.hasError {
		return
	}
	e.hasError = true
	e.p.pusherrtok(tok, err, args...)
}

func (e *eval) pusherrs(errs ...xlog.CompilerLog) {
	if e.hasError {
		return
	}
	e.hasError = true
	e.p.pusherrs(errs...)
}

func (e *eval) toks(toks Toks) (value, iExpr) {
	return e.expr(new(ast.Builder).Expr(toks))
}

func (e *eval) expr(expr Expr) (value, iExpr) {
	processes := make([]Toks, len(expr.Processes))
	copy(processes, expr.Processes)
	return e.processes(processes)
}

func (e *eval) processes(processes []Toks) (v value, expr iExpr) {
	defer func() {
		if typeIsVoid(v.data.Type) {
			v.data.Type.Id = xtype.Void
			v.data.Type.Kind = xtype.VoidTypeStr
		}
	}()
	if processes == nil || e.hasError {
		return
	}
	if len(processes) == 1 {
		m := newExprModel(processes)
		expr = m
		v = e.process(processes[0], m)
		return
	}
	valProcesses := make([]any, len(processes))
	hasError := e.hasError
	for i, process := range processes {
		if isOperator(process) {
			valProcesses[i] = nil
			continue
		}
		val, model := e.p.evalToks(process)
		hasError = hasError || e.hasError
		valProcesses[i] = []any{val.data, model}
	}
	if hasError {
		e.hasError = true
		return
	}
	return e.valProcesses(valProcesses, processes)
}

func (e *eval) valProcesses(exprs []any, processes []Toks) (v value, model iExpr) {
	switch len(exprs) {
	case 0:
		v.data.Type.Id = xtype.Void
		v.data.Type.Kind = xtype.VoidTypeStr
		return
	case 1:
		expr := exprs[0].([]any)
		v.data, model = expr[0].(models.Data), expr[1].(iExpr)
		v.lvalue = typeIsLvalue(v.data.Type)
		return
	}
	i := e.nextOperator(processes)
	process := solver{p: e.p}
	process.operator = processes[i][0]
	left := exprs[i-1].([]any)
	leftV, leftExpr := left[0].(models.Data), left[1].(iExpr)
	right := exprs[i+1].([]any)
	rightV, rightExpr := right[0].(models.Data), right[1].(iExpr)
	process.left = processes[i-1]
	process.leftVal = leftV
	process.right = processes[i+1]
	process.rightVal = rightV
	val := process.solve()
	expr := serieExpr{}
	expr.exprs = make([]any, 5)
	expr.exprs[0] = exprNode{tokens.LPARENTHESES}
	expr.exprs[1] = leftExpr
	expr.exprs[2] = exprNode{process.operator.Kind}
	expr.exprs[3] = rightExpr
	expr.exprs[4] = exprNode{tokens.RPARENTHESES}
	processes = append(processes[:i-1], append([]Toks{{}}, processes[i+2:]...)...)
	exprs = append(exprs[:i-1], append([]any{[]any{val, expr}}, exprs[i+2:]...)...)
	return e.valProcesses(exprs, processes)
}

// nextOperator find index of priority operator and returns index of operator
// if found, returns -1 if not.
func (e *eval) nextOperator(processes []Toks) int {
	prec := precedencer{}
	for i, process := range processes {
		switch {
		case !isOperator(process),
			processes[i-1] == nil && processes[i+1] == nil:
			continue
		}
		switch process[0].Kind {
		case tokens.LSHIFT, tokens.RSHIFT:
			prec.set(1, i)
		case tokens.STAR, tokens.SOLIDUS, tokens.PERCENT:
			prec.set(2, i)
		case tokens.AMPER:
			prec.set(3, i)
		case tokens.CARET:
			prec.set(4, i)
		case tokens.VLINE:
			prec.set(5, i)
		case tokens.PLUS, tokens.MINUS:
			prec.set(6, i)
		case tokens.LESS, tokens.LESS_EQUAL,
			tokens.GREAT, tokens.GREAT_EQUAL:
			prec.set(7, i)
		case tokens.EQUALS, tokens.NOT_EQUALS:
			prec.set(8, i)
		case tokens.AND:
			prec.set(9, i)
		case tokens.OR:
			prec.set(10, i)
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

func (e *eval) single(tok Tok, m *exprModel) (v value, ok bool) {
	eval := valueEvaluator{tok, m, e.p}
	v.data.Type.Id = xtype.Void
	v.data.Type.Kind = xtype.VoidTypeStr
	v.data.Tok = tok
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

func (e *eval) unary(toks Toks, m *exprModel) value {
	var v value
	//? Length is 1 cause all length of operator tokens is 1.
	//? Change "1" with length of token's value
	//? if all operators length is not 1.
	exprToks := toks[1:]
	processor := unary{toks[0], exprToks, m, e.p}
	m.appendSubNode(exprNode{processor.tok.Kind})
	if processor.toks == nil {
		e.pusherrtok(processor.tok, "invalid_syntax")
		return v
	}
	switch processor.tok.Kind {
	case tokens.MINUS:
		v = processor.minus()
	case tokens.PLUS:
		v = processor.plus()
	case tokens.TILDE:
		v = processor.tilde()
	case tokens.EXCLAMATION:
		v = processor.logicalNot()
	case tokens.STAR:
		v = processor.star()
	case tokens.AMPER:
		v = processor.amper()
	default:
		e.pusherrtok(processor.tok, "invalid_syntax")
	}
	v.data.Tok = processor.tok
	return v
}

func (e *eval) betweenParentheses(toks Toks, m *exprModel) value {
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

func (e *eval) dataTypeFunc(expr Tok, callRange Toks, m *exprModel) (v value, isret bool) {
	switch expr.Id {
	case tokens.DataType:
		switch expr.Kind {
		case tokens.STR:
			m.appendSubNode(exprNode{"tostr"})
			// Val: "()" for accept DataType as function.
			v.data.Type = DataType{Id: xtype.Func, Kind: "()", Tag: strDefaultFunc}
			isret = true
		default:
			dt := DataType{
				Tok:  expr,
				Id:   xtype.TypeFromId(expr.Kind),
				Kind: expr.Kind,
			}
			isret = true
			v = e.castExpr(dt, callRange, m, expr)
		}
	case tokens.Id:
		def, _, _, _ := e.p.defById(expr.Kind)
		if def == nil {
			break
		}
		switch t := def.(type) {
		case *Type:
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

func (e *eval) parenthesesRange(toks Toks, m *exprModel) (v value) {
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
			val, ok = e.tryAssign(toks, m)
			if ok {
				v = val
				return
			}
		}
	}
	exprToks, rangeToks := ast.RangeLast(toks)
	if len(exprToks) == 0 {
		return e.betweenParentheses(rangeToks, m)
	}
	// Below is call expression
	var genericsToks Toks
	if tok := exprToks[len(exprToks)-1]; tok.Id == tokens.Brace && tok.Kind == tokens.RBRACKET {
		exprToks, genericsToks = ast.RangeLast(exprToks)
	}
	switch tok := exprToks[0]; tok.Id {
	case tokens.DataType, tokens.Id:
		if len(exprToks) == 1 && len(genericsToks) == 0 {
			v, isret := e.dataTypeFunc(exprToks[0], rangeToks, m)
			if isret {
				return v
			}
		}
		fallthrough
	default:
		v = e.process(exprToks, m)
	}
	switch {
	case typeIsFunc(v.data.Type):
		f := v.data.Type.Tag.(*Func)
		return e.p.callFunc(f, genericsToks, rangeToks, m)
	case valIsStructType(v):
		s := v.data.Type.Tag.(*xstruct)
		return e.p.callStructConstructor(s, genericsToks, rangeToks, m)
	}
	e.pusherrtok(exprToks[len(exprToks)-1], "invalid_syntax")
	return
}

func (e *eval) process(toks Toks, m *exprModel) (v value) {
	defer func() {
		if typeIsVoid(v.data.Type) {
			v.data.Type.Kind = xtype.VoidTypeStr
		}
	}()
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

func (e *eval) subId(toks Toks, m *exprModel) (v value) {
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
		toks = Toks{tok}
	case 1:
		tok := toks[0]
		if tok.Id == tokens.DataType {
			return e.typeSubId(tok, idTok, m)
		} else if tok.Id == tokens.Id {
			t, _, _ := e.p.typeById(tok.Kind)
			if t != nil {
				return e.typeSubId(t.Type.Tok, idTok, m)
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
		case checkType.Id == xtype.Str:
			return e.strObjSubId(val, idTok, m)
		case valIsEnumType(val):
			return e.enumSubId(val, idTok, m)
		case valIsStructIns(val):
			return e.structObjSubId(val, idTok, m)
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

func (e *eval) castExpr(dt DataType, exprToks Toks, m *exprModel, errTok Tok) value {
	val, model := e.toks(exprToks)
	m.appendSubNode(exprNode{tokens.LPARENTHESES + dt.String() + tokens.RPARENTHESES})
	m.appendSubNode(exprNode{tokens.LPARENTHESES})
	m.appendSubNode(model)
	m.appendSubNode(exprNode{tokens.RPARENTHESES})
	val = e.cast(val, dt, errTok)
	return val
}

func (e *eval) tryCast(toks Toks, m *exprModel) (v value, _ bool) {
	braceCount := 0
	errTok := toks[0]
	for i, tok := range toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
				continue
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
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
		exprToks, ok = e.p.getRange(tokens.LPARENTHESES, tokens.RPARENTHESES, exprToks)
		if !ok {
			return
		}
		val := e.castExpr(dt, exprToks, m, errTok)
		return val, true
	}
	return
}

func (e *eval) cast(v value, t DataType, errtok Tok) value {
	switch {
	case typeIsPtr(t):
		e.castPtr(v.data.Type, errtok)
	case typeIsSlice(t):
		e.castSlice(t, v.data.Type, errtok)
	case typeIsPure(t):
		v.lvalue = false
		e.castSingle(t, v.data.Type, errtok)
	default:
		e.pusherrtok(errtok, "type_notsupports_casting", t.Kind)
	}
	v.data.Value = t.Kind
	v.data.Type = t
	v.constant = false
	v.volatile = false
	return v
}

func (e *eval) castSingle(t, vt DataType, errtok Tok) {
	switch t.Id {
	case xtype.Any:
		return
	case xtype.Str:
		e.castStr(vt, errtok)
		return
	case xtype.Enum:
		e.castEnum(t, vt, errtok)
		return
	}
	switch {
	case xtype.IsIntegerType(t.Id):
		e.castInteger(t, vt, errtok)
	case xtype.IsNumericType(t.Id):
		e.castNumeric(t, vt, errtok)
	default:
		e.pusherrtok(errtok, "type_notsupports_casting", t.Kind)
	}
}

func (e *eval) castStr(vt DataType, errtok Tok) {
	if !typeIsSlice(vt) {
		e.pusherrtok(errtok, "type_notsupports_casting", vt.Kind)
		return
	}
	vt = typeOfSliceComponents(vt)
	if !typeIsPure(vt) || vt.Id != xtype.U8 {
		e.pusherrtok(errtok, "type_notsupports_casting", vt.Kind)
	}
}

func (e *eval) castEnum(t, vt DataType, errtok Tok) {
	enum := t.Tag.(*Enum)
	t = enum.Type
	t.Kind = enum.Id
	e.castNumeric(t, vt, errtok)
}

func (e *eval) castInteger(t, vt DataType, errtok Tok) {
	if typeIsPtr(vt) &&
		(t.Id == xtype.I64 || t.Id == xtype.U64 ||
			t.Id == xtype.Intptr || t.Id == xtype.UIntptr) {
		return
	}
	if typeIsPure(vt) && xtype.IsNumericType(vt.Id) {
		return
	}
	e.pusherrtok(errtok, "type_notsupports_casting_to", vt.Kind, t.Kind)
}

func (e *eval) castNumeric(t, vt DataType, errtok Tok) {
	if typeIsPure(vt) && xtype.IsNumericType(vt.Id) {
		return
	}
	e.pusherrtok(errtok, "type_notsupports_casting_to", vt.Kind, t.Kind)
}

func (e *eval) castPtr(vt DataType, errtok Tok) {
	if typeIsPtr(vt) {
		return
	}
	if typeIsPure(vt) && xtype.IsIntegerType(vt.Id) {
		return
	}
	e.pusherrtok(errtok, "type_notsupports_casting", vt.Kind)
}

func (e *eval) castSlice(t, vt DataType, errtok Tok) {
	if !typeIsPure(vt) || vt.Id != xtype.Str {
		e.pusherrtok(errtok, "type_notsupports_casting", vt.Kind)
		return
	}
	t = typeOfSliceComponents(t)
	if !typeIsPure(t) || t.Id != xtype.U8 {
		e.pusherrtok(errtok, "type_notsupports_casting", vt.Kind)
	}
}

func (e *eval) tryAssign(toks Toks, m *exprModel) (v value, ok bool) {
	b := ast.NewBuilder(nil)
	toks = toks[1 : len(toks)-1] // Remove first-last parentheses
	assign, ok := b.AssignExpr(toks, true)
	if !ok {
		return
	}
	ok = true
	if len(b.Errors) > 0 {
		e.pusherrs(b.Errors...)
		return
	}
	v, _ = e.expr(assign.Left[0].Expr)
	if v.lvalue && ast.IsSuffixOperator(assign.Setter.Kind) {
		v.lvalue = false
	}
	e.p.assign(&assign)
	m.appendSubNode(assignExpr{assign})
	return
}

func (e *eval) xTypeSubId(dm *Defmap, idTok Tok, m *exprModel) (v value) {
	i, dm, t := dm.defById(idTok.Kind, nil)
	if i == -1 {
		e.pusherrtok(idTok, "obj_have_not_id", idTok.Kind)
		return
	}
	v.lvalue = false
	v.data.Value = idTok.Kind
	switch t {
	case 'g':
		g := dm.Globals[i]
		m.appendSubNode(exprNode{g.Tag.(string)})
		v.data.Type = g.Type
		v.constant = g.Const
	}
	return
}

func (e *eval) i8SubId(idTok Tok, m *exprModel) value {
	return e.xTypeSubId(i8statics, idTok, m)
}

func (e *eval) i16SubId(idTok Tok, m *exprModel) value {
	return e.xTypeSubId(i16statics, idTok, m)
}

func (e *eval) i32SubId(idTok Tok, m *exprModel) value {
	return e.xTypeSubId(i32statics, idTok, m)
}

func (e *eval) i64SubId(idTok Tok, m *exprModel) value {
	return e.xTypeSubId(i64statics, idTok, m)
}

func (e *eval) u8SubId(idTok Tok, m *exprModel) value {
	return e.xTypeSubId(u8statics, idTok, m)
}

func (e *eval) u16SubId(idTok Tok, m *exprModel) value {
	return e.xTypeSubId(u16statics, idTok, m)
}

func (e *eval) u32SubId(idTok Tok, m *exprModel) value {
	return e.xTypeSubId(u32statics, idTok, m)
}

func (e *eval) u64SubId(idTok Tok, m *exprModel) value {
	return e.xTypeSubId(u64statics, idTok, m)
}

func (e *eval) uintSubId(idTok Tok, m *exprModel) value {
	return e.xTypeSubId(uintStatics, idTok, m)
}

func (e *eval) intSubId(idTok Tok, m *exprModel) value {
	return e.xTypeSubId(intStatics, idTok, m)
}

func (e *eval) f32SubId(idTok Tok, m *exprModel) value {
	return e.xTypeSubId(f32statics, idTok, m)
}

func (e *eval) f64SubId(idTok Tok, m *exprModel) value {
	return e.xTypeSubId(f64statics, idTok, m)
}

func (e *eval) strSubId(idTok Tok, m *exprModel) value {
	return e.xTypeSubId(strStatics, idTok, m)
}

func (e *eval) typeSubId(typeTok, idTok Tok, m *exprModel) (v value) {
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
	case tokens.STR:
		return e.strSubId(idTok, m)
	}
	e.pusherrtok(typeTok, "obj_not_support_sub_fields", typeTok.Kind)
	return
}

func (e *eval) typeId(toks Toks, m *exprModel) (v value) {
	tok := toks[0]
	t, _, _ := e.p.typeById(tok.Kind)
	if t == nil {
		v.data.Type.Id = xtype.Void
		v.data.Type.Kind = xtype.VoidTypeStr
		return
	}
	toks = toks[1:]
	return e.enumerable(toks, t.Type, m)
}

func (e *eval) xObjSubId(dm *Defmap, val value, idTok Tok, m *exprModel) (v value) {
	i, dm, t := dm.defById(idTok.Kind, idTok.File)
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
		if g.Tag == nil {
			m.appendSubNode(exprNode{xapi.OutId(g.Id, g.DefTok.File)})
		} else {
			m.appendSubNode(exprNode{g.Tag.(string)})
		}
		v.data.Type = g.Type
		v.lvalue = true
		v.constant = g.Const
	case 'f':
		f := dm.Funcs[i]
		f.used = true
		v.data.Type.Id = xtype.Func
		v.data.Type.Tag = f.Ast
		v.data.Type.Kind = f.Ast.DataTypeString()
		v.data.Tok = f.Ast.Tok
		m.appendSubNode(exprNode{f.Ast.Id})
	}
	return
}

func (e *eval) strObjSubId(val value, idTok Tok, m *exprModel) value {
	return e.xObjSubId(strDefs, val, idTok, m)
}

func (e *eval) sliceObjSubId(val value, idTok Tok, m *exprModel) value {
	readySliceDefs(val.data.Type)
	return e.xObjSubId(sliceDefs, val, idTok, m)
}

func (e *eval) arrayObjSubId(val value, idTok Tok, m *exprModel) value {
	return e.xObjSubId(arrayDefs, val, idTok, m)
}

func (e *eval) mapObjSubId(val value, idTok Tok, m *exprModel) value {
	readyMapDefs(val.data.Type)
	return e.xObjSubId(mapDefs, val, idTok, m)
}

func (e *eval) enumSubId(val value, idTok Tok, m *exprModel) (v value) {
	enum := val.data.Type.Tag.(*Enum)
	v = val
	v.data.Type.Tok = enum.Tok
	v.constant = true
	v.lvalue = false
	v.isType = false
	m.appendSubNode(exprNode{"::"})
	m.appendSubNode(exprNode{xapi.OutId(idTok.Kind, enum.Tok.File)})
	if enum.ItemById(idTok.Kind) == nil {
		e.pusherrtok(idTok, "obj_have_not_id", idTok.Kind)
	}
	return
}

func (e *eval) structObjSubId(val value, idTok Tok, m *exprModel) value {
	s := val.data.Type.Tag.(*xstruct)
	val.constant = false
	val.lvalue = false
	val.isType = false
	return e.xObjSubId(s.Defs, val, idTok, m)
}

type nsFind interface {
	nsById(string, bool) *namespace
}

func (e *eval) nsSubId(toks Toks, m *exprModel) (v value) {
	var prev nsFind = e.p
	for i, tok := range toks {
		if (i+1)%2 != 0 {
			if tok.Id != tokens.Id {
				e.pusherrtok(tok, "invalid_syntax")
				continue
			}
			src := prev.nsById(tok.Kind, false)
			if src == nil {
				if i > 0 {
					toks = toks[i:]
					goto eval
				}
				e.pusherrtok(tok, "namespace_not_exist", tok.Kind)
				return
			}
			prev = src.Defs
			m.appendSubNode(exprNode{xapi.OutId(src.Id, src.Tok.File)})
			continue
		}
		switch tok.Id {
		case tokens.DoubleColon:
			m.appendSubNode(exprNode{tokens.DOUBLE_COLON})
		default:
			goto eval
		}
	}
eval:
	pdefs := e.p.Defs
	e.p.Defs = prev.(*Defmap)
	parent := e.p.Defs.parent
	e.p.Defs.parent = nil
	defer func() {
		e.p.Defs.parent = parent
		e.p.Defs = pdefs
	}()
	return e.process(toks, m)
}

func (e *eval) id(toks Toks, m *exprModel) (v value) {
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

func (e *eval) operatorRight(toks Toks, m *exprModel) (v value) {
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

func (e *eval) variadic(toks Toks, m *exprModel, errtok Tok) (v value) {
	v = e.process(toks, m)
	if !typeIsVariadicable(v.data.Type) {
		e.pusherrtok(errtok, "variadic_with_nonvariadicable", v.data.Type.Kind)
		return
	}
	v.data.Type = typeOfSliceComponents(v.data.Type)
	v.variadic = true
	return
}

func (e *eval) bracketRange(toks Toks, m *exprModel) (v value) {
	errTok := toks[0]
	var exprToks Toks
	braceCount := 0
	for i := len(toks) - 1; i >= 0; i-- {
		tok := toks[i]
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.RBRACE, tokens.RBRACKET, tokens.RPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount == 0 {
			exprToks = toks[:i]
			break
		}
	}
	if len(exprToks) == 0 || braceCount > 0 {
		e.pusherrtok(errTok, "invalid_syntax")
		return
	}
	var model iExpr
	v, model = e.toks(exprToks)
	m.appendSubNode(model)
	toks = toks[len(exprToks):] // Tokens of [...]
	if toks, colon := ast.SplitColon(toks, new(int)); colon != -1 {
		var leftV, rightV value
		leftToks := toks[:colon]
		rightToks := toks[colon+1:]
		m.appendSubNode(exprNode{".___slice("})
		if len(leftToks) > 0 {
			var model iExpr
			leftV, model = e.p.evalToks(leftToks)
			m.appendSubNode(model)
			e.p.wg.Add(1)
			go assignChecker{
				p:      e.p,
				t:      DataType{Id: xtype.UInt, Kind: xtype.TypeMap[xtype.UInt]},
				v:      leftV,
				errtok: errTok,
			}.checkAssignType()
		} else {
			m.appendSubNode(exprNode{"0"})
		}
		if len(rightToks) > 0 {
			m.appendSubNode(exprNode{","})
			var model iExpr
			rightV, model = e.p.evalToks(rightToks)
			m.appendSubNode(model)
			e.p.wg.Add(1)
			go assignChecker{
				p:      e.p,
				t:      DataType{Id: xtype.UInt, Kind: xtype.TypeMap[xtype.UInt]},
				v:      rightV,
				errtok: errTok,
			}.checkAssignType()
		}
		m.appendSubNode(exprNode{")"})
		return e.slicing(v, errTok)
	}
	m.appendSubNode(exprNode{tokens.LBRACKET})
	leftv, model := e.toks(toks[1 : len(toks)-1])
	m.appendSubNode(model)
	m.appendSubNode(exprNode{tokens.RBRACKET})
	return e.indexing(v, leftv, errTok)
}

func (e *eval) indexing(enumv, leftv value, errtok Tok) (v value) {
	switch {
	case typeIsArray(enumv.data.Type):
		return e.indexingArray(enumv, leftv, errtok)
	case typeIsSlice(enumv.data.Type):
		return e.indexingSlice(enumv, leftv, errtok)
	case typeIsMap(enumv.data.Type):
		return e.indexingMap(enumv, leftv, errtok)
	case typeIsPure(enumv.data.Type):
		return e.indexingStr(enumv, leftv, errtok)
	case typeIsExplicitPtr(enumv.data.Type):
		return e.indexingPtr(enumv, leftv, errtok)
	}
	e.pusherrtok(errtok, "not_supports_indexing", enumv.data.Type.Kind)
	return
}

func (e *eval) indexingSlice(slicev, leftv value, errtok Tok) value {
	slicev.data.Type = typeOfSliceComponents(slicev.data.Type)
	e.p.wg.Add(1)
	go assignChecker{
		p:      e.p,
		t:      DataType{Id: xtype.UInt, Kind: tokens.UINT},
		v:      leftv,
		errtok: errtok,
	}.checkAssignType()
	return slicev
}

func (e *eval) indexingArray(arrv, leftv value, errtok Tok) value {
	arrv.data.Type = typeOfArrayComponents(arrv.data.Type)
	e.p.wg.Add(1)
	go assignChecker{
		p:      e.p,
		t:      DataType{Id: xtype.UInt, Kind: tokens.UINT},
		v:      leftv,
		errtok: errtok,
	}.checkAssignType()
	return arrv
}

func (e *eval) indexingMap(mapv, leftv value, errtok Tok) value {
	types := mapv.data.Type.Tag.([]DataType)
	keyType := types[0]
	valType := types[1]
	mapv.data.Type = valType
	e.p.wg.Add(1)
	go e.p.checkType(keyType, leftv.data.Type, false, errtok)
	return mapv
}

func (e *eval) indexingStr(strv, leftv value, errtok Tok) value {
	strv.data.Type.Id = xtype.U8
	strv.data.Type.Kind = xtype.TypeMap[strv.data.Type.Id]
	e.p.wg.Add(1)
	go assignChecker{
		p:      e.p,
		t:      DataType{Id: xtype.UInt, Kind: tokens.UINT},
		v:      leftv,
		errtok: errtok,
	}.checkAssignType()
	return strv
}

func (e *eval) indexingPtr(ptrv, leftv value, errtok Tok) value {
	ptrv.lvalue = true
	// Remove pointer mark.
	ptrv.data.Type.Kind = ptrv.data.Type.Kind[1:]
	e.p.wg.Add(1)
	go assignChecker{
		p:      e.p,
		t:      DataType{Id: xtype.UInt, Kind: tokens.UINT},
		v:      leftv,
		errtok: errtok,
	}.checkAssignType()
	return ptrv
}

func (e *eval) slicing(enumv value, errtok Tok) (v value) {
	switch {
	case typeIsArray(enumv.data.Type):
		return e.slicingArray(enumv, errtok)
	case typeIsSlice(enumv.data.Type):
		return e.slicingSlice(enumv, errtok)
	case typeIsPure(enumv.data.Type):
		return e.slicingStr(enumv, errtok)
	}
	e.pusherrtok(errtok, "not_supports_slicing", enumv.data.Type.Kind)
	return
}

func (e *eval) slicingSlice(v value, errtok Tok) value {
	v.lvalue = false
	return v
}

func (e *eval) slicingArray(v value, errtok Tok) value {
	v.lvalue = false
	v.data.Type = typeOfArrayComponents(v.data.Type)
	v.data.Type.Kind = x.Prefix_Slice + xtype.TypeMap[v.data.Type.Id]
	return v
}

func (e *eval) slicingStr(v value, errtok Tok) value {
	v.lvalue = false
	v.data.Type.Id = xtype.Str
	v.data.Type.Kind = xtype.TypeMap[v.data.Type.Id]
	return v
}

//! IMPORTANT: Tokens is should be store enumerable parentheses.
func (e *eval) enumerableParts(toks Toks) []Toks {
	toks = toks[1 : len(toks)-1]
	parts, errs := ast.Parts(toks, tokens.Comma, true)
	e.p.pusherrs(errs...)
	return parts
}

func (e *eval) buildArray(parts []Toks, t DataType, errtok Tok) (value, iExpr) {
	if !arrayIsAutoSized(t) {
		n := t.Tag.([][]any)[0][0].(uint64)
		if uint64(len(parts)) > n {
			e.p.pusherrtok(errtok, "overflow_limits")
		}
	} else {
		tag := t.Tag.([][]any)[0]
		n := uint64(len(parts))
		tag[0] = n
		tag[1] = models.Expr{
			Model: exprNode{
				value: xtype.TypeMap[xtype.UInt] + "{" + strconv.FormatUint(n, 10) + "}",
			},
		}
	}
	var v value
	v.data.Value = t.Kind
	v.data.Type = t
	model := sliceExpr{dataType: t}
	elemType := typeOfArrayComponents(t)
	for _, part := range parts {
		partVal, expModel := e.toks(part)
		model.expr = append(model.expr, expModel)
		e.p.wg.Add(1)
		go assignChecker{
			p:      e.p,
			t:      elemType,
			v:      partVal,
			errtok: part[0],
		}.checkAssignType()
	}
	return v, model
}

func (e *eval) buildSlice(parts []Toks, t DataType, errtok Tok) (value, iExpr) {
	var v value
	v.data.Value = t.Kind
	v.data.Type = t
	model := sliceExpr{dataType: t}
	elemType := typeOfSliceComponents(t)
	for _, part := range parts {
		partVal, expModel := e.toks(part)
		model.expr = append(model.expr, expModel)
		e.p.wg.Add(1)
		go assignChecker{
			p:      e.p,
			t:      elemType,
			v:      partVal,
			errtok: part[0],
		}.checkAssignType()
	}
	return v, model
}

func (e *eval) buildMap(parts []Toks, t DataType, errtok Tok) (value, iExpr) {
	var v value
	v.data.Value = t.Kind
	v.data.Type = t
	model := mapExpr{dataType: t}
	types := t.Tag.([]DataType)
	keyType := types[0]
	valType := types[1]
	for _, part := range parts {
		braceCount := 0
		colon := -1
		for i, tok := range part {
			if tok.Id == tokens.Brace {
				switch tok.Kind {
				case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
					braceCount++
				default:
					braceCount--
				}
			}
			if braceCount != 0 {
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
		e.p.wg.Add(1)
		go assignChecker{
			p:      e.p,
			t:      keyType,
			v:      key,
			errtok: colonTok,
		}.checkAssignType()
		e.p.wg.Add(1)
		go assignChecker{
			p:      e.p,
			t:      valType,
			v:      val,
			errtok: colonTok,
		}.checkAssignType()
	}
	return v, model
}

func (e *eval) enumerable(exprToks Toks, t DataType, m *exprModel) (v value) {
	var model iExpr
	if typeIsArray(t) && arrayIsAutoSized(t) {
		exprs := t.Tag.([][]any)[0]
		t = typeOfArrayComponents(t)
		var ok bool
		t, ok = e.p.realType(t, true)
		if !ok {
			return
		}
		t.Kind = x.Prefix_Array + t.Kind
		if t.Tag != nil {
			t.Tag = append([][]any{exprs}, t.Tag.([][]any)...)
		} else {
			t.Tag = [][]any{exprs}
		}
		v, model = e.buildArray(e.enumerableParts(exprToks), t, exprToks[0])
		goto ret
	} else {
		var ok bool
		t, ok = e.p.realType(t, true)
		if !ok {
			return
		}
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
ret:
	m.appendSubNode(model)
	return
}

func (e *eval) braceRange(toks Toks, m *exprModel) (v value) {
	var exprToks Toks
	braceCount := 0
	for i := len(toks) - 1; i >= 0; i-- {
		tok := toks[i]
		if tok.Id != tokens.Brace {
			continue
		}
		switch tok.Kind {
		case tokens.RBRACE, tokens.RBRACKET, tokens.RPARENTHESES:
			braceCount++
		default:
			braceCount--
		}
		if braceCount != 0 {
			continue
		}
		exprToks = toks[:i]
		break
	}
	valToksLen := len(exprToks)
	if valToksLen == 0 || braceCount > 0 {
		e.pusherrtok(toks[0], "invalid_syntax")
		return
	}
	tok := exprToks[0]
	switch exprToks[0].Id {
	case tokens.Id:
		if len(exprToks) > 1 {
			e.pusherrtok(tok, "invalid_syntax")
			return
		}
		return e.typeId(toks, m)
	case tokens.Brace:
		switch exprToks[0].Kind {
		case tokens.LBRACKET:
			b := ast.NewBuilder(nil)
			i := new(int)
			t, ok := b.DataType(exprToks, i, true, true)
			b.Wait()
			if !ok {
				e.p.pusherrs(b.Errors...)
				return
			} else if *i+1 < len(exprToks) {
				e.pusherrtok(toks[*i+1], "invalid_syntax")
			}
			exprToks = toks[len(exprToks):]
			return e.enumerable(exprToks, t, m)
		case tokens.LPARENTHESES:
			b := ast.NewBuilder(toks)
			f := b.Func(b.Toks, true, false)
			b.Wait()
			if len(b.Errors) > 0 {
				e.p.pusherrs(b.Errors...)
				return
			}
			e.p.checkAnonFunc(&f)
			v.data.Value = f.Id
			v.data.Type.Tag = &f
			v.data.Type.Id = xtype.Func
			v.data.Type.Kind = f.DataTypeString()
			m.appendSubNode(anonFuncExpr{f, xapi.LambdaByCopy})
			return
		default:
			e.pusherrtok(exprToks[0], "invalid_syntax")
		}
	default:
		e.pusherrtok(exprToks[0], "invalid_syntax")
	}
	return
}
