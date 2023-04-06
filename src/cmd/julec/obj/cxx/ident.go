package cxx

import (
	"strconv"

	"github.com/julelang/jule"
	"github.com/julelang/jule/ast"
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
	case f.Ident == jule.ENTRY_POINT:
		return as_ident(f.Ident)

	case f.Is_method():
		return "_method_" + f.Ident

	default:
		return as_out_ident(f.Ident, f.Token.File.Addr())
	}
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
	return as_out_ident(f.Ident, f.Token.File.Addr())
}
