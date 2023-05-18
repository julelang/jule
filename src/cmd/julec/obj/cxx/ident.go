package cxx

import (
	"strconv"

	"github.com/julelang/jule"
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/sema"
)

// Extension of Jule data types.
const TYPE_EXT = "_jt"

// Identifier of initialize function caller function.
const INIT_CALLER_IDENT = "__julec_call_initializers"

// Returns specified identifer as JuleC identifer.
// Equavalents: "JULEC_ID(" + ident + ")" of JuleC API.
func as_ident(ident string) string { return "_" + ident }

// Returns given identifier as Jule type identifier.
func as_jt(id string) string { return id + TYPE_EXT }

// Returns cpp output identifier form of pointer address.
func get_ptr_as_ident(ptr uintptr) string {
	addr := "_" + strconv.FormatUint(uint64(ptr), 16)
	for i, r := range addr {
		if r != '0' {
			addr = addr[i:]
			break
		}
	}
	return addr
}

// Returns cpp output identifier form of given identifier.
//
// Parameters:
//  - ident: Identifier.
//  - ptr:   Pointer address of package file handler.
func as_out_ident(ident string, ptr uintptr) string {
	if ptr != 0 {
		return get_ptr_as_ident(ptr) + "_" + ident
	}
	return as_ident(ident)
}

// Returns cpp output local identifier form of fiven identifier.
//
// Parameters:
//  - row:   Row of definition.
//  - col:   Column of definition.
//  - ident: Identifier of definition.
func as_local_ident(row int, col int, ident string) string {
	ident = strconv.Itoa(row) + strconv.Itoa(col) + "_" + ident
	return as_ident(ident)
}

// Returns output identifier of function.
func fn_out_ident(f *sema.Fn) string {
	switch {
	case f.Cpp_linked:
		return f.Ident

	case f.Ident == jule.ENTRY_POINT:
		return as_ident(f.Ident)

	case f.Owner != nil:
		return "_method_" + f.Ident

	default:
		return as_out_ident(f.Ident, f.Token.File.Addr())
	}
}

func fn_ins_out_ident(f *sema.FnIns) string {
	if f.Decl.Cpp_linked || len(f.Generics) == 0 || f.Decl.Parameters_uses_generics() {
		return fn_out_ident(f.Decl)
	}

	kind := f.To_str()
	for i, ins := range f.Decl.Instances {
		if kind == ins.To_str() {
			return fn_out_ident(f.Decl) + "_" + strconv.Itoa(i)
		}
	}

	return "__?__"
}

// Returns output identifier of trait.
func trait_out_ident(t *sema.Trait) string {
	return as_out_ident(t.Ident, t.Token.File.Addr())
}

// Returns output identifier of parameter.
func param_out_ident(p *sema.Param) string {
	return as_local_ident(p.Token.Row, p.Token.Column, p.Ident)
}

// Returns output identifier of structure.
func struct_out_ident(s *sema.Struct) string {
	if s.Cpp_linked {
		return s.Ident
	}
	return as_out_ident(s.Ident, s.Token.File.Addr())
}

// Returns output identifier of generic type declaration.
func generic_decl_out_ident(g *ast.Generic) string {
	return as_ident(g.Ident)
}

// Returns output identifier of field.
func field_out_ident(f *sema.Field) string {
	return "_field_" + f.Ident
}

// Returns output identifier of variable.
func var_out_ident(v *sema.Var) string {
	switch {
	case v.Cpp_linked:
		return v.Ident

	case v.Ident == lex.KND_SELF:
		return "self"

	case v.Scope != nil:
		return as_local_ident(v.Token.Row, v.Token.Column, v.Ident)

	default:
		return as_out_ident(v.Ident, v.Token.File.Addr())
	}
}

// Returns begin label identifier of iteration.
func iter_begin_label_ident(it uintptr) string {
	return "_iter_begin_" + strconv.Itoa(int(it))
}

// Returns end label identifier of iteration.
func iter_end_label_ident(it uintptr) string {
	return "_iter_end_" + strconv.Itoa(int(it))
}

// Returns next label identifier of iteration.
func iter_next_label_ident(it uintptr) string {
	return "_iter_next_" + strconv.Itoa(int(it))
}

// Returns label identifier.
func label_ident(ident string) string {
	return "_julec_label_" + ident
}

// Returns end label identifier of match-case.
func match_end_label_ident(m uintptr) string {
	return "_match_end_" + strconv.Itoa(int(m))
}

// Returns begin label identifier of case.
func case_begin_label_ident(c uintptr) string {
	return "_case_begin_" + strconv.Itoa(int(c))
}

// Returns end label identifier of case.
func case_end_label_ident(c uintptr) string {
	return "_case_end_" + strconv.Itoa(int(c))
}
