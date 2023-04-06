package cxx

import "github.com/julelang/jule/sema"

// Ignore expression for std::tie function.
const CPP_IGNORE = "std::ignore"

// Represents default expression for type.
const CPP_DEFAULT_EXPR = "{}"

func gen_expr(d *sema.Value) string {
	return ""
}

func get_init_expr(t *sema.TypeKind) string {
	enm := t.Enm()
	if enm == nil {
		return CPP_DEFAULT_EXPR
	}
	return "{" + gen_expr(enm.Items[0].Value) + "}"
}
