package parser

import "github.com/julelang/jule/ast/models"

func can_be_order(s *models.Struct) bool {
	for _, d := range s.Origin.Depends {
		if d.Origin.Order < s.Origin.Order {
			return false
		}
	}
	return true
}

func order_structures(structures []*models.Struct) {
	for i, s := range structures {
		s.Order = i
	}

	n := len(structures)
	for i := 0; i < n; i++ {
		swapped := false
		for j := 0; j < n - i - 1; j++ {
			curr := &structures[j]
			if can_be_order(*curr) {
				(*curr).Origin.Order = j+1
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
