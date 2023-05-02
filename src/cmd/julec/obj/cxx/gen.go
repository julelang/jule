package cxx

import (
	"github.com/julelang/jule"
	"github.com/julelang/jule/ast"
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
func get_all_structures(pkg *sema.Package, used []*sema.Package) []*sema.Struct {
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

	for _, p := range used {
		append_structs(p)
	}

	return buffer
}

// Generates all C/C++ include directives.
func gen_links(used []*sema.Package) string {
	obj := ""
	for _, pkg := range used {
		if !pkg.Cpp {
			continue
		}

		obj += "#include "
		if build.Is_std_header_path(pkg.Path) {
			obj += pkg.Path
		} else {
			obj += `"` + pkg.Path + `"`
		}
		obj += "\n"
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
func gen_type_aliases(pkg *sema.Package, used []*sema.Package) string {
	obj := ""
	for _, pkg := range used {
		if !pkg.Cpp {
			obj += gen_type_aliases_pkg(pkg)
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
		obj += " "
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
func gen_traits(pkg *sema.Package, used []*sema.Package) string {
	obj := ""
	for _, pkg := range used {
		if !pkg.Cpp {
			obj += gen_traits_pkg(pkg)
		}
	}
	obj += gen_traits_pkg(pkg)
	return obj
}

// Generates C++ declaration code of generic.
func gen_generic_decl(g *ast.Generic) string {
	obj := "typename "
	obj += generic_decl_out_ident(g)
	return obj
}

// Generates C++ declaration code of all generics.
func gen_generic_decls(generics []*ast.Generic) string {
	if len(generics) == 0 {
		return ""
	}

	obj := "template<"
	for _, g := range generics {
		obj += gen_generic_decl(g) + ","
	}

	obj = obj[:len(obj)-1] // Remove last comma.
	obj += ">"
	return obj
}

// Generates C++ plain-prototype code of structure.
func gen_struct_plain_prototype(s *sema.Struct) string {
	obj := ""
	obj += gen_generic_decls(s.Generics)
	obj += "\nstruct "
	obj += struct_out_ident(s)
	obj += CPP_ST_TERM
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

// Generates C++ code of structure generics.
// Returns declaration and definition code.
func gen_struct_generics(generics []*ast.Generic) (decl string, def string) {
	if len(generics) == 0 {
		return "", ""
	}

	decl = "template<"
	def = "<"

	for _, g := range generics {
		decl += gen_generic_decl(g)
		decl += ","

		def += generic_decl_out_ident(g)
		def += ","
	}

	decl = decl[:len(decl)-1] // Remove last comma.
	decl += ">\n"

	def = def[:len(def)-1] // Remove last comma.
	def += ">"
	return
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

func gen_struct_self_field_type_kind(s *sema.Struct) string {
	return as_ref_kind(gen_struct_kind(s))
}

// Generates C++ field declaration code of structure's self field.
func gen_struct_self_field(s *sema.Struct) string {
	obj := gen_struct_self_field_type_kind(s)
	obj += " self{};"
	return obj
}

// Generates C++ declaration code of field.
func gen_field_decl(f *sema.Field) string {
	obj := gen_type_kind(f.Kind.Kind) + " "
	obj += field_out_ident(f)
	obj += get_init_expr(f.Kind.Kind)
	obj += CPP_ST_TERM
	return obj
}

func gen_struct_self_field_init_st(s *sema.Struct) string {
	obj := "this->self = "
	obj += gen_struct_self_field_type_kind(s)
	obj += "::make(this, nil);"
	return obj
}

func gen_struct_constructor(s *sema.Struct) string {
	obj := struct_out_ident(s)

	obj += "("
	if len(s.Fields) > 0 {
		for _, f := range s.Fields {
			obj += gen_type_kind(f.Kind.Kind)
			obj += " __param_" + f.Ident + ", "
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
			obj += field_out_ident(f)
			obj += " = "
			obj += "__param_" + f.Ident
			obj += CPP_ST_TERM
		}
	}

	done_indent()
	obj += "\n" + indent() + "}"
	return obj
}

func gen_struct_destructor(s *sema.Struct) string {
	obj := "~"
	obj += struct_out_ident(s)
	obj += "(void) noexcept { /* heap allocations managed by traits or references */ this->self.__ref = nil; }"
	return obj
}

func gen_struct_operators(s *sema.Struct) string {
	out_ident := struct_out_ident(s)
	_, def := gen_struct_generics(s.Generics)
	obj := ""

	obj += indent()
	obj += "inline bool operator==(const "
	obj += out_ident
	obj += def
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
			f_ident := field_out_ident(f)
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
	obj += def
	obj += " &_Src) { return !this->operator==(_Src); }"
	return obj
}

// Generates C++ declaration code of structure.
func gen_struct_prototype(s *sema.Struct) string {
	obj := gen_generic_decls(s.Generics) + "\n"
	obj += "struct "
	out_ident := struct_out_ident(s)
	obj += out_ident
	obj += gen_struct_traits(s)
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
		obj += indent()
		f.Owner = nil // Ignore structure identifier prefix.
		obj += gen_fn_prototype(f)
		f.Owner = s
		obj += "\n\n"
	}

	obj += gen_struct_operators(s)
	obj += "\n"

	done_indent()
	obj += indent() + "};"
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

func gen_fn_decl_head(f *sema.FnIns) string {
	obj := ""

	if f.Decl.Owner != nil {
		generics := gen_generic_decls(f.Decl.Owner.Generics)
		if generics != "" {
			obj += generics
			obj += "\n" + indent()
		}
	}

	generics := gen_generic_decls(f.Decl.Generics)
	if generics != "" {
		obj += generics
		obj += "\n" + indent()
	}

	if !f.Decl.Is_entry_point() {
		obj += "inline "
	}
	obj += gen_fn_ins_result(f) + " "

	if f.Decl.Owner != nil {
		_, def := gen_struct_generics(f.Decl.Owner.Generics)
		obj += struct_out_ident(f.Decl.Owner)
		obj += def + lex.KND_DBLCOLON
	}
	obj += fn_out_ident(f.Decl)
	return obj
}

// Generates C++ declaration code of function's combinations.
func gen_fn_prototype(f *sema.Fn) string {
	obj := ""
	for _, c := range f.Combines {
		obj += gen_fn_decl_head(c)
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
				obj += gen_fn_prototype(f)
			}
		}
	}
	return obj
}

// Generates C++ code of all can-be-prototyped declarations.
func gen_prototypes(pkg *sema.Package, used []*sema.Package, structs []*sema.Struct) string {
	obj := ""

	obj += gen_struct_plain_prototypes(structs)
	obj += gen_struct_prototypes(structs)

	for _, p := range used {
		if !p.Cpp {
			obj += gen_fn_prototypes(p)
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
	if v.Value.Expr != nil {
		obj += " = "
		obj += gen_expr(v.Value)
	} else {
		obj += get_init_expr(v.Kind.Kind)
	}
	obj += CPP_ST_TERM
	return obj
}

// Generates C++ code of all globals of package.
func gen_pkg_globals(p *sema.Package) string {
	obj := ""
	for _, f := range p.Files {
		for _, v := range f.Vars {
			if !v.Constant && v.Token.Id != lex.ID_NA {
				obj += gen_var(v) + "\n"
			}
		}
	}
	return obj
}

// Generates C++ code of all globals.
func gen_globals(pkg *sema.Package, used []*sema.Package) string {
	obj := ""

	for _, p := range used {
		if !p.Cpp {
			obj += gen_pkg_globals(p)
		}
	}
	obj += gen_pkg_globals(pkg)

	return obj
}

// Generates C++ code of function.
func gen_fn(f *sema.Fn) string {
	obj := ""
	for _, c := range f.Combines {
		obj += gen_fn_decl_head(c)
		obj += gen_params_ins(c.Params) + " "
		obj += gen_fn_scope(c)
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
func gen_struct_method_defs(s *sema.Struct) string {
	obj := ""
	for _, f := range s.Methods {
		obj += indent()
		obj += gen_fn(f)
		obj += "\n\n"
	}
	return obj
}

// Generates C++ code of structure's ostream.
func gen_struct_ostream(s *sema.Struct) string {
	obj := ""
	generics_decl, generics_def := gen_struct_generics(s.Generics)
	obj += indent()
	if generics_decl != "" {
		obj += generics_decl + "\n"
		obj += indent()
	}
	obj += "std::ostream &operator<<(std::ostream &_Stream, const "
	obj += struct_out_ident(s)
	obj += generics_def
	obj += " &_Src) {\n"
	add_indent()
	obj += indent()
	obj += `_Stream << "`
	obj += struct_out_ident(s)
	obj += "{\";\n"

	for i, field := range s.Fields {
		obj += indent()
		obj += `_Stream << "`
		obj += field.Ident
		obj += `:" << _Src.`
		obj += field_out_ident(field)
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

// Generates C++ code of structure definition.
func gen_struct(s *sema.Struct) string {
	obj := ""
	obj += gen_struct_method_defs(s)
	obj += "\n\n"
	obj += gen_struct_ostream(s)
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
func gen_fns(pkg *sema.Package, used []*sema.Package) string {
	obj := ""

	for _, p := range used {
		if !p.Cpp {
			obj += gen_pkg_fns(p)
		}
	}
	obj += gen_pkg_fns(pkg)

	return obj
}

// Generated C++ code of all initializer functions.
func gen_init_caller(pkg *sema.Package, used []*sema.Package) string {
	const INDENTION = "\t"

	obj := "void "
	obj += INIT_CALLER_IDENT
	obj += "(void) {"

	push_init := func(pkg *sema.Package) {
		const CPP_LINKED = false
		f := pkg.Find_fn(jule.INIT_FN, CPP_LINKED)
		if f == nil {
			return
		}

		obj += "\n" + INDENTION + fn_out_ident(f) + "();"
	}

	for _, u := range used {
		if !u.Cpp {
			push_init(u)
		}
	}
	push_init(pkg)

	obj += "\n}"
	return obj
}

// Generates C++ codes from SymbolTables.
func Gen(pkg *sema.Package, used []*sema.Package) string {
	structs := get_all_structures(pkg, used)
	order_structures(structs)

	obj := ""
	obj += gen_links(used) + "\n"
	obj += gen_type_aliases(pkg, used) + "\n"
	obj += gen_traits(pkg, used) + "\n"
	obj += gen_prototypes(pkg, used, structs) + "\n\n"
	obj += gen_globals(pkg, used) + "\n"
	obj += gen_structs(structs)
	obj += gen_fns(pkg, used) + "\n"
	obj += gen_init_caller(pkg, used) + "\n"
	return obj
}
