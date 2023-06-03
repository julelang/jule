package cxx

import (
	"unsafe"

	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/sema"
)

type _RangeSetter interface {
	setup_vars(key_a *sema.Var, key_b *sema.Var) string
	next_steps(key_a *sema.Var, key_b *sema.Var, begin string) string
}

type _IndexRangeSetter struct {}

func (*_IndexRangeSetter) setup_vars(key_a *sema.Var, key_b *sema.Var) string {
	indent := indent()

	obj := ""
	if key_a != nil {
		obj += gen_var(key_a)
		obj += var_out_ident(key_a)
		obj += " = 0;\n"
		obj += indent
	}

	if key_b != nil {
		obj += gen_var(key_b)
		obj += var_out_ident(key_b)
		obj += " = *__julec_range_begin;\n"
		obj += indent
	}

	return obj
}

func (*_IndexRangeSetter) next_steps(key_a *sema.Var, key_b *sema.Var, begin string) string {
	indent := indent()
	
	obj := "++__julec_range_begin;\n"
	obj += indent
	
	obj += "if (__julec_range_begin != __julec_range_end) { "
	if key_a != nil {
		obj += "++" + var_out_ident(key_a) + "; "
	}
	if key_b != nil {
		obj += var_out_ident(key_b) + " = *__julec_range_begin; "
	}

	obj += "goto " + begin + "; }\n"
	return obj
}

type _MapRangeSetter struct {}

func (*_MapRangeSetter) setup_vars(key_a *sema.Var, key_b *sema.Var) string {
	indent := indent()
	obj := ""

	if key_a != nil {
		obj += gen_var(key_a)
		obj += var_out_ident(key_a)
		obj += " = __julec_range_begin->first;\n"
		obj += indent
	}

	if key_b != nil {
		obj += gen_var(key_b)
		obj += var_out_ident(key_b)
		obj += " = __julec_range_begin->second;\n"
		obj += indent
	}

	return obj
}

func (*_MapRangeSetter) next_steps(key_a *sema.Var, key_b *sema.Var, begin string) string {
	indent := indent()

	obj := "++__julec_range_begin;\n"
	obj += indent
	
	obj += "if (__julec_range_begin != __julec_range_end) { "
	if key_a != nil {
		obj += var_out_ident(key_a)
		obj += " = __julec_range_begin->first; "
	}
	if key_b != nil {
		obj += var_out_ident(key_b)
		obj += " = __julec_range_begin->second; "
	}

	obj += "goto " + begin + "; }\n"

	return obj
}

// In Jule: (uintptr)(PTR)
func _uintptr[T any](t *T) uintptr { return uintptr(unsafe.Pointer(t)) }

func gen_if(i *sema.If) string {
	obj := "if ("
	obj += gen_expr(i.Expr)
	obj += ") "
	obj += gen_scope(i.Scope)
	return obj
}

func gen_conditional(c *sema.Conditional) string {
	obj := gen_if(c.Elifs[0])

	for _, elif := range c.Elifs[1:] {
		obj += " else "
		obj += gen_if(elif)
	}

	if c.Default != nil {
		obj += " else "
		obj += gen_scope(c.Default.Scope)
	}

	return obj
}

func gen_inf_iter(it *sema.InfIter) string {
	begin := iter_begin_label_ident(_uintptr(it))
	end := iter_end_label_ident(_uintptr(it))
	next := iter_next_label_ident(_uintptr(it))
	indent := indent()

	obj := begin + ":;\n"
	obj += indent
	obj += gen_scope(it.Scope)
	obj += "\n"
	obj += indent
	obj += next + ":;\n"
	obj += indent
	obj += "goto " + begin + ";\n"
	obj += indent
	obj += end + ":;"

	return obj
}

func gen_while_iter(it *sema.WhileIter) string {
	begin := iter_begin_label_ident(_uintptr(it))
	end := iter_end_label_ident(_uintptr(it))
	next := iter_next_label_ident(_uintptr(it))
	indent := indent()

	obj := begin + ":;\n"
	obj += indent
	if it.Expr != nil {
		obj += "if (!("
		obj += gen_expr(it.Expr)
		obj += ")) { goto "
		obj += end
		obj += "; }\n"
		obj += indent
	}
	obj += gen_scope(it.Scope)
	obj += "\n"
	obj += indent
	obj += next + ":;\n"
	obj += indent
	if it.Next != nil {
		obj += gen_st(it.Next)
		obj += "\n"
		obj += indent
	}
	obj += "goto " + begin + ";\n"
	obj += indent
	obj += end + ":;"

	return obj
}

func get_range_setter(it *sema.RangeIter) _RangeSetter {
	switch {
	case it.Expr.Kind.Slc() != nil:
		return &_IndexRangeSetter{}

	case it.Expr.Kind.Arr() != nil:
		return &_IndexRangeSetter{}

	case it.Expr.Kind.Map() != nil:
		return &_MapRangeSetter{}

	default: // Str
		return &_IndexRangeSetter{}
	}
}

func gen_range_iter(it *sema.RangeIter) string {
	add_indent()

	begin := iter_begin_label_ident(_uintptr(it))
	end := iter_end_label_ident(_uintptr(it))
	next := iter_next_label_ident(_uintptr(it))
	_indent := indent()
	setter := get_range_setter(it)

	obj := "{\n"
	obj += _indent
	obj += "auto __julec_range_expr = "
	obj += gen_expr(it.Expr.Model) + ";\n"
	obj += _indent
	obj += "if (__julec_range_expr.begin() != __julec_range_expr.end()) {\n"

	add_indent()
	_indent = indent()

	obj += _indent
	obj += "auto __julec_range_begin = __julec_range_expr.begin();\n"
	obj += _indent
	obj += "const auto __julec_range_end = __julec_range_expr.end();\n"
	obj += _indent
	obj += setter.setup_vars(it.Key_a, it.Key_b)
	obj += begin + ":;\n"
	obj += _indent
	obj += gen_scope(it.Scope)
	obj += "\n"
	obj += _indent
	obj += next + ":;\n"
	obj += _indent
	obj += setter.next_steps(it.Key_a, it.Key_b, begin)
	obj += _indent
	obj += end + ":;\n"

	done_indent()
	_indent = indent()

	obj += _indent
	obj += "}\n"

	done_indent()
	_indent = indent()

	obj += _indent
	obj += "}"

	return obj
}

func gen_cont(c *sema.ContSt) string {
	return "goto " + iter_next_label_ident(c.It) + CPP_ST_TERM
}

func gen_label(l *sema.Label) string {
	return label_ident(l.Ident) + ":;"
}

func gen_goto(gt *sema.GotoSt) string {
	return "goto " + label_ident(gt.Ident) + CPP_ST_TERM
}

func gen_postfix(p *sema.Postfix) string {
	return "(" + gen_expr(p.Expr) + ")" + p.Op + CPP_ST_TERM
}

func gen_assign(a *sema.Assign) string {
	obj := gen_expr(a.L)
	obj += a.Op
	obj += gen_expr(a.R)
	obj += CPP_ST_TERM
	return obj
}

func gen_multi_assign(a *sema.MultiAssign) string {
	obj := "std::tie("
	
	for _, l := range a.L {
		if l == nil {
			obj += CPP_IGNORE + ","
		} else {
			obj += gen_expr(l) + ","
		}
	}
	obj = obj[:len(obj)-1] // Remove last comma.

	obj += ") = "
	obj += gen_expr(a.R)
	obj += CPP_ST_TERM
	return obj
}

func gen_case(m *sema.Match, c *sema.Case) string {
	const MATCH_EXPR = "_match_expr"

	end := case_end_label_ident(_uintptr(c))
	obj := ""

	if len(c.Exprs) > 0 {
		obj += "if (!("
		for i, expr := range c.Exprs {
			if !m.Type_match {
				obj += gen_expr(expr)
				obj += " == "
			}

			obj += MATCH_EXPR

			if m.Type_match {
				obj += ".__type_is<" + gen_expr(expr)  + ">()"
			}

			if i+1 < len(c.Exprs) {
				obj += " || "
			}
		}
		obj += ")) { goto "
		obj += end + "; }\n"
	}

	if len(c.Scope.Stmts) > 0 {
		obj += indent()
		obj += case_begin_label_ident(_uintptr(c)) + ":;\n"
		obj += indent()
		obj += gen_scope(c.Scope)
		obj += "\n"
	}

	obj += indent()
	obj += "goto "
	obj += match_end_label_ident(_uintptr(m)) + CPP_ST_TERM
	obj += "\n"
	obj += indent()
	obj += end + ":;"
	return obj
}

func gen_match(m *sema.Match) string {
	obj := "{\n"

	add_indent()

	obj += indent()
	obj += "auto _match_expr{ "
	obj += gen_expr(m.Expr)
	obj += " };\n"
	obj += indent()

	if len(m.Cases) > 0 {
		obj += gen_case(m, m.Cases[0])
		for _, c := range m.Cases[1:] {
			obj += "\n"
			obj += indent()
			obj += gen_case(m, c)
		}
	}

	if m.Default != nil {
		obj += "\n"
		obj += gen_case(m, m.Default)
	}

	obj += "\n"
	obj += indent()
	obj += match_end_label_ident(_uintptr(m)) + ":;"
	obj += "\n"
	
	done_indent()

	obj += indent()
	obj += "}"

	return obj
}

func gen_fall_st(f *sema.FallSt) string {
	return "goto " + case_begin_label_ident(f.Dest_case) + CPP_ST_TERM
}

func gen_break_st(b *sema.BreakSt) string {
	obj := "goto "
	if b.It != 0 {
		obj += iter_end_label_ident(b.It)
	} else {
		obj += match_end_label_ident(b.Mtch)
	}

	obj += CPP_ST_TERM
	return obj
}

func gen_ret_vars(r *sema.RetSt) string {
	obj := ""
	for _, v := range r.Vars {
		if lex.Is_ignore_ident(v.Ident) {
			obj += get_init_expr(v.Kind.Kind)
		} else {
			obj += var_out_ident(v)
		}

		obj += lex.KND_COMMA
	}

	obj = obj[:len(obj)-1] // Remove last comma.

	if len(r.Vars) > 1 {
		obj = "return std::make_tuple(" + obj + ")"
	} else {
		obj = "return " + obj
	}

	obj += CPP_ST_TERM
	return obj
}

func gen_ret_expr_tuple(r *sema.RetSt) string {
	switch r.Expr.(type) {
	case *sema.FnCallExprModel:
		return "return " + gen_expr_model(r.Expr) + CPP_ST_TERM
	}

	datas := r.Expr.(*sema.TupleExprModel).Datas
	obj := ""

	for i, v := range r.Vars {
		if !lex.Is_ignore_ident(v.Ident) {
			ident := var_out_ident(v)
			obj += ident + " = " + gen_expr(datas[i].Model) + ";\n"
			obj += indent()
		}
	}

	obj += "return std::make_tuple("
	for i, d := range datas {
		v := r.Vars[i]
		if lex.Is_ignore_ident(v.Ident) {
			obj += gen_expr(d.Model)
		} else {
			obj += var_out_ident(v)
		}

		obj += ","
	}
	obj = obj[:len(obj)-1] // Remove last comma.
	obj += ");"

	return obj
}

func gen_ret_expr(r *sema.RetSt) string {
	if len(r.Vars) == 0 {
		return "return " + gen_expr(r.Expr) + CPP_ST_TERM
	}

	if len(r.Vars) > 1 {
		return gen_ret_expr_tuple(r)
	}

	obj := ""
	if !lex.Is_ignore_ident(r.Vars[0].Ident) {
		ident := var_out_ident(r.Vars[0])
		obj += ident + " = " + gen_expr(r.Expr) + ";\n"
		obj += indent()
		obj += "return " + ident + CPP_ST_TERM
		return obj
	}

	return "return " + gen_expr(r.Expr) + CPP_ST_TERM
}

func gen_ret_st(r *sema.RetSt) string {
	if r.Expr == nil && len(r.Vars) == 0 {
		return "return;"
	}

	if r.Expr == nil {
		return gen_ret_vars(r)
	}

	return gen_ret_expr(r)
}

func gen_recover(r *sema.Recover) string {
	obj := "try "
	obj += gen_scope(r.Scope)
	obj += " catch(jule::Exception e) "
	if r.Handler.Is_anon() {
		// Anonymous function.
		// Parse body as catch block.
		//
		// NOTICE:
		//  If passed anonymous function from variable, field, or something
		//  like that, parses block. Not calls variable, fields or whatever.

		handler_param := r.Handler.Decl.Params[0]
		if !lex.Is_ignore_ident(handler_param.Ident) && !lex.Is_anon_ident(handler_param.Ident) {
			add_indent()
			obj += "{\n"
			obj += indent()
			obj += "auto "
			obj += param_out_ident(handler_param)
			obj += "{ jule::exception_to_error(e) };\n"
			obj += indent()
			obj += gen_scope(r.Handler.Scope)
			done_indent()
			obj += "\n"
			obj += indent()
			obj += "}"
		} else {
			obj += gen_scope(r.Handler.Scope)
		}
	} else {
		// Passed defined function.
		// Therefore, call passed function with error.

		obj += "{ "
		obj += gen_expr(r.Handler_expr)
		obj += "(jule::exception_to_error(e)); }"
	}

	return obj
}

// Generates C++ code of statement.
func gen_st(st sema.St) string {
	switch st.(type) {
	case *sema.Scope:
		return gen_scope(st.(*sema.Scope))

	case *sema.Var:
		return gen_var(st.(*sema.Var))

	case *sema.TypeAlias:
		return "// " + gen_type_alias(st.(*sema.TypeAlias))

	case *sema.Data:
		return gen_expr(st.(*sema.Data).Model) + CPP_ST_TERM

	case *sema.Conditional:
		return gen_conditional(st.(*sema.Conditional))

	case *sema.InfIter:
		return gen_inf_iter(st.(*sema.InfIter))

	case *sema.WhileIter:
		return gen_while_iter(st.(*sema.WhileIter))

	case *sema.RangeIter:
		return gen_range_iter(st.(*sema.RangeIter))

	case *sema.ContSt:
		return gen_cont(st.(*sema.ContSt))

	case *sema.Label:
		return gen_label(st.(*sema.Label))

	case *sema.GotoSt:
		return gen_goto(st.(*sema.GotoSt))

	case *sema.Postfix:
		return gen_postfix(st.(*sema.Postfix))

	case *sema.Assign:
		return gen_assign(st.(*sema.Assign))

	case *sema.MultiAssign:
		return gen_multi_assign(st.(*sema.MultiAssign))

	case *sema.Match:
		return gen_match(st.(*sema.Match))

	case *sema.FallSt:
		return gen_fall_st(st.(*sema.FallSt))

	case *sema.BreakSt:
		return gen_break_st(st.(*sema.BreakSt))

	case *sema.RetSt:
		return gen_ret_st(st.(*sema.RetSt))

	case *sema.Recover:
		return gen_recover(st.(*sema.Recover))

	default:
		return "<unimplemented stmt>"
	}
}

// Generates C++ code of scope.
func gen_scope(s *sema.Scope) string {
	obj := "{\n"
	add_indent()

	for _, st := range s.Stmts {
		obj += indent()
		obj += gen_st(st)
		obj += "\n"
	}

	done_indent()
	obj += indent()
	obj += "}"
	
	if s.Deferred {
		obj = "__JULE_DEFER(" + obj + ");"
	}

	return obj
}

// Generates C++ code of function's scope.
func gen_fn_scope(f *sema.FnIns) string {
	if f.Owner != nil {
		return gen_method_scope(f)
	}

	return gen_scope(f.Scope)
}

// Generates C++ code of method's scope.
func gen_method_scope(f *sema.FnIns) string {
	obj := gen_scope(f.Scope)

	slf_decl := ""
	slf := f.Decl.Params[0]
	if slf.Is_ref() {
		slf_decl = "auto self{ this->self };"
	} else if !slf.Mutable {
		slf_decl = gen_struct_kind_ins(f.Owner) + " self{ *this };"
	}

	robj := "{\n"
	add_indent()
	robj += indent()
	done_indent()
	robj += slf_decl
	robj += "\n"
	robj += obj[2:] // Skip left brace and new line.

	return robj
}
