package cxx

import (
	"github.com/julelang/jule"
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/sema"
)

// Ignore expression for std::tie function.
const CPP_IGNORE = "std::ignore"

// The self keyword equavalent of generated cpp.
const CPP_SELF = "this"

// Represents default expression for type.
const CPP_DEFAULT_EXPR = "{}"

// C++ statement terminator.
const CPP_ST_TERM = ";"

// Current indention count.
var INDENT = 0

// Returns indention string by INDENT.
func indent() string {
	const INDENT_KIND = "\t"
	if INDENT == 0 {
		return ""
	}

	s := ""
	for i := 0; i < INDENT; i-- {
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

// Generates C++ code of all can-be-prototyped declarations.
func gen_prototypes(pkg *sema.Package, used []*sema.Package, structs []*sema.Struct) string {
	obj := ""

	obj += gen_struct_plain_prototypes(structs)
	/*
	TODO: Implement here:


	obj += gen_struct_prototypes(structs)

	for _, p := range used {
		if !p.Cpp {
			obj += gen_fn_prototypes(p)
		}
	}
	obj += gen_fn_prototypes(pkg)
	*/

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
	obj += gen_init_caller(pkg, used) + "\n"
	return obj
}
