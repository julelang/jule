package cxx

import (
	"fmt"
	"strings"
	"time"

	"github.com/julelang/jule"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/sema"
)

// The self keyword equavalent of generated cpp.
const CPP_SELF = "this"
// C++ statement terminator.
const CPP_ST_TERM = ";"

// Current indention count.
var INDENT = 0

type _OrderedDecls struct {
	structs []*sema.Struct
	globals []*sema.Var
}

// Increase indentation.
func add_indent() { INDENT++ }
// Decrase indentation.
func done_indent() { INDENT-- }

// Returns indention string by INDENT.
func indent() string {
	const INDENT_KIND = "\t"
	if INDENT == 0 {
		return ""
	}

	s := ""
	for i := 0; i < INDENT; i++ {
		s += INDENT_KIND
	}
	return s
}

// Returns all structures of main package and used pakcages.
// Ignores cpp-linked declarations.
func get_all_structures(pkg *sema.Package, used []*sema.ImportInfo) []*sema.Struct {
	buffer := []*sema.Struct{}

	append_structs := func(p *sema.Package) {
		for _, f := range p.Files {
			for _, s := range f.Structs {
				if !s.Cpp_linked {
					buffer = append(buffer, s)
				}
			}
		}
	}

	append_structs(pkg)

	for _, u := range used {
		if !u.Cpp {
			append_structs(u.Package)
		}
	}

	return buffer
}

// Returns all variables of main package and used pakcages.
// Ignores cpp-linked declarations.
func get_all_variables(pkg *sema.Package, used []*sema.ImportInfo) []*sema.Var {
	buffer := []*sema.Var{}

	append_vars := func(p *sema.Package) {
		for _, f := range p.Files {
			for _, v := range f.Vars {
				if !v.Cpp_linked {
					buffer = append(buffer, v)
				}
			}
		}
	}

	append_vars(pkg)

	for _, u := range used {
		if !u.Cpp {
			append_vars(u.Package)
		}
	}

	return buffer
}

// Generates all C/C++ include directives.
func gen_links(used []*sema.ImportInfo) string {
	obj := ""
	for _, pkg := range used {
		switch {
		case !pkg.Cpp:
			continue

		case build.Is_std_header_path(pkg.Path):
			obj += "#include " + pkg.Path + "\n"

		case is_cpp_header_file(pkg.Path):
			obj += `#include "` + pkg.Path + "\"\n"
		}
	}
	return obj
}

// Generates C++ code of type aliase.
func gen_type_alias(ta *sema.TypeAlias) string {
	obj := "typedef "
	obj += gen_type_kind(ta.Kind.Kind)
	obj += " "
	obj += as_out_ident(ta.Ident, ta.Token.File.Addr())
	obj += CPP_ST_TERM
	return obj
}

// Generates C++ code of SymbolTable's all type aliases.
func gen_type_aliases_tbl(tbl *sema.SymbolTable) string {
	obj := ""
	for _, ta := range tbl.Type_aliases {
		if !ta.Cpp_linked {
			obj += gen_type_alias(ta) + "\n"
		}
	}
	return obj
}

// Generates C++ code of package's all type aliases.
func gen_type_aliases_pkg(pkg *sema.Package) string {
	obj := ""
	for _, tbl := range pkg.Files {
		obj += gen_type_aliases_tbl(tbl)
	}
	return obj
}

// Generates C++ code of all type aliases.
func gen_type_aliases(pkg *sema.Package, used []*sema.ImportInfo) string {
	obj := ""
	for _, u := range used {
		if !u.Cpp {
			obj += gen_type_aliases_pkg(u.Package)
		}
	}
	obj += gen_type_aliases_pkg(pkg)
	return obj
}

// Generates C++ code of function's result type.
func gen_fn_result(f *sema.Fn) string {
	if f.Is_void() {
		return "void"
	}
	return gen_type_kind(f.Result.Kind.Kind)
}

// Generates C++ code of function instance's result type.
func gen_fn_ins_result(f *sema.FnIns) string {
	if f.Decl.Is_void() {
		return "void"
	}
	return gen_type_kind(f.Result)
}

// Generates C++ prototype code of parameter.
func gen_param_prototype(p *sema.Param) string {
	obj := ""
	if p.Variadic {
		obj += as_jt("slice")
		obj += "<"
		obj += gen_type_kind(p.Kind.Kind)
		obj += ">"
	} else {
		obj += gen_type_kind(p.Kind.Kind)
	}
	return obj
}

// Generates C++ code of parameter instance.
func gen_param_ins(p *sema.ParamIns) string {
	obj := gen_param_ins_prototype(p)
	obj += " "
	obj += param_out_ident(p.Decl)
	return obj
}

// Generates C++ prototype code of parameter instance.
func gen_param_ins_prototype(p *sema.ParamIns) string {
	obj := ""
	if p.Decl.Variadic {
		obj += as_jt("slice")
		obj += "<"
		obj += gen_type_kind(p.Kind)
		obj += ">"
	} else {
		obj += gen_type_kind(p.Kind)
	}
	return obj
}

// Generates C++ code of parameter.
func gen_param(p *sema.Param) string {
	obj := gen_param_prototype(p)
	if p.Ident != "" && !lex.Is_ignore_ident(p.Ident) && !lex.Is_anon_ident(p.Ident) {
		obj += " " + param_out_ident(p)
	}
	return obj
}

// Generates C++ code of parameters.
func gen_params(params []*sema.Param) string {
	switch {
	case len(params) == 0:
		return "(void)"
	
	case len(params) == 1 && params[0].Is_self():
		return "(void)"
	}

	obj := "("
	for _, p := range params {
		if !p.Is_self() {
			obj += gen_param(p) + ","
		}
	}

	// Remove comma.
	obj = obj[:len(obj)-1]
	return obj + ")"
}

func gen_params_ins(params []*sema.ParamIns) string {
	switch {
	case len(params) == 0:
		return "(void)"
	
	case len(params) == 1 && params[0].Decl.Is_self():
		return "(void)"
	}

	obj := "("
	for _, p := range params {
		if !p.Decl.Is_self() {
			obj += gen_param_ins(p) + ","
		}
	}

	// Remove comma.
	obj = obj[:len(obj)-1]
	return obj + ")"
}

// Generates C++ declaration code of parameters.
func gen_params_prototypes(params []*sema.ParamIns) string {
	switch {
	case len(params) == 0:
		return "(void)"
	
	case len(params) == 1 && params[0].Decl.Is_self():
		return "(void)"
	}

	obj := "("
	for _, p := range params {
		if !p.Decl.Is_self() {
			obj += gen_param_ins_prototype(p) + ","
		}
	}

	// Remove comma.
	obj = obj[:len(obj)-1]
	return obj + ")"
}

// Generates C++ code of trait.
func gen_trait(t *sema.Trait) string {
	const INDENTION = "\t"
	outid := trait_out_ident(t)

	obj := "struct "
	obj += outid
	obj += " {\n"
	obj += INDENTION
	obj += "virtual ~"
	obj += outid
	obj += "(void) noexcept {}\n\n"
	for _, f := range t.Methods {
		obj += INDENTION
		obj += "virtual "
		obj += gen_fn_result(f)
		obj += " _method_"
		obj += f.Ident
		obj += gen_params(f.Params)
		obj += " {"
		if !f.Is_void() {
			obj += " return {}; "
		}
		obj += "}\n"
	}
	obj += "};"
	return obj
}

// Generates C++ code of SymbolTable's all traits.
func gen_traits_tbl(tbl *sema.SymbolTable) string {
	obj := ""
	for _, t := range tbl.Traits {
		obj += gen_trait(t) + "\n\n"
	}
	return obj
}

// Generates C++ code of package's all traits.
func gen_traits_pkg(pkg *sema.Package) string {
	obj := ""
	for _, tbl := range pkg.Files {
		obj += gen_traits_tbl(tbl)
	}
	return obj
}

// Generates C++ code of all traits.
func gen_traits(pkg *sema.Package, used []*sema.ImportInfo) string {
	obj := ""
	for _, u := range used {
		if !u.Cpp {
			obj += gen_traits_pkg(u.Package)
		}
	}
	obj += gen_traits_pkg(pkg)
	return obj
}

// Generates C++ plain-prototype code of structure.
func gen_struct_plain_prototype(s *sema.Struct) string {
	obj := ""
	for _, ins := range s.Instances {
		obj += "\nstruct "
		obj += struct_ins_out_ident(ins)
		obj += CPP_ST_TERM
		obj += "\n"
	}
	return obj
}

// Generates C++ plain-prototype code of all structures.
func gen_struct_plain_prototypes(structs []*sema.Struct) string {
	obj := ""
	for _, s := range structs {
		if !s.Cpp_linked && s.Token.Id != lex.ID_NA {
			obj += gen_struct_plain_prototype(s) + "\n"
		}
	}
	return obj
}

// Generates C++ derive code of structure's implemented traits.
func gen_struct_traits(s *sema.Struct) string {
	if len(s.Implements) == 0 {
		return ""
	}

	obj := ": "
	for _, i := range s.Implements {
		obj += "public "
		obj += trait_out_ident(i)
		obj += ","
	}
	obj = obj[:len(obj)-1] // Remove last comma.
	return obj
}

func gen_struct_self_field_type_kind(s *sema.StructIns) string {
	return as_ref_kind(gen_struct_kind_ins(s))
}

// Generates C++ field declaration code of structure's self field.
func gen_struct_self_field(s *sema.StructIns) string {
	obj := gen_struct_self_field_type_kind(s)
	obj += " self{};"
	return obj
}

// Generates C++ declaration code of field.
func gen_field_decl(f *sema.FieldIns) string {
	obj := gen_type_kind(f.Kind) + " "
	obj += field_out_ident(f.Decl)
	obj += get_init_expr(f.Kind)
	obj += CPP_ST_TERM
	return obj
}

func gen_struct_self_field_init_st(s *sema.StructIns) string {
	obj := "this->self = "
	obj += gen_struct_self_field_type_kind(s)
	obj += "::make(this, nullptr);"
	return obj
}

func gen_struct_constructor(s *sema.StructIns) string {
	obj := struct_ins_out_ident(s)

	obj += "("
	if len(s.Fields) > 0 {
		for _, f := range s.Fields {
			obj += gen_type_kind(f.Kind)
			obj += " __param_" + f.Decl.Ident + ", "
		}
		obj = obj[:len(obj)-2] // Remove last comma.
	} else {
		obj += "void"
	}

	obj += ") noexcept {\n"
	add_indent()
	obj += indent()
	obj += gen_struct_self_field_init_st(s)
	obj += "\n"

	if len(s.Fields) > 0 {
		for _, f := range s.Fields {
			obj += "\n"
			obj += indent()
			obj += "this->"
			obj += field_out_ident(f.Decl)
			obj += " = "
			obj += "__param_" + f.Decl.Ident
			obj += CPP_ST_TERM
		}
	}

	done_indent()
	obj += "\n" + indent() + "}"
	return obj
}

func gen_struct_destructor(s *sema.StructIns) string {
	obj := "~"
	obj += struct_ins_out_ident(s)
	obj += "(void) noexcept { /* heap allocations managed by traits or references */ this->self.ref = nullptr; }"
	return obj
}

func gen_struct_operators(s *sema.StructIns) string {
	out_ident := struct_ins_out_ident(s)
	obj := ""

	obj += indent()
	obj += "inline bool operator==(const "
	obj += out_ident
	obj += " &_Src) {"
	if len(s.Fields) > 0 {
		add_indent()
		obj += "\n"
		obj += indent()
		obj += "return "
		add_indent()
		for _, f := range s.Fields {
			obj += "\n"
			obj += indent()
			obj += "this->"
			f_ident := field_out_ident(f.Decl)
			obj += f_ident
			obj += " == _Src."
			obj += f_ident
			obj += " &&"
		}
		done_indent()
		obj = obj[:len(obj)-3] // Remove last suffix " &&"
		obj += ";\n"
		done_indent()
		obj += indent()
		obj += "}"
	} else {
		obj += " return true; }"
	}
	obj += "\n\n"
	obj += indent()
	obj += "inline bool operator!=(const "
	obj += out_ident
	obj += " &_Src) { return !this->operator==(_Src); }"
	return obj
}

func gen_struct_derive_defs_prototypes(s *sema.StructIns) string {
	obj := ""

	if s.Decl.Is_derives(build.DERIVE_CLONE) {
		obj += indent()
		obj += get_derive_fn_decl_clone(s.Decl)
		obj += ";\n\n"
	}

	return obj
}

func gen_struct_ins_prototype(s *sema.StructIns) string {
	obj := "struct "
	out_ident := struct_ins_out_ident(s)
	obj += out_ident
	obj += gen_struct_traits(s.Decl)
	obj += " {\n"

	add_indent()
	obj += indent()
	obj += gen_struct_self_field(s)
	obj += "\n\n"
	if len(s.Fields) > 0 {
		for _, f := range s.Fields {
			obj += indent()
			obj += gen_field_decl(f)
			obj += "\n"
		}
		obj += "\n\n"
		obj += indent()
		obj += gen_struct_constructor(s)
		obj += "\n\n"
	}

	obj += indent()
	obj += gen_struct_destructor(s)
	obj += "\n\n"

	obj += indent()
	obj += out_ident
	obj += "(void) noexcept { "
	obj += gen_struct_self_field_init_st(s)
	obj += " }\n\n"

	for _, f := range s.Methods {
		obj += gen_fn_prototype(f, true)
		obj += "\n\n"
	}

	obj += gen_struct_derive_defs_prototypes(s)

	obj += gen_struct_operators(s)
	obj += "\n"

	done_indent()
	obj += indent() + "};"
	return obj
}

// Generates C++ declaration code of structure.
func gen_struct_prototype(s *sema.Struct) string {
	obj := ""
	for _, ins := range s.Instances {
		obj += gen_struct_ins_prototype(ins) + "\n\n"
	}
	return obj
}

// Generates C++ declaration code of all structures.
func gen_struct_prototypes(structs []*sema.Struct) string {
	obj := ""
	for _, s := range structs {
		if !s.Cpp_linked && s.Token.Id != lex.ID_NA {
			obj += gen_struct_prototype(s) + "\n"
		}
	}
	return obj
}

func gen_fn_decl_head(f *sema.FnIns, method bool) string {
	obj := ""
	if !f.Decl.Is_entry_point() {
		obj += "inline "
	}

	obj += gen_fn_ins_result(f) + " "

	if !method && f.Decl.Owner != nil {
		obj += struct_ins_out_ident(f.Owner) + lex.KND_DBLCOLON
	}
	obj += fn_ins_out_ident(f)
	return obj
}

// Generates C++ declaration code of function's combinations.
func gen_fn_prototype(f *sema.Fn, method bool) string {
	obj := ""
	for _, c := range f.Instances {
		obj += indent()
		obj += gen_fn_decl_head(c, method)
		obj += gen_params_prototypes(c.Params)
		obj += CPP_ST_TERM + "\n"
	}
	return obj
}

// Generates C++ declaration code of all functions.
func gen_fn_prototypes(pkg *sema.Package) string {
	obj := ""
	for _, file := range pkg.Files {
		for _, f := range file.Funcs {
			if !f.Cpp_linked && f.Token.Id != lex.ID_NA {
				obj += gen_fn_prototype(f, false)
			}
		}
	}
	return obj
}

// Generates C++ code of all can-be-prototyped declarations.
func gen_prototypes(pkg *sema.Package, used []*sema.ImportInfo, structs []*sema.Struct) string {
	obj := ""

	obj += gen_struct_plain_prototypes(structs)
	obj += gen_struct_prototypes(structs)

	for _, u := range used {
		if !u.Cpp {
			obj += gen_fn_prototypes(u.Package)
		}
	}
	obj += gen_fn_prototypes(pkg)

	return obj
}

// Generates C++ code of variable.
func gen_var(v *sema.Var) string {
	if lex.Is_ignore_ident(v.Ident) {
		return ""
	}
	if v.Constant {
		return ""
	}

	obj := gen_type_kind(v.Kind.Kind) + " "
	obj += var_out_ident(v)
	if v.Value != nil && v.Value.Expr != nil {
		if v.Value.Data.Model != nil {
			obj += " = "
			obj += gen_val(v.Value)
		} else {
			obj += CPP_DEFAULT_EXPR
		}
	} else {
		obj += get_init_expr(v.Kind.Kind)
	}
	obj += CPP_ST_TERM
	return obj
}

// Generates C++ code of all globals.
func gen_globals(globals []*sema.Var) string {
	obj := ""

	for _, v := range globals {
		if !v.Constant && v.Token.Id != lex.ID_NA {
			obj += gen_var(v) + "\n"
		}
	}

	return obj
}

// Generates C++ code of function.
func gen_fn(f *sema.Fn) string {
	obj := ""
	for _, c := range f.Instances {
		obj += gen_fn_decl_head(c, false)
		obj += gen_params_ins(c.Params) + " "
		obj += gen_fn_scope(c)
		obj += "\n\n"
	}
	return obj
}

// Generates C++ code of all functions of package.
func gen_pkg_fns(p *sema.Package) string {
	obj := ""
	for _, f := range p.Files {
		for _, f := range f.Funcs {
			if !f.Cpp_linked && f.Token.Id != lex.ID_NA {
				obj += gen_fn(f) + "\n\n"
			}
		}
	}
	return obj
}

// Generates C++ code of structure's methods.
func gen_struct_method_defs(s *sema.StructIns) string {
	obj := ""
	for _, f := range s.Methods {
		obj += indent()
		obj += gen_fn(f)
		obj += "\n\n"
	}
	return obj
}

// Generates C++ code of structure's ostream.
func gen_struct_ostream(s *sema.StructIns) string {
	obj := ""
	obj += indent()
	obj += "std::ostream &operator<<(std::ostream &_Stream, const "
	obj += struct_ins_out_ident(s)
	obj += " &_Src) {\n"
	add_indent()
	obj += indent()
	obj += `_Stream << "`
	obj += s.Decl.Ident
	obj += "{\";\n"

	for i, field := range s.Fields {
		obj += indent()
		obj += `_Stream << "`
		obj += field.Decl.Ident
		obj += `:" << _Src.`
		obj += field_out_ident(field.Decl)
		if i+1 < len(s.Fields) {
			obj += " << \", \""
		}
		obj += ";\n"
	}

	obj += indent()
	obj += "_Stream << \"}\";\n"
	obj += indent()
	obj += "return _Stream;\n"
	done_indent()
	obj += indent()
	obj += "}"
	return obj
}

func gen_struct_derive_defs(s *sema.StructIns) string {
	obj := ""

	if s.Decl.Is_derives(build.DERIVE_CLONE) {
		obj += indent()
		obj += get_derive_fn_def_clone(s.Decl)
		obj += "{\n"
		add_indent()
		obj += indent()
		obj += gen_struct_kind_ins(s)
		obj += " clone;\n"
		for _, f := range s.Fields {
			ident := field_out_ident(f.Decl)

			obj += indent()
			obj += "clone."
			obj += ident
			obj += " = jule::clone(this->"
			obj += ident
			obj += ");\n"
		}
		obj += indent()
		obj += "return clone;\n"
		done_indent()
		obj += indent()
		obj += "}"
	}

	return obj
}

// Generates C++ code of structure instance definition.
func gen_struct_ins(s *sema.StructIns) string {
	obj := gen_struct_method_defs(s)
	obj += "\n\n"
	obj += gen_struct_derive_defs(s)
	obj += "\n\n"
	obj += gen_struct_ostream(s)
	return obj
}

// Generates C++ code of structure definition.
func gen_struct(s *sema.Struct) string {
	obj := ""
	for _, ins := range s.Instances {
		obj += gen_struct_ins(ins) + "\n\n"
	}
	return obj
}

// Generates C++ code of all structures.
func gen_structs(structs []*sema.Struct) string {
	obj := ""
	for _, s := range structs {
		if !s.Cpp_linked && s.Token.Id != lex.ID_NA {
			obj += gen_struct(s)
			obj += "\n\n"
		}
	}
	return obj
}

// Generates C++ code of all functions.
func gen_fns(pkg *sema.Package, used []*sema.ImportInfo) string {
	obj := ""

	for _, u := range used {
		if !u.Cpp {
			obj += gen_pkg_fns(u.Package)
		}
	}
	obj += gen_pkg_fns(pkg)

	return obj
}

// Generated C++ code of all initializer functions.
func gen_init_caller(pkg *sema.Package, used []*sema.ImportInfo) string {
	const INDENTION = "\t"

	obj := "void "
	obj += INIT_CALLER_IDENT
	obj += "(void) {"

	push_init := func(pkg *sema.Package) {
		const CPP_LINKED = false
		f := pkg.Find_fn(build.INIT_FN, CPP_LINKED)
		if f == nil {
			return
		}

		obj += "\n" + INDENTION + fn_out_ident(f) + "();"
	}

	for _, u := range used {
		if !u.Cpp {
			push_init(u.Package)
		}
	}
	push_init(pkg)

	obj += "\n}"
	return obj
}

func append_standard(obj_code *string, compiler string, compiler_cmd string) {
	y, m, d := time.Now().Date()
	h, min, _ := time.Now().Clock()
	timeStr := fmt.Sprintf("%d/%d/%d %d.%d (DD/MM/YYYY) (HH.MM)", d, m, y, h, min)
	var sb strings.Builder
	sb.WriteString("// Auto generated by JuleC.\n")
	sb.WriteString("// JuleC version: ")
	sb.WriteString(jule.VERSION)
	sb.WriteByte('\n')
	sb.WriteString("// Date: ")
	sb.WriteString(timeStr)
	sb.WriteString(`
//
// Recommended Compile Command;
// `)
	sb.WriteString(compiler)
	sb.WriteByte(' ')
	sb.WriteString(compiler_cmd)
	sb.WriteString("\n\n#include \"")
	sb.WriteString(build.PATH_API)
	sb.WriteString("\"\n\n")
	sb.WriteString(*obj_code)
	sb.WriteString(`
int main(int argc, char *argv[]) {
	std::set_terminate(&jule::terminate_handler);
	jule::set_sig_handler(jule::signal_handler);
	jule::setup_command_line_args(argc, argv);
	__jule_call_initializers();
	entry_point();

	return EXIT_SUCCESS;
}`)
	*obj_code = sb.String()
}

// Generates C++ codes from SymbolTables.
func Gen(pkg *sema.Package, used []*sema.ImportInfo) string {
	od := &_OrderedDecls{}
	od.structs = get_all_structures(pkg, used)
	order_structures(od.structs)

	od.globals = get_all_variables(pkg, used)
	order_variables(od.globals)

	obj := ""
	obj += gen_links(used) + "\n"
	obj += gen_type_aliases(pkg, used) + "\n"
	obj += gen_traits(pkg, used) + "\n"
	obj += gen_prototypes(pkg, used, od.structs) + "\n\n"
	obj += gen_globals(od.globals) + "\n"
	obj += gen_structs(od.structs)
	obj += gen_fns(pkg, used) + "\n"
	obj += gen_init_caller(pkg, used) + "\n"
	return obj
}
