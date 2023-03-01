package types

import "github.com/julelang/jule/ast"

// IsStructOrdered Reports whether struct in correct order by dependencies.
func IsStructOrdered(s *ast.Struct) bool {
	for _, d := range s.Origin.Depends {
		if d.Origin.Order > s.Origin.Order {
			return true
		}
	}
	return false
}

// OrderStructures orders structures by their dependencies.
// Struct's dependencies always comes first itself.
func OrderStructures(structures []*ast.Struct) {
	for i, s := range structures {
		s.Order = i
	}

	n := len(structures)
	for i := 0; i < n; i++ {
		swapped := false
		for j := 0; j < n-i-1; j++ {
			curr := &structures[j]
			if !IsStructOrdered(*curr) {
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
