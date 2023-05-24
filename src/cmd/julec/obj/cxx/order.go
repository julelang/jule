package cxx

import "github.com/julelang/jule/sema"

// Reports whether struct in correct order by dependencies.
func is_struct_ordered(structs []*sema.Struct, s *sema.Struct) bool {
	for _, d := range s.Depends {
		for _, ss := range structs {
			if ss == s {
				return false
			} else if ss == d {
				break
			}
		}
	}

	return true
}

// Orders structures by their dependencies.
// Struct's dependencies always comes first itself.
func order_structures(structs []*sema.Struct) {
	n := len(structs)
	for i := 0; i < n; i++ {
		swapped := false

		for j := 0; j < n-i-1; j++ {
			curr := &structs[j]
			if !is_struct_ordered(structs, *curr) {
				next := &structs[j+1]
				*curr, *next = *next, *curr
				swapped = true
			}
		}

		if !swapped {
			break
		}
	}
}

// Reports whether variable in correct order by dependencies.
func is_var_ordered(vars []*sema.Var, v *sema.Var) bool {
	for _, d := range v.Depends {
		for _, vv := range vars {
			if vv == v {
				return false
			} else if vv == d {
				break
			}
		}
	}

	return true
}

// Orders variables by their dependencies.
// Variable's dependencies always comes first itself.
func order_variables(vars []*sema.Var) {
	n := len(vars)
	for i := 0; i < n; i++ {
		swapped := false

		for j := 0; j < n-i-1; j++ {
			curr := &vars[j]
			if !is_var_ordered(vars, *curr) {
				next := &vars[j+1]
				*curr, *next = *next, *curr
				swapped = true
			}
		}

		if !swapped {
			break
		}
	}
}
