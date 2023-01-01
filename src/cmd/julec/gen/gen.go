package gen

import (
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/julelang/jule"
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

// indent is indention count.
// This should be manuplate atomic.
var indent uint32 = 0

// indentation string.
var indentation = "\t"

// init_caller represents initializer caller function identifier.
const init_caller = "__julec_call_package_initializers"

// Returns indent space of current block.
func indent_string() string {
	return strings.Repeat(indentation, int(indent))
}

// Adds new indent to indent_string.
func add_indent() { atomic.AddUint32(&indent, 1) }

// Removes last indent from indent_string.
func done_indent() {
	atomic.SwapUint32(&indent, atomic.LoadUint32(&indent)-1)
}

func gen_params(params []ast.Param) string {
	if len(params) == 0 {
		return "(void)"
	}
	var cpp strings.Builder
	cpp.WriteByte('(')
	for _, p := range params {
		cpp.WriteString(p.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ")"
}

func gen_generics(generics []*ast.GenericType) string {
	if len(generics) == 0 {
		return ""
	}
	var cpp strings.Builder
	cpp.WriteString("template<")
	for _, g := range generics {
		cpp.WriteString(g.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ">"
}

func gen_assign_left(as *ast.AssignLeft) string {
	switch {
	case as.Var.New:
		return as.Var.OutId()
	case as.Ignore:
		return build.CPP_IGNORE
	}
	return as.Expr.String()
}

func gen_single_assign(a *ast.Assign) string {
	left := a.Left[0]
	if left.Var.New {
		left.Var.Expr = a.Right[0]
		s := left.Var.String()
		return s[:len(s)-1] // Remove statement terminator
	}
	var cpp strings.Builder
	if len(left.Expr.Tokens) != 1 ||
		!lex.IsIgnoreId(left.Expr.Tokens[0].Kind) {
		cpp.WriteString(gen_assign_left(&left))
		cpp.WriteString(a.Setter.Kind)
	}
	cpp.WriteString(a.Right[0].String())
	return cpp.String()
}

func assign_has_left(a *ast.Assign) bool {
	for _, s := range a.Left {
		if !s.Ignore {
			return true
		}
	}
	return false
}

func gen_multiple_assign(a *ast.Assign) string {
	var cpp strings.Builder
	if !assign_has_left(a) {
		for _, right := range a.Right {
			cpp.WriteString(right.String())
			cpp.WriteByte(';')
		}
		return cpp.String()[:cpp.Len()-1] // Remove last semicolon
	}
	cpp.WriteString(gen_assign_new_defines(a))
	cpp.WriteString("std::tie(")
	var exprCpp strings.Builder
	exprCpp.WriteString("std::make_tuple(")
	for i := range a.Left {
		left := &a.Left[i]
		cpp.WriteString(gen_assign_left(left))
		cpp.WriteByte(',')
		exprCpp.WriteString(a.Right[i].String())
		exprCpp.WriteByte(',')
	}
	str := cpp.String()[:cpp.Len()-1] + ")"
	cpp.Reset()
	cpp.WriteString(str)
	cpp.WriteString(a.Setter.Kind)
	cpp.WriteString(exprCpp.String()[:exprCpp.Len()-1] + ")")
	return cpp.String()
}

func gen_assign_multi_ret(a *ast.Assign) string {
	var cpp strings.Builder
	cpp.WriteString(gen_assign_new_defines(a))
	cpp.WriteString("std::tie(")
	for i := range a.Left {
		left := &a.Left[i]
		if left.Ignore {
			cpp.WriteString(build.CPP_IGNORE)
			cpp.WriteByte(',')
			continue
		}
		cpp.WriteString(gen_assign_left(left))
		cpp.WriteByte(',')
	}
	str := cpp.String()[:cpp.Len()-1]
	cpp.Reset()
	cpp.WriteString(str)
	cpp.WriteByte(')')
	cpp.WriteString(a.Setter.Kind)
	cpp.WriteString(a.Right[0].String())
	return cpp.String()
}

func gen_assign_new_defines(a *ast.Assign) string {
	var cpp strings.Builder
	for _, left := range a.Left {
		if left.Ignore || !left.Var.New {
			continue
		}
		cpp.WriteString(left.Var.String() + " ")
	}
	return cpp.String()
}

func gen_assign_postfix(a *ast.Assign) string {
	var cpp strings.Builder
	cpp.WriteString(a.Left[0].Expr.String())
	cpp.WriteString(a.Setter.Kind)
	return cpp.String()
}

func gen_assign(a *ast.Assign) string {
	var cpp strings.Builder
	switch {
	case len(a.Right) == 0:
		cpp.WriteString(gen_assign_postfix(a))
	case a.MultipleRet:
		cpp.WriteString(gen_assign_multi_ret(a))
	case len(a.Left) == 1:
		cpp.WriteString(gen_single_assign(a))
	default:
		cpp.WriteString(gen_multiple_assign(a))
	}
	if !a.IsExpr {
		cpp.WriteByte(';')
	}
	return cpp.String()
}

func gen_block(b *ast.Block) string {
	add_indent()
	s := ""
	if b.Deferred {
		s = "__JULEC_DEFER("
	}
	s += gen_parse_block(b)
	done_indent()
	if b.Deferred {
		s += ");"
	}
	return s
}

func gen_parse_block(b *ast.Block) string {
	// Space count per indent.
	var cpp strings.Builder
	cpp.WriteByte('{')
	for _, s := range b.Tree {
		if s.Data == nil {
			continue
		}
		cpp.WriteByte('\n')
		cpp.WriteString(indent_string())
		cpp.WriteString(gen_st(&s))
	}
	cpp.WriteByte('\n')
	indent := strings.Repeat(indentation, int(indent-1))
	cpp.WriteString(indent)
	cpp.WriteByte('}')
	return cpp.String()
}

func gen_concurrent_call(cc *ast.ConcurrentCall) string {
	var cpp strings.Builder
	cpp.WriteString("__JULEC_CO(")
	cpp.WriteString(cc.Expr.String())
	cpp.WriteString(");")
	return cpp.String()
}

func gen_if(i *ast.If) string {
	var cpp strings.Builder
	cpp.WriteString("if (")
	cpp.WriteString(i.Expr.String())
	cpp.WriteString(") ")
	cpp.WriteString(gen_block(i.Block))
	return cpp.String()
}

func gen_else(e *ast.Else) string {
	var cpp strings.Builder
	cpp.WriteString("else ")
	cpp.WriteString(gen_block(e.Block))
	return cpp.String()
}

func gen_conditional(c *ast.Conditional) string {
	var cpp strings.Builder
	cpp.WriteString(gen_if(c.If))
	for _, elif := range c.Elifs {
		cpp.WriteString(" else ")
		cpp.WriteString(gen_if(elif))
	}
	if c.Default != nil {
		cpp.WriteByte(' ')
		cpp.WriteString(gen_else(c.Default))
	}
	return cpp.String()
}

func gen_iter_while(w *ast.IterWhile, i *ast.Iter) string {
	var cpp strings.Builder
	indent := indent_string()
	begin := i.BeginLabel()
	next := i.NextLabel()
	end := i.EndLabel()
	cpp.WriteString(begin)
	cpp.WriteString(":;\n")
	cpp.WriteString(indent)
	if !w.Expr.IsEmpty() {
		cpp.WriteString("if (!(")
		cpp.WriteString(w.Expr.String())
		cpp.WriteString(")) { goto ")
		cpp.WriteString(end)
		cpp.WriteString("; }\n")
		cpp.WriteString(indent)
	}
	cpp.WriteString(gen_block(i.Block))
	cpp.WriteByte('\n')
	cpp.WriteString(indent)
	cpp.WriteString(next)
	cpp.WriteString(":;\n")
	cpp.WriteString(indent)
	if w.Next.Data != nil {
		cpp.WriteString(gen_st(&w.Next))
		cpp.WriteByte('\n')
		cpp.WriteString(indent)
	}
	cpp.WriteString("goto ")
	cpp.WriteString(begin)
	cpp.WriteString(";\n")
	cpp.WriteString(indent)
	cpp.WriteString(end)
	cpp.WriteString(":;")
	return cpp.String()
}

func gen_iter_inf(i *ast.Iter) string {
	var cpp strings.Builder
	indent := indent_string()
	begin := i.BeginLabel()
	cpp.WriteString(begin)
	cpp.WriteString(":;\n")
	cpp.WriteString(indent)
	cpp.WriteString(gen_block(i.Block))
	cpp.WriteByte('\n')
	cpp.WriteString(indent)
	cpp.WriteString(i.NextLabel())
	cpp.WriteString(":;\n")
	cpp.WriteString(indent)
	cpp.WriteString("goto ")
	cpp.WriteString(begin)
	cpp.WriteString(";\n")
	cpp.WriteString(indent)
	cpp.WriteString(i.EndLabel())
	cpp.WriteString(":;")
	return cpp.String()
}

type foreach_setter interface {
	setup_vars(key_a *ast.Var, key_b *ast.Var) string
	next_steps(ket_a *ast.Var, key_b *ast.Var, begin string) string
}

type index_setter struct{}

func (index_setter) setup_vars(key_a *ast.Var, key_b *ast.Var) string {
	var cpp strings.Builder
	indent := indent_string()
	if !lex.IsIgnoreId(key_a.Id) {
		if key_a.New {
			cpp.WriteString(key_a.String())
			cpp.WriteByte(' ')
		}
		cpp.WriteString(key_a.OutId())
		cpp.WriteString(" = 0;\n")
		cpp.WriteString(indent)
	}
	if !lex.IsIgnoreId(key_b.Id) {
		if key_b.New {
			cpp.WriteString(key_b.String())
			cpp.WriteByte(' ')
		}
		cpp.WriteString(key_b.OutId())
		cpp.WriteString(" = *__julec_foreach_begin;\n")
		cpp.WriteString(indent)
	}
	return cpp.String()
}

func (index_setter) next_steps(key_a *ast.Var, key_b *ast.Var, begin string) string {
	var cpp strings.Builder
	indent := indent_string()
	cpp.WriteString("++__julec_foreach_begin;\n")
	cpp.WriteString(indent)
	cpp.WriteString("if (__julec_foreach_begin != __julec_foreach_end) { ")
	if !lex.IsIgnoreId(key_a.Id) {
		cpp.WriteString("++")
		cpp.WriteString(key_a.OutId())
		cpp.WriteString("; ")
	}
	if !lex.IsIgnoreId(key_b.Id) {
		cpp.WriteString(key_b.OutId())
		cpp.WriteString(" = *__julec_foreach_begin; ")
	}
	cpp.WriteString("goto ")
	cpp.WriteString(begin)
	cpp.WriteString("; }\n")
	return cpp.String()
}

type map_setter struct{}

func (map_setter) setup_vars(key_a *ast.Var, key_b *ast.Var) string {
	var cpp strings.Builder
	indent := indent_string()
	if !lex.IsIgnoreId(key_a.Id) {
		if key_a.New {
			cpp.WriteString(key_a.String())
			cpp.WriteByte(' ')
		}
		cpp.WriteString(key_a.OutId())
		cpp.WriteString(" = __julec_foreach_begin->first;\n")
		cpp.WriteString(indent)
	}
	if !lex.IsIgnoreId(key_b.Id) {
		if key_b.New {
			cpp.WriteString(key_b.String())
			cpp.WriteByte(' ')
		}
		cpp.WriteString(key_b.OutId())
		cpp.WriteString(" = __julec_foreach_begin->second;\n")
		cpp.WriteString(indent)
	}
	return cpp.String()
}

func (map_setter) next_steps(key_a *ast.Var, key_b *ast.Var, begin string) string {
	var cpp strings.Builder
	indent := indent_string()
	cpp.WriteString("++__julec_foreach_begin;\n")
	cpp.WriteString(indent)
	cpp.WriteString("if (__julec_foreach_begin != __julec_foreach_end) { ")
	if !lex.IsIgnoreId(key_a.Id) {
		cpp.WriteString(key_a.OutId())
		cpp.WriteString(" = __julec_foreach_begin->first; ")
	}
	if !lex.IsIgnoreId(key_b.Id) {
		cpp.WriteString(key_b.OutId())
		cpp.WriteString(" = __julec_foreach_begin->second; ")
	}
	cpp.WriteString("goto ")
	cpp.WriteString(begin)
	cpp.WriteString("; }\n")
	return cpp.String()
}

func gen_iter_foreach(f *ast.IterForeach, i *ast.Iter) string {
	switch f.ExprType.Id {
	case types.STR, types.SLICE, types.ARRAY:
		return gen_foreach_iter(f, i, index_setter{})
	case types.MAP:
		return gen_foreach_iter(f, i, map_setter{})
	}
	return ""
}

func gen_foreach_iter(f *ast.IterForeach, i *ast.Iter, setter foreach_setter) string {
	var cpp strings.Builder
	cpp.WriteString("{\n")
	add_indent()
	indent := indent_string()
	cpp.WriteString(indent)
	cpp.WriteString("auto __julec_foreach_expr = ")
	cpp.WriteString(f.Expr.String())
	cpp.WriteString(";\n")
	cpp.WriteString(indent)
	cpp.WriteString("if (__julec_foreach_expr.begin() != __julec_foreach_expr.end()) {\n")
	add_indent()
	indent = indent_string()
	cpp.WriteString(indent)
	cpp.WriteString("auto __julec_foreach_begin = __julec_foreach_expr.begin();\n")
	cpp.WriteString(indent)
	cpp.WriteString("const auto __julec_foreach_end = __julec_foreach_expr.end();\n")
	cpp.WriteString(indent)
	cpp.WriteString(setter.setup_vars(&f.KeyA, &f.KeyB))
	begin := i.BeginLabel()
	cpp.WriteString(begin)
	cpp.WriteString(":;\n")
	cpp.WriteString(indent)
	cpp.WriteString(gen_block(i.Block))
	cpp.WriteByte('\n')
	cpp.WriteString(indent)
	cpp.WriteString(i.NextLabel())
	cpp.WriteString(":;\n")
	cpp.WriteString(indent)
	cpp.WriteString(setter.next_steps(&f.KeyA, &f.KeyB, begin))
	cpp.WriteString(indent)
	cpp.WriteString(i.EndLabel())
	cpp.WriteString(":;")
	cpp.WriteByte('\n')
	done_indent()
	cpp.WriteString(indent_string())
	cpp.WriteString("}\n")
	done_indent()
	cpp.WriteString(indent_string())
	cpp.WriteByte('}')
	return cpp.String()
}

func gen_iter(i *ast.Iter) string {
	if i.Profile == nil {
		return gen_iter_inf(i)
	}
	switch t := i.Profile.(type) {
	case ast.IterForeach:
		return gen_iter_foreach(&t, i)
	case ast.IterWhile:
		return gen_iter_while(&t, i)
	default:
		return ""
	}
}

func gen_type_alias(t *ast.TypeAlias) string {
	var cpp strings.Builder
	cpp.WriteString("typedef ")
	cpp.WriteString(t.Type.String())
	cpp.WriteByte(' ')
	if t.Generic {
		cpp.WriteString(build.AsId(t.Id))
	} else {
		cpp.WriteString(build.OutId(t.Id, t.Token.File.Addr()))
	}
	cpp.WriteByte(';')
	return cpp.String()
}

func gen_st(s *ast.Statement) string {
	switch t := s.Data.(type) {
	case ast.ExprStatement:
		return gen_expr_st(&t)
	case ast.Var:
		return t.String()
	case ast.Assign:
		return gen_assign(&t)
	case ast.Break:
		return t.String()
	case ast.Continue:
		return t.String()
	case *ast.Match:
		return gen_match(t)
	case ast.TypeAlias:
		return gen_type_alias(&t)
	case *ast.Block:
		return gen_block(t)
	case ast.ConcurrentCall:
		return gen_concurrent_call(&t)
	case ast.Comment:
		return t.String()
	case ast.Iter:
		return gen_iter(&t)
	case ast.Fallthrough:
		return gen_fallthrough(&t)
	case ast.Conditional:
		return gen_conditional(&t)
	case ast.Ret:
		return gen_ret_st(&t)
	case ast.Goto:
		return t.String()
	case ast.Label:
		return t.String()
	default:
		return ""
	}
}

func gen_expr_st(be *ast.ExprStatement) string {
	var cpp strings.Builder
	cpp.WriteString(be.Expr.String())
	cpp.WriteByte(';')
	return cpp.String()
}

func gen_ret_st(r *ast.Ret) string {
	if r.Expr.Model == nil {
		return "return;"
	}
	var cpp strings.Builder
	cpp.WriteString(r.Expr.String())
	cpp.WriteByte(';')
	return cpp.String()
}

func gen_fallthrough(f *ast.Fallthrough) string {
	var cpp strings.Builder
	cpp.WriteString("goto ")
	cpp.WriteString(f.Case.Next.BeginLabel())
	cpp.WriteByte(';')
	return cpp.String()
}

func gen_case(c *ast.Case, matchExpr string) string {
	endlabel := c.EndLabel()
	var cpp strings.Builder
	if len(c.Exprs) > 0 {
		cpp.WriteString("if (!(")
		for i, expr := range c.Exprs {
			cpp.WriteString(expr.String())
			if matchExpr != "" {
				cpp.WriteString(" == ")
				cpp.WriteString(matchExpr)
			}
			if i+1 < len(c.Exprs) {
				cpp.WriteString(" || ")
			}
		}
		cpp.WriteString(")) { goto ")
		cpp.WriteString(endlabel)
		cpp.WriteString("; }\n")
	}
	if len(c.Block.Tree) > 0 {
		cpp.WriteString(indent_string())
		cpp.WriteString(c.BeginLabel())
		cpp.WriteString(":;\n")
		cpp.WriteString(indent_string())
		cpp.WriteString(gen_block(c.Block))
		cpp.WriteByte('\n')
		cpp.WriteString(indent_string())
		cpp.WriteString("goto ")
		cpp.WriteString(c.Match.EndLabel())
		cpp.WriteString(";")
		cpp.WriteByte('\n')
	}
	cpp.WriteString(indent_string())
	cpp.WriteString(endlabel)
	cpp.WriteString(":;")
	return cpp.String()
}

func gen_match_expr(m *ast.Match) string {
	if len(m.Cases) == 0 {
		if m.Default != nil {
			return gen_case(m.Default, "")
		}
		return ""
	}
	var cpp strings.Builder
	cpp.WriteString("{\n")
	add_indent()
	cpp.WriteString(indent_string())
	cpp.WriteString(m.ExprType.String())
	cpp.WriteString(" expr{")
	cpp.WriteString(m.Expr.String())
	cpp.WriteString("};\n")
	cpp.WriteString(indent_string())
	if len(m.Cases) > 0 {
		cpp.WriteString(gen_case(&m.Cases[0], "expr"))
		for _, c := range m.Cases[1:] {
			cpp.WriteByte('\n')
			cpp.WriteString(indent_string())
			cpp.WriteString(gen_case(&c, "expr"))
		}
	}
	if m.Default != nil {
		cpp.WriteString(gen_case(m.Default, ""))
	}
	cpp.WriteByte('\n')
	done_indent()
	cpp.WriteString(indent_string())
	cpp.WriteByte('}')
	return cpp.String()
}

func gen_match_bool(m *ast.Match) string {
	var cpp strings.Builder
	if len(m.Cases) > 0 {
		cpp.WriteString(gen_case(&m.Cases[0], ""))
		for _, c := range m.Cases[1:] {
			cpp.WriteByte('\n')
			cpp.WriteString(indent_string())
			cpp.WriteString(gen_case(&c, ""))
		}
	}
	if m.Default != nil {
		cpp.WriteByte('\n')
		cpp.WriteString(gen_case(m.Default, ""))
		cpp.WriteByte('\n')
	}
	return cpp.String()
}

func gen_match(m *ast.Match) string {
	var cpp strings.Builder
	if m.Expr.Model != nil {
		cpp.WriteString(gen_match_expr(m))
	} else {
		cpp.WriteString(gen_match_bool(m))
	}
	cpp.WriteByte('\n')
	cpp.WriteString(indent_string())
	cpp.WriteString(m.EndLabel())
	cpp.WriteString(":;")
	return cpp.String()
}

func gen_struct_ostream(s *ast.Struct) string {
	var cpp strings.Builder
	genericsDef, genericsSerie := gen_struct_generics(s)
	cpp.WriteString(indent_string())
	if l, _ := cpp.WriteString(genericsDef); l > 0 {
		cpp.WriteString(indent_string())
	}
	cpp.WriteString("std::ostream &operator<<(std::ostream &_Stream, const ")
	cpp.WriteString(s.OutId())
	cpp.WriteString(genericsSerie)
	cpp.WriteString(" &_Src) {\n")
	add_indent()
	cpp.WriteString(indent_string())
	cpp.WriteString(`_Stream << "`)
	cpp.WriteString(s.Id)
	cpp.WriteString("{\";\n")
	for i, field := range s.Fields {
		cpp.WriteString(indent_string())
		cpp.WriteString(`_Stream << "`)
		cpp.WriteString(field.Id)
		cpp.WriteString(`:" << _Src.`)
		cpp.WriteString(field.OutId())
		if i+1 < len(s.Fields) {
			cpp.WriteString(" << \", \"")
		}
		cpp.WriteString(";\n")
	}
	cpp.WriteString(indent_string())
	cpp.WriteString("_Stream << \"}\";\n")
	cpp.WriteString(indent_string())
	cpp.WriteString("return _Stream;\n")
	done_indent()
	cpp.WriteString(indent_string())
	cpp.WriteString("}")
	return cpp.String()
}

func gen_struct_generics(s *ast.Struct) (def string, serie string) {
	if len(s.Generics) == 0 {
		return "", ""
	}
	var cppDef strings.Builder
	cppDef.WriteString("template<typename ")
	var cppSerie strings.Builder
	cppSerie.WriteByte('<')
	for i := range s.Generics {
		cppSerie.WriteByte('T')
		cppSerie.WriteString(strconv.Itoa(i))
		cppSerie.WriteByte(',')
	}
	serie = cppSerie.String()[:cppSerie.Len()-1] + ">"
	cppDef.WriteString(serie[1:])
	cppDef.WriteByte('\n')
	return cppDef.String(), serie
}

func gen_struct_operators(s *ast.Struct) string {
	outid := s.OutId()
	genericsDef, genericsSerie := gen_struct_generics(s)
	var cpp strings.Builder
	cpp.WriteString(indent_string())
	if l, _ := cpp.WriteString(genericsDef); l > 0 {
		cpp.WriteString(indent_string())
	}
	cpp.WriteString("inline bool operator==(const ")
	cpp.WriteString(outid)
	cpp.WriteString(genericsSerie)
	cpp.WriteString(" &_Src) {")
	if len(s.Defines.Globals) > 0 {
		add_indent()
		cpp.WriteByte('\n')
		cpp.WriteString(indent_string())
		var expr strings.Builder
		expr.WriteString("return ")
		add_indent()
		for _, g := range s.Defines.Globals {
			expr.WriteByte('\n')
			expr.WriteString(indent_string())
			expr.WriteString("this->")
			gid := g.OutId()
			expr.WriteString(gid)
			expr.WriteString(" == _Src.")
			expr.WriteString(gid)
			expr.WriteString(" &&")
		}
		done_indent()
		cpp.WriteString(expr.String()[:expr.Len()-3])
		cpp.WriteString(";\n")
		done_indent()
		cpp.WriteString(indent_string())
		cpp.WriteByte('}')
	} else {
		cpp.WriteString(" return true; }")
	}
	cpp.WriteString("\n\n")
	cpp.WriteString(indent_string())
	if l, _ := cpp.WriteString(genericsDef); l > 0 {
		cpp.WriteString(indent_string())
	}
	cpp.WriteString("inline bool operator!=(const ")
	cpp.WriteString(outid)
	cpp.WriteString(genericsSerie)
	cpp.WriteString(" &_Src) { return !this->operator==(_Src); }")
	return cpp.String()
}

func gen_struct_traits(s *ast.Struct) string {
	if len(s.Traits) == 0 {
		return ""
	}
	var cpp strings.Builder
	cpp.WriteString(": ")
	for _, t := range s.Traits {
		cpp.WriteString("public ")
		cpp.WriteString(t.OutId())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1]
}

func gen_struct_self_var(s *ast.Struct) string {
	var cpp strings.Builder
	cpp.WriteString(s.GetSelfRefVarType().String())
	cpp.WriteString(" self{ nil };")
	return cpp.String()
}

func gen_struct_self_var_init_st(s *ast.Struct) string {
	var cpp strings.Builder
	cpp.WriteString("this->self = ")
	cpp.WriteString(s.GetSelfRefVarType().String())
	cpp.WriteString("(this, nil);")
	return cpp.String()
}

func gen_struct_constructor(s *ast.Struct) string {
	var cpp strings.Builder
	cpp.WriteString(indent_string())
	cpp.WriteString(s.OutId())
	cpp.WriteString(gen_params(s.Constructor.Params))
	cpp.WriteString(" noexcept {\n")
	add_indent()
	cpp.WriteString(indent_string())
	cpp.WriteString(gen_struct_self_var_init_st(s))
	cpp.WriteByte('\n')
	if len(s.Defines.Globals) > 0 {
		for i, g := range s.Defines.Globals {
			cpp.WriteByte('\n')
			cpp.WriteString(indent_string())
			cpp.WriteString("this->")
			cpp.WriteString(g.OutId())
			cpp.WriteString(" = ")
			cpp.WriteString(s.Constructor.Params[i].OutId())
			cpp.WriteByte(';')
		}
	}
	done_indent()
	cpp.WriteByte('\n')
	cpp.WriteString(indent_string())
	cpp.WriteByte('}')
	return cpp.String()
}

func gen_struct_destructor(s *ast.Struct) string {
	var cpp strings.Builder
	cpp.WriteByte('~')
	cpp.WriteString(s.OutId())
	cpp.WriteString("(void) noexcept { /* heap allocations managed by traits or references */ this->self._ref = nil; }")
	return cpp.String()
}

func gen_struct_prototype(s *ast.Struct) string {
	var cpp strings.Builder
	cpp.WriteString(gen_generics(s.Generics))
	cpp.WriteByte('\n')
	cpp.WriteString("struct ")
	outid := s.OutId()
	cpp.WriteString(outid)
	cpp.WriteString(gen_struct_traits(s))
	cpp.WriteString(" {\n")
	add_indent()
	cpp.WriteString(indent_string())
	cpp.WriteString(gen_struct_self_var(s))
	cpp.WriteString("\n\n")
	if len(s.Defines.Globals) > 0 {
		for _, g := range s.Defines.Globals {
			cpp.WriteString(indent_string())
			cpp.WriteString(g.FieldString())
			cpp.WriteByte('\n')
		}
		cpp.WriteString("\n\n")
		cpp.WriteString(indent_string())
		cpp.WriteString(gen_struct_constructor(s))
		cpp.WriteString("\n\n")
	}
	cpp.WriteString(indent_string())
	cpp.WriteString(gen_struct_destructor(s))
	cpp.WriteString("\n\n")
	cpp.WriteString(indent_string())
	cpp.WriteString(outid)
	cpp.WriteString("(void) noexcept { ")
	cpp.WriteString(gen_struct_self_var_init_st(s))
	cpp.WriteString(" }\n\n")
	for _, f := range s.Defines.Fns {
		if f.Used {
			cpp.WriteString(indent_string())
			cpp.WriteString(gen_fn_prototype(f, ""))
			cpp.WriteString("\n\n")
		}
	}
	cpp.WriteString(gen_struct_operators(s))
	cpp.WriteByte('\n')
	done_indent()
	cpp.WriteString(indent_string())
	cpp.WriteString("};")
	return cpp.String()
}

func gen_struct_plain_prototype(s *ast.Struct) string {
	var cpp strings.Builder
	cpp.WriteString(gen_generics(s.Generics))
	cpp.WriteByte('\n')
	cpp.WriteString("struct ")
	cpp.WriteString(s.OutId())
	cpp.WriteByte(';')
	return cpp.String()
}

func gen_struct_decl(s *ast.Struct) string {
	var cpp strings.Builder
	for _, f := range s.Defines.Fns {
		if f.Used {
			cpp.WriteString(indent_string())
			cpp.WriteString(gen_fn_owner(f, s.OutId()))
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

func gen_struct(s *ast.Struct) string {
	var cpp strings.Builder
	cpp.WriteString(gen_struct_decl(s))
	cpp.WriteString("\n\n")
	cpp.WriteString(gen_struct_ostream(s))
	return cpp.String()
}

func gen_fn_decl_head(f *ast.Fn, owner string) string {
	var cpp strings.Builder
	cpp.WriteString(gen_generics(f.Generics))
	if cpp.Len() > 0 {
		cpp.WriteByte('\n')
		cpp.WriteString(indent_string())
	}
	if !f.IsEntryPoint {
		cpp.WriteString("inline ")
	}
	cpp.WriteString(f.RetType.String())
	cpp.WriteByte(' ')
	if owner != "" {
		cpp.WriteString(owner)
		cpp.WriteString(lex.KND_DBLCOLON)
	}
	cpp.WriteString(f.OutId())
	return cpp.String()
}

func gen_fn_head(f *ast.Fn, owner string) string {
	var cpp strings.Builder
	cpp.WriteString(gen_fn_decl_head(f, owner))
	cpp.WriteString(gen_params(f.Params))
	return cpp.String()
}

func gen_fn_owner(f *ast.Fn, owner string) string {
	var cpp strings.Builder
	cpp.WriteString(gen_fn_head(f, owner))
	cpp.WriteByte(' ')
	vars := f.RetType.Vars(f.Block)
	cpp.WriteString(gen_fn_block(vars, f.Block))
	return cpp.String()
}

func gen_fn_block(vars []*ast.Var, b *ast.Block) string {
	var cpp strings.Builder
	if vars != nil {
		statements := make([]ast.Statement, len(vars))
		for i, v := range vars {
			statements[i] = ast.Statement{Token: v.Token, Data: *v}
		}
		b.Tree = append(statements, b.Tree...)
	}
	cpp.WriteString(gen_block(b))
	return cpp.String()
}

func gen_fn(f *ast.Fn) string { return gen_fn_owner(f, "") }

func gen_fn_prototype(f *ast.Fn, owner string) string {
	var cpp strings.Builder
	cpp.WriteString(gen_fn_decl_head(f, owner))
	cpp.WriteString(f.PrototypeParams())
	cpp.WriteByte(';')
	return cpp.String()
}

func gen_links(used *[]*ast.UseDecl) string {
	var cpp strings.Builder
	for _, u := range *used {
		if u.Cpp {
			cpp.WriteString("#include ")
			if build.IsStdHeaderPath(u.Path) {
				cpp.WriteString(u.Path)
			} else {
				cpp.WriteByte('"')
				cpp.WriteString(u.Path)
				cpp.WriteByte('"')
			}
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func _gen_types(dm *ast.Defmap) string {
	var cpp strings.Builder
	for _, t := range dm.Types {
		if t.Used && t.Token.Id != lex.ID_NA {
			cpp.WriteString(gen_type_alias(t))
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func gen_types(tree *ast.Defmap, used *[]*ast.UseDecl) string {
	var cpp strings.Builder
	for _, u := range *used {
		if !u.Cpp {
			cpp.WriteString(_gen_types(u.Defines))
		}
	}
	cpp.WriteString(_gen_types(tree))
	return cpp.String()
}

func _gen_traits(dm *ast.Defmap) string {
	var cpp strings.Builder
	for _, t := range dm.Traits {
		if t.Used && t.Token.Id != lex.ID_NA {
			cpp.WriteString(gen_trait(t))
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

func gen_traits(tree *ast.Defmap, used *[]*ast.UseDecl) string {
	var cpp strings.Builder
	for _, u := range *used {
		if !u.Cpp {
			cpp.WriteString(_gen_traits(u.Defines))
		}
	}
	cpp.WriteString(_gen_traits(tree))
	return cpp.String()
}

func gen_structs(structs []*ast.Struct) string {
	var cpp strings.Builder
	for _, s := range structs {
		if s.Used && s.Token.Id != lex.ID_NA {
			cpp.WriteString(gen_struct(s))
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

func gen_struct_plain_prototypes(structs []*ast.Struct) string {
	var cpp strings.Builder
	for _, s := range structs {
		if s.Used && s.Token.Id != lex.ID_NA {
			cpp.WriteString(gen_struct_plain_prototype(s))
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func gen_struct_prototypes(structs []*ast.Struct) string {
	var cpp strings.Builder
	for _, s := range structs {
		if s.Used && s.Token.Id != lex.ID_NA {
			cpp.WriteString(gen_struct_prototype(s))
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func gen_fn_prototypes(dm *ast.Defmap) string {
	var cpp strings.Builder
	for _, f := range dm.Fns {
		if f.Used && f.Token.Id != lex.ID_NA {
			cpp.WriteString(gen_fn_prototype(f, ""))
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func gen_prototypes(tree *ast.Defmap, used *[]*ast.UseDecl, structs []*ast.Struct) string {
	var cpp strings.Builder
	cpp.WriteString(gen_struct_plain_prototypes(structs))
	cpp.WriteString(gen_struct_prototypes(structs))
	for _, u := range *used {
		if !u.Cpp {
			cpp.WriteString(gen_fn_prototypes(u.Defines))
		}
	}
	cpp.WriteString(gen_fn_prototypes(tree))
	return cpp.String()
}

func _gen_globals(dm *ast.Defmap) string {
	var cpp strings.Builder
	for _, g := range dm.Globals {
		if !g.Const && g.Used && g.Token.Id != lex.ID_NA {
			cpp.WriteString(g.String())
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func gen_globals(tree *ast.Defmap, used *[]*ast.UseDecl) string {
	var cpp strings.Builder
	for _, u := range *used {
		if !u.Cpp {
			cpp.WriteString(_gen_globals(u.Defines))
		}
	}
	cpp.WriteString(_gen_globals(tree))
	return cpp.String()
}

func _gen_fns(dm *ast.Defmap) string {
	var cpp strings.Builder
	for _, f := range dm.Fns {
		if f.Used && f.Token.Id != lex.ID_NA {
			cpp.WriteString(gen_fn(f))
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

func gen_fns(tree *ast.Defmap, used *[]*ast.UseDecl) string {
	var cpp strings.Builder
	for _, u := range *used {
		if !u.Cpp {
			cpp.WriteString(_gen_fns(u.Defines))
		}
	}
	cpp.WriteString(_gen_fns(tree))
	return cpp.String()
}

func gen_init_caller(tree *ast.Defmap, used *[]*ast.UseDecl) string {
	var cpp strings.Builder
	cpp.WriteString("void ")
	cpp.WriteString(init_caller)
	cpp.WriteString("(void) {")
	indent := "\t"
	push_init := func(defs *ast.Defmap) {
		f, dm, _ := defs.FnById(jule.INIT_FN, nil)
		if f == nil || dm != defs {
			return
		}
		cpp.WriteByte('\n')
		cpp.WriteString(indent)
		cpp.WriteString(f.OutId())
		cpp.WriteString("();")
	}
	for _, u := range *used {
		if !u.Cpp {
			push_init(u.Defines)
		}
	}
	push_init(tree)
	cpp.WriteString("\n}")
	return cpp.String()
}

func get_all_structs(tree *ast.Defmap, used *[]*ast.UseDecl) []*ast.Struct {
	order := make([]*ast.Struct, 0, len(tree.Structs))
	order = append(order, tree.Structs...)
	for _, u := range *used {
		if !u.Cpp {
			order = append(order, u.Defines.Structs...)
		}
	}
	return order
}

func gen_trait(t *ast.Trait) string {
	var cpp strings.Builder
	cpp.WriteString("struct ")
	outid := t.OutId()
	cpp.WriteString(outid)
	cpp.WriteString(" {\n")
	is := "\t"
	cpp.WriteString(is)
	cpp.WriteString("virtual ~")
	cpp.WriteString(outid)
	cpp.WriteString("(void) noexcept {}\n\n")
	for _, f := range t.Funcs {
		cpp.WriteString(is)
		cpp.WriteString("virtual ")
		cpp.WriteString(f.RetType.String())
		cpp.WriteByte(' ')
		cpp.WriteString(f.Id)
		cpp.WriteString(gen_params(f.Params))
		cpp.WriteString(" {")
		if !types.IsVoid(f.RetType.Type) {
			cpp.WriteString(" return {}; ")
		}
		cpp.WriteString("}\n")
	}
	cpp.WriteString("};")
	return cpp.String()
}

// Gen generates object code from parse tree.
func Gen(tree *ast.Defmap, used *[]*ast.UseDecl) string {
	structs := get_all_structs(tree, used)
	order_structures(structs)
	var cpp strings.Builder
	cpp.WriteString(gen_links(used))
	cpp.WriteByte('\n')
	cpp.WriteString(gen_types(tree, used))
	cpp.WriteByte('\n')
	cpp.WriteString(gen_traits(tree, used))
	cpp.WriteString(gen_prototypes(tree, used, structs))
	cpp.WriteString("\n\n")
	cpp.WriteString(gen_globals(tree, used))
	cpp.WriteString(gen_structs(structs))
	cpp.WriteString("\n\n")
	cpp.WriteString(gen_fns(tree, used))
	cpp.WriteString(gen_init_caller(tree, used))
	return cpp.String()
}

func can_be_order(s *ast.Struct) bool {
	for _, d := range s.Origin.Depends {
		if d.Origin.Order < s.Origin.Order {
			return false
		}
	}
	return true
}

func order_structures(structures []*ast.Struct) {
	for i, s := range structures {
		s.Order = i
	}

	n := len(structures)
	for i := 0; i < n; i++ {
		swapped := false
		for j := 0; j < n-i-1; j++ {
			curr := &structures[j]
			if can_be_order(*curr) {
				(*curr).Origin.Order = j + 1
				next := &structures[j+1]
				(*next).Origin.Order = j
				*curr, *next = *next, *curr
				swapped = true
			}
		}
		if !swapped {
			break
		}
	}
}
